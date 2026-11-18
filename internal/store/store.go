// Package store keeps Spot's state on disk and reads the config file.
//
// State lives in $SPOT_HOME, or $XDG_STATE_HOME/spot, or ~/.local/state/spot
// (%LOCALAPPDATA%\spot on Windows). Config lives in $SPOT_CONFIG, or
// $XDG_CONFIG_HOME/spot, or ~/.config/spot (%APPDATA%\spot on Windows).
package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/Gitguy268/spot-cli/internal/pet"
)

func dir(override, xdg, win, home string) (string, error) {
	if d := os.Getenv(override); d != "" {
		return d, nil
	}
	if d := os.Getenv(xdg); d != "" {
		return filepath.Join(d, "spot"), nil
	}
	if runtime.GOOS == "windows" {
		if d := os.Getenv(win); d != "" {
			return filepath.Join(d, "spot"), nil
		}
	}
	h, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(h, home, "spot"), nil
}

// StateDir is where state.json lives.
func StateDir() (string, error) {
	return dir("SPOT_HOME", "XDG_STATE_HOME", "LOCALAPPDATA", filepath.Join(".local", "state"))
}

// ConfigDir is where config.toml lives.
func ConfigDir() (string, error) {
	return dir("SPOT_CONFIG", "XDG_CONFIG_HOME", "APPDATA", ".config")
}

// Store reads and writes one state file.
type Store struct {
	Dir string
}

// Open returns the store in the default place.
func Open() (*Store, error) {
	d, err := StateDir()
	if err != nil {
		return nil, err
	}
	return &Store{Dir: d}, nil
}

// Path is the state file.
func (s *Store) Path() string { return filepath.Join(s.Dir, "state.json") }

// ErrCorrupt is returned with a fresh puppy when the old file could not be read.
var ErrCorrupt = errors.New("state file was unreadable")

// Load reads the state. A missing file gives a new puppy and no error. A
// corrupt one is moved aside to state.json.bad-<time>, and Load returns a new
// puppy together with an error wrapping ErrCorrupt.
func (s *Store) Load(now time.Time) (*pet.State, error) {
	b, err := os.ReadFile(s.Path())
	if errors.Is(err, os.ErrNotExist) {
		return pet.New("", now), nil
	}
	if err != nil {
		return nil, err
	}
	var st pet.State
	if err := json.Unmarshal(b, &st); err != nil || st.Version < 1 || st.Version > pet.Version {
		bad := s.Path() + ".bad-" + now.Format("20060102-150405")
		_ = os.Rename(s.Path(), bad)
		why := "not valid JSON"
		if err == nil {
			why = "unknown version " + strconv.Itoa(st.Version)
		}
		return pet.New("", now), fmt.Errorf("%w (%s), kept as %s", ErrCorrupt, why, bad)
	}
	if st.LastResult == nil {
		st.LastResult = map[string]bool{}
	}
	if st.Badges == nil {
		st.Badges = map[string]time.Time{}
	}
	return &st, nil
}

// Save writes the state atomically: a temp file in the same folder, then a rename.
func (s *Store) Save(st *pet.State) error {
	if err := os.MkdirAll(s.Dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(s.Dir, ".state-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(append(b, '\n')); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), s.Path())
}

// Update loads, calls fn, and saves, holding a lock file so two shells that
// finish a command at the same moment don't lose each other's changes. A
// lock older than 5 seconds is treated as left behind and taken over. A
// corrupt-state error from Load is returned after a successful save.
func (s *Store) Update(now time.Time, fn func(*pet.State) error) (*pet.State, error) {
	unlock, err := s.lock()
	if err != nil {
		return nil, err
	}
	defer unlock()
	st, loadErr := s.Load(now)
	if st == nil {
		return nil, loadErr
	}
	if err := fn(st); err != nil {
		return st, err
	}
	if err := s.Save(st); err != nil {
		return st, err
	}
	return st, loadErr
}

func (s *Store) lock() (func(), error) {
	if err := os.MkdirAll(s.Dir, 0o755); err != nil {
		return nil, err
	}
	p := filepath.Join(s.Dir, "state.lock")
	deadline := time.Now().Add(3 * time.Second)
	for {
		f, err := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			f.Close()
			return func() { os.Remove(p) }, nil
		}
		if !errors.Is(err, os.ErrExist) && !os.IsPermission(err) {
			return nil, err
		}
		if fi, serr := os.Stat(p); serr == nil && time.Since(fi.ModTime()) > 5*time.Second {
			os.Remove(p)
			continue
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("state is locked by another spot (%s)", p)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
