#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Constants
GITHUB_REPO="dimasma0305/gzcli"
BINARY_NAME="gzcli"
INSTALL_DIR_SYSTEM="/usr/local/bin"
INSTALL_DIR_USER="$HOME/.local/bin"

# Utility helpers
is_root() {
    [ "$(id -u)" -eq 0 ]
}

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

run_with_privileges() {
    if is_root; then
        "$@"
    elif command_exists sudo; then
        sudo "$@"
    else
        return 1
    fi
}

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

# Detect OS and architecture
detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        OS="Linux"
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        OS="Darwin"
    elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]]; then
        OS="Windows"
    else
        OS="unknown"
    fi

    ARCH=$(uname -m)
    case $ARCH in
        x86_64)
            ARCH="x86_64"
            GO_ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            GO_ARCH="arm64"
            ;;
        armv7l)
            ARCH="armv6"
            GO_ARCH="armv6l"
            ;;
        armv6l)
            ARCH="armv6"
            GO_ARCH="armv6l"
            ;;
        i386|i686)
            ARCH="i386"
            GO_ARCH="386"
            ;;
        *)
            print_warning "Unsupported architecture: $ARCH"
            ARCH="unknown"
            GO_ARCH="amd64"
            ;;
    esac

    print_info "Detected OS: $OS, Architecture: $ARCH"
}

