package tasks

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

type Status string

const (
	StatusSuccess Status = "success"
	StatusFailed  Status = "failed"
)

type Task struct {
	ID        string
	RequestID string
	Status    Status
	Result    string
	Model     string
	Error     string
	LatencyMS int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Store struct {
	mu    sync.RWMutex
	tasks map[string]Task
}

func NewStore() *Store {
	return &Store{
		tasks: make(map[string]Task),
	}
}

func NewID(prefix string) string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
	}
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(b[:]))
}

func (s *Store) Save(task Task) Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	if task.CreatedAt.IsZero() {
		task.CreatedAt = now
	}
	task.UpdatedAt = now
	s.tasks[task.ID] = task
	return task
}

func (s *Store) Get(id string) (Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, ok := s.tasks[id]
	return task, ok
}
