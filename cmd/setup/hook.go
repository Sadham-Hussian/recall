package setup

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var hookCmd = &cobra.Command{
	Use:   "hook [shell]",
	Short: "Initialize recall shell integration",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		shell := strings.ToLower(args[0])

		switch shell {

		case "zsh":
			fmt.Print(zshHook())

		case "bash":
			fmt.Print(bashHook())

		case "fish":
			fmt.Print(fishHook())

		default:
			fmt.Println("Unsupported shell. Supported: zsh, bash, fish")
		}
	},
}

func GetHookCmd() *cobra.Command {
	return hookCmd
}

func zshHook() string {
	return `
autoload -Uz add-zsh-hook

recall_preexec() {
  if [[ "$1" == recall* || "$1" == */recall* ]]; then
    return
  fi
  export RECALL_LAST_COMMAND="$1"
}

recall_precmd() {
  local exit_code=$?
  local cwd="$PWD"
  local timestamp=$(date +%s)

  if [[ -n "$RECALL_LAST_TIMESTAMP" ]]; then
    local gap=$((timestamp - RECALL_LAST_TIMESTAMP))
    if [[ $gap -gt 600 ]]; then
      export RECALL_SESSION_ID="$$-$timestamp"
    fi
  fi

  if [[ -z "$RECALL_SESSION_ID" ]]; then
    export RECALL_SESSION_ID="$$-$timestamp"
  fi

  if [[ -n "$RECALL_LAST_COMMAND" ]]; then
    recall record \
      --cmd "$RECALL_LAST_COMMAND" \
      --exit "$exit_code" \
      --cwd "$cwd" \
      --ts "$timestamp" \
      --shell-pid "$$" \
      --session-id "$RECALL_SESSION_ID" \
      >/dev/null 2>&1 &!

    unset RECALL_LAST_COMMAND
  fi

  export RECALL_LAST_TIMESTAMP=$timestamp
}

add-zsh-hook preexec recall_preexec
add-zsh-hook precmd recall_precmd
`
}

func bashHook() string {
	return `
recall_preexec() {
  case "$BASH_COMMAND" in
    recall*|*/recall*) return ;;
  esac
  export RECALL_LAST_COMMAND="$BASH_COMMAND"
}

recall_precmd() {
  local exit_code=$?
  local cwd="$PWD"
  local timestamp=$(date +%s)

  if [[ -n "$RECALL_LAST_TIMESTAMP" ]]; then
    local gap=$((timestamp - RECALL_LAST_TIMESTAMP))
    if [[ $gap -gt 600 ]]; then
      export RECALL_SESSION_ID="$$-$timestamp"
    fi
  fi

  if [[ -z "$RECALL_SESSION_ID" ]]; then
    export RECALL_SESSION_ID="$$-$timestamp"
  fi

  if [[ -n "$RECALL_LAST_COMMAND" ]]; then
    ( recall record \
      --cmd "$RECALL_LAST_COMMAND" \
      --exit "$exit_code" \
      --cwd "$cwd" \
      --ts "$timestamp" \
      --shell-pid "$$" \
      --session-id "$RECALL_SESSION_ID" \
      >/dev/null 2>&1 & )

    unset RECALL_LAST_COMMAND
  fi

  export RECALL_LAST_TIMESTAMP=$timestamp
}

trap 'recall_preexec' DEBUG
PROMPT_COMMAND="recall_precmd;$PROMPT_COMMAND"
`
}

func fishHook() string {
	return `
function recall_preexec --on-event fish_preexec
    set -gx RECALL_LAST_COMMAND $argv
end

function recall_precmd --on-event fish_prompt
    set exit_code $status
    set cwd (pwd)
    set timestamp (date +%s)

    if test -n "$RECALL_LAST_TIMESTAMP"
        set gap (math $timestamp - $RECALL_LAST_TIMESTAMP)
        if test $gap -gt 600
            set -gx RECALL_SESSION_ID "$fish_pid-$timestamp"
        end
    end

    if test -z "$RECALL_SESSION_ID"
        set -gx RECALL_SESSION_ID "$fish_pid-$timestamp"
    end

    if test -n "$RECALL_LAST_COMMAND"
        recall record \
            --cmd "$RECALL_LAST_COMMAND" \
            --exit "$exit_code" \
            --cwd "$cwd" \
            --ts "$timestamp" \
            --shell-pid "$fish_pid" \
            --session-id "$RECALL_SESSION_ID" \
            >/dev/null 2>&1 &
    end

    set -gx RECALL_LAST_TIMESTAMP $timestamp
end
`
}
