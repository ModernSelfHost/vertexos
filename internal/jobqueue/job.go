package jobqueue

import (
	"database/sql"
	"fmt"
	"sync"
	"time"
)

type Status string

const (
	StatusQueued    Status = "queued"
	StatusRunning   Status = "running"
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
)

type Job struct {
	ID        string
	Type      string
	Status    Status
	Progress  int
	Message   string
	CreatedAt time.Time
	UpdatedAt time.Time

	fn     JobFunc
	cancel chan struct{}
	logs   []string
	mu     sync.Mutex
}

func (j *Job) Log(line string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.logs = append(j.logs, line)
}

func (j *Job) Logs() []string {
	j.mu.Lock()
	defer j.mu.Unlock()
	out := make([]string, len(j.logs))
	copy(out, j.logs)
	return out
}

func (j *Job) Cancel() { close(j.cancel) }

type Manager struct {
	DB    *sql.DB
	mu    sync.Mutex
	jobs  map[string]*Job
	subs  map[chan Event]struct{}
	subMu sync.Mutex
}

type Event struct {
	Type string `json:"type"`
	Job  *Job   `json:"job,omitempty"`
	Data any    `json:"data,omitempty"`
}

func NewManager(db *sql.DB) *Manager {
	return &Manager{DB: db, jobs: map[string]*Job{}, subs: map[chan Event]struct{}{}}
}

func (m *Manager) newID() string {
	return fmt.Sprintf("job-%d", time.Now().UnixNano())
}
