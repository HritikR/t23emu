package cpu

const (

	// SPECIAL opcode for R-type instructions
	OP_SPECIAL uint8 = 0

	// ADD immediate
	OP_ADDI uint8 = 8

	// Memory
	OP_LW uint8 = 35
	OP_SW uint8 = 43
)

const (

	// R-type function codes

	FUNCT_SLL uint8 = 0

	FUNCT_ADD uint8 = 32
)
