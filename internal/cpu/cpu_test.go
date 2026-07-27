package cpu

import (
	"testing"

	"github.com/HritikR/t23emu/internal/bus"
	"github.com/HritikR/t23emu/internal/memory"
)

func createTestCPU() *CPU {

	ram := memory.NewRAM(1024)

	b := bus.New(ram)

	cpu := New(b)

	return cpu
}

func TestCPUReset(t *testing.T) {

	cpu := createTestCPU()

	// Modify state
	cpu.Regs[1] = 123
	cpu.PC = 0x1000
	cpu.Instruction = 0xDEADBEEF
	cpu.Running = true

	// Reset
	cpu.Reset()

	if cpu.PC != 0 {
		t.Fatalf(
			"expected PC=0, got 0x%08X",
			cpu.PC,
		)
	}

	if cpu.Instruction != 0 {
		t.Fatalf(
			"expected instruction=0, got 0x%08X",
			cpu.Instruction,
		)
	}

	if cpu.Running {
		t.Fatalf(
			"expected CPU stopped after reset",
		)
	}

	for i, reg := range cpu.Regs {

		if reg != 0 {
			t.Fatalf(
				"register %d not cleared: %d",
				i,
				reg,
			)
		}
	}
}

func TestCPUFetch(t *testing.T) {

	ram := memory.NewRAM(1024)

	b := bus.New(ram)

	// Put fake instruction at address 0
	ram.Write32(
		0,
		0x12345678,
	)

	cpu := New(b)

	instruction := cpu.Fetch()

	if instruction != 0x12345678 {

		t.Fatalf(
			"expected instruction 0x12345678, got 0x%08X",
			instruction,
		)
	}

	if cpu.Instruction != 0x12345678 {

		t.Fatalf(
			"CPU instruction register not updated",
		)
	}

	if cpu.PC != 4 {

		t.Fatalf(
			"expected PC=4, got %d",
			cpu.PC,
		)
	}
}

func TestCPUFetchMultipleInstructions(t *testing.T) {

	ram := memory.NewRAM(1024)

	b := bus.New(ram)

	ram.Write32(0, 0x11111111)
	ram.Write32(4, 0x22222222)
	ram.Write32(8, 0x33333333)

	cpu := New(b)

	values := []uint32{
		cpu.Fetch(),
		cpu.Fetch(),
		cpu.Fetch(),
	}

	expected := []uint32{
		0x11111111,
		0x22222222,
		0x33333333,
	}

	for i := range expected {

		if values[i] != expected[i] {

			t.Fatalf(
				"fetch %d: expected 0x%08X got 0x%08X",
				i,
				expected[i],
				values[i],
			)
		}
	}

	if cpu.PC != 12 {

		t.Fatalf(
			"expected PC=12, got %d",
			cpu.PC,
		)
	}
}

func TestCPUStop(t *testing.T) {

	cpu := createTestCPU()

	cpu.Running = true

	cpu.Stop()

	if cpu.Running {

		t.Fatalf(
			"CPU should be stopped",
		)
	}
}
