package memory

import (
	"encoding/binary"
	"fmt"
)

type RAM struct {
	data []byte
	size uint32
}

func NewRAM(size uint32) *RAM {
	return &RAM{
		data: make([]byte, size),
		size: size,
	}
}

func (r *RAM) Size() uint32 {
	return r.size
}

func (r *RAM) Bytes() []byte {
	return r.data
}

func (r *RAM) Read8(addr uint32) byte {
	r.checkAddress(addr)

	return r.data[addr]
}

func (r *RAM) Write8(addr uint32, value byte) {
	r.checkAddress(addr)

	r.data[addr] = value
}

func (r *RAM) Read32(addr uint32) uint32 {
	r.checkAddress(addr + 3)

	return binary.LittleEndian.Uint32(
		r.data[addr : addr+4],
	)
}

func (r *RAM) Write32(addr uint32, value uint32) {
	r.checkAddress(addr + 3)

	binary.LittleEndian.PutUint32(
		r.data[addr:addr+4],
		value,
	)
}

func (r *RAM) checkAddress(addr uint32) {
	if addr >= r.size {
		panic(fmt.Sprintf(
			"RAM access out of range: 0x%08X",
			addr,
		))
	}
}
