package cpu

import (
	"testing"
)

func TestRegisterWriteRead(t *testing.T) {

	cpu := createTestCPU()

	cpu.WriteRegister(
		1,
		0x12345678,
	)

	value := cpu.ReadRegister(1)

	if value != 0x12345678 {

		t.Fatalf(
			"expected 0x12345678 got 0x%08X",
			value,
		)
	}
}

func TestZeroRegisterAlwaysZero(t *testing.T) {

	cpu := createTestCPU()

	cpu.WriteRegister(
		0,
		0xFFFFFFFF,
	)

	value := cpu.ReadRegister(0)

	if value != 0 {

		t.Fatalf(
			"$zero register changed: 0x%08X",
			value,
		)
	}
}

func TestZeroRegisterUnderlyingValue(t *testing.T) {

	cpu := createTestCPU()

	// Even if the internal array changes,
	// reads should still return zero.

	cpu.Regs[0] = 12345

	value := cpu.ReadRegister(0)

	if value != 0 {

		t.Fatalf(
			"expected zero register to return 0",
		)
	}
}
