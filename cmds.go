package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/term"

	"github.com/Gitguy268/spot-cli/internal/art"
	"github.com/Gitguy268/spot-cli/internal/hook"
	"github.com/Gitguy268/spot-cli/internal/pet"
	"github.com/Gitguy268/spot-cli/internal/react"
	"github.com/Gitguy268/spot-cli/internal/store"
	"github.com/Gitguy268/spot-cli/internal/tui"
)

// stdoutOpts draws on stdout, for commands you run by hand.
func (e *env) stdoutOpts() react.Opts {
	o := e.opts()
	o.Out = os.Stdout
	o.Profile = colorprofile.Detect(os.Stdout, os.Environ())
	if !term.IsTerminal(os.Stdout.Fd()) {
		o.Mode = "short"
	}
	if w, _, err := term.GetSize(os.Stdout.Fd()); err == nil {
		o.Width = w
	}
	return o
}

// update changes Spot's state and reports a corrupt old file once.
func (e *env) update(fn func(*pet.State) error) (*pet.State, error) {
	st, err := e.st.Update(e.now, fn)
	if errors.Is(err, store.ErrCorrupt) {
		fmt.Fprintln(os.Stderr, "spot:", err, "· Spot starts over as a puppy")
		err = nil
	}
	return st, err
}

// peek reads Spot without saving, for the prompt.
func (e *env) peek() *pet.State {
	st, err := e.st.Load(e.now)
	if st == nil || err != nil {
		st = pet.New("", e.now)
	}
	st.Tick(e.now)
	return st
}

func cmdTUI() int {
	e, err := newEnv()
	if err != nil {
		return fail(err)
	}
	if !term.IsTerminal(os.Stdout.Fd()) {
		return cmdStatus(nil)
	}
	return tui.Run(e.st, e.cfg)
}

// interruptCodes are exits that aren't really a test result: Ctrl-C,
// Ctrl-\, suspend, and "command not found" or "not executable".
var interruptCodes = []int{126, 127, 130, 131, 146, 148}

func cmdRun(args []string) int {
	quiet := false
	for len(args) > 0 && strings.HasPrefix(args[0], "-") {
		a := args[0]
		args = args[1:]
		if a == "--" {
			break
		}
		if a == "-q" || a == "--quiet" {
			quiet = true
			continue
		}
		fmt.Fprintf(os.Stderr, "spot run: unknown flag %s (put your command after --)\n", a)
		return 2
	}
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "spot run: what should Spot watch? e.g. spot run -- go test ./...")
		return 2
	}
	e, err := newEnv()
	if err != nil {
		return fail(err)
	}
	o := e.opts()
	if quiet {
		o.Mode = "off"
	}
	line := strings.Join(args, " ")
	if o.Mode != "off" {
		st := e.peek()
		fmt.Fprintln(os.Stderr, react.Face(pet.Content, "face", "", o.Profile)+"  "+st.Name+" sits and watches "+line)
	}

	c := exec.Command(args[0], args[1:]...)
	c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM) // the child gets them from the terminal too
	start := time.Now()
	err = c.Run()
	signal.Stop(sig)
	code := exitCode(c, err)
	if code == 127 {
		fmt.Fprintf(os.Stderr, "spot run: %v\n", err)
	}
	if slices.Contains(interruptCodes, code) {
		return code
	}
	e.now = time.Now()
	var r pet.Reaction
	st, uerr := e.update(func(st *pet.State) error {
		if code == 0 {
			r = st.TestPassed(line, time.Since(start), e.now)
		} else {
			r = st.TestFailed(line, time.Since(start), e.now)
		}
		return nil
	})
	if uerr == nil {
		react.Play(o, r, st)
	}
	return code
}

func exitCode(c *exec.Cmd, err error) int {
	if err == nil {
		return 0
	}
	if c.ProcessState == nil {
		return 127
	}
	if code := c.ProcessState.ExitCode(); code >= 0 {
		return code
	}
	if ws, ok := c.ProcessState.Sys().(interface {
		Signaled() bool
		Signal() syscall.Signal
	}); ok && ws.Signaled() {
		return 128 + int(ws.Signal())
	}
	return 1
}

func cmdInit(args []string) int {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	prompt := fs.Bool("prompt", false, "also put Spot's face in front of your prompt")
	shell := ""
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		shell, args = args[0], args[1:]
	}
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if shell == "" && fs.NArg() > 0 {
		shell = fs.Arg(0)
	}
	if shell == "" {
		fmt.Fprintln(os.Stderr, "spot init: which shell? zsh, bash, fish or pwsh. e.g. eval \"$(spot init zsh)\"")
		return 2
	}
	exe, err := os.Executable()
	if err != nil {
		exe = "spot"
	} else if r, err := filepath.EvalSymlinks(exe); err == nil && !strings.Contains(r, "/Cellar/") {
		exe = r // Homebrew's bin/spot symlink survives upgrades; the Cellar path doesn't
	}
	s, err := hook.Init(shell, exe, *prompt)
	if err != nil {
		return fail(err)
	}
	fmt.Print(s)
	return 0
}

