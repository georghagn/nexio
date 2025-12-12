// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

package schedule

import (
	//"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Logger Interface: Damit der Scheduler nicht vom konkreten gsflog abhängt.
type Logger interface {
	Errorf(format string, args ...interface{})
	Infof(format string, args ...interface{})
}

// JobInfo ist ein DTO (Data Transfer Object) für die Außenwelt (read-only).
type JobInfo struct {
	ID        int64         `json:"id"`
	Interval  time.Duration `json:"interval"`
	NextRun   time.Time     `json:"next_run"`
	IsRunning bool          `json:"running"`
}

// JobID identifiziert einen laufenden Task eindeutig.
type JobID int64

// Job repräsentiert eine geplante Aufgabe.
type Job struct {
	ID       JobID
	Interval time.Duration // 0 wenn One-Shot
	NextRun  time.Time     // Wann läuft er das nächste Mal? (für Status-Abfragen)
	Fn       func()        // Die eigentliche Funktion

	quit    chan struct{} // Kanal zum Stoppen dieses spezifischen Jobs
	running bool          // Läuft er gerade aktiv?
}

// Scheduler verwaltet alle Jobs.
type Scheduler struct {
	jobs   map[JobID]*Job
	lastID int64
	mu     sync.RWMutex
	wg     sync.WaitGroup // Wartet auf laufende Jobs beim Shutdown

	logger Logger // Kann nil sein!
}

// New erstellt einen neuen Scheduler.
func New() *Scheduler {
	return &Scheduler{
		jobs: make(map[JobID]*Job),
	}
}

// SetLogger injiziert den Logger nachträglich.
func (s *Scheduler) SetLogger(l Logger) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logger = l
}

// --- Public API ---

// Every führt `task` alle `interval` Dauer aus. Gibt die JobID zurück.
func (s *Scheduler) Every(interval time.Duration, task func()) JobID {
	return s.addJob(interval, task, false)
}

// At führt `task` einmalig zum Zeitpunkt `t` aus.
func (s *Scheduler) At(t time.Time, task func()) JobID {
	// Zeit bis dahin berechnen
	duration := time.Until(t)
	if duration < 0 {
		duration = 0 // Sofort ausführen
	}
	return s.addJob(duration, task, true)
}

// Cancel stoppt einen spezifischen Job.
func (s *Scheduler) Cancel(id JobID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[id]
	if !exists {
		return fmt.Errorf("job %d not found", id)
	}

	// Signal zum Stoppen senden
	close(job.quit)
	delete(s.jobs, id)
	return nil
}

// WaitForShutdown wartet, bis alle Jobs beendet sind.
// Sollte aufgerufen werden, nachdem Cancel() für alle Jobs oder Kontext-Cancellation erfolgt ist.
// Hinweis: In dieser Implementierung stoppen wir Jobs manuell oder lassen sie auslaufen.
func (s *Scheduler) StopAll() {
	s.mu.Lock()
	// Wir kopieren die IDs, um Deadlocks zu vermeiden beim Iterieren + Löschen
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

// List gibt eine Kopie aller aktuellen Jobs zurück.
func (s *Scheduler) List() []JobInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]JobInfo, 0, len(s.jobs))
	for _, job := range s.jobs {
		list = append(list, JobInfo{
			ID:        int64(job.ID), // Cast auf int64 für JSON Freundlichkeit
			Interval:  job.Interval,
			NextRun:   job.NextRun,
			IsRunning: job.running,
		})
	}
	return list
}

// --- Interne Logik ---

func (s *Scheduler) addJob(duration time.Duration, task func(), oneShot bool) JobID {
	id := JobID(atomic.AddInt64(&s.lastID, 1))

	job := &Job{
		ID:       id,
		Interval: duration, // Bei OneShot ist das die Wartezeit
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

	// Initial setzen
	s.mu.Lock()
	job.NextRun = time.Now().Add(initialDelay)
	s.mu.Unlock()

	// Initialer Timer (Wartezeit bis zum ersten Start)
	timer := time.NewTimer(initialDelay)

	for {
		select {
		case <-job.quit:
			// Job wurde manuell abgebrochen
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return

		case <-timer.C:
			// 1. Task ausführen (mit Panic Protection)
			s.safeExecute(job)

			// 2. Wenn One-Shot -> Aufräumen und Ende
			if oneShot {
				s.mu.Lock()
				delete(s.jobs, job.ID)
				s.mu.Unlock()
				return
			}

			s.mu.Lock()
			job.NextRun = time.Now().Add(job.Interval)
			s.mu.Unlock()

			// 3. Wenn Intervall -> Timer neu setzen
			timer.Reset(job.Interval)
		}
	}
}

// safeExecute fängt Panics ab, damit der Scheduler nicht crasht
func (s *Scheduler) safeExecute(job *Job) {
	defer func() {
		if r := recover(); r != nil {
			if s.logger != nil {
				s.logger.Errorf("SCHEDULER PANIC in Job %d: %v", job.ID, r)
			} else {
				// Fallback, falls kein Logger gesetzt wurde
				fmt.Printf("SCHEDULER PANIC (No Logger): %v\n", r)
			}
		}
	}()

	job.Fn()
}
