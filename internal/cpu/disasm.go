package cpu

import "fmt"

// RegNames holds the conventional MIPS assembler names for the 32
// general purpose registers, indexed by register number.
var RegNames = [32]string{
	"zero", "at", "v0", "v1", "a0", "a1", "a2", "a3",
	"t0", "t1", "t2", "t3", "t4", "t5", "t6", "t7",
	"s0", "s1", "s2", "s3", "s4", "s5", "s6", "s7",
	"t8", "t9", "k0", "k1", "gp", "sp", "fp", "ra",
}

// cp0Names holds the conventional names of the CP0 registers, used to
// annotate mfc0/mtc0 in disassembly.
var cp0Names = [32]string{
	"Index", "Random", "EntryLo0", "EntryLo1",
	"Context", "PageMask", "Wired", "HWREna",
	"BadVAddr", "Count", "EntryHi", "Compare",
	"Status", "Cause", "EPC", "PRId",
	"Config", "LLAddr", "WatchLo", "WatchHi",
	"XContext", "cp0r21", "cp0r22", "Debug",
	"DEPC", "PerfCnt", "ErrCtl", "CacheErr",
	"TagLo", "TagHi", "ErrorEPC", "DESAVE",
}

func reg(index uint8) string {
	return "$" + RegNames[index&0x1F]
}

// simm formats a sign-extended 16-bit immediate the way an assembler
// listing would, so that negative offsets read as negative.
func simm(value uint16) string {
	v := int16(value)
	if v < 0 {
		return fmt.Sprintf("-0x%x", -int32(v))
	}
	return fmt.Sprintf("0x%x", v)
}

