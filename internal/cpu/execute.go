package cpu

// Execute runs a decoded instruction.
func (c *CPU) Execute(inst Instruction) {

	switch inst.Opcode {

	case OP_SPECIAL:
		c.executeRType(inst)

	case OP_REGIMM:
		c.executeRegimm(inst)

	case OP_SPECIAL2:
		c.executeSpecial2(inst)

	case OP_SPECIAL3:
		c.executeSpecial3(inst)

	case OP_J:
		c.executeJ(inst)

	case OP_JAL:
		c.executeJAL(inst)

	case OP_BEQ:
		c.executeBEQ(inst)

	case OP_BNE:
		c.executeBNE(inst)

	case OP_BLEZ:
		c.executeBLEZ(inst)

	case OP_BGTZ:
		c.executeBGTZ(inst)

	case OP_BEQL:
		c.executeBEQL(inst)

	case OP_BNEL:
		c.executeBNEL(inst)

	case OP_BLEZL:
		c.executeBLEZL(inst)

	case OP_BGTZL:
		c.executeBGTZL(inst)

	case OP_ADDI:
		c.executeADDI(inst)

	case OP_ADDIU:
		c.executeADDIU(inst)

	case OP_SLTI:
		c.executeSLTI(inst)

	case OP_SLTIU:
		c.executeSLTIU(inst)

	case OP_ANDI:
		c.executeANDI(inst)

	case OP_ORI:
		c.executeORI(inst)

	case OP_XORI:
		c.executeXORI(inst)

	case OP_LUI:
		c.executeLUI(inst)

	case OP_LB:
		c.executeLB(inst)

	case OP_LBU:
		c.executeLBU(inst)

	case OP_LH:
		c.executeLH(inst)

	case OP_LHU:
		c.executeLHU(inst)

	case OP_LW:
		c.executeLW(inst)

	case OP_LWL:
		c.executeLWL(inst)

	case OP_LWR:
		c.executeLWR(inst)

	case OP_SB:
		c.executeSB(inst)

	case OP_SH:
		c.executeSH(inst)

	case OP_SW:
		c.executeSW(inst)

	case OP_SWL:
		c.executeSWL(inst)

	case OP_SWR:
		c.executeSWR(inst)

	case OP_LL:
		c.executeLL(inst)

	case OP_LWC1:
		c.executeLWC1(inst)

	case OP_LDC1:
		c.executeLDC1(inst)

	case OP_SC:
		c.executeSC(inst)

	case OP_SWC1:
		c.executeSWC1(inst)

	case OP_SDC1:
		c.executeSDC1(inst)

	case OP_CACHE:
		c.executeCACHE(inst)

	case OP_PREF:
		// Prefetch is a hint with no architectural effect.
		c.retire()

	case OP_COP0:
		c.executeCOP0(inst)

	case OP_COP1:
		c.executeCOP1(inst)

	case OP_COP2:
		c.coprocessorUnusable(inst.Opcode - OP_COP0)

	default:
		c.Exception(EXC_RI, 0)
	}
}

// executeRType handles opcode 0 instructions.
//
// R-type operations are selected by the funct field.
func (c *CPU) executeRType(inst Instruction) {

	switch inst.Funct {

	case FUNCT_SLL:
		c.executeSLL(inst)

	case FUNCT_SRL:
		c.executeSRL(inst)

	case FUNCT_SRA:
		c.executeSRA(inst)

	case FUNCT_SLLV:
		c.executeSLLV(inst)

	case FUNCT_SRLV:
		c.executeSRLV(inst)

	case FUNCT_SRAV:
		c.executeSRAV(inst)

	case FUNCT_JR:
		c.executeJR(inst)

	case FUNCT_JALR:
		c.executeJALR(inst)

	case FUNCT_MOVZ:
		c.executeMOVZ(inst)

	case FUNCT_MOVN:
		c.executeMOVN(inst)

	case FUNCT_SYNC:
		// Ordering is trivially satisfied by an in-order interpreter.
		c.retire()

	case FUNCT_MFHI:
		c.WriteRegister(inst.Rd, c.HI)
		c.retire()

	case FUNCT_MTHI:
		c.HI = c.ReadRegister(inst.Rs)
		c.retire()

	case FUNCT_MFLO:
		c.WriteRegister(inst.Rd, c.LO)
		c.retire()

	case FUNCT_MTLO:
		c.LO = c.ReadRegister(inst.Rs)
		c.retire()

	case FUNCT_MULT:
		c.executeMULT(inst)

	case FUNCT_MULTU:
		c.executeMULTU(inst)

	case FUNCT_DIV:
		c.executeDIV(inst)

	case FUNCT_DIVU:
		c.executeDIVU(inst)

	case FUNCT_ADD:
		c.executeADD(inst)

	case FUNCT_ADDU:
		c.executeADDU(inst)

	case FUNCT_SUB:
		c.executeSUB(inst)

	case FUNCT_SUBU:
		c.executeSUBU(inst)

	case FUNCT_AND:
		c.executeAND(inst)

	case FUNCT_OR:
		c.executeOR(inst)

	case FUNCT_XOR:
		c.executeXOR(inst)

	case FUNCT_NOR:
		c.executeNOR(inst)

	case FUNCT_SLT:
		c.executeSLT(inst)

	case FUNCT_SLTU:
		c.executeSLTU(inst)

	case FUNCT_TGE, FUNCT_TGEU, FUNCT_TLT, FUNCT_TLTU, FUNCT_TEQ, FUNCT_TNE:
		c.executeTrap(inst)

	case FUNCT_SYSCALL:
		c.Exception(EXC_SYS, 0)

	case FUNCT_BREAK:
		c.Exception(EXC_BP, 0)

	default:
		c.Exception(EXC_RI, 0)
	}
}

