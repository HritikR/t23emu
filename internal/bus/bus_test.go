package bus

import (
	"testing"

	"github.com/HritikR/t23emu/internal/memory"
)

func TestBusReadWrite32(t *testing.T) {

	ram := memory.NewRAM(1024)

	b := New(ram)

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

	ram := memory.NewRAM(1024)

	b := New(ram)

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
