package device

import (
	"fmt"
)

// Ingenic Serial Flash Controller register offsets, relative to the SFC
// physical base at 0x13440000.
const (
	SFC_GLB           uint32 = 0x000
	SFC_DEV_CONF      uint32 = 0x004
	SFC_DEV_STA_EXP   uint32 = 0x008
	SFC_DEV_STA_RT    uint32 = 0x00C
	SFC_DEV_STA_MSK   uint32 = 0x010
	SFC_TRAN_CONF     uint32 = 0x014
	SFC_TRAN_LEN      uint32 = 0x02C
	SFC_DEV_ADDR      uint32 = 0x030
	SFC_DEV_ADDR_PLUS uint32 = 0x048
	SFC_MEM_ADDR      uint32 = 0x060
	SFC_TRIG          uint32 = 0x064
	SFC_SR            uint32 = 0x068
	SFC_SCR           uint32 = 0x06C
	SFC_INTC          uint32 = 0x070
	SFC_FSM           uint32 = 0x074
	SFC_CGE           uint32 = 0x078
	SFC_DR            uint32 = 0x1000
)

const (
	SFC_TRIG_START uint32 = 1 << 0
	SFC_TRIG_STOP  uint32 = 1 << 1
	SFC_TRIG_FLUSH uint32 = 1 << 2

	// SFC_SR_RECE_REQ indicates that receive data is available.
	SFC_SR_RECE_REQ uint32 = 1 << 2
	SFC_SR_TRAN_REQ uint32 = 1 << 3
	SFC_SR_END      uint32 = 1 << 4

	SFC_SCR_CLR_RREQ uint32 = 1 << 2
	SFC_SCR_CLR_END  uint32 = 1 << 4

	sfcFIFOBytes uint32 = 32 * 4
)

// SFC is a model of the Ingenic serial flash controller.
type SFC struct {
	*RegisterBlock

	flash []byte
	reply []byte

	Interrupt func(assert bool)

	command byte

	active    bool
	done      bool
	wel       bool
	addr      uint32
	index     uint32
	remaining uint32
}

// NewSFC creates the serial flash controller.
// If size is 0 or smaller than len(flash), capacity defaults to len(flash).
func NewSFC(flash []byte, size uint32) *SFC {
	if uint32(len(flash)) > size {
		size = uint32(len(flash))
	}
	buf := make([]byte, size)
	for i := range buf {
		buf[i] = 0xFF
	}
	copy(buf, flash)

	sfc := &SFC{
		RegisterBlock: NewRegisterBlock("SFC", 0x2000),
		flash:         buf,
	}

	names := map[uint32]string{
		SFC_GLB:           "GLB",
		SFC_DEV_CONF:      "DEV_CONF",
		SFC_DEV_STA_EXP:   "DEV_STA_EXP",
		SFC_DEV_STA_RT:    "DEV_STA_RT",
		SFC_DEV_STA_MSK:   "DEV_STA_MSK",
		SFC_TRAN_CONF:     "TRAN_CONF",
		SFC_TRAN_LEN:      "TRAN_LEN",
		SFC_DEV_ADDR:      "DEV_ADDR",
		SFC_DEV_ADDR_PLUS: "DEV_ADDR_PLUS",
		SFC_MEM_ADDR:      "MEM_ADDR",
		SFC_TRIG:          "TRIG",
		SFC_SR:            "SR",
		SFC_SCR:           "SCR",
		SFC_INTC:          "INTC",
		SFC_FSM:           "FSM",
		SFC_CGE:           "CGE",
		SFC_DR:            "DR",
	}
	for offset, name := range names {
		sfc.SetName(offset, name)
	}

	return sfc
}

func (s *SFC) Read32(addr uint32) uint32 {
	offset := addr &^ 3
	s.readCounts[offset]++

	var value uint32
	switch offset {
	case SFC_SR:
		value = s.status()
	case SFC_DR:
		value = s.readData()
	default:
		value = s.regs[offset] | s.readOnes[offset]
	}

	if s.Trace {
		fmt.Fprintf(s.Out, "  %s read  %s => 0x%08x\n", s.Name, s.RegName(offset), value)
	}

	return value
}

func (s *SFC) Write32(addr uint32, value uint32) {
	offset := addr &^ 3
	s.writeCounts[offset]++
	s.regs[offset] = value

	switch offset {
	case SFC_TRIG:
		start := value&SFC_TRIG_START != 0
		if value&SFC_TRIG_FLUSH != 0 {
			s.active = false
			s.done = false
			s.remaining = 0
			s.setInterrupt(false)
		}
		if start {
			s.startTransfer()
		}
		if value&SFC_TRIG_STOP != 0 && !start {
			s.active = false
		}
	case SFC_SCR:
		if value&SFC_SCR_CLR_END != 0 {
			s.done = false
			s.setInterrupt(false)
		}
	case SFC_DR:
		if s.active && s.remaining > 0 && s.reply == nil && !isReadCommand(s.command) {
			for i := uint32(0); i < 4 && s.remaining > 0; i++ {
				b := byte(value >> (8 * i))
				if s.addr < uint32(len(s.flash)) {
					s.flash[s.addr] = b
				}
				s.addr++
				s.index++
				s.remaining--
			}
			if s.remaining == 0 {
				s.active = false
				s.done = true
				s.wel = false
				s.setInterrupt(true)
			}
		}
	}

	if s.Trace {
		fmt.Fprintf(s.Out, "  %s write %s <= 0x%08x\n", s.Name, s.RegName(offset), value)
	}
}