# Get latest release version from GitHub
get_latest_release() {
    print_info "Fetching latest release version..."

    # Try to get latest release from GitHub API
    LATEST_VERSION=$(curl -sf "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"tag_name": "([^"]+)".*/\1/')

    if [ -z "$LATEST_VERSION" ]; then
        print_warning "Could not fetch latest release version from GitHub API"
        return 1
    fi

    print_info "Latest version: $LATEST_VERSION"
    return 0
}

# Download binary from GitHub releases
download_binary() {
    print_info "Attempting to download pre-built binary..."

    # Construct archive name based on OS and architecture
    ARCHIVE_NAME="${BINARY_NAME}_${LATEST_VERSION#v}_${OS}_${ARCH}"

    if [[ "$OS" == "Windows" ]]; then
        ARCHIVE_EXT="zip"
    else
        ARCHIVE_EXT="tar.gz"
    fi

    ARCHIVE_FILE="${ARCHIVE_NAME}.${ARCHIVE_EXT}"
    DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download/${LATEST_VERSION}/${ARCHIVE_FILE}"

    print_info "Downloading from: $DOWNLOAD_URL"

    # Create temporary directory
    TMP_DIR=$(mktemp -d)
    cd "$TMP_DIR"

    # Download archive
    if ! curl -sfL "$DOWNLOAD_URL" -o "$ARCHIVE_FILE"; then
        print_warning "Failed to download binary archive"
        cd - > /dev/null
        rm -rf "$TMP_DIR"
        return 1
    fi

    # Download checksums
    CHECKSUM_URL="https://github.com/${GITHUB_REPO}/releases/download/${LATEST_VERSION}/checksums.txt"
    if ! curl -sfL "$CHECKSUM_URL" -o checksums.txt; then
        print_warning "Failed to download checksums file"
        cd - > /dev/null
        rm -rf "$TMP_DIR"
        return 1
    fi

    print_success "Binary archive downloaded successfully"
    return 0
}

# Verify checksum
verify_checksum() {
    print_info "Verifying checksum..."

    # Get expected checksum from checksums.txt
    EXPECTED_CHECKSUM=$(grep "$ARCHIVE_FILE" checksums.txt | awk '{print $1}')

    if [ -z "$EXPECTED_CHECKSUM" ]; then
        print_warning "Could not find checksum for $ARCHIVE_FILE"
        return 1
    fi

    # Calculate actual checksum
    if command -v sha256sum > /dev/null; then
        ACTUAL_CHECKSUM=$(sha256sum "$ARCHIVE_FILE" | awk '{print $1}')
    elif command -v shasum > /dev/null; then
        ACTUAL_CHECKSUM=$(shasum -a 256 "$ARCHIVE_FILE" | awk '{print $1}')
    else
        print_warning "No checksum utility found (sha256sum or shasum)"
        return 1
    fi

    # Compare checksums
    if [ "$EXPECTED_CHECKSUM" != "$ACTUAL_CHECKSUM" ]; then
        print_error "Checksum verification failed!"
        print_error "Expected: $EXPECTED_CHECKSUM"
        print_error "Got:      $ACTUAL_CHECKSUM"
        return 1
    fi

    print_success "Checksum verified successfully"
    return 0
}

# Extract and install binary
install_binary() {
    print_info "Extracting binary..."

    # Extract archive
    if [[ "$ARCHIVE_EXT" == "tar.gz" ]]; then
        if ! tar -xzf "$ARCHIVE_FILE"; then
            print_error "Failed to extract archive"
            return 1
        fi
    elif [[ "$ARCHIVE_EXT" == "zip" ]]; then
        if ! unzip -q "$ARCHIVE_FILE"; then
            print_error "Failed to extract archive"
            return 1
        fi
    fi

    # Find the binary
    if [ ! -f "$BINARY_NAME" ]; then
        print_error "Binary not found in archive"
        return 1
    fi

    # Make binary executable
    chmod +x "$BINARY_NAME"

    # Determine install directory
    if [ -w "$INSTALL_DIR_SYSTEM" ]; then
        INSTALL_DIR="$INSTALL_DIR_SYSTEM"
    elif [ -d "$INSTALL_DIR_USER" ] || mkdir -p "$INSTALL_DIR_USER" 2>/dev/null; then
        INSTALL_DIR="$INSTALL_DIR_USER"
        # Add to PATH if not already there
        add_user_bin_to_path
    else
        print_error "No suitable installation directory found"
        return 1
    fi

    # Move binary to install directory
    print_info "Installing to $INSTALL_DIR..."
    if ! mv "$BINARY_NAME" "$INSTALL_DIR/"; then
        if ! run_with_privileges mv "$BINARY_NAME" "$INSTALL_DIR/"; then
            print_error "Failed to install binary (insufficient permissions)"
            print_info "Try running this script as root or install manually by moving $BINARY_NAME to $INSTALL_DIR"
            return 1
        fi
    fi

    # Cleanup
    cd - > /dev/null
    rm -rf "$TMP_DIR"

    print_success "Binary installed successfully to $INSTALL_DIR/$BINARY_NAME"
    return 0
}

# Add ~/.local/bin to PATH
add_user_bin_to_path() {
    print_info "Adding ~/.local/bin to PATH for all available shells..."

    # Export for current session
    export PATH=$PATH:$HOME/.local/bin

    # Add to bash if available
    if command -v bash &> /dev/null; then
        SHELL_CONFIG="$HOME/.bashrc"
        if ! grep -q ".local/bin" "$SHELL_CONFIG" 2>/dev/null; then
            print_info "  Adding to $SHELL_CONFIG"
            touch "$SHELL_CONFIG" 2>/dev/null || true
            echo "" >> "$SHELL_CONFIG"
            echo "# User binaries" >> "$SHELL_CONFIG"
            echo 'export PATH=$PATH:$HOME/.local/bin' >> "$SHELL_CONFIG"
        fi
    fi

    # Add to zsh if available
    if command -v zsh &> /dev/null; then
        SHELL_CONFIG="$HOME/.zshrc"
        if ! grep -q ".local/bin" "$SHELL_CONFIG" 2>/dev/null; then
            print_info "  Adding to $SHELL_CONFIG"
            touch "$SHELL_CONFIG" 2>/dev/null || true
            echo "" >> "$SHELL_CONFIG"
            echo "# User binaries" >> "$SHELL_CONFIG"
            echo 'export PATH=$PATH:$HOME/.local/bin' >> "$SHELL_CONFIG"
        fi
    fi

    # Add to fish if available
    if command -v fish &> /dev/null; then
        SHELL_CONFIG="$HOME/.config/fish/config.fish"
        if ! grep -q ".local/bin" "$SHELL_CONFIG" 2>/dev/null; then
            print_info "  Adding to $SHELL_CONFIG"
            mkdir -p "$HOME/.config/fish" 2>/dev/null || true
            touch "$SHELL_CONFIG" 2>/dev/null || true
            echo "" >> "$SHELL_CONFIG"
            echo "# User binaries" >> "$SHELL_CONFIG"
            echo 'set -gx PATH $PATH $HOME/.local/bin' >> "$SHELL_CONFIG"
        fi
    fi

    # Add to profile as fallback
    SHELL_CONFIG="$HOME/.profile"
    if ! grep -q ".local/bin" "$SHELL_CONFIG" 2>/dev/null; then
        print_info "  Adding to $SHELL_CONFIG"
        touch "$SHELL_CONFIG" 2>/dev/null || true
        echo "" >> "$SHELL_CONFIG"
        echo "# User binaries" >> "$SHELL_CONFIG"
        echo 'export PATH=$PATH:$HOME/.local/bin' >> "$SHELL_CONFIG"
    fi
}

# Check if Go is installed and get version
check_go_version() {
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
        print_info "Go $GO_VERSION is already installed"
        return 0
    else
        print_warning "Go is not installed"
        return 1
    fi
}

# Get latest Go version
get_latest_go_version() {
    print_info "Fetching latest Go version..."
    LATEST_GO_VERSION=$(curl -s https://go.dev/VERSION?m=text | head -n 1)
    if [ -z "$LATEST_GO_VERSION" ]; then
        print_error "Failed to fetch latest Go version"
        exit 1
    fi
    print_info "Latest Go version: $LATEST_GO_VERSION"
}

# Install Go
install_go() {
    print_info "Installing Go $LATEST_GO_VERSION..."

    # Map OS names for Go downloads
    case "$OS" in
        Linux)
            GO_OS="linux"
            ;;
        Darwin)
            GO_OS="darwin"
            ;;
        Windows)
            GO_OS="windows"
            ;;
        *)
            print_error "Unsupported OS for Go installation: $OS"
            return 1
            ;;
    esac

    # Download Go
    GO_TARBALL="${LATEST_GO_VERSION}.${GO_OS}-${GO_ARCH}.tar.gz"
    DOWNLOAD_URL="https://go.dev/dl/${GO_TARBALL}"

    print_info "Downloading from $DOWNLOAD_URL..."

    TMP_DIR=$(mktemp -d)
    cd "$TMP_DIR"

    if ! curl -LO "$DOWNLOAD_URL"; then
        print_error "Failed to download Go"
        rm -rf "$TMP_DIR"
        exit 1
    fi

    # Remove old Go installation if exists
    if [ -d "/usr/local/go" ]; then
        print_info "Removing old Go installation..."
        if ! run_with_privileges rm -rf /usr/local/go; then
            print_error "Failed to remove old Go installation (insufficient permissions)"
            cd - > /dev/null
            rm -rf "$TMP_DIR"
            exit 1
        fi
    fi

    # Extract and install
    print_info "Extracting Go..."
    if ! run_with_privileges tar -C /usr/local -xzf "$GO_TARBALL"; then
        print_error "Failed to extract Go archive (insufficient permissions)"
        cd - > /dev/null
        rm -rf "$TMP_DIR"
        exit 1
    fi

    # Cleanup
    cd - > /dev/null
    rm -rf "$TMP_DIR"

    # Add Go to PATH
    add_go_to_path

    print_success "Go installed successfully!"
}

