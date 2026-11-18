// Command spot is a terminal pet: Spot the black labrador lives in your
// shell, gets hungry, and wags when your tests pass.
package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/term"

	"github.com/Gitguy268/spot-cli/internal/art"
	"github.com/Gitguy268/spot-cli/internal/react"
	"github.com/Gitguy268/spot-cli/internal/store"
)

// version is set at build time with -ldflags "-X main.version=…".
var version = ""

func main() { os.Exit(run(os.Args[1:])) }

func run(args []string) int {
	cmd, rest := "", args
	if len(args) > 0 {
		cmd, rest = args[0], args[1:]
	}
	switch cmd {
	case "":
		return cmdTUI()
	case "run":
		return cmdRun(rest)
	case "init":
		return cmdInit(rest)
	case "prompt":
		return cmdPrompt(rest)
	case "_event":
		return cmdEvent(rest)
	case "status", "st":
		return cmdStatus(rest)
	case "feed", "treat":
		return cmdFeed()
	case "pet":
		return cmdPet()
	case "trick", "tricks":
		return cmdTrick(rest)
	case "badges":
		return cmdBadges()
	case "stats":
		return cmdStats()
	case "rename":
		return cmdRename(rest)
	case "reset":
		return cmdReset(rest)
	case "show":
		return cmdShow(rest)
	case "config":
		return cmdConfig(rest)
	case "tests":
		return cmdTests()
	case "version", "--version", "-v":
		fmt.Println("spot", getVersion())
		return 0
	case "help", "--help", "-h":
		fmt.Print(help)
		return 0
	}
	fmt.Fprintf(os.Stderr, "spot: %q isn't a command Spot knows. Try: spot help\n", cmd)
	return 2
}

const help = `spot: a black labrador who lives in your terminal

  spot                 open Spot's room (pet, feed, fetch, tricks)
  spot run -- CMD      run your tests with Spot watching; keeps CMD's exit code
  spot init SHELL      print the shell hook (zsh, bash, fish, pwsh)
                       add --prompt to also put Spot's face in your prompt
  spot status          how Spot is doing (--json for scripts)
  spot feed            give him a treat, or his free kibble
  spot pet             scratch behind the ears
  spot tricks          tricks he knows; spot trick NAME to see one
  spot badges          badges you've earned together
  spot stats           your test history, as Spot remembers it
  spot rename NAME     he answers to anything said kindly
  spot reset           start over with a new puppy
  spot config [init]   show the config path, or write an example config
  spot tests           the commands Spot treats as tests
  spot show [ANIM]     play one of Spot's animations
  spot version

Set up (zsh):  echo 'eval "$(spot init zsh --prompt)"' >> ~/.zshrc
Quiet for now: SPOT_QUIET=1   Less motion: SPOT_REDUCED_MOTION=1
`

func getVersion() string {
	if version != "" {
		return version
	}
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		return strings.TrimPrefix(bi.Main.Version, "v")
	}
	return "dev"
}

// env bundles what most commands need.
type env struct {
	cfg     store.Config
	cfgErr  error
	st      *store.Store
	now     time.Time
	profile colorprofile.Profile // for drawing on stderr
	tty     bool                 // stderr is a terminal
	width   int
}

func newEnv() (*env, error) {
	cfg, cfgErr := store.LoadConfig()
	s, err := store.Open()
	if err != nil {
		return nil, err
	}
	e := &env{cfg: cfg, cfgErr: cfgErr, st: s, now: time.Now()}
	e.tty = term.IsTerminal(os.Stderr.Fd())
	e.profile = colorprofile.Detect(os.Stderr, os.Environ())
	if w, _, err := term.GetSize(os.Stderr.Fd()); err == nil {
		e.width = w
	}
	return e, nil
}

// light follows the theme setting; "auto" reads COLORFGBG ("15;0" is
// light text on dark), and assumes dark when it isn't set.
func (e *env) light() bool {
	switch e.cfg.Theme {
	case "light":
		return true
	case "dark":
		return false
	}
	if v := os.Getenv("COLORFGBG"); v != "" {
		parts := strings.Split(v, ";")
		switch parts[len(parts)-1] {
		case "7", "15":
			return true
		}
	}
	return false
}

func (e *env) palette() art.Palette {
	if e.light() {
		return art.Light
	}
	return art.Dark
}

// opts draws on stderr, which stays the terminal even when stdout is
// captured by the prompt hook.
func (e *env) opts() react.Opts {
	mode := e.cfg.Reactions
	if !e.tty {
		mode = "off"
	}
	return react.Opts{Out: os.Stderr, Profile: e.profile, Pal: e.palette(), Light: e.light(), Width: e.width, Mode: mode}
}

func fail(err error) int {
	fmt.Fprintln(os.Stderr, "spot:", err)
	return 1
}
