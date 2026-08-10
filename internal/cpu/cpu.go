package cpu

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/HritikR/t23emu/internal/bus"
)

const HistorySize = 512

// spinForwardThreshold is the number of Steps without an interrupt
// before the CPU fast-forwards through a spin loop (e.g. the kernel
// reboot loop where interrupts are disabled). The OST tick fires every
// ~12M cycles, so this threshold is well clear of normal operation.
const spinForwardThreshold = 100000

type CPU struct {
	// General purpose registers
	// MIPS has 32 registers: $zero-$ra
	Regs [32]uint32

	// HI and LO hold the results of multiply and divide operations.
	HI uint32
	LO uint32

	// Floating-point registers are modelled as raw 32-bit lanes.
	FPR  [32]uint32
	FCSR uint32

	// PC is the address of the next instruction to fetch.
	//
	// MIPS has a one-instruction branch delay slot, so a taken branch
	// cannot simply overwrite PC: the instruction already sitting in the
	// delay slot must execute first. The delay is modelled by keeping two
	// addresses, PC and NextPC. A branch writes its target to NextPC,
	// which leaves the delay slot at PC to execute on the following step.
	PC uint32

	// NextPC is the address of the instruction after the one at PC.
	NextPC uint32

	// CurrentPC is the address the instruction being executed was fetched
	// from. Exception handling needs it because PC has already advanced.
	CurrentPC uint32

	// InDelaySlot reports whether the instruction being executed sits in
	// the delay slot of a taken branch. Exceptions raised here must report
	// the branch as the faulting instruction, not the delay slot.
	InDelaySlot bool

	// branchTaken records that the instruction just executed was a taken
	// branch, so the next instruction is a delay slot.
	branchTaken bool

	// Current fetched instruction
	Instruction uint32

	// Memory interface
	Bus *bus.Bus

	// CPU execution state
	Running bool

	// Waiting reports that WAIT has stopped instruction fetch until an
	// interrupt or implementation-specific wake event resumes the core.
	Waiting bool

	// CPU halt status
	HaltReason HaltReason

	// HaltDetail carries a human-readable explanation of why the CPU
	// halted, which is otherwise lost when execution simply stops.
	HaltDetail string

	// Cycle counter
	Cycles uint64

	// Reset PC address
	ResetPC uint32

	// Coprocessor 0 registers
	CP0 [32]uint32

	// UserLocal is CP0 register 4 select 2, read by userspace through
	// RDHWR register 29 for TLS.
	UserLocal uint32

	// CP0 Count is derived from Cycles, but writes reset the visible base.
	countBaseCycle uint64
	countBaseValue uint32
	compareSet     bool

	// TLB contains the CP0-managed virtual mappings used by kuseg/kseg2.
	TLB [32]TLBEntry

	// TLB statistics for the debug report.
	TLBHits   uint64
	TLBMisses uint64
	TLBWI     uint64
	TLBWR     uint64

	// InterruptPending returns CP0 Cause.IP bits currently asserted by
	// external interrupt hardware.
	InterruptPending func() uint32

	// WakePending reports implementation-specific activity that can resume
	// the core from WAIT without necessarily being a deliverable interrupt.
	WakePending func() bool

	// NextWakeCycle returns the cycle count of the earliest pending wake
	// event (e.g. the next OST tick), or 0 if none. When the CPU is
	// WAITing and no interrupt is immediately deliverable, Step() uses
	// this to fast-forward Cycles to the next event instead of spinning.
	NextWakeCycle func() uint64

	// WatchdogCheck is called on every Step(). If it returns true, the
	// machine halts for a watchdog reset (reboot).
	WatchdogCheck func() bool

	// LLBit is the load-linked bit set by LL and tested by SC.
	LLBit bool

	// TLS caching: when the kernel emulates rdhwr $29 via an RI trap,
	// we observe the value it writes and cache it so subsequent calls
	// return instantly without trapping. pendingTLSRt is the register
	// the kernel will write; 0 means no pending observation.
	// cachedTLSASID records the ASID (EntryHi bits 7:0) of the process
	// whose TLS we cached; if a context switch changes the ASID, the
	// cache is invalidated automatically.
	pendingTLSPC    uint32
	pendingTLSRt    uint8
	cachedTLSASID   uint32

	// RITraceOut, when non-nil, receives one line per Reserved
	// Instruction exception with the faulting PC, raw word, decoded
	// fields, and a disassembly. Booting real firmware can otherwise
	// surface unimplemented instructions only as kernel "Illegal
	// instruction" messages, with no indication of the offending opcode.
	RITraceOut io.Writer
	riCounters map[string]*RICounter

	// TraceADE, when set, halts on the first userspace AdEL/AdES
	// exception and dumps the instruction, registers, and history.
	TraceADE bool

	// Instruction tracing
	Trace bool

	// TraceOut receives trace output. Defaults to os.Stderr so that a
	// trace can be redirected independently of emulated UART output.
	TraceOut io.Writer

	// exceptionRun counts exceptions taken without an intervening
	// successful instruction retire, used to detect a fault storm.
	exceptionRun int

	// MaxExceptionRun is the number of back-to-back exceptions tolerated
	// before the CPU halts. A MIPS core with a bad exception vector will
	// otherwise spin forever, burning the whole cycle budget and hiding
	// the original fault.
	MaxExceptionRun int

	// Instruction execution history (saved last HistorySize instructions)
	History          [HistorySize]HistoryEntry
	HistoryIndex     int
	HistoryFull      bool
	RecordHistory    bool
	currentMemAddr   uint32
	currentMemVal    uint32
	currentMemAccess string

	// stepsSinceInterrupt counts Steps taken without delivering an
	// interrupt. During normal operation the OST fires every ~12M
	// cycles, so this rarely exceeds a few thousand. During a reboot
	// spin loop (interrupts disabled), it climbs indefinitely. When it
	// passes spinForwardThreshold and the watchdog is armed, we
	// fast-forward to the watchdog deadline so the reboot fires
	// quickly.
	stepsSinceInterrupt int

	// Debugger breakpoints, watchpoints, and single stepping
	Breakpoints   map[uint32]bool
	Watchpoints   map[string]Watchpoint
	HitWatchpoint uint32
	SingleStep    bool

	icache [65536]icacheEntry
}

