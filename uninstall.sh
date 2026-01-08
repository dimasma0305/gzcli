#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color


# Constants
BINARY_NAME="gzcli"
INSTALL_DIR_SYSTEM="/usr/local/bin"
INSTALL_DIR_USER="$HOME/.local/bin"
INSTALL_DIR_GO="$HOME/go/bin"

# Print functions
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Remove binary from a specific directory
remove_binary_from() {
    local dir="$1"
    local binary_path="$dir/$BINARY_NAME"

    if [ -f "$binary_path" ]; then
        print_info "Found $BINARY_NAME in $dir"
        if rm "$binary_path" 2>/dev/null; then
            print_success "  ✓ Removed $binary_path"
            return 0
        elif sudo rm "$binary_path" 2>/dev/null; then
            print_success "  ✓ Removed $binary_path (with sudo)"
            return 0
        else
            print_error "  ✗ Failed to remove $binary_path"
            return 1
        fi
    fi
    return 1
}

# Remove gzcli binary
remove_binary() {
    print_info "Searching for $BINARY_NAME binary..."

    local found=false

    # Check system directory
    if remove_binary_from "$INSTALL_DIR_SYSTEM"; then
        found=true
    fi

    # Check user directory
    if remove_binary_from "$INSTALL_DIR_USER"; then
        found=true
    fi

    # Check go bin directory
    if remove_binary_from "$INSTALL_DIR_GO"; then
        found=true
    fi

    # Check if binary is in PATH but not in standard locations
    if command -v $BINARY_NAME &> /dev/null; then
        local binary_location=$(which $BINARY_NAME)
        if [ -f "$binary_location" ]; then
            print_info "Found $BINARY_NAME at non-standard location: $binary_location"
            print_info "Removing it automatically..."
            if remove_binary_from "$(dirname "$binary_location")"; then
                found=true
            fi
        fi
    fi

    if [ "$found" = false ]; then
        print_warning "No $BINARY_NAME binary found in standard locations"
    fi
}

# Remove bash completion
remove_bash_completion() {
    print_info "Checking for bash completion..."

    local completion_file="$HOME/.bash_completion.d/gzcli"
    local bashrc="$HOME/.bashrc"
    local removed=false

    # Remove completion file
    if [ -f "$completion_file" ]; then
        if rm "$completion_file" 2>/dev/null; then
            print_success "  ✓ Removed bash completion file"
            removed=true

            # Remove directory if empty
            if [ -d "$HOME/.bash_completion.d" ] && [ -z "$(ls -A "$HOME/.bash_completion.d")" ]; then
                rmdir "$HOME/.bash_completion.d" 2>/dev/null
                print_info "  ✓ Removed empty completion directory"
            fi
        fi
    fi

    # Remove from bashrc
    if [ -f "$bashrc" ] && grep -q "gzcli completion" "$bashrc" 2>/dev/null; then
        # Create backup
        cp "$bashrc" "${bashrc}.backup.$(date +%Y%m%d_%H%M%S)"

        # Remove gzcli completion section
        sed -i '/# gzcli completion/,+1d' "$bashrc" 2>/dev/null || {
            print_warning "  Could not automatically remove from $bashrc"
            print_info "  Please manually remove the gzcli completion section"
        }
        print_success "  ✓ Removed bash completion from $bashrc"
        print_info "  Backup created: ${bashrc}.backup.*"
        removed=true
    fi

    if [ "$removed" = false ]; then
        print_info "  No bash completion found"
    fi
}

