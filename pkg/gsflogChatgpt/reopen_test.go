package gsflog

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReopenableWriter(t *testing.T) {
	// Setup: Temporäres Verzeichnis
	dir := t.TempDir()
	logPath := filepath.Join(dir, "service.log")

	// 1. Writer starten
	w, err := NewReopenableWriter(logPath)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer w.Close()

	// 2. Erste Zeile schreiben
	w.Write([]byte("Line 1\n"))

	// 3. Simulation: Logrotate (Datei verschieben)
	// Das Betriebssystem hält das Filehandle für "w" noch offen auf die alte Inode!
	backupPath := filepath.Join(dir, "service.log.bak")
	if err := os.Rename(logPath, backupPath); err != nil {
		t.Fatalf("Rename failed: %v", err)
	}

	// 4. Schreiben OHNE Reopen (landet in der .bak Datei, weil Handle noch alt ist)
	w.Write([]byte("Line 2 (Old Handle)\n"))

	// 5. Reopen ausführen
	if err := w.Reopen(); err != nil {
		t.Fatalf("Reopen failed: %v", err)
	}

	// 6. Schreiben MIT Reopen (landet in der neuen service.log)
	w.Write([]byte("Line 3 (New Handle)\n"))

	// --- Überprüfung ---

	// A. Backup Datei prüfen
	contentBak, _ := os.ReadFile(backupPath)
	strBak := string(contentBak)
	if strBak != "Line 1\nLine 2 (Old Handle)\n" {
		t.Errorf("Backup file content wrong. Got: %q", strBak)
	}

	// B. Neue Datei prüfen
	contentNew, _ := os.ReadFile(logPath)
	strNew := string(contentNew)
	if strNew != "Line 3 (New Handle)\n" {
		t.Errorf("New file content wrong. Got: %q", strNew)
	}
}