# Add Go to PATH
add_go_to_path() {
    GO_PATH="/usr/local/go/bin"
    GOPATH_BIN="$HOME/go/bin"

    print_info "Adding Go to PATH for all available shells..."

    # Export for current session
    export PATH=$PATH:/usr/local/go/bin
    export PATH=$PATH:$HOME/go/bin

    # Add to bash if available
    if command -v bash &> /dev/null; then
        SHELL_CONFIG="$HOME/.bashrc"
        if ! grep -q "/usr/local/go/bin" "$SHELL_CONFIG" 2>/dev/null; then
            print_info "  Adding to $SHELL_CONFIG"
            touch "$SHELL_CONFIG" 2>/dev/null || true
            echo "" >> "$SHELL_CONFIG"
            echo "# Go environment" >> "$SHELL_CONFIG"
            echo 'export PATH=$PATH:/usr/local/go/bin' >> "$SHELL_CONFIG"
            echo 'export PATH=$PATH:$HOME/go/bin' >> "$SHELL_CONFIG"
        fi
    fi

    # Add to zsh if available
    if command -v zsh &> /dev/null; then
        SHELL_CONFIG="$HOME/.zshrc"
        if ! grep -q "/usr/local/go/bin" "$SHELL_CONFIG" 2>/dev/null; then
            print_info "  Adding to $SHELL_CONFIG"
            touch "$SHELL_CONFIG" 2>/dev/null || true
            echo "" >> "$SHELL_CONFIG"
            echo "# Go environment" >> "$SHELL_CONFIG"
            echo 'export PATH=$PATH:/usr/local/go/bin' >> "$SHELL_CONFIG"
            echo 'export PATH=$PATH:$HOME/go/bin' >> "$SHELL_CONFIG"
        fi
    fi

    # Add to fish if available
    if command -v fish &> /dev/null; then
        SHELL_CONFIG="$HOME/.config/fish/config.fish"
        if ! grep -q "/usr/local/go/bin" "$SHELL_CONFIG" 2>/dev/null; then
            print_info "  Adding to $SHELL_CONFIG"
            mkdir -p "$HOME/.config/fish" 2>/dev/null || true
            touch "$SHELL_CONFIG" 2>/dev/null || true
            echo "" >> "$SHELL_CONFIG"
            echo "# Go environment" >> "$SHELL_CONFIG"
            echo 'set -gx PATH $PATH /usr/local/go/bin' >> "$SHELL_CONFIG"
            echo 'set -gx PATH $PATH $HOME/go/bin' >> "$SHELL_CONFIG"
        fi
    fi

    # Add to profile as fallback
    SHELL_CONFIG="$HOME/.profile"
    if ! grep -q "/usr/local/go/bin" "$SHELL_CONFIG" 2>/dev/null; then
        print_info "  Adding to $SHELL_CONFIG"
        touch "$SHELL_CONFIG" 2>/dev/null || true
        echo "" >> "$SHELL_CONFIG"
        echo "# Go environment" >> "$SHELL_CONFIG"
        echo 'export PATH=$PATH:/usr/local/go/bin' >> "$SHELL_CONFIG"
        echo 'export PATH=$PATH:$HOME/go/bin' >> "$SHELL_CONFIG"
    fi
}