// executeRegimm handles opcode 1, whose operation is selected by the Rt
// field rather than by funct.
func (c *CPU) executeRegimm(inst Instruction) {

	rs := int32(c.ReadRegister(inst.Rs))

	switch inst.Rt {

	case REGIMM_BLTZ:
		c.conditionalBranch(rs < 0, inst)

	case REGIMM_BGEZ:
		c.conditionalBranch(rs >= 0, inst)

	case REGIMM_BLTZL:
		c.conditionalBranchLikely(rs < 0, inst)

	case REGIMM_BGEZL:
		c.conditionalBranchLikely(rs >= 0, inst)

	case REGIMM_BLTZAL:
		// The link register is written whether or not the branch is
		// taken, and is captured before the branch so that "bltzal $ra"
		// still behaves predictably.
		c.WriteRegister(31, c.CurrentPC+8)
		c.conditionalBranch(rs < 0, inst)

	case REGIMM_BGEZAL:
		c.WriteRegister(31, c.CurrentPC+8)
		c.conditionalBranch(rs >= 0, inst)

	default:
		c.Exception(EXC_RI, 0)
	}
}

// executeSpecial2 handles opcode 28: the multiply-accumulate and
// count-leading-bit instructions.
func (c *CPU) executeSpecial2(inst Instruction) {

	switch inst.Funct {

	case FUNCT2_MUL:
		// MUL writes a GPR and leaves HI/LO architecturally undefined;
		// clearing them keeps behaviour deterministic.
		rs := int32(c.ReadRegister(inst.Rs))
		rt := int32(c.ReadRegister(inst.Rt))
		c.WriteRegister(inst.Rd, uint32(rs*rt))
		c.HI = 0
		c.LO = 0
		c.retire()

	case FUNCT2_MADD:
		rs := int64(int32(c.ReadRegister(inst.Rs)))
		rt := int64(int32(c.ReadRegister(inst.Rt)))
		acc := int64(c.HI)<<32 | int64(c.LO)
		c.setHILO(uint64(acc + rs*rt))
		c.retire()

	case FUNCT2_MADDU:
		rs := uint64(c.ReadRegister(inst.Rs))
		rt := uint64(c.ReadRegister(inst.Rt))
		acc := uint64(c.HI)<<32 | uint64(c.LO)
		c.setHILO(acc + rs*rt)
		c.retire()

	case FUNCT2_MSUB:
		rs := int64(int32(c.ReadRegister(inst.Rs)))
		rt := int64(int32(c.ReadRegister(inst.Rt)))
		acc := int64(c.HI)<<32 | int64(c.LO)
		c.setHILO(uint64(acc - rs*rt))
		c.retire()

	case FUNCT2_MSUBU:
		rs := uint64(c.ReadRegister(inst.Rs))
		rt := uint64(c.ReadRegister(inst.Rt))
		acc := uint64(c.HI)<<32 | uint64(c.LO)
		c.setHILO(acc - rs*rt)
		c.retire()

	case FUNCT2_CLZ:
		c.WriteRegister(inst.Rd, countLeadingZeros(c.ReadRegister(inst.Rs)))
		c.retire()

	case FUNCT2_CLO:
		c.WriteRegister(inst.Rd, countLeadingZeros(^c.ReadRegister(inst.Rs)))
		c.retire()

	default:
		c.Exception(EXC_RI, 0)
	}
}

