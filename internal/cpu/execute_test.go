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

func TestCP0ConfigAdvertisesConfig1(t *testing.T) {
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
