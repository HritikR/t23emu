package device

// OST register offsets for the timer block Linux maps at 0x12000000.
const (
	OST_CTRL  uint32 = 0x00
	OST_CNTH  uint32 = 0x08
	OST_CNTL  uint32 = 0x0C
	OST_CMPH  uint32 = 0x10
	OST_CMPL  uint32 = 0x14
	OST_FLAG  uint32 = 0x34
	OST_CLEAR uint32 = 0x38
)

// NewOST creates the OS timer block used by the Linux clocksource path.
func NewOST(ticks func() uint64) *RegisterBlock {
	ost := NewRegisterBlock("OST", 0x1000)

	names := map[uint32]string{
		OST_CTRL:  "CTRL",
		OST_CNTH:  "CNTH",
		OST_CNTL:  "CNTL",
		OST_CMPH:  "CMPH",
		OST_CMPL:  "CMPL",
		OST_FLAG:  "FLAG",
		OST_CLEAR: "CLEAR",
	}
	for offset, name := range names {
		ost.SetName(offset, name)
	}

	var hiBuf uint32
	ost.SetReadFunc(OST_CNTL, func() uint32 {
		now := ticks()
		hiBuf = uint32(now >> 32)
		return uint32(now)
	})
	ost.SetReadFunc(OST_CNTH, func() uint32 {
		return uint32(ticks() >> 32)
	})
	ost.SetReadFunc(OST_CMPH, func() uint32 {
		return hiBuf
	})

	return ost
}
