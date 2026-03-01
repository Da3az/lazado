#!/usr/bin/env bash
# =============================================================================
# lazado — utils.sh
# Shared helper functions
# =============================================================================

# -- Output helpers -----------------------------------------------------------
_lazado_err() {
  echo -e "\033[31mlazado: $*\033[0m" >&2
}

_lazado_warn() {
  echo -e "\033[33mlazado: $*\033[0m" >&2
}

_lazado_info() {
  echo -e "\033[36m$*\033[0m" >&2
}

# -- Dependency checking ------------------------------------------------------
_lazado_require_cmd() {
  local cmd="$1"
  local hint="${2:-}"
  if ! command -v "$cmd" &>/dev/null; then
    _lazado_err "'$cmd' is required but not found."
    [[ -n "$hint" ]] && _lazado_err "  $hint"
    return 1
  fi
}

_lazado_check_deps() {
  local missing=0
  _lazado_require_cmd az "Install: https://learn.microsoft.com/en-us/cli/azure/install-azure-cli" || ((missing++))
  _lazado_require_cmd jq "Install: sudo apt install jq  (or brew install jq)" || ((missing++))
  _lazado_require_cmd fzf "Install: https://github.com/junegunn/fzf#installation" || ((missing++))
  _lazado_require_cmd git "Install: sudo apt install git  (or brew install git)" || ((missing++))

  if ((missing > 0)); then
    _lazado_err "$missing required tool(s) missing."
    return 1
  fi

  # Check azure-devops extension
  if ! az extension show --name azure-devops &>/dev/null 2>&1; then
    _lazado_warn "Azure DevOps CLI extension not found. Install with: az extension add --name azure-devops"
  fi
}

# -- Git helpers --------------------------------------------------------------

# Extract repo name from git remote origin URL
_lazado_repo_name() {
  local url
  url=$(git config --get remote.origin.url 2>/dev/null)
  if [[ -z "$url" ]]; then
    _lazado_err "Not in a git repo with an origin remote."
    return 1
  fi
  echo "$url" | sed -E 's#.*/##' | sed 's/\.git$//'
}

# Extract work item ID from branch name
# Supports patterns like: feature/1234-description, solve/4089-prod, bug/5678
_lazado_branch_work_item() {
  local branch
  branch=$(git branch --show-current 2>/dev/null)
  echo "$branch" | grep -oP "${LAZADO_BRANCH_ID_PATTERN}" | head -1
}

# -- Azure DevOps helpers ----------------------------------------------------

# Get current authenticated user
_lazado_me() {
  az account show --query user.name -o tsv 2>/dev/null
}

# Open a URL in the user's browser
_lazado_open_url() {
  local url="$1"
  if [[ -n "${LAZADO_BROWSER_CMD:-}" ]]; then
    "$LAZADO_BROWSER_CMD" "$url" 2>/dev/null
  else
    _lazado_err "No browser command found. Set LAZADO_BROWSER_CMD in your config."
    _lazado_info "URL: $url"
  fi
}

# Build SSH clone URL from template
_lazado_ssh_url() {
  local repo="$1"
  local org_name

  # Extract org name from URL (https://dev.azure.com/OrgName -> OrgName)
  org_name=$(echo "$LAZADO_ORG" | sed -E 's#.*/##')

  echo "$LAZADO_SSH_URL_TEMPLATE" \
    | sed "s/{org}/$org_name/g" \
    | sed "s/{project}/$LAZADO_PROJECT/g" \
    | sed "s/{repo}/$repo/g"
}

# Check deps on load (warn only, don't block sourcing)
_lazado_check_deps 2>/dev/null || true
