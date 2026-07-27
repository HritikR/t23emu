// Command imgscan recovers the link address of a raw MIPS boot image.
//
// It works by testing candidate image-offset mappings against the call
// targets in the image: with the correct mapping, nearly every JAL lands
// on a function prologue, and with a wrong one they land on noise.
package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/HritikR/t23emu/internal/cpu"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: imgscan <image> [dumpOffset count]")
		os.Exit(1)
	}

	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if len(os.Args) >= 4 {
		off, _ := strconv.ParseUint(os.Args[2], 0, 32)
		n, _ := strconv.ParseUint(os.Args[3], 0, 32)
		dump(data, uint32(off), int(n))
		return
	}

	if len(os.Args) >= 3 && os.Args[2] == "regions" {
		regions(data)
		return
	}

	scan(data)
}

// regions reports how instruction-like each 1 KB block of the image is,
// which separates real code from data and padding. Offset scoring is only
// meaningful over blocks that are actually code.
func regions(data []byte) {
	const block = 1024

	fmt.Println("offset     decode%  jal  cache  mtc0  note")

	for base := uint32(0); base < uint32(len(data)); base += block {
		var total, ok, jal, cache, mtc0 int

		for off := base; off+4 <= base+block && off+4 <= uint32(len(data)); off += 4 {
			raw := word(data, off)
			total++
			if raw == 0 {
				// A nop is valid code but also the commonest padding, so
				// it is counted as decodable without being evidence.
				ok++
				continue
			}
			if plausible(raw) {
				ok++
			}
			inst := cpu.Decode(raw)
			switch {
			case inst.Opcode == cpu.OP_JAL:
				jal++
			case inst.Opcode == cpu.OP_CACHE:
				cache++
			case inst.Opcode == cpu.OP_COP0 && inst.Rs == cpu.COP0_MTC0:
				mtc0++
			}
		}

		if total == 0 {
			continue
		}

		pct := 100 * float64(ok) / float64(total)

		note := ""
		if pct > 90 {
			note = "CODE"
		} else if pct < 40 {
			note = "data"
		}

		fmt.Printf("0x%06x   %5.1f%%  %3d  %5d  %4d  %s\n",
			base, pct, jal, cache, mtc0, note)
	}
}

func word(data []byte, off uint32) uint32 {
	if off+4 > uint32(len(data)) {
		return 0
	}
	return uint32(data[off]) | uint32(data[off+1])<<8 |
		uint32(data[off+2])<<16 | uint32(data[off+3])<<24
}

// isPrologue reports whether an instruction looks like the first
// instruction of a compiled function. Almost all non-leaf MIPS functions
// open by making stack room.
func isPrologue(raw uint32) bool {
	inst := cpu.Decode(raw)

	// addiu $sp, $sp, -N
	if inst.Opcode == cpu.OP_ADDIU && inst.Rs == 29 && inst.Rt == 29 &&
		int16(inst.Immediate) < 0 {
		return true
	}
	return false
}

// plausible reports whether a word decodes to something that could be an
// instruction at all, used as a weaker signal than isPrologue.
func plausible(raw uint32) bool {
	if raw == 0 {
		return false
	}
	return cpu.Disassemble(raw, 0)[0] != '.'
}

// codeStart and codeEnd bound the region of the image that the region
// report identifies as real instructions. Scoring outside it just
// measures how often random data happens to decode.
const (
	codeStart uint32 = 0x800
	codeEnd   uint32 = 0x2c00
)

// scan recovers the mapping between image offsets and virtual addresses.
//
// It sweeps the single unknown, D, defined by
//
//	virtual address low 28 bits of image offset X = X + D
//
// and scores each candidate by how many JAL targets inside the code
// region resolve to a function prologue. The correct D produces a sharp
// peak because every compiled call in the image agrees on it.
func scan(data []byte) {

	type result struct {
		d              int64
		calls, inRange int
		prologues      int
	}

	var best []result

	for d := int64(-0x8000); d <= 0x8000; d += 4 {

		r := result{d: d}

		for off := codeStart; off+4 <= codeEnd; off += 4 {
			inst := cpu.Decode(word(data, off))

			if inst.Opcode != cpu.OP_JAL {
				continue
			}

			r.calls++

			// Target virtual low 28 bits back to an image offset.
			t := int64(inst.Target<<2) - d
			if t < int64(codeStart) || t+4 > int64(codeEnd) {
				continue
			}

			r.inRange++
			if isPrologue(word(data, uint32(t))) {
				r.prologues++
			}
		}

		best = append(best, r)
	}

	sort.Slice(best, func(i, j int) bool {
		if best[i].prologues != best[j].prologues {
			return best[i].prologues > best[j].prologues
		}
		return best[i].inRange > best[j].inRange
	})

	fmt.Printf("JAL instructions in code region: %d\n\n", best[0].calls)
	fmt.Println("       D   targets in code   prologues")
	for _, r := range best[:10] {
		fmt.Printf("%+#08x   %15d   %9d\n", r.d, r.inRange, r.prologues)
	}
}

func dump(data []byte, off uint32, n int) {
	for i := 0; i < n; i++ {
		o := off + uint32(i*4)
		raw := word(data, o)
		fmt.Printf("%08x: %08x  %s\n", o, raw, cpu.Disassemble(raw, o))
	}
}
