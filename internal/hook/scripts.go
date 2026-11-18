package hook

import (
	"embed"
	"fmt"
	"strings"
)

//go:embed shell
var scripts embed.FS

// Shells are the shells spot init knows.
var Shells = []string{"zsh", "bash", "fish", "pwsh"}

var prompts = map[string]string{
	"zsh":  "setopt prompt_subst\nPROMPT='${SPOT_FACE} '\"$PROMPT\"",
	"bash": "PS1='${SPOT_FACE} '\"$PS1\"",
	"fish": "functions -q _spot_orig_prompt; or functions -c fish_prompt _spot_orig_prompt\nfunction fish_prompt\n    printf '%s ' $SPOT_FACE\n    _spot_orig_prompt\nend",
}

// Init returns the hook script for a shell. exe is the path of the spot
// binary, so the hook doesn't depend on PATH. With prompt, the script also
// puts Spot's face in front of your prompt.
func Init(shell, exe string, prompt bool) (string, error) {
	file := map[string]string{"zsh": "spot.zsh", "bash": "spot.bash", "fish": "spot.fish", "pwsh": "spot.ps1", "powershell": "spot.ps1"}[shell]
	if file == "" {
		return "", fmt.Errorf("spot doesn't know the %q shell yet; try one of: %s", shell, strings.Join(Shells, ", "))
	}
	b, err := scripts.ReadFile("shell/" + file)
	if err != nil {
		return "", err
	}
	p := ""
	if prompt {
		p = prompts[shell]
	}
	show := "$false"
	if prompt {
		show = "$true"
	}
	return strings.NewReplacer(
		"{{SPOT}}", shQuote(exe),
		"{{SPOTPS}}", strings.ReplaceAll(exe, "'", "''"),
		"{{PROMPT}}", p,
		"{{SHOWFACE}}", show,
	).Replace(string(b)), nil
}

// shQuote quotes a path for sh, zsh and fish.
func shQuote(s string) string {
	if s != "" && strings.Trim(s, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789/._-+") == "" {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
