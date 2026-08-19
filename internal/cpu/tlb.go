package cpu

type TLBEntry struct {
	PageMask uint32
	EntryHi  uint32
	EntryLo0 uint32
	EntryLo1 uint32
}

const (
	entryLoG    uint32 = 1 << 0
	entryLoV    uint32 = 1 << 1
	entryLoD    uint32 = 1 << 2
	entryLoPFN  uint32 = 0x03FFFFC0
	entryHiVPN  uint32 = 0xFFFFE000
	entryHiASID uint32 = 0x000000FF

	contextPTEBase uint32 = 0xFF800000
	contextBadVPN2 uint32 = 0x007FFFF0
)

// TranslateAddress applies the fixed unmapped kernel segments and the CP0 TLB.
func (c *CPU) TranslateAddress(addr uint32) (uint32, bool) {
	if addr < 0x80000000 && c.CP0[CP0_STATUS]&STATUS_ERL != 0 {
		return addr, true
	}
	if addr >= 0x80000000 && addr < 0xC0000000 {
		return addr & 0x1FFFFFFF, true
	}
	if addr >= 0xE0000000 {
		return addr, true
	}

	if phys, ok, _ := c.lookupTLB(addr, false); ok {
		return phys, true
	}
	return addr, false
}

func isTLBMappedSegment(addr uint32) bool {
	return addr < 0x80000000 || (addr >= 0xC0000000 && addr < 0xE0000000)
}

func (c *CPU) requiresTLB(addr uint32) bool {
	if addr < 0x80000000 && c.CP0[CP0_STATUS]&STATUS_ERL != 0 {
		return false
	}
	return isTLBMappedSegment(addr)
}

func (c *CPU) updateTLBExceptionState(badVAddr uint32) {
	c.CP0[CP0_BADVADDR] = badVAddr
	c.CP0[CP0_ENTRYHI] = (c.CP0[CP0_ENTRYHI] & entryHiASID) | (badVAddr & entryHiVPN)
	c.CP0[CP0_CONTEXT] = (c.CP0[CP0_CONTEXT] & contextPTEBase) | ((badVAddr >> 9) & contextBadVPN2)
}

func (c *CPU) lookupTLB(addr uint32, write bool) (uint32, bool, int) {
	asid := c.CP0[CP0_ENTRYHI] & entryHiASID
	for i := range c.TLB {
		entry := c.TLB[i]
		if !entry.matches(addr, asid) {
			continue
		}

		entryLo := entry.entryLo(addr)
		if entryLo&entryLoV == 0 {
			return 0, false, i
		}
		if write && entryLo&entryLoD == 0 {
			return 0, false, i
		}

		return ((entryLo & entryLoPFN) << 6) | entry.pageOffset(addr), true, i
	}
	return 0, false, -1
}

func (e TLBEntry) matches(addr uint32, asid uint32) bool {
	global := e.EntryLo0&entryLoG != 0 && e.EntryLo1&entryLoG != 0
	if !global && e.EntryHi&entryHiASID != asid {
		return false
	}

	mask := e.PageMask | 0x1FFF
	return (addr &^ mask) == (e.EntryHi & entryHiVPN &^ mask)
}

func (e TLBEntry) entryLo(addr uint32) uint32 {
	if addr&e.oddPageBit() == 0 {
		return e.EntryLo0
	}
	return e.EntryLo1
}

func (e TLBEntry) pageOffset(addr uint32) uint32 {
	return addr & ((e.PageMask >> 1) | 0xFFF)
}

func (e TLBEntry) oddPageBit() uint32 {
	return (e.PageMask >> 1) + 0x1000
}

func (c *CPU) writeIndexedTLB(index int) {
	if index < 0 || index >= len(c.TLB) {
		return
	}
	c.TLB[index] = TLBEntry{
		PageMask: c.CP0[CP0_PAGEMASK],
		EntryHi:  c.CP0[CP0_ENTRYHI],
		EntryLo0: c.CP0[CP0_ENTRYLO0],
		EntryLo1: c.CP0[CP0_ENTRYLO1],
	}
	if c.Dynarec != nil {
		c.Dynarec.Invalidate()
	}
}

func (c *CPU) readIndexedTLB(index int) {
	if index < 0 || index >= len(c.TLB) {
		return
	}
	entry := c.TLB[index]
	c.CP0[CP0_PAGEMASK] = entry.PageMask
	c.CP0[CP0_ENTRYHI] = entry.EntryHi
	c.CP0[CP0_ENTRYLO0] = entry.EntryLo0
	c.CP0[CP0_ENTRYLO1] = entry.EntryLo1
}

func (c *CPU) probeTLB() {
	_, _, index := c.lookupTLB(c.CP0[CP0_ENTRYHI], false)
	if index < 0 {
		c.CP0[CP0_INDEX] = 0x80000000
		return
	}
	c.CP0[CP0_INDEX] = uint32(index)
}

func (c *CPU) randomTLBIndex() int {
	random := int(c.CP0[CP0_RANDOM] & 31)
	wired := int(c.CP0[CP0_WIRED] & 31)
	if random < wired {
		random = 31
	}
	next := random - 1
	if next < wired {
		next = 31
	}
	c.CP0[CP0_RANDOM] = uint32(next)
	return random
}
