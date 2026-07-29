package device

import "fmt"

// Ingenic interrupt controller register offsets. The controller is split into
// banks of 32 IRQs. Bank 0 starts at 0x00, bank 1 at 0x20.
const (
	INTC_ISR  uint32 = 0x00
	INTC_IMR  uint32 = 0x04
	INTC_IMSR uint32 = 0x08
	INTC_IMCR uint32 = 0x0C
	INTC_IPR  uint32 = 0x10

	intcBankStride uint32 = 0x20
	intcBanks             = 2
)

// INTC is a small interrupt controller model with mask and pending state.
type INTC struct {
	*RegisterBlock

	pending [intcBanks]uint32
	mask    [intcBanks]uint32
}

func NewINTC() *INTC {
	intc := &INTC{
		RegisterBlock: NewRegisterBlock("INTC", 0x1000),
	}
	for bank := range intc.mask {
		intc.mask[bank] = ^uint32(0)
	}

	names := map[uint32]string{
		INTC_ISR:  "ISR",
		INTC_IMR:  "IMR",
		INTC_IMSR: "IMSR",
		INTC_IMCR: "IMCR",
		INTC_IPR:  "IPR",
	}
	for offset, name := range names {
		intc.SetName(offset, name)
		intc.SetName(offset+intcBankStride, name+"1")
	}

	return intc
}

func (i *INTC) Read32(addr uint32) uint32 {
	offset := addr &^ 3
	i.readCounts[offset]++
	bank, reg := intcBank(offset)

	var value uint32
	if bank < 0 {
		value = i.regs[offset] | i.readOnes[offset]
	} else {
		switch reg {
		case INTC_ISR:
			value = i.pending[bank] &^ i.mask[bank]
		case INTC_IPR:
			value = i.pending[bank]
		case INTC_IMR:
			value = i.mask[bank]
		default:
			value = i.regs[offset] | i.readOnes[offset]
		}
	}

	if i.Trace {
		fmt.Fprintf(i.Out, "  %s read  %s => 0x%08x\n", i.Name, i.RegName(offset), value)
	}

	return value
}

// Add Deassert to allow clearing pending interrupt lines
func (i *INTC) Deassert(irq uint8) {
	bank, bit, ok := irqBankBit(irq)
	if ok {
		i.pending[bank] &^= bit
	}
}

func (i *INTC) Write32(addr uint32, value uint32) {
	offset := addr &^ 3
	i.writeCounts[offset]++
	i.regs[offset] = value
	bank, reg := intcBank(offset)

	if bank >= 0 {
		switch reg {
		case INTC_IMSR:
			i.mask[bank] |= value
		case INTC_IMCR:
			i.mask[bank] &^= value
		case INTC_IPR:
			i.pending[bank] &^= value
		}
	}

	if i.Trace {
		fmt.Fprintf(i.Out, "  %s write %s <= 0x%08x\n", i.Name, i.RegName(offset), value)
	}
}

func (i *INTC) Assert(irq uint8) {
	bank, bit, ok := irqBankBit(irq)
	if ok {
		i.pending[bank] |= bit
	}
}

func (i *INTC) Pending() uint32 {
	var pending uint32
	for bank := range i.pending {
		if i.pending[bank]&^i.mask[bank] != 0 {
			pending = 1
			break
		}
	}
	return pending
}

func (i *INTC) RawPending() uint32 {
	var pending uint32
	for bank := range i.pending {
		pending |= i.pending[bank]
	}
	return pending
}

func intcBank(offset uint32) (int, uint32) {
	bank := int(offset / intcBankStride)
	if bank >= intcBanks {
		return -1, offset
	}
	return bank, offset % intcBankStride
}

func irqBankBit(irq uint8) (int, uint32, bool) {
	bank := int(irq / 32)
	if bank >= intcBanks {
		return 0, 0, false
	}
	return bank, 1 << (irq % 32), true
}

var _ Device = (*INTC)(nil)
