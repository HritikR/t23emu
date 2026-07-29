package device

import "testing"

func TestINTCPendingRespectsMask(t *testing.T) {
	intc := NewINTC()
	intc.Assert(3)

	if got := intc.Pending(); got != 0 {
		t.Fatalf("expected masked pending IRQ to be hidden, got 0x%08X", got)
	}

	intc.Write32(INTC_IMCR, 1<<3)

	if got := intc.Read32(INTC_ISR); got != 1<<3 {
		t.Fatalf("expected IRQ3 pending in bank 0 ISR, got 0x%08X", got)
	}
	if got := intc.Pending(); got != 1 {
		t.Fatalf("expected unmasked pending IRQ, got 0x%08X", got)
	}
}

func TestINTCStatusReadDoesNotConsumePending(t *testing.T) {
	intc := NewINTC()
	intc.Write32(INTC_IMCR, 1<<3)
	intc.Assert(3)

	if got := intc.Read32(INTC_ISR); got != 1<<3 {
		t.Fatalf("expected IRQ3 pending, got 0x%08X", got)
	}
	if got := intc.Pending(); got != 1 {
		t.Fatalf("expected pending IRQ to remain asserted, got 0x%08X", got)
	}
}

func TestINTCBank1PendingRespectsMask(t *testing.T) {
	intc := NewINTC()
	intc.Assert(40)

	if got := intc.Pending(); got != 0 {
		t.Fatalf("expected masked bank 1 IRQ to be hidden, got 0x%08X", got)
	}

	intc.Write32(INTC_IMCR+intcBankStride, 1<<8)

	if got := intc.Read32(INTC_ISR + intcBankStride); got != 1<<8 {
		t.Fatalf("expected bank 1 IRQ8 pending, got 0x%08X", got)
	}
	if got := intc.Pending(); got != 1 {
		t.Fatalf("expected an unmasked pending IRQ, got 0x%08X", got)
	}
}
