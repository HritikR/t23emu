package device

const (
	GMAC_MDIO_ADDR    uint32 = 0x10
	GMAC_MDIO_DATA    uint32 = 0x14
	GMAC_DMA_BUS_MODE uint32 = 0x1000
)

const (
	GMAC_MDIO_BUSY      uint32 = 1 << 0
	GMAC_DMA_SOFT_RESET uint32 = 1 << 0
)

// NewGMAC creates a minimal Ethernet MAC register block.
func NewGMAC() *RegisterBlock {
	gmac := NewRegisterBlock("GMAC", 0x10000)

	gmac.SetName(GMAC_MDIO_ADDR, "MDIO_ADDR")
	gmac.SetReadFunc(GMAC_MDIO_ADDR, func() uint32 {
		return gmac.regs[GMAC_MDIO_ADDR] &^ GMAC_MDIO_BUSY
	})
	gmac.SetName(GMAC_MDIO_DATA, "MDIO_DATA")
	gmac.SetReadFunc(GMAC_MDIO_DATA, func() uint32 {
		return 0xffff
	})
	gmac.SetName(GMAC_DMA_BUS_MODE, "DMA_BUS_MODE")
	gmac.SetReadFunc(GMAC_DMA_BUS_MODE, func() uint32 {
		return gmac.regs[GMAC_DMA_BUS_MODE] &^ GMAC_DMA_SOFT_RESET
	})

	return gmac
}
