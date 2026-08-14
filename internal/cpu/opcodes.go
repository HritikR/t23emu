package cpu

const (

	// SPECIAL opcode for R-type instructions
	OP_SPECIAL uint8 = 0

	// REGIMM opcode: sub-decoded by the Rt field
	OP_REGIMM uint8 = 1

	// Jumps and Branches
	OP_J    uint8 = 2
	OP_JAL  uint8 = 3
	OP_BEQ  uint8 = 4
	OP_BNE  uint8 = 5
	OP_BLEZ uint8 = 6
	OP_BGTZ uint8 = 7

	// Arithmetic/Logical immediate
	OP_ADDI  uint8 = 8
	OP_ADDIU uint8 = 9
	OP_SLTI  uint8 = 10
	OP_SLTIU uint8 = 11
	OP_ANDI  uint8 = 12
	OP_ORI   uint8 = 13
	OP_XORI  uint8 = 14
	OP_LUI   uint8 = 15

	// Coprocessors
	OP_COP0 uint8 = 16
	OP_COP1 uint8 = 17
	OP_COP2 uint8 = 18

	// Branch likely variants (nullify the delay slot when not taken)
	OP_BEQL  uint8 = 20
	OP_BNEL  uint8 = 21
	OP_BLEZL uint8 = 22
	OP_BGTZL uint8 = 23

	// SPECIAL2 opcode: sub-decoded by the Funct field
	OP_SPECIAL2 uint8 = 28

	// SPECIAL3 opcode: sub-decoded by the Funct field
	OP_SPECIAL3 uint8 = 31

	// Memory
	OP_LB    uint8 = 32
	OP_LH    uint8 = 33
	OP_LWL   uint8 = 34
	OP_LW    uint8 = 35
	OP_LBU   uint8 = 36
	OP_LHU   uint8 = 37
	OP_LWR   uint8 = 38
	OP_SB    uint8 = 40
	OP_SH    uint8 = 41
	OP_SWL   uint8 = 42
	OP_SW    uint8 = 43
	OP_SWR   uint8 = 46
	OP_CACHE uint8 = 47
	OP_LL    uint8 = 48
	OP_LWC1  uint8 = 49
	OP_PREF  uint8 = 51
	OP_LDC1  uint8 = 53
	OP_SC    uint8 = 56
	OP_SWC1  uint8 = 57
	OP_SDC1  uint8 = 61
)

// REGIMM sub-opcodes, held in the Rt field.
const (
	REGIMM_BLTZ   uint8 = 0
	REGIMM_BGEZ   uint8 = 1
	REGIMM_BLTZL  uint8 = 2
	REGIMM_BGEZL  uint8 = 3
	REGIMM_BLTZAL uint8 = 16
	REGIMM_BGEZAL uint8 = 17
)

// Coprocessor 0 register indices
const (
	CP0_INDEX    uint8 = 0
	CP0_RANDOM   uint8 = 1
	CP0_ENTRYLO0 uint8 = 2
	CP0_ENTRYLO1 uint8 = 3
	CP0_CONTEXT  uint8 = 4
	CP0_PAGEMASK uint8 = 5
	CP0_WIRED    uint8 = 6
	CP0_HWRENA   uint8 = 7
	CP0_BADVADDR uint8 = 8
	CP0_COUNT    uint8 = 9
	CP0_ENTRYHI  uint8 = 10
	CP0_COMPARE  uint8 = 11
	CP0_STATUS   uint8 = 12
	CP0_CAUSE    uint8 = 13
	CP0_EPC      uint8 = 14
	CP0_PRID     uint8 = 15
	CP0_CONFIG   uint8 = 16
	CP0_LLADDR   uint8 = 17
	CP0_ERROREPC uint8 = 30
)

// CP0 Config register bits and reset values used by the emulated T23 core.
const (
	CONFIG_M  uint32 = 1 << 31
	CONFIG_AR uint32 = 1 << 10
	CONFIG_K0 uint32 = 3

	// Config1 describes a small MIPS32/XBurst-class core with 32 TLB
	// entries, separate 16 KiB I/D caches, an FPU, and Config2 present.
	CP0_CONFIG1_RESET uint32 = CONFIG_M | (31 << 25) | (3 << 22) | (4 << 19) | (1 << 16) |
		(3 << 13) | (4 << 10) | (1 << 7) | CONFIG1_FP
	CP0_CONFIG2_RESET uint32 = CONFIG_M
	CP0_CONFIG3_RESET uint32 = 1 << 13
)

// Config1 feature bits.
const (
	CONFIG1_FP uint32 = 1 << 3 // FPU implemented
)

