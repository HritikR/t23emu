package device

import (
	"bytes"
	"testing"
)

func TestUARTWriteAndCapture(t *testing.T) {
	var out bytes.Buffer
	uart := NewUART(&out)

	// Write characters to UART TX (addr 0)
	uart.Write8(0, 'H')
	uart.Write8(0, 'e')
	uart.Write8(0, 'l')
	uart.Write8(0, 'l')
	uart.Write8(0, 'o')

	// Check buffer output
	if out.String() != "Hello" {
		t.Fatalf("expected printed output to be 'Hello', got '%s'", out.String())
	}

	// Check captured output buffer
	if uart.GetCapturedOutput() != "Hello" {
		t.Fatalf("expected captured output to be 'Hello', got '%s'", uart.GetCapturedOutput())
	}

	// Write 32-bit character
	uart.Write32(0, '!')
	if out.String() != "Hello!" {
		t.Fatalf("expected printed output to be 'Hello!', got '%s'", out.String())
	}

	// Reset and check
	uart.Reset()
	if uart.GetCapturedOutput() != "" {
		t.Fatalf("expected captured output to be empty after reset, got '%s'", uart.GetCapturedOutput())
	}
}

func TestUARTStatusRegister(t *testing.T) {
	uart := NewUART(nil)

	// Reading offset 0x14 (status register) should return 0x60 (ready to TX)
	status := uart.Read8(0x14)
	if status != 0x60 {
		t.Fatalf("expected status 0x60, got 0x%02X", status)
	}

	status32 := uart.Read32(0x14)
	if status32 != 0x60 {
		t.Fatalf("expected 32-bit status 0x60, got 0x%08X", status32)
	}
}