type icacheEntry struct {
	valid bool
	raw   uint32
	inst  Instruction
}

type WatchpointType int

const (
	WatchWrite  WatchpointType = 2
	WatchRead   WatchpointType = 3
	WatchAccess WatchpointType = 4
)

type Watchpoint struct {
	Addr uint32
	Len  uint32
	Type WatchpointType
}

// RICounter tallies Reserved Instruction exceptions that share a key.
// The key groups together identical encodings (opcode/funct/register
// fields) so the summary can distinguish one missing instruction from a
// flood of identical traps.
type RICounter struct {
	Count  uint64
	LastPC uint32
	Raw    uint32
	Key    string
}

// New creates a new CPU instance
func New(b *bus.Bus) *CPU {
	cpu := &CPU{
		Bus:             b,
		TraceOut:        os.Stderr,
		MaxExceptionRun: 16,
	}

	cpu.Reset()
	b.SetTranslator(cpu.TranslateAddress)

	return cpu
}

// Reset places CPU into initial state
func (c *CPU) Reset() {

	// Clear general purpose registers
	for i := range c.Regs {
		c.Regs[i] = 0
	}

	// Clear Coprocessor 0 registers
	for i := range c.CP0 {
		c.CP0[i] = 0
	}
	for i := range c.FPR {
		c.FPR[i] = 0
	}
	c.FCSR = 0
	c.UserLocal = 0
	for i := range c.TLB {
		c.TLB[i] = TLBEntry{}
	}

	// Match the T23/XBurst PRId expected by the vendor kernel.
	c.CP0[CP0_PRID] = 0x00d00100

	// Advertise Config1 through Config.M. Linux checks this before reading
	// the cache/TLB geometry from CP0 Config select 1.
	c.CP0[CP0_CONFIG] = CONFIG_M | CONFIG_AR | CONFIG_K0
	c.CP0[CP0_RANDOM] = 31
	c.countBaseCycle = 0
	c.countBaseValue = 0
	c.compareSet = false

	// After reset a MIPS core starts in kernel mode with BEV set and the
	// error level flag asserted, so exceptions use the ROM vectors.
	c.CP0[CP0_STATUS] = STATUS_BEV | STATUS_ERL

	c.HI = 0
	c.LO = 0

	c.PC = c.ResetPC
	c.NextPC = c.ResetPC + 4
	c.CurrentPC = c.ResetPC

	c.InDelaySlot = false
	c.branchTaken = false

	c.Instruction = 0

	c.Running = false
	c.Waiting = false

	c.HaltReason = HaltNone
	c.HaltDetail = ""

	c.LLBit = false

	c.exceptionRun = 0

	c.Cycles = 0

	for i := range c.icache {
		c.icache[i].valid = false
	}
}