// CP0 Status register bits
const (
	STATUS_IE  uint32 = 1 << 0  // Interrupt enable
	STATUS_EXL uint32 = 1 << 1  // Exception level
	STATUS_ERL uint32 = 1 << 2  // Error level
	STATUS_KSU uint32 = 0x18    // Mode bits 4:3 (0=kernel, 1=sup, 2=user)
	STATUS_CU0 uint32 = 1 << 28 // Coprocessor 0 usable in user mode
	STATUS_CU1 uint32 = 1 << 29 // Coprocessor 1 (FPU) usable in user mode
	STATUS_BEV uint32 = 1 << 22 // Bootstrap exception vectors
	STATUS_IM  uint32 = 0x0000FF00
)

// CP0 Cause register bits
const (
	CAUSE_EXCCODE uint32 = 0x7C       // ExcCode field, bits 6:2
	CAUSE_BD      uint32 = 1 << 31    // Exception occurred in a branch delay slot
	CAUSE_CE      uint32 = 0x30000000 // Coprocessor number for a CpU exception
	CAUSE_IP      uint32 = 0x0000FF00 // Interrupt pending bits, bits 15:8
	CAUSE_IP2     uint32 = 1 << 10    // Hardware interrupt line 0
	CAUSE_IP7     uint32 = 1 << 15    // CP0 timer interrupt
)

const (

	// R-type function codes (SPECIAL)

	FUNCT_SLL   uint8 = 0
	FUNCT_MOVCI uint8 = 1
	FUNCT_SRL   uint8 = 2
	FUNCT_SRA   uint8 = 3
	FUNCT_SLLV  uint8 = 4
	FUNCT_SRLV  uint8 = 6
	FUNCT_SRAV  uint8 = 7

	FUNCT_JR   uint8 = 8
	FUNCT_JALR uint8 = 9

	FUNCT_MOVZ uint8 = 10
	FUNCT_MOVN uint8 = 11

	FUNCT_SYNC uint8 = 15

	FUNCT_MFHI uint8 = 16
	FUNCT_MTHI uint8 = 17
	FUNCT_MFLO uint8 = 18
	FUNCT_MTLO uint8 = 19

	FUNCT_MULT  uint8 = 24
	FUNCT_MULTU uint8 = 25
	FUNCT_DIV   uint8 = 26
	FUNCT_DIVU  uint8 = 27

	FUNCT_ADD  uint8 = 32
	FUNCT_ADDU uint8 = 33
	FUNCT_SUB  uint8 = 34
	FUNCT_SUBU uint8 = 35

	FUNCT_AND uint8 = 36
	FUNCT_OR  uint8 = 37
	FUNCT_XOR uint8 = 38
	FUNCT_NOR uint8 = 39

	FUNCT_SLT  uint8 = 42
	FUNCT_SLTU uint8 = 43

	FUNCT_TGE  uint8 = 48
	FUNCT_TGEU uint8 = 49
	FUNCT_TLT  uint8 = 50
	FUNCT_TLTU uint8 = 51
	FUNCT_TEQ  uint8 = 52
	FUNCT_TNE  uint8 = 54

	// System/exception functions
	FUNCT_SYSCALL uint8 = 12
	FUNCT_BREAK   uint8 = 13
)

// SPECIAL2 function codes
const (
	FUNCT2_MADD  uint8 = 0
	FUNCT2_MADDU uint8 = 1
	FUNCT2_MUL   uint8 = 2
	FUNCT2_MSUB  uint8 = 4
	FUNCT2_MSUBU uint8 = 5
	FUNCT2_CLZ   uint8 = 32
	FUNCT2_CLO   uint8 = 33
	FUNCT2_SDBBP uint8 = 63
)

// SPECIAL3 function codes
const (
	FUNCT3_EXT   uint8 = 0
	FUNCT3_INS   uint8 = 4
	FUNCT3_BSHFL uint8 = 32
	FUNCT3_RDHWR uint8 = 59
)

// BSHFL sub-operations, held in the Shamt field of a SPECIAL3 BSHFL instruction.
const (
	BSHFL_WSBH uint8 = 2
	BSHFL_SEB  uint8 = 16
	BSHFL_SEH  uint8 = 24
)

// COP0 sub-operations, held in the Rs field.
const (
	COP0_MFC0 uint8 = 0
	COP0_MTC0 uint8 = 4
	COP0_CO   uint8 = 16 // Coprocessor operation: TLB ops, ERET, WAIT, DERET
)

// COP1 register transfer sub-operations, held in the Rs field.
const (
	COP1_MFC1 uint8 = 0
	COP1_CFC1 uint8 = 2
	COP1_MTC1 uint8 = 4
	COP1_CTC1 uint8 = 6
)

// COP1 format codes (rs field of a fmt-formatted COP1 instruction).
const (
	COP1_FMT_S uint8 = 16 // single (32-bit float)
	COP1_FMT_D uint8 = 17 // double (64-bit float)
	COP1_FMT_W uint8 = 20 // word (32-bit signed int)
	COP1_FMT_L uint8 = 21 // longword (64-bit signed int, MIPS32r2)
	COP1_BC    uint8 = 8  // branch on FP condition
)

