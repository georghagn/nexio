package rotate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRotator_Integration_Gzip(t *testing.T) {
	// 1. Temporäres Verzeichnis für diesen Test (wird automatisch gelöscht)
	dir := t.TempDir()
	logPath := filepath.Join(dir, "app.log")

	// 2. Setup: Rotieren nach 10 Bytes, mit Gzip Kompression
	w := New(logPath,
		&SizePolicy{MaxBytes: 10},
		&GzipCompression{},
		&KeepAll{}, // Retention prüfen wir separat
	)
	defer w.Close()

	// 3. Schreiben (unter dem Limit)
	// "Hello" = 5 Bytes
	if _, err := w.Write([]byte("Hello")); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Prüfen ob Datei existiert
	checkFileExists(t, logPath)

	// 4. Limit überschreiten
	// "World!" = 6 Bytes. Total 11 > 10.
	// Wichtig: Rotation passiert VOR dem Schreiben, wenn Limit schon erreicht WÄRE,
	// oder BEIM nächsten Write.
	// Unser Code prüft: current + new > max.
	// 5 + 6 = 11 > 10. Also Rotation JETZT.
	if _, err := w.Write([]byte("World!")); err != nil {
		t.Fatalf("Write 2 failed: %v", err)
	}

	// Jetzt sollte "app.log" nur "World!" enthalten.
	// Und es sollte eine .gz Datei geben mit "Hello".
	content, _ := os.ReadFile(logPath)
	if string(content) != "World!" {
		t.Errorf("Expected current log to contain 'World!', got '%s'", string(content))
	}

	// 5. Nach .gz Datei suchen
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

	// Wir wollen max 2 Backups behalten
	maxBackups := 2
	w := New(logPath,
		&SizePolicy{MaxBytes: 5}, // Sehr kleines Limit
		&NoCompression{},         // Einfaches Umbenennen reicht
		&MaxFiles{MaxBackups: maxBackups},
	)
	defer w.Close()

	// Wir erzeugen 5 Rotationen
	for i := 0; i < 5; i++ {
		w.Write([]byte("123456")) // Trigger Rotation

		// Kleiner Sleep, damit die Zeitstempel der Dateien unterschiedlich sind
		time.Sleep(10 * time.Millisecond)
	}

	// Analyse des Ordners
	entries, _ := os.ReadDir(dir)

	// Wir erwarten: 1 aktives File + 2 Backups = 3 Dateien total
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
			<-start // Warten auf Startschuss
			for j := 0; j < writesPerRoutine; j++ {
				w.Write([]byte("Data\n"))
			}
			done <- struct{}{}
		}(i)
	}

	close(start) // Alle gleichzeitig loslassen

	// Warten bis alle fertig
	for i := 0; i < goroutines; i++ {
		<-done
	}

	// Check: Wir haben nicht gecrasht.
	// Optional: Größe prüfen.
	// 10 Routines * 100 Writes * 5 Bytes ("Data\n") = 5000 Bytes Total.
	// Da wir bei 1000 Bytes rotieren, sollten wir ca. 5 Dateien haben.
	//time.Sleep(200 * time.Millisecond)
	entries, _ := os.ReadDir(dir)
	if len(entries) < 4 {
		t.Errorf("Expected rotations under load, got only %d files", len(entries))
	}
}

// Helper Funktion
func checkFileExists(t *testing.T, path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("File %s does not exist", path)
	}
}
