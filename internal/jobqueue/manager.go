package jobqueue

import (
	"database/sql"
	"fmt"
	"time"
)

// JobFunc is the work performed by a job. It receives the job and a logger.
type JobFunc func(job *Job, log func(string)) error

// Enqueue creates and starts a job.
func (m *Manager) Enqueue(jobType string, fn JobFunc) *Job {
	now := time.Now().UTC()
	id := m.newID()
	j := &Job{
		ID:        id,
		Type:      jobType,
		Status:    StatusQueued,
		Progress:  0,
		Message:   "",
		CreatedAt: now,
		UpdatedAt: now,
		fn:        fn,
		cancel:    make(chan struct{}),
	}
	m.mu.Lock()
	m.jobs[id] = j
	m.mu.Unlock()

	// Persist queued.
	m.persist(j)

	go m.run(j, fn)
	return j
}

func (m *Manager) run(j *Job, fn JobFunc) {
	m.setStatus(j, StatusRunning, 0, "started")
	logf := func(line string) {
		j.Log(line)
		m.persistLog(j.ID, line)
		m.broadcast(j)
	}
	select {
	case <-j.cancel:
		m.setStatus(j, StatusCancelled, j.Progress, "cancelled")
		m.broadcast(j)
		return
	default:
	}
	err := fn(j, logf)
	if err != nil {
		m.setStatus(j, StatusFailed, j.Progress, err.Error())
	} else {
		m.setStatus(j, StatusSucceeded, 100, "done")
	}
	m.broadcast(j)
}

func (m *Manager) setStatus(j *Job, s Status, progress int, msg string) {
	j.mu.Lock()
	j.Status = s
	j.Progress = progress
	j.Message = msg
	j.UpdatedAt = time.Now().UTC()
	j.mu.Unlock()
	m.persist(j)
}

func (m *Manager) persist(j *Job) {
	if m.DB == nil {
		return
	}
	_, _ = m.DB.Exec(
		`INSERT INTO jobs (id,type,status,progress,message,created_at,updated_at)
		 VALUES (?,?,?,?,?,?,?)
		 ON CONFLICT(id) DO UPDATE SET status=excluded.status, progress=excluded.progress, message=excluded.message, updated_at=excluded.updated_at`,
		j.ID, j.Type, string(j.Status), j.Progress, j.Message,
		j.CreatedAt.Format(time.RFC3339), j.UpdatedAt.Format(time.RFC3339),
	)
}

func (m *Manager) persistLog(jobID, line string) {
	if m.DB == nil {
		return
	}
	_, _ = m.DB.Exec(
		`INSERT INTO job_logs (job_id,line,ts) VALUES (?,?,?)`,
		jobID, line, time.Now().UTC().Format(time.RFC3339),
	)
}

func (m *Manager) Get(id string) (*Job, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	j, ok := m.jobs[id]
	return j, ok
}

func (m *Manager) List() []*Job {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]*Job, 0, len(m.jobs))
	for _, j := range m.jobs {
		out = append(out, j)
	}
	return out
}

func (m *Manager) SetProgress(j *Job, progress int, msg string) {
	m.setStatus(j, j.Status, progress, msg)
	m.broadcast(j)
}

// Retry re-runs a failed/cancelled job using its stored function.
func (m *Manager) Retry(id string) bool {
	m.mu.Lock()
	j, ok := m.jobs[id]
	m.mu.Unlock()
	if !ok || j.fn == nil {
		return false
	}
	if j.Status != StatusFailed && j.Status != StatusCancelled {
		return false
	}
	// reset
	j.mu.Lock()
	j.Status = StatusQueued
	j.Progress = 0
	j.Message = ""
	j.logs = nil
	j.cancel = make(chan struct{})
	j.UpdatedAt = time.Now().UTC()
	j.mu.Unlock()
	m.persist(j)
	go m.run(j, j.fn)
	return true
}

// Cancel requests cancellation. JobFunc should check j.cancel channel.
func (m *Manager) Cancel(id string) bool {
	m.mu.Lock()
	j, ok := m.jobs[id]
	m.mu.Unlock()
	if !ok {
		return false
	}
	select {
	case <-j.cancel:
		// already closed
	default:
		j.Cancel()
	}
	return true
}

// Subscribe returns a channel receiving events.
func (m *Manager) Subscribe() chan Event {
	ch := make(chan Event, 64)
	m.subMu.Lock()
	m.subs[ch] = struct{}{}
	m.subMu.Unlock()
	return ch
}

func (m *Manager) Unsubscribe(ch chan Event) {
	m.subMu.Lock()
	delete(m.subs, ch)
	m.subMu.Unlock()
	close(ch)
}

func (m *Manager) broadcast(j *Job) {
	ev := Event{Type: "job", Job: j}
	m.subMu.Lock()
	defer m.subMu.Unlock()
	for ch := range m.subs {
		select {
		case ch <- ev:
		default:
		}
	}
}

// LoadFromDB hydrates in-memory jobs (status only) on startup.
func (m *Manager) LoadFromDB() error {
	if m.DB == nil {
		return nil
	}
	rows, err := m.DB.Query(`SELECT id,type,status,progress,message,created_at,updated_at FROM jobs`)
	if err != nil {
		return err
	}
	defer rows.Close()
	m.mu.Lock()
	defer m.mu.Unlock()
	for rows.Next() {
		j := &Job{}
		var status string
		var created, updated string
		if err := rows.Scan(&j.ID, &j.Type, &status, &j.Progress, &j.Message, &created, &updated); err != nil {
			return err
		}
		j.Status = Status(status)
		j.CreatedAt, _ = time.Parse(time.RFC3339, created)
		j.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
		j.cancel = make(chan struct{})
		m.jobs[j.ID] = j
	}
	return rows.Err()
}

var _ = sql.Open
var _ = fmt.Sprintf
