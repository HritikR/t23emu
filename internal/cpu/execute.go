package cpu

import "fmt"

func (c *CPU) Execute(inst Instruction) {

	switch inst.Opcode {

	case OP_SPECIAL:

		c.executeSpecial(inst)

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

func (c *CPU) executeSpecial(inst Instruction) {

	switch inst.Funct {

	case 0:

		// NOP
		//
		// In MIPS:
		// sll $zero,$zero,0
		//
		// It does nothing.

		return

	default:

		panic(
			fmt.Sprintf(
				"unsupported funct: %d",
				inst.Funct,
			),
		)
	}
}

func (c *CPU) executeADDI(inst Instruction) {

	// Sign extend immediate
	value := int32(
		int16(inst.Immediate),
	)

	rs := c.ReadRegister(
		inst.Rs,
	)

	result := uint32(
		int32(rs) + value,
	)

	c.WriteRegister(
		inst.Rt,
		result,
	)
}
