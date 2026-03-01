#!/usr/bin/env bash
# =============================================================================
# lazado — config.sh
# Configuration loading with hierarchy:
#   1. Built-in defaults
#   2. Global config:  ~/.config/lazado/config
#   3. Project config: .adoconfig in git repo root
#   4. Environment variables (LAZADO_*)
#   5. Auto-detect from git remote URL
# =============================================================================

# -- Parse a key=value config file safely (no eval/source) -------------------
_lazado_parse_config() {
  local file="$1"
  [[ -f "$file" ]] || return 0

  local line key value
  while IFS= read -r line || [[ -n "$line" ]]; do
    # skip comments and blank lines
    [[ "$line" =~ ^[[:space:]]*# ]] && continue
    [[ "$line" =~ ^[[:space:]]*$ ]] && continue

    key="${line%%=*}"
    value="${line#*=}"

    # trim whitespace
    key="${key#"${key%%[![:space:]]*}"}"
    key="${key%"${key##*[![:space:]]}"}"
    value="${value#"${value%%[![:space:]]*}"}"
    value="${value%"${value##*[![:space:]]}"}"

    # strip surrounding quotes from value
    if [[ "$value" =~ ^\"(.*)\"$ ]] || [[ "$value" =~ ^\'(.*)\'$ ]]; then
      value="${BASH_REMATCH[1]}"
    fi

    # expand $HOME in values
    value="${value//\$HOME/$HOME}"
    value="${value//\~/$HOME}"

    # map config keys to LAZADO_ variables (only if env var not already set)
    case "$key" in
      org)                    LAZADO_ORG="${LAZADO_ORG:-$value}" ;;
      project)                LAZADO_PROJECT="${LAZADO_PROJECT:-$value}" ;;
      clone_dir)              LAZADO_CLONE_DIR="${LAZADO_CLONE_DIR:-$value}" ;;
      branch_prefix)          LAZADO_BRANCH_PREFIX="${LAZADO_BRANCH_PREFIX:-$value}" ;;
      branch_id_pattern)      LAZADO_BRANCH_ID_PATTERN="${LAZADO_BRANCH_ID_PATTERN:-$value}" ;;
      default_target_branch)  LAZADO_DEFAULT_TARGET_BRANCH="${LAZADO_DEFAULT_TARGET_BRANCH:-$value}" ;;
      cache_ttl)              LAZADO_CACHE_TTL="${LAZADO_CACHE_TTL:-$value}" ;;
      cache_dir)              LAZADO_CACHE_DIR="${LAZADO_CACHE_DIR:-$value}" ;;
      browser_cmd)            LAZADO_BROWSER_CMD="${LAZADO_BROWSER_CMD:-$value}" ;;
      fzf_opts)               LAZADO_FZF_OPTS="${LAZADO_FZF_OPTS:-$value}" ;;
      ssh_url_template)       LAZADO_SSH_URL_TEMPLATE="${LAZADO_SSH_URL_TEMPLATE:-$value}" ;;
    esac
  done < "$file"
}

# -- Detect org and project from git remote URL ------------------------------
_lazado_detect_org_project() {
  local url
  url=$(git config --get remote.origin.url 2>/dev/null) || return 1

  local org project

  # SSH: git@ssh.dev.azure.com:v3/OrgName/ProjectName/RepoName
  if [[ "$url" =~ ssh\.dev\.azure\.com.*v3/([^/]+)/([^/]+)/ ]]; then
    org="${BASH_REMATCH[1]}"
    project="${BASH_REMATCH[2]}"
  # HTTPS: https://dev.azure.com/OrgName/ProjectName/_git/RepoName
  elif [[ "$url" =~ dev\.azure\.com/([^/]+)/([^/]+)/ ]]; then
    org="${BASH_REMATCH[1]}"
    project="${BASH_REMATCH[2]}"
  # HTTPS (older): https://OrgName.visualstudio.com/ProjectName/_git/RepoName
  elif [[ "$url" =~ ([^/]+)\.visualstudio\.com/([^/]+)/ ]]; then
    org="${BASH_REMATCH[1]}"
    project="${BASH_REMATCH[2]}"
  else
    return 1
  fi

  [[ -z "$LAZADO_ORG" ]] && LAZADO_ORG="https://dev.azure.com/$org"
  [[ -z "$LAZADO_PROJECT" ]] && LAZADO_PROJECT="$project"
}

# -- Main config loader ------------------------------------------------------
_lazado_load_config() {
  # 1. Built-in defaults (only set if not already in environment)
  LAZADO_CLONE_DIR="${LAZADO_CLONE_DIR:-$HOME/dev}"
  LAZADO_BRANCH_PREFIX="${LAZADO_BRANCH_PREFIX:-feature}"
  LAZADO_BRANCH_ID_PATTERN="${LAZADO_BRANCH_ID_PATTERN:-(?<=/)\d+}"
  LAZADO_DEFAULT_TARGET_BRANCH="${LAZADO_DEFAULT_TARGET_BRANCH:-main}"
  LAZADO_CACHE_TTL="${LAZADO_CACHE_TTL:-86400}"
  LAZADO_CACHE_DIR="${LAZADO_CACHE_DIR:-$HOME/.cache/lazado}"
  LAZADO_FZF_OPTS="${LAZADO_FZF_OPTS:-}"
  LAZADO_SSH_URL_TEMPLATE="${LAZADO_SSH_URL_TEMPLATE:-git@ssh.dev.azure.com:v3/{org}/{project}/{repo}}"

  # Auto-detect browser command
  if [[ -z "${LAZADO_BROWSER_CMD:-}" ]]; then
    if command -v xdg-open &>/dev/null; then
      LAZADO_BROWSER_CMD="xdg-open"
    elif command -v open &>/dev/null; then
      LAZADO_BROWSER_CMD="open"
    fi
  fi

  # 2. Load global config
  local global_config="${XDG_CONFIG_HOME:-$HOME/.config}/lazado/config"
  _lazado_parse_config "$global_config"

  # 3. Load project-level config (find .adoconfig at git root)
  local git_root
  git_root=$(git rev-parse --show-toplevel 2>/dev/null)
  if [[ -n "$git_root" && -f "$git_root/.adoconfig" ]]; then
    _lazado_parse_config "$git_root/.adoconfig"
  fi

  # 4. Environment variables already take priority (via ${VAR:-} pattern above)

  # 5. Auto-detect from git remote if org/project still unset
  if [[ -z "${LAZADO_ORG:-}" || -z "${LAZADO_PROJECT:-}" ]]; then
    _lazado_detect_org_project 2>/dev/null
  fi
}

# -- Getters with validation -------------------------------------------------
_lazado_org() {
  if [[ -z "${LAZADO_ORG:-}" ]]; then
    _lazado_err "Organization not configured. Run 'ado-init' or set LAZADO_ORG."
    return 1
  fi
  echo "$LAZADO_ORG"
}

_lazado_project() {
  if [[ -z "${LAZADO_PROJECT:-}" ]]; then
    _lazado_err "Project not configured. Run 'ado-init' or set LAZADO_PROJECT."
    return 1
  fi
  echo "$LAZADO_PROJECT"
}

# Load config on source
_lazado_load_config
