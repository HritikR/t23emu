package cpu

import "fmt"

// Execute runs a decoded instruction.
func (c *CPU) Execute(inst Instruction) {

	switch inst.Opcode {

	case OP_SPECIAL:

		c.executeRType(inst)

	case OP_J:
		c.executeJ(inst)

	case OP_JAL:
		c.executeJAL(inst)

	case OP_BEQ:
		c.executeBEQ(inst)

	case OP_BNE:
		c.executeBNE(inst)

	case OP_ADDI:

		c.executeADDI(inst)

	case OP_ANDI:
		c.executeANDI(inst)

	case OP_ORI:
		c.executeORI(inst)

	case OP_LUI:
		c.executeLUI(inst)

	case OP_LW:
		c.executeLW(inst)

	case OP_SW:
		c.executeSW(inst)

	case OP_COP0:
		c.executeCOP0(inst)

	default:
		c.Exception(EXC_RI, 0)
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

	case FUNCT_JR:
		c.executeJR(inst)

	case FUNCT_ADD:

		c.executeADD(inst)

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

	case FUNCT_SYSCALL:
		c.Exception(EXC_SYS, 0)

	case FUNCT_BREAK:
		c.Exception(EXC_BP, 0)

	default:
		c.Exception(EXC_RI, 0)
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

// LW
//
// rt = memory[rs + immediate]
func (c *CPU) executeLW(inst Instruction) {

	base := c.ReadRegister(
		inst.Rs,
	)

	offset := int32(
		int16(inst.Immediate),
	)

	address := uint32(
		int32(base) + offset,
	)

	if !c.Bus.HasMapping(address) {
		c.Exception(EXC_ADEL, address)
		return
	}

	value := c.Bus.Read32(
		address,
	)

	c.WriteRegister(
		inst.Rt,
		value,
	)
}

// SW
//
// memory[rs + immediate] = rt
func (c *CPU) executeSW(inst Instruction) {

	base := c.ReadRegister(
		inst.Rs,
	)

	offset := int32(
		int16(inst.Immediate),
	)

	address := uint32(
		int32(base) + offset,
	)

	if !c.Bus.HasMapping(address) {
		c.Exception(EXC_ADES, address)
		return
	}

	value := c.ReadRegister(
		inst.Rt,
	)

	c.Bus.Write32(
		address,
		value,
	)
}

func (c *CPU) executeJ(inst Instruction) {
	c.PC = (c.PC & 0xF0000000) | (inst.Target << 2)
}

func (c *CPU) executeJAL(inst Instruction) {
	c.WriteRegister(31, c.PC)
	c.PC = (c.PC & 0xF0000000) | (inst.Target << 2)
}

func (c *CPU) executeJR(inst Instruction) {
	c.PC = c.ReadRegister(inst.Rs)
}

func (c *CPU) executeBEQ(inst Instruction) {
	rs := c.ReadRegister(inst.Rs)
	rt := c.ReadRegister(inst.Rt)
	if rs == rt {
		offset := int32(int16(inst.Immediate)) << 2
		c.PC = uint32(int32(c.PC) + offset)
	}
}

func (c *CPU) executeBNE(inst Instruction) {
	rs := c.ReadRegister(inst.Rs)
	rt := c.ReadRegister(inst.Rt)
	if rs != rt {
		offset := int32(int16(inst.Immediate)) << 2
		c.PC = uint32(int32(c.PC) + offset)
	}
}

func (c *CPU) executeLUI(inst Instruction) {
	value := uint32(inst.Immediate) << 16
	c.WriteRegister(inst.Rt, value)
}

func (c *CPU) executeANDI(inst Instruction) {
	rs := c.ReadRegister(inst.Rs)
	imm := uint32(inst.Immediate)
	c.WriteRegister(inst.Rt, rs & imm)
}

func (c *CPU) executeORI(inst Instruction) {
	rs := c.ReadRegister(inst.Rs)
	imm := uint32(inst.Immediate)
	c.WriteRegister(inst.Rt, rs | imm)
}

func (c *CPU) executeAND(inst Instruction) {
	rs := c.ReadRegister(inst.Rs)
	rt := c.ReadRegister(inst.Rt)
	c.WriteRegister(inst.Rd, rs & rt)
}

func (c *CPU) executeOR(inst Instruction) {
	rs := c.ReadRegister(inst.Rs)
	rt := c.ReadRegister(inst.Rt)
	c.WriteRegister(inst.Rd, rs | rt)
}

func (c *CPU) executeXOR(inst Instruction) {
	rs := c.ReadRegister(inst.Rs)
	rt := c.ReadRegister(inst.Rt)
	c.WriteRegister(inst.Rd, rs ^ rt)
}

func (c *CPU) executeNOR(inst Instruction) {
	rs := c.ReadRegister(inst.Rs)
	rt := c.ReadRegister(inst.Rt)
	c.WriteRegister(inst.Rd, ^(rs | rt))
}

func (c *CPU) executeSLT(inst Instruction) {
	rs := int32(c.ReadRegister(inst.Rs))
	rt := int32(c.ReadRegister(inst.Rt))
	if rs < rt {
		c.WriteRegister(inst.Rd, 1)
	} else {
		c.WriteRegister(inst.Rd, 0)
	}
}

func (c *CPU) executeSLL(inst Instruction) {
	rt := c.ReadRegister(inst.Rt)
	c.WriteRegister(inst.Rd, rt << inst.Shamt)
}

func (c *CPU) executeSRL(inst Instruction) {
	rt := c.ReadRegister(inst.Rt)
	c.WriteRegister(inst.Rd, rt >> inst.Shamt)
}

func (c *CPU) executeSRA(inst Instruction) {
	rt := int32(c.ReadRegister(inst.Rt))
	c.WriteRegister(inst.Rd, uint32(rt >> inst.Shamt))
}

func (c *CPU) executeCOP0(inst Instruction) {
	// ERET instruction has funct code 24 and Rs=16
	if inst.Funct == 24 && inst.Rs == 16 {
		c.executeERET(inst)
		return
	}

	switch inst.Rs {
	case 0: // MFC0 rt, rd
		c.executeMFC0(inst)
	case 4: // MTC0 rt, rd
		c.executeMTC0(inst)
	default:
		panic(fmt.Sprintf("unsupported COP0 sub-op (Rs): %d", inst.Rs))
	}
}

func (c *CPU) executeMFC0(inst Instruction) {
	value := c.CP0[inst.Rd]
	c.WriteRegister(inst.Rt, value)
}

func (c *CPU) executeMTC0(inst Instruction) {
	value := c.ReadRegister(inst.Rt)
	c.CP0[inst.Rd] = value
}

func (c *CPU) executeERET(inst Instruction) {
	// Set PC to CP0 Exception Program Counter (EPC)
	c.PC = c.CP0[CP0_EPC]
	// Clear the EXL (Exception Level) bit (bit 1) in Status register
	c.CP0[CP0_STATUS] &= ^uint32(0x2)
}
