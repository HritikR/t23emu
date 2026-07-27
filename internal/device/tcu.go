package device

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

// NewTCU creates the timer block, with the OS timer counter driven by
// ticks.
//
// The OS timer has to actually advance. Firmware implements udelay by
// sampling this counter and spinning until it passes a target, so a
// counter stuck at zero turns every delay in the boot path into an
// infinite loop. Driving it from the instruction count means delays
// terminate in proportion to work done rather than to wall time, which is
// what a cycle-stepped interpreter can offer.
func NewTCU(ticks func() uint64) *RegisterBlock {
	tcu := NewRegisterBlock("TCU", 0x1000)

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
