package gsflog

import (
	"os"
	"path/filepath"
	"testing"

	reopen "github.com/georghagn/gsf-suite/pkg/gsflog/writer/reopen"
)

func TestReopenableWriter(t *testing.T) {
	// Setup: Temporary Directory
	dir := t.TempDir()
	logPath := filepath.Join(dir, "service.log")

	// 1. Start Writer
	w, err := reopen.NewReopenableWriter(logPath)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer w.Close()

	// 2. Write first line
	w.Write([]byte("Line 1\n"))

	// 3. Simulation: Logrotate (move file)
	// The operating system is still keeping the file handle for "w" open on the old inode!
	backupPath := filepath.Join(dir, "service.log.bak")
	if err := os.Rename(logPath, backupPath); err != nil {
		t.Fatalf("Rename failed: %v", err)
	}

	// 4. Write WITHOUT Reopen (It ends up in the .bak file because the handle is still old.)
	w.Write([]byte("Line 2 (Old Handle)\n"))

	// 5. Reopen ausführen
	if err := w.Reopen(); err != nil {
		t.Fatalf("Reopen failed: %v", err)
	}

	// 6. Write WITH Reopen (It ends up in the new service.log)
	w.Write([]byte("Line 3 (New Handle)\n"))

	// --- Examination ---

	// A. Check backup file
	contentBak, _ := os.ReadFile(backupPath)
	strBak := string(contentBak)
	if strBak != "Line 1\nLine 2 (Old Handle)\n" {
		t.Errorf("Backup file content wrong. Got: %q", strBak)
	}

	// B. Check new file
	contentNew, _ := os.ReadFile(logPath)
	strNew := string(contentNew)
	if strNew != "Line 3 (New Handle)\n" {
		t.Errorf("New file content wrong. Got: %q", strNew)
	}
}
