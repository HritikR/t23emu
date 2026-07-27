package machine

import (
	"testing"

	"github.com/HritikR/t23emu/internal/cpu"
)

func TestMachineCreation(t *testing.T) {

	m := New(1024, nil)

	if m.CPU == nil {
		t.Fatalf(
			"CPU was not created",
		)
	}

	if m.Bus == nil {
		t.Fatalf(
			"Bus was not created",
		)
	}

	if m.RAM == nil {
		t.Fatalf(
			"RAM was not created",
		)
	}

	if m.UART == nil {
		t.Fatalf(
			"UART was not created",
		)
	}
}

func TestMachineLoadProgram(t *testing.T) {

	m := New(1024, nil)

	program := []uint32{

		cpu.EncodeI(
			cpu.OP_ADDI,
			0,
			8,
			42,
		),
	}

	m.LoadProgram(
		0,
		program,
	)

	value := m.RAM.Read32(0)

	expected := cpu.EncodeI(
		cpu.OP_ADDI,
		0,
		8,
		42,
	)

	if value != expected {

		t.Fatalf(
			"expected instruction 0x%08X got 0x%08X",
			expected,
			value,
		)
	}
}

func TestMachineReset(t *testing.T) {

	m := New(1024, nil)

	m.CPU.WriteRegister(
		8,
		123,
	)

	m.CPU.PC = 0x100

	m.CPU.Running = true

	m.Reset()

	if m.CPU.ReadRegister(8) != 0 {

		t.Fatalf(
			"register was not reset",
		)
	}

	if m.CPU.PC != 0 {

		t.Fatalf(
			"PC was not reset",
		)
	}

	if m.CPU.Running {

		t.Fatalf(
			"CPU should not be running after reset",
		)
	}
}

func TestMachineRunADDIProgram(t *testing.T) {

	m := New(1024, nil)

	program := []uint32{

		// addi $t0,$zero,42
		cpu.EncodeI(
			cpu.OP_ADDI,
			0,
			8,
			42,
		),
	}

	m.LoadProgram(
		0,
		program,
	)

	cycles := m.Run(1)

	if cycles != 1 {

		t.Fatalf(
			"expected 1 cycle got %d",
			cycles,
		)
	}

	value := m.CPU.ReadRegister(8)

	if value != 42 {

		t.Fatalf(
			"expected register value 42 got %d",
			value,
		)
	}
}

func TestMachineRunMultipleInstructions(t *testing.T) {

	m := New(1024, nil)

	program := []uint32{

		// t0 = 10
		cpu.EncodeI(
			cpu.OP_ADDI,
			0,
			8,
			10,
		),

		// t1 = 20
		cpu.EncodeI(
			cpu.OP_ADDI,
			0,
			9,
			20,
		),

		// t2 = t0 + t1
		cpu.EncodeR(
			cpu.OP_SPECIAL,
			8,
			9,
			10,
			0,
			cpu.FUNCT_ADD,
		),
	}

	m.LoadProgram(
		0,
		program,
	)

	cycles := m.Run(3)

	if cycles != 3 {

		t.Fatalf(
			"expected 3 cycles got %d",
			cycles,
		)
	}

	value := m.CPU.ReadRegister(10)

	if value != 30 {

		t.Fatalf(
			"expected 30 got %d",
			value,
		)
	}
}

func TestMachineRunWithCycleLimit(t *testing.T) {

	m := New(1024, nil)

	program := []uint32{

		cpu.EncodeI(
			cpu.OP_ADDI,
			8,
			8,
			1,
		),
	}

	m.LoadProgram(
		0,
		program,
	)

	cycles := m.Run(5)

	if cycles != 5 {

		t.Fatalf(
			"expected 5 cycles got %d",
			cycles,
		)
	}

	if m.CPU.Cycles != 5 {

		t.Fatalf(
			"expected CPU cycles=5 got %d",
			m.CPU.Cycles,
		)
	}
}

