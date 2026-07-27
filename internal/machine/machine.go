package machine

import (
	"github.com/HritikR/t23emu/internal/bus"
	"github.com/HritikR/t23emu/internal/cpu"
	"github.com/HritikR/t23emu/internal/device"
	"github.com/HritikR/t23emu/internal/memory"
)

const (
	ROMStart  uint32 = 0x1fc00000 // Physical ROM start
	CPMStart  uint32 = 0x10000000
	CPMEnd    uint32 = 0x1000FFFF
	GPIOStart uint32 = 0x10010000
	GPIOEnd   uint32 = 0x1001FFFF
	UARTStart uint32 = 0x10030000 // Ingenic UART standard physical address
	UARTEnd   uint32 = 0x1003FFFF
)

type stubDevice struct{}

func (s *stubDevice) Read8(addr uint32) byte             { return 0 }
func (s *stubDevice) Write8(addr uint32, value byte)     {}
func (s *stubDevice) Read32(addr uint32) uint32          { return 0 }
func (s *stubDevice) Write32(addr uint32, value uint32)  {}

var _ device.Device = (*stubDevice)(nil)

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

	isIngenicSPL := len(romData) > 0x800 && romData[4] == 0x02 && romData[5] == 0x55 && romData[6] == 0xAA && romData[7] == 0x55 && romData[8] == 0xAA
	isFirmwareDump := len(romData) == 8388608

	// Map specific devices first so they take precedence over ROM
	if !isIngenicSPL || isFirmwareDump {
		// Map Clock and Power Management (CPM) physical stub
		b.Map(
			CPMStart,
			CPMEnd,
			&stubDevice{},
		)
	}

	// Map GPIO physical stub
	b.Map(
		GPIOStart,
		GPIOEnd,
		&stubDevice{},
	)

	uart := device.NewUART(nil)
	b.Map(
		UARTStart,
		UARTEnd,
		uart,
	)

	var rom *device.ROM
	romStart := ROMStart
	resetPC := uint32(0xbfc00000)
	var sram *memory.RAM
	var sramSize uint32

	if isIngenicSPL {
		romStart = 0x10000000 // Map SPL at physical SRAM base
		resetPC = 0xb0000800  // Reset PC inside SRAM (kseg1)
		sramSize = 262144     // 256KB SRAM
		if uint32(len(romData)) > sramSize {
			sramSize = uint32(len(romData))
		}
	}

	if len(romData) > 0 {
		if isIngenicSPL {
			sram = memory.NewRAM(sramSize)
			for i, val := range romData {
				sram.Write8(uint32(i), val)
			}
			b.Map(
				romStart,
				romStart+sramSize-1,
				sram,
			)
		} else {
			rom = device.NewROM(romData)
			b.Map(
				romStart,
				romStart+uint32(len(romData))-1,
				rom,
			)
		}
	}

	c := cpu.New(
		b,
	)

	if len(romData) > 0 {
		c.ResetPC = resetPC
		c.Reset()
		if isIngenicSPL {
			// Initialize stack pointer to standard top of SRAM (virtual 0xb0010000)
			c.WriteRegister(29, 0xb0010000)
			// Initialize return address register ($ra) to BootROM handoff hook address
			c.WriteRegister(31, 0x90000000)
		}
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
