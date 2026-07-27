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

func TestExecuteJ(t *testing.T) {
	cpu := createTestCPU()
	cpu.PC = 0x10000004

	inst := Instruction{
		Opcode: OP_J,
		Target: 0x00A0000, // target << 2 = 0x00280000
	}

	cpu.Execute(inst)

	if cpu.PC != 0x10280000 {
		t.Fatalf("expected PC 0x10280000, got 0x%08X", cpu.PC)
	}
}

func TestExecuteJAL(t *testing.T) {
	cpu := createTestCPU()
	cpu.PC = 0x10000008

	inst := Instruction{
		Opcode: OP_JAL,
		Target: 0x00A0000, // target << 2 = 0x00280000
	}

	cpu.Execute(inst)

	if cpu.PC != 0x10280000 {
		t.Fatalf("expected PC 0x10280000, got 0x%08X", cpu.PC)
	}

	if cpu.ReadRegister(31) != 0x10000008 {
		t.Fatalf("expected $ra (R31) to be 0x10000008, got 0x%08X", cpu.ReadRegister(31))
	}
}

func TestExecuteJR(t *testing.T) {
	cpu := createTestCPU()
	cpu.WriteRegister(31, 0x00400000)

	inst := Instruction{
		Opcode: OP_SPECIAL,
		Rs:     31,
		Funct:  FUNCT_JR,
	}

	cpu.Execute(inst)

	if cpu.PC != 0x00400000 {
		t.Fatalf("expected PC 0x00400000, got 0x%08X", cpu.PC)
	}
}

func TestExecuteBEQ(t *testing.T) {
	// Case 1: Equal (branch taken)
	cpu := createTestCPU()
	cpu.PC = 0x1000
	cpu.WriteRegister(8, 5)
	cpu.WriteRegister(9, 5)

	inst := Instruction{
		Opcode:    OP_BEQ,
		Rs:        8,
		Rt:        9,
		Immediate: 4, // 4 instructions forward = 16 bytes
	}

	cpu.Execute(inst)

	if cpu.PC != 0x1010 {
		t.Fatalf("BEQ taken: expected PC 0x1010, got 0x%08X", cpu.PC)
	}

	// Case 2: Not Equal (branch not taken)
	cpu = createTestCPU()
	cpu.PC = 0x1000
	cpu.WriteRegister(8, 5)
	cpu.WriteRegister(9, 10)

	cpu.Execute(inst)

	if cpu.PC != 0x1000 {
		t.Fatalf("BEQ not taken: expected PC 0x1000, got 0x%08X", cpu.PC)
	}
}

func TestExecuteBNE(t *testing.T) {
	// Case 1: Not Equal (branch taken)
	cpu := createTestCPU()
	cpu.PC = 0x1000
	cpu.WriteRegister(8, 5)
	cpu.WriteRegister(9, 10)

	inst := Instruction{
		Opcode:    OP_BNE,
		Rs:        8,
		Rt:        9,
		Immediate: 4,
	}

	cpu.Execute(inst)

	if cpu.PC != 0x1010 {
		t.Fatalf("BNE taken: expected PC 0x1010, got 0x%08X", cpu.PC)
	}

	// Case 2: Equal (branch not taken)
	cpu = createTestCPU()
	cpu.PC = 0x1000
	cpu.WriteRegister(8, 5)
	cpu.WriteRegister(9, 5)

	cpu.Execute(inst)

	if cpu.PC != 0x1000 {
		t.Fatalf("BNE not taken: expected PC 0x1000, got 0x%08X", cpu.PC)
	}
}

func TestExecuteLUI(t *testing.T) {
	cpu := createTestCPU()

	inst := Instruction{
		Opcode:    OP_LUI,
		Rt:        8,
		Immediate: 0x1234,
	}

	cpu.Execute(inst)

	if cpu.ReadRegister(8) != 0x12340000 {
		t.Fatalf("expected 0x12340000, got 0x%08X", cpu.ReadRegister(8))
	}
}

func TestExecuteLogicalOps(t *testing.T) {
	cpu := createTestCPU()
	cpu.WriteRegister(8, 0x0F0F0F0F)
	cpu.WriteRegister(9, 0xF0F0F0F0)

	// AND
	inst := Instruction{
		Opcode: OP_SPECIAL,
		Rs:     8,
		Rt:     9,
		Rd:     10,
		Funct:  FUNCT_AND,
	}
	cpu.Execute(inst)
	if cpu.ReadRegister(10) != 0 {
		t.Fatalf("AND failed: got 0x%08X", cpu.ReadRegister(10))
	}

	// OR
	inst.Funct = FUNCT_OR
	cpu.Execute(inst)
	if cpu.ReadRegister(10) != 0xFFFFFFFF {
		t.Fatalf("OR failed: got 0x%08X", cpu.ReadRegister(10))
	}

	// XOR
	inst.Funct = FUNCT_XOR
	cpu.Execute(inst)
	if cpu.ReadRegister(10) != 0xFFFFFFFF {
		t.Fatalf("XOR failed: got 0x%08X", cpu.ReadRegister(10))
	}

	// NOR
	inst.Funct = FUNCT_NOR
	cpu.Execute(inst)
	if cpu.ReadRegister(10) != 0 {
		t.Fatalf("NOR failed: got 0x%08X", cpu.ReadRegister(10))
	}
}

