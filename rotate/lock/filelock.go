// NOT used yet. Reminder for later development
package lock

import (
	"errors"
	"os"
	"strconv"
	"time"
)

var ErrLockTimeout = errors.New("could not acquire lock: timeout")

type FileLock struct {
	Path    string
	Timeout time.Duration
	Expiry  time.Duration // Zeit, nach der ein Lock als "verwaist" gilt
}

// Acquire versucht den Lock zu erhalten
func (l *FileLock) Acquire() error {
	deadline := time.Now().Add(l.Timeout)

	for {
		// 1. Versuch: Atomares Erstellen
		f, err := os.OpenFile(l.Path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if err == nil {
			f.WriteString(strconv.Itoa(os.Getpid()))
			f.Close()
			return nil
		}

		// 2. Falls Datei existiert: Prüfen auf "Stale Lock" (Absturz-Schutz)
		if os.IsExist(err) && l.Expiry > 0 {
			if info, err := os.Stat(l.Path); err == nil {
				if time.Since(info.ModTime()) > l.Expiry {
					// Lock ist zu alt -> Löschen und neu versuchen
					os.Remove(l.Path)
					continue
				}
			}
		}

		// 3. Timeout Check
		if time.Now().After(deadline) {
			return ErrLockTimeout
		}

		// Kurze Pause vor dem nächsten Versuch
		time.Sleep(50 * time.Millisecond)
	}
}

func (l *FileLock) Release() error {
	return os.Remove(l.Path)
}

/*
	func (l *FileLock) Lock(path string) error {
		l.Path = path // Falls der Pfad erst hier übergeben wird
		return l.Acquire()
	}
*/
func (l *FileLock) Lock() error {
	return l.Acquire()
}

func (l *FileLock) Unlock() error {
	return l.Release()
}
