#!/usr/bin/env bash
# =============================================================================
# lazado — work-items.sh
# Work item and user story commands + fzf navigator functions
# =============================================================================

# =============================================================================
# DIRECT COMMANDS
# =============================================================================

# View a work item by ID (table format)
ado-wi() {
  local id="${1:?Usage: ado-wi <work-item-id>}"
  az boards work-item show --id "$id" \
    --org "$(_lazado_org)" \
    --output table
}

# View full detail of a work item (formatted JSON)
ado-wi-detail() {
  local id="${1:?Usage: ado-wi-detail <work-item-id>}"
  az boards work-item show --id "$id" \
    --org "$(_lazado_org)" \
    --output json | jq '{
      id: .id,
      title: .fields["System.Title"],
      state: .fields["System.State"],
      type: .fields["System.WorkItemType"],
      assignedTo: (.fields["System.AssignedTo"].displayName // "Unassigned"),
      description: .fields["System.Description"]
    }'
}

# View work item linked to current branch (extracts ID from branch name)
ado-wi-branch() {
  local id
  id=$(_lazado_branch_work_item)
  if [[ -z "$id" ]]; then
    _lazado_err "No work item ID found in branch name: $(git branch --show-current 2>/dev/null)"
    return 1
  fi
  ado-wi-detail "$id"
}

# Update work item state
ado-wi-state() {
  local id="${1:?Usage: ado-wi-state <work-item-id> <state>}"
  local state="${2:?Usage: ado-wi-state <work-item-id> <state>}"
  az boards work-item update --id "$id" \
    --state "$state" \
    --org "$(_lazado_org)" \
    --output table
}

# Search work items by title keyword
ado-wi-search() {
  local keyword="${1:?Usage: ado-wi-search <keyword>}"
  local org project
  org=$(_lazado_org) || return 1
  project=$(_lazado_project) || return 1
  az boards query \
    --wiql "SELECT [System.Id], [System.Title], [System.State], [System.WorkItemType] FROM WorkItems WHERE [System.Title] CONTAINS '$keyword' AND [System.State] <> 'Closed' ORDER BY [System.CreatedDate] DESC" \
    --org "$org" \
    --project "$project" \
    --output table
}

# =============================================================================
# NAVIGATOR FUNCTIONS
# =============================================================================

# List items with status filter — shared for all work item types
_lazado_nav_list_items() {
  local item_type="$1"  # "Task", "User Story", "Bug", etc.
  local label="$item_type"
  [[ "$item_type" == "Task" ]] && label="Work Items"
  [[ "$item_type" == "User Story" ]] && label="User Stories"

  local org project
  org=$(_lazado_org) || return 1
  project=$(_lazado_project) || return 1

  # Status filter
  local filter
  filter=$(printf '%s\n' \
    "New/Active" \
    "Done" \
    "All" \
    | fzf --header="$label  ─  Filter by status" \
          --prompt="Filter > " \
          --height=~30% --reverse --border $LAZADO_FZF_OPTS)

  [[ -z "$filter" ]] && ado && return 0

  local state_clause
  state_clause=$(_lazado_state_filter "$filter" "$item_type")

  _lazado_info "Fetching $label ($filter)..."
  local items
  items=$(az boards query \
    --wiql "SELECT [System.Id], [System.Title], [System.State], [System.WorkItemType], [System.CreatedDate] FROM WorkItems WHERE [System.AssignedTo] = @Me AND [System.WorkItemType] = '$item_type' $state_clause ORDER BY [System.CreatedDate] DESC" \
    --org "$org" \
    --project "$project" \
    --output json 2>/dev/null)

  if [[ -z "$items" || "$items" == "[]" ]]; then
    echo "No $label found ($filter)."
    read -rp "Press Enter to go back..." _
    ado
    return 0
  fi

  local selected
  selected=$(echo "$items" | \
    jq -r '.[] | "\(.id)\t\(.fields["System.State"])\t\(.fields["System.Title"])"' | \
    column -t -s$'\t' | \
    fzf --header="$label ($filter)  ─  Select to view, ESC to go back" \
        --prompt="$label > " \
        --height=~60% --reverse --border $LAZADO_FZF_OPTS)

  [[ -z "$selected" ]] && ado && return 0

  local id
  id=$(echo "$selected" | awk '{print $1}')
  _lazado_nav_item_actions "$id" "$item_type"
}

# Actions menu for a single work item
_lazado_nav_item_actions() {
  local id="$1"
  local item_type="$2"

  echo ""
  ado-wi-detail "$id"
  echo ""

  local action
  action=$(printf '%s\n' \
    "Update Status" \
    "Open in Browser" \
    "Create Branch" \
    "Back to Menu" \
    | fzf --header="#$id  ─  Actions" \
          --prompt="Action > " \
          --height=~40% --reverse --border $LAZADO_FZF_OPTS)

  case "$action" in
    "Update Status")  _lazado_nav_update_state "$id" "$item_type" ;;
    "Open in Browser")
      _lazado_open_url "$(_lazado_org)/$(_lazado_project)/_workitems/edit/$id" ;;
    "Create Branch")
      local title
      title=$(az boards work-item show --id "$id" --org "$(_lazado_org)" --output json 2>/dev/null | \
        jq -r '.fields["System.Title"]' | \
        tr '[:upper:]' '[:lower:]' | sed -E 's/[^a-z0-9]+/-/g; s/^-|-$//g')
      local branch_name="${LAZADO_BRANCH_PREFIX}/${id}-${title}"
      read -rp "Branch name [$branch_name]: " custom_name
      branch_name="${custom_name:-$branch_name}"
      git checkout -b "$branch_name"
      ;;
    "Back to Menu"|"") ado ;;
  esac
}

