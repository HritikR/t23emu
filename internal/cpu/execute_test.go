package cpu

import (
	"math"
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

func TestExecuteRotates(t *testing.T) {
	cpu := createTestCPU()
	cpu.WriteRegister(8, 0x80000001)
	cpu.WriteRegister(10, 4)

	cpu.Execute(Instruction{
		Opcode: OP_SPECIAL,
		Rs:     1,
		Rt:     8,
		Rd:     9,
		Shamt:  4,
		Funct:  FUNCT_SRL,
	})
	if got := cpu.ReadRegister(9); got != 0x18000000 {
		t.Fatalf("ROTR failed: expected 0x18000000, got 0x%08X", got)
	}

	cpu.Execute(Instruction{
		Opcode: OP_SPECIAL,
		Rs:     10,
		Rt:     8,
		Rd:     9,
		Shamt:  1,
		Funct:  FUNCT_SRLV,
	})
	if got := cpu.ReadRegister(9); got != 0x18000000 {
		t.Fatalf("ROTRV failed: expected 0x18000000, got 0x%08X", got)
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

func TestCP0ConfigAdvertisesUserLocal(t *testing.T) {
	cpu := createTestCPU()

	if got := cpu.CP0[CP0_CONFIG]; got&CONFIG_M == 0 {
		t.Fatalf("expected Config.M to be set, got 0x%08X", got)
	}

	// mfc0 $t0, Config, 1
	cpu.Execute(Instruction{
		Raw:    1,
		Opcode: OP_COP0,
		Rs:     COP0_MFC0,
		Rt:     8,
		Rd:     CP0_CONFIG,
	})

	if got := cpu.ReadRegister(8); got != CP0_CONFIG1_RESET {
		t.Fatalf("expected Config1 0x%08X, got 0x%08X", CP0_CONFIG1_RESET, got)
	}
	if got := cpu.ReadRegister(8); got&CONFIG_M == 0 {
		t.Fatalf("expected Config1.M to be set, got 0x%08X", got)
	}

	cpu.Execute(Instruction{
		Raw:    2,
		Opcode: OP_COP0,
		Rs:     COP0_MFC0,
		Rt:     8,
		Rd:     CP0_CONFIG,
	})

	if got := cpu.ReadRegister(8); got != CP0_CONFIG2_RESET {
		t.Fatalf("expected Config2 0x%08X, got 0x%08X", CP0_CONFIG2_RESET, got)
	}

	cpu.Execute(Instruction{
		Raw:    3,
		Opcode: OP_COP0,
		Rs:     COP0_MFC0,
		Rt:     8,
		Rd:     CP0_CONFIG,
	})

	if got := cpu.ReadRegister(8); got != CP0_CONFIG3_RESET {
		t.Fatalf("expected Config3 0x%08X, got 0x%08X", CP0_CONFIG3_RESET, got)
	}
}

func TestCP0UserLocalAndRDHWR(t *testing.T) {
	cpu := createTestCPU()
	cpu.CP0[CP0_HWRENA] = 1 << 29
	cpu.CP0[CP0_CONTEXT] = 0x80002000
	cpu.WriteRegister(8, 0x76543210)

	cpu.Execute(Instruction{
		Raw:    2,
		Opcode: OP_COP0,
		Rs:     COP0_MTC0,
		Rt:     8,
		Rd:     CP0_CONTEXT,
	})

	if cpu.CP0[CP0_CONTEXT] != 0x80002000 {
		t.Fatalf("expected Context to stay unchanged, got 0x%08X", cpu.CP0[CP0_CONTEXT])
	}
	if cpu.UserLocal != 0x76543210 {
		t.Fatalf("expected UserLocal 0x76543210, got 0x%08X", cpu.UserLocal)
	}

	cpu.Execute(Instruction{
		Opcode: OP_SPECIAL3,
		Rt:     9,
		Rd:     29,
		Funct:  FUNCT3_RDHWR,
	})

	if got := cpu.ReadRegister(9); got != 0x76543210 {
		t.Fatalf("expected RDHWR UserLocal value, got 0x%08X", got)
	}
}

func TestRDHWRUserLocalDoesNotInferFromStack(t *testing.T) {
	cpu := createTestCPU()
	cpu.CurrentPC = 0x00401000
	cpu.CP0[CP0_HWRENA] = 1 << 29
	cpu.UserLocal = 0x775b8000
	cpu.WriteRegister(29, 0x7fffe000)

	cpu.Execute(Instruction{
		Opcode: OP_SPECIAL3,
		Rt:     9,
		Rd:     29,
		Funct:  FUNCT3_RDHWR,
	})

	if got := cpu.ReadRegister(9); got != 0x775b8000 {
		t.Fatalf("expected RDHWR to return UserLocal, got 0x%08X", got)
	}
	if cpu.UserLocal != 0x775b8000 {
		t.Fatalf("expected RDHWR not to mutate UserLocal, got 0x%08X", cpu.UserLocal)
	}
}

func TestUserRDHWRAfterKernelSetsHWRENA(t *testing.T) {
	cpu := createTestCPU()
	cpu.CurrentPC = 0x00401000
	cpu.UserLocal = 0x775b8000
	cpu.CP0[CP0_HWRENA] = 1 << 29 // kernel enabled rdhwr $29

	cpu.Execute(Instruction{
		Opcode: OP_SPECIAL3,
		Rt:     9,
		Rd:     29,
		Funct:  FUNCT3_RDHWR,
	})

	if got := cpu.ReadRegister(9); got != 0x775b8000 {
		t.Fatalf("expected UserLocal from RDHWR, got 0x%08X", got)
	}
}

func TestUserRDHWRWithoutHWRENATrapsRI(t *testing.T) {
	cpu := createTestCPU()
	cpu.CurrentPC = 0x00401000
	cpu.UserLocal = 0 // no cached TLS value

	cpu.Execute(Instruction{
		Opcode: OP_SPECIAL3,
		Rt:     9,
		Rd:     29,
		Funct:  FUNCT3_RDHWR,
	})

	if got := cpu.ReadRegister(9); got != 0 {
		t.Fatalf("expected disabled RDHWR not to write target, got 0x%08X", got)
	}
	if excCode := (cpu.CP0[CP0_CAUSE] & CAUSE_EXCCODE) >> 2; excCode != uint32(EXC_RI) {
		t.Fatalf("expected RI exception, got %d", excCode)
	}
}

func TestCOP1LoadStoreMemoryMoves(t *testing.T) {
	cpu, ram := createCPUWithRAM()
	cpu.WriteRegister(4, 0x80000100)
	cpu.FPR[20] = 0x11223344
	cpu.FPR[21] = 0x55667788

	cpu.Execute(Instruction{
		Opcode:    OP_SDC1,
		Rs:        4,
		Rt:        20,
		Immediate: 0x38,
	})

	if got := ram.Read32(0x138); got != 0x11223344 {
		t.Fatalf("expected low word stored, got 0x%08X", got)
	}
	if got := ram.Read32(0x13c); got != 0x55667788 {
		t.Fatalf("expected high word stored, got 0x%08X", got)
	}

	ram.Write32(0x140, 0xaabbccdd)
	ram.Write32(0x144, 0xeeff0011)
	cpu.Execute(Instruction{
		Opcode:    OP_LDC1,
		Rs:        4,
		Rt:        22,
		Immediate: 0x40,
	})

	if cpu.FPR[22] != 0xaabbccdd || cpu.FPR[23] != 0xeeff0011 {
		t.Fatalf("expected double load into FPRs, got 0x%08X 0x%08X", cpu.FPR[22], cpu.FPR[23])
	}

	cpu.FPR[24] = 0x01020304
	cpu.Execute(Instruction{
		Opcode:    OP_SWC1,
		Rs:        4,
		Rt:        24,
		Immediate: 0x48,
	})
	if got := ram.Read32(0x148); got != 0x01020304 {
		t.Fatalf("expected single word stored, got 0x%08X", got)
	}

	ram.Write32(0x14c, 0xfeedface)
	cpu.Execute(Instruction{
		Opcode:    OP_LWC1,
		Rs:        4,
		Rt:        25,
		Immediate: 0x4c,
	})
	if cpu.FPR[25] != 0xfeedface {
		t.Fatalf("expected single word loaded, got 0x%08X", cpu.FPR[25])
	}
}

func TestCOP1RegisterTransfers(t *testing.T) {
	cpu := createTestCPU()
	cpu.WriteRegister(8, 0x12345678)

	cpu.Execute(Instruction{Opcode: OP_COP1, Rs: COP1_MTC1, Rt: 8, Rd: 4})
	if cpu.FPR[4] != 0x12345678 {
		t.Fatalf("expected MTC1 to write FPR, got 0x%08X", cpu.FPR[4])
	}

	cpu.Execute(Instruction{Opcode: OP_COP1, Rs: COP1_MFC1, Rt: 9, Rd: 4})
	if got := cpu.ReadRegister(9); got != 0x12345678 {
		t.Fatalf("expected MFC1 value, got 0x%08X", got)
	}

	cpu.WriteRegister(10, 0x00020001)
	cpu.Execute(Instruction{Opcode: OP_COP1, Rs: COP1_CTC1, Rt: 10, Rd: 31})
	// Cause.E (bit 17) is read-only zero in hardware; CTC1 keeps the rest.
	if got := cpu.FCSR&0x0000FFFF; got != 0x00000001 {
		t.Fatalf("expected CTC1 to write FCSR (Cause.E masked), got 0x%08X", cpu.FCSR)
	}

	cpu.Execute(Instruction{Opcode: OP_COP1, Rs: COP1_CFC1, Rt: 11, Rd: 31})
	// Cause.E was masked on the write, so it reads back zero too.
	if got := cpu.ReadRegister(11); got != 0x00000001 {
		t.Fatalf("expected CFC1 FCSR value, got 0x%08X", got)
	}
}

func TestCP0CountAdvancesFromCycles(t *testing.T) {
	cpu, ram := createCPUWithRAM()
	ram.Write32(0, 0)
	ram.Write32(4, 0)
	cpu.Running = true

	cpu.writeCP0(CP0_COUNT, 0, 100)
	cpu.Step()
	cpu.Step()

	if got := cpu.readCP0(CP0_COUNT, 0); got != 101 {
		t.Fatalf("expected Count to advance to 101, got %d", got)
	}
}

func TestCP0CompareRaisesTimerInterrupt(t *testing.T) {
	cpu, ram := createCPUWithRAM()
	ram.Write32(0, 0)
	cpu.Running = true
	cpu.CP0[CP0_STATUS] = STATUS_IE | CAUSE_IP7
	cpu.writeCP0(CP0_COUNT, 0, 0)
	cpu.writeCP0(CP0_COMPARE, 0, 1)

	cpu.Step()
	cpu.Step()
	cpu.Step()

	if cpu.PC != 0x80000180 {
		t.Fatalf("expected CP0 timer interrupt vector, got 0x%08X", cpu.PC)
	}
	if got := (cpu.CP0[CP0_CAUSE] >> 2) & 0x1F; got != uint32(EXC_INT) {
		t.Fatalf("expected interrupt exception, got %d", got)
	}

	cpu.writeCP0(CP0_COMPARE, 0, 10)
	if cpu.CP0[CP0_CAUSE]&CAUSE_IP7 != 0 {
		t.Fatalf("writing Compare should clear IP7, Cause=0x%08X", cpu.CP0[CP0_CAUSE])
	}
}

func TestStepTakesExternalInterrupt(t *testing.T) {
	cpu, ram := createCPUWithRAM()
	ram.Write32(0, 0)
	cpu.PC = 0
	cpu.NextPC = 4
	cpu.CP0[CP0_STATUS] = STATUS_IE | CAUSE_IP2
	cpu.InterruptPending = func() uint32 { return CAUSE_IP2 }
	cpu.Running = true

	cpu.Step()

	if cpu.PC != 0x80000180 {
		t.Fatalf("expected interrupt vector 0x80000180, got 0x%08X", cpu.PC)
	}
	if got := uint8((cpu.CP0[CP0_CAUSE] & CAUSE_EXCCODE) >> 2); got != EXC_INT {
		t.Fatalf("expected interrupt exception code, got %d", got)
	}
	if cpu.CP0[CP0_EPC] != 0 {
		t.Fatalf("expected EPC 0, got 0x%08X", cpu.CP0[CP0_EPC])
	}
}

func TestExternalInterruptWaitsForBranchDelaySlot(t *testing.T) {
	cpu, ram := createCPUWithRAM()
	ram.Write32(0, 0x10000001) // beq $zero, $zero, 0x8
	ram.Write32(4, 0)          // delay slot
	cpu.PC = 0
	cpu.NextPC = 4
	cpu.CP0[CP0_STATUS] = STATUS_IE | CAUSE_IP2
	checks := 0
	cpu.InterruptPending = func() uint32 {
		checks++
		if checks >= 2 {
			return CAUSE_IP2
		}
		return 0
	}
	cpu.Running = true

	cpu.Step()
	cpu.Step()

	if cpu.PC != 8 {
		t.Fatalf("expected delay slot to run before interrupt, got PC 0x%08X", cpu.PC)
	}

	cpu.Step()

	if cpu.PC != 0x80000180 {
		t.Fatalf("expected interrupt after delay slot, got PC 0x%08X", cpu.PC)
	}
	if cpu.CP0[CP0_EPC] != 8 {
		t.Fatalf("expected EPC at branch target, got 0x%08X", cpu.CP0[CP0_EPC])
	}
	if cpu.CP0[CP0_CAUSE]&CAUSE_BD != 0 {
		t.Fatalf("expected interrupt after delay slot to leave BD clear, got Cause 0x%08X", cpu.CP0[CP0_CAUSE])
	}
}

func TestWAITStopsFetchUntilEnabledInterrupt(t *testing.T) {
	cpu, ram := createCPUWithRAM()
	ram.Write32(0, 0x42000020) // wait
	ram.Write32(4, EncodeI(OP_ADDI, 0, 8, 1))
	cpu.PC = 0
	cpu.NextPC = 4
	cpu.Running = true

	cpu.Step()

	if !cpu.Waiting {
		t.Fatalf("expected CPU to enter WAIT state")
	}
	if !cpu.Running {
		t.Fatalf("WAIT should not halt the CPU")
	}
	if cpu.PC != 4 {
		t.Fatalf("expected PC after WAIT to be 0x00000004, got 0x%08X", cpu.PC)
	}

	cpu.Step()

	if cpu.ReadRegister(8) != 0 {
		t.Fatalf("WAIT fetched the following instruction without an interrupt")
	}
	if cpu.PC != 4 {
		t.Fatalf("expected PC to remain at 0x00000004 while waiting, got 0x%08X", cpu.PC)
	}

	cpu.InterruptPending = func() uint32 { return CAUSE_IP2 }
	cpu.Step()

	if cpu.Waiting {
		t.Fatalf("pending interrupt should wake WAIT even when IE is clear")
	}
	if cpu.ReadRegister(8) != 0 {
		t.Fatalf("WAIT should wake before fetching the following instruction")
	}
	if cpu.PC != 4 {
		t.Fatalf("expected PC to remain at instruction after WAIT, got 0x%08X", cpu.PC)
	}

	cpu.InterruptPending = func() uint32 { return 0 }
	cpu.Step()

	if cpu.ReadRegister(8) != 1 {
		t.Fatalf("expected execution to resume after WAIT, got r8=0x%08X", cpu.ReadRegister(8))
	}
	if cpu.PC != 8 {
		t.Fatalf("expected PC after resumed instruction to be 0x00000008, got 0x%08X", cpu.PC)
	}
}

func TestWAITTakesEnabledInterrupt(t *testing.T) {
	cpu, ram := createCPUWithRAM()
	ram.Write32(0, 0x42000020) // wait
	cpu.PC = 0
	cpu.NextPC = 4
	cpu.CP0[CP0_STATUS] = STATUS_IE | CAUSE_IP2
	cpu.Running = true

	cpu.Step()

	cpu.InterruptPending = func() uint32 { return CAUSE_IP2 }
	cpu.Step()

	if cpu.Waiting {
		t.Fatalf("expected interrupt to leave WAIT state")
	}
	if cpu.PC != 0x80000180 {
		t.Fatalf("expected interrupt vector 0x80000180, got 0x%08X", cpu.PC)
	}
	if cpu.CP0[CP0_EPC] != 4 {
		t.Fatalf("expected EPC at instruction after WAIT, got 0x%08X", cpu.CP0[CP0_EPC])
	}
}

func TestWAITWakesWithoutDeliverableInterrupt(t *testing.T) {
	cpu, ram := createCPUWithRAM()
	ram.Write32(0, 0x42000020) // wait
	ram.Write32(4, 0x24080001) // addiu t0, zero, 1
	cpu.PC = 0
	cpu.NextPC = 4
	cpu.Running = true

	cpu.Step()
	if !cpu.Waiting {
		t.Fatalf("expected CPU to enter WAIT state")
	}

	cpu.WakePending = func() bool { return true }
	cpu.Step()

	if cpu.Waiting {
		t.Fatalf("expected wake event to leave WAIT state")
	}
	if cpu.CP0[CP0_CAUSE]&CAUSE_IP != 0 {
		t.Fatalf("wake event should not force Cause.IP bits, got 0x%08X", cpu.CP0[CP0_CAUSE])
	}

	cpu.Step()
	if cpu.ReadRegister(8) != 1 {
		t.Fatalf("expected execution to resume after wake, got r8=0x%08X", cpu.ReadRegister(8))
	}
}

func TestPollLoopFastForward(t *testing.T) {
	cpu, ram := createCPUWithRAM()
	// A poll loop: load a zero flag and branch back while it is zero.
	ram.Write32(0, 0x8FA80000) // lw $t0, 0($sp)
	ram.Write32(4, 0x1000FFFE) // beq $t0, $zero, 0
	ram.Write32(8, 0x00000000) // nop (delay slot)
	cpu.Regs[29] = 0x100      // $sp into RAM
	cpu.CP0[CP0_STATUS] = STATUS_IE | STATUS_ERL
	cpu.Running = true
	cpu.NextWakeCycle = func() uint64 { return 1_000_000 }

	for i := 0; i < 200; i++ {
		cpu.Step()
	}

	if cpu.Cycles < 1_000_000 {
		t.Fatalf("expected poll loop to fast-forward to wake cycle, got %d", cpu.Cycles)
	}
	if cpu.PC > 12 {
		t.Fatalf("expected execution to stay inside the loop, PC=0x%08X", cpu.PC)
	}
}

func TestCountingLoopNotSkipped(t *testing.T) {
	cpu, ram := createCPUWithRAM()
	// A progress loop: decrement a counter and branch back while
	// nonzero. Registers change every iteration, so it must never be
	// fast-forwarded.
	ram.Write32(0, 0x2508FFFF) // addiu $t0, $t0, -1
	ram.Write32(4, 0x1500FFFE) // bne $t0, $zero, 0
	ram.Write32(8, 0x00000000) // nop (delay slot)
	cpu.Regs[8] = 5
	cpu.CP0[CP0_STATUS] = STATUS_IE | STATUS_ERL
	cpu.Running = true
	cpu.NextWakeCycle = func() uint64 { return 1_000_000 }

	for i := 0; i < 200; i++ {
		cpu.Step()
	}

	if cpu.Cycles >= 1_000_000 {
		t.Fatalf("counting loop must not be fast-forwarded, cycles=%d", cpu.Cycles)
	}
	if cpu.ReadRegister(8) != 0 {
		t.Fatalf("expected counter to reach zero, got r8=0x%08X", cpu.ReadRegister(8))
	}
}

func TestPollLoopWithInnerBranch(t *testing.T) {
	cpu, ram := createCPUWithRAM()
	// A poll loop whose body contains a nested counting loop: the
	// inner branch's registers change every iteration, but the outer
	// branch always executes with identical state, so the outer loop
	// must still be detected and skipped.
	ram.Write32(0, 0x8FA80000)  // lw   $t0, 0($sp)
	ram.Write32(4, 0x24090003)  // addiu $t1, $zero, 3
	ram.Write32(8, 0x2529FFFF)  // addiu $t1, $t1, -1
	ram.Write32(12, 0x1520FFFE) // bne  $t1, $zero, 8
	ram.Write32(16, 0x00000000) // nop
	ram.Write32(20, 0x1000FFFA) // beq  $t0, $zero, 0
	ram.Write32(24, 0x00000000) // nop
	cpu.Regs[29] = 0x100
	cpu.CP0[CP0_STATUS] = STATUS_IE | STATUS_ERL
	cpu.Running = true
	cpu.NextWakeCycle = func() uint64 { return 1_000_000 }

	for i := 0; i < 600; i++ {
		cpu.Step()
	}

	if cpu.Cycles < 1_000_000 {
		t.Fatalf("expected outer poll loop to fast-forward despite inner branch, got %d cycles", cpu.Cycles)
	}
}

func TestPollLoopInterruptsDisabledSleeps(t *testing.T) {
	cpu, ram := createCPUWithRAM()
	ram.Write32(0, 0x8FA80000) // lw $t0, 0($sp)
	ram.Write32(4, 0x1000FFFE) // beq $t0, $zero, 0
	ram.Write32(8, 0x00000000) // nop
	cpu.Regs[29] = 0x100
	cpu.CP0[CP0_STATUS] = STATUS_ERL // IE=0
	cpu.Running = true
	cpu.NextWakeCycle = func() uint64 { return 100_000_000 }

	for i := 0; i < 400; i++ {
		cpu.Step()
	}

	// The IE=0 path advances Cycles by pollSleepCycles per skip
	// instead of jumping to the (undeliverable) wake event.
	if cpu.Cycles < pollSleepCycles {
		t.Fatalf("expected IE=0 poll loop to advance cycles past sleep, got %d", cpu.Cycles)
	}
	if cpu.Cycles >= 100_000_000 {
		t.Fatalf("IE=0 poll loop must not jump to wake cycle, got %d", cpu.Cycles)
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

func TestExecuteSyscallException(t *testing.T) {
	cpu := createTestCPU()
	cpu.PC = 0x00400004

	// Syscall instruction: funct = 12, Opcode = 0
	inst := Instruction{
		Opcode: OP_SPECIAL,
		Funct:  FUNCT_SYSCALL,
	}

	cpu.Execute(inst)

	// Jump to exception vector
	if cpu.PC != 0x80000180 {
		t.Fatalf("expected PC to be 0x80000180, got 0x%08X", cpu.PC)
	}

	// ExcCode should be 8 (SYS)
	excCode := (cpu.CP0[13] >> 2) & 0x1F
	if excCode != 8 {
		t.Fatalf("expected ExcCode to be 8, got %d", excCode)
	}

	// EPC should be 0x00400000 (c.PC - 4)
	if cpu.CP0[14] != 0x00400000 {
		t.Fatalf("expected EPC to be 0x00400000, got 0x%08X", cpu.CP0[14])
	}

	// EXL bit should be set
	if (cpu.CP0[12] & 0x2) == 0 {
		t.Fatalf("expected Status EXL bit to be set")
	}
}

func TestExecuteBreakpointException(t *testing.T) {
	cpu := createTestCPU()
	cpu.PC = 0x00400004

	// Break instruction: funct = 13, Opcode = 0
	inst := Instruction{
		Opcode: OP_SPECIAL,
		Funct:  FUNCT_BREAK,
	}

	cpu.Execute(inst)

	if cpu.PC != 0x80000180 {
		t.Fatalf("expected PC to be 0x80000180, got 0x%08X", cpu.PC)
	}

	excCode := (cpu.CP0[13] >> 2) & 0x1F
	if excCode != 9 {
		t.Fatalf("expected ExcCode to be 9, got %d", excCode)
	}
}

func TestExecuteRIException(t *testing.T) {
	cpu := createTestCPU()
	cpu.PC = 0x00400004

	// Invalid instruction: Opcode = 99 (invalid)
	inst := Instruction{
		Opcode: 99,
	}

	cpu.Execute(inst)

	if cpu.PC != 0x80000180 {
		t.Fatalf("expected PC to be 0x80000180, got 0x%08X", cpu.PC)
	}

	excCode := (cpu.CP0[13] >> 2) & 0x1F
	if excCode != 10 {
		t.Fatalf("expected ExcCode to be 10, got %d", excCode)
	}
}

func TestExecuteAdEExceptions(t *testing.T) {
	cpu := createTestCPU() // RAM size is 1024 (0x0 to 0x3FF)
	cpu.PC = 0x00400004

	// 1. LW Address Error (AdEL)
	cpu.WriteRegister(1, 0x10000) // Out of bounds base
	instLw := Instruction{
		Opcode:    OP_LW,
		Rs:        1,
		Rt:        2,
		Immediate: 0,
	}

	cpu.Execute(instLw)

	if cpu.PC != 0x80000180 {
		t.Fatalf("expected PC to be 0x80000180, got 0x%08X", cpu.PC)
	}

	excCode := (cpu.CP0[13] >> 2) & 0x1F
	if excCode != 4 {
		t.Fatalf("expected ExcCode to be 4 (ADEL), got %d", excCode)
	}

	if cpu.CP0[8] != 0x10000 {
		t.Fatalf("expected BadVAddr to be 0x10000, got 0x%08X", cpu.CP0[8])
	}

	// Reset Status EXL so next exception can capture EPC
	cpu.CP0[12] &= ^uint32(0x2)
	cpu.PC = 0x00400004

	// 2. SW Address Error (AdES)
	cpu.WriteRegister(1, 0x20000)
	instSw := Instruction{
		Opcode:    OP_SW,
		Rs:        1,
		Rt:        2,
		Immediate: 0,
	}

	cpu.Execute(instSw)

	excCodeSw := (cpu.CP0[13] >> 2) & 0x1F
	if excCodeSw != 5 {
		t.Fatalf("expected ExcCode to be 5 (ADES), got %d", excCodeSw)
	}

	if cpu.CP0[8] != 0x20000 {
		t.Fatalf("expected BadVAddr to be 0x20000, got 0x%08X", cpu.CP0[8])
	}
}

func TestStepAddressErrorException(t *testing.T) {
	cpu := createTestCPU()
	cpu.PC = 0x99999999 // Completely unmapped fetch PC
	cpu.Running = true

	cpu.Step()

	// Step should have intercepted unmapped PC, triggered AdEL exception, and PC is now vector
	if cpu.PC != 0x80000180 {
		t.Fatalf("expected PC to jump to exception vector, got 0x%08X", cpu.PC)
	}

	excCode := (cpu.CP0[13] >> 2) & 0x1F
	if excCode != 4 {
		t.Fatalf("expected ExcCode to be 4 (ADEL), got %d", excCode)
	}

	if cpu.CP0[8] != 0x99999999 {
		t.Fatalf("expected BadVAddr to hold unmapped fetch address, got 0x%08X", cpu.CP0[8])
	}

	// EPC should be set to the invalid fetch address
	if cpu.CP0[14] != 0x99999999 {
		t.Fatalf("expected EPC to be invalid fetch PC, got 0x%08X", cpu.CP0[14])
	}
}

func TestExecuteByteAccess(t *testing.T) {
	cpu, ram := createCPUWithRAM()

	// Write 0x8A to address 10 (RAM is byte-addressable via RAM.Data slice)
	// RAM struct has Data field, but RAM.Write32 writes 4 bytes. We can just write using Write32 or Write8.
	// Wait, memory.RAM does not expose Write8/Read8 publicly? Let's check ram.go.
	// Oh, RAM does expose Write32/Read32 but wait, does RAM support Write8?
	// Let's check TestRAMReadWrite8 in ram_test.go. Yes, ram.Write8 is called there!
	// So we can write directly using ram.Write8(10, 0x8A)!
	ram.Write8(10, 0x8A)

	cpu.WriteRegister(1, 8)

	// LBU $t0, 2($t1) => reads offset 10 (8 + 2)
	instLbu := Instruction{
		Opcode:    OP_LBU,
		Rs:        1,
		Rt:        8,
		Immediate: 2,
	}

	cpu.Execute(instLbu)
	if cpu.ReadRegister(8) != 0x8A {
		t.Fatalf("LBU failed: expected 0x8A, got 0x%08X", cpu.ReadRegister(8))
	}

	// LB $t0, 2($t1) => reads offset 10, sign extends to 0xFFFFFF8A
	instLb := Instruction{
		Opcode:    OP_LB,
		Rs:        1,
		Rt:        8,
		Immediate: 2,
	}

	cpu.Execute(instLb)
	if cpu.ReadRegister(8) != 0xFFFFFF8A {
		t.Fatalf("LB failed: expected 0xFFFFFF8A, got 0x%08X", cpu.ReadRegister(8))
	}

	// SB $t0, 3($t1) => writes bottom byte 0x8A to offset 11 (8 + 3)
	cpu.WriteRegister(8, 0x1234568A)
	instSb := Instruction{
		Opcode:    OP_SB,
		Rs:        1,
		Rt:        8,
		Immediate: 3,
	}

	cpu.Execute(instSb)
	val8 := ram.Read8(11)
	if val8 != 0x8A {
		t.Fatalf("SB failed: expected byte at 11 to be 0x8A, got 0x%02X", val8)
	}
}

func TestUnalignedLoadPairLittleEndian(t *testing.T) {
	cpu, ram := createCPUWithRAM()
	for i, value := range []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88} {
		ram.Write8(uint32(0x100+i), value)
	}

	cpu.WriteRegister(4, 0x80000101)
	cpu.WriteRegister(8, 0xa5a5a5a5)

	cpu.Execute(Instruction{
		Opcode:    OP_LWL,
		Rs:        4,
		Rt:        8,
		Immediate: 3,
	})
	cpu.Execute(Instruction{
		Opcode:    OP_LWR,
		Rs:        4,
		Rt:        8,
		Immediate: 0,
	})

	if got := cpu.ReadRegister(8); got != 0x55443322 {
		t.Fatalf("expected unaligned load pair to produce 0x55443322, got 0x%08X", got)
	}
}

func TestUnalignedStorePairLittleEndian(t *testing.T) {
	cpu, ram := createCPUWithRAM()
	for i := 0; i < 8; i++ {
		ram.Write8(uint32(0x100+i), 0xaa)
	}

	cpu.WriteRegister(4, 0x80000101)
	cpu.WriteRegister(8, 0x55443322)

	cpu.Execute(Instruction{
		Opcode:    OP_SWL,
		Rs:        4,
		Rt:        8,
		Immediate: 3,
	})
	cpu.Execute(Instruction{
		Opcode:    OP_SWR,
		Rs:        4,
		Rt:        8,
		Immediate: 0,
	})

	want := []byte{0xaa, 0x22, 0x33, 0x44, 0x55, 0xaa, 0xaa, 0xaa}
	for i, expected := range want {
		if got := ram.Read8(uint32(0x100 + i)); got != expected {
			t.Fatalf("byte %d: expected 0x%02X, got 0x%02X", i, expected, got)
		}
	}
}

func TestExecuteCACHE(t *testing.T) {
	cpu := createTestCPU()
	inst := Instruction{
		Opcode: OP_CACHE,
	}
	// Verify it executes cleanly without panic or exception
	cpu.Execute(inst)
	if cpu.PC != 0 {
		t.Fatalf("CACHE should be treated as NOP, PC changed: 0x%08X", cpu.PC)
	}
}

// enableCU1 enables FPU access in user mode for COP1 tests.
func enableCU1(cpu *CPU) {
	cpu.CP0[CP0_STATUS] |= STATUS_CU1
}

func TestCOP1UnusableInUserMode(t *testing.T) {
	cpu := createTestCPU()
	// Clear CU1 and force user mode (KSU=2, EXL/ERL clear).
	cpu.CP0[CP0_STATUS] = (cpu.CP0[CP0_STATUS] &^ STATUS_CU1 &^ STATUS_EXL &^ STATUS_ERL) | 0x10
	cpu.CurrentPC = 0x100

	cpu.Execute(Instruction{Opcode: OP_COP1, Rs: COP1_MTC1, Rt: 8, Rd: 4})

	cause := cpu.CP0[CP0_CAUSE]
	if got := (cause & CAUSE_EXCCODE) >> 2; got != uint32(EXC_CPU) {
		t.Fatalf("expected CpU exception, got ExcCode %d", got)
	}
	if got := (cause & CAUSE_CE) >> 28; got != 1 {
		t.Fatalf("expected CE=1 for COP1, got CE=%d", got)
	}
}

func TestCOP1Config1AdvertisesFPU(t *testing.T) {
	cpu := createTestCPU()
	if CP0_CONFIG1_RESET&CONFIG1_FP == 0 {
		t.Fatalf("Config1 reset value should advertise FP")
	}
	// readCP0 should return the reset value for Config1 select 1.
	if got := cpu.readCP0(CP0_CONFIG, 1); got&CONFIG1_FP == 0 {
		t.Fatalf("expected Config1.FP set, got 0x%08x", got)
	}
}

func TestCOP1ArithSingle(t *testing.T) {
	cpu := createTestCPU()
	enableCU1(cpu)

	// $f0 = 2.0, $f2 = 3.0
	cpu.writeFPR_S(0, 2.0)
	cpu.writeFPR_S(2, 3.0)

	tests := []struct {
		name   string
		funct  uint8
		expect float32
	}{
		{"add.s", COP1_ADD, 5.0},
		{"sub.s", COP1_SUB, -1.0},
		{"mul.s", COP1_MUL, 6.0},
		{"div.s", COP1_DIV, 2.0 / 3.0},
	}
	for _, tc := range tests {
		cpu.FPR[4] = 0
		cpu.Execute(Instruction{
			Opcode: OP_COP1, Rs: COP1_FMT_S, Rt: 2, Rd: 0, Shamt: 4, Funct: tc.funct,
		})
		if got := cpu.readFPR_S(4); got != tc.expect {
			t.Fatalf("%s expected %g, got %g", tc.name, tc.expect, got)
		}
	}
}

func TestCOP1ArithDouble(t *testing.T) {
	cpu := createTestCPU()
	enableCU1(cpu)

	// $f0 = 2.5, $f4 = 4.0 (paired lanes)
	cpu.writeFPR_D(0, 2.5)
	cpu.writeFPR_D(4, 4.0)

	cpu.Execute(Instruction{
		Opcode: OP_COP1, Rs: COP1_FMT_D, Rt: 4, Rd: 0, Shamt: 2, Funct: COP1_MUL,
	})
	if got := cpu.readFPR_D(2); got != 10.0 {
		t.Fatalf("mul.d expected 10.0, got %g", got)
	}
}

func TestCOP1MulDMatchesTrace(t *testing.T) {
	// Reproduces the 0x46340082 instruction from the user's RI trace:
	//   mul.d $f2, $f0, $f20  -> $f2 = $f0 * $f20
	cpu := createTestCPU()
	enableCU1(cpu)

	cpu.writeFPR_D(0, 1.5)
	cpu.writeFPR_D(20, 7.0)

	cpu.Execute(Instruction{
		Opcode: OP_COP1, Rs: COP1_FMT_D, Rt: 20, Rd: 0, Shamt: 2, Funct: COP1_MUL,
	})
	if got := cpu.readFPR_D(2); got != 10.5 {
		t.Fatalf("expected 10.5, got %g", got)
	}
}

func TestCOP1CvtDWMatchesTrace(t *testing.T) {
	// Reproduces 0x46800021: cvt.d.w $f0, $f0 (32-bit int -> double).
	cpu := createTestCPU()
	enableCU1(cpu)

	cpu.writeFPR_W(0, 42)
	cpu.Execute(Instruction{
		Opcode: OP_COP1, Rs: COP1_FMT_W, Rd: 0, Shamt: 0, Funct: COP1_CVT_D,
	})
	if got := cpu.readFPR_D(0); got != 42.0 {
		t.Fatalf("cvt.d.w expected 42.0, got %g", got)
	}
}

func TestCOP1MovDMatchesTrace(t *testing.T) {
	// Reproduces 0x46200386: mov.d $f14, $f7.
	cpu := createTestCPU()
	enableCU1(cpu)

	cpu.writeFPR_D(7, 3.14159)
	cpu.Execute(Instruction{
		Opcode: OP_COP1, Rs: COP1_FMT_D, Rd: 7, Shamt: 14, Funct: COP1_MOV,
	})
	if got := cpu.readFPR_D(14); got != 3.14159 {
		t.Fatalf("mov.d expected 3.14159, got %g", got)
	}
	// Source lane must remain intact.
	if got := cpu.readFPR_D(7); got != 3.14159 {
		t.Fatalf("mov.d clobbered source, got %g", got)
	}
}

func TestCOP1CvtDSAndCvtSD(t *testing.T) {
	cpu := createTestCPU()
	enableCU1(cpu)

	cpu.writeFPR_S(2, 1.5)
	cpu.Execute(Instruction{
		Opcode: OP_COP1, Rs: COP1_FMT_S, Rd: 2, Shamt: 4, Funct: COP1_CVT_D,
	})
	if got := cpu.readFPR_D(4); got != 1.5 {
		t.Fatalf("cvt.d.s expected 1.5, got %g", got)
	}

	cpu.Execute(Instruction{
		Opcode: OP_COP1, Rs: COP1_FMT_D, Rd: 4, Shamt: 6, Funct: COP1_CVT_S,
	})
	if got := cpu.readFPR_S(6); got != 1.5 {
		t.Fatalf("cvt.s.d expected 1.5, got %g", got)
	}
}

func TestCOP1CvtWRoundModes(t *testing.T) {
	cpu := createTestCPU()
	enableCU1(cpu)

	cpu.writeFPR_D(0, 2.5)
	for _, tc := range []struct {
		name string
		rm   uint32
		funct uint8
		want int32
	}{
		{"trunc.w.d rz", FP_RZ, COP1_TRUNC_W, 2},
		{"ceil.w.d rp", FP_RP, COP1_CEIL_W, 3},
		{"floor.w.d rm", FP_RM, COP1_FLOOR_W, 2},
		{"round.w.d rn", FP_RN, COP1_ROUND_W, 2}, // IEEE 754 / MIPS FP_RN: ties to even (2.5 -> 2)
		{"cvt.w.d default rn", FP_RN, COP1_CVT_W, 2},

		{"cvt.w.d rp", FP_RP, COP1_CVT_W, 3},
		{"cvt.w.d rm", FP_RM, COP1_CVT_W, 2},
	} {
		cpu.FCSR = (cpu.FCSR &^ FCSR_RMMASK) | tc.rm
		cpu.FPR[2] = 0
		cpu.FPR[3] = 0
		cpu.Execute(Instruction{
			Opcode: OP_COP1, Rs: COP1_FMT_D, Rd: 0, Shamt: 2, Funct: tc.funct,
		})
		if got := cpu.readFPR_W(2); got != tc.want {
			t.Fatalf("%s expected %d, got %d", tc.name, tc.want, got)
		}
	}
}

func TestCOP1AbsNegSqrtD(t *testing.T) {
	cpu := createTestCPU()
	enableCU1(cpu)

	cpu.writeFPR_D(0, -4.0)
	cpu.Execute(Instruction{Opcode: OP_COP1, Rs: COP1_FMT_D, Rd: 0, Shamt: 2, Funct: COP1_ABS})
	if got := cpu.readFPR_D(2); got != 4.0 {
		t.Fatalf("abs.d expected 4.0, got %g", got)
	}

	cpu.writeFPR_D(0, 4.0)
	cpu.Execute(Instruction{Opcode: OP_COP1, Rs: COP1_FMT_D, Rd: 0, Shamt: 4, Funct: COP1_NEG})
	if got := cpu.readFPR_D(4); got != -4.0 {
		t.Fatalf("neg.d expected -4.0, got %g", got)
	}

	cpu.writeFPR_D(0, 16.0)
	cpu.Execute(Instruction{Opcode: OP_COP1, Rs: COP1_FMT_D, Rd: 0, Shamt: 6, Funct: COP1_SQRT})
	if got := cpu.readFPR_D(6); got != 4.0 {
		t.Fatalf("sqrt.d expected 4.0, got %g", got)
	}
}

func TestCOP1CompareAndBranch(t *testing.T) {
	cpu := createTestCPU()
	enableCU1(cpu)
	// Lay out: 0x1000: bc1t $f0, +4; 0x1004: nop (delay slot); 0x1008: target.
	cpu.CurrentPC = 0x1000
	cpu.PC = 0x1000
	cpu.NextPC = 0x1004

	cpu.writeFPR_D(0, 1.0)
	cpu.writeFPR_D(2, 1.0)

	// c.eq.d $f0, $f2  (funct=0x32=COP1_C_EQ, fmt=D, fs=0→Rd, ft=2→Rt, cc=0→Shamt)
	cpu.Execute(Instruction{
		Opcode: OP_COP1, Rs: COP1_FMT_D, Rt: 2, Rd: 0, Shamt: 0, Funct: COP1_C_EQ,
	})
	if !cpu.readFCC(0) {
		t.Fatalf("c.eq.d should set FCC0")
	}

	// bc1t to 0x1008: rt bit 0 = tf=1, no likely. Delay slot is at 0x1004,
	// so immediate=1 targets 0x1004 + 1*4 = 0x1008.
	cpu.CurrentPC = 0x1000
	cpu.PC = 0x1000
	cpu.NextPC = 0x1004
	cpu.branchTaken = false
	cpu.Execute(Instruction{
		Opcode: OP_COP1, Rs: COP1_BC, Rt: 0x01, Immediate: 1,
	})
	if !cpu.branchTaken {
		t.Fatalf("bc1t should mark branch taken when FCC0 is set")
	}
	if cpu.NextPC != 0x1008 {
		t.Fatalf("branch target expected 0x1008, got 0x%08x", cpu.NextPC)
	}
}

func TestCOP1CompareLess(t *testing.T) {
	cpu := createTestCPU()
	enableCU1(cpu)

	cpu.writeFPR_D(0, 1.0)
	cpu.writeFPR_D(2, 2.0)

	// c.lt.d $f0, $f2 (signaling, funct=0x3C=60, ft=2 in Rt, fs=0 in Rd)
	cpu.Execute(Instruction{
		Opcode: OP_COP1, Rs: COP1_FMT_D, Rt: 2, Rd: 0, Shamt: 0, Funct: COP1_C_LT,
	})
	if !cpu.readFCC(0) {
		t.Fatalf("c.lt.d 1.0 < 2.0 should set FCC0")
	}

	// Reverse: c.lt.d $f2, $f0 → 2.0 < 1.0 → false
	cpu.Execute(Instruction{
		Opcode: OP_COP1, Rs: COP1_FMT_D, Rt: 0, Rd: 2, Shamt: 0, Funct: COP1_C_LT,
	})
	if cpu.readFCC(0) {
		t.Fatalf("c.lt.d 2.0 < 1.0 should clear FCC0")
	}
}

func TestCOP1CompareNaN(t *testing.T) {
	cpu := createTestCPU()
	enableCU1(cpu)

	cpu.writeFPR_D(0, math.NaN())
	cpu.writeFPR_D(2, 1.0)

	cpu.FCSR &^= FCSR_CAUSE_V
	// c.un.d $f0, $f2 (funct=0x31=49) should be true without signaling.
	cpu.Execute(Instruction{
		Opcode: OP_COP1, Rs: COP1_FMT_D, Rt: 2, Rd: 0, Shamt: 0, Funct: COP1_C_UN,
	})
	if !cpu.readFCC(0) {
		t.Fatalf("c.un.d with NaN operand should set FCC0 (unordered)")
	}
	if cpu.FCSR&FCSR_CAUSE_V != 0 {
		t.Fatalf("c.un.d should not signal invalid")
	}

	// c.lt.d (signaling) with NaN should signal invalid.
	cpu.FCSR &^= FCSR_CAUSE_V
	cpu.Execute(Instruction{
		Opcode: OP_COP1, Rs: COP1_FMT_D, Rt: 2, Rd: 0, Shamt: 0, Funct: COP1_C_LT,
	})
	if cpu.FCSR&FCSR_CAUSE_V == 0 {
		t.Fatalf("c.lt.d with NaN operand should signal invalid")
	}
}

func TestCOP1CFC1ReturnsFIR(t *testing.T) {
	cpu := createTestCPU()
	enableCU1(cpu)

	cpu.Execute(Instruction{Opcode: OP_COP1, Rs: COP1_CFC1, Rt: 5, Rd: 0})
	if got := cpu.ReadRegister(5); got&0x0001C000 != 0x0001C000 {
		t.Fatalf("CFC1 FIR expected S/D/W support bits, got 0x%08x", got)
	}
}

func TestCOP1CFC1CTC1FCSRRoundTrip(t *testing.T) {
	cpu := createTestCPU()
	enableCU1(cpu)

	cpu.FCSR = 0
	cpu.WriteRegister(8, uint32(FP_RM)|FCSR_FCC0)
	cpu.Execute(Instruction{Opcode: OP_COP1, Rs: COP1_CTC1, Rt: 8, Rd: 31})

	cpu.Execute(Instruction{Opcode: OP_COP1, Rs: COP1_CFC1, Rt: 9, Rd: 31})
	if got := cpu.ReadRegister(9); got&FCSR_RMMASK != FP_RM {
		t.Fatalf("FCSR RM did not round-trip, got 0x%08x", got)
	}
	if got := cpu.ReadRegister(9); got&FCSR_FCC0 == 0 {
		t.Fatalf("FCSR FCC0 did not round-trip, got 0x%08x", got)
	}
}

func TestDelayLoopFastForward(t *testing.T) {
	cpu, ram := createCPUWithRAM()

	// __delay() pattern at address 0:
	//   bne $a0, $zero, -1   (self-branch)
	//   addiu $a0, $a0, -1   (delay slot)
	//   jr  $ra
	//   nop
	ram.Write32(0, 0x1480FFFF) // bne $a0, $zero, -1
	ram.Write32(4, 0x2484FFFF) // addiu $a0, $a0, -1
	ram.Write32(8, 0x03E00008) // jr $ra

	cpu.Regs[4] = 1000 // $a0 = 1000
	cpu.PC = 0
	cpu.NextPC = 4
	cpu.Running = true

	// Step 1: BNE — should be fast-forwarded
	cpu.Step()

	if cpu.Regs[4] != 0 {
		t.Fatalf("expected $a0=0 after fast-forward, got 0x%08X", cpu.Regs[4])
	}
	// Fast-forward adds 1000*2=2000 cycles, then Step adds 1 for BNE
	if cpu.Cycles != 2001 {
		t.Fatalf("expected cycles=2001, got %d", cpu.Cycles)
	}
	// PC should point to delay slot
	if cpu.PC != 4 {
		t.Fatalf("expected PC=4 (delay slot), got 0x%08X", cpu.PC)
	}

	// Step 2: delay slot ADDIU — decrements $a0 from 0 to 0xFFFFFFFF
	cpu.Step()

	if cpu.Regs[4] != 0xFFFFFFFF {
		t.Fatalf("expected $a0=0xFFFFFFFF after delay slot, got 0x%08X", cpu.Regs[4])
	}
	if cpu.Cycles != 2002 {
		t.Fatalf("expected cycles=2002, got %d", cpu.Cycles)
	}
	// PC should now be at JR $ra
	if cpu.PC != 8 {
		t.Fatalf("expected PC=8 (jr $ra), got 0x%08X", cpu.PC)
	}
}

func TestDelayLoopFastForwardZeroCount(t *testing.T) {
	cpu, ram := createCPUWithRAM()

	ram.Write32(0, 0x1480FFFF) // bne $a0, $zero, -1
	ram.Write32(4, 0x2484FFFF) // addiu $a0, $a0, -1

	cpu.Regs[4] = 0 // $a0 = 0: BNE not taken, no fast-forward
	cpu.PC = 0
	cpu.NextPC = 4
	cpu.Running = true

	cpu.Step()

	// BNE not taken, $a0 unchanged
	if cpu.Regs[4] != 0 {
		t.Fatalf("expected $a0=0, got 0x%08X", cpu.Regs[4])
	}
	if cpu.Cycles != 1 {
		t.Fatalf("expected cycles=1, got %d", cpu.Cycles)
	}
}

func TestCOP1XAndExtendedCOP1(t *testing.T) {
	cpu, ram := createCPUWithRAM()
	enableCU1(cpu)

	// Test MFHC1 / MTHC1
	cpu.WriteRegister(1, 0x12345678)
	cpu.Execute(Instruction{
		Opcode: OP_COP1, Rs: COP1_MTHC1, Rt: 1, Rd: 2,
	})
	cpu.Execute(Instruction{
		Opcode: OP_COP1, Rs: COP1_MFHC1, Rt: 3, Rd: 2,
	})
	if cpu.ReadRegister(3) != 0x12345678 {
		t.Fatalf("MFHC1/MTHC1 expected 0x12345678, got 0x%08X", cpu.ReadRegister(3))
	}

	// Test COP1X MADD.S: (2.0 * 3.0) + 4.0 = 10.0
	cpu.writeFPR_S(0, 4.0) // fr
	cpu.writeFPR_S(1, 3.0) // ft
	cpu.writeFPR_S(2, 2.0) // fs
	cpu.Execute(Instruction{
		Opcode: OP_COP1X, Rs: 0, Rt: 1, Rd: 2, Shamt: 3, Funct: COP1X_MADD_S,
	})
	if got := cpu.readFPR_S(3); got != 10.0 {
		t.Fatalf("MADD.S expected 10.0, got %g", got)
	}

	// Test COP1X SWXC1 / LWXC1
	cpu.WriteRegister(4, 0x100) // base
	cpu.WriteRegister(5, 0x10)  // index (addr = 0x110)
	cpu.writeFPR_S(8, 42.5)     // fs
	cpu.Execute(Instruction{
		Opcode: OP_COP1X, Rs: 4, Rt: 5, Rd: 8, Funct: COP1X_SWXC1,
	})
	if bits := ram.Read32(0x110); math.Float32frombits(bits) != 42.5 {
		t.Fatalf("SWXC1 memory expected 42.5, got bits 0x%08X", bits)
	}
	cpu.Execute(Instruction{
		Opcode: OP_COP1X, Rs: 4, Rt: 5, Shamt: 9, Funct: COP1X_LWXC1,
	})
	if got := cpu.readFPR_S(9); got != 42.5 {
		t.Fatalf("LWXC1 expected 42.5, got %g", got)
	}

	// Test FUNCT_MOVCI (MOVF / MOVT on GPR): raw 0x002d1001 is MOVT $2, $1, cc=3 (rs=1, rt=13, rd=2)
	cpu.setFCC(3, true)
	cpu.WriteRegister(2, 99)
	cpu.WriteRegister(1, 77)
	cpu.Execute(Decode(0x002d1001)) // MOVT $2, $1, cc=3
	if cpu.ReadRegister(2) != 77 {
		t.Fatalf("MOVT GPR expected 77, got %d", cpu.ReadRegister(2))
	}

}


