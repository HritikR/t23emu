package bus

import "github.com/HritikR/t23emu/internal/memory"

type Bus struct {
	ram *memory.RAM
}

func New(ram *memory.RAM) *Bus {
	return &Bus{
		ram: ram,
	}
}

func (b *Bus) Read8(addr uint32) byte {
	return b.ram.Read8(addr)
}

func (b *Bus) Write8(addr uint32, value byte) {
	b.ram.Write8(addr, value)
}

func (b *Bus) Read32(addr uint32) uint32 {
	return b.ram.Read32(addr)
}

func (b *Bus) Write32(addr uint32, value uint32) {
	b.ram.Write32(addr, value)
}
