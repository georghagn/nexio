// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

package main

import (
	//"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/georghagn/gsf-go/pkg/gsflog"
	"github.com/georghagn/gsf-go/pkg/nexIO"
	"github.com/georghagn/gsf-go/pkg/rotate"
	"github.com/georghagn/gsf-go/pkg/schedule"
)

// --- 1. Task Registry (Was kann unser System?) ---
// Da wir via RPC keine Go-Funktionen senden können, mappen wir Strings auf Funktionen.
var taskRegistry = map[string]func(){

	"print_hello": func() { fmt.Println("Task: Hello World!") },
	"db_cleanup":  func() { fmt.Println("Task: Cleaning DB... done.") },
	"heavy_job": func() {
		fmt.Println("Task: Working hard...")
		time.Sleep(2 * time.Second)
		fmt.Println("Task: Working hard finished.")
	},
}

// --- 2. RPC Service Definition ---
// SchedulerRPC ist das Objekt, das wir via NexIO exposen.
// Seine Methoden (Start, Stop, List) werden automatisch zu JSON-RPC.
type SchedulerRPC struct {
	Sched *schedule.Scheduler
	Log   *gsflog.Sink
}

// Argumente für "Scheduler.Start"
type StartArgs struct {
	TaskName string `json:"task_name"` // Welcher Task?
	Interval string `json:"interval"`  // Z.B. "5s", "1m", "500ms"
}

// Start startet einen neuen Job.
func (s *SchedulerRPC) Start(args StartArgs) (int64, error) {

	// 1. Task suchen
	taskFunc, exists := taskRegistry[args.TaskName]
	if !exists {
		return 0, fmt.Errorf("unknown task: %s", args.TaskName)
	}

	// 2. Intervall parsen
	duration, err := time.ParseDuration(args.Interval)
	if err != nil {
		return 0, fmt.Errorf("invalid duration: %v", err)
	}

	// 3. Scheduler beauftragen
	s.Log.Infof("RPC: Starte Task '%s' alle %v", args.TaskName, duration)
	jobID := s.Sched.Every(duration, taskFunc)

	return int64(jobID), nil
}

// Argumente für "Scheduler.Stop"
type StopArgs struct {
	ID int64 `json:"id"`
}

// Stop beendet einen Job.
func (s *SchedulerRPC) Stop(args StopArgs) (string, error) {
	err := s.Sched.Cancel(schedule.JobID(args.ID))
	if err != nil {
		return "", err
	}
	s.Log.Infof("RPC: Job %d gestoppt", args.ID)
	return "Success", nil
}

// List zeigt alle laufenden Jobs.
// Wir brauchen keine Argumente, also leeres Struct oder Ignorieren.
type EmptyArgs struct{}

func (s *SchedulerRPC) List(args EmptyArgs) ([]schedule.JobInfo, error) {
	return s.Sched.List(), nil
}

// --- 3. Main ---
func main() {

	// 1. Logger Setup
	// ACHTUNG: Für den Rotator nutzen wir einen Console-Logger,
	// A. Debug-Logger nur für den Rotator, um die Endlosschleife (Logger -> Rotator -> Logger) zu verhindern.
	consoleLog := gsflog.NewConsole(gsflog.LevelDebug)

	rotator := rotate.New("app.log", nil, nil, nil)
	rotator.SetLogger(consoleLog) // Rotator "spricht" zur Konsole
	defer rotator.Close()

	// B. Haupt-Logger (mainLogger): Soll in Datei UND Konsole schreiben
	// Wir kombinieren Stdout und Rotator
	multiOutput := io.MultiWriter(os.Stdout, rotator)

	// mainLog schreibt jetzt Console UND Datei
	mainLog := gsflog.NewJSON(multiOutput, gsflog.LevelInfo)
	mainLog.Info("GSF System Start")

	// 2. Scheduler Setup & Injection
	sched := schedule.New()
	sched.SetLogger(mainLog) // Scheduler nutzt Haupt-Logger

	// Test Panic Recovery mit Log
	sched.Every(1*time.Second, func() {
		// mainLog.Info("Tick") // Normaler Log
	})

	// 3. NexIO Setup & Injection
	server := nexio.New()
	server.SetLogger(mainLog) // RPC Server nutzt Haupt-Logger

	// ... Services registrieren ...

	mainLog.Info("Setup complete.")

	// Den RPC Service initialisieren und registrieren
	rpcService := &SchedulerRPC{
		Sched: sched,
		Log:   mainLog,
	}

	// NexIO Magic: Macht aus "Start" -> "Scheduler.Start"
	// Wir nennen den Service im RPC explizit "Scheduler" (sonst hieße er "SchedulerRPC")
	// Da RegisterService den Struct-Namen nimmt, müssen wir aufpassen.
	// Unser Reflect-Code nimmt den Struct Namen. Also "SchedulerRPC.Start".
	if err := server.RegisterService(rpcService); err != nil {
		panic(err)
	}

	// HTTP Server Setup
	http.Handle("/rpc", server)
	// WebSocket Support (optional, falls Client das will)
	http.Handle("/ws", http.HandlerFunc(server.ServeWS))

	mainLog.Info("GSF Server bereit auf :8080")
	mainLog.Info("Verfügbare Tasks: 'print_hello', 'db_cleanup', 'heavy_job'")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