// Disassemble renders a single instruction as assembly text.
//
// pc is the address the instruction was fetched from; it is needed to
// resolve the absolute targets of PC-relative branches and of the
// region-relative J/JAL forms.
func Disassemble(raw uint32, pc uint32) string {
	inst := Decode(raw)

	if raw == 0 {
		return "nop"
	}

	// Branch targets are relative to the delay slot, i.e. pc+4.
	branchTarget := uint32(int32(pc+4) + (int32(int16(inst.Immediate)) << 2))

	// J/JAL replace the low 28 bits of the delay slot's address.
	jumpTarget := ((pc + 4) & 0xF0000000) | (inst.Target << 2)

	switch inst.Opcode {

	case OP_SPECIAL:
		return disasmSpecial(inst)

	case OP_REGIMM:
		switch inst.Rt {
		case REGIMM_BLTZ:
			return fmt.Sprintf("bltz    %s, 0x%08x", reg(inst.Rs), branchTarget)
		case REGIMM_BGEZ:
			return fmt.Sprintf("bgez    %s, 0x%08x", reg(inst.Rs), branchTarget)
		case REGIMM_BLTZL:
			return fmt.Sprintf("bltzl   %s, 0x%08x", reg(inst.Rs), branchTarget)
		case REGIMM_BGEZL:
			return fmt.Sprintf("bgezl   %s, 0x%08x", reg(inst.Rs), branchTarget)
		case REGIMM_BLTZAL:
			return fmt.Sprintf("bltzal  %s, 0x%08x", reg(inst.Rs), branchTarget)
		case REGIMM_BGEZAL:
			return fmt.Sprintf("bgezal  %s, 0x%08x", reg(inst.Rs), branchTarget)
		}

	case OP_J:
		return fmt.Sprintf("j       0x%08x", jumpTarget)

	case OP_JAL:
		return fmt.Sprintf("jal     0x%08x", jumpTarget)

	case OP_BEQ:
		// beq $x, $zero is conventionally written beqz.
		if inst.Rt == 0 {
			return fmt.Sprintf("beqz    %s, 0x%08x", reg(inst.Rs), branchTarget)
		}
		if inst.Rs == 0 && inst.Rt == 0 {
			return fmt.Sprintf("b       0x%08x", branchTarget)
		}
		return fmt.Sprintf("beq     %s, %s, 0x%08x", reg(inst.Rs), reg(inst.Rt), branchTarget)

	case OP_BNE:
		if inst.Rt == 0 {
			return fmt.Sprintf("bnez    %s, 0x%08x", reg(inst.Rs), branchTarget)
		}
		return fmt.Sprintf("bne     %s, %s, 0x%08x", reg(inst.Rs), reg(inst.Rt), branchTarget)

	case OP_BLEZ:
		return fmt.Sprintf("blez    %s, 0x%08x", reg(inst.Rs), branchTarget)

	case OP_BGTZ:
		return fmt.Sprintf("bgtz    %s, 0x%08x", reg(inst.Rs), branchTarget)

	case OP_BEQL:
		return fmt.Sprintf("beql    %s, %s, 0x%08x", reg(inst.Rs), reg(inst.Rt), branchTarget)

	case OP_BNEL:
		return fmt.Sprintf("bnel    %s, %s, 0x%08x", reg(inst.Rs), reg(inst.Rt), branchTarget)

	case OP_BLEZL:
		return fmt.Sprintf("blezl   %s, 0x%08x", reg(inst.Rs), branchTarget)

	case OP_BGTZL:
		return fmt.Sprintf("bgtzl   %s, 0x%08x", reg(inst.Rs), branchTarget)

	case OP_ADDI:
		return fmt.Sprintf("addi    %s, %s, %s", reg(inst.Rt), reg(inst.Rs), simm(inst.Immediate))

	case OP_ADDIU:
		// addiu $x, $zero, imm is the canonical "li".
		if inst.Rs == 0 {
			return fmt.Sprintf("li      %s, %s", reg(inst.Rt), simm(inst.Immediate))
		}
		return fmt.Sprintf("addiu   %s, %s, %s", reg(inst.Rt), reg(inst.Rs), simm(inst.Immediate))

	case OP_SLTI:
		return fmt.Sprintf("slti    %s, %s, %s", reg(inst.Rt), reg(inst.Rs), simm(inst.Immediate))

	case OP_SLTIU:
		return fmt.Sprintf("sltiu   %s, %s, %s", reg(inst.Rt), reg(inst.Rs), simm(inst.Immediate))

	case OP_ANDI:
		return fmt.Sprintf("andi    %s, %s, 0x%x", reg(inst.Rt), reg(inst.Rs), inst.Immediate)

	case OP_ORI:
		return fmt.Sprintf("ori     %s, %s, 0x%x", reg(inst.Rt), reg(inst.Rs), inst.Immediate)

	case OP_XORI:
		return fmt.Sprintf("xori    %s, %s, 0x%x", reg(inst.Rt), reg(inst.Rs), inst.Immediate)

	case OP_LUI:
		return fmt.Sprintf("lui     %s, 0x%x", reg(inst.Rt), inst.Immediate)

	case OP_COP0:
		return disasmCOP0(inst)

	case OP_SPECIAL2:
		return disasmSpecial2(inst)

	case OP_SPECIAL3:
		return disasmSpecial3(inst)

	case OP_CACHE:
		return fmt.Sprintf("cache   0x%x, %s(%s)", inst.Rt, simm(inst.Immediate), reg(inst.Rs))

	case OP_PREF:
		return fmt.Sprintf("pref    0x%x, %s(%s)", inst.Rt, simm(inst.Immediate), reg(inst.Rs))
	}

	// The remaining opcodes are all load/store forms sharing one layout.
	if name, ok := memMnemonics[inst.Opcode]; ok {
		return fmt.Sprintf("%-7s %s, %s(%s)", name, reg(inst.Rt), simm(inst.Immediate), reg(inst.Rs))
	}

	return fmt.Sprintf(".word   0x%08x", raw)
}

