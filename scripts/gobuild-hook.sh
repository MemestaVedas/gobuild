#!/usr/bin/env bash
# gobuild-hook.sh
# 
# Source this file in your ~/.bashrc or ~/.zshrc to enable the goBuild
# command interception feature.
#
# Usage:
#   source /path/to/gobuild-hook.sh

_gobuild_bold() { echo -e "\033[1m$1\033[0m"; }
_gobuild_info() { echo -e "\033[36m$1\033[0m"; }
_gobuild_success() { echo -e "\033[32m$1\033[0m"; }

_gobuild_intercept() {
    local cmd=$1
    shift
    
    # Check if we are inside a watched directory
    local cwd=$(pwd)
    local watch_file="$HOME/.config/gobuild/watch.json"
    
    # Construct full command string
    local full_cmd="$cmd"
    if [ $# -gt 0 ]; then
        full_cmd="$cmd $*"
    fi
    
    # If watch file doesn't exist, just run the command normally
    if [ ! -f "$watch_file" ]; then
        command "$cmd" "$@"
        return $?
    fi
    
    # Check if current directory (or its parents) is watched
    if command -v jq &> /dev/null; then
        # Normalise paths by removing trailing slashes for comparison
        local norm_cwd="${cwd%/}"
        
        local is_watched=$(jq -r --arg cwd "$norm_cwd" --arg cmd "$full_cmd" '
            .Directories[] | 
            .Path as $p | 
            ($p | sub("/$"; "")) as $norm_p |
            select($cwd == $norm_p or ($cwd | startswith($norm_p + "/"))) | 
            .Commands[] | select(. == $cmd)
        ' "$watch_file" 2>/dev/null)
        
        if [ -n "$is_watched" ]; then
            # Command is watched! Proxy it to goBuild daemon.
            if command -v gobuild &> /dev/null; then
                _gobuild_info "[goBuild] Intercepting: $full_cmd"
                gobuild proxy "$cmd" "$@"
                return $?
            else
                # Check for dev path
                local dev_bin="$HOME/project/goBuild/gobuild/gobuild"
                if [ -f "$dev_bin" ]; then
                     _gobuild_info "[goBuild] Intercepting (dev): $full_cmd"
                     "$dev_bin" proxy "$cmd" "$@"
                     return $?
                fi
            fi
        fi
    fi
    
    # Not watched or jq missing, run normally. 
    # Use a hidden/faint color for skip message to avoid noise.
    # To debug why something isn't showing, uncomment the line below:
    # _gobuild_info "  [goBuild] Skipped: not in watch list" &> /dev/null
    
    command "$cmd" "$@"
    return $?
}

# Aliases for common dev tools we might want to watch
alias npm='_gobuild_intercept npm'
alias yarn='_gobuild_intercept yarn'
alias pnpm='_gobuild_intercept pnpm'
alias cargo='_gobuild_intercept cargo'
alias make='_gobuild_intercept make'
alias go='_gobuild_intercept go'

_gobuild_success "✔ goBuild shell integration active."
_gobuild_info "  Watching commands in directories defined in ~/.config/gobuild/watch.json"
