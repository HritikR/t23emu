package cpu

import (
	"fmt"
	"io"
	"os"

	"github.com/HritikR/t23emu/internal/bus"
)

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

	// InterruptPending returns CP0 Cause.IP bits currently asserted by
	// external interrupt hardware.
	InterruptPending func() uint32

	// WakePending reports implementation-specific activity that can resume
	// the core from WAIT without necessarily being a deliverable interrupt.
	WakePending func() bool

	// LLBit is the load-linked bit set by LL and tested by SC.
	LLBit bool

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

	// Instruction execution history (saved last 40 instructions)
	History          [40]HistoryEntry
	HistoryIndex     int
	HistoryFull      bool
	RecordHistory    bool
	currentMemAddr   uint32
	currentMemVal    uint32
	currentMemAccess string
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

	if c.Waiting {
		if pending := c.updateInterruptPending(); pending != 0 {
			c.Waiting = false
			if c.interruptEnabled(pending) {
				c.takeInterrupt()
			}
		} else if c.WakePending != nil && c.WakePending() {
			c.Waiting = false
		}
		c.Cycles++
		return
	}

	if !c.branchTaken && c.checkInterrupts() {
		c.Cycles++
		return
	}

	// The instruction about to run is a delay slot if the previous one
	// was a taken branch.
	c.InDelaySlot = c.branchTaken
	c.branchTaken = false

	pc := c.PC

	// Address Error check for Fetch
	if !c.Bus.HasMapping(pc) {
		c.CurrentPC = pc
		if isTLBMappedSegment(pc) {
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

	inst := Decode(raw)

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
	pending := c.updateInterruptPending()
	if !c.interruptEnabled(pending) {
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

	if code == EXC_ADEL || code == EXC_ADES || code == EXC_TLBL || code == EXC_TLBS {
		c.updateTLBExceptionState(badVAddr)
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
	c.HistoryIndex = (c.HistoryIndex + 1) % 40
	if c.HistoryIndex == 0 {
		c.HistoryFull = true
	}
}

func (c *CPU) GetHistory() []HistoryEntry {
	var entries []HistoryEntry
	limit := 40
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

func (c *CPU) read8(addr uint32) byte {
	val := c.Bus.Read8(addr)
	c.currentMemAddr = addr
	c.currentMemVal = uint32(val)
	c.currentMemAccess = "R"
	return val
}

func (c *CPU) write8(addr uint32, val byte) {
	c.Bus.Write8(addr, val)
	c.currentMemAddr = addr
	c.currentMemVal = uint32(val)
	c.currentMemAccess = "W"
}

func (c *CPU) read16(addr uint32) uint16 {
	val := c.Bus.Read16(addr)
	c.currentMemAddr = addr
	c.currentMemVal = uint32(val)
	c.currentMemAccess = "R"
	return val
}

func (c *CPU) write16(addr uint32, val uint16) {
	c.Bus.Write16(addr, val)
	c.currentMemAddr = addr
	c.currentMemVal = uint32(val)
	c.currentMemAccess = "W"
}

func (c *CPU) read32(addr uint32) uint32 {
	val := c.Bus.Read32(addr)
	c.currentMemAddr = addr
	c.currentMemVal = val
	c.currentMemAccess = "R"
	return val
}

func (c *CPU) write32(addr uint32, val uint32) {
	c.Bus.Write32(addr, val)
	c.currentMemAddr = addr
	c.currentMemVal = val
	c.currentMemAccess = "W"
}
