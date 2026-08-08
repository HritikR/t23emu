package device

// Ingenic I2C controller register offsets. The block is DesignWare
// derived, and the kernel's own boot message pins the map down: the
// driver prints the setup and hold times it just programmed, and
// "set:249 hold:250" is exactly the 0xF9 written to SDASU and the 0xFA
// written to SDAHD.
const (
	I2C_CTRL   uint32 = 0x00 // master enable, addressing mode, speed
	I2C_TAR    uint32 = 0x04 // target address
	I2C_DC     uint32 = 0x10 // data and command
	I2C_SHCNT  uint32 = 0x14 // standard-speed SCL high count
	I2C_SLCNT  uint32 = 0x18 // standard-speed SCL low count
	I2C_INTST  uint32 = 0x2C // interrupt status
	I2C_INTM   uint32 = 0x30 // interrupt mask
	I2C_RAWST  uint32 = 0x34 // raw interrupt status
	I2C_TXTL   uint32 = 0x3C // transmit FIFO threshold
	I2C_CLRIN  uint32 = 0x40 // clear combined interrupt
	I2C_CLRTXO uint32 = 0x4C // clear transmit overflow
	I2C_CLRTXA uint32 = 0x54 // clear transmit abort
	I2C_CLRSTP uint32 = 0x60 // clear stop detected
	I2C_TXFLR  uint32 = 0x74 // transmit FIFO level
	I2C_RXFLR  uint32 = 0x78 // receive FIFO level
	I2C_ENB    uint32 = 0x6C // enable request
	I2C_STA    uint32 = 0x70 // controller status
	I2C_SDAHD  uint32 = 0x7C // SDA hold time
	I2C_SDASU  uint32 = 0x94 // SDA setup time
	I2C_ENSTA  uint32 = 0x9C // enable status
)

// I2C_ENB_ENABLE is the enable request bit, and the bit the driver waits
// to see come back in ENSTA.
const I2C_ENB_ENABLE uint32 = 1 << 0

const (
	i2cDataCmdRead uint32 = 1 << 8

	i2cStatusTXNotFull  uint32 = 1 << 1
	i2cStatusTXEmpty    uint32 = 1 << 2
	i2cStatusRXNotEmpty uint32 = 1 << 3

	i2cIntTXOver  uint32 = 1 << 3
	i2cIntTXEmpty uint32 = 1 << 4
	i2cIntTXAbort uint32 = 1 << 6
	i2cIntRXFull  uint32 = 1 << 2
)

// I2CDevice defines the interface for devices connected to the I2C bus.
type I2CDevice interface {
	WriteI2C(val byte)
	ReadI2C() byte
}

type I2C struct {
	*RegisterBlock

	Interrupt func(assert bool)

	devices    map[uint32]I2CDevice
	targetAddr uint32

	intMask  uint32
	pending  uint32
	irqState bool
}

// AttachDevice registers an I2CDevice at a specific target address.
func (i *I2C) AttachDevice(targetAddress uint32, dev I2CDevice) {
	i.devices[targetAddress] = dev
}

