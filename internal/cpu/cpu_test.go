package cpu

import (
	"testing"

	"github.com/HritikR/t23emu/internal/bus"
	"github.com/HritikR/t23emu/internal/memory"
)

func createTestCPU() *CPU {

	ram := memory.NewRAM(1024)

	b := bus.New()

	b.Map(
		0x00000000,
		0x000003FF,
		ram,
	)

	cpu := New(b)

	return cpu
}

func TestCPUReset(t *testing.T) {

	cpu := createTestCPU()

	cpu.Regs[1] = 123
	cpu.PC = 0x1000
	cpu.Instruction = 0xDEADBEEF
	cpu.Running = true

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

	b := bus.New()

	b.Map(
		0,
		0x3FF,
		ram,
	)

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

	b := bus.New()

	b.Map(
		0,
		0x3FF,
		ram,
	)

	ram.Write32(
		0,
		0x11111111,
	)

	ram.Write32(
		4,
		0x22222222,
	)

	ram.Write32(
		8,
		0x33333333,
	)

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

func TestTLBTranslatesKseg2Address(t *testing.T) {
	ram := memory.NewRAM(0x2000)
	b := bus.New()
	b.Map(0, 0x1FFF, ram)

	cpu := New(b)
	cpu.CP0[CP0_INDEX] = 0
	cpu.CP0[CP0_ENTRYHI] = 0xC0000000
	cpu.CP0[CP0_ENTRYLO0] = entryLoV | entryLoD | entryLoG
	cpu.CP0[CP0_ENTRYLO1] = (1 << 6) | entryLoV | entryLoD | entryLoG
	cpu.writeIndexedTLB(0)

	b.Write32(0xC0000000, 0x12345678)
	if got := ram.Read32(0); got != 0x12345678 {
		t.Fatalf("expected TLB store to physical 0, got 0x%08X", got)
	}

	ram.Write32(0x1000, 0x89ABCDEF)
	if got := b.Read32(0xC0001000); got != 0x89ABCDEF {
		t.Fatalf("expected odd TLB page read from physical 0x1000, got 0x%08X", got)
	}
}

func TestCOP0TLBProbeAndRead(t *testing.T) {
	cpu := createTestCPU()

	cpu.CP0[CP0_INDEX] = 3
	cpu.CP0[CP0_PAGEMASK] = 0
	cpu.CP0[CP0_ENTRYHI] = 0xC0000000
	cpu.CP0[CP0_ENTRYLO0] = entryLoV | entryLoD | entryLoG
	cpu.CP0[CP0_ENTRYLO1] = (1 << 6) | entryLoV | entryLoD | entryLoG
	cpu.Execute(Instruction{Opcode: OP_COP0, Rs: COP0_CO, Funct: COP0CO_TLBWI})

	cpu.CP0[CP0_ENTRYHI] = 0xC0000123
	cpu.Execute(Instruction{Opcode: OP_COP0, Rs: COP0_CO, Funct: COP0CO_TLBP})
	if cpu.CP0[CP0_INDEX] != 3 {
		t.Fatalf("expected TLBP to find index 3, got 0x%08X", cpu.CP0[CP0_INDEX])
	}

	cpu.CP0[CP0_ENTRYHI] = 0
	cpu.CP0[CP0_ENTRYLO0] = 0
	cpu.CP0[CP0_ENTRYLO1] = 0
	cpu.Execute(Instruction{Opcode: OP_COP0, Rs: COP0_CO, Funct: COP0CO_TLBR})
	if cpu.CP0[CP0_ENTRYHI]&entryHiVPN != 0xC0000000 {
		t.Fatalf("expected TLBR to restore EntryHi, got 0x%08X", cpu.CP0[CP0_ENTRYHI])
	}
	if cpu.CP0[CP0_ENTRYLO0]&entryLoV == 0 {
		t.Fatalf("expected TLBR to restore valid EntryLo0, got 0x%08X", cpu.CP0[CP0_ENTRYLO0])
	}
}

func TestKseg2StoreMissRaisesTLBSRefill(t *testing.T) {
	cpu := createTestCPU()
	cpu.CP0[CP0_STATUS] = 0
	cpu.PC = 0x80020000
	cpu.CurrentPC = cpu.PC
	cpu.WriteRegister(1, 0xC0000000)
	cpu.WriteRegister(2, 0x12345678)

	cpu.Execute(Instruction{Opcode: OP_SW, Rs: 1, Rt: 2})

	if got := (cpu.CP0[CP0_CAUSE] >> 2) & 0x1F; got != uint32(EXC_TLBS) {
		t.Fatalf("expected TLBS, got %d", got)
	}
	if cpu.CP0[CP0_BADVADDR] != 0xC0000000 {
		t.Fatalf("expected BadVAddr c0000000, got 0x%08X", cpu.CP0[CP0_BADVADDR])
	}
	if cpu.PC != 0x80000000 {
		t.Fatalf("expected TLB refill vector 0x80000000, got 0x%08X", cpu.PC)
	}
}

func TestKseg2InvalidStoreRaisesTLBSGeneralException(t *testing.T) {
	cpu := createTestCPU()
	cpu.CP0[CP0_STATUS] = 0
	cpu.CP0[CP0_INDEX] = 0
	cpu.CP0[CP0_ENTRYHI] = 0xC0000000
	cpu.CP0[CP0_ENTRYLO0] = entryLoG
	cpu.CP0[CP0_ENTRYLO1] = entryLoG
	cpu.writeIndexedTLB(0)

	cpu.PC = 0x80020000
	cpu.CurrentPC = cpu.PC
	cpu.WriteRegister(1, 0xC0000000)
	cpu.WriteRegister(2, 0x12345678)

	cpu.Execute(Instruction{Opcode: OP_SW, Rs: 1, Rt: 2})

	if got := (cpu.CP0[CP0_CAUSE] >> 2) & 0x1F; got != uint32(EXC_TLBS) {
		t.Fatalf("expected TLBS, got %d", got)
	}
	if cpu.PC != 0x80000180 {
		t.Fatalf("expected general exception vector 0x80000180, got 0x%08X", cpu.PC)
	}
}