// Fetch reads the next instruction from memory and advances the program
// counters by one slot.
func (c *CPU) Fetch() uint32 {

	instruction := c.Bus.Read32(c.PC)

	c.Instruction = instruction

	// Advance the pipeline. Doing this before execution is what gives a
	// branch handler somewhere to write its target: it sets NextPC, and
	// the instruction now at PC (the delay slot) still runs first.
	c.CurrentPC = c.PC
	c.PC = c.NextPC
	c.NextPC = c.PC + 4

	return instruction
}

// Step executes one CPU cycle.
func (c *CPU) Step() {

	if !c.Running {
		return
	}

	// Watchdog check — must run even when WAITing so a watchdog timeout
	// can break out of a reboot spin loop.
	if c.WatchdogCheck != nil && c.WatchdogCheck() {
		return
	}

	if c.Waiting {
		if pending := c.updateInterruptPending(); pending != 0 {
			c.Waiting = false
			c.stepsSinceInterrupt = 0
			if c.interruptEnabled(pending) {
				c.takeInterrupt()
			}
			c.Cycles++
		} else if c.WakePending != nil && c.WakePending() {
			c.Waiting = false
			c.Cycles++
		} else if c.NextWakeCycle != nil && c.NextWakeCycle() > c.Cycles {
			c.Cycles = c.NextWakeCycle()
		} else {
			c.Cycles++
		}
		return
	}

	if !c.branchTaken && c.checkInterrupts() {
		c.stepsSinceInterrupt = 0
		c.Cycles++
		return
	}

	// Spin-loop fast-forward: if no interrupt has fired in a long time
	// and the watchdog is armed, the CPU is likely in the kernel reboot
	// spin loop (interrupts disabled). Jump directly to the watchdog
	// deadline instead of emulating billions of printk cycles.
	c.stepsSinceInterrupt++
	if c.stepsSinceInterrupt >= spinForwardThreshold && c.NextWakeCycle != nil {
		if next := c.NextWakeCycle(); next > c.Cycles {
			c.Cycles = next
		}
		c.stepsSinceInterrupt = 0
	}

	// The instruction about to run is a delay slot if the previous one
	// was a taken branch.
	c.InDelaySlot = c.branchTaken
	c.branchTaken = false

	pc := c.PC

	if !c.SingleStep && c.Breakpoints != nil && c.Breakpoints[pc] {
		c.Running = false
		return
	}

	// Address Error check for Fetch
	if !c.Bus.HasMapping(pc) {
		c.CurrentPC = pc
		if c.requiresTLB(pc) {
			if _, _, index := c.lookupTLB(pc, false); index >= 0 {
				c.exceptionNoRefill(EXC_TLBL, pc)
			} else {
				c.Exception(EXC_TLBL, pc)
			}
		} else {
			c.Exception(EXC_ADEL, pc)
		}
		if c.RecordHistory {
			c.currentMemAddr = 0
			c.currentMemVal = 0
			c.currentMemAccess = ""
			c.RecordHistoryEntry(pc, 0, c.InDelaySlot)
		}
		c.Cycles++
		return
	}

	// Reset memory transaction tracking
	c.currentMemAddr = 0
	c.currentMemVal = 0
	c.currentMemAccess = ""
	inDelaySlot := c.InDelaySlot

	raw := c.Fetch()

	icIndex := (pc >> 2) & 0xFFFF
	entry := &c.icache[icIndex]
	var inst Instruction
	if entry.valid && entry.raw == raw {
		inst = entry.inst
	} else {
		inst = Decode(raw)
		entry.valid = true
		entry.raw = raw
		entry.inst = inst
	}

	if c.Trace {
		marker := " "
		if c.InDelaySlot {
			// Mark delay slots so that traces of branchy code can be
			// read without mentally re-deriving the pipeline.
			marker = "+"
		}
		fmt.Fprintf(c.TraceOut, "[%08d]%s %08x: %08x  %s\n",
			c.Cycles, marker, pc, raw, Disassemble(raw, pc))
	}

	c.Execute(inst)

	if c.RecordHistory {
		c.RecordHistoryEntry(pc, raw, inDelaySlot)
	}

	c.Cycles++
}

