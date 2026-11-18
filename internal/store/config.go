package store

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config is ~/.config/spot/config.toml. Every field is optional.
type Config struct {
	// Reactions after a test run: "full" (animation), "short" (one still
	// picture and a line) or "off".
	Reactions string `toml:"reactions"`
	// Theme picks Spot's colours: "auto", "dark" or "light". Auto looks at
	// COLORFGBG in the shell and asks the terminal in the full-screen app.
	Theme string `toml:"theme"`
	// Prompt face style: "face" (coloured kaomoji), "emoji" or "ascii".
	Prompt string `toml:"prompt"`
	// TestCommands replaces the built-in list of test commands.
	TestCommands []string `toml:"test_commands"`
	// ExtraTestCommands is added to the built-in list.
	ExtraTestCommands []string `toml:"extra_test_commands"`
	// IgnoreCommands are never treated as tests, even when they match.
	IgnoreCommands []string `toml:"ignore_commands"`
}

// ConfigPath is the config file location.
func ConfigPath() (string, error) {
	d, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "config.toml"), nil
}

// LoadConfig reads the config; a missing file is an empty config. On a
// broken file it still returns usable defaults alongside the error.
// Environment variables win: SPOT_QUIET=1 turns reactions off and
// SPOT_REDUCED_MOTION=1 makes them short.
func LoadConfig() (Config, error) {
	var c Config
	p, err := ConfigPath()
	if err == nil {
		if _, err = toml.DecodeFile(p, &c); errors.Is(err, os.ErrNotExist) {
			err = nil
		}
	}
	if err != nil {
		c = Config{}
	}
	switch c.Reactions {
	case "full", "short", "off":
	default:
		c.Reactions = "full"
	}
	if os.Getenv("SPOT_REDUCED_MOTION") != "" && c.Reactions == "full" {
		c.Reactions = "short"
	}
	if os.Getenv("SPOT_QUIET") != "" {
		c.Reactions = "off"
	}
	switch c.Theme {
	case "auto", "dark", "light":
	default:
		c.Theme = "auto"
	}
	switch c.Prompt {
	case "face", "emoji", "ascii":
	default:
		c.Prompt = "face"
	}
	return c, err
}

// Example is written by `spot config init`.
const Example = `# spot config. Every line is optional.

# What Spot does after a test run: "full" (a short animation),
# "short" (one still picture and a line) or "off".
reactions = "full"

# Colours: "auto", "dark" or "light". Spot has a light rim on dark
# terminals so a black dog doesn't disappear into a black background.
theme = "auto"

# Prompt face: "face" (coloured), "emoji" or "ascii".
prompt = "face"

# Add your own test commands. A word ending in * matches any word that
# starts with it, so "make test*" matches "make test-unit".
extra_test_commands = [
  # "./scripts/test.sh",
  # "task test*",
]

# Commands that look like tests but aren't.
ignore_commands = [
  # "go test -run XXX",
]

# Replace the built-in list completely (spot tests prints it).
# test_commands = ["go test", "npm test"]
`
