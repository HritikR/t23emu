package device

// OST register offsets for the timer block Linux maps at 0x12000000.
//
// These are the Ingenic OST registers, not a guess: every offset the
// kernel writes during timer init decodes cleanly against them. It
// programs OSTCCR with the prescaler, OSTMR with 0xFFFFFFFE to unmask
// only OST1, OST1DFR with 14999 for the 100 Hz tick, then OSTESR with
// 0x2 to start OST2 and OSTECR with 0x1 to stop OST1.
const (
	OST_OSTCCR      uint32 = 0x00 // clock/prescaler control
	OST_OSTER       uint32 = 0x04 // enable status
	OST_OSTCR       uint32 = 0x08 // counter clear
	OST_OSTFR       uint32 = 0x0C // interrupt flags
	OST_OSTMR       uint32 = 0x10 // interrupt mask
	OST_OST1DFR     uint32 = 0x14 // OST1 compare value
	OST_OST1CNT     uint32 = 0x18 // OST1 32-bit counter
	OST_OST2CNTL    uint32 = 0x20 // OST2 free-running counter, low half
	OST_OSTCNT2HBUF uint32 = 0x24 // OST2 high half, latched on low read
	OST_OSTESR      uint32 = 0x34 // enable set
	OST_OSTECR      uint32 = 0x38 // enable clear
)

// Bits shared by OSTER, OSTESR, OSTECR, OSTFR and OSTMR: one per timer.
const (
	OST_BIT_OST1 uint32 = 1 << 0
	OST_BIT_OST2 uint32 = 1 << 1
)

// Clear bits in OSTCR.
const (
	OST_CLR_OST1 uint32 = 1 << 0
	OST_CLR_OST2 uint32 = 1 << 1
)

// OSTCyclesPerTick converts emulated instruction cycles into OST ticks.
//
// OST2 counts EXTAL divided by the OSTCCR prescaler, which works out to
// 1.5 MHz on this board: the kernel programs OST1DFR to 14999 for its
// 100 Hz tick, and 1500000/15000 is exactly 100. The core it is told
// about runs at CCLK = 1188 MHz, so one OST tick is 792 core cycles.
//
// The ratio matters because this counter is what the kernel's
// sched_clock reads, and sched_clock is what stamps every printk line.
// Feeding it raw cycles would run the clock 792x fast and make every
// timestamp in the boot log a fiction.
const OSTCyclesPerTick uint64 = 792

// OST is the OS timer block behind the Linux clocksource and tick.
//
// Two things hang off it, and both have to work or the boot stalls in a
// different place:
//
//   - OST2 is the free-running 64-bit counter read by the jz clocksource
//     and by sched_clock. A counter stuck at zero does not hang anything,
//     it just silently pins every kernel timestamp at [    0.000000].
//   - OST1 is the periodic tick. Its interrupt is what advances jiffies,
//     and the handler decides whether to dispatch by reading the OST1 bit
//     in OSTFR. A flag that never sets leaves jiffies frozen, which hangs
//     the first thing that waits on them: calibrate_delay().
//
// So the flag and the interrupt are generated from one place here, from
// the compare value the kernel itself programmed, rather than the two
// being wired up separately and left free to disagree.
type OST struct {
	*RegisterBlock

	ticks func() uint64

	// origin is the tick value OST2 counts from, moved forward when
	// software clears the counter through OSTCR.
	origin uint64

	// hiBuf latches the high half of OST2 when the low half is read,
	// which is how the hardware lets software sample all 64 bits without
	// the halves tearing across a carry.
	hiBuf uint32

	// enable is the OSTER state, driven by the OSTESR/OSTECR aliases.
	enable uint32

	// period is OST1DFR + 1, the tick interval in OST ticks.
	period uint64

	// nextCompare is the counter value at which OST1 next fires, and
	// pending is whether it has fired and not yet been acknowledged.
	nextCompare uint64
	pending     bool
}