var memMnemonics = map[uint8]string{
	OP_LB:  "lb",
	OP_LH:  "lh",
	OP_LWL: "lwl",
	OP_LW:  "lw",
	OP_LBU: "lbu",
	OP_LHU: "lhu",
	OP_LWR: "lwr",
	OP_SB:  "sb",
	OP_SH:  "sh",
	OP_SWL: "swl",
	OP_SW:  "sw",
	OP_SWR: "swr",
	OP_LL:  "ll",
	OP_SC:  "sc",
}

func disasmSpecial(inst Instruction) string {
	switch inst.Funct {

	case FUNCT_SLL:
		return fmt.Sprintf("sll     %s, %s, %d", reg(inst.Rd), reg(inst.Rt), inst.Shamt)
	case FUNCT_SRL:
		return fmt.Sprintf("srl     %s, %s, %d", reg(inst.Rd), reg(inst.Rt), inst.Shamt)
	case FUNCT_SRA:
		return fmt.Sprintf("sra     %s, %s, %d", reg(inst.Rd), reg(inst.Rt), inst.Shamt)
	case FUNCT_SLLV:
		return fmt.Sprintf("sllv    %s, %s, %s", reg(inst.Rd), reg(inst.Rt), reg(inst.Rs))
	case FUNCT_SRLV:
		return fmt.Sprintf("srlv    %s, %s, %s", reg(inst.Rd), reg(inst.Rt), reg(inst.Rs))
	case FUNCT_SRAV:
		return fmt.Sprintf("srav    %s, %s, %s", reg(inst.Rd), reg(inst.Rt), reg(inst.Rs))

	case FUNCT_JR:
		return fmt.Sprintf("jr      %s", reg(inst.Rs))
	case FUNCT_JALR:
		if inst.Rd == 31 {
			return fmt.Sprintf("jalr    %s", reg(inst.Rs))
		}
		return fmt.Sprintf("jalr    %s, %s", reg(inst.Rd), reg(inst.Rs))

	case FUNCT_MOVZ:
		return fmt.Sprintf("movz    %s, %s, %s", reg(inst.Rd), reg(inst.Rs), reg(inst.Rt))
	case FUNCT_MOVN:
		return fmt.Sprintf("movn    %s, %s, %s", reg(inst.Rd), reg(inst.Rs), reg(inst.Rt))

	case FUNCT_SYNC:
		return "sync"

	case FUNCT_MFHI:
		return fmt.Sprintf("mfhi    %s", reg(inst.Rd))
	case FUNCT_MTHI:
		return fmt.Sprintf("mthi    %s", reg(inst.Rs))
	case FUNCT_MFLO:
		return fmt.Sprintf("mflo    %s", reg(inst.Rd))
	case FUNCT_MTLO:
		return fmt.Sprintf("mtlo    %s", reg(inst.Rs))

	case FUNCT_MULT:
		return fmt.Sprintf("mult    %s, %s", reg(inst.Rs), reg(inst.Rt))
	case FUNCT_MULTU:
		return fmt.Sprintf("multu   %s, %s", reg(inst.Rs), reg(inst.Rt))
	case FUNCT_DIV:
		return fmt.Sprintf("div     %s, %s", reg(inst.Rs), reg(inst.Rt))
	case FUNCT_DIVU:
		return fmt.Sprintf("divu    %s, %s", reg(inst.Rs), reg(inst.Rt))

	case FUNCT_SYSCALL:
		return "syscall"
	case FUNCT_BREAK:
		return "break"
	}

	if name, ok := aluMnemonics[inst.Funct]; ok {
		// move $d, $s is the canonical form of "or $d, $s, $zero" and
		// "addu $d, $s, $zero", both of which compilers emit constantly.
		if inst.Rt == 0 && (inst.Funct == FUNCT_OR || inst.Funct == FUNCT_ADDU) {
			return fmt.Sprintf("move    %s, %s", reg(inst.Rd), reg(inst.Rs))
		}
		return fmt.Sprintf("%-7s %s, %s, %s", name, reg(inst.Rd), reg(inst.Rs), reg(inst.Rt))
	}

	return fmt.Sprintf(".word   0x%08x", inst.Raw)
}

