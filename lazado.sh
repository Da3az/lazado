#!/usr/bin/env bash
# =============================================================================
# lazado — Interactive Azure DevOps CLI
# https://github.com/da3az/lazado
#
# Source this file from your shell profile:
#   export LAZADO_DIR="$HOME/.lazado"
#   [ -s "$LAZADO_DIR/lazado.sh" ] && \. "$LAZADO_DIR/lazado.sh"
# =============================================================================

# Bash-only guard
if [[ -z "${BASH_VERSION:-}" ]]; then
  echo "lazado requires bash. Current shell is not supported." >&2
  return 1 2>/dev/null || exit 1
fi

# Resolve install directory
if [[ -z "${LAZADO_DIR:-}" ]]; then
  if [[ -n "${BASH_SOURCE[0]:-}" ]]; then
    LAZADO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  else
    LAZADO_DIR="$HOME/.lazado"
  fi
fi
export LAZADO_DIR

LAZADO_VERSION="0.1.0"
export LAZADO_VERSION

# Source modules in order
source "$LAZADO_DIR/lib/config.sh"
source "$LAZADO_DIR/lib/utils.sh"
source "$LAZADO_DIR/lib/states.sh"
source "$LAZADO_DIR/lib/work-items.sh"
source "$LAZADO_DIR/lib/pull-requests.sh"
source "$LAZADO_DIR/lib/pipelines.sh"
source "$LAZADO_DIR/lib/repos.sh"
source "$LAZADO_DIR/lib/navigator.sh"
source "$LAZADO_DIR/lib/init.sh"
