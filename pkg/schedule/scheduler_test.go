package schedule

import (
	"testing"
	"time"
)

func TestScheduler_Every_Cancel(t *testing.T) {
	sched := New()
	counter := 0

	// Job: Zähle hoch
	id := sched.Every(10*time.Millisecond, func() {
		counter++
	})

	// Lass ihn kurz laufen (sollte ca 5-10 mal feuern)
	time.Sleep(100 * time.Millisecond)

	// Stoppen
	if err := sched.Cancel(id); err != nil {
		t.Errorf("Cancel failed: %v", err)
	}

	// Merken wie hoch er ist
	valAfterCancel := counter

	// Warten... er sollte NICHT weiter hochzählen
	time.Sleep(50 * time.Millisecond)

	if counter != valAfterCancel {
		t.Errorf("Job continued running after cancel! Got %d, expected %d", counter, valAfterCancel)
	}
}

func TestScheduler_OneShot(t *testing.T) {
	sched := New()
	done := make(chan bool)

	// One Shot in 50ms
	sched.At(time.Now().Add(50*time.Millisecond), func() {
		done <- true
	})

	select {
	case <-done:
		// Success
	case <-time.After(200 * time.Millisecond):
		t.Fatal("OneShot job did not fire in time")
	}

	// Prüfen ob er aus der Map gelöscht wurde (Cleanup)
	sched.mu.Lock()
	count := len(sched.jobs)
	sched.mu.Unlock()

	if count != 0 {
		t.Errorf("OneShot job was not removed from map. Count: %d", count)
	}
}

func TestScheduler_PanicRecovery(t *testing.T) {
	sched := New()

	// Dieser Job stürzt absichtlich ab
	sched.Every(10*time.Millisecond, func() {
		panic("Boom!")
	})

	// Wenn der Scheduler keine Recovery hätte, würde der Test hier crashen.
	time.Sleep(50 * time.Millisecond)

	sched.StopAll()
	// Wenn wir hier ankommen, hat recover() funktioniert
}