func (c *CPU) checkInterrupts() bool {
	status := c.CP0[CP0_STATUS]
	if status&STATUS_IE == 0 || status&(STATUS_EXL|STATUS_ERL) != 0 {
		return false
	}
	if (c.Cycles & 31) != 0 {
		return false
	}
	pending := c.updateInterruptPending()
	if pending&status&STATUS_IM == 0 {
		return false
	}

	c.takeInterrupt()
	return true
}

func (c *CPU) updateInterruptPending() uint32 {
	pending := uint32(0)
	if c.InterruptPending != nil {
		pending = c.InterruptPending() & CAUSE_IP
	}
	if c.cp0TimerPending() {
		pending |= CAUSE_IP7
	}

	c.CP0[CP0_CAUSE] = (c.CP0[CP0_CAUSE] & ^CAUSE_IP) | pending
	return pending
}

func (c *CPU) interruptEnabled(pending uint32) bool {
	status := c.CP0[CP0_STATUS]
	if status&STATUS_IE == 0 || status&(STATUS_EXL|STATUS_ERL) != 0 {
		return false
	}
	if pending&status&STATUS_IM == 0 {
		return false
	}
	return true
}

func (c *CPU) takeInterrupt() {
	c.CurrentPC = c.PC
	c.InDelaySlot = false
	c.Exception(EXC_INT, 0)
}

func (c *CPU) cp0Count() uint32 {
	return c.countBaseValue + uint32((c.Cycles-c.countBaseCycle)/2)
}

func (c *CPU) cp0TimerPending() bool {
	if !c.compareSet {
		return false
	}
	return int32(c.cp0Count()-c.CP0[CP0_COMPARE]) >= 0
}

// Run executes the CPU loop.
func (c *CPU) Run() {

	c.Running = true

	for c.Running {
		c.Step()
	}
}

// Stop stops CPU execution.
func (c *CPU) Stop() {

	c.Running = false
}

func (c *CPU) Halt(reason HaltReason) {

	c.Running = false
	c.Waiting = false

	c.HaltReason = reason
}

// HaltWith stops the CPU and records why.
func (c *CPU) HaltWith(reason HaltReason, format string, args ...any) {

	c.Running = false
	c.Waiting = false

	c.HaltReason = reason

	c.HaltDetail = fmt.Sprintf(format, args...)
}

