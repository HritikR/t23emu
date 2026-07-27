package device

// Ingenic Ethernet MAC register offsets, relative to the GMAC physical base at
// 0x134B0000. Only the MDIO control path is modeled here because U-Boot polls
// it during PHY discovery.
const (
	GMAC_MDIO_ADDR uint32 = 0x10
)

const (
	GMAC_MDIO_BUSY uint32 = 1 << 0
)

// NewGMAC creates a minimal Ethernet MAC register block.
func NewGMAC() *RegisterBlock {
	gmac := NewRegisterBlock("GMAC", 0x10000)

	gmac.SetName(GMAC_MDIO_ADDR, "MDIO_ADDR")
	gmac.SetReadFunc(GMAC_MDIO_ADDR, func() uint32 {
		return gmac.regs[GMAC_MDIO_ADDR] &^ GMAC_MDIO_BUSY
	})

	return gmac
}
