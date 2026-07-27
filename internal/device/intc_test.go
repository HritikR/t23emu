package device

import "testing"

func TestINTCPendingRespectsMask(t *testing.T) {
	intc := NewINTC()
	intc.Assert(3)

	if got := intc.Pending(); got != 0 {
		t.Fatalf("expected masked pending IRQ to be hidden, got 0x%08X", got)
	}

	intc.Write32(INTC_IMCR, 1<<3)

	if got := intc.Pending(); got != 1<<3 {
		t.Fatalf("expected unmasked pending IRQ, got 0x%08X", got)
	}
}

func TestINTCStatusReadConsumesPending(t *testing.T) {
	intc := NewINTC()
	intc.Write32(INTC_IMCR, 1<<3)
	intc.Assert(3)

	if got := intc.Read32(INTC_ISR); got != 1<<3 {
		t.Fatalf("expected IRQ3 pending, got 0x%08X", got)
	}
	if got := intc.Pending(); got != 0 {
		t.Fatalf("expected pending IRQ to be consumed, got 0x%08X", got)
	}
}