// branch redirects execution to target after the delay slot runs.
func (c *CPU) branch(target uint32) {
	c.NextPC = target
	c.branchTaken = true
}

// nullifyDelaySlot skips the instruction in the delay slot. The
// "branch likely" instructions do this when their condition is false.
func (c *CPU) nullifyDelaySlot() {
	c.PC = c.NextPC
	c.NextPC = c.PC + 4
}

// EnableRITrace enables logging of every Reserved Instruction exception
// to w and starts a fresh per-key tally. Pass nil to disable. The tally
// is kept in memory so it is lost if the process is killed without a
// clean shutdown; the per-event log written to w is durable as soon as
// each line is emitted.
func (c *CPU) EnableRITrace(w io.Writer) {
	c.RITraceOut = w
	if w != nil {
		c.riCounters = make(map[string]*RICounter)
	} else {
		c.riCounters = nil
	}
}

// reservedInstruction raises a Reserved Instruction exception and, when
// RI tracing is enabled, records the faulting word for offline analysis.
// Use this at every site where the emulator falls off the end of its
// instruction decode tables.
func (c *CPU) reservedInstruction(inst Instruction) {
	if c.RITraceOut != nil {
		key := fmt.Sprintf("op=0x%02x funct=0x%02x rs=%d rt=%d rd=%d shamt=%d",
			inst.Opcode, inst.Funct, inst.Rs, inst.Rt, inst.Rd, inst.Shamt)
		fmt.Fprintf(c.RITraceOut, "[%d] pc=0x%08x raw=0x%08x %s  %s\n",
			c.Cycles, c.CurrentPC, inst.Raw, key, Disassemble(inst.Raw, c.CurrentPC))
		if existing, ok := c.riCounters[key]; ok {
			existing.Count++
			existing.LastPC = c.CurrentPC
			existing.Raw = inst.Raw
		} else {
			c.riCounters[key] = &RICounter{
				Count:  1,
				LastPC: c.CurrentPC,
				Raw:    inst.Raw,
				Key:    key,
			}
		}
	}
	c.Exception(EXC_RI, 0)
}

// RICounters returns the accumulated Reserved Instruction tally sorted by
// descending count, or nil if RI tracing was never enabled.
func (c *CPU) RICounters() []RICounter {
	if c.riCounters == nil {
		return nil
	}
	out := make([]RICounter, 0, len(c.riCounters))
	for _, v := range c.riCounters {
		out = append(out, *v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Count > out[j].Count })
	return out
}

// Exception handles CPU exception processing: updates Cause, EPC, Status EXL,
// BadVAddr and jumps to the exception vector.
func (c *CPU) Exception(code uint8, badVAddr uint32) {
	c.exception(code, badVAddr, true)
}

func (c *CPU) exceptionNoRefill(code uint8, badVAddr uint32) {
	c.exception(code, badVAddr, false)
}

