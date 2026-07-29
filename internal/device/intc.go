package device

import "fmt"

// Ingenic interrupt controller register offsets. The controller is split into
// banks of 32 IRQs; this model implements the first bank, which is enough for
// the TCU/OST interrupt used during early Linux boot.
const (
	INTC_ISR  uint32 = 0x00
	INTC_IMR  uint32 = 0x04
	INTC_IMSR uint32 = 0x08
	INTC_IMCR uint32 = 0x0C
	INTC_IPR  uint32 = 0x10
)

// INTC is a small interrupt controller model with mask and pending state.
type INTC struct {
	*RegisterBlock

	pending uint32
	mask    uint32
}

func NewINTC() *INTC {
	intc := &INTC{
		RegisterBlock: NewRegisterBlock("INTC", 0x1000),
		mask:          ^uint32(0),
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
	}

	return intc
}

func (i *INTC) Read32(addr uint32) uint32 {
	offset := addr &^ 3
	i.readCounts[offset]++

	var value uint32
	switch offset {
	case INTC_ISR:
		value = i.pending &^ i.mask
	case INTC_IPR:
		value = i.pending // REMOVED: i.pending &^= value (Reads must be side-effect free)
	case INTC_IMR:
		value = i.mask
	default:
		value = i.regs[offset] | i.readOnes[offset]
	}

	if i.Trace {
		fmt.Fprintf(i.Out, "  %s read  %s => 0x%08x\n", i.Name, i.RegName(offset), value)
	}

	return value
}

// Add Deassert to allow clearing pending interrupt lines
func (i *INTC) Deassert(irq uint8) {
	if irq < 32 {
		i.pending &^= (1 << irq)
	}
}

func (i *INTC) Write32(addr uint32, value uint32) {
	offset := addr &^ 3
	i.writeCounts[offset]++
	i.regs[offset] = value

	switch offset {
	case INTC_IMSR:
		i.mask |= value
	case INTC_IMCR:
		i.mask &^= value
	case INTC_IPR:
		i.pending &^= value
	}

	if i.Trace {
		fmt.Fprintf(i.Out, "  %s write %s <= 0x%08x\n", i.Name, i.RegName(offset), value)
	}
}

func (i *INTC) Assert(irq uint8) {
	if irq < 32 {
		i.pending |= 1 << irq
	}
}

func (i *INTC) Pending() uint32 {
	return i.pending &^ i.mask
}

func (i *INTC) RawPending() uint32 {
	return i.pending
}

var _ Device = (*INTC)(nil)
