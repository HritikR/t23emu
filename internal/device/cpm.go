package device

// Ingenic Clock and Power Management register offsets, relative to the
// CPM physical base at 0x10000000.
const (
	CPM_CPCCR  uint32 = 0x00 // Clock control
	CPM_CPCSR  uint32 = 0x04 // Clock status: divider change in progress
	CPM_DDRCDR uint32 = 0x0C // DDR clock divider
	CPM_CPAPCR uint32 = 0x10 // APLL control
	CPM_CPMPCR uint32 = 0x14 // MPLL control
	CPM_CPVPCR uint32 = 0x18 // VPLL control
	CPM_CPEPCR uint32 = 0x1C // EPLL control
	CPM_I2SCDR uint32 = 0x24
	CPM_LPCDR  uint32 = 0x2C
	CPM_MSC0CDR uint32 = 0x30
	CPM_SSICDR uint32 = 0x40
	CPM_CIMCDR uint32 = 0x54
	CPM_LCR    uint32 = 0xA0 // Low power control
	CPM_CLKGR  uint32 = 0xB0 // Clock gate
	CPM_OPCR   uint32 = 0xB8 // Oscillator and power control
	CPM_SRBC   uint32 = 0xC4 // Soft reset and bus control
	CPM_SLBC   uint32 = 0xC8
	CPM_SLPC   uint32 = 0xCC
)

// PLL control register bits. The M/N/OD dividers occupy the high bits
// and are software owned; the low bits are the enable and status flags.
const (
	// PLL_ENABLE is set by software to start the PLL.
	PLL_ENABLE uint32 = 1 << 0

	// PLL_BYPASS routes the input clock straight through.
	PLL_BYPASS uint32 = 1 << 1

	// PLL_ON is the read-only flag reporting that the PLL is running and
	// stable. Boot firmware spins on it after programming the dividers.
	//
	// The bit position comes from the firmware itself: after writing
	// CPAPCR the SPL loops on `lw $v0, 0x10($v1); andi $v0, $v0, 0x8`,
	// so the flag this part reports lock through is bit 3.
	PLL_ON uint32 = 1 << 3
)

// pllRegisters lists the control registers that carry lock status.
var pllRegisters = []struct {
	offset uint32
	name   string
}{
	{CPM_CPAPCR, "CPAPCR"},
	{CPM_CPMPCR, "CPMPCR"},
	{CPM_CPVPCR, "CPVPCR"},
	{CPM_CPEPCR, "CPEPCR"},
}

// NewCPM creates the Clock and Power Management block.
//
// Every PLL reports itself as on and locked. Real silicon takes hundreds
// of microseconds to lock and firmware polls for it; since the emulator
// has no notion of PLL settling time, reporting an immediate lock is both
// correct and what lets the poll loop terminate.
func NewCPM() *RegisterBlock {
	cpm := NewRegisterBlock("CPM", 0x1000)

	names := map[uint32]string{
		CPM_CPCCR:   "CPCCR",
		CPM_CPCSR:   "CPCSR",
		CPM_DDRCDR:  "DDRCDR",
		CPM_CPAPCR:  "CPAPCR",
		CPM_CPMPCR:  "CPMPCR",
		CPM_CPVPCR:  "CPVPCR",
		CPM_CPEPCR:  "CPEPCR",
		CPM_I2SCDR:  "I2SCDR",
		CPM_LPCDR:   "LPCDR",
		CPM_MSC0CDR: "MSC0CDR",
		CPM_SSICDR:  "SSICDR",
		CPM_CIMCDR:  "CIMCDR",
		CPM_LCR:     "LCR",
		CPM_CLKGR:   "CLKGR",
		CPM_OPCR:    "OPCR",
		CPM_SRBC:    "SRBC",
		CPM_SLBC:    "SLBC",
		CPM_SLPC:    "SLPC",
	}
	for offset, name := range names {
		cpm.SetName(offset, name)
	}

	for _, pll := range pllRegisters {
		cpm.SetReadOnes(pll.offset, PLL_ON)
	}

	// CPCSR reports clock divider changes still in flight. Divider writes
	// complete instantly here, so it must read as idle or firmware waits
	// forever for a change that already happened.
	cpm.SetInitial(CPM_CPCSR, 0)

	// The APLL is already running at reset on this part, so firmware that
	// only reads the dividers without programming them still sees a sane
	// clock tree. M=0x60, N=1, OD=1 corresponds to a 24 MHz input.
	cpm.SetInitial(CPM_CPAPCR, 0x60<<24|1<<18|1<<16|PLL_ENABLE)

	return cpm
}
