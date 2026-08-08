package sensor

import "testing"

func TestSC2336RegisterRead(t *testing.T) {
	s := NewSC2336()

	s.WriteI2C(0x31)
	s.WriteI2C(0x07)
	if got := s.ReadI2C(); got != 0xCB {
		t.Fatalf("expected 0xCB for register 0x3107, got 0x%02X", got)
	}

	s.WriteI2C(0x31)
	s.WriteI2C(0x08)
	if got := s.ReadI2C(); got != 0x3A {
		t.Fatalf("expected 0x3A for register 0x3108, got 0x%02X", got)
	}

	s.WriteI2C(0x00)
	s.WriteI2C(0x00)
	if got := s.ReadI2C(); got != 0x00 {
		t.Fatalf("expected 0x00 for unhandled register, got 0x%02X", got)
	}
}
