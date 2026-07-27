package device

import "testing"

func TestDDRPPIRInitBitSelfClears(t *testing.T) {
	ddrp := NewDDRP()

	ddrp.Write32(DDRP_PIR, 0x00400001)

	if got := ddrp.Read32(DDRP_PIR); got&DDRP_PIR_INIT != 0 {
		t.Fatalf("expected PIR init bit to self-clear, got 0x%08X", got)
	}
}

func TestDDRPPIRPreservesOtherWrittenBits(t *testing.T) {
	ddrp := NewDDRP()

	ddrp.Write32(DDRP_PIR, 0x00400001)

	if got := ddrp.Read32(DDRP_PIR); got != 0x00400000 {
		t.Fatalf("expected PIR to preserve non-init bits, got 0x%08X", got)
	}
}
