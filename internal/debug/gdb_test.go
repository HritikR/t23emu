// Copyright (c) 2026 Hritik R
package debug

import (
	"testing"

	"github.com/HritikR/t23emu/internal/bus"
	"github.com/HritikR/t23emu/internal/cpu"
	"github.com/HritikR/t23emu/internal/memory"
)

func TestChecksum(t *testing.T) {
	cs := CalculateChecksum("qSupported")
	if cs != 0x37 {
		t.Errorf("Expected checksum 0x37 for qSupported, got 0x%02x", cs)
	}
}

func TestHex32LEEncoding(t *testing.T) {
	val := uint32(0x80001800)
	encoded := encodeHex32LE(val)
	if encoded != "00180080" {
		t.Fatalf("Expected 00180080, got %s", encoded)
	}

	decoded, err := decodeHex32LE(encoded)
	if err != nil {
		t.Fatalf("Unexpected error decoding: %v", err)
	}
	if decoded != val {
		t.Fatalf("Expected 0x%08x, got 0x%08x", val, decoded)
	}
}

func TestRegisterEncodingDecoding(t *testing.T) {
	b := bus.New()
	ram := memory.NewRAM(1024 * 1024)
	b.Map(0x00000000, 1024*1024-1, ram)
	c := cpu.New(b)

	c.WriteRegister(1, 0x12345678)
	c.PC = 0x80001800

	s := NewServer(":1234", c)
	hexRegs := s.encodeRegisters()

	// Modify register
	c.WriteRegister(1, 0x0)
	s.decodeRegisters(hexRegs)

	if c.ReadRegister(1) != 0x12345678 {
		t.Errorf("Expected R1 to be restored to 0x12345678, got 0x%08x", c.ReadRegister(1))
	}
	if c.PC != 0x80001800 {
		t.Errorf("Expected PC to be 0x80001800, got 0x%08x", c.PC)
	}
}

func TestWatchpointTrigger(t *testing.T) {
	b := bus.New()
	ram := memory.NewRAM(1024 * 1024)
	b.Map(0x00000000, 1024*1024-1, ram)
	c := cpu.New(b)

	s := NewServer(":1234", c)
	s.dispatchPacket("Z2,00001000,4") // Write watchpoint at 0x1000

	if len(c.Watchpoints) != 1 {
		t.Fatalf("Expected 1 watchpoint, got %d", len(c.Watchpoints))
	}

	c.Running = true
	c.CheckWatchpoint(0x1000, 4, true)
	if c.HitWatchpoint != 0x1000 {
		t.Errorf("Expected HitWatchpoint 0x1000, got 0x%08x", c.HitWatchpoint)
	}
	if c.Running {
		t.Errorf("Expected CPU to pause on watchpoint hit")
	}
}
