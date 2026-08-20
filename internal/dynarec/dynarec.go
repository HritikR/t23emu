package dynarec

import (
	"sync"
)

// CPUContext defines the interface required by the Dynarec engine to execute CPU basic blocks.
type CPUContext interface {
	GetPC() uint32
	SetPC(pc uint32)
	GetNextPC() uint32
	SetNextPC(pc uint32)
	SetCurrentPC(pc uint32)
	SetInstruction(raw uint32)
	IsSingleStep() bool
	IsInDelaySlot() bool
	IsBranchTaken() bool
	SetBranchTaken(taken bool)
	SetInDelaySlot(inDelay bool)
	CheckInterrupts() bool
	IncCycles(count uint64)
	HasMapping(addr uint32) bool
	Read32(addr uint32) uint32
	Read8(addr uint32) byte
	Write32(addr uint32, val uint32)
	Write8(addr uint32, val byte)
	ReadReg(reg uint8) uint32
	WriteReg(reg uint8, val uint32)
	ExecuteRaw(raw uint32) bool
	RaiseException(exc uint32, addr uint32)
	IsRunning() bool
	Retire()
}

// BasicBlock represents a compiled sequence of instructions.
type BasicBlock struct {
	StartPC   uint32
	InstCount int
	Exec      func(c CPUContext) bool
}

// Engine manages basic block compilation and caching.
type Engine struct {
	mu     sync.RWMutex
	cache  map[uint32]*BasicBlock
	lookup [65536]*BasicBlock
}

// NewEngine creates a new Dynarec execution engine.
func NewEngine() *Engine {
	return &Engine{
		cache: make(map[uint32]*BasicBlock),
	}
}

// Invalidate clears all cached basic blocks (e.g. on TLB/MMU changes or self-modifying code).
func (e *Engine) Invalidate() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.cache = make(map[uint32]*BasicBlock)
	for i := range e.lookup {
		e.lookup[i] = nil
	}
}

// Lookup finds or compiles a basic block for the given PC.
func (e *Engine) Lookup(c CPUContext, pc uint32) *BasicBlock {
	index := (pc >> 2) & 0xFFFF

	// Fast path: direct lock-free check in the lookup table
	block := e.lookup[index]
	if block != nil && block.StartPC == pc {
		return block
	}

	// Slow path: acquire lock to search secondary cache or compile new block
	e.mu.Lock()
	defer e.mu.Unlock()

	// Double-check lookup table under lock
	block = e.lookup[index]
	if block != nil && block.StartPC == pc {
		return block
	}

	block = e.cache[pc]
	if block != nil {
		e.lookup[index] = block
		return block
	}

	block = e.compileThreadedBlock(c, pc)
	e.cache[pc] = block
	e.lookup[index] = block

	return block
}

// StepBlock attempts to execute one basic block via Dynarec.
func (e *Engine) StepBlock(c CPUContext) bool {
	pc := c.GetPC()
	if c.IsSingleStep() || c.IsInDelaySlot() {
		return false
	}

	block := e.Lookup(c, pc)
	if block == nil || block.InstCount == 0 {
		return false
	}

	return block.Exec(c)
}
