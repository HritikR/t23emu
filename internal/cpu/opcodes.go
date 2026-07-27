package cpu

const (

	// SPECIAL opcode for R-type instructions
	OP_SPECIAL uint8 = 0

	// Jumps and Branches
	OP_J   uint8 = 2
	OP_JAL uint8 = 3
	OP_BEQ uint8 = 4
	OP_BNE uint8 = 5

	// Arithmetic/Logical immediate
	OP_ADDI  uint8 = 8
	OP_ADDIU uint8 = 9
	OP_ANDI  uint8 = 12
	OP_ORI   uint8 = 13
	OP_LUI   uint8 = 15

	// Coprocessor 0
	OP_COP0 uint8 = 16

	// Memory
	OP_LW uint8 = 35
	OP_SW uint8 = 43
)

// Coprocessor 0 register indices
const (
	CP0_BADVADDR uint8 = 8
	CP0_COUNT    uint8 = 9
	CP0_COMPARE  uint8 = 11
	CP0_STATUS   uint8 = 12
	CP0_CAUSE    uint8 = 13
	CP0_EPC      uint8 = 14
	CP0_PRID     uint8 = 15
)

const (

	// R-type function codes

	FUNCT_SLL uint8 = 0
	FUNCT_SRL uint8 = 2
	FUNCT_SRA uint8 = 3

	FUNCT_JR uint8 = 8

	FUNCT_ADD uint8 = 32

	FUNCT_AND uint8 = 36
	FUNCT_OR  uint8 = 37
	FUNCT_XOR uint8 = 38
	FUNCT_NOR uint8 = 39

	FUNCT_SLT uint8 = 42

	// System/exception functions
	FUNCT_SYSCALL uint8 = 12
	FUNCT_BREAK   uint8 = 13
)

// Exception Codes
const (
	EXC_INT   uint8 = 0
	EXC_ADEL  uint8 = 4
	EXC_ADES  uint8 = 5
	EXC_SYS   uint8 = 8
	EXC_BP    uint8 = 9
	EXC_RI    uint8 = 10
	EXC_OV    uint8 = 12
)
