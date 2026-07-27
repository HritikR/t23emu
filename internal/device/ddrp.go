package device

// Ingenic DDR PHY register offsets, relative to the DDR PHY physical
// base at 0x134f0000.
const (
	DDRP_PGCR uint32 = 0x004
	DDRP_PGSR uint32 = 0x008
	DDRP_PIR  uint32 = 0x00C
)

const (
	// DDRP_PIR_INIT is a command/start bit. The SPL writes it to launch a
	// PHY initialization command and then polls until hardware clears it.
	DDRP_PIR_INIT uint32 = 1 << 0
)

// NewDDRP creates the DDR PHY register block.
func NewDDRP() *RegisterBlock {
	ddrp := NewRegisterBlock("DDRP", 0x10000)

	names := map[uint32]string{
		DDRP_PGCR: "PGCR",
		DDRP_PGSR: "PGSR",
		DDRP_PIR:  "PIR",
	}
	for offset, name := range names {
		ddrp.SetName(offset, name)
	}

	ddrp.SetReadFunc(DDRP_PIR, func() uint32 {
		return ddrp.regs[DDRP_PIR] &^ DDRP_PIR_INIT
	})

	return ddrp
}