// executeSpecial3 handles opcode 31: the MIPS32r2 bitfield and sign
// extension instructions that compilers emit freely.
func (c *CPU) executeSpecial3(inst Instruction) {

	switch inst.Funct {

	case FUNCT3_EXT:
		// Extract size bits starting at pos from rs.
		pos := uint32(inst.Shamt)
		size := uint32(inst.Rd) + 1
		if pos+size > 32 {
			c.Exception(EXC_RI, 0)
			return
		}
		value := c.ReadRegister(inst.Rs) >> pos
		if size < 32 {
			value &= (1 << size) - 1
		}
		c.WriteRegister(inst.Rt, value)
		c.retire()

	case FUNCT3_INS:
		// Insert bits [msb:pos] of rs into the same field of rt.
		pos := uint32(inst.Shamt)
		msb := uint32(inst.Rd)
		if msb < pos || msb > 31 {
			c.Exception(EXC_RI, 0)
			return
		}
		size := msb - pos + 1
		var mask uint32 = ^uint32(0)
		if size < 32 {
			mask = (1 << size) - 1
		}
		field := (c.ReadRegister(inst.Rs) & mask) << pos
		kept := c.ReadRegister(inst.Rt) & ^(mask << pos)
		c.WriteRegister(inst.Rt, kept|field)
		c.retire()

	case FUNCT3_BSHFL:
		rt := c.ReadRegister(inst.Rt)
		switch inst.Shamt {
		case BSHFL_WSBH:
			// Swap the bytes within each halfword.
			c.WriteRegister(inst.Rd,
				((rt&0x00FF00FF)<<8)|((rt&0xFF00FF00)>>8))
			c.retire()
		case BSHFL_SEB:
			c.WriteRegister(inst.Rd, uint32(int32(int8(rt))))
			c.retire()
		case BSHFL_SEH:
			c.WriteRegister(inst.Rd, uint32(int32(int16(rt))))
			c.retire()
		default:
			c.Exception(EXC_RI, 0)
		}

	case FUNCT3_RDHWR:
		c.executeRDHWR(inst)

	default:
		c.Exception(EXC_RI, 0)
	}
}

// executeRDHWR reads a hardware register. Only the registers a boot
// loader plausibly touches are provided.
func (c *CPU) executeRDHWR(inst Instruction) {
	if c.CurrentPC < 0x80000000 && c.CP0[CP0_HWRENA]&(uint32(1)<<inst.Rd) == 0 {
		c.Exception(EXC_RI, 0)
		return
	}

	switch inst.Rd {
	case 0: // CPU number
		c.WriteRegister(inst.Rt, 0)
	case 1: // SYNCI step size
		c.WriteRegister(inst.Rt, 32)
	case 2: // Cycle counter
		c.WriteRegister(inst.Rt, uint32(c.Cycles))
	case 3: // Cycle counter resolution
		c.WriteRegister(inst.Rt, 1)
	case 29: // UserLocal/TLS pointer
		c.WriteRegister(inst.Rt, c.UserLocal)
	default:
		c.Exception(EXC_RI, 0)
		return
	}
	c.retire()
}

// setHILO splits a 64-bit result across the HI and LO registers.
func (c *CPU) setHILO(value uint64) {
	c.LO = uint32(value)
	c.HI = uint32(value >> 32)
}

func countLeadingZeros(value uint32) uint32 {
	var n uint32
	for i := 31; i >= 0; i-- {
		if value&(1<<uint(i)) != 0 {
			break
		}
		n++
	}
	return n
}

// coprocessorUnusable raises the trap taken when software touches a
// coprocessor that is not present.
func (c *CPU) coprocessorUnusable(unit uint8) {
	c.CP0[CP0_CAUSE] = (c.CP0[CP0_CAUSE] & ^CAUSE_CE) | (uint32(unit) << 28)
	c.Exception(EXC_CPU, 0)
}

// ---------------------------------------------------------------------
// Arithmetic
// ---------------------------------------------------------------------

// ADD
//
// rd = rs + rt, trapping on signed overflow.
func (c *CPU) executeADD(inst Instruction) {

	rs := int32(c.ReadRegister(inst.Rs))
	rt := int32(c.ReadRegister(inst.Rt))

	result := rs + rt

	// Overflow occurred if the operands share a sign that the result does
	// not.
	if (rs < 0) == (rt < 0) && (result < 0) != (rs < 0) {
		c.Exception(EXC_OV, 0)
		return
	}

	c.WriteRegister(inst.Rd, uint32(result))
	c.retire()
}

// ADDU
//
// rd = rs + rt, without overflow detection.
func (c *CPU) executeADDU(inst Instruction) {
	c.WriteRegister(inst.Rd, c.ReadRegister(inst.Rs)+c.ReadRegister(inst.Rt))
	c.retire()
}

// SUB
//
// rd = rs - rt, trapping on signed overflow.
func (c *CPU) executeSUB(inst Instruction) {

	rs := int32(c.ReadRegister(inst.Rs))
	rt := int32(c.ReadRegister(inst.Rt))

	result := rs - rt

	if (rs < 0) != (rt < 0) && (result < 0) != (rs < 0) {
		c.Exception(EXC_OV, 0)
		return
	}

	c.WriteRegister(inst.Rd, uint32(result))
	c.retire()
}

// SUBU
//
// rd = rs - rt, without overflow detection.
func (c *CPU) executeSUBU(inst Instruction) {
	c.WriteRegister(inst.Rd, c.ReadRegister(inst.Rs)-c.ReadRegister(inst.Rt))
	c.retire()
}

// ADDI
//
// rt = rs + immediate, trapping on signed overflow.
func (c *CPU) executeADDI(inst Instruction) {

	rs := int32(c.ReadRegister(inst.Rs))

	// Sign extend 16-bit immediate
	immediate := int32(int16(inst.Immediate))

	result := rs + immediate

	if (rs < 0) == (immediate < 0) && (result < 0) != (rs < 0) {
		c.Exception(EXC_OV, 0)
		return
	}

	c.WriteRegister(inst.Rt, uint32(result))
	c.retire()
}

