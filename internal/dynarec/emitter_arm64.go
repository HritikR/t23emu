package dynarec

import (
	"encoding/binary"
)

// ARM64Emitter generates native AArch64 machine code instructions into a byte slice.
type ARM64Emitter struct {
	Buf []byte
}

func NewARM64Emitter(buf []byte) *ARM64Emitter {
	return &ARM64Emitter{Buf: buf}
}

func (e *ARM64Emitter) Emit32(inst uint32) {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], inst)
	e.Buf = append(e.Buf, b[:]...)
}

// RET emits `ret` (0xD65F03C0)
func (e *ARM64Emitter) RET() {
	e.Emit32(0xD65F03C0)
}

// NOP emits `nop` (0xD503201F)
func (e *ARM64Emitter) NOP() {
	e.Emit32(0xD503201F)
}

// LDR_W_X0 emits `ldr w<rd>, [x0, #<offset>]`
func (e *ARM64Emitter) LDR_W_X0(rd uint8, offset uint32) {
	imm12 := (offset / 4) & 0xFFF
	inst := uint32(0xB9400000) | (uint32(imm12) << 10) | (uint32(rd & 0x1F))
	e.Emit32(inst)
}

// STR_W_X0 emits `str w<rs>, [x0, #<offset>]`
func (e *ARM64Emitter) STR_W_X0(rs uint8, offset uint32) {
	imm12 := (offset / 4) & 0xFFF
	inst := uint32(0xB9000000) | (uint32(imm12) << 10) | (uint32(rs & 0x1F))
	e.Emit32(inst)
}

// ADD_W emits `add w<rd>, w<rn>, w<rm>`
func (e *ARM64Emitter) ADD_W(rd, rn, rm uint8) {
	inst := uint32(0x0B000000) | (uint32(rm&0x1F) << 16) | (uint32(rn&0x1F) << 5) | uint32(rd&0x1F)
	e.Emit32(inst)
}

// SUB_W emits `sub w<rd>, w<rn>, w<rm>`
func (e *ARM64Emitter) SUB_W(rd, rn, rm uint8) {
	inst := uint32(0x4B000000) | (uint32(rm&0x1F) << 16) | (uint32(rn&0x1F) << 5) | uint32(rd&0x1F)
	e.Emit32(inst)
}

// AND_W emits `and w<rd>, w<rn>, w<rm>`
func (e *ARM64Emitter) AND_W(rd, rn, rm uint8) {
	inst := uint32(0x0A000000) | (uint32(rm&0x1F) << 16) | (uint32(rn&0x1F) << 5) | uint32(rd&0x1F)
	e.Emit32(inst)
}

// ORR_W emits `orr w<rd>, w<rn>, w<rm>`
func (e *ARM64Emitter) ORR_W(rd, rn, rm uint8) {
	inst := uint32(0x2A000000) | (uint32(rm&0x1F) << 16) | (uint32(rn&0x1F) << 5) | uint32(rd&0x1F)
	e.Emit32(inst)
}

// EOR_W emits `eor w<rd>, w<rn>, w<rm>`
func (e *ARM64Emitter) EOR_W(rd, rn, rm uint8) {
	inst := uint32(0x4A000000) | (uint32(rm&0x1F) << 16) | (uint32(rn&0x1F) << 5) | uint32(rd&0x1F)
	e.Emit32(inst)
}

// MOVZ_W emits `movz w<rd>, #<imm16>`
func (e *ARM64Emitter) MOVZ_W(rd uint8, imm16 uint16) {
	inst := uint32(0x52800000) | (uint32(imm16) << 5) | uint32(rd&0x1F)
	e.Emit32(inst)
}