func TestMachineWithROM(t *testing.T) {

	romData := []byte{
		0x78,
		0x56,
		0x34,
		0x12,
	}

	m := New(1024, romData)

	if m.ROM == nil {
		t.Fatalf(
			"ROM was not created",
		)
	}

	value := m.Bus.Read32(ROMStart)

	if value != 0x12345678 {
		t.Fatalf(
			"expected 0x12345678, got 0x%08X",
			value,
		)
	}

	m.Bus.Write32(
		ROMStart,
		0x99999999,
	)

	valueAfterWrite := m.Bus.Read32(ROMStart)

	if valueAfterWrite != 0x12345678 {
		t.Fatalf(
			"ROM data changed after write: got 0x%08X",
			valueAfterWrite,
		)
	}
}

func TestMachineBootFromROM(t *testing.T) {

	// Build a small program to run from ROM:
	// 1. addi $t0, $zero, 42  => Write 42 to $t0 (register 8)
	// 2. sw $t0, 100($zero)   => Write value of $t0 to RAM address 100
	program := []uint32{
		cpu.EncodeI(
			cpu.OP_ADDI,
			0,
			8,
			42,
		),
		cpu.EncodeI(
			cpu.OP_SW,
			0,
			8,
			100,
		),
	}

	// Convert program to little-endian bytes for ROM
	romData := make([]byte, len(program)*4)
	for i, inst := range program {
		romData[i*4] = byte(inst)
		romData[i*4+1] = byte(inst >> 8)
		romData[i*4+2] = byte(inst >> 16)
		romData[i*4+3] = byte(inst >> 24)
	}

	m := New(1024, romData)

	// Verify CPU PC is initialized to virtual boot address 0xbfc00000
	if m.CPU.PC != 0xbfc00000 {
		t.Fatalf(
			"expected PC to be 0xbfc00000, got 0x%08X",
			m.CPU.PC,
		)
	}

	// Run the CPU for 2 cycles
	cycles := m.Run(2)

	if cycles != 2 {
		t.Fatalf(
			"expected to run 2 cycles, got %d",
			cycles,
		)
	}

	// Verify the result of executing ROM instructions:
	// Register 8 should be 42
	regVal := m.CPU.ReadRegister(8)
	if regVal != 42 {
		t.Fatalf(
			"expected register 8 to be 42, got %d",
			regVal,
		)
	}

	// RAM address 100 should contain 42
	ramVal := m.RAM.Read32(100)
	if ramVal != 42 {
		t.Fatalf(
			"expected RAM address 100 to contain 42, got %d",
			ramVal,
		)
	}
}

func TestMachineUARTConsole(t *testing.T) {

	// Build a program to write to UART:
	// 1. lui $t1, 0x1000       => Set $t1 (register 9) to UART base address (0x10000000)
	// 2. addi $t0, $zero, 0x41  => Load 'A' into $t0 (register 8)
	// 3. sw $t0, 0($t1)        => Write 'A' to UART TX
	// 4. addi $t0, $zero, 0x42  => Load 'B' into $t0 (register 8)
	// 5. sw $t0, 0($t1)        => Write 'B' to UART TX
	program := []uint32{
		cpu.EncodeI(
			cpu.OP_LUI,
			0,
			9,
			0x1003,
		),
		cpu.EncodeI(
			cpu.OP_ADDI,
			0,
			8,
			0x41,
		),
		cpu.EncodeI(
			cpu.OP_SW,
			9,
			8,
			0,
		),
		cpu.EncodeI(
			cpu.OP_ADDI,
			0,
			8,
			0x42,
		),
		cpu.EncodeI(
			cpu.OP_SW,
			9,
			8,
			0,
		),
	}

	m := New(1024, nil)
	m.LoadProgram(0, program)

	// Run the program (5 instructions)
	cycles := m.Run(5)

	if cycles != 5 {
		t.Fatalf(
			"expected 5 cycles, got %d",
			cycles,
		)
	}

	// Verify the captured UART output
	output := m.UART.GetCapturedOutput()
	if output != "AB" {
		t.Fatalf(
			"expected UART output 'AB', got '%s'",
			output,
		)
	}
}
