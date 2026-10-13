package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"syscall"
)

const (
	stateDir      = "agenthook"
	stateDirMode  = 0o750
	stateFileMode = 0o600
)

var unsafeFileNameCharacters = regexp.MustCompile(`[^A-Za-z0-9_-]`)

// Store keeps one session's state in a file under a checkout's git dir, so linked worktrees and parallel sessions
// never share it.
type Store struct {
	path string
}

func Open(gitDir, sessionID string) Store {
	name := unsafeFileNameCharacters.ReplaceAllString(sessionID, "_") + ".json"

	return Store{path: filepath.Join(gitDir, stateDir, name)}
}

func (s Store) Load() (State, error) {
	content, err := os.ReadFile(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		return emptyState(), nil
	}

	if err != nil {
		return emptyState(), fmt.Errorf("read session state: %w", err)
	}

	state := emptyState()

	err = json.Unmarshal(content, &state)
	if err != nil {
		return emptyState(), fmt.Errorf("decode session state %s: %w", s.path, err)
	}

	return state, nil
}

// Update applies change under a lock, because the agent can run several tools, and so several hooks, at once.
func (s Store) Update(change func(*State)) error {
	unlock, err := s.lock()
	if err != nil {
		return err
	}
	defer unlock()

	state, err := s.Load()
	if err != nil {
		return err
	}

	change(&state)

	return s.save(state)
}

func (s Store) RecordEdit(path string) error {
	return s.Update(func(state *State) { state.EditedPaths = appendNew(state.EditedPaths, path) })
}

// Seen treats unreadable state as unseen, so the comment guard denies rather than asks.
func (s Store) Seen(comment string) bool {
	state, err := s.Load()

	return err == nil && slices.Contains(state.SeenComments, comment)
}

func (s Store) Remember(comments []string) error {
	return s.Update(func(state *State) { state.SeenComments = appendNew(state.SeenComments, comments...) })
}

func (s Store) lock() (func(), error) {
	err := os.MkdirAll(filepath.Dir(s.path), stateDirMode)
	if err != nil {
		return nil, fmt.Errorf("create session state directory: %w", err)
	}

	file, err := os.OpenFile(s.path+".lock", os.O_CREATE|os.O_RDWR, stateFileMode)
	if err != nil {
		return nil, fmt.Errorf("open session state lock: %w", err)
	}

	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX)
	if err != nil {
		_ = file.Close()

		return nil, fmt.Errorf("lock session state: %w", err)
	}

	return func() { _ = file.Close() }, nil
}

// save writes through a rename, so a hook reading without the lock never sees half a file.
func (s Store) save(state State) error {
	content, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("encode session state: %w", err)
	}

	temporary := s.path + ".tmp"

	err = os.WriteFile(temporary, content, stateFileMode)
	if err != nil {
		return fmt.Errorf("write session state: %w", err)
	}

	err = os.Rename(temporary, s.path)
	if err != nil {
		return fmt.Errorf("replace session state: %w", err)
	}

	return nil
}
