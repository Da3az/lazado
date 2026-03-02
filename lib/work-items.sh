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

# Update work item fields (title, description, assignment)
ado-wi-update() {
  local id="${1:?Usage: ado-wi-update <work-item-id>}"
  shift

  local org title description assigned_to
  org=$(_lazado_org) || return 1

  while [[ $# -gt 0 ]]; do
    case "$1" in
      --title)       title="$2"; shift 2 ;;
      --description) description="$2"; shift 2 ;;
      --assigned-to) assigned_to="$2"; shift 2 ;;
      *) _lazado_err "Unknown option: $1"; return 1 ;;
    esac
  done

  local cmd=(az boards work-item update --id "$id" --org "$org" --output table)
  local has_update=0

  if [[ -n "$title" ]]; then
    cmd+=(--title "$title")
    has_update=1
  fi
  if [[ -n "$description" ]]; then
    cmd+=(--description "$description")
    has_update=1
  fi
  if [[ -n "${assigned_to+set}" ]]; then
    cmd+=(--assigned-to "$assigned_to")
    has_update=1
  fi

  if [[ "$has_update" -eq 0 ]]; then
    _lazado_err "No fields to update. Use --title, --description, or --assigned-to."
    return 1
  fi

  "${cmd[@]}"
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
    "Update Item" \
    "Open in Browser" \
    "Create Branch" \
    "Back to Menu" \
    | fzf --header="#$id  ─  Actions" \
          --prompt="Action > " \
          --height=~40% --reverse --border $LAZADO_FZF_OPTS)

  case "$action" in
    "Update Status")  _lazado_nav_update_state "$id" "$item_type" ;;
    "Update Item")    _lazado_nav_update_item_fields "$id" ;;
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

# Field-level update menu for a work item (called when ID is already known)
_lazado_nav_update_item_fields() {
  local id="$1"
  local org
  org=$(_lazado_org) || return 1

  echo ""
  ado-wi-detail "$id"
  echo ""

  local field
  field=$(printf '%s\n' \
    "Title" \
    "Description" \
    "Assign to me" \
    "Unassign" \
    "Back" \
    | fzf --header="#$id  ─  Update field" \
          --prompt="Field > " \
          --height=~40% --reverse --border $LAZADO_FZF_OPTS)

  case "$field" in
    "Title")
      local new_title
      read -rp "New title: " new_title
      if [[ -z "$new_title" ]]; then
        _lazado_err "Aborted: title cannot be empty."
      else
        az boards work-item update --id "$id" --title "$new_title" \
          --org "$org" --output none 2>/dev/null
        echo "Updated title for #$id."
      fi
      _lazado_nav_update_item_fields "$id"
      ;;
    "Description")
      local new_desc
      read -rp "New description: " new_desc
      az boards work-item update --id "$id" --description "$new_desc" \
        --org "$org" --output none 2>/dev/null
      echo "Updated description for #$id."
      _lazado_nav_update_item_fields "$id"
      ;;
    "Assign to me")
      az boards work-item update --id "$id" --assigned-to "$(_lazado_me)" \
        --org "$org" --output none 2>/dev/null
      echo "Assigned #$id to you."
      _lazado_nav_update_item_fields "$id"
      ;;
    "Unassign")
      az boards work-item update --id "$id" --assigned-to "" \
        --org "$org" --output none 2>/dev/null
      echo "Unassigned #$id."
      _lazado_nav_update_item_fields "$id"
      ;;
    "Back"|"") return 0 ;;
  esac
}

# Standalone update — select from your items then pick fields to update
_lazado_nav_update_item() {
  local item_type="$1"
  local label="$item_type"
  [[ "$item_type" == "Task" ]] && label="Work Items"
  [[ "$item_type" == "User Story" ]] && label="User Stories"

  local org project
  org=$(_lazado_org) || return 1
  project=$(_lazado_project) || return 1

  _lazado_info "Fetching $label (New/Active)..."
  local state_clause
  state_clause=$(_lazado_state_filter "New/Active" "$item_type")

  local items
  items=$(az boards query \
    --wiql "SELECT [System.Id], [System.Title], [System.State], [System.WorkItemType], [System.CreatedDate] FROM WorkItems WHERE [System.AssignedTo] = @Me AND [System.WorkItemType] = '$item_type' $state_clause ORDER BY [System.CreatedDate] DESC" \
    --org "$org" \
    --project "$project" \
    --output json 2>/dev/null)

  if [[ -z "$items" || "$items" == "[]" ]]; then
    echo "No $label found."
    read -rp "Press Enter to go back..." _
    ado
    return 0
  fi

  local selected
  selected=$(echo "$items" | \
    jq -r '.[] | "\(.id)\t\(.fields["System.State"])\t\(.fields["System.Title"])"' | \
    column -t -s$'\t' | \
    fzf --header="$label  ─  Select item to update, ESC to go back" \
        --prompt="$label > " \
        --height=~60% --reverse --border $LAZADO_FZF_OPTS)

  [[ -z "$selected" ]] && ado && return 0

  local id
  id=$(echo "$selected" | awk '{print $1}')
  _lazado_nav_update_item_fields "$id"
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
