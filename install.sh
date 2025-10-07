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
        # Try with sudo if regular move fails
        if ! sudo mv "$BINARY_NAME" "$INSTALL_DIR/"; then
            print_error "Failed to install binary"
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
    # Detect shell config file
    if [ -n "$BASH_VERSION" ]; then
        SHELL_CONFIG="$HOME/.bashrc"
    elif [ -n "$ZSH_VERSION" ]; then
        SHELL_CONFIG="$HOME/.zshrc"
    else
        case "$SHELL" in
            */bash)
                SHELL_CONFIG="$HOME/.bashrc"
                ;;
            */zsh)
                SHELL_CONFIG="$HOME/.zshrc"
                ;;
            */fish)
                SHELL_CONFIG="$HOME/.config/fish/config.fish"
                ;;
            *)
                SHELL_CONFIG="$HOME/.profile"
                ;;
        esac
    fi

    # Check if ~/.local/bin is already in PATH config
    if ! grep -q ".local/bin" "$SHELL_CONFIG" 2>/dev/null; then
        print_info "Adding ~/.local/bin to PATH in $SHELL_CONFIG"
        echo "" >> "$SHELL_CONFIG"
        echo "# User binaries" >> "$SHELL_CONFIG"
        echo 'export PATH=$PATH:$HOME/.local/bin' >> "$SHELL_CONFIG"
    fi

    # Export for current session
    export PATH=$PATH:$HOME/.local/bin
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
        sudo rm -rf /usr/local/go
    fi

    # Extract and install
    print_info "Extracting Go..."
    sudo tar -C /usr/local -xzf "$GO_TARBALL"

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

    # Detect shell config file
    if [ -n "$BASH_VERSION" ]; then
        SHELL_CONFIG="$HOME/.bashrc"
    elif [ -n "$ZSH_VERSION" ]; then
        SHELL_CONFIG="$HOME/.zshrc"
    else
        case "$SHELL" in
            */bash)
                SHELL_CONFIG="$HOME/.bashrc"
                ;;
            */zsh)
                SHELL_CONFIG="$HOME/.zshrc"
                ;;
            */fish)
                SHELL_CONFIG="$HOME/.config/fish/config.fish"
                ;;
            *)
                SHELL_CONFIG="$HOME/.profile"
                ;;
        esac
    fi

    # Check if Go paths are already in config
    if ! grep -q "/usr/local/go/bin" "$SHELL_CONFIG" 2>/dev/null; then
        print_info "Adding Go to PATH in $SHELL_CONFIG"
        echo "" >> "$SHELL_CONFIG"
        echo "# Go environment" >> "$SHELL_CONFIG"
        echo 'export PATH=$PATH:/usr/local/go/bin' >> "$SHELL_CONFIG"
        echo 'export PATH=$PATH:$HOME/go/bin' >> "$SHELL_CONFIG"
    fi

    # Export for current session
    export PATH=$PATH:/usr/local/go/bin
    export PATH=$PATH:$HOME/go/bin
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

    print_info "Detected shell: $CURRENT_SHELL"
}

# Setup shell completion
setup_completion() {
    print_info "Setting up shell completion for $CURRENT_SHELL..."

    # Verify gzcli is in PATH
    if ! command -v gzcli &> /dev/null; then
        print_warning "gzcli not found in PATH, skipping completion setup"
        return 1
    fi

    case "$CURRENT_SHELL" in
        bash)
            # Create bash completion directory if it doesn't exist
            COMPLETION_DIR="$HOME/.bash_completion.d"
            mkdir -p "$COMPLETION_DIR"

            # Generate completion script
            gzcli completion bash > "$COMPLETION_DIR/gzcli"

            # Add to bashrc if not already there
            if ! grep -q "bash_completion.d/gzcli" "$SHELL_CONFIG" 2>/dev/null; then
                echo "" >> "$SHELL_CONFIG"
                echo "# gzcli completion" >> "$SHELL_CONFIG"
                echo "source $COMPLETION_DIR/gzcli" >> "$SHELL_CONFIG"
            fi

            print_success "Bash completion installed!"
            print_info "Run 'source ~/.bashrc' or restart your shell to enable completion"
            ;;

        zsh)
            # Create zsh completion directory if it doesn't exist
            COMPLETION_DIR="$HOME/.zsh/completion"
            mkdir -p "$COMPLETION_DIR"

            # Generate completion script
            gzcli completion zsh > "$COMPLETION_DIR/_gzcli"

            # Add to zshrc if not already there
            if ! grep -q "fpath.*zsh/completion" "$SHELL_CONFIG" 2>/dev/null; then
                echo "" >> "$SHELL_CONFIG"
                echo "# gzcli completion" >> "$SHELL_CONFIG"
                echo "fpath=($COMPLETION_DIR \$fpath)" >> "$SHELL_CONFIG"
                echo "autoload -Uz compinit && compinit" >> "$SHELL_CONFIG"
            fi

            print_success "Zsh completion installed!"
            print_info "Run 'source ~/.zshrc' or restart your shell to enable completion"
            ;;

        fish)
            # Create fish completion directory if it doesn't exist
            COMPLETION_DIR="$HOME/.config/fish/completions"
            mkdir -p "$COMPLETION_DIR"

            # Generate completion script
            gzcli completion fish > "$COMPLETION_DIR/gzcli.fish"

            print_success "Fish completion installed!"
            print_info "Restart your fish shell to enable completion"
            ;;

        *)
            print_warning "Shell completion not available for $CURRENT_SHELL"
            print_info "Supported shells: bash, zsh, fish"
            return 1
            ;;
    esac
}

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

    # Fall back to source installation if binary installation failed
    if [ "$BINARY_INSTALL_SUCCESS" = false ]; then
        echo ""
        print_warning "Binary installation failed or not available"
        print_info "Falling back to source installation..."
        echo ""

        # Check and install Go if needed
        if ! check_go_version; then
            print_warning "Go is required to install gzcli from source"
            read -p "Do you want to install the latest Go version? (y/n) " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                get_latest_go_version
                install_go
            else
                print_error "Go is required. Please install Go manually from https://go.dev/dl/"
                exit 1
            fi
        fi

        # Verify Go installation
        if ! command -v go &> /dev/null; then
            print_error "Go installation failed or not in PATH"
            print_info "Please restart your shell and run this script again"
            exit 1
        fi

        # Install gzcli from source
        install_from_source
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

    # Setup shell completion
    if detect_shell; then
        read -p "Do you want to setup shell completion for $CURRENT_SHELL? (y/n) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            setup_completion
        fi
    fi

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