func cmdPrompt(args []string) int {
	shell := shellFlag(args)
	e, err := newEnv()
	if err != nil {
		return 0 // never break a prompt
	}
	fmt.Print(react.Face(e.peek().MoodAt(e.now), e.cfg.Prompt, shell, e.profile))
	return 0
}

func shellFlag(args []string) string {
	for i, a := range args {
		if a == "--shell" && i+1 < len(args) {
			return args[i+1]
		}
		if v, ok := strings.CutPrefix(a, "--shell="); ok {
			return v
		}
	}
	return ""
}

// cmdEvent is called by the shell hook after every command. It must be
// quick and must never fail: it prints the prompt face and exits 0.
func cmdEvent(args []string) (code int) {
	defer func() {
		if recover() != nil {
			code = 0
		}
	}()
	var shell, line string
	exit, dur := -1, time.Duration(0)
	for i := 0; i < len(args); i++ {
		next := func() string {
			if i+1 < len(args) {
				i++
				return args[i]
			}
			return ""
		}
		switch args[i] {
		case "--shell":
			shell = next()
		case "--exit":
			exit, _ = strconv.Atoi(next())
		case "--dur":
			if f, err := strconv.ParseFloat(next(), 64); err == nil {
				dur = time.Duration(f * float64(time.Second))
			}
		case "--since":
			if f, err := strconv.ParseFloat(strings.ReplaceAll(next(), ",", "."), 64); err == nil && f > 1e9 {
				dur = time.Since(time.Unix(0, int64(f*1e9)))
			}
		case "--":
			line = strings.Join(args[i+1:], " ")
			i = len(args)
		}
	}
	e, err := newEnv()
	if err != nil {
		return 0
	}
	m := hook.NewMatcher(e.cfg.TestCommands, e.cfg.ExtraTestCommands, e.cfg.IgnoreCommands)
	if exit < 0 || !m.IsTest(line) || slices.Contains(interruptCodes, exit) {
		fmt.Print(react.Face(e.peek().MoodAt(e.now), e.cfg.Prompt, shell, e.profile))
		return 0
	}
	var r pet.Reaction
	st, err := e.update(func(st *pet.State) error {
		if exit == 0 {
			r = st.TestPassed(line, dur, e.now)
		} else {
			r = st.TestFailed(line, dur, e.now)
		}
		return nil
	})
	if err != nil {
		st = e.peek()
	} else {
		react.Play(e.opts(), r, st)
	}
	fmt.Print(react.Face(st.MoodAt(e.now), e.cfg.Prompt, shell, e.profile))
	return 0
}

type statusJSON struct {
	Name       string   `json:"name"`
	Mood       pet.Mood `json:"mood"`
	Food       int      `json:"food"`
	Joy        int      `json:"joy"`
	Energy     int      `json:"energy"`
	Bond       int      `json:"bond"`
	Treats     int      `json:"treats"`
	Streak     int      `json:"streak"`
	BestStreak int      `json:"bestStreak"`
	Runs       int      `json:"runs"`
	Passes     int      `json:"passes"`
	Tricks     []string `json:"tricks"`
	Badges     []string `json:"badges"`
	State      string   `json:"stateFile"`
}

func cmdStatus(args []string) int {
	e, err := newEnv()
	if err != nil {
		return fail(err)
	}
	st := e.peek()
	mood := st.MoodAt(e.now)
	if slices.Contains(args, "--json") {
		j := statusJSON{
			Name: st.Name, Mood: mood, Food: int(st.Food + .5), Joy: int(st.Joy + .5), Energy: int(st.Energy + .5),
			Bond: int(st.Bond + .5), Treats: st.Treats, Streak: st.Streak, BestStreak: st.BestStreak,
			Runs: st.Runs, Passes: st.Passes, Tricks: append([]string{}, st.Tricks...), State: e.st.Path(),
		}
		for _, b := range st.EarnedBadges() {
			j.Badges = append(j.Badges, b.ID)
		}
		if j.Badges == nil {
			j.Badges = []string{}
		}
		out, _ := json.MarshalIndent(j, "", "  ")
		fmt.Println(string(out))
		return 0
	}
	o := e.stdoutOpts()
	fmt.Fprintln(o.Out, strings.Join(statusCard(st, mood, o), "\n"))
	return 0
}

