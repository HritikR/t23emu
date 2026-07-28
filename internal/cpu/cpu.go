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

	// InterruptPending returns CP0 Cause.IP bits currently asserted by
	// external interrupt hardware.
	InterruptPending func() uint32

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
}

// New creates a new CPU instance
func New(b *bus.Bus) *CPU {
	cpu := &CPU{
		Bus:             b,
		TraceOut:        os.Stderr,
		MaxExceptionRun: 16,
	}

	cpu.Reset()

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

	// Initialize Processor Identification (PRId) register at index 15
	// A standard MIPS32 processor value.
	c.CP0[CP0_PRID] = 0x00018000

	// Advertise Config1 through Config.M. Linux checks this before reading
	// the cache/TLB geometry from CP0 Config select 1.
	c.CP0[CP0_CONFIG] = CONFIG_M | CONFIG_K0

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

	if !c.branchTaken && c.checkInterrupts() {
		c.Cycles++
		return
	}

	// Address Error check for Fetch
	if !c.Bus.HasMapping(c.PC) {
		c.CurrentPC = c.PC
		c.Exception(EXC_ADEL, c.PC)
		c.Cycles++
		return
	}

	// The instruction about to run is a delay slot if the previous one
	// was a taken branch.
	c.InDelaySlot = c.branchTaken
	c.branchTaken = false

	pc := c.PC
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

	c.Cycles++
}

func (c *CPU) checkInterrupts() bool {
	pending := uint32(0)
	if c.InterruptPending != nil {
		pending = c.InterruptPending() & CAUSE_IP
	}

	c.CP0[CP0_CAUSE] = (c.CP0[CP0_CAUSE] & ^CAUSE_IP) | pending

	status := c.CP0[CP0_STATUS]
	if status&STATUS_IE == 0 || status&(STATUS_EXL|STATUS_ERL) != 0 {
		return false
	}
	if pending&status&STATUS_IM == 0 {
		return false
	}

	c.CurrentPC = c.PC
	c.InDelaySlot = false
	c.Exception(EXC_INT, 0)
	return true
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

	c.HaltReason = reason
}

// HaltWith stops the CPU and records why.
func (c *CPU) HaltWith(reason HaltReason, format string, args ...any) {

	c.Running = false

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

	c.exceptionRun++
	if c.MaxExceptionRun > 0 && c.exceptionRun > c.MaxExceptionRun {
		// The handler itself is faulting. Report the original cause
		// rather than letting the core spin on the vector address.
		c.HaltWith(HaltExceptionStorm,
			"%d consecutive exceptions (last: %s at 0x%08X, vector 0x%08X unhandled)",
			c.exceptionRun, ExceptionName(code), c.CurrentPC, c.exceptionVector())
		return
	}

	if code == EXC_ADEL || code == EXC_ADES {
		c.CP0[CP0_BADVADDR] = badVAddr
	}

	// EPC and the BD flag are only meaningful for the outermost
	// exception; a fault taken with EXL already set must not clobber the
	// original return address.
	if (c.CP0[CP0_STATUS] & STATUS_EXL) == 0 {

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

	vector := c.exceptionVector()

	c.PC = vector
	c.NextPC = vector + 4

	// Entering the handler cancels any pending branch.
	c.branchTaken = false
	c.InDelaySlot = false
}

// exceptionVector returns the general exception vector address selected
// by the BEV bit in Status.
func (c *CPU) exceptionVector() uint32 {
	if (c.CP0[CP0_STATUS] & STATUS_BEV) != 0 {
		return 0xbfc00380
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
