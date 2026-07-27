package bus

import (
	"testing"

	"github.com/HritikR/t23emu/internal/memory"
)

func createTestBus() *Bus {

	ram := memory.NewRAM(1024)

	b := New()

	b.Map(
		0x00000000,
		0x000003FF,
		ram,
	)

	return b
}

func TestBusReadWrite32(t *testing.T) {

	b := createTestBus()

	b.Write32(
		0x100,
		0x12345678,
	)

	value := b.Read32(0x100)

	if value != 0x12345678 {

		t.Fatalf(
			"expected 0x12345678, got 0x%08X",
			value,
		)
	}
}

func TestBusReadWrite8(t *testing.T) {

	b := createTestBus()

	b.Write8(
		0x20,
		0xAA,
	)

	value := b.Read8(0x20)

	if value != 0xAA {

		t.Fatalf(
			"expected 0xAA, got 0x%02X",
			value,
		)
	}
}

func TestBusUnmappedRead(t *testing.T) {

	b := New()

	defer func() {

		if recover() == nil {

			t.Fatalf(
				"expected panic for unmapped read",
			)
		}

	}()

	b.Read32(
		0x1000,
	)
}

func TestBusMultipleMappings(t *testing.T) {

	ram1 := memory.NewRAM(256)

	ram2 := memory.NewRAM(256)

	b := New()

	b.Map(
		0x00000000,
		0x000000FF,
		ram1,
	)

	b.Map(
		0x10000000,
		0x100000FF,
		ram2,
	)

	b.Write32(
		0x00000000,
		0x11111111,
	)

	b.Write32(
		0x10000000,
		0x22222222,
	)

	if b.Read32(0x00000000) != 0x11111111 {

		t.Fatalf(
			"first mapping failed",
		)
	}

	if b.Read32(0x10000000) != 0x22222222 {

		t.Fatalf(
			"second mapping failed",
		)
	}
}
