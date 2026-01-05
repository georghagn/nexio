package nexlog

import "sync"

type TestSink struct {
	mu     sync.Mutex
	Calls  []string
	Levels []Level
}

func NewTestSink() *TestSink {
	return &TestSink{}
}

func (t *TestSink) record(level Level, msg string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Calls = append(t.Calls, msg)
	t.Levels = append(t.Levels, level)
}

func (t *TestSink) Debug(msg string) { t.record(LevelDebug, msg) }
func (t *TestSink) Info(msg string)  { t.record(LevelInfo, msg) }
func (t *TestSink) Warn(msg string)  { t.record(LevelWarn, msg) }
func (t *TestSink) Error(msg string) { t.record(LevelError, msg) }

func (t *TestSink) SetLevel(Level) {}

func (t *TestSink) With(key string, value any) LogSink {
	return t // its ok for tests
}
