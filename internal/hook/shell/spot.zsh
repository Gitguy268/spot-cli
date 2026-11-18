# spot shell hook for zsh. Add this line to the end of ~/.zshrc:
#   eval "$(spot init zsh)"
# Then put $SPOT_FACE in your prompt, or use `spot init zsh --prompt`.

zmodload zsh/datetime 2>/dev/null
typeset -g _spot_cmd="" _spot_start="" SPOT_FACE=""

_spot_preexec() {
  _spot_cmd="$1"
  _spot_start="$EPOCHREALTIME"
}

_spot_precmd() {
  local code=$?
  if [[ -n "$_spot_cmd" ]]; then
    SPOT_FACE="$({{SPOT}} _event --shell zsh --exit "$code" --since "$_spot_start" -- "$_spot_cmd")"
    _spot_cmd=""
  else
    SPOT_FACE="$({{SPOT}} prompt --shell zsh)"
  fi
  return $code
}

spot_prompt() { print -rn -- "$SPOT_FACE"; }

autoload -Uz add-zsh-hook
add-zsh-hook preexec _spot_preexec
# First in line, so it sees the exit code of your command, not of another hook.
precmd_functions=(_spot_precmd ${precmd_functions:#_spot_precmd})
{{PROMPT}}
