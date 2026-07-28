package device

import "testing"

func TestOSTCounterReadsFromTickSource(t *testing.T) {
	var ticks uint64 = 10
	ost := NewOST(func() uint64 { return ticks })

	if got := ost.Read32(OST_CNTL); got != 10 {
		t.Fatalf("expected counter low 10, got 0x%08X", got)
	}

	ticks = 42
	if got := ost.Read32(OST_CNTL); got != 42 {
		t.Fatalf("expected counter low 42, got 0x%08X", got)
	}
}
