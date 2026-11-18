# spot shell hook for bash. Add this line to the end of ~/.bashrc:
#   eval "$(spot init bash)"
# Then put $SPOT_FACE in PS1, or use `spot init bash --prompt`.

_spot_armed="" _spot_start="" _spot_first="" _spot_hist="" SPOT_FACE=""
_spot_re='^ *([0-9]+)\*? +(.*)$'

_spot_now() {
  if [[ -n "$EPOCHREALTIME" ]]; then printf '%s' "$EPOCHREALTIME"; else printf '%s' "$SECONDS"; fi
}

# Runs before each command (bash-preexec, or our DEBUG trap). $1 is the
# command line (bash-preexec) or its first simple command (DEBUG trap).
_spot_preexec() {
  [[ -n "$_spot_armed" ]] || return 0
  _spot_armed=""
  _spot_start="$(_spot_now)"
  _spot_first="$1"
}

_spot_precmd() {
  local code=$? line num="" cmd=""
  line="$(HISTTIMEFORMAT='' builtin history 1)"
  if [[ "$line" =~ $_spot_re ]]; then num="${BASH_REMATCH[1]}" cmd="${BASH_REMATCH[2]}"; fi
  # A new history entry is your command. With ignoredups or erasedups a
  # repeat adds none, so also accept the last entry when it starts with what
  # just ran; a command hidden by a leading space won't match it.
  if [[ -n "$_spot_start" && -n "$cmd" ]] && { [[ -n "$num" && "$num" != "$_spot_hist" ]] || [[ -n "$_spot_first" && "$cmd" == "$_spot_first"* ]]; }; then
    SPOT_FACE="$({{SPOT}} _event --shell bash --exit "$code" --since "$_spot_start" -- "$cmd")"
  else
    SPOT_FACE="$({{SPOT}} prompt --shell bash)"
  fi
  _spot_hist="$num" _spot_start="" _spot_first=""
  return $code
}

# Last in PROMPT_COMMAND: from here on, the next DEBUG trap is your command.
_spot_arm() { _spot_armed=1; }

spot_prompt() { printf '%s' "$SPOT_FACE"; }

if [[ -n "${bash_preexec_imported:-}${__bp_imported:-}" ]]; then
  preexec_functions+=(_spot_preexec)
  precmd_functions=(_spot_precmd "${precmd_functions[@]}" _spot_arm)
else
  _spot_debug() {
    [[ "$BASH_COMMAND" == _spot_* || -n "$COMP_LINE" ]] && return 0
    _spot_preexec "$BASH_COMMAND"
  }
  _spot_old_trap="$(trap -p DEBUG)"
  _spot_old_trap="${_spot_old_trap#trap -- \'}"; _spot_old_trap="${_spot_old_trap%\' DEBUG}"
  if [[ -n "$_spot_old_trap" ]]; then
    trap "_spot_debug; $_spot_old_trap" DEBUG
  else
    trap '_spot_debug' DEBUG
  fi
  if [[ "$(declare -p PROMPT_COMMAND 2>/dev/null)" == "declare -a"* ]]; then
    PROMPT_COMMAND=(_spot_precmd "${PROMPT_COMMAND[@]}" _spot_arm)
  else
    PROMPT_COMMAND="_spot_precmd${PROMPT_COMMAND:+; $PROMPT_COMMAND}; _spot_arm"
  fi
fi
[[ "$(HISTTIMEFORMAT='' builtin history 1)" =~ $_spot_re ]] && _spot_hist="${BASH_REMATCH[1]}"
{{PROMPT}}