// COP1 arithmetic function codes. Held in the funct field of a
// fmt-formatted instruction.
const (
	COP1_ADD     uint8 = 0  // add.fmt
	COP1_SUB     uint8 = 1  // sub.fmt
	COP1_MUL     uint8 = 2  // mul.fmt
	COP1_DIV     uint8 = 3  // div.fmt
	COP1_SQRT    uint8 = 4  // sqrt.fmt
	COP1_ABS     uint8 = 5  // abs.fmt
	COP1_MOV     uint8 = 6  // mov.fmt
	COP1_NEG     uint8 = 7  // neg.fmt
	COP1_ROUND_W uint8 = 12 // round.w.fmt (round to nearest, away on tie)
	COP1_TRUNC_W uint8 = 13 // trunc.w.fmt (round toward zero)
	COP1_CEIL_W  uint8 = 14 // ceil.w.fmt (round toward +inf)
	COP1_FLOOR_W uint8 = 15 // floor.w.fmt (round toward -inf)
	COP1_MOVCF   uint8 = 17 // movcf.fmt (move on FP condition)
	COP1_MOVZ    uint8 = 18 // movz.fmt
	COP1_MOVN    uint8 = 19 // movn.fmt
	COP1_CVT_S   uint8 = 32 // cvt.s.fmt
	COP1_CVT_D   uint8 = 33 // cvt.d.fmt
	COP1_CVT_W   uint8 = 36 // cvt.w.fmt (uses FCSR.RM)
)

// COP1 compare condition codes (funct field). C.cond.fmt encodes a
// 16-entry predicate table combining less/equal/unordered results.
const (
	COP1_C_F    uint8 = 48
	COP1_C_UN   uint8 = 49
	COP1_C_EQ   uint8 = 50
	COP1_C_UEQ  uint8 = 51
	COP1_C_OLT  uint8 = 52
	COP1_C_ULT  uint8 = 53
	COP1_C_OLE  uint8 = 54
	COP1_C_ULE  uint8 = 55
	COP1_C_SF   uint8 = 56
	COP1_C_NGLE uint8 = 57
	COP1_C_SEQ  uint8 = 58
	COP1_C_NGL  uint8 = 59
	COP1_C_LT   uint8 = 60
	COP1_C_NGE  uint8 = 61
	COP1_C_LE   uint8 = 62
	COP1_C_NGT  uint8 = 63
)

// FCSR bit fields and rounding modes.
//
// FCC0 lives at bit 23; FCC1..FCC7 occupy bits 24..30 respectively. The
// cc field of a C.cond or BC1 instruction therefore maps directly to
// bit position 23+cc.
const (
	FCSR_FCC0   uint32 = 1 << 23
	FCSR_ALLFCC uint32 = 0x7F800000 // FCC7..FCC0, bits 30..23
	FCSR_CAUSE  uint32 = 0x0003F000 // E/V/Z/O/U/I sticky causes, bits 17:12
	FCSR_ENABLE uint32 = 0x00000F80 // V/Z/O/U/I trap enables, bits 11:7
	FCSR_FLAGS  uint32 = 0x0000007C // V/Z/O/U/I sticky flags, bits 6:2
	FCSR_RMMASK uint32 = 0x00000003 // rounding mode, bits 1:0

	FCSR_CAUSE_E uint32 = 1 << 17 // Unimplemented operation
	FCSR_CAUSE_V uint32 = 1 << 16 // Invalid operation
	FCSR_CAUSE_Z uint32 = 1 << 15 // Division by zero
	FCSR_CAUSE_O uint32 = 1 << 14 // Overflow
	FCSR_CAUSE_U uint32 = 1 << 13 // Underflow
	FCSR_CAUSE_I uint32 = 1 << 12 // Inexact

	FP_RN uint32 = 0 // Round to nearest, ties to even
	FP_RZ uint32 = 1 // Round toward zero
	FP_RP uint32 = 2 // Round toward +infinity
	FP_RM uint32 = 3 // Round toward -infinity
)

// COP0 "CO" function codes (valid when Rs == COP0_CO).
const (
	COP0CO_TLBR  uint8 = 1
	COP0CO_TLBWI uint8 = 2
	COP0CO_TLBWR uint8 = 6
	COP0CO_TLBP  uint8 = 8
	COP0CO_ERET  uint8 = 24
	COP0CO_DERET uint8 = 31
	COP0CO_WAIT  uint8 = 32
)

// Exception Codes
const (
	EXC_INT  uint8 = 0
	EXC_MOD  uint8 = 1
	EXC_TLBL uint8 = 2
	EXC_TLBS uint8 = 3
	EXC_ADEL uint8 = 4
	EXC_ADES uint8 = 5
	EXC_IBE  uint8 = 6
	EXC_DBE  uint8 = 7
	EXC_SYS  uint8 = 8
	EXC_BP   uint8 = 9
	EXC_RI   uint8 = 10
	EXC_CPU  uint8 = 11
	EXC_OV   uint8 = 12
	EXC_TR   uint8 = 13
)
