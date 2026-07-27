package cpu

import (
	"testing"

	"github.com/HritikR/t23emu/internal/bus"
	"github.com/HritikR/t23emu/internal/memory"
)

func createCPUWithRAM() (*CPU, *memory.RAM) {

	ram := memory.NewRAM(1024)

	b := bus.New()

	b.Map(
		0x00000000,
		0x000003FF,
		ram,
	)

	return New(b), ram
}

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

	if cpu.ReadRegister(8) != 100 {

		t.Fatalf(
			"expected 100 got %d",
			cpu.ReadRegister(8),
		)
	}
}

func TestExecuteADD(t *testing.T) {

	cpu := createTestCPU()

	cpu.WriteRegister(9, 10)
	cpu.WriteRegister(10, 20)

	inst := Instruction{
		Opcode: OP_SPECIAL,
		Rs:     9,
		Rt:     10,
		Rd:     8,
		Funct:  FUNCT_ADD,
	}

	cpu.Execute(inst)

	if cpu.ReadRegister(8) != 30 {

		t.Fatalf(
			"expected 30 got %d",
			cpu.ReadRegister(8),
		)
	}
}

func TestExecuteADDZeroRegister(t *testing.T) {

	cpu := createTestCPU()

	cpu.WriteRegister(1, 10)
	cpu.WriteRegister(2, 20)

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

	cpu.WriteRegister(1, 50)

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

	cpu.WriteRegister(1, 100)

	inst := Instruction{
		Opcode:    OP_ADDI,
		Rs:        1,
		Rt:        2,
		Immediate: uint16(0xFFFF),
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

	cpu, ram := createCPUWithRAM()

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

	if cpu.ReadRegister(2) != 0x12345678 {

		t.Fatalf(
			"expected 0x12345678 got 0x%08X",
			cpu.ReadRegister(2),
		)
	}
}

func TestExecuteSW(t *testing.T) {

	cpu, ram := createCPUWithRAM()

	cpu.WriteRegister(1, 200)
	cpu.WriteRegister(2, 0xDEADBEEF)

	inst := Instruction{
		Opcode:    OP_SW,
		Rs:        1,
		Rt:        2,
		Immediate: 0,
	}

	cpu.Execute(inst)

	if ram.Read32(200) != 0xDEADBEEF {

		t.Fatalf(
			"expected 0xDEADBEEF got 0x%08X",
			ram.Read32(200),
		)
	}
}

func TestCPUFullADDIFlow(t *testing.T) {

	cpu, ram := createCPUWithRAM()

	ram.Write32(
		0,
		EncodeI(
			OP_ADDI,
			0,
			8,
			42,
		),
	)

	cpu.Running = true

	cpu.Step()

	if cpu.ReadRegister(8) != 42 {

		t.Fatalf(
			"expected 42 got %d",
			cpu.ReadRegister(8),
		)
	}
}

func TestCPUFullADDFlow(t *testing.T) {

	cpu, ram := createCPUWithRAM()

	ram.Write32(
		0,
		EncodeR(
			OP_SPECIAL,
			9,
			10,
			8,
			0,
			FUNCT_ADD,
		),
	)

	cpu.WriteRegister(9, 10)
	cpu.WriteRegister(10, 20)

	cpu.Running = true

	cpu.Step()

	if cpu.ReadRegister(8) != 30 {

		t.Fatalf(
			"expected 30 got %d",
			cpu.ReadRegister(8),
		)
	}
}
