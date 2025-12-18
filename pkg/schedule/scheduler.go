// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

package schedule

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Logger Interface: So that the scheduler does not depend on the specific gsflog.
type Logger interface {
	Error(format string, args ...interface{})
	Info(format string, args ...interface{})
}

// JobInfo is a DTO (Data Transfer Object) for the outside world (read-only).
type JobInfo struct {
	ID        int64         `json:"id"`
	Interval  time.Duration `json:"interval"`
	NextRun   time.Time     `json:"next_run"`
	IsRunning bool          `json:"running"`
}

// JobID uniquely identifies a running task.
type JobID int64

// A job represents a planned task..
type Job struct {
	ID       JobID
	Interval time.Duration // 0 when One-Shot
	NextRun  time.Time     // When will it run next? (for status queries)
	Fn       func()        // The actual function

	quit    chan struct{} // Channel to stop this specific job
	running bool          // Is it currently running?
}

// The scheduler manages all jobs.
type Scheduler struct {
	jobs   map[JobID]*Job
	lastID int64
	mu     sync.RWMutex
	wg     sync.WaitGroup // Wait for ongoing jobs during the shutdown

	logger Logger // Can be nil!
}

// New creates a new scheduler.
func New() *Scheduler {
	return &Scheduler{
		jobs: make(map[JobID]*Job),
	}
}

// SetLogger injects the logger afterwards.
func (s *Scheduler) SetLogger(l Logger) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logger = l
}

// --- Public API ---

// Every executes `task` every `interval` duration. Returns the JobID.
func (s *Scheduler) Every(interval time.Duration, task func()) JobID {
	return s.addJob(interval, task, false)
}

// At executes `task` once at time `t`.
func (s *Scheduler) At(t time.Time, task func()) JobID {
	// Calculate the time until then
	duration := time.Until(t)
	if duration < 0 {
		duration = 0 // eecute immediately
	}
	return s.addJob(duration, task, true)
}

// Cancel stops a specific job.
func (s *Scheduler) Cancel(id JobID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[id]
	if !exists {
		return fmt.Errorf("job %d not found", id)
	}

	// Send a signal to stop
	close(job.quit)
	delete(s.jobs, id)
	return nil
}

// WaitForShutdown waits until all jobs have finished.
// Should be called after Cancel() has been used for all jobs or after context cancellation.
// Note: In this implementation, we stop jobs manually or let them expire.
func (s *Scheduler) StopAll() {
	s.mu.Lock()
	// We copy the IDs to avoid deadlocks when iterating and deleting.
	ids := make([]JobID, 0, len(s.jobs))
	for id := range s.jobs {
		ids = append(ids, id)
	}
	s.mu.Unlock()

	for _, id := range ids {
		s.Cancel(id)
	}

	s.wg.Wait()
}

// List returns a copy of all current jobs.
func (s *Scheduler) List() []JobInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]JobInfo, 0, len(s.jobs))
	for _, job := range s.jobs {
		list = append(list, JobInfo{
			ID:        int64(job.ID), // Cast to int64 for JSON friendliness
			Interval:  job.Interval,
			NextRun:   job.NextRun,
			IsRunning: job.running,
		})
	}
	return list
}

// --- Interne Logic ---

func (s *Scheduler) addJob(duration time.Duration, task func(), oneShot bool) JobID {
	id := JobID(atomic.AddInt64(&s.lastID, 1))

	job := &Job{
		ID:       id,
		Interval: duration, // In OneShot, that's the waiting time.
		Fn:       task,
		quit:     make(chan struct{}),
		running:  true,
	}

	s.mu.Lock()
	s.jobs[id] = job
	s.mu.Unlock()

	s.wg.Add(1)
	go s.run(job, duration, oneShot)

	return id
}

func (s *Scheduler) run(job *Job, initialDelay time.Duration, oneShot bool) {
	defer s.wg.Done()

	// set initial
	s.mu.Lock()
	job.NextRun = time.Now().Add(initialDelay)
	s.mu.Unlock()

	// Initial timer (waiting time until first start)
	timer := time.NewTimer(initialDelay)

	for {
		select {
		case <-job.quit:
			// The job was manually cancelled.
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return

		case <-timer.C:
			// 1. Run task (with panic protection)
			s.safeExecute(job)

			// 2. If it's a one-shot -> clean up and end it
			if oneShot {
				s.mu.Lock()
				delete(s.jobs, job.ID)
				s.mu.Unlock()
				return
			}

			s.mu.Lock()
			job.NextRun = time.Now().Add(job.Interval)
			s.mu.Unlock()

			// 3.If interval -> reset timer
			timer.Reset(job.Interval)
		}
	}
}

// safeExecute catches panics to prevent the scheduler from crashing.
func (s *Scheduler) safeExecute(job *Job) {
	defer func() {
		if r := recover(); r != nil {
			msg := fmt.Sprintf("SCHEDULER PANIC in Job %d: %v", job.ID, r)
			if s.logger != nil {
				s.logger.Error(msg)
			} else {
				//Fallback if no logger was set
				fmt.Println(msg)
			}
		}
	}()

	job.Fn()
}
