#!/usr/bin/env bash
# gobuild-hook.sh
# 
# Source this file in your ~/.bashrc or ~/.zshrc to enable the goBuild
# command interception feature.
#
# Usage:
#   source /path/to/gobuild-hook.sh

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
    
    # Very basic check: 
    # Use jq if available to check if the current directory is watched for this command.
    if command -v jq &> /dev/null; then
        local is_watched=$(jq -r --arg cwd "$cwd" --arg cmd "$full_cmd" '
            .directories[] | .path as $p | select($cwd | startswith($p)) | .commands[] | select(. == $cmd)
        ' "$watch_file" 2>/dev/null)
        
        if [ -n "$is_watched" ]; then
            # Command is watched! Proxy it to goBuild daemon.
            # We assume `gobuild` is in the PATH.
            if command -v gobuild &> /dev/null; then
                echo -e "\033[36m[goBuild] Intercepting: $full_cmd\033[0m"
                gobuild proxy "$cmd" "$@"
                return $?
            else
                # Fallback if gobuild is not in PATH but gobuild can be run via go run
                # (For development testing)
                if [ -d "$HOME/project/goBuild/gobuild" ]; then
                    echo -e "\033[36m[goBuild] Intercepting (dev mode): $full_cmd\033[0m"
                    (cd "$HOME/project/goBuild/gobuild" && go run cmd/gobuild/main.go proxy "$cmd" "$@")
                    return $?
                fi
            fi
        fi
    fi
    
    # Not watched or jq missing, run normally
    command "$cmd" "$@"
}

# Aliases for common dev tools we might want to watch
alias npm='_gobuild_intercept npm'
alias yarn='_gobuild_intercept yarn'
alias pnpm='_gobuild_intercept pnpm'
alias cargo='_gobuild_intercept cargo'
alias make='_gobuild_intercept make'
alias go='_gobuild_intercept go'
