package cpu

import "fmt"

// Execute runs a decoded instruction.
func (c *CPU) Execute(inst Instruction) {

	switch inst.Opcode {

	case OP_SPECIAL:

		c.executeRType(inst)

	case OP_ADDI:

		c.executeADDI(inst)

	default:

		panic(
			fmt.Sprintf(
				"unsupported opcode: %d",
				inst.Opcode,
			),
		)
	}
}

// executeRType handles opcode 0 instructions.
//
// R-type operations are selected by the funct field.
func (c *CPU) executeRType(inst Instruction) {

	switch inst.Funct {

	case FUNCT_SLL:

		// NOP is encoded as:
		//
		// sll $zero,$zero,0
		//
		// Because writing to $zero has no effect,
		// this naturally does nothing.

		return

	case FUNCT_ADD:

		c.executeADD(inst)

	default:

		panic(
			fmt.Sprintf(
				"unsupported R-type funct: %d",
				inst.Funct,
			),
		)
	}
}

// ADD
//
// rd = rs + rt
func (c *CPU) executeADD(inst Instruction) {

	rs := c.ReadRegister(
		inst.Rs,
	)

	rt := c.ReadRegister(
		inst.Rt,
	)

	result := rs + rt

	c.WriteRegister(
		inst.Rd,
		result,
	)
}

// ADDI
//
// rt = rs + immediate
func (c *CPU) executeADDI(inst Instruction) {

	rs := c.ReadRegister(
		inst.Rs,
	)

	// Sign extend 16-bit immediate
	immediate := int32(
		int16(inst.Immediate),
	)

	result := uint32(
		int32(rs) + immediate,
	)

	c.WriteRegister(
		inst.Rt,
		result,
	)
}
