package cpu

// Instruction represents a decoded MIPS instruction.
type Instruction struct {

	// Original 32-bit instruction
	Raw uint32

	// Opcode field [31:26]
	Opcode uint8

	// Source register [25:21]
	Rs uint8

	// Target register [20:16]
	Rt uint8

	// Destination register [15:11]
	Rd uint8

	// Shift amount [10:6]
	Shamt uint8

	// Function code [5:0]
	Funct uint8

	// Immediate value [15:0]
	Immediate uint16

	// Jump Target [25:0]
	Target uint32
}

// Decode converts a raw 32-bit MIPS instruction
// into its component fields.
func Decode(raw uint32) Instruction {

	return Instruction{

		Raw: raw,

		Opcode: uint8(
			(raw >> 26) & 0x3F,
		),

		Rs: uint8(
			(raw >> 21) & 0x1F,
		),

		Rt: uint8(
			(raw >> 16) & 0x1F,
		),

		Rd: uint8(
			(raw >> 11) & 0x1F,
		),

		Shamt: uint8(
			(raw >> 6) & 0x1F,
		),

		Funct: uint8(
			raw & 0x3F,
		),

		Immediate: uint16(
			raw & 0xFFFF,
		),

		Target: raw & 0x03FFFFFF,
	}
}
