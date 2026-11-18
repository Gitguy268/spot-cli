// Package react draws Spot outside the full-screen app: the little face in
// your prompt and the short animation after a test run.
package react

import (
	"strings"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"

	"github.com/Gitguy268/spot-cli/internal/pet"
)

type face struct{ eyes, extra, emoji, ascii string }

var faces = map[pet.Mood]face{
	pet.Content: {"•ᴥ•", "", "🐶", "U.w.U"},
	pet.Happy:   {"^ᴥ^", "", "🐶💕", "U^w^U"},
	pet.Proud:   {"^ᴥ^", "✦", "🐶✨", "U^w^U*"},
	pet.Sad:     {"╥ᴥ╥", "", "🐶💧", "U;w;U"},
	pet.Hungry:  {"°ᴥ°", "🦴", "🐶🦴", "U.w.U?"},
	pet.Sleepy:  {"=ᴥ=", "", "🐶🥱", "U=w=U"},
	pet.Asleep:  {"-ᴥ-", "ᶻᶻ", "🐶💤", "U-w-Uz"},
}

// Face is Spot's prompt face for a mood. style is "face", "emoji" or
// "ascii"; shell decides how colour codes are wrapped so the shell can count
// the prompt width: %{…%} for zsh, \001…\002 for bash.
func Face(m pet.Mood, style, shell string, p colorprofile.Profile) string {
	f, ok := faces[m]
	if !ok {
		f = faces[pet.Content]
	}
	switch style {
	case "emoji":
		return f.emoji
	case "ascii":
		return f.ascii
	}
	if p <= colorprofile.ASCII {
		return strings.TrimSpace("U" + f.eyes + "U" + f.extra)
	}
	ear := colourise(p, "#8a90b8", shell)
	eye := colourise(p, "#f2f4ff", shell)
	nose := colourise(p, "#ff8fb4", shell)
	tag := colourise(p, "#ffc94a", shell)
	r := []rune(f.eyes)
	out := ear("U") + eye(string(r[0])) + nose(string(r[1])) + eye(string(r[2])) + ear("U")
	if f.extra != "" {
		out += tag(f.extra)
	}
	return out
}

func colourise(p colorprofile.Profile, hex, shell string) func(string) string {
	c := p.Convert(ansi.XParseColor(hex))
	start := ansi.Style{}.ForegroundColor(c).String()
	end := ansi.ResetStyle
	wrap := func(s string) string { return s }
	switch shell {
	case "zsh":
		wrap = func(s string) string { return "%{" + s + "%}" }
	case "bash":
		wrap = func(s string) string { return "\x01" + s + "\x02" }
	}
	return func(s string) string { return wrap(start) + s + wrap(end) }
}