// NewOST creates the OS timer block used by the Linux clocksource path.
func NewOST(ticks func() uint64) *OST {
	o := &OST{
		RegisterBlock: NewRegisterBlock("OST", 0x1000),
		ticks:         ticks,
	}

	names := map[uint32]string{
		OST_OSTCCR:      "OSTCCR",
		OST_OSTER:       "OSTER",
		OST_OSTCR:       "OSTCR",
		OST_OSTFR:       "OSTFR",
		OST_OSTMR:       "OSTMR",
		OST_OST1DFR:     "OST1DFR",
		OST_OST1CNT:     "OST1CNT",
		OST_OST2CNTL:    "OST2CNTL",
		OST_OSTCNT2HBUF: "OSTCNT2HBUF",
		OST_OSTESR:      "OSTESR",
		OST_OSTECR:      "OSTECR",
	}
	for offset, name := range names {
		o.SetName(offset, name)
	}

	o.SetReadFunc(OST_OST2CNTL, func() uint32 {
		now := o.counter()
		o.hiBuf = uint32(now >> 32)
		return uint32(now)
	})

	o.SetReadFunc(OST_OSTCNT2HBUF, func() uint32 {
		return o.hiBuf
	})

	// OST1 counts from the last compare rather than from reset.
	o.SetReadFunc(OST_OST1CNT, func() uint32 {
		now := o.counter()
		if o.period == 0 || now+o.period < o.nextCompare {
			return 0
		}
		return uint32(now + o.period - o.nextCompare)
	})

	// OSTFR reports the pending tick. Reading it is treated as consuming
	// it, which real hardware does not do; it is a deliberate backstop so
	// that a handler acknowledging the tick in some way this model does
	// not recognise still gets exactly one dispatch per period instead of
	// an interrupt storm that starves the boot.
	o.SetReadFunc(OST_OSTFR, func() uint32 {
		value := o.regs[OST_OSTFR]
		if o.expired() {
			value |= OST_BIT_OST1
			o.ack()
		}
		return value
	})

	// Writing OSTFR acknowledges the tick, which is the ordinary path.
	o.SetWriteFunc(OST_OSTFR, func(uint32) {
		o.ack()
	})

	// OSTESR and OSTECR are set/clear aliases onto the enable state that
	// reads back through OSTER; neither holds a value of its own.
	o.SetReadFunc(OST_OSTER, func() uint32 {
		return o.enable
	})
	o.SetWriteFunc(OST_OSTESR, func(value uint32) {
		o.enable |= value
	})
	o.SetWriteFunc(OST_OSTECR, func(value uint32) {
		o.enable &^= value
	})

	// The tick interval is only known once the kernel programs it.
	o.SetWriteFunc(OST_OST1DFR, func(value uint32) {
		o.period = uint64(value) + 1
		o.nextCompare = o.counter() + o.period
		o.pending = false
	})

	// OSTCR is write-one-to-clear on each counter. OST2 is left
	// free-running regardless of the enable bits, so that a driver which
	// reads it before enabling it still sees time pass rather than
	// silently getting zeroes.
	o.SetWriteFunc(OST_OSTCR, func(value uint32) {
		if value&OST_CLR_OST2 != 0 {
			o.origin = o.ticks()
			o.hiBuf = 0
		}
		if value&OST_CLR_OST1 != 0 {
			o.ack()
		}
	})

	return o
}

// counter returns the OST2 free-running counter in OST ticks.
func (o *OST) counter() uint64 {
	now := o.ticks()
	if now < o.origin {
		return 0
	}
	return (now - o.origin) / OSTCyclesPerTick
}

// ack clears a pending tick and arms the next one.
func (o *OST) ack() {
	o.pending = false
	if o.period != 0 {
		now := o.counter()
		if o.nextCompare <= now {
			o.nextCompare = now + o.period
		}
	}
}

// expired reports whether OST1 has reached its compare value and not yet
// been acknowledged.
//
// The compare is deliberately not gated on the OST1 enable bit. The tick
// is the only thing advancing jiffies in this machine, and stopping it
// because of an enable write whose ordering this model may have wrong
// would wedge the boot with no visible cause.
func (o *OST) expired() bool {
	if o.period == 0 {
		return false
	}
	if !o.pending && o.counter() >= o.nextCompare {
		o.pending = true
	}
	return o.pending
}

// OST1Expired reports a pending periodic tick, for the machine to raise
// as an interrupt.
func (o *OST) OST1Expired() bool {
	return o.expired()
}

var _ Device = (*OST)(nil)
