package bus

import "github.com/HritikR/t23emu/internal/device"

type Bus struct {
	mappings []Mapping
}

func New() *Bus {

	return &Bus{
		mappings: make([]Mapping, 0),
	}
}

func (b *Bus) Map(
	start uint32,
	end uint32,
	dev device.Device,
) {

	b.mappings = append(
		b.mappings,
		Mapping{
			Start:  start,
			End:    end,
			Device: dev,
		},
	)
}

func (b *Bus) translate(addr uint32) uint32 {
	// kseg0 (0x80000000 - 0x9FFFFFFF) and kseg1 (0xA0000000 - 0xBFFFFFFF)
	// map directly to physical memory starting at 0x00000000 (clear top 3 bits)
	if addr >= 0x80000000 && addr < 0xC0000000 {
		return addr & 0x1FFFFFFF
	}
	return addr
}

func (b *Bus) find(addr uint32) *Mapping {
	phys := b.translate(addr)
	for i := range b.mappings {
		m := &b.mappings[i]
		if phys >= m.Start && phys <= m.End {
			return m
		}
	}
	return nil
}

// HasMapping returns true if the address has a mapped device.
func (b *Bus) HasMapping(addr uint32) bool {
	return b.find(addr) != nil
}

func (b *Bus) Read8(addr uint32) byte {
	m := b.find(addr)
	if m == nil {
		panic("bus: unmapped read8")
	}
	phys := b.translate(addr)
	return m.Device.Read8(phys - m.Start)
}

func (b *Bus) Write8(
	addr uint32,
	value byte,
) {
	m := b.find(addr)
	if m == nil {
		panic("bus: unmapped write8")
	}
	phys := b.translate(addr)
	m.Device.Write8(
		phys - m.Start,
		value,
	)
}

func (b *Bus) Read32(addr uint32) uint32 {
	m := b.find(addr)
	if m == nil {
		panic("bus: unmapped read32")
	}
	phys := b.translate(addr)
	return m.Device.Read32(phys - m.Start)
}

func (b *Bus) Write32(
	addr uint32,
	value uint32,
) {
	m := b.find(addr)
	if m == nil {
		panic("bus: unmapped write32")
	}
	phys := b.translate(addr)
	m.Device.Write32(
		phys - m.Start,
		value,
	)
}

// Read16 reads a halfword. The device interface only exposes byte and
// word access, so halfwords are assembled from two little-endian byte
// accesses. This keeps every device halfword-addressable for free.
func (b *Bus) Read16(addr uint32) uint16 {
	lo := uint16(b.Read8(addr))
	hi := uint16(b.Read8(addr + 1))
	return lo | hi<<8
}

// Write16 writes a halfword as two little-endian byte accesses.
func (b *Bus) Write16(
	addr uint32,
	value uint16,
) {
	b.Write8(addr, byte(value))
	b.Write8(addr+1, byte(value>>8))
}

