package machine

import (
	"github.com/HritikR/t23emu/internal/bus"
	"github.com/HritikR/t23emu/internal/cpu"
	"github.com/HritikR/t23emu/internal/device"
	"github.com/HritikR/t23emu/internal/memory"
)

const ROMStart uint32 = 0xbfc00000
const UARTStart uint32 = 0x10000000
const UARTEnd uint32 = 0x1000000F

type Machine struct {
	CPU *cpu.CPU

	RAM *memory.RAM

	ROM *device.ROM

	UART *device.UART

	Bus *bus.Bus
}

// New creates a new T23 emulator machine.
func New(ramSize uint32, romData []byte) *Machine {

	ram := memory.NewRAM(
		ramSize,
	)

	b := bus.New()

	b.Map(
		0x00000000,
		ramSize-1,
		ram,
	)

	var rom *device.ROM
	if len(romData) > 0 {
		rom = device.NewROM(romData)
		b.Map(
			ROMStart,
			ROMStart+uint32(len(romData))-1,
			rom,
		)
	}

	uart := device.NewUART(nil)
	b.Map(
		UARTStart,
		UARTEnd,
		uart,
	)

	c := cpu.New(
		b,
	)

	if rom != nil {
		resetPC := ROMStart
		// Detect Ingenic boot header signature (typically 2KB offset 0x800)
		if len(romData) > 0x800 && romData[4] == 0x02 && romData[5] == 0x55 && romData[6] == 0xAA && romData[7] == 0x55 && romData[8] == 0xAA {
			resetPC = ROMStart + 0x800
		}
		c.ResetPC = resetPC
		c.Reset()
	}

	return &Machine{
		CPU:  c,
		RAM:  ram,
		ROM:  rom,
		UART: uart,
		Bus:  b,
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
func (m *Machine) Run(maxCycles uint64) uint64 {

	m.CPU.Running = true

	start := m.CPU.Cycles

	for m.CPU.Running {

		if m.CPU.Cycles-start >= maxCycles {
			break
		}

		m.CPU.Step()
	}

	m.CPU.Stop()

	return m.CPU.Cycles - start
}
