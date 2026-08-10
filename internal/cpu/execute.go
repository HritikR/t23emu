package cpu

import "math"

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
		if !c.cop1Usable() {
			c.coprocessorUnusable(1)
			break
		}
		c.executeLWC1(inst)

	case OP_LDC1:
		if !c.cop1Usable() {
			c.coprocessorUnusable(1)
			break
		}
		c.executeLDC1(inst)

	case OP_SC:
		c.executeSC(inst)

	case OP_SWC1:
		if !c.cop1Usable() {
			c.coprocessorUnusable(1)
			break
		}
		c.executeSWC1(inst)

	case OP_SDC1:
		if !c.cop1Usable() {
			c.coprocessorUnusable(1)
			break
		}
		c.executeSDC1(inst)

	case OP_CACHE:
		c.executeCACHE(inst)

	case OP_PREF:
		// Prefetch is a hint with no architectural effect.
		c.retire()

	case OP_COP0:
		c.executeCOP0(inst)

	case OP_COP1:
		if !c.cop1Usable() {
			c.coprocessorUnusable(1)
			break
		}
		c.executeCOP1(inst)

	case OP_COP2:
		c.coprocessorUnusable(inst.Opcode - OP_COP0)

	default:
		c.reservedInstruction(inst)
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
		c.reservedInstruction(inst)
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
		c.reservedInstruction(inst)
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
		c.reservedInstruction(inst)
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
			c.reservedInstruction(inst)
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
			c.reservedInstruction(inst)
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
			c.reservedInstruction(inst)
		}

	case FUNCT3_RDHWR:
		c.executeRDHWR(inst)

	default:
		c.reservedInstruction(inst)
	}
}

