package machine

import (
	"github.com/HritikR/t23emu/internal/bus"
	"github.com/HritikR/t23emu/internal/cpu"
	"github.com/HritikR/t23emu/internal/memory"
)

type Machine struct {
	CPU *cpu.CPU

	RAM *memory.RAM

	Bus *bus.Bus
}

// New creates a new T23 emulator machine.
func New(ramSize uint32) *Machine {

	ram := memory.NewRAM(
		ramSize,
	)

	b := bus.New(
		ram,
	)

	c := cpu.New(
		b,
	)

	return &Machine{
		CPU: c,
		RAM: ram,
		Bus: b,
	}
}

// Reset resets the complete machine state.
func (m *Machine) Reset() {

	m.CPU.Reset()

}

// LoadProgram copies a program into RAM.
//
// address:
//
//	Starting memory address
//
// program:
//
//	Slice of 32-bit MIPS instructions
func (m *Machine) LoadProgram(
	address uint32,
	program []uint32,
) {

	for i, instruction := range program {

		offset := address + uint32(i*4)

		m.RAM.Write32(
			offset,
			instruction,
		)
	}
}

// Run executes the CPU for a number of cycles.
//
// This prevents tests from accidentally creating
// infinite loops.
func (m *Machine) Run(cycles int) {

	m.CPU.Running = true

	for i := 0; i < cycles; i++ {

		m.CPU.Step()

	}

	m.CPU.Stop()
}