var aluMnemonics = map[uint8]string{
	FUNCT_ADD:  "add",
	FUNCT_ADDU: "addu",
	FUNCT_SUB:  "sub",
	FUNCT_SUBU: "subu",
	FUNCT_AND:  "and",
	FUNCT_OR:   "or",
	FUNCT_XOR:  "xor",
	FUNCT_NOR:  "nor",
	FUNCT_SLT:  "slt",
	FUNCT_SLTU: "sltu",
}

func disasmSpecial2(inst Instruction) string {
	switch inst.Funct {
	case FUNCT2_MUL:
		return fmt.Sprintf("mul     %s, %s, %s", reg(inst.Rd), reg(inst.Rs), reg(inst.Rt))
	case FUNCT2_MADD:
		return fmt.Sprintf("madd    %s, %s", reg(inst.Rs), reg(inst.Rt))
	case FUNCT2_MADDU:
		return fmt.Sprintf("maddu   %s, %s", reg(inst.Rs), reg(inst.Rt))
	case FUNCT2_MSUB:
		return fmt.Sprintf("msub    %s, %s", reg(inst.Rs), reg(inst.Rt))
	case FUNCT2_MSUBU:
		return fmt.Sprintf("msubu   %s, %s", reg(inst.Rs), reg(inst.Rt))
	case FUNCT2_CLZ:
		return fmt.Sprintf("clz     %s, %s", reg(inst.Rd), reg(inst.Rs))
	case FUNCT2_CLO:
		return fmt.Sprintf("clo     %s, %s", reg(inst.Rd), reg(inst.Rs))
	}
	return fmt.Sprintf(".word   0x%08x", inst.Raw)
}

func disasmSpecial3(inst Instruction) string {
	switch inst.Funct {
	case FUNCT3_EXT:
		// pos is in Shamt, size-1 is in Rd.
		return fmt.Sprintf("ext     %s, %s, %d, %d",
			reg(inst.Rt), reg(inst.Rs), inst.Shamt, uint32(inst.Rd)+1)
	case FUNCT3_INS:
		// pos is in Shamt, msb is in Rd, so size is msb-pos+1.
		return fmt.Sprintf("ins     %s, %s, %d, %d",
			reg(inst.Rt), reg(inst.Rs), inst.Shamt, uint32(inst.Rd)-uint32(inst.Shamt)+1)
	case FUNCT3_BSHFL:
		switch inst.Shamt {
		case BSHFL_WSBH:
			return fmt.Sprintf("wsbh    %s, %s", reg(inst.Rd), reg(inst.Rt))
		case BSHFL_SEB:
			return fmt.Sprintf("seb     %s, %s", reg(inst.Rd), reg(inst.Rt))
		case BSHFL_SEH:
			return fmt.Sprintf("seh     %s, %s", reg(inst.Rd), reg(inst.Rt))
		}
	case FUNCT3_RDHWR:
		return fmt.Sprintf("rdhwr   %s, %d", reg(inst.Rt), inst.Rd)
	}
	return fmt.Sprintf(".word   0x%08x", inst.Raw)
}

func disasmCOP0(inst Instruction) string {
	switch inst.Rs {
	case COP0_MFC0:
		return fmt.Sprintf("mfc0    %s, %s", reg(inst.Rt), cp0Names[inst.Rd&0x1F])
	case COP0_MTC0:
		return fmt.Sprintf("mtc0    %s, %s", reg(inst.Rt), cp0Names[inst.Rd&0x1F])
	case COP0_CO:
		switch inst.Funct {
		case COP0CO_TLBR:
			return "tlbr"
		case COP0CO_TLBWI:
			return "tlbwi"
		case COP0CO_TLBWR:
			return "tlbwr"
		case COP0CO_TLBP:
			return "tlbp"
		case COP0CO_ERET:
			return "eret"
		case COP0CO_DERET:
			return "deret"
		case COP0CO_WAIT:
			return "wait"
		}
	}
	return fmt.Sprintf(".word   0x%08x", inst.Raw)
}
