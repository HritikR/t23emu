package cpu

import (
	"sync"
)

// BasicBlock represents a compiled sequence of instructions.
type BasicBlock struct {
	StartPC   uint32
	InstCount int
	// Exec executes the basic block on the target CPU.
	// Returns true if the block completed fully, false if interrupted or faulted early.
	Exec func(c *CPU) bool
}

// DynarecEngine manages basic block compilation and caching.
type DynarecEngine struct {
	mu           sync.RWMutex
	cache        map[uint32]*BasicBlock
	CompileBlock func(c *CPU, pc uint32) *BasicBlock
	Enabled      bool

	// Stats
	BlockHits   uint64
	BlockMisses uint64
	ExecCount   uint64
}

// NewDynarecEngine creates a new Dynarec execution engine.
func NewDynarecEngine() *DynarecEngine {
	engine := &DynarecEngine{
		cache:   make(map[uint32]*BasicBlock),
		Enabled: true,
	}
	engine.CompileBlock = engine.compileThreadedBlock
	return engine
}

// Invalidate clears all cached basic blocks (e.g. on TLB flush or code modification).
func (d *DynarecEngine) Invalidate() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.cache = make(map[uint32]*BasicBlock)
}

// Lookup finds a cached basic block or compiles a new one.
func (d *DynarecEngine) Lookup(c *CPU, pc uint32) *BasicBlock {
	d.mu.RLock()
	block, ok := d.cache[pc]
	d.mu.RUnlock()

	if ok {
		d.BlockHits++
		return block
	}

	d.BlockMisses++
	block = d.CompileBlock(c, pc)

	d.mu.Lock()
	d.cache[pc] = block
	d.mu.Unlock()

	return block
}

// StepBlock attempts to execute one basic block via Dynarec.
// Returns true if a block was executed, false if fallen back to interpreter.
func (d *DynarecEngine) StepBlock(c *CPU) bool {
	if !d.Enabled || !c.Running || c.Waiting {
		return false
	}

	// Do not use Dynarec in single-step, delay slot, or when breakpoints/trace are enabled
	if c.SingleStep || c.InDelaySlot || c.Trace || c.RecordHistory || (c.Breakpoints != nil && len(c.Breakpoints) > 0) {
		return false
	}

	pc := c.PC

	// Quick address mapping check
	if !c.Bus.HasMapping(pc) {
		return false
	}

	block := d.Lookup(c, pc)
	if block == nil || block.Exec == nil {
		return false
	}

	d.ExecCount++
	return block.Exec(c)
}

// isTerminatingInst reports whether an instruction ends a basic block.
func isTerminatingInst(inst Instruction) bool {
	switch inst.Opcode {
	case OP_J, OP_JAL, OP_BEQ, OP_BNE, OP_BLEZ, OP_BGTZ, OP_BEQL, OP_BNEL, OP_BLEZL, OP_BGTZL:
		return true

	case OP_REGIMM:
		switch inst.Rt {
		case REGIMM_BLTZ, REGIMM_BGEZ, REGIMM_BLTZL, REGIMM_BGEZL, REGIMM_BLTZAL, REGIMM_BGEZAL:
			return true
		}

	case OP_SPECIAL:
		switch inst.Funct {
		case FUNCT_JR, FUNCT_JALR, FUNCT_SYSCALL, FUNCT_BREAK:
			return true
		}

	case OP_COP0:
		// ERET, WAIT, or CP0 control instructions end basic blocks
		return true
	}

	return false
}