func (c *CPU) exception(code uint8, badVAddr uint32, allowRefill bool) {
	c.exceptionRun++
	if c.MaxExceptionRun > 0 && c.exceptionRun > c.MaxExceptionRun {
		// The handler itself is faulting. Report the original cause
		// rather than letting the core spin on the vector address.
		c.HaltWith(HaltExceptionStorm,
			"%d consecutive exceptions (last: %s at 0x%08X, vector 0x%08X unhandled)",
			c.exceptionRun, ExceptionName(code), c.CurrentPC, c.exceptionVector(code, false))
		return
	}

	status := c.CP0[CP0_STATUS]
	refill := allowRefill && (code == EXC_TLBL || code == EXC_TLBS) && status&STATUS_EXL == 0

	if code == EXC_MOD || code == EXC_ADEL || code == EXC_ADES || code == EXC_TLBL || code == EXC_TLBS {
		c.updateTLBExceptionState(badVAddr)
	}

	// If TraceADE is set, halt on the first userspace address error
	// (AdEL/AdES) so the caller can inspect the faulting instruction.
	if c.TraceADE && (code == EXC_ADEL || code == EXC_ADES) && c.CurrentPC < 0x80000000 {
		var sb strings.Builder
		fmt.Fprintf(&sb, "userspace %s at PC=0x%08X  inst=0x%08X  %s\n",
			ExceptionName(code), c.CurrentPC, c.Instruction,
			Disassemble(c.Instruction, c.CurrentPC))
		fmt.Fprintf(&sb, "BadVAddr=0x%08X  Status=0x%08X  Cause=0x%08X  UserLocal=0x%08X\n",
			badVAddr, c.CP0[CP0_STATUS], c.CP0[CP0_CAUSE], c.UserLocal)
		for i := 0; i < 32; i += 4 {
			fmt.Fprintf(&sb, "  r%-2d=%08X  r%-2d=%08X  r%-2d=%08X  r%-2d=%08X\n",
				i, c.Regs[i], i+1, c.Regs[i+1],
				i+2, c.Regs[i+2], i+3, c.Regs[i+3])
		}
		fmt.Fprintf(&sb, "  hi=%08X  lo=%08X  sp=%08X\n\n",
			c.HI, c.LO, c.Regs[29])
		fmt.Fprintln(&sb, "--- last 40 instructions ---")
		for _, e := range c.GetHistory() {
			marker := " "
			if e.InDelaySlot {
				marker = "+"
			}
			fmt.Fprintf(&sb, "  %s 0x%08X  %08X  %s\n",
				marker, e.PC, e.Instruction,
				Disassemble(e.Instruction, e.PC))
		}
		c.HaltWith(HaltExceptionStorm, "%s", sb.String())
		return
	}

	// EPC and the BD flag are only meaningful for the outermost
	// exception; a fault taken with EXL already set must not clobber the
	// original return address.
	if status&STATUS_EXL == 0 {

		if c.InDelaySlot {
			// A delay slot cannot be restarted on its own, so EPC points
			// at the branch and Cause.BD tells the handler why.
			c.CP0[CP0_EPC] = c.CurrentPC - 4
			c.CP0[CP0_CAUSE] |= CAUSE_BD
		} else {
			c.CP0[CP0_EPC] = c.CurrentPC
			c.CP0[CP0_CAUSE] &= ^CAUSE_BD
		}

		c.CP0[CP0_STATUS] |= STATUS_EXL
	}

	// Set Cause ExcCode bits 6:2
	c.CP0[CP0_CAUSE] = (c.CP0[CP0_CAUSE] & ^CAUSE_EXCCODE) | (uint32(code) << 2)

	vector := c.exceptionVector(code, refill)

	c.PC = vector
	c.NextPC = vector + 4

	// Entering the handler cancels any pending branch.
	c.branchTaken = false
	c.InDelaySlot = false
	c.Waiting = false
}

// exceptionVector returns the general exception vector address selected
// by the BEV bit in Status.
func (c *CPU) exceptionVector(code uint8, refill bool) uint32 {
	if (c.CP0[CP0_STATUS] & STATUS_BEV) != 0 {
		return 0xbfc00380
	}
	if refill {
		return 0x80000000
	}
	return 0x80000180
}

// retire records that an instruction completed without faulting, which
// resets the exception storm detector.
func (c *CPU) retire() {
	c.exceptionRun = 0
}

// ExceptionName returns the mnemonic for a MIPS exception code.
func ExceptionName(code uint8) string {
	switch code {
	case EXC_INT:
		return "Int"
	case EXC_MOD:
		return "Mod"
	case EXC_TLBL:
		return "TLBL"
	case EXC_TLBS:
		return "TLBS"
	case EXC_ADEL:
		return "AdEL"
	case EXC_ADES:
		return "AdES"
	case EXC_IBE:
		return "IBE"
	case EXC_DBE:
		return "DBE"
	case EXC_SYS:
		return "Sys"
	case EXC_BP:
		return "Bp"
	case EXC_RI:
		return "RI"
	case EXC_CPU:
		return "CpU"
	case EXC_OV:
		return "Ov"
	case EXC_TR:
		return "Tr"
	}
	return fmt.Sprintf("Exc%d", code)
}

