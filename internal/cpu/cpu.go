package cpu

import (
	"github.com/HritikR/t23emu/internal/bus"
)

type CPU struct {
	// General purpose registers
	// MIPS has 32 registers: $zero-$ra
	Regs [32]uint32

	// Program counter
	PC uint32

	// Current fetched instruction
	Instruction uint32

	// Memory interface
	Bus *bus.Bus

	// CPU execution state
	Running bool

	// CPU halt status
	HaltReason HaltReason

	// Cycle counter
	Cycles uint64

	// Reset PC address
	ResetPC uint32

	// Coprocessor 0 registers
	CP0 [32]uint32
}

// New creates a new CPU instance
func New(b *bus.Bus) *CPU {
	cpu := &CPU{
		Bus: b,
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
	c.CP0[15] = 0x00018000

	c.PC = c.ResetPC

	c.Instruction = 0

	c.Running = false

	c.HaltReason = HaltNone

	c.Cycles = 0
}

// Fetch reads the next instruction from memory
func (c *CPU) Fetch() uint32 {

	instruction := c.Bus.Read32(c.PC)

	c.Instruction = instruction

	// MIPS instructions are 4 bytes
	c.PC += 4

	return instruction
}

// Step executes one CPU cycle.
func (c *CPU) Step() {

	if !c.Running {
		return
	}

	// Address Error check for Fetch
	if !c.Bus.HasMapping(c.PC) {
		c.Exception(EXC_ADEL, c.PC)
		c.Cycles++
		return
	}

	raw := c.Fetch()

	inst := Decode(raw)

	c.Execute(inst)

	c.Cycles++
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

// Exception handles CPU exception processing: updates Cause, EPC, Status EXL, BadVAddr and jumps to vector.
func (c *CPU) Exception(code uint8, badVAddr uint32) {
	if code == EXC_ADEL || code == EXC_ADES {
		c.CP0[CP0_BADVADDR] = badVAddr
	}

	// Status register is CP0[12]
	if (c.CP0[CP0_STATUS] & 0x2) == 0 {
		// Set EPC CP0[14]. If EXC_ADEL on Fetch, c.PC points to invalid address.
		// Else inside execution, EPC = c.PC - 4.
		if code == EXC_ADEL && badVAddr == c.PC {
			c.CP0[CP0_EPC] = c.PC
		} else {
			c.CP0[CP0_EPC] = c.PC - 4
		}
		// Set EXL bit (bit 1)
		c.CP0[CP0_STATUS] |= 0x2
	}

	// Set Cause ExcCode bits 6:2
	c.CP0[CP0_CAUSE] = (c.CP0[CP0_CAUSE] & ^uint32(0x7C)) | (uint32(code) << 2)

	// Jump to exception vector
	// BEV is bit 22 (0x00400000) of Status
	if (c.CP0[CP0_STATUS] & 0x00400000) != 0 {
		c.PC = 0xbfc00380
	} else {
		c.PC = 0x80000180
	}
}