# Context-aware state selector
_lazado_nav_update_state() {
  local id="$1"
  local item_type="$2"

  # If type is unknown, fetch it
  if [[ -z "$item_type" ]]; then
    item_type=$(az boards work-item show --id "$id" --org "$(_lazado_org)" --output json 2>/dev/null | \
      jq -r '.fields["System.WorkItemType"]')
  fi

  local state
  state=$(_lazado_states_all "$item_type" | \
    fzf --header="Set status for #$id ($item_type)" \
        --prompt="Status > " \
        --height=~40% --reverse --border $LAZADO_FZF_OPTS)

  if [[ -n "$state" ]]; then
    ado-wi-state "$id" "$state"
    echo "Updated #$id to: $state"
  fi
}

# Standalone status update — enter any ID
_lazado_nav_update_status_standalone() {
  local id
  read -rp "Work item ID: " id
  [[ -z "$id" ]] && ado && return 0

  echo ""
  ado-wi-detail "$id"
  echo ""

  local item_type
  item_type=$(az boards work-item show --id "$id" --org "$(_lazado_org)" --output json 2>/dev/null | \
    jq -r '.fields["System.WorkItemType"]')

  if [[ -z "$item_type" || "$item_type" == "null" ]]; then
    echo "Could not find item #$id"
    read -rp "Press Enter to go back..." _
    ado
    return 0
  fi

  _lazado_nav_update_state "$id" "$item_type"
}

# Create a new work item interactively
_lazado_nav_create_item() {
  local item_type="$1"
  local label="$item_type"
  [[ "$item_type" == "Task" ]] && label="Work Item"

  local org project
  org=$(_lazado_org) || return 1
  project=$(_lazado_project) || return 1

  # Title
  local title
  read -rp "$label title: " title
  if [[ -z "$title" ]]; then
    _lazado_err "Aborted: title is required."
    return 1
  fi

  # Description
  local description
  read -rp "Description (optional, Enter to skip): " description

  # State
  local state
  state=$(printf '%s\n' "New" "Active" "Closed" | \
    fzf --header="$label  ─  Initial status" \
        --prompt="Status > " \
        --height=~30% --reverse --border $LAZADO_FZF_OPTS)
  state="${state:-New}"

  # Assignment
  local assign_to
  assign_to=$(printf '%s\n' "Assign to me" "Unassigned" | \
    fzf --header="$label  ─  Assignment" \
        --prompt="Assign > " \
        --height=~25% --reverse --border $LAZADO_FZF_OPTS)

  # Build command
  local cmd=(az boards work-item create
    --type "$item_type"
    --title "$title"
    --org "$org"
    --project "$project"
    --output json
  )

  if [[ -n "$description" ]]; then
    cmd+=(--description "$description")
  fi

  if [[ "$assign_to" == "Assign to me" ]]; then
    cmd+=(--assigned-to "$(_lazado_me)")
  fi

  echo ""
  _lazado_info "Creating $label..."
  local result
  result=$("${cmd[@]}" 2>&1)

  local new_id
  new_id=$(echo "$result" | jq -r '.id' 2>/dev/null)

  if [[ -z "$new_id" || "$new_id" == "null" ]]; then
    _lazado_err "Failed to create $label:"
    echo "$result"
    return 1
  fi

  # Set state if not New
  if [[ "$state" != "New" ]]; then
    az boards work-item update --id "$new_id" \
      --state "$state" \
      --org "$org" \
      --output none 2>/dev/null
  fi

  echo ""
  echo "Created $label #$new_id"
  ado-wi-detail "$new_id"
}

# Search across all work item types
_lazado_nav_search() {
  local org project
  org=$(_lazado_org) || return 1
  project=$(_lazado_project) || return 1

  local keyword
  read -rp "Search keyword: " keyword
  [[ -z "$keyword" ]] && ado && return 0

  _lazado_info "Searching for '$keyword'..."
  local items
  items=$(az boards query \
    --wiql "SELECT [System.Id], [System.Title], [System.State], [System.WorkItemType] FROM WorkItems WHERE [System.Title] CONTAINS '$keyword' ORDER BY [System.CreatedDate] DESC" \
    --org "$org" \
    --project "$project" \
    --output json 2>/dev/null)

  if [[ -z "$items" || "$items" == "[]" ]]; then
    echo "No results for '$keyword'."
    read -rp "Press Enter to go back..." _
    ado
    return 0
  fi

  local selected
  selected=$(echo "$items" | \
    jq -r '.[] | "\(.id)\t\(.fields["System.WorkItemType"])\t\(.fields["System.State"])\t\(.fields["System.Title"])"' | \
    column -t -s$'\t' | \
    fzf --header="Search: '$keyword'  ─  Select to view" \
        --prompt="Result > " \
        --height=~60% --reverse --border $LAZADO_FZF_OPTS)

  [[ -z "$selected" ]] && ado && return 0

  local id item_type
  id=$(echo "$selected" | awk '{print $1}')
  item_type=$(echo "$selected" | awk '{print $2}')
  # Handle multi-word types like "User Story"
  if [[ "$item_type" == "User" ]]; then
    item_type="User Story"
  fi
  _lazado_nav_item_actions "$id" "$item_type"
}
