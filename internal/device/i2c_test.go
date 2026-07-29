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

func TestI2CStatusReportsTransmitReady(t *testing.T) {
	i2c := NewI2C("I2C0")

	got := i2c.Read32(I2C_STA)
	if got&i2cStatusTXNotFull == 0 {
		t.Fatalf("expected TX-not-full status bit, got 0x%08X", got)
	}
	if got&i2cStatusTXEmpty == 0 {
		t.Fatalf("expected TX-empty status bit, got 0x%08X", got)
	}
	if got&i2cStatusRXNotEmpty != 0 {
		t.Fatalf("expected RX FIFO empty before read command, got 0x%08X", got)
	}
	if got := i2c.Read32(I2C_TXFLR); got != 0 {
		t.Fatalf("expected empty TX FIFO level, got %d", got)
	}
}

func TestI2CSensorIDReadSequence(t *testing.T) {
	i2c := NewI2C("I2C0")

	i2c.Write32(I2C_DC, 0x31)
	i2c.Write32(I2C_DC, 0x07)
	i2c.Write32(I2C_DC, i2cDataCmdRead)

	if got := i2c.Read32(I2C_RXFLR); got != 1 {
		t.Fatalf("expected one byte in RX FIFO, got %d", got)
	}
	if got := i2c.Read32(I2C_STA); got&i2cStatusRXNotEmpty == 0 {
		t.Fatalf("expected RX-not-empty status bit, got 0x%08X", got)
	}
	if got := i2c.Read32(I2C_DC); got != 0xCB {
		t.Fatalf("expected SC2336 high ID 0xCB, got 0x%02X", got)
	}
	if got := i2c.Read32(I2C_RXFLR); got != 0 {
		t.Fatalf("expected RX FIFO to drain, got %d", got)
	}

	i2c.Write32(I2C_DC, 0x31)
	i2c.Write32(I2C_DC, 0x08)
	i2c.Write32(I2C_DC, i2cDataCmdRead)

	if got := i2c.Read32(I2C_DC); got != 0x3A {
		t.Fatalf("expected SC2336 low ID 0x3A, got 0x%02X", got)
	}
}

func TestI2CTransmitInterruptFollowsMask(t *testing.T) {
	i2c := NewI2C("I2C0")
	var asserted bool
	i2c.Interrupt = func(assert bool) {
		asserted = assert
	}

	i2c.Write32(I2C_DC, 0x31)
	if asserted {
		t.Fatalf("expected masked TX interrupt to stay deasserted")
	}
	if got := i2c.Read32(I2C_RAWST); got&i2cIntTXEmpty == 0 {
		t.Fatalf("expected raw TX-empty pending bit, got 0x%08X", got)
	}

	i2c.Write32(I2C_INTM, i2cIntTXEmpty)
	if !asserted {
		t.Fatalf("expected unmasked TX-empty interrupt to assert")
	}
	if got := i2c.Read32(I2C_INTST); got != i2cIntTXEmpty {
		t.Fatalf("expected masked interrupt status 0x%08X, got 0x%08X", i2cIntTXEmpty, got)
	}

	i2c.Write32(I2C_INTM, 0)
	if asserted {
		t.Fatalf("expected masking interrupt to deassert line")
	}
}
