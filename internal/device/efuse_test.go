package device

import "testing"

func TestEFUSEProvidesT23NSOCInfo(t *testing.T) {
	efuse := NewEFUSE()

	if got := efuse.Read32(EFUSE_SOC_INFO); got != EFUSE_SOC_T23N {
		t.Fatalf("SOC_INFO = 0x%08x, want 0x%08x", got, EFUSE_SOC_T23N)
	}
}
