package cpu

import (
	"testing"

	"github.com/HritikR/t23emu/internal/bus"
	"github.com/HritikR/t23emu/internal/memory"
)

func TestExecuteNOP(t *testing.T) {

	cpu := createTestCPU()

	before := cpu.Regs

	inst := Instruction{
		Opcode: OP_SPECIAL,
		Funct:  FUNCT_SLL,
	}

	cpu.Execute(inst)

	for i := range before {

		if cpu.Regs[i] != before[i] {

			t.Fatalf(
				"NOP changed register %d",
				i,
			)
		}
	}
}

func TestExecuteADDI(t *testing.T) {

	cpu := createTestCPU()

	inst := Instruction{
		Opcode:    OP_ADDI,
		Rs:        0,
		Rt:        8,
		Immediate: 100,
	}

	cpu.Execute(inst)

	value := cpu.ReadRegister(8)

	if value != 100 {

		t.Fatalf(
			"expected 100 got %d",
			value,
		)
	}
}

func TestExecuteADD(t *testing.T) {

	cpu := createTestCPU()

	cpu.WriteRegister(
		9,
		10,
	)

	cpu.WriteRegister(
		10,
		20,
	)

	inst := Instruction{
		Opcode: OP_SPECIAL,
		Rs:     9,
		Rt:     10,
		Rd:     8,
		Funct:  FUNCT_ADD,
	}

	cpu.Execute(inst)

	value := cpu.ReadRegister(8)

	if value != 30 {

		t.Fatalf(
			"expected 30 got %d",
			value,
		)
	}
}

func TestExecuteADDZeroRegister(t *testing.T) {

	cpu := createTestCPU()

	cpu.WriteRegister(
		1,
		10,
	)

	cpu.WriteRegister(
		2,
		20,
	)

	inst := Instruction{
		Opcode: OP_SPECIAL,
		Rs:     1,
		Rt:     2,
		Rd:     0,
		Funct:  FUNCT_ADD,
	}

	cpu.Execute(inst)

	if cpu.ReadRegister(0) != 0 {

		t.Fatalf(
			"$zero register changed",
		)
	}
}

func TestExecuteADDIPositive(t *testing.T) {

	cpu := createTestCPU()

	cpu.WriteRegister(
		1,
		50,
	)

	inst := Instruction{
		Opcode:    OP_ADDI,
		Rs:        1,
		Rt:        2,
		Immediate: 25,
	}

	cpu.Execute(inst)

	if cpu.ReadRegister(2) != 75 {

		t.Fatalf(
			"expected 75 got %d",
			cpu.ReadRegister(2),
		)
	}
}

func TestExecuteADDINegativeImmediate(t *testing.T) {

	cpu := createTestCPU()

	cpu.WriteRegister(
		1,
		100,
	)

	inst := Instruction{
		Opcode:    OP_ADDI,
		Rs:        1,
		Rt:        2,
		Immediate: uint16(0xFFFF), // -1
	}

	cpu.Execute(inst)

	if cpu.ReadRegister(2) != 99 {

		t.Fatalf(
			"expected 99 got %d",
			cpu.ReadRegister(2),
		)
	}
}

func TestExecuteLW(t *testing.T) {

	ram := memory.NewRAM(1024)

	b := bus.New(ram)

	cpu := New(b)

	ram.Write32(
		100,
		0x12345678,
	)

	cpu.WriteRegister(
		1,
		100,
	)

	inst := Instruction{
		Opcode:    OP_LW,
		Rs:        1,
		Rt:        2,
		Immediate: 0,
	}

	cpu.Execute(inst)

	value := cpu.ReadRegister(2)

	if value != 0x12345678 {

		t.Fatalf(
			"expected 0x12345678 got 0x%08X",
			value,
		)
	}
}

func TestExecuteSW(t *testing.T) {

	ram := memory.NewRAM(1024)

	b := bus.New(ram)

	cpu := New(b)

	cpu.WriteRegister(
		1,
		200,
	)

	cpu.WriteRegister(
		2,
		0xDEADBEEF,
	)

	inst := Instruction{
		Opcode:    OP_SW,
		Rs:        1,
		Rt:        2,
		Immediate: 0,
	}

	cpu.Execute(inst)

	value := ram.Read32(200)

	if value != 0xDEADBEEF {

		t.Fatalf(
			"expected 0xDEADBEEF got 0x%08X",
			value,
		)
	}
}

func TestCPUFullADDIFlow(t *testing.T) {

	ram := memory.NewRAM(1024)

	b := bus.New(ram)

	// addi $t0,$zero,42
	//
	// 0x2008002A

	ram.Write32(
		0,
		0x2008002A,
	)

	cpu := New(b)

	cpu.Running = true

	cpu.Step()

	value := cpu.ReadRegister(8)

	if value != 42 {

		t.Fatalf(
			"expected 42 got %d",
			value,
		)
	}
}

func TestCPUFullADDFlow(t *testing.T) {

	ram := memory.NewRAM(1024)

	b := bus.New(ram)

	// add $t0,$t1,$t2
	//
	// 0x012A4020

	ram.Write32(
		0,
		0x012A4020,
	)

	cpu := New(b)

	cpu.WriteRegister(
		9,
		10,
	)

	cpu.WriteRegister(
		10,
		20,
	)

	cpu.Running = true

	cpu.Step()

	value := cpu.ReadRegister(8)

	if value != 30 {

		t.Fatalf(
			"expected 30 got %d",
			value,
		)
	}
}
