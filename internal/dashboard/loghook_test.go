package dashboard

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestLogHook_FireAppendsAndCaps(t *testing.T) {
	resetLogHookForTest()
	hook := NewLogHook()

	for i := 0; i < maxLogLines+10; i++ {
		if err := hook.Fire(&logrus.Entry{Level: logrus.InfoLevel, Message: "line"}); err != nil {
			t.Fatalf("Fire returned error: %v", err)
		}
	}

	got := GetRecentLogs()
	if len(got) != maxLogLines {
		t.Fatalf("len(GetRecentLogs()) = %d, want %d", len(got), maxLogLines)
	}
}

func TestLogHook_FireNotifiesSubscriber(t *testing.T) {
	resetLogHookForTest()
	hook := NewLogHook()

	var got []LogLine
	OnLogLine(func(l LogLine) { got = append(got, l) })

	_ = hook.Fire(&logrus.Entry{Level: logrus.WarnLevel, Message: "careful"})

	if len(got) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(got))
	}
	if got[0].Level != "warning" || got[0].Message != "careful" {
		t.Fatalf("unexpected log line: %+v", got[0])
	}
}

func TestLogHook_LevelsExcludesDebugAndTrace(t *testing.T) {
	hook := NewLogHook()
	levels := hook.Levels()

	for _, l := range levels {
		if l == logrus.DebugLevel || l == logrus.TraceLevel {
			t.Fatalf("Levels() unexpectedly includes %s; debug/trace should be excluded to avoid flooding the dashboard", l)
		}
	}
	if len(levels) == 0 {
		t.Fatal("Levels() returned no levels")
	}
}
