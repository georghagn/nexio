package lockWriter

import (
	"os"
	"testing"
	"time"

	"github.com/georghagn/nexio/rotate/lock"
)

func TestFileLock_AcquireAndRelease(t *testing.T) {
	lockPath := "test.lock"
	defer os.Remove(lockPath)

	l := &lock.FileLock{
		Path:    lockPath,
		Timeout: 500 * time.Millisecond,
	}

	// 1. Lock erfolgreich holen
	err := l.Acquire()
	if err != nil {
		t.Fatalf("Expected to acquire lock, got: %v", err)
	}

	// Prüfen, ob Datei existiert
	if _, err := os.Stat(lockPath); os.IsNotExist(err) {
		t.Fatal("Lock file should exist but doesn't")
	}

	// 2. Lock wieder freigeben
	err = l.Release()
	if err != nil {
		t.Fatalf("Expected to release lock, got: %v", err)
	}

	// Prüfen, ob Datei weg ist
	if _, err := os.Stat(lockPath); err == nil {
		t.Fatal("Lock file should be deleted but still exists")
	}
}

func TestFileLock_Conflict(t *testing.T) {
	lockPath := "conflict.lock"
	defer os.Remove(lockPath)

	l1 := &lock.FileLock{Path: lockPath, Timeout: 100 * time.Millisecond}
	l2 := &lock.FileLock{Path: lockPath, Timeout: 100 * time.Millisecond}

	// L1 holt den Lock
	if err := l1.Acquire(); err != nil {
		t.Fatal(err)
	}

	// L2 versucht es und sollte ins Timeout laufen
	err := l2.Acquire()
	if err != lock.ErrLockTimeout {
		t.Errorf("Expected ErrLockTimeout, got: %v", err)
	}

	l1.Release()
}

func TestFileLock_StaleLock(t *testing.T) {
	lockPath := "stale.lock"
	defer os.Remove(lockPath)

	// 1. Simuliere einen abgestürzten Prozess: Datei manuell erstellen
	err := os.WriteFile(lockPath, []byte("9999"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Zeit in die Vergangenheit setzen (Simuliere Alter von 1 Stunde)
	oldTime := time.Now().Add(-1 * time.Hour)
	os.Chtimes(lockPath, oldTime, oldTime)

	l := &lock.FileLock{
		Path:    lockPath,
		Timeout: 100 * time.Millisecond,
		Expiry:  1 * time.Minute, // Lock gilt nur 1 Minute
	}

	// Sollte den alten Lock "stehlen" (löschen und neu anlegen)
	err = l.Acquire()
	if err != nil {
		t.Fatalf("Should have acquired stale lock, got: %v", err)
	}
	l.Release()
}

// chaos test
func TestLockWriter_FileStolenBetweenWrites(t *testing.T) {
	filename := "stolen_test.log"
	defer os.Remove(filename)
	defer os.Remove(filename + ".LOCK")

	lw, err := NewLockingWriter(filename)
	if err != nil {
		t.Fatalf("Failed to create LockWriter: %v", err)
	}

	// 1. Erster Schreibvorgang (Datei wird erstellt)
	msg1 := "First entry\n"
	_, err = lw.Write([]byte(msg1))
	if err != nil {
		t.Fatalf("First write failed: %v", err)
	}

	// 2. CHAOS: Wir löschen die Datei einfach weg (wie ein externer Prozess)
	err = os.Remove(filename)
	if err != nil {
		t.Fatalf("Failed to delete file: %v", err)
	}

	// 3. Zweiter Schreibvorgang
	// Dank deines Designs MUSS der LockWriter die Datei einfach neu erstellen.
	msg2 := "Second entry after theft\n"
	_, err = lw.Write([]byte(msg2))
	if err != nil {
		t.Fatalf("Second write failed: %v", err)
	}

	// 4. Verifikation
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Es darf NUR die zweite Nachricht drinstehen
	if string(content) != msg2 {
		t.Errorf("Content mismatch.\nExpected: %q\nGot: %q", msg2, string(content))
	}
}