// ADDIU
//
// rt = rs + immediate. Despite the name the immediate is sign extended;
// only the overflow trap is suppressed.
func (c *CPU) executeADDIU(inst Instruction) {

	rs := c.ReadRegister(inst.Rs)

	immediate := uint32(int32(int16(inst.Immediate)))

	c.WriteRegister(inst.Rt, rs+immediate)
	c.retire()
}

func (c *CPU) executeMULT(inst Instruction) {
	rs := int64(int32(c.ReadRegister(inst.Rs)))
	rt := int64(int32(c.ReadRegister(inst.Rt)))
	c.setHILO(uint64(rs * rt))
	c.retire()
}

func (c *CPU) executeMULTU(inst Instruction) {
	rs := uint64(c.ReadRegister(inst.Rs))
	rt := uint64(c.ReadRegister(inst.Rt))
	c.setHILO(rs * rt)
	c.retire()
}

func (c *CPU) executeDIV(inst Instruction) {
	rs := int32(c.ReadRegister(inst.Rs))
	rt := int32(c.ReadRegister(inst.Rt))

	// MIPS leaves HI/LO undefined rather than trapping on a zero divisor
	// or on the one overflowing case, so no exception is raised here.
	if rt == 0 {
		c.HI = uint32(rs)
		if rs >= 0 {
			c.LO = 0xFFFFFFFF
		} else {
			c.LO = 1
		}
		c.retire()
		return
	}

	if rs == -0x80000000 && rt == -1 {
		c.LO = 0x80000000
		c.HI = 0
		c.retire()
		return
	}

	c.LO = uint32(rs / rt)
	c.HI = uint32(rs % rt)
	c.retire()
}

func (c *CPU) executeDIVU(inst Instruction) {
	rs := c.ReadRegister(inst.Rs)
	rt := c.ReadRegister(inst.Rt)

	if rt == 0 {
		c.LO = 0xFFFFFFFF
		c.HI = rs
		c.retire()
		return
	}

	c.LO = rs / rt
	c.HI = rs % rt
	c.retire()
}

// ---------------------------------------------------------------------
// Logical and comparison
// ---------------------------------------------------------------------

func (c *CPU) executeLUI(inst Instruction) {
	c.WriteRegister(inst.Rt, uint32(inst.Immediate)<<16)
	c.retire()
}

func (c *CPU) executeANDI(inst Instruction) {
	c.WriteRegister(inst.Rt, c.ReadRegister(inst.Rs)&uint32(inst.Immediate))
	c.retire()
}

func (c *CPU) executeORI(inst Instruction) {
	c.WriteRegister(inst.Rt, c.ReadRegister(inst.Rs)|uint32(inst.Immediate))
	c.retire()
}

func (c *CPU) executeXORI(inst Instruction) {
	c.WriteRegister(inst.Rt, c.ReadRegister(inst.Rs)^uint32(inst.Immediate))
	c.retire()
}

func (c *CPU) executeAND(inst Instruction) {
	c.WriteRegister(inst.Rd, c.ReadRegister(inst.Rs)&c.ReadRegister(inst.Rt))
	c.retire()
}

func (c *CPU) executeOR(inst Instruction) {
	c.WriteRegister(inst.Rd, c.ReadRegister(inst.Rs)|c.ReadRegister(inst.Rt))
	c.retire()
}

func (c *CPU) executeXOR(inst Instruction) {
	c.WriteRegister(inst.Rd, c.ReadRegister(inst.Rs)^c.ReadRegister(inst.Rt))
	c.retire()
}

func (c *CPU) executeNOR(inst Instruction) {
	c.WriteRegister(inst.Rd, ^(c.ReadRegister(inst.Rs) | c.ReadRegister(inst.Rt)))
	c.retire()
}

func (c *CPU) executeSLT(inst Instruction) {
	rs := int32(c.ReadRegister(inst.Rs))
	rt := int32(c.ReadRegister(inst.Rt))
	c.WriteRegister(inst.Rd, boolToWord(rs < rt))
	c.retire()
}

func (c *CPU) executeSLTU(inst Instruction) {
	rs := c.ReadRegister(inst.Rs)
	rt := c.ReadRegister(inst.Rt)
	c.WriteRegister(inst.Rd, boolToWord(rs < rt))
	c.retire()
}

func (c *CPU) executeSLTI(inst Instruction) {
	rs := int32(c.ReadRegister(inst.Rs))
	imm := int32(int16(inst.Immediate))
	c.WriteRegister(inst.Rt, boolToWord(rs < imm))
	c.retire()
}

// SLTIU compares without sign, but the immediate is still sign extended
// before the comparison.
func (c *CPU) executeSLTIU(inst Instruction) {
	rs := c.ReadRegister(inst.Rs)
	imm := uint32(int32(int16(inst.Immediate)))
	c.WriteRegister(inst.Rt, boolToWord(rs < imm))
	c.retire()
}

func boolToWord(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}

// ---------------------------------------------------------------------
// Shifts
// ---------------------------------------------------------------------

