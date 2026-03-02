#!/usr/bin/env bash
# =============================================================================
# lazado — navigator.sh
# Main interactive fzf menu + help command
# =============================================================================

# Main entry point — type `ado` to launch the interactive navigator
ado() {
  local repo_name
  repo_name=$(basename "$(git rev-parse --show-toplevel 2>/dev/null)" 2>/dev/null || echo "no repo")
  local project="${LAZADO_PROJECT:-not configured}"

  local section
  section=$(printf '%s\n' \
    "My Work Items" \
    "My User Stories" \
    "Create Work Item" \
    "Create User Story" \
    "Update Work Item" \
    "Update User Story" \
    "Update Item Status" \
    "Search Items" \
    "Current Branch Ticket" \
    "─────────────────────" \
    "Pull Requests (this repo)" \
    "My Pull Requests (all repos)" \
    "Create Pull Request" \
    "─────────────────────" \
    "Pipelines" \
    "Repos" \
    "Open in Browser" \
    | fzf --header="lazado  ─  $repo_name  ─  $project" \
          --prompt="Navigate > " \
          --height=~50% --reverse --border --border-label=" ado " $LAZADO_FZF_OPTS)

  [[ -z "$section" || "$section" == "─────────────────────" ]] && return 0

  case "$section" in
    "My Work Items")                _lazado_nav_list_items "Task" ;;
    "My User Stories")              _lazado_nav_list_items "User Story" ;;
    "Create Work Item")             _lazado_nav_create_item "Task" ;;
    "Create User Story")            _lazado_nav_create_item "User Story" ;;
    "Update Work Item")             _lazado_nav_update_item "Task" ;;
    "Update User Story")            _lazado_nav_update_item "User Story" ;;
    "Update Item Status")           _lazado_nav_update_status_standalone ;;
    "Search Items")                 _lazado_nav_search ;;
    "Current Branch Ticket")        ado-wi-branch ;;
    "Pull Requests (this repo)")    _lazado_nav_prs ;;
    "My Pull Requests (all repos)") _lazado_nav_my_prs ;;
    "Create Pull Request")          ado-pr-create ;;
    "Pipelines")                    ado-pipelines ;;
    "Repos")                        _lazado_nav_repos ;;
    "Open in Browser")              ado-web ;;
  esac
}

# Help command
ado-help() {
  cat <<'HELP'
lazado — Interactive Azure DevOps CLI

INTERACTIVE:
  ado                        Launch interactive navigator (fzf)

WORK ITEMS / USER STORIES:
  ado-wi <id>                View work item
  ado-wi-detail <id>         View work item (full detail)
  ado-wi-branch              View work item linked to current branch
  ado-wi-search <keyword>    Search work items by title
  ado-wi-update <id>         Update work item fields (title, desc, assignment)
  ado-wi-state <id> <state>  Update work item state

PULL REQUESTS:
  ado-prs                    List open PRs for current repo
  ado-my-prs                 List my open PRs across all repos
  ado-pr <id>                Show PR details
  ado-pr-create [target]     Create PR from current branch
  ado-pr-approve <id>        Approve a PR
  ado-pr-complete <id>       Complete/merge a PR
  ado-pr-web <id>            Open PR in browser

PIPELINES:
  ado-pipelines              List recent pipeline runs

REPOS:
  ado-repos                  List all repos in the project
  ado-web                    Open current repo in browser

CONFIG:
  ado-init                   Setup wizard (configure org, project, etc.)
  ado-refresh-states         Refresh cached work item states from API

MORE INFO:
  https://github.com/da3az/lazado
HELP
}
