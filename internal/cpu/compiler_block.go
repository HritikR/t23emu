package cpu

// CompiledInst is a pre-decoded instruction with a specialized fast execution handler.
type CompiledInst struct {
	Raw       uint32
	Opcode    uint8
	Rs        uint8
	Rt        uint8
	Rd        uint8
	Shamt     uint8
	Funct     uint8
	Immediate uint16
	Target    uint32
	SignImm   uint32
	Inst      Instruction
	Exec      func(c *CPU, ci *CompiledInst) bool
}

const maxBlockInstructions = 256

func (d *DynarecEngine) compileThreadedBlock(c *CPU, startPC uint32) *BasicBlock {
	var compiled []CompiledInst
	pc := startPC
	maxPC := pc + (maxBlockInstructions * 4)

	for pc < maxPC {
		if !c.Bus.HasMapping(pc) {
			break
		}

		raw := c.Bus.Read32(pc)
		inst := Decode(raw)

		ci := CompiledInst{
			Raw:       raw,
			Opcode:    inst.Opcode,
			Rs:        inst.Rs,
			Rt:        inst.Rt,
			Rd:        inst.Rd,
			Shamt:     inst.Shamt,
			Funct:     inst.Funct,
			Immediate: inst.Immediate,
			Target:    inst.Target,
			SignImm:   uint32(int32(int16(inst.Immediate))),
			Inst:      inst,
		}
		ci.Exec = selectFastExec(ci)
		compiled = append(compiled, ci)

		pc += 4

		if isTerminatingInst(inst) {
			// Include the delay slot instruction in the basic block if available
			if c.Bus.HasMapping(pc) {
				delayRaw := c.Bus.Read32(pc)
				delayInst := Decode(delayRaw)
				delayCI := CompiledInst{
					Raw:       delayRaw,
					Opcode:    delayInst.Opcode,
					Rs:        delayInst.Rs,
					Rt:        delayInst.Rt,
					Rd:        delayInst.Rd,
					Shamt:     delayInst.Shamt,
					Funct:     delayInst.Funct,
					Immediate: delayInst.Immediate,
					Target:    delayInst.Target,
					SignImm:   uint32(int32(int16(delayInst.Immediate))),
					Inst:      delayInst,
				}
				delayCI.Exec = selectFastExec(delayCI)
				compiled = append(compiled, delayCI)
			}
			break
		}
	}

	if len(compiled) == 0 {
		return &BasicBlock{
			StartPC:   startPC,
			InstCount: 0,
			Exec: func(c *CPU) bool {
				return false
			},
		}
	}

	return &BasicBlock{
		StartPC:   startPC,
		InstCount: len(compiled),
		Exec: func(c *CPU) bool {
			for i := 0; i < len(compiled); i++ {
				if !c.Running {
					return false
				}

				if !c.branchTaken && c.checkInterrupts() {
					c.Cycles++
					return false
				}

				c.InDelaySlot = c.branchTaken
				c.branchTaken = false

				ci := &compiled[i]

				// Fast pipeline advance matching c.Fetch()
				c.Instruction = ci.Raw
				c.CurrentPC = c.PC
				c.PC = c.NextPC
				c.NextPC = c.PC + 4

				if !ci.Exec(c, ci) {
					return false
				}

				c.Cycles++
			}
			return true
		},
	}
}

func selectFastExec(ci CompiledInst) func(c *CPU, ci *CompiledInst) bool {
	switch ci.Opcode {
	case OP_ADDIU:
		return fastADDIU
	case OP_ADDI:
		return fastADDI
	case OP_LUI:
		return fastLUI
	case OP_ANDI:
		return fastANDI
	case OP_ORI:
		return fastORI
	case OP_XORI:
		return fastXORI
	case OP_SLTI:
		return fastSLTI
	case OP_SLTIU:
		return fastSLTIU
	case OP_LW:
		return fastLW
	case OP_SW:
		return fastSW
	case OP_LB:
		return fastLB
	case OP_LBU:
		return fastLBU
	case OP_SB:
		return fastSB
	case OP_SPECIAL:
		switch ci.Funct {
		case FUNCT_ADDU:
			return fastADDU
		case FUNCT_SUBU:
			return fastSUBU
		case FUNCT_AND:
			return fastAND
		case FUNCT_OR:
			return fastOR
		case FUNCT_XOR:
			return fastXOR
		case FUNCT_NOR:
			return fastNOR
		case FUNCT_SLL:
			return fastSLL
		case FUNCT_SRL:
			return fastSRL
		case FUNCT_SRA:
			return fastSRA
		case FUNCT_SLLV:
			return fastSLLV
		case FUNCT_SRLV:
			return fastSRLV
		case FUNCT_SRAV:
			return fastSRAV
		case FUNCT_SLT:
			return fastSLT
		case FUNCT_SLTU:
			return fastSLTU
		}
	}

	// Fallback to standard Execute on decoded instruction
	return fallbackExec
}

