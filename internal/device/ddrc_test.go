package device

import "testing"

func TestDDRCStatusReportsReady(t *testing.T) {
	ddrc := NewDDRC()

	if got := ddrc.Read32(DDRC_STATUS); got&DDRC_STATUS_READY == 0 {
		t.Fatalf("expected DDRC status ready bit set, got 0x%08X", got)
	}
}

func TestDDRCRegistersReadBackWrites(t *testing.T) {
	ddrc := NewDDRC()

	ddrc.Write32(DDRC_DLP, 0x18)

	if got := ddrc.Read32(DDRC_DLP); got != 0x18 {
		t.Fatalf("expected DLP to read back 0x18, got 0x%08X", got)
	}
}
