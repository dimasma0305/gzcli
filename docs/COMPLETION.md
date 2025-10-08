# Shell Completion

gzcli provides intelligent shell completion for all commands and flags, including dynamic event name completion.

## Quick Setup

### Bash

```bash
# Load completions for current session
source <(gzcli completion bash)

# Install completions permanently
# Linux:
gzcli completion bash > /etc/bash_completion.d/gzcli

# macOS:
gzcli completion bash > $(brew --prefix)/etc/bash_completion.d/gzcli
```

### Zsh

```bash
# Enable completion in your shell (if not already enabled)
echo "autoload -U compinit; compinit" >> ~/.zshrc

# Install completions
gzcli completion zsh > "${fpath[1]}/_gzcli"

# Reload shell
exec zsh
```

### Fish

```bash
# Load completions for current session
gzcli completion fish | source

# Install completions permanently
gzcli completion fish > ~/.config/fish/completions/gzcli.fish
```

### PowerShell

```powershell
# Load completions for current session
gzcli completion powershell | Out-String | Invoke-Expression

# Install completions permanently
gzcli completion powershell > gzcli.ps1
# Then add this to your PowerShell profile
```

## Features

### Dynamic Event Completion

The `--event` flag provides intelligent completion by scanning the `events/` directory:

```bash
# Type and press TAB
gzcli watch start --event <TAB>
# Shows: ctf2024  ctf2025  training

gzcli watch stop --event <TAB>
# Shows: ctf2024  ctf2025  training

gzcli event switch <TAB>
# Shows: ctf2024  ctf2025  training
```

### Multi-Event Completion

The watch start command supports multiple events with completion:

```bash
gzcli watch start --event ctf2024 --event <TAB>
# Shows remaining available events
```

### Completion Features

1. **Command Completion**: All commands and subcommands
   ```bash
   gzcli <TAB>
   # Shows: sync, watch, event, team, etc.
   ```

2. **Flag Completion**: All flags with descriptions
   ```bash
   gzcli watch start --<TAB>
   # Shows: --event, --foreground, --debounce, etc.
   ```

3. **Event Name Completion**: Dynamic from `events/` directory
   ```bash
   gzcli --event <TAB>
   # Shows: ctf2024, ctf2025, training
   ```

4. **Argument Completion**: Context-aware argument suggestions
   ```bash
   gzcli event switch <TAB>
   # Shows available event names
   ```

## How It Works

### Event Discovery

The completion system automatically discovers events by:
1. Scanning the `events/` directory
2. Checking for valid `.gzevent` configuration files
3. Listing only directories with valid event configurations

### Real-Time Updates

Completion reflects the current state:
- New events appear automatically after creation
- Removed events disappear from suggestions
- No caching - always shows current state

## Troubleshooting

### Completions Not Working

**Bash:**
```bash
# Check if bash-completion is installed
apt-get install bash-completion  # Debian/Ubuntu
brew install bash-completion@2    # macOS

# Reload completions
source ~/.bashrc
```

**Zsh:**
```bash
# Verify completion is enabled
echo $fpath

# Clear completion cache
rm -f ~/.zcompdump*
compinit
```

**Fish:**
```bash
# Check completion file exists
ls ~/.config/fish/completions/gzcli.fish

# Reload fish config
source ~/.config/fish/config.fish
```

### Completions Show Wrong Events

The completion system reads from the current working directory's `events/` folder. Ensure you're in the correct project directory:

```bash
# Check current directory
pwd

# Verify events directory exists
ls -la events/
```

## Examples

### Complete Workflow

```bash
# 1. Generate and install completions
gzcli completion bash > ~/.local/share/bash-completion/completions/gzcli
source ~/.bashrc

# 2. Use completion for event selection
gzcli watch start --event <TAB>
→ ctf2024  ctf2025

# 3. Select event
gzcli watch start --event ctf2024

# 4. Use completion for status
gzcli watch status --event <TAB>
→ ctf2024  ctf2025

# 5. Use completion for switching
gzcli event switch <TAB>
→ ctf2024  ctf2025
```

### Multiple Events

```bash
# Watch multiple events with completion
gzcli watch start --event ctf<TAB>
→ ctf2024  ctf2025

gzcli watch start --event ctf2024 --event ctf<TAB>
→ ctf2025  # (ctf2024 already selected, shows others)
```

## Advanced Usage

### Custom Completion Scripts

You can extend the completion by modifying the generated script:

```bash
# Generate base completion
gzcli completion bash > ~/.gzcli-completion.bash

# Add custom completions
echo 'complete -W "dev staging prod" gzcli-deploy' >> ~/.gzcli-completion.bash

# Source in .bashrc
echo 'source ~/.gzcli-completion.bash' >> ~/.bashrc
```

### Debugging Completions

**Bash:**
```bash
# Enable debug mode
export BASH_COMP_DEBUG_FILE=/tmp/bash-completion-debug.log
# Use completions
# Check log
cat /tmp/bash-completion-debug.log
```

**Zsh:**
```bash
# Verbose completion
setopt BASH_COMPLETION_DEBUG
# Use completions
```

## Performance

The completion system is optimized for speed:
- **Event Discovery**: < 10ms for typical projects
- **Caching**: No caching (always fresh)
- **Lazy Loading**: Only scans when completion is triggered

For projects with 100+ events, completion may take 20-50ms (still fast).

## See Also

- [Multi-Event Management](MULTI_EVENT.md)
- [Watch Command Reference](../README.md#file-watcher)
- [Event Commands](../README.md#event-management)
