package device

import "testing"

func TestGMACMDIOBusySelfClears(t *testing.T) {
	gmac := NewGMAC()

	gmac.Write32(GMAC_MDIO_ADDR, 0x81)

	if got := gmac.Read32(GMAC_MDIO_ADDR); got&GMAC_MDIO_BUSY != 0 {
		t.Fatalf("expected MDIO busy bit to clear, got 0x%08X", got)
	}
	if got := gmac.Read32(GMAC_MDIO_ADDR); got != 0x80 {
		t.Fatalf("expected other MDIO address bits preserved, got 0x%08X", got)
	}
}

func TestGMACMDIODataReadsNoPHY(t *testing.T) {
	gmac := NewGMAC()

	gmac.Write32(GMAC_MDIO_DATA, 0)

	if got := gmac.Read32(GMAC_MDIO_DATA); got != 0xffff {
		t.Fatalf("expected MDIO data to report no PHY, got 0x%08X", got)
	}
}

func TestGMACDMASoftResetSelfClears(t *testing.T) {
	gmac := NewGMAC()

	gmac.Write32(GMAC_DMA_BUS_MODE, 0x201)

	if got := gmac.Read32(GMAC_DMA_BUS_MODE); got&GMAC_DMA_SOFT_RESET != 0 {
		t.Fatalf("expected DMA soft reset bit to clear, got 0x%08X", got)
	}
	if got := gmac.Read32(GMAC_DMA_BUS_MODE); got != 0x200 {
		t.Fatalf("expected other DMA bus mode bits preserved, got 0x%08X", got)
	}
}
