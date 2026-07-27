package machine

import (
	"testing"

	"github.com/HritikR/t23emu/internal/cpu"
)

func TestMachineCreation(t *testing.T) {

	m := New(1024)

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
}

func TestMachineLoadProgram(t *testing.T) {

	m := New(1024)

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

	m := New(1024)

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

	m := New(1024)

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

	m := New(1024)

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

	m := New(1024)

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
