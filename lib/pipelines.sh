#!/usr/bin/env bash
# =============================================================================
# lazado — pipelines.sh
# Pipeline commands
# =============================================================================

# List recent pipeline runs
ado-pipelines() {
  local org project
  org=$(_lazado_org) || return 1
  project=$(_lazado_project) || return 1
  az pipelines run list \
    --org "$org" \
    --project "$project" \
    --top 10 \
    --output table
}
