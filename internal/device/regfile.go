package device

import (
	"fmt"
	"io"
	"os"
	"sort"
)

// RegisterBlock is a generic memory-mapped register file.
//
// It backs a peripheral whose behaviour the emulator does not model,
// but which firmware still expects to behave like memory: a value
// written to a register must read back. A silent stub that always
// returns zero breaks any driver that configures a peripheral and then
// checks its own configuration.
//
// On top of read/write storage it supports two things needed to get past
// hardware handshakes:
//
//   - ReadOnes forces selected bits high on read, which is how status
//     flags such as "PLL locked" are represented.
//   - Access counting records how often each register is read, which
//     identifies the register a stalled boot is spinning on.
type RegisterBlock struct {
	// Name identifies the block in trace output.
	Name string

	size uint32

	// regs holds one stored value per word-aligned register, keyed by
	// offset. A map rather than a slice so that a block can span a whole
	// peripheral window without allocating storage for every address in
	// it; firmware only ever touches a sparse handful.
	regs map[uint32]uint32

	// readOnes maps a register offset to a mask of bits that always read
	// as 1, regardless of what was written.
	readOnes map[uint32]uint32

	// names maps a register offset to a symbolic name for tracing.
	names map[uint32]string

	// readFuncs maps a register offset to a function supplying its value.
	// It is how live registers such as a free-running counter are modelled
	// without giving every peripheral its own device type.
	readFuncs map[uint32]func() uint32

	// writeFuncs maps a register offset to a side effect run on write. It
	// covers registers that act rather than store: set/clear aliases onto
	// another register, and write-one-to-clear controls.
	writeFuncs map[uint32]func(uint32)

	// readCounts and writeCounts record accesses per offset.
	readCounts  map[uint32]uint64
	writeCounts map[uint32]uint64

	// Trace enables per-access logging.
	Trace bool

	// Out receives trace output; defaults to os.Stderr.
	Out io.Writer
}

// NewRegisterBlock creates a register file covering size bytes.
func NewRegisterBlock(name string, size uint32) *RegisterBlock {
	return &RegisterBlock{
		Name:        name,
		size:        size,
		regs:        make(map[uint32]uint32),
		readOnes:    make(map[uint32]uint32),
		names:       make(map[uint32]string),
		readFuncs:   make(map[uint32]func() uint32),
		writeFuncs:  make(map[uint32]func(uint32)),
		readCounts:  make(map[uint32]uint64),
		writeCounts: make(map[uint32]uint64),
		Out:         os.Stderr,
	}
}

// SetReadOnes marks bits that always read as 1 at the given offset.
// Use it for read-only status flags that firmware polls.
func (r *RegisterBlock) SetReadOnes(offset uint32, mask uint32) {
	r.readOnes[offset&^3] |= mask
}

// SetInitial presets a register's stored value, for registers with a
// meaningful hardware reset value.
func (r *RegisterBlock) SetInitial(offset uint32, value uint32) {
	if offset < r.size {
		r.regs[offset&^3] = value
	}
}

// SetReadFunc installs a function that supplies a register's value on
// every read, for registers whose value is not simply what was last
// written to them.
func (r *RegisterBlock) SetReadFunc(offset uint32, fn func() uint32) {
	r.readFuncs[offset&^3] = fn
}

// SetWriteFunc installs a side effect run whenever a register is
// written, for registers that act on a write rather than simply hold
// the value. The value is still stored, so the write remains visible in
// a register dump.
func (r *RegisterBlock) SetWriteFunc(offset uint32, fn func(uint32)) {
	r.writeFuncs[offset&^3] = fn
}

// SetName attaches a symbolic name to a register offset for tracing.
func (r *RegisterBlock) SetName(offset uint32, name string) {
	r.names[offset&^3] = name
}

// RegName returns the symbolic name of an offset, or a hex offset when
// the register is unnamed.
func (r *RegisterBlock) RegName(offset uint32) string {
	if name, ok := r.names[offset&^3]; ok {
		return name
	}
	return fmt.Sprintf("+0x%03x", offset&^3)
}

func (r *RegisterBlock) Read32(addr uint32) uint32 {
	offset := addr &^ 3

	r.readCounts[offset]++

	var value uint32
	if fn, ok := r.readFuncs[offset]; ok {
		value = fn()
	} else {
		value = r.regs[offset]
	}

	value |= r.readOnes[offset]

	if r.Trace {
		fmt.Fprintf(r.Out, "  %s read  %s => 0x%08x\n", r.Name, r.RegName(offset), value)
	}

	return value
}

func (r *RegisterBlock) Write32(addr uint32, value uint32) {
	offset := addr &^ 3

	r.writeCounts[offset]++

	r.regs[offset] = value

	if fn, ok := r.writeFuncs[offset]; ok {
		fn(value)
	}

	if r.Trace {
		fmt.Fprintf(r.Out, "  %s write %s <= 0x%08x\n", r.Name, r.RegName(offset), value)
	}
}

func (r *RegisterBlock) Read8(addr uint32) byte {
	word := r.Read32(addr &^ 3)
	return byte(word >> ((addr & 3) * 8))
}

func (r *RegisterBlock) Write8(addr uint32, value byte) {
	offset := addr &^ 3
	shift := (addr & 3) * 8

	word := r.regs[offset]

	word = (word & ^(uint32(0xFF) << shift)) | (uint32(value) << shift)

	r.Write32(offset, word)
}

// Access describes how often one register was touched.
type Access struct {
	Offset uint32
	Name   string
	Reads  uint64
	Writes uint64
	Value  uint32
}

func (r *RegisterBlock) value(offset uint32) uint32 {
	offset = offset &^ 3

	var value uint32
	if fn, ok := r.readFuncs[offset]; ok {
		value = fn()
	} else {
		value = r.regs[offset]
	}

	return value | r.readOnes[offset]
}

// HotRegisters returns the registers read at least min times, most read
// first. A register read thousands of times is almost always a poll loop
// waiting on a status bit the emulator does not implement.
func (r *RegisterBlock) HotRegisters(min uint64) []Access {
	out := make([]Access, 0, len(r.readCounts))

	for offset, reads := range r.readCounts {
		if reads < min {
			continue
		}
		out = append(out, Access{
			Offset: offset,
			Name:   r.RegName(offset),
			Reads:  reads,
			Writes: r.writeCounts[offset],
			Value:  r.value(offset),
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Reads != out[j].Reads {
			return out[i].Reads > out[j].Reads
		}
		return out[i].Offset < out[j].Offset
	})

	return out
}

// Written returns every register that firmware wrote, in offset order.
// It is the quickest way to see what a boot stage actually configured.
func (r *RegisterBlock) Written() []Access {
	out := make([]Access, 0, len(r.writeCounts))

	for offset, writes := range r.writeCounts {
		out = append(out, Access{
			Offset: offset,
			Name:   r.RegName(offset),
			Reads:  r.readCounts[offset],
			Writes: writes,
			Value:  r.regs[offset],
		})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Offset < out[j].Offset })

	return out
}

var _ Device = (*RegisterBlock)(nil)
