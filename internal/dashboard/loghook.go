package dashboard

import (
	"sync"

	"github.com/sirupsen/logrus"
)

const maxLogLines = 500

// LogLine is one captured log entry, formatted for display.
type LogLine struct {
	Time    string
	Level   string
	Message string
}

var (
	logMu    sync.Mutex
	logLines []LogLine
	logEmit  func(LogLine)
)

// LogHook is a logrus.Hook that keeps the most recent maxLogLines log
// lines in memory and, if a subscriber is registered, forwards each new
// line as it arrives.
type LogHook struct{}

// NewLogHook returns a LogHook ready to be registered with log.AddHook.
func NewLogHook() *LogHook {
	return &LogHook{}
}

// Levels implements logrus.Hook: fire for Info level and more severe,
// excluding Debug and Trace. Trace/Debug logging (e.g. per-packet trace
// lines in client/listener.go) would otherwise flood the dashboard's
// in-memory ring buffer and live event stream when a user enables
// -trace/-debug for bug reports; this only affects what the dashboard
// captures, not what's written to the log file/console.
func (h *LogHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
	}
}

// Fire implements logrus.Hook.
func (h *LogHook) Fire(entry *logrus.Entry) error {
	line := LogLine{
		Time:    entry.Time.Format("15:04:05"),
		Level:   entry.Level.String(),
		Message: entry.Message,
	}

	logMu.Lock()
	logLines = append(logLines, line)
	if len(logLines) > maxLogLines {
		logLines = logLines[len(logLines)-maxLogLines:]
	}
	emit := logEmit
	logMu.Unlock()

	if emit != nil {
		emit(line)
	}
	return nil
}

// OnLogLine registers the callback invoked once per new log line.
func OnLogLine(fn func(LogLine)) {
	logMu.Lock()
	logEmit = fn
	logMu.Unlock()
}

// GetRecentLogs returns a copy of the currently buffered log lines,
// oldest first.
func GetRecentLogs() []LogLine {
	logMu.Lock()
	defer logMu.Unlock()
	out := make([]LogLine, len(logLines))
	copy(out, logLines)
	return out
}

// resetLogHookForTest clears all package state. Test-only.
func resetLogHookForTest() {
	logMu.Lock()
	logLines = nil
	logEmit = nil
	logMu.Unlock()
}