var moodWords = map[pet.Mood]string{
	pet.Asleep: "fast asleep", pet.Hungry: "hungry", pet.Sad: "a bit blue", pet.Proud: "proud of you",
	pet.Happy: "happy", pet.Content: "content", pet.Sleepy: "sleepy",
}

var moodAnim = map[pet.Mood]string{
	pet.Asleep: "mini-sleep", pet.Sleepy: "mini-sleep", pet.Sad: "mini-sad", pet.Proud: "mini-love",
	pet.Happy: "mini-wag", pet.Hungry: "mini-idle", pet.Content: "mini-idle",
}

func statusCard(st *pet.State, mood pet.Mood, o react.Opts) []string {
	bar := func(v float64) string {
		n := int(v/10 + .5)
		return strings.Repeat("█", n) + strings.Repeat("░", 10-n) + fmt.Sprintf(" %3.0f", v)
	}
	text := []string{
		st.Name + " is " + moodWords[mood],
		"",
		"food    " + bar(st.Food),
		"joy     " + bar(st.Joy),
		"energy  " + bar(st.Energy),
		"bond    " + bar(st.Bond),
		"",
		fmt.Sprintf("🦴 %d treats · streak %d · best %d", st.Treats, st.Streak, st.BestStreak),
	}
	if o.Profile <= colorprofile.ASCII || o.Width < 60 {
		return text
	}
	return react.Beside(art.Render(art.Get(moodAnim[mood]).Frames[0], o.Pal, o.Profile), text, 3)
}

func cmdFeed() int {
	return simple(func(st *pet.State, now time.Time) pet.Reaction { return st.Feed(now) })
}
func cmdPet() int {
	return simple(func(st *pet.State, now time.Time) pet.Reaction { return st.Pet(now) })
}
func simple(fn func(*pet.State, time.Time) pet.Reaction) int {
	e, err := newEnv()
	if err != nil {
		return fail(err)
	}
	var r pet.Reaction
	st, err := e.update(func(st *pet.State) error { r = fn(st, e.now); return nil })
	if err != nil {
		return fail(err)
	}
	o := e.stdoutOpts()
	r.Treats = 0
	react.Play(o, r, st)
	return 0
}

func cmdTrick(args []string) int {
	e, err := newEnv()
	if err != nil {
		return fail(err)
	}
	st := e.peek()
	if len(args) == 0 {
		fmt.Println(st.Name + "'s tricks:")
		for _, t := range pet.Tricks {
			if st.Knows(t.ID) {
				fmt.Printf("  ✦ %-12s spot trick %s\n", t.Name, t.ID)
			} else {
				fmt.Printf("  · %-12s learns it with %s\n", t.Name, t.How)
			}
		}
		return 0
	}
	t, ok := pet.TrickByID(args[0])
	if !ok {
		fmt.Fprintf(os.Stderr, "spot: there's no trick called %q (spot tricks lists them)\n", args[0])
		return 2
	}
	if !st.Knows(t.ID) {
		fmt.Printf("%s tilts his head. He'll learn %s with %s.\n", st.Name, t.Name, t.How)
		return 0
	}
	react.PlayBig(e.stdoutOpts(), t.Anim, 2500*time.Millisecond, st.Name+": "+t.Name+"!")
	return 0
}

func cmdBadges() int {
	e, err := newEnv()
	if err != nil {
		return fail(err)
	}
	st := e.peek()
	fmt.Printf("%d of %d badges\n", len(st.Badges), len(pet.Badges))
	for _, b := range pet.Badges {
		if at, ok := st.Badges[b.ID]; ok {
			fmt.Printf("  %s %-13s %s\n", b.Icon, b.Name, at.Format("2 Jan 2006"))
		} else {
			fmt.Printf("  ·  %-13s %s\n", b.Name, b.How)
		}
	}
	return 0
}

func cmdStats() int {
	e, err := newEnv()
	if err != nil {
		return fail(err)
	}
	st := e.peek()
	fmt.Printf("%s has been with you since %s\n\n", st.Name, st.Born.Format("2 January 2006"))
	fmt.Printf("  test runs     %d (%d green, %d red)\n", st.Runs, st.Passes, st.Fails)
	if st.Runs > 0 {
		fmt.Printf("  pass rate     %.0f%%\n", st.PassRate()*100)
	}
	fmt.Printf("  streak        %d (best %d)\n", st.Streak, st.BestStreak)
	fmt.Printf("  green days    %d in a row\n", st.GreenDays)
	fmt.Printf("  pets          %d\n", st.Pets)
	fmt.Printf("  fetch record  %d\n", st.FetchBest)
	if st.Last != nil {
		res := "green"
		if !st.Last.OK {
			res = "red"
		}
		fmt.Printf("  last run      %s, %s ago: %s\n", res, ago(e.now.Sub(st.Last.At)), st.Last.Cmd)
	}
	if n := len(st.Log); n > 0 {
		fmt.Println("\n  recently:")
		for _, m := range st.Log[max(0, n-6):] {
			fmt.Printf("    %s  %s\n", m.At.Format("Mon 15:04"), m.Text)
		}
	}
	return 0
}

