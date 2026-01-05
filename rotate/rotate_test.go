package rotate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRotator_Integration_Gzip(t *testing.T) {
	// 1. Temporary directory for this test (will be automatically deleted)
	dir := t.TempDir()
	logPath := filepath.Join(dir, "app.log")

	// 2. Setup: Rotate after 10 bytes, using Gzip compression
	w := New(logPath,
		&SizePolicy{MaxBytes: 10},
		&GzipCompression{},
		&KeepAll{}, // We will review retention separately.
	)
	defer w.Close()

	// 3. Writing (below the limit)
	//    "Hello" = 5 Bytes
	if _, err := w.Write([]byte("Hello")); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Check if file exists
	checkFileExists(t, logPath)

	// 4. Limit exceeded
	//    "World!" = 6 bytes. Total 11 > 10.
	//    Important: Rotation happens BEFORE writing, if the limit would already be reached,
	//    or on the next write.
	//    Our code checks: current + new > max.
	//    5 + 6 = 11 > 10. So rotation happens NOW.
	if _, err := w.Write([]byte("World!")); err != nil {
		t.Fatalf("Write 2 failed: %v", err)
	}

	// Now, "app.log" should only contain "World!".
	// And there should be a .gz file containing "Hello".
	content, _ := os.ReadFile(logPath)
	if string(content) != "World!" {
		t.Errorf("Expected current log to contain 'World!', got '%s'", string(content))
	}

	// 5. Search for .gz files
	files, _ := os.ReadDir(dir)
	foundGzip := false
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".gz") {
			foundGzip = true
			t.Logf("Found backup file: %s", f.Name())
		}
	}

	if !foundGzip {
		t.Error("No .gz backup file found after rotation")
	}
}

func TestRotator_Retention(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "server.log")

	// We want to keep a maximum of 2 backups.
	maxBackups := 2
	w := New(logPath,
		&SizePolicy{MaxBytes: 5}, // Very small limit
		&NoCompression{},         // Simply renaming is enough
		&MaxFiles{MaxBackups: maxBackups},
	)
	defer w.Close()

	// Wir erzeugen 5 Rotationen
	for i := 0; i < 5; i++ {
		w.Write([]byte("123456")) // Trigger Rotation

		// Small sleep timer so that the timestamps of the files are different.
		time.Sleep(10 * time.Millisecond)
	}

	// Analysis of the folder
	entries, _ := os.ReadDir(dir)

	// We expect: 1 active file + 2 backups = 3 files total
	expectedTotal := 1 + maxBackups
	if len(entries) != expectedTotal {
		t.Errorf("Retention failed. Expected %d files total, got %d", expectedTotal, len(entries))
		for _, e := range entries {
			t.Logf("Found: %s", e.Name())
		}
	}
}

func TestRotator_Concurrency(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "concurrent.log")

	w := New(logPath, &SizePolicy{MaxBytes: 1000}, nil, nil)
	defer w.Close()

	// Sync Mechanismen
	start := make(chan struct{})
	done := make(chan struct{})
	goroutines := 10
	writesPerRoutine := 100

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			<-start // Waiting for the starting signal
			for j := 0; j < writesPerRoutine; j++ {
				w.Write([]byte("Data\n"))
			}
			done <- struct{}{}
		}(i)
	}

	close(start) // Let go all at once

	// Wait until everyone is ready
	for i := 0; i < goroutines; i++ {
		<-done
	}

	// Check: We didn't crash.
	// Optional: Check size.
	// 10 routines * 100 writes * 5 bytes ("Data\n") = 5000 bytes total.
	// Since we rotate at 1000 bytes, we should have approximately 5 files.
	// time.Sleep(200 * time.Millisecond)
	entries, _ := os.ReadDir(dir)
	if len(entries) < 4 {
		t.Errorf("Expected rotations under load, got only %d files", len(entries))
	}
}

// Helper Function
func checkFileExists(t *testing.T, path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("File %s does not exist", path)
	}
}
