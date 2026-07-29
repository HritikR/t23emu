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

func TestDWC2HardwareConfigReportsEndpointsAndFIFO(t *testing.T) {
	dwc2 := NewDWC2()

	if got := dwc2.Read32(DWC2_GHWCFG2); got != DWC2_GHWCFG2_VALUE {
		t.Fatalf("expected GHWCFG2 reset value, got 0x%08X", got)
	}
	if got := dwc2.Read32(DWC2_GHWCFG3); got != DWC2_GHWCFG3_VALUE {
		t.Fatalf("expected GHWCFG3 reset value, got 0x%08X", got)
	}
	if got := dwc2.Read32(DWC2_GHWCFG4); got != DWC2_GHWCFG4_VALUE {
		t.Fatalf("expected GHWCFG4 reset value, got 0x%08X", got)
	}
	if got := dwc2.Read32(DWC2_GRXFSIZ); got != DWC2_GRXFSIZ_VALUE {
		t.Fatalf("expected GRXFSIZ reset value, got 0x%08X", got)
	}
}