func (c *CPU) executeSLL(inst Instruction) {
	c.WriteRegister(inst.Rd, c.ReadRegister(inst.Rt)<<inst.Shamt)
	c.retire()
}

func (c *CPU) executeSRL(inst Instruction) {
	value := c.ReadRegister(inst.Rt)
	if inst.Rs == 1 {
		c.WriteRegister(inst.Rd, rotateRight32(value, uint32(inst.Shamt)))
	} else {
		c.WriteRegister(inst.Rd, value>>inst.Shamt)
	}
	c.retire()
}

func (c *CPU) executeSRA(inst Instruction) {
	c.WriteRegister(inst.Rd, uint32(int32(c.ReadRegister(inst.Rt))>>inst.Shamt))
	c.retire()
}

func (c *CPU) executeSLLV(inst Instruction) {
	shift := c.ReadRegister(inst.Rs) & 0x1F
	c.WriteRegister(inst.Rd, c.ReadRegister(inst.Rt)<<shift)
	c.retire()
}

func (c *CPU) executeSRLV(inst Instruction) {
	shift := c.ReadRegister(inst.Rs) & 0x1F
	value := c.ReadRegister(inst.Rt)
	if inst.Shamt == 1 {
		c.WriteRegister(inst.Rd, rotateRight32(value, shift))
	} else {
		c.WriteRegister(inst.Rd, value>>shift)
	}
	c.retire()
}

func (c *CPU) executeSRAV(inst Instruction) {
	shift := c.ReadRegister(inst.Rs) & 0x1F
	c.WriteRegister(inst.Rd, uint32(int32(c.ReadRegister(inst.Rt))>>shift))
	c.retire()
}

func rotateRight32(value uint32, shift uint32) uint32 {
	shift &= 0x1F
	if shift == 0 {
		return value
	}
	return (value >> shift) | (value << (32 - shift))
}

// ---------------------------------------------------------------------
// Conditional moves
// ---------------------------------------------------------------------

func (c *CPU) executeMOVZ(inst Instruction) {
	if c.ReadRegister(inst.Rt) == 0 {
		c.WriteRegister(inst.Rd, c.ReadRegister(inst.Rs))
	}
	c.retire()
}

func (c *CPU) executeMOVN(inst Instruction) {
	if c.ReadRegister(inst.Rt) != 0 {
		c.WriteRegister(inst.Rd, c.ReadRegister(inst.Rs))
	}
	c.retire()
}

// ---------------------------------------------------------------------
// Jumps and branches
//
// All of these set the branch target through c.branch so that the delay
// slot instruction executes before control transfers.
// ---------------------------------------------------------------------

// jumpTarget computes the absolute destination of a J-format
// instruction. The high 4 bits come from the delay slot's address, which
// is CurrentPC+4.
func (c *CPU) jumpTarget(inst Instruction) uint32 {
	return ((c.CurrentPC + 4) & 0xF0000000) | (inst.Target << 2)
}

// branchTargetOf computes the absolute destination of an I-format
// branch, whose offset is relative to the delay slot.
func (c *CPU) branchTargetOf(inst Instruction) uint32 {
	return uint32(int32(c.CurrentPC+4) + (int32(int16(inst.Immediate)) << 2))
}

func (c *CPU) executeJ(inst Instruction) {
	c.branch(c.jumpTarget(inst))
	c.retire()
}

func (c *CPU) executeJAL(inst Instruction) {
	// The return address skips the delay slot.
	c.WriteRegister(31, c.CurrentPC+8)
	c.branch(c.jumpTarget(inst))
	c.retire()
}

func (c *CPU) executeJR(inst Instruction) {
	c.branch(c.ReadRegister(inst.Rs))
	c.retire()
}

func (c *CPU) executeJALR(inst Instruction) {
	// The target is read before the link register is written, so that
	// "jalr $ra, $ra" works.
	target := c.ReadRegister(inst.Rs)

	// An omitted destination defaults to $ra.
	rd := inst.Rd
	if rd == 0 {
		rd = 31
	}

	c.WriteRegister(rd, c.CurrentPC+8)
	c.branch(target)
	c.retire()
}

// conditionalBranch takes the branch if cond holds, and otherwise simply
// falls through into the delay slot.
func (c *CPU) conditionalBranch(cond bool, inst Instruction) {
	if cond {
		c.branch(c.branchTargetOf(inst))
	}
	c.retire()
}

// conditionalBranchLikely implements the "branch likely" forms, which
// discard the delay slot instruction when the branch is not taken.
func (c *CPU) conditionalBranchLikely(cond bool, inst Instruction) {
	if cond {
		c.branch(c.branchTargetOf(inst))
	} else {
		c.nullifyDelaySlot()
	}
	c.retire()
}

func (c *CPU) executeBEQ(inst Instruction) {
	c.conditionalBranch(c.ReadRegister(inst.Rs) == c.ReadRegister(inst.Rt), inst)
}

func (c *CPU) executeBNE(inst Instruction) {
	c.conditionalBranch(c.ReadRegister(inst.Rs) != c.ReadRegister(inst.Rt), inst)
}

