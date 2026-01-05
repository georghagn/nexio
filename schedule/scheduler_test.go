package schedule

import (
	"testing"
	"time"
)

func TestScheduler_Every_Cancel(t *testing.T) {
	sched := New()
	counter := 0

	// Job: Count up
	id := sched.Every(10*time.Millisecond, func() {
		counter++
	})

	// Let it run briefly (it should fire about 5-10 times)
	time.Sleep(100 * time.Millisecond)

	// Stop
	if err := sched.Cancel(id); err != nil {
		t.Errorf("Cancel failed: %v", err)
	}

	// Remember how high it is
	valAfterCancel := counter

	// Wait... he should NOT count up any further.
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

	// Check if it has been deleted from the map (cleanup)
	sched.mu.Lock()
	count := len(sched.jobs)
	sched.mu.Unlock()

	if count != 0 {
		t.Errorf("OneShot job was not removed from map. Count: %d", count)
	}
}

func TestScheduler_PanicRecovery(t *testing.T) {
	sched := New()

	// This job is deliberately crashing.
	sched.Every(10*time.Millisecond, func() {
		panic("Boom!")
	})

	// If the scheduler didn't have a recovery function, the test would crash here.
	time.Sleep(50 * time.Millisecond)

	sched.StopAll()
	// If we get here, recover() has worked.
}
