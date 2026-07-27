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
		Opcode: 0,
		Funct:  0,
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
