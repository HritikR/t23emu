package cpu

import (
	"encoding/binary"
)

// ARM64Emitter constructs AArch64 host machine instructions.
type ARM64Emitter struct {
	code []byte
}

// NewARM64Emitter creates a new emitter instance.
func NewARM64Emitter() *ARM64Emitter {
	return &ARM64Emitter{
		code: make([]byte, 0, 1024),
	}
}

// Bytes returns the assembled machine code.
func (e *ARM64Emitter) Bytes() []byte {
	return e.code
}

// Emit32 appends a 32-bit ARM64 instruction (little-endian).
func (e *ARM64Emitter) Emit32(inst uint32) {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], inst)
	e.code = append(e.code, buf[:]...)
}

// RET emits ARM64 ret instruction.
func (e *ARM64Emitter) RET() {
	e.Emit32(0xD65F03C0)
}

// NOP emits ARM64 nop.
func (e *ARM64Emitter) NOP() {
	e.Emit32(0xD503201F)
}

// LDR_W_X0 emits LDR Wd, [X0, #imm12] (Load 32-bit word from struct offset).
func (e *ARM64Emitter) LDR_W_X0(wd uint32, offset uint32) {
	imm12 := (offset / 4) & 0xFFF
	inst := uint32(0xB9400000) | (imm12 << 10) | (0 << 5) | (wd & 0x1F)
	e.Emit32(inst)
}

// STR_W_X0 emits STR Wd, [X0, #imm12] (Store 32-bit word into struct offset).
func (e *ARM64Emitter) STR_W_X0(wd uint32, offset uint32) {
	imm12 := (offset / 4) & 0xFFF
	inst := uint32(0xB9000000) | (imm12 << 10) | (0 << 5) | (wd & 0x1F)
	e.Emit32(inst)
}

// ADD_W emits ADD Wd, Wn, Wm
func (e *ARM64Emitter) ADD_W(wd, wn, wm uint32) {
	inst := uint32(0x0B000000) | ((wm & 0x1F) << 16) | ((wn & 0x1F) << 5) | (wd & 0x1F)
	e.Emit32(inst)
}

// SUB_W emits SUB Wd, Wn, Wm
func (e *ARM64Emitter) SUB_W(wd, wn, wm uint32) {
	inst := uint32(0x4B000000) | ((wm & 0x1F) << 16) | ((wn & 0x1F) << 5) | (wd & 0x1F)
	e.Emit32(inst)
}

// AND_W emits AND Wd, Wn, Wm
func (e *ARM64Emitter) AND_W(wd, wn, wm uint32) {
	inst := uint32(0x0A000000) | ((wm & 0x1F) << 16) | ((wn & 0x1F) << 5) | (wd & 0x1F)
	e.Emit32(inst)
}

// ORR_W emits ORR Wd, Wn, Wm
func (e *ARM64Emitter) ORR_W(wd, wn, wm uint32) {
	inst := uint32(0x2A000000) | ((wm & 0x1F) << 16) | ((wn & 0x1F) << 5) | (wd & 0x1F)
	e.Emit32(inst)
}

// EOR_W emits EOR Wd, Wn, Wm
func (e *ARM64Emitter) EOR_W(wd, wn, wm uint32) {
	inst := uint32(0x4A000000) | ((wm & 0x1F) << 16) | ((wn & 0x1F) << 5) | (wd & 0x1F)
	e.Emit32(inst)
}

// MOVZ_W emits MOVZ Wd, #imm16
func (e *ARM64Emitter) MOVZ_W(wd uint32, imm16 uint16) {
	inst := uint32(0x52800000) | (uint32(imm16) << 5) | (wd & 0x1F)
	e.Emit32(inst)
}