# Install gzcli from source
install_from_source() {
    print_info "Installing gzcli from source..."

    if ! go install github.com/${GITHUB_REPO}@latest; then
        print_error "Failed to install gzcli from source"
        exit 1
    fi

    print_success "gzcli installed successfully from source!"
}

# Detect shell
detect_shell() {
    if [ -n "$BASH_VERSION" ]; then
        CURRENT_SHELL="bash"
        SHELL_CONFIG="$HOME/.bashrc"
    elif [ -n "$ZSH_VERSION" ]; then
        CURRENT_SHELL="zsh"
        SHELL_CONFIG="$HOME/.zshrc"
    else
        case "$SHELL" in
            */bash)
                CURRENT_SHELL="bash"
                SHELL_CONFIG="$HOME/.bashrc"
                ;;
            */zsh)
                CURRENT_SHELL="zsh"
                SHELL_CONFIG="$HOME/.zshrc"
                ;;
            */fish)
                CURRENT_SHELL="fish"
                SHELL_CONFIG="$HOME/.config/fish/config.fish"
                ;;
            *)
                CURRENT_SHELL="unknown"
                print_warning "Unknown shell: $SHELL"
                return 1
                ;;
        esac
    fi

    print_info "Detected current shell: $CURRENT_SHELL"
}

