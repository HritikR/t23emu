package dynarec

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
	Exec      func(c CPUContext, ci *CompiledInst) bool
}

const maxBlockInstructions = 256

func (e *Engine) compileThreadedBlock(c CPUContext, startPC uint32) *BasicBlock {
	var compiled []CompiledInst
	pc := startPC
	maxPC := pc + (maxBlockInstructions * 4)

	for pc < maxPC {
		if !c.HasMapping(pc) {
			break
		}

		raw := c.Read32(pc)
		inst := DecodeInst(raw)

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
			if c.HasMapping(pc) {
				delayRaw := c.Read32(pc)
				delayInst := DecodeInst(delayRaw)
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
			Exec: func(c CPUContext) bool {
				return false
			},
		}
	}

	return &BasicBlock{
		StartPC:   startPC,
		InstCount: len(compiled),
		Exec: func(c CPUContext) bool {
			for i := 0; i < len(compiled); i++ {
				if !c.IsRunning() {
					return false
				}

				if !c.IsBranchTaken() && c.CheckInterrupts() {
					c.IncCycles(1)
					return false
				}

				c.SetInDelaySlot(c.IsBranchTaken())
				c.SetBranchTaken(false)

				ci := &compiled[i]

				c.SetInstruction(ci.Raw)
				currPC := c.GetPC()
				nextPC := c.GetNextPC()
				c.SetCurrentPC(currPC)
				c.SetPC(nextPC)
				c.SetNextPC(nextPC + 4)

				if !ci.Exec(c, ci) {
					return false
				}

				c.IncCycles(1)
			}
			return true
		},
	}
}

func selectFastExec(ci CompiledInst) func(c CPUContext, ci *CompiledInst) bool {
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

	return fallbackExec
}

func fallbackExec(c CPUContext, ci *CompiledInst) bool {
	return c.ExecuteRaw(ci.Raw)
}

func fastADDIU(c CPUContext, ci *CompiledInst) bool {
	if ci.Rt != 0 {
		c.WriteReg(ci.Rt, c.ReadReg(ci.Rs)+ci.SignImm)
	}
	c.Retire()
	return true
}

func fastADDI(c CPUContext, ci *CompiledInst) bool {
	rs := int32(c.ReadReg(ci.Rs))
	imm := int32(int16(ci.Immediate))
	result := rs + imm
	if (rs < 0) == (imm < 0) && (result < 0) != (rs < 0) {
		c.RaiseException(EXC_OV, 0)
		return false
	}
	if ci.Rt != 0 {
		c.WriteReg(ci.Rt, uint32(result))
	}
	c.Retire()
	return true
}

func fastLUI(c CPUContext, ci *CompiledInst) bool {
	if ci.Rt != 0 {
		c.WriteReg(ci.Rt, uint32(ci.Immediate)<<16)
	}
	c.Retire()
	return true
}

func fastANDI(c CPUContext, ci *CompiledInst) bool {
	if ci.Rt != 0 {
		c.WriteReg(ci.Rt, c.ReadReg(ci.Rs)&uint32(ci.Immediate))
	}
	c.Retire()
	return true
}

func fastORI(c CPUContext, ci *CompiledInst) bool {
	if ci.Rt != 0 {
		c.WriteReg(ci.Rt, c.ReadReg(ci.Rs)|uint32(ci.Immediate))
	}
	c.Retire()
	return true
}

func fastXORI(c CPUContext, ci *CompiledInst) bool {
	if ci.Rt != 0 {
		c.WriteReg(ci.Rt, c.ReadReg(ci.Rs)^uint32(ci.Immediate))
	}
	c.Retire()
	return true
}

func fastSLTI(c CPUContext, ci *CompiledInst) bool {
	if ci.Rt != 0 {
		if int32(c.ReadReg(ci.Rs)) < int32(int16(ci.Immediate)) {
			c.WriteReg(ci.Rt, 1)
		} else {
			c.WriteReg(ci.Rt, 0)
		}
	}
	c.Retire()
	return true
}

func fastSLTIU(c CPUContext, ci *CompiledInst) bool {
	if ci.Rt != 0 {
		if c.ReadReg(ci.Rs) < ci.SignImm {
			c.WriteReg(ci.Rt, 1)
		} else {
			c.WriteReg(ci.Rt, 0)
		}
	}
	c.Retire()
	return true
}

func fastADDU(c CPUContext, ci *CompiledInst) bool {
	if ci.Rd != 0 {
		c.WriteReg(ci.Rd, c.ReadReg(ci.Rs)+c.ReadReg(ci.Rt))
	}
	c.Retire()
	return true
}

func fastSUBU(c CPUContext, ci *CompiledInst) bool {
	if ci.Rd != 0 {
		c.WriteReg(ci.Rd, c.ReadReg(ci.Rs)-c.ReadReg(ci.Rt))
	}
	c.Retire()
	return true
}