func fallbackExec(c *CPU, ci *CompiledInst) bool {
	c.Execute(ci.Inst)
	return true
}

func fastADDIU(c *CPU, ci *CompiledInst) bool {
	if ci.Rt != 0 {
		c.Regs[ci.Rt] = c.ReadRegister(ci.Rs) + ci.SignImm
	}
	c.retire()
	return true
}

func fastADDI(c *CPU, ci *CompiledInst) bool {
	rs := int32(c.ReadRegister(ci.Rs))
	imm := int32(int16(ci.Immediate))
	result := rs + imm
	if (rs < 0) == (imm < 0) && (result < 0) != (rs < 0) {
		c.Exception(EXC_OV, 0)
		return false
	}
	if ci.Rt != 0 {
		c.Regs[ci.Rt] = uint32(result)
	}
	c.retire()
	return true
}

func fastLUI(c *CPU, ci *CompiledInst) bool {
	if ci.Rt != 0 {
		c.Regs[ci.Rt] = uint32(ci.Immediate) << 16
	}
	c.retire()
	return true
}

func fastANDI(c *CPU, ci *CompiledInst) bool {
	if ci.Rt != 0 {
		c.Regs[ci.Rt] = c.ReadRegister(ci.Rs) & uint32(ci.Immediate)
	}
	c.retire()
	return true
}

func fastORI(c *CPU, ci *CompiledInst) bool {
	if ci.Rt != 0 {
		c.Regs[ci.Rt] = c.ReadRegister(ci.Rs) | uint32(ci.Immediate)
	}
	c.retire()
	return true
}

func fastXORI(c *CPU, ci *CompiledInst) bool {
	if ci.Rt != 0 {
		c.Regs[ci.Rt] = c.ReadRegister(ci.Rs) ^ uint32(ci.Immediate)
	}
	c.retire()
	return true
}

func fastSLTI(c *CPU, ci *CompiledInst) bool {
	if ci.Rt != 0 {
		if int32(c.ReadRegister(ci.Rs)) < int32(int16(ci.Immediate)) {
			c.Regs[ci.Rt] = 1
		} else {
			c.Regs[ci.Rt] = 0
		}
	}
	c.retire()
	return true
}

func fastSLTIU(c *CPU, ci *CompiledInst) bool {
	if ci.Rt != 0 {
		if c.ReadRegister(ci.Rs) < ci.SignImm {
			c.Regs[ci.Rt] = 1
		} else {
			c.Regs[ci.Rt] = 0
		}
	}
	c.retire()
	return true
}

func fastADDU(c *CPU, ci *CompiledInst) bool {
	if ci.Rd != 0 {
		c.Regs[ci.Rd] = c.ReadRegister(ci.Rs) + c.ReadRegister(ci.Rt)
	}
	c.retire()
	return true
}

func fastSUBU(c *CPU, ci *CompiledInst) bool {
	if ci.Rd != 0 {
		c.Regs[ci.Rd] = c.ReadRegister(ci.Rs) - c.ReadRegister(ci.Rt)
	}
	c.retire()
	return true
}

func fastAND(c *CPU, ci *CompiledInst) bool {
	if ci.Rd != 0 {
		c.Regs[ci.Rd] = c.ReadRegister(ci.Rs) & c.ReadRegister(ci.Rt)
	}
	c.retire()
	return true
}

func fastOR(c *CPU, ci *CompiledInst) bool {
	if ci.Rd != 0 {
		c.Regs[ci.Rd] = c.ReadRegister(ci.Rs) | c.ReadRegister(ci.Rt)
	}
	c.retire()
	return true
}

func fastXOR(c *CPU, ci *CompiledInst) bool {
	if ci.Rd != 0 {
		c.Regs[ci.Rd] = c.ReadRegister(ci.Rs) ^ c.ReadRegister(ci.Rt)
	}
	c.retire()
	return true
}

