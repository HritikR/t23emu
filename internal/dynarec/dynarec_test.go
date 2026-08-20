package dynarec

import (
	"testing"
)

type mockCPUContext struct {
	pc          uint32
	nextPC      uint32
	currentPC   uint32
	inst        uint32
	regs        [32]uint32
	memory      map[uint32]uint32
	cycles      uint64
	running     bool
	inDelaySlot bool
	branchTaken bool
}

func newMockCPUContext() *mockCPUContext {
	return &mockCPUContext{
		memory:  make(map[uint32]uint32),
		running: true,
	}
}

func (m *mockCPUContext) GetPC() uint32                    { return m.pc }
func (m *mockCPUContext) SetPC(pc uint32)                  { m.pc = pc }
func (m *mockCPUContext) GetNextPC() uint32                { return m.nextPC }
func (m *mockCPUContext) SetNextPC(pc uint32)              { m.nextPC = pc }
func (m *mockCPUContext) SetCurrentPC(pc uint32)           { m.currentPC = pc }
func (m *mockCPUContext) SetInstruction(raw uint32)        { m.inst = raw }
func (m *mockCPUContext) IsSingleStep() bool               { return false }
func (m *mockCPUContext) IsInDelaySlot() bool              { return m.inDelaySlot }
func (m *mockCPUContext) IsBranchTaken() bool              { return m.branchTaken }
func (m *mockCPUContext) SetBranchTaken(taken bool)        { m.branchTaken = taken }
func (m *mockCPUContext) SetInDelaySlot(inDelay bool)      { m.inDelaySlot = inDelay }
func (m *mockCPUContext) CheckInterrupts() bool            { return false }
func (m *mockCPUContext) IncCycles(count uint64)           { m.cycles += count }
func (m *mockCPUContext) HasMapping(addr uint32) bool      { return true }
func (m *mockCPUContext) Read32(addr uint32) uint32        { return m.memory[addr] }
func (m *mockCPUContext) Read8(addr uint32) byte           { return byte(m.memory[addr&^3] >> ((addr & 3) * 8)) }
func (m *mockCPUContext) Write32(addr uint32, val uint32)  { m.memory[addr] = val }
func (m *mockCPUContext) Write8(addr uint32, val byte)     {}
func (m *mockCPUContext) ReadReg(reg uint8) uint32         { return m.regs[reg&31] }
func (m *mockCPUContext) WriteReg(reg uint8, val uint32)   { if reg != 0 { m.regs[reg&31] = val } }
func (m *mockCPUContext) ExecuteRaw(raw uint32) bool       { return true }
func (m *mockCPUContext) RaiseException(exc, addr uint32)  { m.running = false }
func (m *mockCPUContext) IsRunning() bool                  { return m.running }
func (m *mockCPUContext) Retire()                          {}

func TestDynarecLookupAndInvalidate(t *testing.T) {
	engine := NewEngine()
	cpu := newMockCPUContext()

	// addiu $1, $0, 42  (0x2401002A)
	// jr $31            (0x03E00008)
	// nop               (0x00000000)
	pc := uint32(0x80001000)
	cpu.memory[pc] = 0x2401002A
	cpu.memory[pc+4] = 0x03E00008
	cpu.memory[pc+8] = 0x00000000
	cpu.pc = pc
	cpu.nextPC = pc + 4

	block1 := engine.Lookup(cpu, pc)
	if block1 == nil {
		t.Fatalf("expected non-nil basic block")
	}
	if block1.StartPC != pc {
		t.Fatalf("expected StartPC 0x%08X, got 0x%08X", pc, block1.StartPC)
	}

	// Second lookup should return same cached block via lock-free fast path
	block2 := engine.Lookup(cpu, pc)
	if block2 != block1 {
		t.Fatalf("expected fast path to return identical cached block pointer")
	}

	// Invalidate clears cache
	engine.Invalidate()
	block3 := engine.Lookup(cpu, pc)
	if block3 == nil {
		t.Fatalf("expected recompiled basic block after invalidate")
	}
}

func TestDynarecStepBlock(t *testing.T) {
	engine := NewEngine()
	cpu := newMockCPUContext()

	// addiu $1, $0, 10 (0x2401000A)
	// addiu $2, $1, 20 (0x24220014)
	pc := uint32(0x80002000)
	cpu.memory[pc] = 0x2401000A
	cpu.memory[pc+4] = 0x24220014
	cpu.pc = pc
	cpu.nextPC = pc + 4

	ok := engine.StepBlock(cpu)
	if !ok {
		t.Fatalf("expected StepBlock to succeed")
	}

	if cpu.regs[1] != 10 {
		t.Fatalf("expected reg[1] == 10, got %d", cpu.regs[1])
	}
	if cpu.regs[2] != 30 {
		t.Fatalf("expected reg[2] == 30, got %d", cpu.regs[2])
	}
}