func fastAND(c CPUContext, ci *CompiledInst) bool {
	if ci.Rd != 0 {
		c.WriteReg(ci.Rd, c.ReadReg(ci.Rs)&c.ReadReg(ci.Rt))
	}
	c.Retire()
	return true
}

func fastOR(c CPUContext, ci *CompiledInst) bool {
	if ci.Rd != 0 {
		c.WriteReg(ci.Rd, c.ReadReg(ci.Rs)|c.ReadReg(ci.Rt))
	}
	c.Retire()
	return true
}

func fastXOR(c CPUContext, ci *CompiledInst) bool {
	if ci.Rd != 0 {
		c.WriteReg(ci.Rd, c.ReadReg(ci.Rs)^c.ReadReg(ci.Rt))
	}
	c.Retire()
	return true
}

func fastNOR(c CPUContext, ci *CompiledInst) bool {
	if ci.Rd != 0 {
		c.WriteReg(ci.Rd, ^(c.ReadReg(ci.Rs) | c.ReadReg(ci.Rt)))
	}
	c.Retire()
	return true
}

func fastSLL(c CPUContext, ci *CompiledInst) bool {
	if ci.Rd != 0 {
		c.WriteReg(ci.Rd, c.ReadReg(ci.Rt)<<ci.Shamt)
	}
	c.Retire()
	return true
}

func fastSRL(c CPUContext, ci *CompiledInst) bool {
	if ci.Rd != 0 {
		c.WriteReg(ci.Rd, c.ReadReg(ci.Rt)>>ci.Shamt)
	}
	c.Retire()
	return true
}

func fastSRA(c CPUContext, ci *CompiledInst) bool {
	if ci.Rd != 0 {
		c.WriteReg(ci.Rd, uint32(int32(c.ReadReg(ci.Rt))>>ci.Shamt))
	}
	c.Retire()
	return true
}

func fastSLLV(c CPUContext, ci *CompiledInst) bool {
	if ci.Rd != 0 {
		c.WriteReg(ci.Rd, c.ReadReg(ci.Rt)<<(c.ReadReg(ci.Rs)&0x1F))
	}
	c.Retire()
	return true
}

func fastSRLV(c CPUContext, ci *CompiledInst) bool {
	if ci.Rd != 0 {
		c.WriteReg(ci.Rd, c.ReadReg(ci.Rt)>>(c.ReadReg(ci.Rs)&0x1F))
	}
	c.Retire()
	return true
}

func fastSRAV(c CPUContext, ci *CompiledInst) bool {
	if ci.Rd != 0 {
		c.WriteReg(ci.Rd, uint32(int32(c.ReadReg(ci.Rt))>>(c.ReadReg(ci.Rs)&0x1F)))
	}
	c.Retire()
	return true
}

func fastSLT(c CPUContext, ci *CompiledInst) bool {
	if ci.Rd != 0 {
		if int32(c.ReadReg(ci.Rs)) < int32(c.ReadReg(ci.Rt)) {
			c.WriteReg(ci.Rd, 1)
		} else {
			c.WriteReg(ci.Rd, 0)
		}
	}
	c.Retire()
	return true
}

func fastSLTU(c CPUContext, ci *CompiledInst) bool {
	if ci.Rd != 0 {
		if c.ReadReg(ci.Rs) < c.ReadReg(ci.Rt) {
			c.WriteReg(ci.Rd, 1)
		} else {
			c.WriteReg(ci.Rd, 0)
		}
	}
	c.Retire()
	return true
}

func fastLW(c CPUContext, ci *CompiledInst) bool {
	addr := c.ReadReg(ci.Rs) + ci.SignImm
	if addr&3 != 0 {
		c.RaiseException(EXC_ADEL, addr)
		return false
	}
	val := c.Read32(addr)
	if ci.Rt != 0 {
		c.WriteReg(ci.Rt, val)
	}
	c.Retire()
	return true
}

func fastSW(c CPUContext, ci *CompiledInst) bool {
	addr := c.ReadReg(ci.Rs) + ci.SignImm
	if addr&3 != 0 {
		c.RaiseException(EXC_ADES, addr)
		return false
	}
	c.Write32(addr, c.ReadReg(ci.Rt))
	c.Retire()
	return true
}

func fastLB(c CPUContext, ci *CompiledInst) bool {
	addr := c.ReadReg(ci.Rs) + ci.SignImm
	val := int32(int8(c.Read8(addr)))
	if ci.Rt != 0 {
		c.WriteReg(ci.Rt, uint32(val))
	}
	c.Retire()
	return true
}

func fastLBU(c CPUContext, ci *CompiledInst) bool {
	addr := c.ReadReg(ci.Rs) + ci.SignImm
	val := c.Read8(addr)
	if ci.Rt != 0 {
		c.WriteReg(ci.Rt, uint32(val))
	}
	c.Retire()
	return true
}

func fastSB(c CPUContext, ci *CompiledInst) bool {
	addr := c.ReadReg(ci.Rs) + ci.SignImm
	c.Write8(addr, byte(c.ReadReg(ci.Rt)))
	c.Retire()
	return true
}
