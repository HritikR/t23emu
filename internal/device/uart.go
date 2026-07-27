package device

import (
	"fmt"
	"io"
	"os"
	"sync"
)

type UART struct {
	mu     sync.Mutex
	output io.Writer
	buf    []byte
}

// NewUART creates a new UART device. If out is nil, it defaults to os.Stdout.
func NewUART(out io.Writer) *UART {
	if out == nil {
		out = os.Stdout
	}
	return &UART{
		output: out,
		buf:    make([]byte, 0),
	}
}

// GetCapturedOutput returns all characters written to UART TX register.
func (u *UART) GetCapturedOutput() string {
	u.mu.Lock()
	defer u.mu.Unlock()
	return string(u.buf)
}

// Reset clears the captured output buffer.
func (u *UART) Reset() {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.buf = make([]byte, 0)
}

func (u *UART) Read8(addr uint32) byte {
	// Status register at offset 4
	if addr == 4 {
		// Bit 5 (0x20): TX FIFO empty
		// Bit 6 (0x40): TX transmitter empty
		return 0x60
	}
	return 0
}

func (u *UART) Write8(addr uint32, value byte) {
	// Data TX register at offset 0
	if addr == 0 {
		u.mu.Lock()
		u.buf = append(u.buf, value)
		u.mu.Unlock()

		fmt.Fprintf(u.output, "%c", value)
	}
}

func (u *UART) Read32(addr uint32) uint32 {
	return uint32(u.Read8(addr))
}

func (u *UART) Write32(addr uint32, value uint32) {
	u.Write8(addr, byte(value))
}

var _ Device = (*UART)(nil)