func ago(d time.Duration) string {
	switch {
	case d < time.Minute:
		return "moments"
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 48*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%d days", int(d.Hours()/24))
}

func cmdRename(args []string) int {
	name := strings.TrimSpace(strings.Join(args, " "))
	if name == "" {
		fmt.Fprintln(os.Stderr, "spot rename: what should he be called?")
		return 2
	}
	e, err := newEnv()
	if err != nil {
		return fail(err)
	}
	if _, err := e.update(func(st *pet.State) error { st.Rename(name); return nil }); err != nil {
		return fail(err)
	}
	fmt.Printf("%s wags. That's him now.\n", name)
	return 0
}

func cmdReset(args []string) int {
	e, err := newEnv()
	if err != nil {
		return fail(err)
	}
	if !slices.Contains(args, "--yes") {
		if !term.IsTerminal(os.Stdin.Fd()) {
			fmt.Fprintln(os.Stderr, "spot reset: add --yes to start over without asking")
			return 2
		}
		fmt.Print("Start over with a new puppy? Spot forgets his tricks, badges and history. [y/N] ")
		ans, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		if a := strings.ToLower(strings.TrimSpace(ans)); a != "y" && a != "yes" {
			fmt.Println("Phew. Spot stays.")
			return 0
		}
	}
	old := e.peek().Name
	if err := e.st.Save(pet.New(old, e.now)); err != nil {
		return fail(err)
	}
	fmt.Printf("A new puppy called %s blinks at you.\n", old)
	return 0
}

func cmdShow(args []string) int {
	if len(args) == 0 {
		names := make([]string, 0, len(art.Anims))
		for n := range art.Anims {
			names = append(names, n)
		}
		sort.Strings(names)
		fmt.Println("spot show NAME, where NAME is one of:\n  " + strings.Join(names, " "))
		return 0
	}
	if _, ok := art.Anims[args[0]]; !ok {
		fmt.Fprintf(os.Stderr, "spot: no animation called %q (spot show lists them)\n", args[0])
		return 2
	}
	e, err := newEnv()
	if err != nil {
		return fail(err)
	}
	secs := 3.0
	if len(args) > 1 {
		if f, err := strconv.ParseFloat(args[1], 64); err == nil {
			secs = f
		}
	}
	o := e.stdoutOpts()
	if o.Mode != "full" {
		o.Mode = "short"
	}
	react.PlayBig(o, args[0], time.Duration(secs*float64(time.Second)), "")
	return 0
}

func cmdConfig(args []string) int {
	p, err := store.ConfigPath()
	if err != nil {
		return fail(err)
	}
	if len(args) > 0 && args[0] == "init" {
		if _, err := os.Stat(p); err == nil {
			fmt.Fprintf(os.Stderr, "spot: %s already exists; leaving it alone\n", p)
			return 1
		}
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return fail(err)
		}
		if err := os.WriteFile(p, []byte(store.Example), 0o644); err != nil {
			return fail(err)
		}
		fmt.Println("wrote", p)
		return 0
	}
	s, _ := store.Open()
	fmt.Println("config  ", p)
	fmt.Println("state   ", s.Path())
	if _, err := store.LoadConfig(); err != nil {
		fmt.Fprintln(os.Stderr, "spot: config has a problem:", err)
		return 1
	}
	return 0
}

func cmdTests() int {
	e, err := newEnv()
	if err != nil {
		return fail(err)
	}
	list := hook.Defaults
	if len(e.cfg.TestCommands) > 0 {
		list = e.cfg.TestCommands
	}
	w := io.Writer(os.Stdout)
	fmt.Fprintln(w, "Spot counts these as test commands (a * matches the rest of a word):")
	for _, p := range append(append([]string{}, list...), e.cfg.ExtraTestCommands...) {
		fmt.Fprintln(w, "  "+p)
	}
	if len(e.cfg.IgnoreCommands) > 0 {
		fmt.Fprintln(w, "\nbut never these:")
		for _, p := range e.cfg.IgnoreCommands {
			fmt.Fprintln(w, "  "+p)
		}
	}
	return 0
}
