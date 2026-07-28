package device

import "testing"

// The driver writes ENB and then waits for the controller to acknowledge
// in ENSTA. An ENSTA stuck at zero costs the boot five seconds of jiffy
// timeout and then "enable i2c0 failed".
func TestI2CEnableStatusMirrorsEnable(t *testing.T) {
	i2c := NewI2C("I2C0")

	if got := i2c.Read32(I2C_ENSTA); got != 0 {
		t.Fatalf("expected ENSTA clear before enable, got 0x%08X", got)
	}

	i2c.Write32(I2C_ENB, I2C_ENB_ENABLE)
	if got := i2c.Read32(I2C_ENSTA); got != I2C_ENB_ENABLE {
		t.Fatalf("expected ENSTA to acknowledge enable, got 0x%08X", got)
	}

	i2c.Write32(I2C_ENB, 0)
	if got := i2c.Read32(I2C_ENSTA); got != 0 {
		t.Fatalf("expected ENSTA clear after disable, got 0x%08X", got)
	}
}

// Only the enable bit is meaningful; the driver must not see the rest of
// what it wrote to ENB reflected back as controller state.
func TestI2CEnableStatusReportsOnlyEnableBit(t *testing.T) {
	i2c := NewI2C("I2C0")

	i2c.Write32(I2C_ENB, 0xFFFFFFFF)
	if got := i2c.Read32(I2C_ENSTA); got != I2C_ENB_ENABLE {
		t.Fatalf("expected only the enable bit in ENSTA, got 0x%08X", got)
	}
}

// The timing registers are plain storage, and the kernel reads back what
// it programmed when it prints "set:249 hold:250".
func TestI2CTimingRegistersReadBack(t *testing.T) {
	i2c := NewI2C("I2C0")

	i2c.Write32(I2C_SDASU, 249)
	i2c.Write32(I2C_SDAHD, 250)

	if got := i2c.Read32(I2C_SDASU); got != 249 {
		t.Fatalf("expected SDASU 249, got %d", got)
	}
	if got := i2c.Read32(I2C_SDAHD); got != 250 {
		t.Fatalf("expected SDAHD 250, got %d", got)
	}
}
