package device

import (
	"fmt"
	"io"
	"os"
	"sync"
)

// Ingenic UART register offsets. The part uses a 16550-compatible layout
// with registers spaced four bytes apart.
const (
	UART_RBR  uint32 = 0x00 // Receive buffer / transmit hold / divisor low
	UART_IER  uint32 = 0x04 // Interrupt enable / divisor high
	UART_FCR  uint32 = 0x08 // FIFO control on write, interrupt ID on read
	UART_LCR  uint32 = 0x0C // Line control
	UART_MCR  uint32 = 0x10 // Modem control
	UART_LSR  uint32 = 0x14 // Line status
	UART_MSR  uint32 = 0x18 // Modem status
	UART_SPR  uint32 = 0x1C // Scratch
	UART_UMR  uint32 = 0x20 // Ingenic clock divider mode
	UART_UACR uint32 = 0x24 // Ingenic clock divider add cycle
)

// Line status register bits.
const (
	LSR_DR   byte = 1 << 0 // Receive data ready
	LSR_OE   byte = 1 << 1 // Overrun
	LSR_THRE byte = 1 << 5 // Transmit hold register empty
	LSR_TEMT byte = 1 << 6 // Transmitter completely empty
)

// Line control register bits.
const (
	LCR_DLAB byte = 1 << 7 // Divisor latch access
)

// FIFO control register bits.
const (
	FCR_FE  byte = 1 << 0 // FIFO enable
	FCR_UME byte = 1 << 4 // Ingenic UART module enable
)

// UART models a 16550-compatible Ingenic serial port.
//
// Transmission is instantaneous, so the line status register always
// reports the transmitter as empty. Boot firmware spins on those bits
// before every character it prints, so a port that does not report them
// stalls the boot on its first line of output.
type UART struct {
	mu     sync.Mutex
	output io.Writer
	buf    []byte

	// Name labels the port in aggregated console output.
	Name string

	// ier through spr hold the writable registers.
	ier byte
	fcr byte
	lcr byte
	mcr byte
	msr byte
	spr byte

	// dll and dlm are the baud rate divisor latches, visible in place of
	// RBR and IER while the divisor latch access bit is set.
	dll byte
	dlm byte

	// umr and uacr are the Ingenic-specific divider registers.
	umr  byte
	uacr byte

	// rx holds bytes queued for the firmware to read.
	rx []byte
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

// NewNamedUART creates a UART labelled for multi-port output.
func NewNamedUART(name string, out io.Writer) *UART {
	u := NewUART(out)
	u.Name = name
	return u
}

// GetCapturedOutput returns all characters written to the UART's
// transmit register.
func (u *UART) GetCapturedOutput() string {
	u.mu.Lock()
	defer u.mu.Unlock()
	return string(u.buf)
}

// SetOutput changes where transmitted characters are echoed live. Captured
// output is unaffected.
func (u *UART) SetOutput(out io.Writer) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.output = out
}

// Reset clears the captured output buffer.
func (u *UART) Reset() {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.buf = make([]byte, 0)
}

// Feed queues bytes for the firmware to receive.
func (u *UART) Feed(data []byte) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.rx = append(u.rx, data...)
}

// dlab reports whether the divisor latches are currently mapped over the
// first two registers.
func (u *UART) dlab() bool {
	return u.lcr&LCR_DLAB != 0
}

func (u *UART) Read8(addr uint32) byte {
	u.mu.Lock()
	defer u.mu.Unlock()

	// Registers are four bytes apart; only the low byte of each slot
	// carries data.
	if addr&3 != 0 {
		return 0
	}

	switch addr & 0xFF {

	case UART_RBR:
		if u.dlab() {
			return u.dll
		}
		if len(u.rx) > 0 {
			b := u.rx[0]
			u.rx = u.rx[1:]
			return b
		}
		return 0

	case UART_IER:
		if u.dlab() {
			return u.dlm
		}
		return u.ier

	case UART_FCR:
		// Read back as the interrupt identification register: no
		// interrupt pending, with the FIFO enable bits mirrored.
		iir := byte(0x01)
		if u.fcr&FCR_FE != 0 {
			iir |= 0xC0
		}
		return iir

	case UART_LCR:
		return u.lcr

	case UART_MCR:
		return u.mcr

	case UART_LSR:
		// The transmitter is always idle because writes complete
		// immediately, so THRE and TEMT are always set.
		status := LSR_THRE | LSR_TEMT
		if len(u.rx) > 0 {
			status |= LSR_DR
		}
		return status

	case UART_MSR:
		return u.msr

	case UART_SPR:
		return u.spr

	case UART_UMR:
		return u.umr

	case UART_UACR:
		return u.uacr
	}

	return 0
}

func (u *UART) Write8(addr uint32, value byte) {
	u.mu.Lock()

	if addr&3 != 0 {
		u.mu.Unlock()
		return
	}

	// emit carries the character to print after the lock is released.
	var emit bool

	switch addr & 0xFF {

	case UART_RBR:
		if u.dlab() {
			// Programming the baud rate divisor, not sending a character.
			u.dll = value
		} else {
			u.buf = append(u.buf, value)
			emit = true
		}

	case UART_IER:
		if u.dlab() {
			u.dlm = value
		} else {
			u.ier = value
		}

	case UART_FCR:
		u.fcr = value

	case UART_LCR:
		u.lcr = value

	case UART_MCR:
		u.mcr = value

	case UART_LSR:
		// Line status is read-only.

	case UART_MSR:
		u.msr = value

	case UART_SPR:
		u.spr = value

	case UART_UMR:
		u.umr = value

	case UART_UACR:
		u.uacr = value
	}

	out := u.output
	u.mu.Unlock()

	if emit {
		fmt.Fprintf(out, "%c", value)
	}
}

func (u *UART) Read32(addr uint32) uint32 {
	return uint32(u.Read8(addr))
}

func (u *UART) Write32(addr uint32, value uint32) {
	u.Write8(addr, byte(value))
}

var _ Device = (*UART)(nil)