// executeRDHWR reads a hardware register. Only the registers a boot
// loader plausibly touches are provided.
func (c *CPU) executeRDHWR(inst Instruction) {
	// Enforce the HWRENA permission gate for user-mode accesses. When
	// HWRENA[rd] is 0, rdhwr traps as RI and the kernel's RI handler
	// emulates it (reading tp_value from thread_info for $29, etc.).
	if c.CurrentPC < 0x80000000 && c.CP0[CP0_HWRENA]&(uint32(1)<<inst.Rd) == 0 {
		// Fast path for rdhwr $29 (TLS pointer): if we've cached a
		// non-zero UserLocal AND the ASID matches (same process),
		// return it directly.
		if inst.Rd == 29 && c.UserLocal != 0 {
			currentASID := c.CP0[CP0_ENTRYHI] & 0xFF
			if currentASID == c.cachedTLSASID {
				c.WriteRegister(inst.Rt, c.UserLocal)
				c.retire()
				return
			}
			// ASID changed — context switch detected. Fall
			// through to re-trap and refresh the cache.
			c.UserLocal = 0
		}
		// For rdhwr $29, mark this as pending so that when the kernel
		// finishes emulating it (returns via eret) we can observe the
		// value and cache it.
		if inst.Rd == 29 {
			c.pendingTLSPC = c.CurrentPC
			c.pendingTLSRt = inst.Rt
		}
		c.reservedInstruction(inst)
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
		c.reservedInstruction(inst)
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

// cop1Usable reports whether the COP1 (FPU) may be accessed from the
// current privilege level. CU1 grants access unconditionally; EXL or ERL
// (exception/error mode) also grants access so the kernel's FPU trap
// handler can save/restore state. Kernel mode (KSU=0) also permits
// access, matching many MIPS implementations that allow the kernel to
// use CP1 without explicitly setting CU1.
func (c *CPU) cop1Usable() bool {
	status := c.CP0[CP0_STATUS]
	if status&STATUS_CU1 != 0 {
		return true
	}
	if status&(STATUS_EXL|STATUS_ERL) != 0 {
		return true
	}
	return status&STATUS_KSU == 0
}

// readFPR_S reads FPR[idx] as a single-precision float.
func (c *CPU) readFPR_S(idx uint8) float32 {
	return math.Float32frombits(c.FPR[idx&31])
}

// writeFPR_S writes a single-precision float into FPR[idx].
func (c *CPU) writeFPR_S(idx uint8, v float32) {
	c.FPR[idx&31] = math.Float32bits(v)
}

// readFPR_D reads the even/odd lane pair at FPR[idx] as a double.
// MIPS requires idx to be even for double access; the LSB is masked so
// that an odd index silently aliases to the previous register, matching
// the behaviour of the FR=0 FPR register file used by 32-bit MIPS.
func (c *CPU) readFPR_D(idx uint8) float64 {
	base := idx & 30
	bits := uint64(c.FPR[base]) | uint64(c.FPR[base+1])<<32
	return math.Float64frombits(bits)
}

// writeFPR_D writes a double into the even/odd lane pair at FPR[idx].
func (c *CPU) writeFPR_D(idx uint8, v float64) {
	base := idx & 30
	bits := math.Float64bits(v)
	c.FPR[base] = uint32(bits)
	c.FPR[base+1] = uint32(bits >> 32)
}

// readFPR_W reads FPR[idx] as a 32-bit signed integer (the W format).
func (c *CPU) readFPR_W(idx uint8) int32 {
	return int32(c.FPR[idx&31])
}

// writeFPR_W writes a 32-bit signed integer into FPR[idx] (the W format).
func (c *CPU) writeFPR_W(idx uint8, v int32) {
	c.FPR[idx&31] = uint32(v)
}

// fccBit returns the FCSR bit mask for condition code cc. FCC0 lives at
// FCSR bit 23 and FCC1..FCC7 occupy the next seven bits upward.
func fccBit(cc uint8) uint32 {
	return uint32(1) << (23 + cc&7)
}

// readFCC returns condition code cc from FCSR.
func (c *CPU) readFCC(cc uint8) bool {
	return c.FCSR&fccBit(cc) != 0
}

// setFCC writes condition code cc into FCSR.
func (c *CPU) setFCC(cc uint8, v bool) {
	if v {
		c.FCSR |= fccBit(cc)
	} else {
		c.FCSR &^= fccBit(cc)
	}
}

// roundWithMode rounds a double to a 32-bit signed integer using the
// rounding mode selected by FCSR.RM. Used by cvt.w.fmt.
func (c *CPU) roundWithMode(v float64) int32 {
	switch c.FCSR & FCSR_RMMASK {
	case FP_RZ:
		return int32(math.Trunc(v))
	case FP_RP:
		return int32(math.Ceil(v))
	case FP_RM:
		return int32(math.Floor(v))
	default: // FP_RN
		return int32(math.Round(v))
	}
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

// tryFastDelayLoop detects the kernel's __delay() busy-wait loop and
// fast-forwards through it. The genuine delay-loop pattern uses plain
// volatile loads/stores as compiler barriers — NEVER LL/SC, which are
// atomic operations. We must reject LL/SC loops to avoid corrupting
// atomic_inc, spin_lock, and similar primitives.
//
// Genuine delay pattern:
//
//	target:   lw/lwc1  $tN, off($base)    ; volatile barrier load
//	target+4: addiu    $v1, $v1, 1         ; 0x24630001
//	target+8: sw/swc1  $tN, off($base)    ; volatile barrier store
//	target+C: beq      $v1, $zero, target  ; backward branch
//	target+10: nop                         ; delay slot
func (c *CPU) tryFastDelayLoop(inst Instruction) bool {
	if inst.Rs != 3 || inst.Rt != 0 {
		return false
	}

	v1 := c.ReadRegister(3)
	if v1 == 0 {
		return false
	}

	target := c.branchTargetOf(inst)

	// Must be a tight backward branch. Only the 3-instruction body
	// (target+0, target+4, target+8) + branch at CurrentPC is supported.
	if c.CurrentPC-target != 12 {
		return false
	}

	// Check addiu $v1, $v1, 1 at target+4.
	if c.Bus.Read32(target+4) != 0x24630001 {
		return false
	}

	// Reject LL/SC atomic loops. Load at target must be LW(0x23) or
	// LWC1(0x31), NOT LL(0x30). Store at target+8 must be SW(0x2B) or
	// SWC1(0x39), NOT SC(0x38).
	loadRaw := c.Bus.Read32(target)
	loadOp := (loadRaw >> 26) & 0x3F
	if loadOp != uint32(OP_LW) && loadOp != uint32(OP_LWC1) {
		return false
	}

	storeRaw := c.Bus.Read32(target + 8)
	storeOp := (storeRaw >> 26) & 0x3F
	if storeOp != uint32(OP_SW) && storeOp != uint32(OP_SWC1) {
		return false
	}

	// This is a delay loop. v1 wraps to 0 after (2^32 - v1) more increments.
	remaining := uint64(uint32(-int32(v1)))

	// Each iteration is 5 instructions (load, addiu, store, beq, nop).
	c.Cycles += remaining * 5

	// Exit the loop.
	c.WriteRegister(3, 0)

	return true
}

func (c *CPU) executeBEQ(inst Instruction) {
	if c.tryFastDelayLoop(inst) {
		c.retire()
		return
	}
	c.conditionalBranch(c.ReadRegister(inst.Rs) == c.ReadRegister(inst.Rt), inst)
}

// tryFastDelayBNE detects the standard MIPS __delay() function:
//
//	__delay:
//	  bne  $R, $zero, __delay    ; self-branch (offset = -1)
//	  addiu $R, $R, -1           ; delay slot: decrement counter
//	  jr   $ra
//	  nop
//
// When detected with $R != 0, advance Cycles by $R × 2 and set $R = 0
// so the branch falls through.  The delay slot still executes normally
// (decrementing $R to 0xFFFFFFFF), which matches real hardware behavior.
func (c *CPU) tryFastDelayBNE(inst Instruction) bool {
	if inst.Rt != 0 || inst.Rs == 0 {
		return false
	}

	offset := int32(int16(inst.Immediate))
	if offset != -1 {
		return false
	}

	regIdx := inst.Rs
	regVal := c.ReadRegister(regIdx)
	if regVal == 0 {
		return false
	}

	// Verify the delay slot is addiu $R, $R, -1.
	delayRaw := c.Bus.Read32(c.CurrentPC + 4)
	expectedDelay := uint32(OP_ADDIU)<<26 |
		uint32(regIdx)<<21 | uint32(regIdx)<<16 | 0xFFFF
	if delayRaw != expectedDelay {
		return false
	}

	// Each iteration: BNE + ADDIU delay slot = 2 cycles.
	// The current BNE and its delay slot will execute normally (the exit
	// iteration), so we only advance for the skipped iterations.
	c.Cycles += uint64(regVal) * 2

	c.WriteRegister(regIdx, 0)

	return true
}

func (c *CPU) executeBNE(inst Instruction) {
	if c.tryFastDelayBNE(inst) {
		c.retire()
		return
	}
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
		if c.requiresTLB(addr) {
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
	if c.requiresTLB(addr) {
		if _, ok, index := c.lookupTLB(addr, true); !ok {
			if index >= 0 {
				entryLo := c.TLB[index].entryLo(addr)
				if entryLo&entryLoV != 0 && entryLo&entryLoD == 0 {
					c.Exception(EXC_MOD, addr)
				} else {
					c.exceptionNoRefill(EXC_TLBS, addr)
				}
			} else {
				c.Exception(EXC_TLBS, addr)
			}
			return false
		}
		if !c.Bus.HasMapping(addr) {
			c.Exception(EXC_ADES, addr)
			return false
		}
		return true
	}
	if !c.Bus.HasMapping(addr) {
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
			c.TLBWI++
			c.writeIndexedTLB(int(c.CP0[CP0_INDEX] & 31))
			c.retire()
		case COP0CO_TLBWR:
			c.TLBWR++
			c.writeIndexedTLB(c.randomTLBIndex())
			c.retire()
		case COP0CO_TLBP:
			c.probeTLB()
			c.retire()
		default:
			c.reservedInstruction(inst)
		}

	default:
		c.reservedInstruction(inst)
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
		// FIR: advertise FPU with S (14), D (15), and W (16) format
		// support, no paired single, no fused multiply-add.
		c.WriteRegister(inst.Rt, 0x0001C000)
		case 25:
			// FEXR: cause/flags lower half.
			c.WriteRegister(inst.Rt, c.FCSR&(FCSR_CAUSE|FCSR_FLAGS))
		case 26:
			// FENR: enables plus rounding mode, and FCC1..FCC7.
			c.WriteRegister(inst.Rt, c.FCSR&(FCSR_ENABLE|FCSR_RMMASK|FCSR_ALLFCC&^FCSR_FCC0))
		case 28:
			// FCCR: FCC0 plus the enables and rounding mode.
			c.WriteRegister(inst.Rt, c.FCSR&(FCSR_FCC0|FCSR_ENABLE|FCSR_RMMASK))
		case 31:
			c.WriteRegister(inst.Rt, c.FCSR)
		default:
			c.WriteRegister(inst.Rt, 0)
		}
		c.retire()
	case COP1_CTC1:
		switch inst.Rd {
		case 25:
			c.FCSR = (c.FCSR & ^(FCSR_CAUSE | FCSR_FLAGS)) | (c.ReadRegister(inst.Rt) & (FCSR_CAUSE | FCSR_FLAGS))
		case 26:
			c.FCSR = (c.FCSR & ^(FCSR_ENABLE | FCSR_RMMASK | FCSR_ALLFCC&^FCSR_FCC0)) |
				(c.ReadRegister(inst.Rt) & (FCSR_ENABLE | FCSR_RMMASK | FCSR_ALLFCC&^FCSR_FCC0))
		case 28:
			c.FCSR = (c.FCSR & ^(FCSR_FCC0 | FCSR_ENABLE | FCSR_RMMASK)) |
				(c.ReadRegister(inst.Rt) & (FCSR_FCC0 | FCSR_ENABLE | FCSR_RMMASK))
		case 31:
			// Reserved/cause bits stay writable for the kernel's FP context
			// save/restore path. Cause.E is read-only-zero in hardware but
			// ignoring that is harmless.
			c.FCSR = c.ReadRegister(inst.Rt) & 0xFFF8FFFF
		}
		c.retire()
	case COP1_BC:
		c.executeCOP1Branch(inst)
	case COP1_FMT_S, COP1_FMT_D, COP1_FMT_W, COP1_FMT_L:
		if inst.Funct >= COP1_C_F {
			c.executeCOP1Compare(inst)
		} else {
			c.executeCOP1Arith(inst)
		}
	default:
		c.reservedInstruction(inst)
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

	// If the kernel just handled a rdhwr $29 RI trap, observe the TLS
	// value it wrote to the destination register and cache it so future
	// rdhwr $29 calls return instantly without trapping.
	if c.pendingTLSPC != 0 {
		// Only cache if eret returns to the rdhwr site (or within
		// a few instructions of it, accounting for delay slots).
		if c.PC >= c.pendingTLSPC && c.PC <= c.pendingTLSPC+8 {
			tlsValue := c.Regs[c.pendingTLSRt]
			if tlsValue != 0 {
				c.UserLocal = tlsValue
				c.cachedTLSASID = c.CP0[CP0_ENTRYHI] & 0xFF
			}
		}
		c.pendingTLSPC = 0
	}

	// ERET is not a branch and has no delay slot.
	c.branchTaken = false
	c.LLBit = false

	c.retire()
}

// ---------------------------------------------------------------------
// Floating-point unit (COP1)
// ---------------------------------------------------------------------

// executeCOP1Branch implements BC1F/BC1T and their likely variants. The
// cc field selects which FCC bit to test, nd selects the "likely" form
// (nullify delay slot when not taken), and tf selects true vs false.
func (c *CPU) executeCOP1Branch(inst Instruction) {
	tf := inst.Rt&1 == 1
	nd := inst.Rt&2 != 0
	cc := uint8(inst.Rt>>2) & 7

	cond := c.readFCC(cc) == tf

	if nd {
		c.conditionalBranchLikely(cond, inst)
	} else {
		c.conditionalBranch(cond, inst)
	}
}

// executeCOP1Arith handles a fmt-formatted arithmetic, conversion, or
// conditional-move instruction. Only S/D source formats and W target for
// fixed-point conversion are supported; L format is reserved.
func (c *CPU) executeCOP1Arith(inst Instruction) {
	ft := inst.Rt
	fs := inst.Rd
	fd := inst.Shamt
	fmt := inst.Rs

	switch inst.Funct {
	case COP1_ADD:
		switch fmt {
		case COP1_FMT_S:
			c.writeFPR_S(fd, c.readFPR_S(fs)+c.readFPR_S(ft))
		case COP1_FMT_D:
			c.writeFPR_D(fd, c.readFPR_D(fs)+c.readFPR_D(ft))
		default:
			c.reservedInstruction(inst)
			return
		}
	case COP1_SUB:
		switch fmt {
		case COP1_FMT_S:
			c.writeFPR_S(fd, c.readFPR_S(fs)-c.readFPR_S(ft))
		case COP1_FMT_D:
			c.writeFPR_D(fd, c.readFPR_D(fs)-c.readFPR_D(ft))
		default:
			c.reservedInstruction(inst)
			return
		}
	case COP1_MUL:
		switch fmt {
		case COP1_FMT_S:
			c.writeFPR_S(fd, c.readFPR_S(fs)*c.readFPR_S(ft))
		case COP1_FMT_D:
			c.writeFPR_D(fd, c.readFPR_D(fs)*c.readFPR_D(ft))
		default:
			c.reservedInstruction(inst)
			return
		}
	case COP1_DIV:
		switch fmt {
		case COP1_FMT_S:
			c.writeFPR_S(fd, c.readFPR_S(fs)/c.readFPR_S(ft))
		case COP1_FMT_D:
			c.writeFPR_D(fd, c.readFPR_D(fs)/c.readFPR_D(ft))
		default:
			c.reservedInstruction(inst)
			return
		}
	case COP1_SQRT:
		switch fmt {
		case COP1_FMT_S:
			c.writeFPR_S(fd, float32(math.Sqrt(float64(c.readFPR_S(fs)))))
		case COP1_FMT_D:
			c.writeFPR_D(fd, math.Sqrt(c.readFPR_D(fs)))
		default:
			c.reservedInstruction(inst)
			return
		}
	case COP1_ABS:
		switch fmt {
		case COP1_FMT_S:
			c.writeFPR_S(fd, float32(math.Abs(float64(c.readFPR_S(fs)))))
		case COP1_FMT_D:
			c.writeFPR_D(fd, math.Abs(c.readFPR_D(fs)))
		default:
			c.reservedInstruction(inst)
			return
		}
	case COP1_MOV:
		switch fmt {
		case COP1_FMT_S:
			c.writeFPR_S(fd, c.readFPR_S(fs))
		case COP1_FMT_D:
			c.writeFPR_D(fd, c.readFPR_D(fs))
		case COP1_FMT_W:
			c.writeFPR_W(fd, c.readFPR_W(fs))
		default:
			c.reservedInstruction(inst)
			return
		}
	case COP1_NEG:
		switch fmt {
		case COP1_FMT_S:
			c.writeFPR_S(fd, -c.readFPR_S(fs))
		case COP1_FMT_D:
			c.writeFPR_D(fd, -c.readFPR_D(fs))
		default:
			c.reservedInstruction(inst)
			return
		}
	case COP1_ROUND_W:
		switch fmt {
		case COP1_FMT_S:
			c.writeFPR_W(fd, int32(math.Round(float64(c.readFPR_S(fs)))))
		case COP1_FMT_D:
			c.writeFPR_W(fd, int32(math.Round(c.readFPR_D(fs))))
		default:
			c.reservedInstruction(inst)
			return
		}
	case COP1_TRUNC_W:
		switch fmt {
		case COP1_FMT_S:
			c.writeFPR_W(fd, int32(math.Trunc(float64(c.readFPR_S(fs)))))
		case COP1_FMT_D:
			c.writeFPR_W(fd, int32(math.Trunc(c.readFPR_D(fs))))
		default:
			c.reservedInstruction(inst)
			return
		}
	case COP1_CEIL_W:
		switch fmt {
		case COP1_FMT_S:
			c.writeFPR_W(fd, int32(math.Ceil(float64(c.readFPR_S(fs)))))
		case COP1_FMT_D:
			c.writeFPR_W(fd, int32(math.Ceil(c.readFPR_D(fs))))
		default:
			c.reservedInstruction(inst)
			return
		}
	case COP1_FLOOR_W:
		switch fmt {
		case COP1_FMT_S:
			c.writeFPR_W(fd, int32(math.Floor(float64(c.readFPR_S(fs)))))
		case COP1_FMT_D:
			c.writeFPR_W(fd, int32(math.Floor(c.readFPR_D(fs))))
		default:
			c.reservedInstruction(inst)
			return
		}
	case COP1_CVT_S:
		switch fmt {
		case COP1_FMT_D:
			c.writeFPR_S(fd, float32(c.readFPR_D(fs)))
		case COP1_FMT_W:
			c.writeFPR_S(fd, float32(c.readFPR_W(fs)))
		default:
			c.reservedInstruction(inst)
			return
		}
	case COP1_CVT_D:
		switch fmt {
		case COP1_FMT_S:
			c.writeFPR_D(fd, float64(c.readFPR_S(fs)))
		case COP1_FMT_W:
			c.writeFPR_D(fd, float64(c.readFPR_W(fs)))
		default:
			c.reservedInstruction(inst)
			return
		}
	case COP1_CVT_W:
		switch fmt {
		case COP1_FMT_S:
			c.writeFPR_W(fd, c.roundWithMode(float64(c.readFPR_S(fs))))
		case COP1_FMT_D:
			c.writeFPR_W(fd, c.roundWithMode(c.readFPR_D(fs)))
		default:
			c.reservedInstruction(inst)
			return
		}
	case COP1_MOVCF:
		// movt/movf: tf lives in bit 16 (Rt[0]). cc lives in bits 20:18
		// (Rt[4:2]). Bit 17 (Rt[1]) is always 0.
		tf := inst.Rt&1 == 1
		cc := uint8(inst.Rt>>2) & 7
		if c.readFCC(cc) == tf {
			switch fmt {
			case COP1_FMT_S:
				c.writeFPR_S(fd, c.readFPR_S(fs))
			case COP1_FMT_D:
				c.writeFPR_D(fd, c.readFPR_D(fs))
			default:
				c.reservedInstruction(inst)
				return
			}
		}
	case COP1_MOVZ:
		if c.ReadRegister(inst.Rt) == 0 {
			switch fmt {
			case COP1_FMT_S:
				c.writeFPR_S(fd, c.readFPR_S(fs))
			case COP1_FMT_D:
				c.writeFPR_D(fd, c.readFPR_D(fs))
			default:
				c.reservedInstruction(inst)
				return
			}
		}
	case COP1_MOVN:
		if c.ReadRegister(inst.Rt) != 0 {
			switch fmt {
			case COP1_FMT_S:
				c.writeFPR_S(fd, c.readFPR_S(fs))
			case COP1_FMT_D:
				c.writeFPR_D(fd, c.readFPR_D(fs))
			default:
				c.reservedInstruction(inst)
				return
			}
		}
	default:
		c.reservedInstruction(inst)
		return
	}
	c.retire()
}

// executeCOP1Compare implements C.cond.fmt. The 16 MIPS predicates are
// the truth-table expansion of (less, equal, unordered) raised to the
// IEEE-754 predicate table. Signaling predicates additionally set the
// Invalid cause when either operand is NaN.
func (c *CPU) executeCOP1Compare(inst Instruction) {
	// C.cond.fmt encoding: ft in bits [20:16] (Rt), fs in bits [15:11]
	// (Rd), cc in bits [10:8] (Shamt >> 2). This is the same ft/fs layout
	// as arithmetic instructions; cc replaces the fd field.
	fs := inst.Rd
	ft := inst.Rt
	cc := uint8(inst.Shamt>>2) & 7

	var aD, bD float64
	var aS, bS float32
	switch inst.Rs {
	case COP1_FMT_S:
		aS = c.readFPR_S(fs)
		bS = c.readFPR_S(ft)
	case COP1_FMT_D:
		aD = c.readFPR_D(fs)
		bD = c.readFPR_D(ft)
	default:
		c.reservedInstruction(inst)
		return
	}

	var unordered, less, equal bool
	switch inst.Rs {
	case COP1_FMT_S:
		if math32IsNaN(aS) || math32IsNaN(bS) {
			unordered = true
		} else {
			less = aS < bS
			equal = aS == bS
		}
	case COP1_FMT_D:
		if math.IsNaN(aD) || math.IsNaN(bD) {
			unordered = true
		} else {
			less = aD < bD
			equal = aD == bD
		}
	}

	var result bool
	signaling := false
	switch inst.Funct {
	case COP1_C_F:
		result = false
	case COP1_C_UN:
		result = unordered
	case COP1_C_EQ:
		result = equal
	case COP1_C_UEQ:
		result = equal || unordered
	case COP1_C_OLT:
		result = less
	case COP1_C_ULT:
		result = less || unordered
	case COP1_C_OLE:
		result = less || equal
	case COP1_C_ULE:
		result = less || equal || unordered
	case COP1_C_SF:
		result, signaling = false, true
	case COP1_C_NGLE:
		result, signaling = unordered, true
	case COP1_C_SEQ:
		result, signaling = equal, true
	case COP1_C_NGL:
		result, signaling = equal || unordered, true
	case COP1_C_LT:
		result, signaling = less, true
	case COP1_C_NGE:
		result, signaling = less || unordered, true
	case COP1_C_LE:
		result, signaling = less || equal, true
	case COP1_C_NGT:
		result, signaling = less || equal || unordered, true
	default:
		c.reservedInstruction(inst)
		return
	}

	if signaling && unordered {
		c.FCSR |= FCSR_CAUSE_V | (FCSR_CAUSE_V >> 5) // cause + flag
	}

	c.setFCC(cc, result)
	c.retire()
}

// math32IsNaN reports whether v is NaN without dragging the runtime's
// reflection-based isNaN path through float64 conversion.
func math32IsNaN(v float32) bool {
	return v != v
}
