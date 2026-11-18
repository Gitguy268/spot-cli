# Sourced (hidden) by the VHS tapes: a throwaway Spot, a tiny Go project
# with one test, and the zsh hook. Run the tapes from the repo root:
#   vhs docs/tapes/hook.tape
# Pass "light" to use Spot's light-terminal colours.
SPOT_REPO=$PWD
go build -o "$SPOT_REPO/spot" "$SPOT_REPO" || return
export SPOT_HOME="$(mktemp -d)"
export SPOT_CONFIG="$SPOT_HOME" PATH="$SPOT_REPO:$PATH"
[ "$1" = light ] && printf 'theme = "light"\n' > "$SPOT_CONFIG/config.toml"
cd "$(mktemp -d)" || return
printf 'module example.com/app\n\ngo 1.22\n' > go.mod
cat > app_test.go <<'GO'
package app

import (
	"os"
	"testing"
)

func TestAdd(t *testing.T) {
	if os.Getenv("BREAK") != "" {
		t.Fatal("2 + 2 = 5")
	}
}
GO
go test ./... > /dev/null 2>&1
PROMPT='%F{blue}~/my-app%f %# '
eval "$(spot init zsh --prompt)"
clear
