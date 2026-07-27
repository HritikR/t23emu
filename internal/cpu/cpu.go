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
// Execution logic will be added later.
func (c *CPU) Step() {

	if !c.Running {
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
