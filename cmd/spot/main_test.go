package main

import (
	"encoding/json"
	"io"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/Gitguy268/spot-cli/internal/hook"
	"github.com/Gitguy268/spot-cli/internal/store"
)

func sandbox(t testing.TB) *store.Store {
	t.Helper()
	d := t.TempDir()
	t.Setenv("SPOT_HOME", d)
	t.Setenv("SPOT_CONFIG", d)
	t.Setenv("SPOT_QUIET", "")
	t.Setenv("SPOT_REDUCED_MOTION", "")
	return &store.Store{Dir: d}
}

// capture runs fn with stdout redirected and returns what it printed.
func capture(t testing.TB, fn func()) string {
	t.Helper()
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w
	done := make(chan string)
	go func() { b, _ := io.ReadAll(r); done <- string(b) }()
	fn()
	w.Close()
	os.Stdout = old
	return <-done
}

func TestRunKeepsExitCode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses sh")
	}
	s := sandbox(t)
	for _, c := range []struct {
		args []string
		code int
	}{
		{[]string{"--", "true"}, 0},
		{[]string{"--", "sh", "-c", "exit 3"}, 3},
		{[]string{"-q", "--", "false"}, 1},
		{[]string{"sh", "-c", "kill -INT $$"}, 130},
		{[]string{"--", "no-such-command-spot"}, 127},
	} {
		if got := run(append([]string{"run"}, c.args...)); got != c.code {
			t.Errorf("spot run %v = %d, want %d", c.args, got, c.code)
		}
	}
	st, _ := s.Load(time.Now())
	if st.Runs != 3 || st.Passes != 1 {
		t.Fatalf("runs %d passes %d; Ctrl-C and missing commands shouldn't count", st.Runs, st.Passes)
	}
}

func TestEventRecordsOnlyTests(t *testing.T) {
	s := sandbox(t)
	face := capture(t, func() { run([]string{"_event", "--shell", "zsh", "--exit", "0", "--dur", "1.5", "--", "ls", "-la"}) })
	if face == "" {
		t.Fatal("no face printed")
	}
	capture(t, func() {
		run([]string{"_event", "--shell", "zsh", "--exit", "1", "--since", "1700000000.5", "--", "BREAK=1", "go", "test", "./..."})
	})
	capture(t, func() { run([]string{"_event", "--shell", "bash", "--exit", "0", "--", "go", "test", "./..."}) })
	capture(t, func() { run([]string{"_event", "--exit", "130", "--", "go", "test"}) })
	st, _ := s.Load(time.Now())
	if st.Runs != 2 || st.Passes != 1 || st.Fails != 1 {
		t.Fatalf("runs %d passes %d fails %d", st.Runs, st.Passes, st.Fails)
	}
	if _, ok := st.Badges["comeback"]; !ok {
		t.Fatal("BREAK=1 go test then go test should be a comeback")
	}
}

func TestEventNeverFails(t *testing.T) {
	sandbox(t)
	for _, args := range [][]string{{}, {"--exit"}, {"--exit", "x", "--", ""}, {"--since", "nope", "--"}} {
		if code := run(append([]string{"_event"}, args...)); code != 0 {
			t.Errorf("_event %v exited %d", args, code)
		}
	}
}

func TestStatusJSON(t *testing.T) {
	sandbox(t)
	out := capture(t, func() { run([]string{"status", "--json"}) })
	var j statusJSON
	if err := json.Unmarshal([]byte(out), &j); err != nil || j.Name != "Spot" || j.Badges == nil {
		t.Fatalf("%v: %s", err, out)
	}
}

func TestInitScripts(t *testing.T) {
	for _, sh := range hook.Shells {
		for _, prompt := range []string{"", "--prompt"} {
			args := []string{"init", sh}
			if prompt != "" {
				args = append(args, prompt)
			}
			out := capture(t, func() {
				if code := run(args); code != 0 {
					t.Errorf("init %s exited %d", sh, code)
				}
			})
			if !strings.Contains(out, "_event") || strings.Contains(out, "{{") {
				t.Errorf("init %s %s: bad script", sh, prompt)
			}
		}
	}
	if run([]string{"init", "tcsh"}) == 0 {
		t.Fatal("unknown shell should fail")
	}
}

func TestUnknownCommand(t *testing.T) {
	if run([]string{"fly"}) != 2 {
		t.Fatal("unknown command should exit 2")
	}
}

func TestRenameAndReset(t *testing.T) {
	s := sandbox(t)
	capture(t, func() { run([]string{"rename", "Spotty", "McSpot"}) })
	st, _ := s.Load(time.Now())
	if st.Name != "Spotty McSpot" {
		t.Fatalf("name %q", st.Name)
	}
	capture(t, func() { run([]string{"reset", "--yes"}) })
	st, _ = s.Load(time.Now())
	if st.Name != "Spotty McSpot" || st.Runs != 0 {
		t.Fatalf("reset should keep the name and clear the rest: %+v", st)
	}
}

// The prompt runs on every Enter, so it has to be quick.
func BenchmarkPrompt(b *testing.B) {
	sandbox(b)
	capture(b, func() {
		for range 3 {
			run([]string{"_event", "--exit", "0", "--", "go", "test"})
		}
		b.ResetTimer()
		for b.Loop() {
			run([]string{"prompt", "--shell", "zsh"})
		}
	})
}

func TestPromptIsFast(t *testing.T) {
	if testing.Short() || os.Getenv("CI") != "" {
		t.Skip("timing test")
	}
	r := testing.Benchmark(BenchmarkPrompt)
	if per := time.Duration(r.NsPerOp()); per > 5*time.Millisecond {
		t.Fatalf("spot prompt takes %s, budget 5ms", per)
	}
}
