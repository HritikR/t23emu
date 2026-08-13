package bus

import (
	"encoding/binary"

	"github.com/HritikR/t23emu/internal/device"
)

type RAMProvider interface {
	Bytes() []byte
}

type Bus struct {
	mappings  []Mapping
	translate func(uint32) (uint32, bool)
	ram       []byte
	ramStart  uint32
	ramEnd    uint32
	hasRAM    bool
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

	// Fast path: if mapping a device that exposes raw RAM slice at physical address 0x0
	if start == 0 {
		if provider, ok := dev.(RAMProvider); ok {
			b.ram = provider.Bytes()
			b.ramStart = start
			b.ramEnd = end
			b.hasRAM = true
		}
	}
}

func (b *Bus) SetTranslator(translate func(uint32) (uint32, bool)) {
	b.translate = translate
}

func fixedTranslate(addr uint32) (uint32, bool) {
	// kseg0 (0x80000000 - 0x9FFFFFFF) and kseg1 (0xA0000000 - 0xBFFFFFFF)
	// map directly to physical memory starting at 0x00000000 (clear top 3 bits)
	if addr >= 0x80000000 && addr < 0xC0000000 {
		return addr & 0x1FFFFFFF, true
	}
	return addr, true
}

func (b *Bus) translateAddr(addr uint32) (uint32, bool) {
	if b.translate != nil {
		return b.translate(addr)
	}
	return fixedTranslate(addr)
}

func (b *Bus) findSlow(phys uint32) *Mapping {
	for i := range b.mappings {
		m := &b.mappings[i]
		if phys >= m.Start && phys <= m.End {
			return m
		}
	}
	return nil
}

func (b *Bus) find(addr uint32) *Mapping {
	phys, ok := b.translateAddr(addr)
	if !ok {
		return nil
	}
	return b.findSlow(phys)
}

// HasMapping returns true if the address has a mapped device.
func (b *Bus) HasMapping(addr uint32) bool {
	phys, ok := b.translateAddr(addr)
	if !ok {
		return false
	}
	if b.hasRAM && phys >= b.ramStart && phys <= b.ramEnd {
		return true
	}
	return b.findSlow(phys) != nil
}

func (b *Bus) Read8(addr uint32) byte {
	phys, ok := b.translateAddr(addr)
	if ok && b.hasRAM && phys >= b.ramStart && phys <= b.ramEnd {
		return b.ram[phys-b.ramStart]
	}
	m := b.findSlow(phys)
	if m == nil {
		panic("bus: unmapped read8")
	}
	return m.Device.Read8(phys - m.Start)
}

func (b *Bus) Write8(
	addr uint32,
	value byte,
) {
	phys, ok := b.translateAddr(addr)
	if ok && b.hasRAM && phys >= b.ramStart && phys <= b.ramEnd {
		b.ram[phys-b.ramStart] = value
		return
	}
	m := b.findSlow(phys)
	if m == nil {
		panic("bus: unmapped write8")
	}
	m.Device.Write8(
		phys-m.Start,
		value,
	)
}

func (b *Bus) Read32(addr uint32) uint32 {
	phys, ok := b.translateAddr(addr)
	if ok && b.hasRAM && phys >= b.ramStart && phys+3 <= b.ramEnd {
		offset := phys - b.ramStart
		return binary.LittleEndian.Uint32(b.ram[offset : offset+4])
	}
	m := b.findSlow(phys)
	if m == nil {
		panic("bus: unmapped read32")
	}
	return m.Device.Read32(phys - m.Start)
}

func (b *Bus) Write32(
	addr uint32,
	value uint32,
) {
	phys, ok := b.translateAddr(addr)
	if ok && b.hasRAM && phys >= b.ramStart && phys+3 <= b.ramEnd {
		offset := phys - b.ramStart
		binary.LittleEndian.PutUint32(
			b.ram[offset:offset+4],
			value,
		)
		return
	}
	m := b.findSlow(phys)
	if m == nil {
		panic("bus: unmapped write32")
	}
	m.Device.Write32(
		phys-m.Start,
		value,
	)
}

// Read16 reads a halfword.
func (b *Bus) Read16(addr uint32) uint16 {
	phys, ok := b.translateAddr(addr)
	if ok && b.hasRAM && phys >= b.ramStart && phys+1 <= b.ramEnd {
		offset := phys - b.ramStart
		return binary.LittleEndian.Uint16(b.ram[offset : offset+2])
	}
	lo := uint16(b.Read8(addr))
	hi := uint16(b.Read8(addr + 1))
	return lo | hi<<8
}

// Write16 writes a halfword.
func (b *Bus) Write16(
	addr uint32,
	value uint16,
) {
	phys, ok := b.translateAddr(addr)
	if ok && b.hasRAM && phys >= b.ramStart && phys+1 <= b.ramEnd {
		offset := phys - b.ramStart
		binary.LittleEndian.PutUint16(
			b.ram[offset:offset+2],
			value,
		)
		return
	}
	b.Write8(addr, byte(value))
	b.Write8(addr+1, byte(value>>8))
}