func (s *SFC) Read8(addr uint32) byte {
	word := s.Read32(addr &^ 3)
	return byte(word >> ((addr & 3) * 8))
}

func (s *SFC) Write8(addr uint32, value byte) {
	offset := addr &^ 3
	if offset == SFC_DR && s.active && s.remaining > 0 && s.reply == nil && !isReadCommand(s.command) {
		if s.addr < uint32(len(s.flash)) {
			s.flash[s.addr] = value
		}
		s.addr++
		s.index++
		s.remaining--
		if s.remaining == 0 {
			s.active = false
			s.done = true
			s.wel = false
			s.setInterrupt(true)
		}
		return
	}
	shift := (addr & 3) * 8
	word := s.regs[offset]
	word = (word & ^(uint32(0xFF) << shift)) | (uint32(value) << shift)
	s.Write32(offset, word)
}

func (s *SFC) startTransfer() {
	s.command = byte(s.regs[SFC_TRAN_CONF])
	s.reply = s.commandReply(s.command)
	s.addr = s.regs[SFC_DEV_ADDR]
	s.index = 0
	s.remaining = s.regs[SFC_TRAN_LEN]

	switch s.command {
	case 0x06: // Write Enable
		s.wel = true
	case 0x04: // Write Disable
		s.wel = false
	case 0x20, 0x21: // Sector Erase 4KB
		s.erase(s.addr&^0xFFF, 4096)
		s.wel = false
	case 0x52: // Block Erase 32KB
		s.erase(s.addr&^0x7FFF, 32768)
		s.wel = false
	case 0xD8, 0xDC: // Block Erase 64KB
		s.erase(s.addr&^0xFFFF, 65536)
		s.wel = false
	case 0xC7, 0x60: // Chip Erase
		s.erase(0, uint32(len(s.flash)))
		s.wel = false
	}

	s.active = s.remaining > 0
	s.done = true
	s.setInterrupt(true)
}

func (s *SFC) erase(start uint32, length uint32) {
	for i := uint32(0); i < length; i++ {
		if start+i < uint32(len(s.flash)) {
			s.flash[start+i] = 0xFF
		}
	}
}

func (s *SFC) status() uint32 {
	value := s.regs[SFC_SR]
	if s.active && s.remaining > 0 {
		if !isReadCommand(s.command) && s.reply == nil {
			value |= SFC_SR_TRAN_REQ
		} else {
			value |= SFC_SR_RECE_REQ
		}
		fifoWords := (min32(s.remaining, sfcFIFOBytes) + 3) / 4
		value |= fifoWords << 16
	}
	if s.done {
		value |= SFC_SR_END
	}
	return value
}

func (s *SFC) readData() uint32 {
	var value uint32
	for i := uint32(0); i < 4; i++ {
		if s.remaining == 0 {
			break
		}
		if b, ok := s.readByte(); ok {
			value |= uint32(b) << (8 * i)
		}
		s.addr++
		s.index++
		s.remaining--
	}

	if s.remaining == 0 {
		s.active = false
		s.done = true
		s.setInterrupt(true)
	}

	s.completeIfFinished()

	return value
}

func (s *SFC) readByte() (byte, bool) {
	if s.reply != nil {
		if s.index < uint32(len(s.reply)) {
			return s.reply[s.index], true
		}
		return 0, true
	}
	if s.addr < uint32(len(s.flash)) {
		return s.flash[s.addr], true
	}
	return 0, false
}

func (s *SFC) completeIfFinished() {
	if s.reply != nil && s.command != 0x05 && s.command != 0x35 && s.index >= uint32(len(s.reply)) {
		s.active = false
		s.done = true
		return
	}

	if s.remaining != 0 {
		return
	}

	s.active = false
	s.done = true
	s.setInterrupt(true)
}

func (s *SFC) setInterrupt(assert bool) {
	if s.Interrupt != nil {
		s.Interrupt(assert)
	}
}

func (s *SFC) commandReply(command byte) []byte {
	switch command {
	case 0x05:
		// Read status register 1. Return WEL bit if enabled.
		val := byte(0x00)
		if s.wel {
			val |= 0x02 // WEL (Write Enable Latch)
		}
		return []byte{val}

	case 0x35:
		// Read status register 2.
		return []byte{0x00}

	case 0x90:
		// Read manufacturer/device ID (REMS).
		// P25Q64H: manufacturer 0x85, device ID 0x16.
		return []byte{0x85, 0x16}

	case 0x9f:
		// JEDEC RDID for Puya P25Q64H 8 MiB SPI NOR.
		return []byte{0x85, 0x60, 0x17}

	default:
		return nil
	}
}

func isReadCommand(cmd byte) bool {
	switch cmd {
	case 0x03, 0x13, 0x0B, 0x0C, 0x3B, 0x3C, 0x6B, 0x6C, 0xEB, 0xEC, 0x5A, 0x05, 0x35, 0x90, 0x9F:
		return true
	default:
		return false
	}
}

func min32(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

var _ Device = (*SFC)(nil)
