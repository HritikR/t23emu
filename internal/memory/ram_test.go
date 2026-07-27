package memory

import "testing"

func TestRAMReadWrite8(t *testing.T) {

	ram := NewRAM(1024)

	ram.Write8(100, 0xAA)

	value := ram.Read8(100)

	if value != 0xAA {
		t.Fatalf(
			"expected 0xAA, got 0x%02X",
			value,
		)
	}
}

func TestRAMReadWrite32(t *testing.T) {

	ram := NewRAM(1024)

	ram.Write32(
		100,
		0x12345678,
	)

	value := ram.Read32(100)

	if value != 0x12345678 {
		t.Fatalf(
			"expected 0x12345678, got 0x%08X",
			value,
		)
	}
}

func TestRAMSize(t *testing.T) {

	ram := NewRAM(64 * 1024 * 1024)

	if ram.Size() != 64*1024*1024 {
		t.Fatalf(
			"wrong RAM size",
		)
	}
}

func TestRAMOutOfBounds(t *testing.T) {

	defer func() {

		if recover() == nil {
			t.Fatalf(
				"expected panic",
			)
		}

	}()

	ram := NewRAM(1024)

	ram.Read32(2000)
}
