package device

import "testing"

func TestOSTCounterReadsFromTickSource(t *testing.T) {
	var ticks uint64
	ost := NewOST(func() uint64 { return ticks })

	ticks = 10 * OSTCyclesPerTick
	if got := ost.Read32(OST_OST2CNTL); got != 10 {
		t.Fatalf("expected counter low 10, got 0x%08X", got)
	}

	ticks = 42 * OSTCyclesPerTick
	if got := ost.Read32(OST_OST2CNTL); got != 42 {
		t.Fatalf("expected counter low 42, got 0x%08X", got)
	}
}

// The counter is what sched_clock reads, so it has to advance at the OST
// rate rather than at the core clock. Reading it back as the raw cycle
// count would run every kernel timestamp 792x fast.
func TestOSTCounterIsScaledToOSTRate(t *testing.T) {
	var ticks uint64
	ost := NewOST(func() uint64 { return ticks })

	ticks = OSTCyclesPerTick - 1
	if got := ost.Read32(OST_OST2CNTL); got != 0 {
		t.Fatalf("expected counter still 0 below one tick, got 0x%08X", got)
	}

	ticks = OSTCyclesPerTick
	if got := ost.Read32(OST_OST2CNTL); got != 1 {
		t.Fatalf("expected counter 1 after one tick, got 0x%08X", got)
	}
}

// Software samples all 64 bits by reading the low half and then the high
// buffer, which must hold the high half as it was at the low read.
func TestOSTHighHalfLatchesOnLowRead(t *testing.T) {
	var ticks uint64
	ost := NewOST(func() uint64 { return ticks })

	ticks = 0x3_0000_0007 * OSTCyclesPerTick
	if got := ost.Read32(OST_OST2CNTL); got != 0x00000007 {
		t.Fatalf("expected counter low 0x00000007, got 0x%08X", got)
	}

	// Advancing the tick source must not disturb the already-latched
	// high half.
	ticks = 0x4_0000_0000 * OSTCyclesPerTick
	if got := ost.Read32(OST_OSTCNT2HBUF); got != 0x00000003 {
		t.Fatalf("expected latched high 0x00000003, got 0x%08X", got)
	}
}

func TestOSTClearRestartsCounter(t *testing.T) {
	var ticks uint64
	ost := NewOST(func() uint64 { return ticks })

	ticks = 100 * OSTCyclesPerTick
	if got := ost.Read32(OST_OST2CNTL); got != 100 {
		t.Fatalf("expected counter 100, got 0x%08X", got)
	}

	ost.Write32(OST_OSTCR, OST_CLR_OST2)
	if got := ost.Read32(OST_OST2CNTL); got != 0 {
		t.Fatalf("expected counter 0 after clear, got 0x%08X", got)
	}

	ticks += 5 * OSTCyclesPerTick
	if got := ost.Read32(OST_OST2CNTL); got != 5 {
		t.Fatalf("expected counter 5 after clear, got 0x%08X", got)
	}
}

// OSTESR and OSTECR are set/clear aliases; the state reads back through
// OSTER. The kernel starts OST2 and stops OST1 through exactly this pair.
func TestOSTEnableSetAndClear(t *testing.T) {
	ost := NewOST(func() uint64 { return 0 })

	ost.Write32(OST_OSTESR, OST_BIT_OST1|OST_BIT_OST2)
	if got := ost.Read32(OST_OSTER); got != OST_BIT_OST1|OST_BIT_OST2 {
		t.Fatalf("expected both timers enabled, got 0x%08X", got)
	}

	ost.Write32(OST_OSTECR, OST_BIT_OST1)
	if got := ost.Read32(OST_OSTER); got != OST_BIT_OST2 {
		t.Fatalf("expected only OST2 enabled, got 0x%08X", got)
	}
}
