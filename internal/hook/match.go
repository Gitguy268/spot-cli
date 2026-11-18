// Package hook decides which shell commands are test runs and holds the
// scripts that `spot init <shell>` prints.
package hook

import (
	"path/filepath"
	"strings"
)

// Defaults are the test commands Spot knows out of the box. A word ending in
// * matches any word that starts with it.
var Defaults = []string{
	"go test", "gotestsum",
	"npm test", "npm t", "npm run test*", "npm run e2e*",
	"pnpm test", "pnpm run test*", "pnpm t",
	"yarn test", "yarn run test*",
	"bun test", "bun run test*",
	"npx vitest", "npx jest", "npx playwright test", "npx mocha", "npx ava",
	"vitest", "jest", "mocha", "ava", "playwright test",
	"pytest", "py.test", "python -m pytest", "python3 -m pytest", "python -m unittest", "python3 -m unittest",
	"uv run pytest", "poetry run pytest", "hatch test", "tox", "nox",
	"cargo test", "cargo nextest",
	"make test*", "make check", "just test*", "task test*",
	"swift test", "xcodebuild test", "xcodebuild test-without-building",
	"mvn test", "mvn verify", "mvnw test", "mvnw verify", "gradle test", "gradlew test", "gradle check", "gradlew check",
	"rspec", "bundle exec rspec", "rake test", "rails test", "bin/rails test",
	"mix test", "dotnet test", "deno test", "zig build test", "ctest",
	"phpunit", "pest", "composer test",
	"flutter test", "dart test", "stack test", "cabal test", "busted", "bats", "rebar3 eunit", "sbt test", "lein test",
	"crystal spec", "nimble test", "dune test",
}

// wrappers are skipped before the real command word.
var wrappers = map[string]bool{
	"sudo": true, "time": true, "command": true, "env": true, "nice": true, "nohup": true,
	"caffeinate": true, "exec": true, "noglob": true, "builtin": true, "doas": true,
}

// Matcher knows which commands are tests.
type Matcher struct {
	patterns [][]string
	ignore   [][]string
}

// NewMatcher builds a matcher. replace swaps out the defaults when it is not
// empty; extra adds to them; ignore wins over both.
func NewMatcher(replace, extra, ignore []string) *Matcher {
	base := Defaults
	if len(replace) > 0 {
		base = replace
	}
	m := &Matcher{}
	for _, p := range append(append([]string{}, base...), extra...) {
		if f := strings.Fields(p); len(f) > 0 {
			m.patterns = append(m.patterns, f)
		}
	}
	for _, p := range ignore {
		if f := strings.Fields(p); len(f) > 0 {
			m.ignore = append(m.ignore, f)
		}
	}
	return m
}

// IsTest reports whether a command line runs tests. The line is split on
// ; && || | and newlines, and any part that matches counts. Lines that start
// with spot itself never count: `spot run` already told Spot.
func (m *Matcher) IsTest(line string) bool {
	segs := segments(line)
	if len(segs) > 0 && len(segs[0]) > 0 && base(segs[0][0]) == "spot" {
		return false
	}
	for _, words := range segs {
		for _, ig := range m.ignore {
			if matches(ig, words) {
				return false
			}
		}
	}
	for _, words := range segs {
		for _, p := range m.patterns {
			if matches(p, words) {
				return true
			}
		}
	}
	return false
}

// segments splits a command line into simple commands, each as its words
// with leading VAR=value assignments and wrappers like sudo removed.
func segments(line string) [][]string {
	r := strings.NewReplacer("&&", "\n", "||", "\n", ";", "\n", "|", "\n", "&", "\n", "(", " ", ")", " ", "{", " ", "}", " ")
	var out [][]string
	for _, part := range strings.Split(r.Replace(line), "\n") {
		words := strings.Fields(part)
		wrapped := false
	strip:
		for len(words) > 0 {
			w := words[0]
			switch {
			case wrappers[base(w)]:
				wrapped = true
			case assignment(w), w == "!":
			case wrapped && strings.HasPrefix(w, "-"): // sudo -E, env -i
			default:
				break strip
			}
			words = words[1:]
		}
		if len(words) > 0 {
			out = append(out, words)
		}
	}
	return out
}

// assignment is a leading NAME=value word.
func assignment(w string) bool {
	i := strings.IndexByte(w, '=')
	if i <= 0 {
		return false
	}
	for j, c := range w[:i] {
		if !(c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || j > 0 && c >= '0' && c <= '9') {
			return false
		}
	}
	return true
}

func base(w string) string {
	w = strings.Trim(w, `"'`)
	return filepath.Base(strings.ReplaceAll(w, `\`, "/"))
}

// matches checks a pattern against the start of a command. The first word is
// compared by its base name (so /usr/local/go/bin/go and ./gradlew work).
// Flags between later pattern words are skipped: `make -j8 test`.
func matches(p, words []string) bool {
	if len(words) == 0 || !(word(p[0], words[0]) || word(p[0], base(words[0]))) {
		return false
	}
	i := 1
	for _, pw := range p[1:] {
		for i < len(words) && strings.HasPrefix(words[i], "-") && !strings.HasPrefix(pw, "-") {
			i++
		}
		if i >= len(words) || !word(pw, words[i]) {
			return false
		}
		i++
	}
	return true
}

func word(pattern, w string) bool {
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(w, strings.TrimSuffix(pattern, "*"))
	}
	return pattern == w
}
