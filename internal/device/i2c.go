package device

// Ingenic I2C controller register offsets. The block is DesignWare
// derived, and the kernel's own boot message pins the map down: the
// driver prints the setup and hold times it just programmed, and
// "set:249 hold:250" is exactly the 0xF9 written to SDASU and the 0xFA
// written to SDAHD.
const (
	I2C_CTRL  uint32 = 0x00 // master enable, addressing mode, speed
	I2C_TAR   uint32 = 0x04 // target address
	I2C_DC    uint32 = 0x10 // data and command
	I2C_SHCNT uint32 = 0x14 // standard-speed SCL high count
	I2C_SLCNT uint32 = 0x18 // standard-speed SCL low count
	I2C_INTST uint32 = 0x2C // interrupt status
	I2C_INTM  uint32 = 0x30 // interrupt mask
	I2C_ENB   uint32 = 0x6C // enable request
	I2C_STA   uint32 = 0x70 // controller status
	I2C_SDAHD uint32 = 0x7C // SDA hold time
	I2C_SDASU uint32 = 0x94 // SDA setup time
	I2C_ENSTA uint32 = 0x9C // enable status
)

// I2C_ENB_ENABLE is the enable request bit, and the bit the driver waits
// to see come back in ENSTA.
const I2C_ENB_ENABLE uint32 = 1 << 0

// NewI2C creates a minimal I2C controller.
//
// It models one thing: the enable handshake. The driver writes ENB and
// then polls ENSTA until the controller acknowledges by setting the same
// bit. Left to the catch-all register file, ENSTA reads as zero forever,
// the poll runs out its jiffy timeout, and the boot loses five seconds
// to "enable i2c0 failed" before giving up on the bus.
//
// Nothing about an actual transfer is modelled. STA, the FIFO level
// registers and the abort register are all still plain storage, so a
// driver that gets past enable and tries to talk to a device on the bus
// will not get an answer from it.
func NewI2C(name string) *RegisterBlock {
	i2c := NewRegisterBlock(name, 0x1000)

	names := map[uint32]string{
		I2C_CTRL:  "CTRL",
		I2C_TAR:   "TAR",
		I2C_DC:    "DC",
		I2C_SHCNT: "SHCNT",
		I2C_SLCNT: "SLCNT",
		I2C_INTST: "INTST",
		I2C_INTM:  "INTM",
		I2C_ENB:   "ENB",
		I2C_STA:   "STA",
		I2C_SDAHD: "SDAHD",
		I2C_SDASU: "SDASU",
		I2C_ENSTA: "ENSTA",
	}
	for offset, name := range names {
		i2c.SetName(offset, name)
	}

	// ENSTA reports the enable actually taking effect. Real hardware
	// takes a few cycles over it; acknowledging immediately is fine,
	// because the driver only ever polls for the bit rather than timing
	// how long it takes to appear.
	var enabled uint32
	i2c.SetWriteFunc(I2C_ENB, func(value uint32) {
		enabled = value & I2C_ENB_ENABLE
	})
	i2c.SetReadFunc(I2C_ENSTA, func() uint32 {
		return enabled
	})

	return i2c
}
