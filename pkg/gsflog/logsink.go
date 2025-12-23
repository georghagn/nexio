package gsflog

type LogSink interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)

	SetLevel(Level)
	With(key string, value any) LogSink
}
