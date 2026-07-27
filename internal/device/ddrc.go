package device

// Ingenic DDR controller register offsets, relative to the DDRC
// physical base at 0x13010000.
const (
	DDRC_AUTOSR_EN   uint32 = 0x1080
	DDRC_DLP         uint32 = 0x1084
	DDRC_REMAP       uint32 = 0x1088
	DDRC_STATUS      uint32 = 0x10C8
	DDRC_STATE       uint32 = 0x10CC
	DDRC_INIT_STATUS uint32 = 0x10D0
	DDRC_PHY_STATUS  uint32 = 0x208C
)

const (
	// DDRC_STATUS_READY is the status bit the SPL waits on after
	// programming the DDR PHY/controller timing registers. The emulator
	// does not model analog DDR training, so readiness is immediate.
	DDRC_STATUS_READY uint32 = 1 << 3

	// DDRC_INIT_DONE is a nonzero controller initialization status value.
	DDRC_INIT_DONE uint32 = 1 << 0

	// DDRC_STATE_READY is the post-initialization state value the SPL
	// waits for before continuing DDR setup.
	DDRC_STATE_READY uint32 = 0x3

	// DDRC_PHY_STATUS_READY is the DDR PHY training/initialization
	// completion state polled by the SPL after enabling the PHY. This SPL
	// waits for bit 2 and then bit 1 from the same status register.
	DDRC_PHY_STATUS_READY uint32 = 1<<2 | 1<<1
)

// NewDDRC creates the DDR controller/PHY register block.
func NewDDRC() *RegisterBlock {
	ddrc := NewRegisterBlock("DDRC", 0x10000)

	names := map[uint32]string{
		DDRC_AUTOSR_EN:   "AUTOSR_EN",
		DDRC_DLP:         "DLP",
		DDRC_REMAP:       "REMAP",
		DDRC_STATUS:      "STATUS",
		DDRC_STATE:       "STATE",
		DDRC_INIT_STATUS: "INIT_STATUS",
		DDRC_PHY_STATUS:  "PHY_STATUS",
	}
	for offset, name := range names {
		ddrc.SetName(offset, name)
	}

	ddrc.SetReadOnes(DDRC_STATUS, DDRC_STATUS_READY)
	ddrc.SetReadOnes(DDRC_STATE, DDRC_STATE_READY)
	ddrc.SetReadOnes(DDRC_INIT_STATUS, DDRC_INIT_DONE)
	ddrc.SetReadOnes(DDRC_PHY_STATUS, DDRC_PHY_STATUS_READY)

	return ddrc
}
