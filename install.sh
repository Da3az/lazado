#!/usr/bin/env bash
# =============================================================================
# lazado installer
# Usage: curl -o- https://raw.githubusercontent.com/da3az/lazado/main/install.sh | bash
# =============================================================================

set -euo pipefail

LAZADO_DIR="${LAZADO_DIR:-$HOME/.lazado}"
REPO_URL="https://github.com/da3az/lazado.git"

echo ""
echo "lazado — Interactive Azure DevOps CLI"
echo "======================================"
echo ""

# -- Check git ----------------------------------------------------------------
if ! command -v git &>/dev/null; then
  echo "Error: git is required to install lazado." >&2
  exit 1
fi

# -- Clone or update ----------------------------------------------------------
if [[ -d "$LAZADO_DIR" ]]; then
  echo "Updating existing installation at $LAZADO_DIR..."
  cd "$LAZADO_DIR" && git pull origin main
else
  echo "Installing to $LAZADO_DIR..."
  git clone "$REPO_URL" "$LAZADO_DIR"
fi

# -- Detect shell profile -----------------------------------------------------
PROFILE=""
if [[ -f "$HOME/.bashrc" ]]; then
  PROFILE="$HOME/.bashrc"
elif [[ -f "$HOME/.bash_profile" ]]; then
  PROFILE="$HOME/.bash_profile"
elif [[ -f "$HOME/.profile" ]]; then
  PROFILE="$HOME/.profile"
fi

# -- Add source line to profile -----------------------------------------------
SOURCE_LINE='export LAZADO_DIR="$HOME/.lazado"'
LOAD_LINE='[ -s "$LAZADO_DIR/lazado.sh" ] && \. "$LAZADO_DIR/lazado.sh"'

if [[ -n "$PROFILE" ]]; then
  if grep -q "lazado.sh" "$PROFILE" 2>/dev/null; then
    echo "Shell profile already configured ($PROFILE)."
  else
    echo "" >> "$PROFILE"
    echo "# lazado — Interactive Azure DevOps CLI" >> "$PROFILE"
    echo "$SOURCE_LINE" >> "$PROFILE"
    echo "$LOAD_LINE" >> "$PROFILE"
    echo "Added lazado to $PROFILE."
  fi
else
  echo "Could not detect shell profile. Add these lines manually:"
  echo ""
  echo "  $SOURCE_LINE"
  echo "  $LOAD_LINE"
fi

# -- Done ---------------------------------------------------------------------
echo ""
echo "Installation complete."
echo ""
echo "Next steps:"
echo "  1. Restart your terminal or run: source $PROFILE"
echo "  2. Run: ado-init"
echo ""
echo "Prerequisites (if not already installed):"
echo "  - Azure CLI:  curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash"
echo "  - DevOps ext: az extension add --name azure-devops"
echo "  - Login:      az login"
echo "  - jq:         sudo apt install jq  (or brew install jq)"
echo "  - fzf:        https://github.com/junegunn/fzf#installation"
echo ""
