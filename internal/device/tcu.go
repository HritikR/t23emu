package device

import "fmt"

// Timer/Counter Unit register offsets, relative to the TCU physical base
// at 0x10002000. The watchdog, the general purpose timers and the OS
// timer all live in this one window.
const (
	// Watchdog.
	TCU_WDT_TDR  uint32 = 0x000
	TCU_WDT_TCER uint32 = 0x004
	TCU_WDT_TCNT uint32 = 0x008
	TCU_WDT_TCSR uint32 = 0x00C

	// Timer enable/status.
	TCU_TER  uint32 = 0x010
	TCU_TESR uint32 = 0x014
	TCU_TECR uint32 = 0x018
	TCU_TFR  uint32 = 0x020
	TCU_TFSR uint32 = 0x024
	TCU_TFCR uint32 = 0x028
	TCU_TMR  uint32 = 0x030
	TCU_TSR  uint32 = 0x01C

	// OS timer: a free-running 64-bit counter used for delay loops.
	TCU_OSTDR      uint32 = 0x0E0
	TCU_OSTCNTL    uint32 = 0x0E4
	TCU_OSTCNTH    uint32 = 0x0E8
	TCU_OSTCSR     uint32 = 0x0EC
	TCU_OSTFLAG    uint32 = 0x0F0
	TCU_OSTCNTHBUF uint32 = 0x0FC
)

// WDT_TCER_TCEN is the Timer Counter Enable bit in the watchdog control
// register. Writing it starts the watchdog countdown.
const WDT_TCER_TCEN uint32 = 1 << 0

// TCU models the Timer/Counter Unit. It embeds RegisterBlock for storage
// and adds watchdog reset detection: when the guest enables the watchdog
// with a zero data register (TDR=0), as the Ingenic kernel does during
// reboot, OnWatchdogReset is called.
type TCU struct {
	*RegisterBlock

	// OnWatchdogReset is called when the guest programs the watchdog
	// for an immediate system reset (TCEN written with TDR=0).
	OnWatchdogReset func()
}

// NewTCU creates the timer block, with the OS timer counter driven by
// ticks.
//
// The OS timer has to actually advance. Firmware implements udelay by
// sampling this counter and spinning until it passes a target, so a
// counter stuck at zero turns every delay in the boot path into an
// infinite loop. Driving it from the instruction count means delays
// terminate in proportion to work done rather than to wall time, which is
// what a cycle-stepped interpreter can offer.
func NewTCU(ticks func() uint64) *TCU {
	tcu := &TCU{
		RegisterBlock: NewRegisterBlock("TCU", 0x1000),
	}

	names := map[uint32]string{
		TCU_WDT_TDR:    "WDT_TDR",
		TCU_WDT_TCER:   "WDT_TCER",
		TCU_WDT_TCNT:   "WDT_TCNT",
		TCU_WDT_TCSR:   "WDT_TCSR",
		TCU_TER:        "TER",
		TCU_TESR:       "TESR",
		TCU_TECR:       "TECR",
		TCU_TSR:        "TSR",
		TCU_TFR:        "TFR",
		TCU_TFSR:       "TFSR",
		TCU_TFCR:       "TFCR",
		TCU_TMR:        "TMR",
		TCU_OSTDR:      "OSTDR",
		TCU_OSTCNTL:    "OSTCNTL",
		TCU_OSTCNTH:    "OSTCNTH",
		TCU_OSTCSR:     "OSTCSR",
		TCU_OSTFLAG:    "OSTFLAG",
		TCU_OSTCNTHBUF: "OSTCNTHBUF",
	}
	for offset, name := range names {
		tcu.SetName(offset, name)
	}

	// hiBuf latches the high half of the counter when the low half is
	// read, which is how the hardware lets software sample all 64 bits
	// without the halves tearing across a carry.
	var hiBuf uint32

	tcu.SetReadFunc(TCU_OSTCNTL, func() uint32 {
		now := ticks()
		hiBuf = uint32(now >> 32)
		return uint32(now)
	})

	tcu.SetReadFunc(TCU_OSTCNTH, func() uint32 {
		return uint32(ticks() >> 32)
	})

	tcu.SetReadFunc(TCU_OSTCNTHBUF, func() uint32 {
		return hiBuf
	})

	return tcu
}

// Write32 intercepts watchdog control writes. When the guest enables the
// watchdog (TCEN) with TDR=0 — the Ingenic kernel's reboot sequence —
// the OnWatchdogReset callback fires so the emulator can halt or restart.
//
// Normal watchdog operation (TDR > 0) is ignored because the emulator
// does not model the countdown. The guest's watchdog daemon kicks TCNT
// before it reaches zero, so no false reset occurs.
func (t *TCU) Write32(addr uint32, value uint32) {
	offset := addr &^ 3
	t.writeCounts[offset]++
	t.regs[offset] = value

	if fn, ok := t.writeFuncs[offset]; ok {
		fn(value)
	}

	if offset == TCU_WDT_TCER && value&WDT_TCER_TCEN != 0 {
		if t.regs[TCU_WDT_TDR] == 0 && t.OnWatchdogReset != nil {
			t.OnWatchdogReset()
		}
	}

	if t.Trace {
		fmt.Fprintf(t.Out, "  %s write %s <= 0x%08x\n", t.Name, t.RegName(offset), value)
	}
}

var _ Device = (*TCU)(nil)
