package cpu

// ReadRegister returns the value of a MIPS register.
func (c *CPU) ReadRegister(index uint8) uint32 {

	// Register 0 ($zero) always returns 0
	if index == 0 {
		return 0
	}

	return c.Regs[index]
}

// WriteRegister writes a value into a MIPS register.
func (c *CPU) WriteRegister(index uint8, value uint32) {

	// Register 0 ($zero) cannot be modified
	if index == 0 {
		return
	}

	c.Regs[index] = value
}