func (c *CPU) executeBLEZ(inst Instruction) {
	c.conditionalBranch(int32(c.ReadRegister(inst.Rs)) <= 0, inst)
}

func (c *CPU) executeBGTZ(inst Instruction) {
	c.conditionalBranch(int32(c.ReadRegister(inst.Rs)) > 0, inst)
}

func (c *CPU) executeBEQL(inst Instruction) {
	c.conditionalBranchLikely(c.ReadRegister(inst.Rs) == c.ReadRegister(inst.Rt), inst)
}

func (c *CPU) executeBNEL(inst Instruction) {
	c.conditionalBranchLikely(c.ReadRegister(inst.Rs) != c.ReadRegister(inst.Rt), inst)
}

func (c *CPU) executeBLEZL(inst Instruction) {
	c.conditionalBranchLikely(int32(c.ReadRegister(inst.Rs)) <= 0, inst)
}

func (c *CPU) executeBGTZL(inst Instruction) {
	c.conditionalBranchLikely(int32(c.ReadRegister(inst.Rs)) > 0, inst)
}

// executeTrap implements the conditional trap instructions.
func (c *CPU) executeTrap(inst Instruction) {

	rsU := c.ReadRegister(inst.Rs)
	rtU := c.ReadRegister(inst.Rt)
	rsS := int32(rsU)
	rtS := int32(rtU)

	var cond bool
	switch inst.Funct {
	case FUNCT_TGE:
		cond = rsS >= rtS
	case FUNCT_TGEU:
		cond = rsU >= rtU
	case FUNCT_TLT:
		cond = rsS < rtS
	case FUNCT_TLTU:
		cond = rsU < rtU
	case FUNCT_TEQ:
		cond = rsU == rtU
	case FUNCT_TNE:
		cond = rsU != rtU
	}

	if cond {
		c.Exception(EXC_TR, 0)
		return
	}
	c.retire()
}

// ---------------------------------------------------------------------
// Loads and stores
// ---------------------------------------------------------------------

// effectiveAddress computes rs + sign-extended offset.
func (c *CPU) effectiveAddress(inst Instruction) uint32 {
	base := c.ReadRegister(inst.Rs)
	offset := uint32(int32(int16(inst.Immediate)))
	return base + offset
}

// checkLoad validates a load address for alignment and mapping,
// raising AdEL and returning false if it is unusable.
func (c *CPU) checkLoad(addr uint32, align uint32) bool {
	if addr&(align-1) != 0 {
		c.Exception(EXC_ADEL, addr)
		return false
	}
	if !c.Bus.HasMapping(addr) {
		if isTLBMappedSegment(addr) {
			if _, _, index := c.lookupTLB(addr, false); index >= 0 {
				c.exceptionNoRefill(EXC_TLBL, addr)
			} else {
				c.Exception(EXC_TLBL, addr)
			}
			return false
		}
		c.Exception(EXC_ADEL, addr)
		return false
	}
	return true
}

// checkStore validates a store address, raising AdES on failure.
func (c *CPU) checkStore(addr uint32, align uint32) bool {
	if addr&(align-1) != 0 {
		c.Exception(EXC_ADES, addr)
		return false
	}
	if !c.Bus.HasMapping(addr) {
		if isTLBMappedSegment(addr) {
			if _, _, index := c.lookupTLB(addr, true); index >= 0 {
				c.exceptionNoRefill(EXC_TLBS, addr)
			} else {
				c.Exception(EXC_TLBS, addr)
			}
			return false
		}
		c.Exception(EXC_ADES, addr)
		return false
	}
	return true
}

// LW
//
// rt = memory[rs + immediate]
func (c *CPU) executeLW(inst Instruction) {
	addr := c.effectiveAddress(inst)
	if !c.checkLoad(addr, 4) {
		return
	}
	c.WriteRegister(inst.Rt, c.read32(addr))
	c.retire()
}

// SW
//
// memory[rs + immediate] = rt
func (c *CPU) executeSW(inst Instruction) {
	addr := c.effectiveAddress(inst)
	if !c.checkStore(addr, 4) {
		return
	}
	c.write32(addr, c.ReadRegister(inst.Rt))
	c.retire()
}

func (c *CPU) executeLB(inst Instruction) {
	addr := c.effectiveAddress(inst)
	if !c.checkLoad(addr, 1) {
		return
	}
	c.WriteRegister(inst.Rt, uint32(int32(int8(c.read8(addr)))))
	c.retire()
}

func (c *CPU) executeLBU(inst Instruction) {
	addr := c.effectiveAddress(inst)
	if !c.checkLoad(addr, 1) {
		return
	}
	c.WriteRegister(inst.Rt, uint32(c.read8(addr)))
	c.retire()
}

func (c *CPU) executeLH(inst Instruction) {
	addr := c.effectiveAddress(inst)
	if !c.checkLoad(addr, 2) {
		return
	}
	c.WriteRegister(inst.Rt, uint32(int32(int16(c.read16(addr)))))
	c.retire()
}

