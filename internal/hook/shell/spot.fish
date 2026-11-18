# spot shell hook for fish. Add this line to ~/.config/fish/config.fish:
#   {{SPOT}} init fish | source
# Then print $SPOT_FACE in fish_prompt, or use `spot init fish --prompt`.

set -g SPOT_FACE ({{SPOT}} prompt --shell fish)

function _spot_postexec --on-event fish_postexec
    set -l code $status
    set -l secs (math "$CMD_DURATION / 1000")
    set -g SPOT_FACE ({{SPOT}} _event --shell fish --exit $code --dur $secs -- $argv[1])
end

function spot_prompt
    printf '%s' $SPOT_FACE
end
{{PROMPT}}
