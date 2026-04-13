package client

import (
	"strings"
	"testing"
)

func TestFormatDebugPhotonParams_collapsesZeros(t *testing.T) {
	zeros := make([]int16, 50)
	zeros[0] = 1
	zeros[49] = 2
	m := map[uint8]interface{}{
		1: zeros,
		2: append(make([]byte, 40), 0xab, 0xcd),
	}
	s := formatDebugPhotonParams(m)
	if strings.Count(s, "0 ") > 15 {
		t.Fatalf("expected collapsed zeros, got: %s", s)
	}
	if !strings.Contains(s, "×") && !strings.Contains(s, "…(") {
		t.Fatalf("expected collapse marker in: %s", s)
	}
}