# Detect all available shells on the system
detect_available_shells() {
    AVAILABLE_SHELLS=()

    print_info "Detecting available shells..."

    # Check for bash
    if command -v bash &> /dev/null; then
        AVAILABLE_SHELLS+=("bash")
        print_info "  ✓ bash detected"
    fi

    # Check for zsh
    if command -v zsh &> /dev/null; then
        AVAILABLE_SHELLS+=("zsh")
        print_info "  ✓ zsh detected"
    fi

    # Check for fish
    if command -v fish &> /dev/null; then
        AVAILABLE_SHELLS+=("fish")
        print_info "  ✓ fish detected"
    fi

    # Check for PowerShell (pwsh)
    if command -v pwsh &> /dev/null; then
        AVAILABLE_SHELLS+=("powershell")
        print_info "  ✓ powershell detected"
    fi

    if [ ${#AVAILABLE_SHELLS[@]} -eq 0 ]; then
        print_warning "No supported shells detected"
        return 1
    fi

    print_info "Found ${#AVAILABLE_SHELLS[@]} supported shell(s)"
    return 0
}

# Setup shell completion for a specific shell
setup_completion_for_shell() {
    local shell_name="$1"

    print_info "Setting up shell completion for $shell_name..."

    # Verify gzcli is in PATH
    if ! command -v gzcli &> /dev/null; then
        print_error "  ✗ gzcli not found in PATH, skipping $shell_name completion"
        return 1
    fi

    case "$shell_name" in
        bash)
            # Create bash completion directory if it doesn't exist
            COMPLETION_DIR="$HOME/.bash_completion.d"
            mkdir -p "$COMPLETION_DIR" 2>/dev/null

            # Generate completion script
            print_info "  Generating bash completion script..."
            if gzcli completion bash > "$COMPLETION_DIR/gzcli" 2>&1; then
                # Add to bashrc if not already there
                SHELL_CONFIG="$HOME/.bashrc"
                # Create bashrc if it doesn't exist
                touch "$SHELL_CONFIG" 2>/dev/null || true

                print_info "  Updating $SHELL_CONFIG..."
                if ! grep -q "bash_completion.d/gzcli" "$SHELL_CONFIG" 2>/dev/null; then
                    {
                        echo ""
                        echo "# gzcli completion"
                        echo "source $COMPLETION_DIR/gzcli"
                    } >> "$SHELL_CONFIG" 2>/dev/null || {
                        print_warning "  Could not update $SHELL_CONFIG automatically"
                    }
                fi
                print_success "  ✓ Bash completion installed!"
                print_info "    Run 'source ~/.bashrc' or restart bash to enable completion"
                return 0
            else
                print_error "  ✗ Failed to generate bash completion"
                return 1
            fi
            ;;

        zsh)
            # Create zsh completion directory if it doesn't exist
            COMPLETION_DIR="$HOME/.zsh/completion"
            mkdir -p "$COMPLETION_DIR" 2>/dev/null

            # Generate completion script
            print_info "  Generating zsh completion script..."
            if gzcli completion zsh > "$COMPLETION_DIR/_gzcli" 2>&1; then
                # Add to zshrc if not already there
                SHELL_CONFIG="$HOME/.zshrc"
                # Create zshrc if it doesn't exist
                touch "$SHELL_CONFIG" 2>/dev/null || true

                print_info "  Updating $SHELL_CONFIG..."
                if ! grep -q "fpath.*zsh/completion" "$SHELL_CONFIG" 2>/dev/null; then
                    {
                        echo ""
                        echo "# gzcli completion"
                        echo "fpath=($COMPLETION_DIR \$fpath)"
                        echo "autoload -Uz compinit && compinit"
                    } >> "$SHELL_CONFIG" 2>/dev/null || {
                        print_warning "  Could not update $SHELL_CONFIG automatically"
                    }
                fi
                print_success "  ✓ Zsh completion installed!"
                print_info "    Run 'source ~/.zshrc' or restart zsh to enable completion"
                return 0
            else
                print_error "  ✗ Failed to generate zsh completion"
                return 1
            fi
            ;;

        fish)
            # Create fish completion directory if it doesn't exist
            COMPLETION_DIR="$HOME/.config/fish/completions"
            mkdir -p "$COMPLETION_DIR" 2>/dev/null

            # Generate completion script
            print_info "  Generating fish completion script..."
            if gzcli completion fish > "$COMPLETION_DIR/gzcli.fish" 2>&1; then
                print_success "  ✓ Fish completion installed!"
                print_info "    Restart fish or run 'source ~/.config/fish/config.fish' to enable completion"
                return 0
            else
                print_error "  ✗ Failed to generate fish completion"
                return 1
            fi
            ;;

        powershell)
            # Create PowerShell profile directory if it doesn't exist
            PWSH_PROFILE_DIR="$HOME/.config/powershell"
            mkdir -p "$PWSH_PROFILE_DIR" 2>/dev/null

            # Generate completion script
            print_info "  Generating PowerShell completion script..."
            if gzcli completion powershell > "$PWSH_PROFILE_DIR/gzcli-completion.ps1" 2>&1; then
                # Add to PowerShell profile if it exists
                PWSH_PROFILE="$PWSH_PROFILE_DIR/Microsoft.PowerShell_profile.ps1"
                if [ ! -f "$PWSH_PROFILE" ]; then
                    touch "$PWSH_PROFILE" 2>/dev/null || true
                fi

                print_info "  Updating PowerShell profile..."
                if ! grep -q "gzcli-completion.ps1" "$PWSH_PROFILE" 2>/dev/null; then
                    {
                        echo ""
                        echo "# gzcli completion"
                        echo ". $PWSH_PROFILE_DIR/gzcli-completion.ps1"
                    } >> "$PWSH_PROFILE" 2>/dev/null || {
                        print_warning "  Could not update PowerShell profile automatically"
                    }
                fi
                print_success "  ✓ PowerShell completion installed!"
                print_info "    Restart PowerShell to enable completion"
                return 0
            else
                print_error "  ✗ Failed to generate PowerShell completion"
                return 1
            fi
            ;;

        *)
            print_warning "Shell completion not available for $shell_name"
            print_info "Supported shells: bash, zsh, fish, powershell"
            return 1
            ;;
    esac

    return 0
}

