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

func (b *Bus) find(addr uint32) *Mapping {

	for i := range b.mappings {

		m := &b.mappings[i]

		if addr >= m.Start &&
			addr <= m.End {

			return m
		}
	}

	return nil
}

func (b *Bus) Read8(addr uint32) byte {

	m := b.find(addr)

	if m == nil {
		panic("bus: unmapped read8")
	}

	return m.Device.Read8(addr - m.Start)
}

func (b *Bus) Write8(
	addr uint32,
	value byte,
) {

	m := b.find(addr)

	if m == nil {
		panic("bus: unmapped write8")
	}

	m.Device.Write8(
		addr - m.Start,
		value,
	)
}

func (b *Bus) Read32(addr uint32) uint32 {

	m := b.find(addr)

	if m == nil {
		panic("bus: unmapped read32")
	}

	return m.Device.Read32(addr - m.Start)
}

func (b *Bus) Write32(
	addr uint32,
	value uint32,
) {

	m := b.find(addr)

	if m == nil {
		panic("bus: unmapped write32")
	}

	m.Device.Write32(
		addr - m.Start,
		value,
	)
}