func (c *CPU) executeLHU(inst Instruction) {
	addr := c.effectiveAddress(inst)
	if !c.checkLoad(addr, 2) {
		return
	}
	c.WriteRegister(inst.Rt, uint32(c.read16(addr)))
	c.retire()
}

func (c *CPU) executeSB(inst Instruction) {
	addr := c.effectiveAddress(inst)
	if !c.checkStore(addr, 1) {
		return
	}
	c.write8(addr, byte(c.ReadRegister(inst.Rt)))
	c.retire()
}

func (c *CPU) executeSH(inst Instruction) {
	addr := c.effectiveAddress(inst)
	if !c.checkStore(addr, 2) {
		return
	}
	c.write16(addr, uint16(c.ReadRegister(inst.Rt)))
	c.retire()
}

func (c *CPU) executeLWC1(inst Instruction) {
	addr := c.effectiveAddress(inst)
	if !c.checkLoad(addr, 4) {
		return
	}
	c.FPR[inst.Rt] = c.read32(addr)
	c.retire()
}

func (c *CPU) executeSWC1(inst Instruction) {
	addr := c.effectiveAddress(inst)
	if !c.checkStore(addr, 4) {
		return
	}
	c.write32(addr, c.FPR[inst.Rt])
	c.retire()
}

func (c *CPU) executeLDC1(inst Instruction) {
	addr := c.effectiveAddress(inst)
	if !c.checkLoad(addr, 8) {
		return
	}
	c.FPR[inst.Rt] = c.read32(addr)
	c.FPR[(inst.Rt+1)&31] = c.read32(addr + 4)
	c.retire()
}

func (c *CPU) executeSDC1(inst Instruction) {
	addr := c.effectiveAddress(inst)
	if !c.checkStore(addr, 8) {
		return
	}
	c.write32(addr, c.FPR[inst.Rt])
	c.write32(addr+4, c.FPR[(inst.Rt+1)&31])
	c.retire()
}

// LWL loads the unaligned left-hand portion of a word, merging it into
// the high bytes of rt.
func (c *CPU) executeLWL(inst Instruction) {
	addr := c.effectiveAddress(inst)
	aligned := addr & ^uint32(3)
	if !c.checkLoad(aligned, 4) {
		return
	}

	word := c.read32(aligned)

	// The addressed byte becomes the most significant byte of rt, with
	// lower-addressed bytes of the same word filling downwards. Offset 0
	// therefore contributes one byte and offset 3 the whole word.
	shift := (addr & 3) * 8

	// keep covers the bytes of rt left untouched.
	keep := uint32(0x00FFFFFF) >> shift

	current := c.ReadRegister(inst.Rt)
	c.WriteRegister(inst.Rt, (current&keep)|(word<<(24-shift)))
	c.retire()
}

// LWR loads the unaligned right-hand portion of a word, merging it into
// the low bytes of rt.
func (c *CPU) executeLWR(inst Instruction) {
	addr := c.effectiveAddress(inst)
	aligned := addr & ^uint32(3)
	if !c.checkLoad(aligned, 4) {
		return
	}

	word := c.read32(aligned)

	// The addressed byte becomes the least significant byte of rt, with
	// higher-addressed bytes of the same word filling upwards. Offset 0
	// therefore contributes the whole word and offset 3 one byte.
	shift := (addr & 3) * 8

	// A shift count of 32 or more yields zero in Go, which is exactly the
	// "keep nothing" case wanted at offset 0.
	keep := uint32(0xFFFFFF00) << (24 - shift)

	current := c.ReadRegister(inst.Rt)
	c.WriteRegister(inst.Rt, (current&keep)|(word>>shift))
	c.retire()
}

// SWL stores the high bytes of rt into the unaligned left-hand portion
// of the addressed word.
func (c *CPU) executeSWL(inst Instruction) {
	addr := c.effectiveAddress(inst)
	aligned := addr & ^uint32(3)
	if !c.checkStore(aligned, 4) {
		return
	}

	shift := (addr & 3) * 8
	value := c.ReadRegister(inst.Rt)

	mask := uint32(0xFFFFFFFF) >> (24 - shift)

	word := c.read32(aligned)
	c.write32(aligned, (word & ^mask)|((value>>(24-shift))&mask))
	c.retire()
}

// SWR stores the low bytes of rt into the unaligned right-hand portion
// of the addressed word.
func (c *CPU) executeSWR(inst Instruction) {
	addr := c.effectiveAddress(inst)
	aligned := addr & ^uint32(3)
	if !c.checkStore(aligned, 4) {
		return
	}

	shift := (addr & 3) * 8
	value := c.ReadRegister(inst.Rt)

	mask := uint32(0xFFFFFFFF) << shift

	word := c.read32(aligned)
	c.write32(aligned, (word & ^mask)|((value<<shift)&mask))
	c.retire()
}

// LL is a plain load that additionally arms the load-linked bit.
func (c *CPU) executeLL(inst Instruction) {
	addr := c.effectiveAddress(inst)
	if !c.checkLoad(addr, 4) {
		return
	}
	c.WriteRegister(inst.Rt, c.read32(addr))
	c.CP0[CP0_LLADDR] = addr >> 4
	c.LLBit = true
	c.retire()
}