# Setup shell completion for all available shells
setup_all_completions() {
    if ! detect_available_shells; then
        print_warning "No supported shells detected, skipping completion setup"
        return 0
    fi

    echo ""
    print_info "Installing shell completions for all available shells..."
    echo ""

    local success_count=0
    local fail_count=0
    local total_shells=${#AVAILABLE_SHELLS[@]}

    print_info "Found $total_shells shell(s) to configure: ${AVAILABLE_SHELLS[*]}"
    echo ""

    for shell in "${AVAILABLE_SHELLS[@]}"; do
        # Always attempt installation, continue even if one fails
        if setup_completion_for_shell "$shell"; then
            success_count=$((success_count + 1))
        else
            fail_count=$((fail_count + 1))
            print_warning "  (Continuing with next shell...)"
        fi
        echo ""
    done

    # Summary
    echo ""
    print_info "═══════════════════════════════════════"
    if [ $success_count -gt 0 ]; then
        print_success "✓ Successfully installed completions for $success_count of $total_shells shell(s)"
    fi

    if [ $fail_count -gt 0 ]; then
        print_warning "✗ Failed to install completions for $fail_count of $total_shells shell(s)"
    fi

    if [ $success_count -eq $total_shells ]; then
        print_success "All shell completions installed successfully!"
    fi
    print_info "═══════════════════════════════════════"

    return 0
}

# Print usage/help
print_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --install-go    Also install Go and build from source if binary installation fails"
    echo "  --with-go       Alias for --install-go"
    echo "  -h, --help      Show this help message"
    echo ""
    echo "By default, only the pre-built binary will be installed."
    echo "Use --install-go to enable Go installation and source build fallback."
}

