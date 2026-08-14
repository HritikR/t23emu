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

// WDTCyclesPerTick converts one watchdog data-register tick into core
// cycles. The watchdog is clocked from RTC (32768 Hz) through the TCSR
// prescaler; on the T23 thingino firmware the effective rate works out
// to approximately 1 Hz, so one tick ≈ 1 second ≈ 1.188 billion cycles
// at the 1188 MHz CCLK.
const WDTCyclesPerTick uint64 = 1_188_000_000

// TCU models the Timer/Counter Unit. It embeds RegisterBlock for storage
// and adds watchdog countdown logic: when the guest enables the
// watchdog, the expiry is computed from TDR. On every Step() the CPU
// checks WatchdogExpired(); if it returns true the machine reboots.
// Writing TCNT (kicking the watchdog) resets the countdown.
type TCU struct {
	*RegisterBlock

	OnWatchdogReset func()

	ticks func() uint64

	// wdtExpiryCycle is the core cycle count at which the watchdog
	// fires, or 0 if the watchdog is not armed.
	wdtExpiryCycle uint64
}

// NewTCU creates the timer block, with the OS timer counter and
// watchdog driven by ticks (the CPU cycle count).
func NewTCU(ticks func() uint64) *TCU {
	tcu := &TCU{
		RegisterBlock: NewRegisterBlock("TCU", 0x1000),
		ticks:         ticks,
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

// Write32 intercepts watchdog control and counter writes to maintain
// the countdown state.
func (t *TCU) Write32(addr uint32, value uint32) {
	offset := addr &^ 3
	t.writeCounts[offset]++
	t.regs[offset] = value

	if fn, ok := t.writeFuncs[offset]; ok {
		fn(value)
	}

	switch offset {
	case TCU_WDT_TCER:
		if value&WDT_TCER_TCEN != 0 {
			t.armWatchdog()
		} else {
			t.wdtExpiryCycle = 0
		}
	case TCU_WDT_TCNT:
		// Writing TCNT reloads the counter and restarts the countdown.
		if t.wdtExpiryCycle != 0 {
			t.armWatchdog()
		}
	case TCU_WDT_TDR:
		// If the watchdog is already armed, update the expiry.
		if t.wdtExpiryCycle != 0 {
			t.armWatchdog()
		}
	}

	if t.Trace {
		fmt.Fprintf(t.Out, "  %s write %s <= 0x%08x\n", t.Name, t.RegName(offset), value)
	}
}

// armWatchdog computes the expiry cycle from the current tick count and
// the data register (TDR).
func (t *TCU) armWatchdog() {
	tdr := t.regs[TCU_WDT_TDR]
	if tdr == 0 {
		tdr = 1
	}
	t.wdtExpiryCycle = t.ticks() + uint64(tdr)*WDTCyclesPerTick
}

// WatchdogExpired returns true if the watchdog countdown has reached
// zero. The CPU calls this on every Step().
func (t *TCU) WatchdogExpired() bool {
	if t.wdtExpiryCycle == 0 {
		return false
	}
	if t.ticks() >= t.wdtExpiryCycle {
		t.wdtExpiryCycle = 0
		t.OnWatchdogReset()
		return true
	}
	return false
}

// WatchdogExpiryCycle returns the cycle count at which the watchdog
// fires, or 0 if not armed. This lets the machine include the watchdog
// deadline in NextWakeCycle so the idle-loop fast-forward does not skip
// past it.
func (t *TCU) WatchdogExpiryCycle() uint64 {
	return t.wdtExpiryCycle
}

// Reset clears all watchdog state so a rebooted kernel starts fresh.
func (t *TCU) Reset() {
	t.wdtExpiryCycle = 0
}

var _ Device = (*TCU)(nil)
