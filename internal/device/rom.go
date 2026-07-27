package device

type ROM struct {
	data []byte
}

func NewROM(data []byte) *ROM {

	romData := make([]byte, len(data))

	copy(
		romData,
		data,
	)

	return &ROM{
		data: romData,
	}
}

func (r *ROM) Read8(addr uint32) byte {

	if addr >= uint32(len(r.data)) {
		panic("ROM read out of bounds")
	}

	return r.data[addr]
}

func (r *ROM) Write8(addr uint32, value byte) {

	// ignore writes
}

func (r *ROM) Read32(addr uint32) uint32 {

	if addr+3 >= uint32(len(r.data)) {
		panic("ROM read32 out of bounds")
	}

	return uint32(r.data[addr]) |
		uint32(r.data[addr+1])<<8 |
		uint32(r.data[addr+2])<<16 |
		uint32(r.data[addr+3])<<24
}

func (r *ROM) Write32(addr uint32, value uint32) {

	// ignore writes
}

var _ Device = (*ROM)(nil)
