#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Detect OS
detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        OS="linux"
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        OS="darwin"
    elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]]; then
        OS="windows"
    else
        OS="unknown"
    fi
    
    ARCH=$(uname -m)
    case $ARCH in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        armv7l)
            ARCH="armv6l"
            ;;
        *)
            print_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac
    
    print_info "Detected OS: $OS, Architecture: $ARCH"
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
    
    # Download Go
    GO_TARBALL="${LATEST_GO_VERSION}.${OS}-${ARCH}.tar.gz"
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
    
    # Add Go to PATH if not already there
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
        # Try to detect from SHELL variable
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

# Install gzcli
install_gzcli() {
    print_info "Installing gzcli..."
    
    if ! go install github.com/dimasma0305/gzcli@latest; then
        print_error "Failed to install gzcli"
        exit 1
    fi
    
    print_success "gzcli installed successfully!"
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
    
    # Check and install Go if needed
    if ! check_go_version; then
        print_warning "Go is required to install gzcli"
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
    
    # Install gzcli
    install_gzcli
    
    # Verify gzcli installation
    if ! command -v gzcli &> /dev/null; then
        print_warning "gzcli is installed but not in PATH"
        print_info "Make sure $HOME/go/bin is in your PATH"
        print_info "Add this to your shell config: export PATH=\$PATH:\$HOME/go/bin"
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
    print_info "You can now use gzcli by running: gzcli --help"
    print_info "If the command is not found, restart your shell or run:"
    echo "  source $SHELL_CONFIG"
    echo ""
}

# Run main function
main

