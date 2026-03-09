#!/usr/bin/env bash
# =============================================================================
# lazado — states.sh
# Dynamic work item state discovery via Azure DevOps REST API with caching
#
# Fallback chain:
#   1. Fresh cache (within TTL)
#   2. API call (if online and authenticated)
#   3. Stale cache (if API fails)
#   4. Built-in defaults (Agile process template)
# =============================================================================

# -- Built-in defaults (Azure DevOps Agile process template) -----------------
_LAZADO_DEFAULT_STATES_Task='[
  {"name":"New","category":"Proposed"},
  {"name":"Active","category":"InProgress"},
  {"name":"Closed","category":"Completed"},
  {"name":"Removed","category":"Removed"}
]'

_LAZADO_DEFAULT_STATES_UserStory='[
  {"name":"New","category":"Proposed"},
  {"name":"Active","category":"InProgress"},
  {"name":"Resolved","category":"Resolved"},
  {"name":"Closed","category":"Completed"},
  {"name":"Removed","category":"Removed"}
]'

_LAZADO_DEFAULT_STATES_Bug='[
  {"name":"New","category":"Proposed"},
  {"name":"Active","category":"InProgress"},
  {"name":"Resolved","category":"Resolved"},
  {"name":"Closed","category":"Completed"}
]'

_LAZADO_DEFAULT_STATES_Epic='[
  {"name":"New","category":"Proposed"},
  {"name":"Active","category":"InProgress"},
  {"name":"Resolved","category":"Resolved"},
  {"name":"Closed","category":"Completed"}
]'

_LAZADO_DEFAULT_STATES_Feature='[
  {"name":"New","category":"Proposed"},
  {"name":"Active","category":"InProgress"},
  {"name":"Resolved","category":"Resolved"},
  {"name":"Closed","category":"Completed"}
]'

# -- Cache helpers ------------------------------------------------------------

_lazado_cache_path() {
  local work_item_type="$1"
  local safe_type
  safe_type=$(echo "$work_item_type" | tr ' ' '_')
  echo "${LAZADO_CACHE_DIR}/states/${safe_type}.json"
}

_lazado_cache_fresh() {
  local cache_file="$1"
  [[ -f "$cache_file" ]] || return 1

  local ttl="${LAZADO_CACHE_TTL:-86400}"
  local now file_age
  now=$(date +%s)
  file_age=$(stat -c %Y "$cache_file" 2>/dev/null || stat -f %m "$cache_file" 2>/dev/null || echo 0)
  (( now - file_age < ttl ))
}

# -- Fetch states from Azure DevOps REST API ---------------------------------

_lazado_fetch_states() {
  local work_item_type="$1"
  local org project encoded_type
  org=$(_lazado_org) || return 1
  project=$(_lazado_project) || return 1
  encoded_type=$(printf '%s' "$work_item_type" | sed 's/ /%20/g')

  az rest --method get \
    --uri "${org}/${project}/_apis/wit/workitemtypes/${encoded_type}/states?api-version=7.1-preview" \
    2>/dev/null | jq '[.value[] | {name: .name, category: .category}]' 2>/dev/null
}

# -- Cache states to disk ----------------------------------------------------

_lazado_cache_states() {
  local work_item_type="$1"
  local data="$2"
  local cache_file
  cache_file=$(_lazado_cache_path "$work_item_type")

  mkdir -p "$(dirname "$cache_file")"
  echo "$data" > "$cache_file"
}

# -- Get states (with full fallback chain) ------------------------------------

_lazado_get_states() {
  local work_item_type="$1"
  local cache_file states

  cache_file=$(_lazado_cache_path "$work_item_type")

  # 1. Fresh cache
  if _lazado_cache_fresh "$cache_file"; then
    cat "$cache_file"
    return 0
  fi

  # 2. API call
  states=$(_lazado_fetch_states "$work_item_type" 2>/dev/null)
  if [[ -n "$states" && "$states" != "null" && "$states" != "[]" ]]; then
    _lazado_cache_states "$work_item_type" "$states"
    echo "$states"
    return 0
  fi

  # 3. Stale cache
  if [[ -f "$cache_file" ]]; then
    cat "$cache_file"
    return 0
  fi

  # 4. Built-in defaults
  local safe_type
  safe_type=$(echo "$work_item_type" | tr ' ' '_' | tr -d '-')
  local var_name="_LAZADO_DEFAULT_STATES_${safe_type}"
  local defaults="${!var_name}"

  if [[ -n "$defaults" ]]; then
    echo "$defaults"
    return 0
  fi

  # Unknown type — return minimal default
  echo '[{"name":"New","category":"Proposed"},{"name":"Active","category":"InProgress"},{"name":"Closed","category":"Completed"}]'
}

# -- Filter states by category -----------------------------------------------

# Get active states (Proposed + InProgress)
_lazado_states_active() {
  local work_item_type="$1"
  _lazado_get_states "$work_item_type" | \
    jq -r '[.[] | select(.category == "Proposed" or .category == "InProgress")] | .[].name'
}

# Get done states (Completed + Resolved + Removed)
_lazado_states_done() {
  local work_item_type="$1"
  _lazado_get_states "$work_item_type" | \
    jq -r '[.[] | select(.category == "Completed" or .category == "Resolved" or .category == "Removed")] | .[].name'
}

# Get all state names
_lazado_states_all() {
  local work_item_type="$1"
  _lazado_get_states "$work_item_type" | jq -r '.[].name'
}

# -- Build WIQL state filter clause ------------------------------------------

_lazado_state_filter() {
  local filter_type="$1"   # "New/Active", "Done", or "All"
  local work_item_type="$2"
  local names=()

  if [[ "$filter_type" == "New/Active" ]]; then
    mapfile -t names < <(_lazado_states_active "$work_item_type")
  elif [[ "$filter_type" == "Done" ]]; then
    mapfile -t names < <(_lazado_states_done "$work_item_type")
  else
    # All — no filter needed
    echo ""
    return 0
  fi

  if [[ ${#names[@]} -eq 0 ]]; then
    echo ""
    return 0
  fi

  local clauses=()
  for s in "${names[@]}"; do
    clauses+=("[System.State] = '$s'")
  done

  local joined
  joined=$(IFS=' OR '; echo "${clauses[*]}")
  echo "AND ($joined)"
}

# -- User-facing command to refresh cache ------------------------------------

ado-refresh-states() {
  local org project
  org=$(_lazado_org) || return 1
  project=$(_lazado_project) || return 1

  local types=("Task" "User Story" "Bug" "Epic" "Feature")
  local type states

  _lazado_info "Refreshing work item states from ${org}..."
  for type in "${types[@]}"; do
    states=$(_lazado_fetch_states "$type" 2>/dev/null)
    if [[ -n "$states" && "$states" != "null" && "$states" != "[]" ]]; then
      _lazado_cache_states "$type" "$states"
      local count
      count=$(echo "$states" | jq 'length')
      echo "  $type: $count states cached"
    else
      echo "  $type: skipped (not found or no access)"
    fi
  done
  _lazado_info "Done."
}
