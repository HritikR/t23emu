package dynarec

// MIPS Opcodes
const (
	OP_SPECIAL  = 0x00
	OP_REGIMM   = 0x01
	OP_J        = 0x02
	OP_JAL      = 0x03
	OP_BEQ      = 0x04
	OP_BNE      = 0x05
	OP_BLEZ     = 0x06
	OP_BGTZ     = 0x07
	OP_ADDI     = 0x08
	OP_ADDIU    = 0x09
	OP_SLTI     = 0x0A
	OP_SLTIU    = 0x0B
	OP_ANDI     = 0x0C
	OP_ORI      = 0x0D
	OP_XORI     = 0x0E
	OP_LUI      = 0x0F
	OP_BEQL     = 0x14
	OP_BNEL     = 0x15
	OP_BLEZL    = 0x16
	OP_BGTZL    = 0x17
	OP_LB       = 0x20
	OP_LH       = 0x21
	OP_LWL      = 0x22
	OP_LW       = 0x23
	OP_LBU      = 0x24
	OP_LHU      = 0x25
	OP_LWR      = 0x26
	OP_SB       = 0x28
	OP_SH       = 0x29
	OP_SWL      = 0x2A
	OP_SW       = 0x2B
	OP_SWR      = 0x2E
	OP_CACHE    = 0x2F
	OP_LL       = 0x30
	OP_LWC1     = 0x31
	OP_PREF     = 0x33
	OP_LDC1     = 0x35
	OP_SC       = 0x38
	OP_SWC1     = 0x39
	OP_SDC1     = 0x3D
)

// SPECIAL Funct codes
const (
	FUNCT_SLL     = 0x00
	FUNCT_SRL     = 0x02
	FUNCT_SRA     = 0x03
	FUNCT_SLLV    = 0x04
	FUNCT_SRLV    = 0x06
	FUNCT_SRAV    = 0x07
	FUNCT_JR      = 0x08
	FUNCT_JALR    = 0x09
	FUNCT_MOVZ    = 0x0A
	FUNCT_MOVN    = 0x0B
	FUNCT_SYSCALL = 0x0C
	FUNCT_BREAK   = 0x0D
	FUNCT_SYNC    = 0x0F
	FUNCT_MFHI    = 0x10
	FUNCT_MTHI    = 0x11
	FUNCT_MFLO    = 0x12
	FUNCT_MTLO    = 0x13
	FUNCT_MULT    = 0x18
	FUNCT_MULTU   = 0x19
	FUNCT_DIV     = 0x1A
	FUNCT_DIVU    = 0x1B
	FUNCT_ADD     = 0x20
	FUNCT_ADDU    = 0x21
	FUNCT_SUB     = 0x22
	FUNCT_SUBU    = 0x23
	FUNCT_AND     = 0x24
	FUNCT_OR      = 0x25
	FUNCT_XOR     = 0x26
	FUNCT_NOR     = 0x27
	FUNCT_SLT     = 0x2B
	FUNCT_SLTU    = 0x2C
)

// REGIMM Rt codes
const (
	REGIMM_BLTZ   = 0x00
	REGIMM_BGEZ   = 0x01
	REGIMM_BLTZL  = 0x02
	REGIMM_BGEZL  = 0x03
	REGIMM_BLTZAL = 0x10
	REGIMM_BGEZAL = 0x11
)

// Exceptions
const (
	EXC_OV   = 12
	EXC_ADEL = 4
	EXC_ADES = 5
)

// Instruction fields
type Instruction struct {
	Raw       uint32
	Opcode    uint8
	Rs        uint8
	Rt        uint8
	Rd        uint8
	Shamt     uint8
	Funct     uint8
	Immediate uint16
	Target    uint32
}

func DecodeInst(raw uint32) Instruction {
	return Instruction{
		Raw:       raw,
		Opcode:    uint8(raw >> 26),
		Rs:        uint8((raw >> 21) & 0x1F),
		Rt:        uint8((raw >> 16) & 0x1F),
		Rd:        uint8((raw >> 11) & 0x1F),
		Shamt:     uint8((raw >> 6) & 0x1F),
		Funct:     uint8(raw & 0x3F),
		Immediate: uint16(raw & 0xFFFF),
		Target:    raw & 0x03FFFFFF,
	}
}

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
	}
	return false
}
