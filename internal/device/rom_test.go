package device

import "testing"

func TestROMRead32(t *testing.T) {

	data := []byte{
		0x78,
		0x56,
		0x34,
		0x12,
	}

	rom := NewROM(data)

	value := rom.Read32(0)

	if value != 0x12345678 {

		t.Fatalf(
			"expected 0x12345678 got 0x%08X",
			value,
		)
	}
}

func TestROMWriteIgnored(t *testing.T) {

	data := []byte{
		0xAA,
		0xBB,
		0xCC,
		0xDD,
	}

	rom := NewROM(data)

	rom.Write32(
		0,
		0x12345678,
	)

	value := rom.Read32(0)

	if value != 0xDDCCBBAA {

		t.Fatalf(
			"ROM contents changed",
		)
	}
}
