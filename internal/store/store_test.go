package store

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/Gitguy268/spot-cli/internal/pet"
)

var now = time.Date(2026, 11, 18, 12, 0, 0, 0, time.Local)

func TestMissingFileIsNewPuppy(t *testing.T) {
	s := &Store{Dir: t.TempDir()}
	st, err := s.Load(now)
	if err != nil || st.Name != "Spot" || st.Version != pet.Version {
		t.Fatalf("got %+v, %v", st, err)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	s := &Store{Dir: filepath.Join(t.TempDir(), "deep", "dir")}
	st := pet.New("Spot", now)
	st.TestPassed("go test ./...", time.Second, now)
	if err := s.Save(st); err != nil {
		t.Fatal(err)
	}
	got, err := s.Load(now)
	if err != nil {
		t.Fatal(err)
	}
	if got.Passes != 1 || got.Treats != st.Treats || got.Last.Cmd != "go test ./..." {
		t.Fatalf("round trip lost data: %+v", got)
	}
	left, _ := filepath.Glob(filepath.Join(s.Dir, ".state-*"))
	if len(left) != 0 {
		t.Fatalf("temp files left behind: %v", left)
	}
}

func TestCorruptFileIsMovedAside(t *testing.T) {
	dir := t.TempDir()
	s := &Store{Dir: dir}
	os.WriteFile(s.Path(), []byte("{nope"), 0o644)
	st, err := s.Load(now)
	if !errors.Is(err, ErrCorrupt) || st == nil || st.Name != "Spot" {
		t.Fatalf("got %+v, %v", st, err)
	}
	bad, _ := filepath.Glob(filepath.Join(dir, "state.json.bad-*"))
	if len(bad) != 1 {
		t.Fatalf("corrupt file not kept: %v", bad)
	}
}

func TestFutureVersionIsKept(t *testing.T) {
	s := &Store{Dir: t.TempDir()}
	os.WriteFile(s.Path(), []byte(`{"version": 99}`), 0o644)
	if _, err := s.Load(now); !errors.Is(err, ErrCorrupt) {
		t.Fatalf("want ErrCorrupt, got %v", err)
	}
}

func TestConcurrentUpdatesDontLoseRuns(t *testing.T) {
	s := &Store{Dir: t.TempDir()}
	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := s.Update(now, func(st *pet.State) error {
				st.TestPassed("go test", 0, now)
				return nil
			}); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	st, err := s.Load(now)
	if err != nil || st.Runs != 20 {
		t.Fatalf("runs %d, err %v", st.Runs, err)
	}
}

func TestStaleLockIsTakenOver(t *testing.T) {
	s := &Store{Dir: t.TempDir()}
	lock := filepath.Join(s.Dir, "state.lock")
	os.WriteFile(lock, nil, 0o644)
	old := time.Now().Add(-time.Minute)
	os.Chtimes(lock, old, old)
	if _, err := s.Update(now, func(*pet.State) error { return nil }); err != nil {
		t.Fatal(err)
	}
}

func TestConfigEnvOverrides(t *testing.T) {
	t.Setenv("SPOT_CONFIG", t.TempDir())
	t.Setenv("SPOT_QUIET", "")
	t.Setenv("SPOT_REDUCED_MOTION", "1")
	c, err := LoadConfig()
	if err != nil || c.Reactions != "short" || c.Prompt != "face" {
		t.Fatalf("got %+v, %v", c, err)
	}
	t.Setenv("SPOT_QUIET", "1")
	if c, _ := LoadConfig(); c.Reactions != "off" {
		t.Fatalf("SPOT_QUIET: %+v", c)
	}
}

func TestConfigFile(t *testing.T) {
	d := t.TempDir()
	t.Setenv("SPOT_CONFIG", d)
	t.Setenv("SPOT_QUIET", "")
	t.Setenv("SPOT_REDUCED_MOTION", "")
	p := filepath.Join(d, "config.toml")
	os.WriteFile(p, []byte(Example), 0o644)
	if c, err := LoadConfig(); err != nil || c.Reactions != "full" {
		t.Fatalf("example config: %+v, %v", c, err)
	}
	os.WriteFile(p, []byte("reactions = [oops"), 0o644)
	if c, err := LoadConfig(); err == nil || c.Reactions != "full" {
		t.Fatalf("broken config should error but keep defaults: %+v, %v", c, err)
	}
	os.WriteFile(p, []byte("reactions = \"off\"\nextra_test_commands = [\"task test*\"]\n"), 0o644)
	c, err := LoadConfig()
	if err != nil || c.Reactions != "off" || len(c.ExtraTestCommands) != 1 {
		t.Fatalf("got %+v, %v", c, err)
	}
}
