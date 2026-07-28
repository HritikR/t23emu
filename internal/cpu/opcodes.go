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
	OP_PREF  uint8 = 51
	OP_SC    uint8 = 56
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
	CONFIG_K0 uint32 = 3

	// Config1 describes a small MIPS32/XBurst-class core with 32 TLB
	// entries, separate 16 KiB I/D caches, no FPU, and no Config2.
	CP0_CONFIG1_RESET uint32 = (31 << 25) | (3 << 22) | (4 << 19) | (1 << 16) |
		(3 << 13) | (4 << 10) | (1 << 7)
)

// CP0 Status register bits
const (
	STATUS_IE  uint32 = 1 << 0  // Interrupt enable
	STATUS_EXL uint32 = 1 << 1  // Exception level
	STATUS_ERL uint32 = 1 << 2  // Error level
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
