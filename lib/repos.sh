#!/usr/bin/env bash
# =============================================================================
# lazado — repos.sh
# Repository commands + fzf navigator
# =============================================================================

# List all repos in the project
ado-repos() {
  local org project
  org=$(_lazado_org) || return 1
  project=$(_lazado_project) || return 1
  az repos list \
    --org "$org" \
    --project "$project" \
    --output table
}

# Open current repo in browser
ado-web() {
  local repo
  repo=$(_lazado_repo_name) || return 1
  _lazado_open_url "$(_lazado_org)/$(_lazado_project)/_git/$repo"
}

# Browse repos interactively
_lazado_nav_repos() {
  local org project
  org=$(_lazado_org) || return 1
  project=$(_lazado_project) || return 1

  _lazado_info "Fetching repos..."
  local repos
  repos=$(az repos list \
    --org "$org" \
    --project "$project" \
    --output json 2>/dev/null)

  if [[ -z "$repos" || "$repos" == "[]" ]]; then
    echo "No repos found."
    return 0
  fi

  local selected
  selected=$(echo "$repos" | \
    jq -r '.[] | "\(.name)\t\(.defaultBranch // "n/a")\t\(.size)"' | \
    column -t -s$'\t' | \
    fzf --header="Repos  ─  $project" \
        --prompt="Repo > " \
        --height=~50% --reverse --border $LAZADO_FZF_OPTS)

  [[ -z "$selected" ]] && ado && return 0

  local repo_name
  repo_name=$(echo "$selected" | awk '{print $1}')

  local action
  action=$(printf '%s\n' \
    "Open in Browser" \
    "Clone (SSH)" \
    "Back to Menu" \
    | fzf --header="$repo_name  ─  Actions" \
          --prompt="Action > " \
          --height=~30% --reverse --border $LAZADO_FZF_OPTS)

  case "$action" in
    "Open in Browser")
      _lazado_open_url "$org/$project/_git/$repo_name" ;;
    "Clone (SSH)")
      local target_dir="$LAZADO_CLONE_DIR/$repo_name"
      read -rp "Clone to [$target_dir]: " custom_dir
      target_dir="${custom_dir:-$target_dir}"
      local ssh_url
      ssh_url=$(_lazado_ssh_url "$repo_name")
      git clone "$ssh_url" "$target_dir" ;;
    "Back to Menu"|"") ado ;;
  esac
}
