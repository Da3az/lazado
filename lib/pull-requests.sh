#!/usr/bin/env bash
# =============================================================================
# lazado — pull-requests.sh
# Pull request commands + fzf navigator functions
# =============================================================================

# =============================================================================
# DIRECT COMMANDS
# =============================================================================

# List open PRs for current repo
ado-prs() {
  local repo org project
  repo=$(_lazado_repo_name) || return 1
  org=$(_lazado_org) || return 1
  project=$(_lazado_project) || return 1
  az repos pr list \
    --repository "$repo" \
    --org "$org" \
    --project "$project" \
    --status active \
    --output table
}

# List my open PRs across all repos
ado-my-prs() {
  local org project
  org=$(_lazado_org) || return 1
  project=$(_lazado_project) || return 1
  az repos pr list \
    --org "$org" \
    --project "$project" \
    --creator "$(_lazado_me)" \
    --status active \
    --output table
}

# Show PR details
ado-pr() {
  local id="${1:?Usage: ado-pr <pr-id>}"
  az repos pr show --id "$id" \
    --org "$(_lazado_org)" \
    --output table
}

# Create a PR from current branch
ado-pr-create() {
  local repo target branch title work_item_id org project
  org=$(_lazado_org) || return 1
  project=$(_lazado_project) || return 1
  repo=$(_lazado_repo_name) || return 1
  target="${1:-$LAZADO_DEFAULT_TARGET_BRANCH}"
  branch=$(git branch --show-current)

  read -rp "PR title: " title
  if [[ -z "$title" ]]; then
    _lazado_err "Aborted: title is required."
    return 1
  fi

  work_item_id=$(_lazado_branch_work_item)

  local cmd=(az repos pr create
    --repository "$repo"
    --source-branch "$branch"
    --target-branch "$target"
    --title "$title"
    --org "$org"
    --project "$project"
    --output table
  )

  if [[ -n "$work_item_id" ]]; then
    cmd+=(--work-items "$work_item_id")
    _lazado_info "Linking work item #$work_item_id"
  fi

  "${cmd[@]}"
}

# Approve a PR
ado-pr-approve() {
  local id="${1:?Usage: ado-pr-approve <pr-id>}"
  az repos pr set-vote --id "$id" --vote approve \
    --org "$(_lazado_org)"
}

# Complete (merge) a PR
ado-pr-complete() {
  local id="${1:?Usage: ado-pr-complete <pr-id>}"
  az repos pr update --id "$id" --status completed \
    --org "$(_lazado_org)"
}

# Open PR in browser
ado-pr-web() {
  local id="${1:?Usage: ado-pr-web <pr-id>}"
  local repo
  repo=$(_lazado_repo_name) || return 1
  _lazado_open_url "$(_lazado_org)/$(_lazado_project)/_git/$repo/pullrequest/$id"
}

# =============================================================================
# NAVIGATOR FUNCTIONS
# =============================================================================

# Browse PRs for current repo
_lazado_nav_prs() {
  local repo org project
  repo=$(_lazado_repo_name) || return 1
  org=$(_lazado_org) || return 1
  project=$(_lazado_project) || return 1

  _lazado_info "Fetching PRs for $repo..."
  local prs
  prs=$(az repos pr list \
    --repository "$repo" \
    --org "$org" \
    --project "$project" \
    --status active \
    --output json 2>/dev/null)

  if [[ -z "$prs" || "$prs" == "[]" ]]; then
    echo "No active PRs for $repo."
    read -rp "Press Enter to go back..." _
    ado
    return 0
  fi

  local selected
  selected=$(echo "$prs" | \
    jq -r '.[] | "\(.pullRequestId)\t\(.createdBy.displayName)\t\(.sourceRefName | sub("refs/heads/"; ""))\t\(.title)"' | \
    column -t -s$'\t' | \
    fzf --header="Pull Requests  ─  $repo" \
        --prompt="PR > " \
        --height=~60% --reverse --border $LAZADO_FZF_OPTS)

  [[ -z "$selected" ]] && ado && return 0

  local pr_id
  pr_id=$(echo "$selected" | awk '{print $1}')
  _lazado_nav_pr_actions "$pr_id"
}

# Browse my PRs across all repos
_lazado_nav_my_prs() {
  local org project
  org=$(_lazado_org) || return 1
  project=$(_lazado_project) || return 1

  _lazado_info "Fetching your PRs..."
  local prs
  prs=$(az repos pr list \
    --org "$org" \
    --project "$project" \
    --creator "$(_lazado_me)" \
    --status active \
    --output json 2>/dev/null)

  if [[ -z "$prs" || "$prs" == "[]" ]]; then
    echo "No active PRs created by you."
    read -rp "Press Enter to go back..." _
    ado
    return 0
  fi

  local selected
  selected=$(echo "$prs" | \
    jq -r '.[] | "\(.pullRequestId)\t\(.repository.name)\t\(.sourceRefName | sub("refs/heads/"; ""))\t\(.title)"' | \
    column -t -s$'\t' | \
    fzf --header="My Pull Requests  ─  All Repos" \
        --prompt="PR > " \
        --height=~60% --reverse --border $LAZADO_FZF_OPTS)

  [[ -z "$selected" ]] && ado && return 0

  local pr_id
  pr_id=$(echo "$selected" | awk '{print $1}')
  _lazado_nav_pr_actions "$pr_id"
}

# PR actions menu
_lazado_nav_pr_actions() {
  local pr_id="$1"

  echo ""
  az repos pr show --id "$pr_id" \
    --org "$(_lazado_org)" \
    --output json 2>/dev/null | jq '{
      id: .pullRequestId,
      title: .title,
      status: .status,
      source: (.sourceRefName | sub("refs/heads/"; "")),
      target: (.targetRefName | sub("refs/heads/"; "")),
      createdBy: .createdBy.displayName,
      reviewers: [.reviewers[]?.displayName] | join(", "),
      mergeStatus: .mergeStatus
    }'
  echo ""

  local action
  action=$(printf '%s\n' \
    "Approve" \
    "Complete (Merge)" \
    "Checkout Source Branch" \
    "Open in Browser" \
    "Back to Menu" \
    | fzf --header="PR #$pr_id  ─  Actions" \
          --prompt="Action > " \
          --height=~40% --reverse --border $LAZADO_FZF_OPTS)

  case "$action" in
    "Approve")
      ado-pr-approve "$pr_id"
      echo "PR #$pr_id approved." ;;
    "Complete (Merge)")
      read -rp "Are you sure you want to complete PR #$pr_id? [y/N]: " confirm
      [[ "$confirm" =~ ^[Yy]$ ]] && ado-pr-complete "$pr_id" ;;
    "Checkout Source Branch")
      local branch
      branch=$(az repos pr show --id "$pr_id" --org "$(_lazado_org)" --output json 2>/dev/null | \
        jq -r '.sourceRefName | sub("refs/heads/"; "")')
      git fetch origin && git checkout "$branch" ;;
    "Open in Browser")
      ado-pr-web "$pr_id" ;;
    "Back to Menu"|"") ado ;;
  esac
}