# Remove zsh completion
remove_zsh_completion() {
    print_info "Checking for zsh completion..."

    local completion_file="$HOME/.zsh/completion/_gzcli"
    local zshrc="$HOME/.zshrc"
    local removed=false

    # Remove completion file
    if [ -f "$completion_file" ]; then
        if rm "$completion_file" 2>/dev/null; then
            print_success "  ✓ Removed zsh completion file"
            removed=true

            # Remove directory if empty
            if [ -d "$HOME/.zsh/completion" ] && [ -z "$(ls -A "$HOME/.zsh/completion")" ]; then
                rmdir "$HOME/.zsh/completion" 2>/dev/null
                print_info "  ✓ Removed empty completion directory"
                if [ -d "$HOME/.zsh" ] && [ -z "$(ls -A "$HOME/.zsh")" ]; then
                    rmdir "$HOME/.zsh" 2>/dev/null
                fi
            fi
        fi
    fi

    # Remove from zshrc
    if [ -f "$zshrc" ] && grep -q "gzcli completion" "$zshrc" 2>/dev/null; then
        # Create backup
        cp "$zshrc" "${zshrc}.backup.$(date +%Y%m%d_%H%M%S)"

        # Remove gzcli completion section (including fpath and autoload lines)
        sed -i '/# gzcli completion/,+2d' "$zshrc" 2>/dev/null || {
            print_warning "  Could not automatically remove from $zshrc"
            print_info "  Please manually remove the gzcli completion section"
        }
        print_success "  ✓ Removed zsh completion from $zshrc"
        print_info "  Backup created: ${zshrc}.backup.*"
        removed=true
    fi

    if [ "$removed" = false ]; then
        print_info "  No zsh completion found"
    fi
}

# Remove fish completion
remove_fish_completion() {
    print_info "Checking for fish completion..."

    local completion_file="$HOME/.config/fish/completions/gzcli.fish"

    if [ -f "$completion_file" ]; then
        if rm "$completion_file" 2>/dev/null; then
            print_success "  ✓ Removed fish completion file"

            # Remove directory if empty
            if [ -d "$HOME/.config/fish/completions" ] && [ -z "$(ls -A "$HOME/.config/fish/completions")" ]; then
                rmdir "$HOME/.config/fish/completions" 2>/dev/null
                print_info "  ✓ Removed empty completion directory"
            fi
        fi
    else
        print_info "  No fish completion found"
    fi
}

# Remove PowerShell completion
remove_powershell_completion() {
    print_info "Checking for PowerShell completion..."

    local completion_file="$HOME/.config/powershell/gzcli-completion.ps1"
    local pwsh_profile="$HOME/.config/powershell/Microsoft.PowerShell_profile.ps1"
    local removed=false

    # Remove completion file
    if [ -f "$completion_file" ]; then
        if rm "$completion_file" 2>/dev/null; then
            print_success "  ✓ Removed PowerShell completion file"
            removed=true
        fi
    fi

    # Remove from PowerShell profile
    if [ -f "$pwsh_profile" ] && grep -q "gzcli-completion.ps1" "$pwsh_profile" 2>/dev/null; then
        # Create backup
        cp "$pwsh_profile" "${pwsh_profile}.backup.$(date +%Y%m%d_%H%M%S)"

        # Remove gzcli completion section
        sed -i '/# gzcli completion/,+1d' "$pwsh_profile" 2>/dev/null || {
            print_warning "  Could not automatically remove from PowerShell profile"
            print_info "  Please manually remove the gzcli completion section"
        }
        print_success "  ✓ Removed PowerShell completion from profile"
        print_info "  Backup created: ${pwsh_profile}.backup.*"
        removed=true
    fi

    if [ "$removed" = false ]; then
        print_info "  No PowerShell completion found"
    fi
}

# Remove all shell completions
remove_all_completions() {
    echo ""
    print_info "Removing shell completions..."
    echo ""

    remove_bash_completion
    echo ""

    remove_zsh_completion
    echo ""

    remove_fish_completion
    echo ""

    remove_powershell_completion
}

# Main uninstallation flow
main() {
    echo ""
    echo "===================================="
    echo "  gzcli Uninstallation Script"
    echo "===================================="
    echo ""

    # No confirmation needed - proceed with uninstallation
    print_info "Removing gzcli and all its shell completions..."
    echo ""

    # Remove binary
    remove_binary

    # Remove completions automatically
    echo ""
    remove_all_completions

    echo ""
    print_success "Uninstallation complete!"
    echo ""

    # Check if binary is still accessible
    if command -v $BINARY_NAME &> /dev/null; then
        print_warning "$BINARY_NAME is still accessible in PATH"
        print_info "Location: $(which $BINARY_NAME)"
        print_info "You may need to manually remove it or restart your shell"
    else
        print_success "$BINARY_NAME has been successfully removed from your system"
    fi

    echo ""
    print_info "Note: Shell configuration backups were created with .backup.* extension"
    print_info "You may need to restart your shell or source your shell config file"
    echo ""
}

# Run main function
main
