package device

import "testing"

func TestDWC2GRSTCTLSoftResetClearsAndAHBIdle(t *testing.T) {
	dwc2 := NewDWC2()

	dwc2.Write32(DWC2_GRSTCTL, DWC2_GRSTCTL_CSFTRST|0x20)

	got := dwc2.Read32(DWC2_GRSTCTL)
	if got&DWC2_GRSTCTL_CSFTRST != 0 {
		t.Fatalf("expected core soft reset bit to clear, got 0x%08X", got)
	}
	if got&DWC2_GRSTCTL_AHBIDLE == 0 {
		t.Fatalf("expected AHB idle bit to be set, got 0x%08X", got)
	}
	if got&0x20 == 0 {
		t.Fatalf("expected other GRSTCTL bits preserved, got 0x%08X", got)
	}
}

func TestDWC2GSNPSIDReports300A(t *testing.T) {
	dwc2 := NewDWC2()

	if got := dwc2.Read32(DWC2_GSNPSID); got != DWC2_GSNPSID_300A {
		t.Fatalf("expected DWC2 3.00a id, got 0x%08X", got)
	}
}