func TestExecuteANDI_ORI(t *testing.T) {
	cpu := createTestCPU()
	cpu.WriteRegister(8, 0x00FF00FF)

	// ANDI
	inst := Instruction{
		Opcode:    OP_ANDI,
		Rs:        8,
		Rt:        9,
		Immediate: 0x000F,
	}
	cpu.Execute(inst)
	if cpu.ReadRegister(9) != 0x0000000F {
		t.Fatalf("ANDI failed: got 0x%08X", cpu.ReadRegister(9))
	}

	// ORI
	inst = Instruction{
		Opcode:    OP_ORI,
		Rs:        8,
		Rt:        10,
		Immediate: 0xF000,
	}
	cpu.Execute(inst)
	if cpu.ReadRegister(10) != 0x00FFF0FF {
		t.Fatalf("ORI failed: got 0x%08X", cpu.ReadRegister(10))
	}
}

func TestExecuteShifts(t *testing.T) {
	cpu := createTestCPU()
	cpu.WriteRegister(8, 0x000000F0)

	// SLL
	inst := Instruction{
		Opcode: OP_SPECIAL,
		Rt:     8,
		Rd:     9,
		Shamt:  4,
		Funct:  FUNCT_SLL,
	}
	cpu.Execute(inst)
	if cpu.ReadRegister(9) != 0x00000F00 {
		t.Fatalf("SLL failed: got 0x%08X", cpu.ReadRegister(9))
	}

	// SRL
	inst.Funct = FUNCT_SRL
	cpu.Execute(inst)
	if cpu.ReadRegister(9) != 0x0000000F {
		t.Fatalf("SRL failed: got 0x%08X", cpu.ReadRegister(9))
	}

	// SRA (arithmetic shift right, sign-preserving)
	cpu.WriteRegister(8, 0xF0000000)
	inst.Rt = 8
	inst.Rd = 9
	inst.Shamt = 4
	inst.Funct = FUNCT_SRA
	cpu.Execute(inst)
	if cpu.ReadRegister(9) != 0xFF000000 {
		t.Fatalf("SRA failed: got 0x%08X", cpu.ReadRegister(9))
	}
}

func TestExecuteSLT(t *testing.T) {
	cpu := createTestCPU()
	cpu.WriteRegister(8, 0xFFFFFFFF) // -1 signed
	cpu.WriteRegister(9, 1)

	inst := Instruction{
		Opcode: OP_SPECIAL,
		Rs:     8,
		Rt:     9,
		Rd:     10,
		Funct:  FUNCT_SLT,
	}
	cpu.Execute(inst)
	if cpu.ReadRegister(10) != 1 {
		t.Fatalf("SLT: expected -1 < 1 to write 1, got %d", cpu.ReadRegister(10))
	}

	cpu.WriteRegister(8, 5)
	cpu.WriteRegister(9, 2)
	cpu.Execute(inst)
	if cpu.ReadRegister(10) != 0 {
		t.Fatalf("SLT: expected 5 < 2 to write 0, got %d", cpu.ReadRegister(10))
	}
}

func TestExecuteCOP0_MTC0_MFC0(t *testing.T) {
	cpu := createTestCPU()
	cpu.WriteRegister(8, 0x12345678)

	// MTC0 $t0, Status ($8 is general reg, $12 is CP0 Status)
	// Rs = 4 (MTC0), Rt = 8, Rd = 12 (CP0 Status)
	instMtc0 := Instruction{
		Opcode: OP_COP0,
		Rs:     4,
		Rt:     8,
		Rd:     12,
	}

	cpu.Execute(instMtc0)

	if cpu.CP0[12] != 0x12345678 {
		t.Fatalf("expected CP0 Status to be 0x12345678, got 0x%08X", cpu.CP0[12])
	}

	// MFC0 $t1, Status
	// Rs = 0 (MFC0), Rt = 9, Rd = 12
	instMfc0 := Instruction{
		Opcode: OP_COP0,
		Rs:     0,
		Rt:     9,
		Rd:     12,
	}

	cpu.Execute(instMfc0)

	if cpu.ReadRegister(9) != 0x12345678 {
		t.Fatalf("expected register 9 to be 0x12345678, got 0x%08X", cpu.ReadRegister(9))
	}
}

func TestExecuteERET(t *testing.T) {
	cpu := createTestCPU()
	cpu.CP0[14] = 0x80001000 // EPC
	cpu.CP0[12] = 0x00000007 // Status (EXL is bit 1)

	// ERET instruction: Opcode=16, Rs=16, Funct=24
	inst := Instruction{
		Opcode: OP_COP0,
		Rs:     16,
		Funct:  24,
	}

	cpu.Execute(inst)

	if cpu.PC != 0x80001000 {
		t.Fatalf("expected PC to be 0x80001000, got 0x%08X", cpu.PC)
	}

	// EXL bit (bit 1, mask 0x2) in CP0 Status should be cleared
	if (cpu.CP0[12] & 0x2) != 0 {
		t.Fatalf("expected Status EXL bit to be cleared, got 0x%08X", cpu.CP0[12])
	}
}