# Parse command line arguments
INSTALL_GO=false
while [[ $# -gt 0 ]]; do
    case $1 in
        --install-go|--with-go)
            INSTALL_GO=true
            shift
            ;;
        -h|--help)
            print_usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            echo ""
            print_usage
            exit 1
            ;;
    esac
done

# Main installation flow
main() {
    echo ""
    echo "=================================="
    echo "  gzcli Installation Script"
    echo "=================================="
    echo ""

    # Detect OS and architecture
    detect_os

    # Try binary installation first
    BINARY_INSTALL_SUCCESS=false

    if [ "$OS" != "unknown" ] && [ "$ARCH" != "unknown" ]; then
        if get_latest_release; then
            if download_binary; then
                if verify_checksum; then
                    if install_binary; then
                        BINARY_INSTALL_SUCCESS=true
                    fi
                fi
            fi
        fi
    fi

    # Fall back to source installation only if requested and binary installation failed
    if [ "$BINARY_INSTALL_SUCCESS" = false ]; then
        if [ "$INSTALL_GO" = true ]; then
            echo ""
            print_warning "Binary installation failed or not available"
            print_info "Falling back to source installation (--install-go flag was set)..."
            echo ""

            # Check and install Go if needed
            if ! check_go_version; then
                print_warning "Go is required to install gzcli from source"
                print_info "Installing latest Go version automatically..."
                get_latest_go_version
                install_go
            fi

            # Verify Go installation
            if ! command -v go &> /dev/null; then
                print_error "Go installation failed or not in PATH"
                print_info "Please restart your shell and run this script again"
                exit 1
            fi

            # Install gzcli from source
            install_from_source
        else
            echo ""
            print_error "Binary installation failed or not available"
            print_info "To install from source, run with --install-go flag:"
            print_info "  $0 --install-go"
            echo ""
            exit 1
        fi
    fi

    # Verify gzcli installation
    if ! command -v gzcli &> /dev/null; then
        print_warning "gzcli is installed but not in PATH"
        if [ -f "$HOME/go/bin/gzcli" ]; then
            print_info "Found gzcli in $HOME/go/bin"
            print_info "Make sure $HOME/go/bin is in your PATH"
            print_info "Add this to your shell config: export PATH=\$PATH:\$HOME/go/bin"
        elif [ -f "$INSTALL_DIR_USER/gzcli" ]; then
            print_info "Found gzcli in $INSTALL_DIR_USER"
            print_info "Make sure $INSTALL_DIR_USER is in your PATH"
            print_info "Restart your shell or run: export PATH=\$PATH:$INSTALL_DIR_USER"
        fi
    fi

    # Setup shell completion for all available shells automatically
    echo ""
    print_info "Setting up shell completions for all available shells..."
    setup_all_completions

    echo ""
    print_success "Installation complete!"
    echo ""

    if command -v gzcli &> /dev/null; then
        print_info "You can now use gzcli by running: gzcli --help"
        print_info "Version: $(gzcli --version 2>/dev/null || echo 'unknown')"
    else
        print_info "After restarting your shell, you can use gzcli by running: gzcli --help"
        print_info "Or source your shell config: source $SHELL_CONFIG"
    fi
    echo ""
}

# Run main function
main