type HistoryEntry struct {
	Cycle       uint64
	PC          uint32
	Instruction uint32
	MemAddr     uint32
	MemVal      uint32
	MemAccess   string // "R" (read) or "W" (write) or "" (none)
	InDelaySlot bool
}

func (c *CPU) RecordHistoryEntry(pc uint32, raw uint32, inDelaySlot bool) {
	c.History[c.HistoryIndex] = HistoryEntry{
		Cycle:       c.Cycles,
		PC:          pc,
		Instruction: raw,
		MemAddr:     c.currentMemAddr,
		MemVal:      c.currentMemVal,
		MemAccess:   c.currentMemAccess,
		InDelaySlot: inDelaySlot,
	}
	c.HistoryIndex = (c.HistoryIndex + 1) % HistorySize
	if c.HistoryIndex == 0 {
		c.HistoryFull = true
	}
}

func (c *CPU) GetHistory() []HistoryEntry {
	var entries []HistoryEntry
	limit := HistorySize
	start := 0
	size := c.HistoryIndex
	if c.HistoryFull {
		size = limit
		start = c.HistoryIndex
	}
	for i := 0; i < size; i++ {
		idx := (start + i) % limit
		entries = append(entries, c.History[idx])
	}
	return entries
}

func (c *CPU) CheckWatchpoint(addr uint32, length uint32, isWrite bool) {
	if len(c.Watchpoints) == 0 {
		return
	}
	accessEnd := addr + length
	for _, wp := range c.Watchpoints {
		wpEnd := wp.Addr + wp.Len
		if addr < wpEnd && wp.Addr < accessEnd {
			if (isWrite && (wp.Type == WatchWrite || wp.Type == WatchAccess)) ||
				(!isWrite && (wp.Type == WatchRead || wp.Type == WatchAccess)) {
				c.HitWatchpoint = wp.Addr
				c.Running = false
				return
			}
		}
	}
}

func (c *CPU) read8(addr uint32) byte {
	c.CheckWatchpoint(addr, 1, false)
	val := c.Bus.Read8(addr)
	c.currentMemAddr = addr
	c.currentMemVal = uint32(val)
	c.currentMemAccess = "R"
	return val
}

func (c *CPU) write8(addr uint32, val byte) {
	c.CheckWatchpoint(addr, 1, true)
	c.Bus.Write8(addr, val)
	c.currentMemAddr = addr
	c.currentMemVal = uint32(val)
	c.currentMemAccess = "W"
}

func (c *CPU) read16(addr uint32) uint16 {
	c.CheckWatchpoint(addr, 2, false)
	val := c.Bus.Read16(addr)
	c.currentMemAddr = addr
	c.currentMemVal = uint32(val)
	c.currentMemAccess = "R"
	return val
}

func (c *CPU) write16(addr uint32, val uint16) {
	c.CheckWatchpoint(addr, 2, true)
	c.Bus.Write16(addr, val)
	c.currentMemAddr = addr
	c.currentMemVal = uint32(val)
	c.currentMemAccess = "W"
}

func (c *CPU) read32(addr uint32) uint32 {
	c.CheckWatchpoint(addr, 4, false)
	val := c.Bus.Read32(addr)
	c.currentMemAddr = addr
	c.currentMemVal = val
	c.currentMemAccess = "R"
	return val
}

func (c *CPU) write32(addr uint32, val uint32) {
	c.CheckWatchpoint(addr, 4, true)
	c.Bus.Write32(addr, val)
	c.currentMemAddr = addr
	c.currentMemVal = val
	c.currentMemAccess = "W"
}