// SC stores conditionally on the load-linked bit. With no competing bus
// masters the store always succeeds if LL armed it.
func (c *CPU) executeSC(inst Instruction) {
	addr := c.effectiveAddress(inst)
	if !c.checkStore(addr, 4) {
		return
	}

	if c.LLBit {
		c.write32(addr, c.ReadRegister(inst.Rt))
		c.WriteRegister(inst.Rt, 1)
	} else {
		c.WriteRegister(inst.Rt, 0)
	}

	c.LLBit = false
	c.retire()
}

func (c *CPU) executeCACHE(inst Instruction) {
	// There is no cache model, so cache maintenance is a no-op. The
	// address is deliberately not validated: boot code routinely
	// invalidates ranges it has not mapped yet.
	c.retire()
}

// ---------------------------------------------------------------------
// Coprocessor 0
// ---------------------------------------------------------------------

func (c *CPU) executeCOP0(inst Instruction) {

	switch inst.Rs {

	case COP0_MFC0:
		c.executeMFC0(inst)

	case COP0_MTC0:
		c.executeMTC0(inst)

	case COP0_CO:
		switch inst.Funct {
		case COP0CO_ERET:
			c.executeERET(inst)
		case COP0CO_DERET:
			c.executeERET(inst)
		case COP0CO_WAIT:
			c.Waiting = true
			c.retire()
		case COP0CO_TLBR:
			c.readIndexedTLB(int(c.CP0[CP0_INDEX] & 31))
			c.retire()
		case COP0CO_TLBWI:
			c.writeIndexedTLB(int(c.CP0[CP0_INDEX] & 31))
			c.retire()
		case COP0CO_TLBWR:
			c.writeIndexedTLB(c.randomTLBIndex())
			c.retire()
		case COP0CO_TLBP:
			c.probeTLB()
			c.retire()
		default:
			c.Exception(EXC_RI, 0)
		}

	default:
		c.Exception(EXC_RI, 0)
	}
}

func (c *CPU) executeCOP1(inst Instruction) {
	switch inst.Rs {
	case COP1_MFC1:
		c.WriteRegister(inst.Rt, c.FPR[inst.Rd])
		c.retire()
	case COP1_MTC1:
		c.FPR[inst.Rd] = c.ReadRegister(inst.Rt)
		c.retire()
	case COP1_CFC1:
		switch inst.Rd {
		case 0:
			c.WriteRegister(inst.Rt, 0)
		case 31:
			c.WriteRegister(inst.Rt, c.FCSR)
		default:
			c.WriteRegister(inst.Rt, 0)
		}
		c.retire()
	case COP1_CTC1:
		if inst.Rd == 31 {
			c.FCSR = c.ReadRegister(inst.Rt)
		}
		c.retire()
	default:
		c.Exception(EXC_RI, 0)
	}
}

func (c *CPU) executeMFC0(inst Instruction) {
	c.WriteRegister(inst.Rt, c.readCP0(inst.Rd, inst.Raw&7))
	c.retire()
}

func (c *CPU) executeMTC0(inst Instruction) {
	c.writeCP0(inst.Rd, inst.Raw&7, c.ReadRegister(inst.Rt))
	c.retire()
}

func (c *CPU) readCP0(rd uint8, sel uint32) uint32 {
	switch rd {
	case CP0_CONTEXT:
		if sel == 2 {
			return c.UserLocal
		}
	case CP0_COUNT:
		return c.cp0Count()
	case CP0_CONFIG:
		switch sel {
		case 1:
			return CP0_CONFIG1_RESET
		case 2:
			return CP0_CONFIG2_RESET
		case 3:
			return CP0_CONFIG3_RESET
		}
	}
	return c.CP0[rd]
}

func (c *CPU) writeCP0(rd uint8, sel uint32, value uint32) {
	switch rd {
	case CP0_CONTEXT:
		if sel == 2 {
			c.UserLocal = value
			return
		}
	case CP0_COUNT:
		c.countBaseValue = value
		c.countBaseCycle = c.Cycles
		c.CP0[CP0_COUNT] = value
		return
	case CP0_COMPARE:
		c.CP0[CP0_COMPARE] = value
		c.compareSet = true
		c.CP0[CP0_CAUSE] &^= CAUSE_IP7
		return
	case CP0_CONFIG:
		if sel != 0 {
			return
		}
	}
	c.CP0[rd] = value
}

func (c *CPU) executeERET(inst Instruction) {
	// ERET returns through ErrorEPC when the error level flag is set,
	// and through EPC otherwise.
	if (c.CP0[CP0_STATUS] & STATUS_ERL) != 0 {
		c.PC = c.CP0[CP0_ERROREPC]
		c.CP0[CP0_STATUS] &= ^STATUS_ERL
	} else {
		c.PC = c.CP0[CP0_EPC]
		c.CP0[CP0_STATUS] &= ^STATUS_EXL
	}

	c.NextPC = c.PC + 4

	// ERET is not a branch and has no delay slot.
	c.branchTaken = false
	c.LLBit = false

	c.retire()
}