func fastNOR(c *CPU, ci *CompiledInst) bool {
	if ci.Rd != 0 {
		c.Regs[ci.Rd] = ^(c.ReadRegister(ci.Rs) | c.ReadRegister(ci.Rt))
	}
	c.retire()
	return true
}

func fastSLL(c *CPU, ci *CompiledInst) bool {
	if ci.Rd != 0 {
		c.Regs[ci.Rd] = c.ReadRegister(ci.Rt) << ci.Shamt
	}
	c.retire()
	return true
}

func fastSRL(c *CPU, ci *CompiledInst) bool {
	if ci.Rd != 0 {
		c.Regs[ci.Rd] = c.ReadRegister(ci.Rt) >> ci.Shamt
	}
	c.retire()
	return true
}

func fastSRA(c *CPU, ci *CompiledInst) bool {
	if ci.Rd != 0 {
		c.Regs[ci.Rd] = uint32(int32(c.ReadRegister(ci.Rt)) >> ci.Shamt)
	}
	c.retire()
	return true
}

func fastSLLV(c *CPU, ci *CompiledInst) bool {
	if ci.Rd != 0 {
		c.Regs[ci.Rd] = c.ReadRegister(ci.Rt) << (c.ReadRegister(ci.Rs) & 0x1F)
	}
	c.retire()
	return true
}

func fastSRLV(c *CPU, ci *CompiledInst) bool {
	if ci.Rd != 0 {
		c.Regs[ci.Rd] = c.ReadRegister(ci.Rt) >> (c.ReadRegister(ci.Rs) & 0x1F)
	}
	c.retire()
	return true
}

func fastSRAV(c *CPU, ci *CompiledInst) bool {
	if ci.Rd != 0 {
		c.Regs[ci.Rd] = uint32(int32(c.ReadRegister(ci.Rt)) >> (c.ReadRegister(ci.Rs) & 0x1F))
	}
	c.retire()
	return true
}

func fastSLT(c *CPU, ci *CompiledInst) bool {
	if ci.Rd != 0 {
		if int32(c.ReadRegister(ci.Rs)) < int32(c.ReadRegister(ci.Rt)) {
			c.Regs[ci.Rd] = 1
		} else {
			c.Regs[ci.Rd] = 0
		}
	}
	c.retire()
	return true
}

func fastSLTU(c *CPU, ci *CompiledInst) bool {
	if ci.Rd != 0 {
		if c.ReadRegister(ci.Rs) < c.ReadRegister(ci.Rt) {
			c.Regs[ci.Rd] = 1
		} else {
			c.Regs[ci.Rd] = 0
		}
	}
	c.retire()
	return true
}

func fastLW(c *CPU, ci *CompiledInst) bool {
	addr := c.ReadRegister(ci.Rs) + ci.SignImm
	if addr&3 != 0 {
		c.Exception(EXC_ADEL, addr)
		return false
	}
	val := c.read32(addr)
	if ci.Rt != 0 {
		c.Regs[ci.Rt] = val
	}
	c.retire()
	return true
}

func fastSW(c *CPU, ci *CompiledInst) bool {
	addr := c.ReadRegister(ci.Rs) + ci.SignImm
	if addr&3 != 0 {
		c.Exception(EXC_ADES, addr)
		return false
	}
	c.write32(addr, c.ReadRegister(ci.Rt))
	c.retire()
	return true
}

func fastLB(c *CPU, ci *CompiledInst) bool {
	addr := c.ReadRegister(ci.Rs) + ci.SignImm
	val := int32(int8(c.read8(addr)))
	if ci.Rt != 0 {
		c.Regs[ci.Rt] = uint32(val)
	}
	c.retire()
	return true
}

func fastLBU(c *CPU, ci *CompiledInst) bool {
	addr := c.ReadRegister(ci.Rs) + ci.SignImm
	val := c.read8(addr)
	if ci.Rt != 0 {
		c.Regs[ci.Rt] = uint32(val)
	}
	c.retire()
	return true
}

func fastSB(c *CPU, ci *CompiledInst) bool {
	addr := c.ReadRegister(ci.Rs) + ci.SignImm
	c.write8(addr, byte(c.ReadRegister(ci.Rt)))
	c.retire()
	return true
}
