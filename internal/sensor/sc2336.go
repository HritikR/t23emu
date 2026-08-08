package sensor

// SC2336 represents the SC2336 image sensor connected via I2C.
type SC2336 struct {
	addrBytes []byte
}

// NewSC2336 creates a new SC2336 sensor instance.
func NewSC2336() *SC2336 {
	return &SC2336{
		addrBytes: make([]byte, 0, 2),
	}
}

// WriteI2C receives a byte written to the sensor (e.g. register address bytes).
func (s *SC2336) WriteI2C(val byte) {
	s.addrBytes = append(s.addrBytes, val)
	if len(s.addrBytes) > 2 {
		s.addrBytes = s.addrBytes[len(s.addrBytes)-2:]
	}
}

// ReadI2C reads a register value based on the current address pointer.
func (s *SC2336) ReadI2C() byte {
	if len(s.addrBytes) < 2 {
		return 0
	}

	addr := uint16(s.addrBytes[len(s.addrBytes)-2])<<8 | uint16(s.addrBytes[len(s.addrBytes)-1])
	switch addr {
	case 0x3107:
		return 0xCB
	case 0x3108:
		return 0x3A
	default:
		return 0
	}
}
