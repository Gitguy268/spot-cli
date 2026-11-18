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

// Scheme is the text colours for one terminal background.
type Scheme struct{ Text, Dim, Name, Gold, Ear, Eye, Nose, Red string }

// Colours returns the text colours for a dark or light terminal.
func Colours(light bool) Scheme {
	if light {
		return Scheme{Text: "#2a2b36", Dim: "#6b6f8f", Name: "#c62f36", Gold: "#a8700c", Ear: "#5d6284", Eye: "#1b1c24", Nose: "#d64a80", Red: "#c62f36"}
	}
	return Scheme{Text: "#d8dcf5", Dim: "#9aa0c0", Name: "#ff6b6b", Gold: "#ffc94a", Ear: "#8a90b8", Eye: "#f2f4ff", Nose: "#ff8fb4", Red: "#ff6b6b"}
}

// Face is Spot's prompt face for a mood. style is "face", "emoji" or
// "ascii"; shell decides how colour codes are wrapped so the shell can count
// the prompt width: %{…%} for zsh, \001…\002 for bash.
func Face(m pet.Mood, style, shell string, p colorprofile.Profile, light bool) string {
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
	c := Colours(light)
	ear := colourise(p, c.Ear, shell)
	eye := colourise(p, c.Eye, shell)
	nose := colourise(p, c.Nose, shell)
	tag := colourise(p, c.Gold, shell)
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