// NewI2C creates a minimal I2C controller.
//
// It models the enable handshake plus the small PIO FIFO behaviour that
// the kernel driver polls.
func NewI2C(name string) *I2C {
	i2c := &I2C{
		RegisterBlock: NewRegisterBlock(name, 0x1000),
		devices:       make(map[uint32]I2CDevice),
	}

	names := map[uint32]string{
		I2C_CTRL:   "CTRL",
		I2C_TAR:    "TAR",
		I2C_DC:     "DC",
		I2C_SHCNT:  "SHCNT",
		I2C_SLCNT:  "SLCNT",
		I2C_INTST:  "INTST",
		I2C_INTM:   "INTM",
		I2C_RAWST:  "RAWST",
		I2C_TXTL:   "TXTL",
		I2C_CLRIN:  "CLRIN",
		I2C_CLRTXO: "CLRTXO",
		I2C_CLRTXA: "CLRTXA",
		I2C_CLRSTP: "CLRSTP",
		I2C_TXFLR:  "TXFLR",
		I2C_RXFLR:  "RXFLR",
		I2C_ENB:    "ENB",
		I2C_STA:    "STA",
		I2C_SDAHD:  "SDAHD",
		I2C_SDASU:  "SDASU",
		I2C_ENSTA:  "ENSTA",
	}
	for offset, name := range names {
		i2c.SetName(offset, name)
	}

	// ENSTA reports the enable actually taking effect. Real hardware
	// takes a few cycles over it; acknowledging immediately is fine,
	// because the driver only ever polls for the bit rather than timing
	// how long it takes to appear.
	var enabled uint32
	var rxFIFO []uint32
	i2c.SetWriteFunc(I2C_ENB, func(value uint32) {
		enabled = value & I2C_ENB_ENABLE
	})
	i2c.SetReadFunc(I2C_ENSTA, func() uint32 {
		return enabled
	})
	i2c.SetWriteFunc(I2C_TAR, func(value uint32) {
		i2c.targetAddr = value
	})
	i2c.SetReadFunc(I2C_TAR, func() uint32 {
		return i2c.targetAddr
	})
	i2c.SetWriteFunc(I2C_INTM, func(value uint32) {
		i2c.intMask = value
		i2c.updateInterrupt()
	})
	i2c.SetReadFunc(I2C_INTST, func() uint32 {
		return i2c.pending & i2c.intMask
	})
	i2c.SetReadFunc(I2C_RAWST, func() uint32 {
		return i2c.pending
	})
	i2c.SetReadFunc(I2C_CLRIN, func() uint32 {
		i2c.pending = 0
		i2c.updateInterrupt()
		return 0
	})
	i2c.SetReadFunc(I2C_CLRTXO, func() uint32 {
		i2c.pending &^= i2cIntTXOver
		i2c.updateInterrupt()
		return 0
	})
	i2c.SetReadFunc(I2C_CLRTXA, func() uint32 {
		i2c.pending &^= i2cIntTXAbort
		i2c.updateInterrupt()
		return 0
	})
	i2c.SetReadFunc(I2C_CLRSTP, func() uint32 {
		i2c.updateInterrupt()
		return 0
	})
	i2c.SetReadFunc(I2C_STA, func() uint32 {
		status := i2cStatusTXNotFull | i2cStatusTXEmpty
		if len(rxFIFO) > 0 {
			status |= i2cStatusRXNotEmpty
		}
		return status
	})
	i2c.SetReadFunc(I2C_TXFLR, func() uint32 {
		return 0
	})
	i2c.SetReadFunc(I2C_RXFLR, func() uint32 {
		return uint32(len(rxFIFO))
	})
	i2c.SetWriteFunc(I2C_DC, func(value uint32) {
		dev := i2c.devices[i2c.targetAddr]
		if value&i2cDataCmdRead != 0 {
			var val byte
			if dev != nil {
				val = dev.ReadI2C()
			}
			rxFIFO = append(rxFIFO, uint32(val))
			i2c.pending |= i2cIntRXFull
			i2c.updateInterrupt()
			return
		}

		if dev != nil {
			dev.WriteI2C(byte(value))
		}
		i2c.pending |= i2cIntTXEmpty
		i2c.updateInterrupt()
	})
	i2c.SetReadFunc(I2C_DC, func() uint32 {
		if len(rxFIFO) == 0 {
			return 0
		}
		value := rxFIFO[0]
		rxFIFO = rxFIFO[1:]
		if len(rxFIFO) == 0 {
			i2c.pending &^= i2cIntRXFull
			i2c.updateInterrupt()
		}
		return value
	})

	return i2c
}

func (i *I2C) updateInterrupt() {
	assert := i.pending&i.intMask != 0
	if assert == i.irqState {
		return
	}
	i.irqState = assert
	if i.Interrupt != nil {
		i.Interrupt(assert)
	}
}
