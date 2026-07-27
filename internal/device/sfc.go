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

	// SFC_SR_RECE_REQ indicates that receive data is available. The SPL
	// polls it and then drains words from SFC_DR.
	SFC_SR_RECE_REQ uint32 = 1 << 2
	SFC_SR_TRAN_REQ uint32 = 1 << 3
	SFC_SR_END      uint32 = 1 << 4

	SFC_SCR_CLR_RREQ uint32 = 1 << 2
	SFC_SCR_CLR_END  uint32 = 1 << 4

	sfcFIFOBytes uint32 = 32 * 4
)

// SFC is a small model of the Ingenic serial flash controller.
//
// It is intentionally narrow: the SPL only needs PIO reads from the data
// register. Configuration registers still read back through RegisterBlock so
// diagnostics show what the firmware programmed.
type SFC struct {
	*RegisterBlock

	flash []byte
	reply []byte

	active    bool
	done      bool
	addr      uint32
	index     uint32
	remaining uint32
}

// NewSFC creates the serial flash controller.
func NewSFC(flash []byte) *SFC {
	sfc := &SFC{
		RegisterBlock: NewRegisterBlock("SFC", 0x2000),
		flash:         append([]byte(nil), flash...),
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
		if value&SFC_TRIG_FLUSH != 0 {
			s.active = false
			s.done = false
			s.remaining = 0
		}
		if value&SFC_TRIG_START != 0 {
			s.startTransfer()
		}
		if value&SFC_TRIG_STOP != 0 {
			s.active = false
		}
	case SFC_SCR:
		if value&SFC_SCR_CLR_END != 0 {
			s.done = false
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
	shift := (addr & 3) * 8
	word := s.regs[offset]
	word = (word & ^(uint32(0xFF) << shift)) | (uint32(value) << shift)
	s.Write32(offset, word)
}

func (s *SFC) startTransfer() {
	s.reply = s.commandReply(byte(s.regs[SFC_TRAN_CONF]))
	s.addr = s.regs[SFC_DEV_ADDR]
	s.index = 0
	s.remaining = s.regs[SFC_TRAN_LEN]
	s.active = s.remaining > 0
	s.done = true
}

func (s *SFC) status() uint32 {
	value := s.regs[SFC_SR]
	if s.active && s.remaining > 0 {
		value |= SFC_SR_RECE_REQ
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
	}

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

func (s *SFC) commandReply(command byte) []byte {
	switch command {
	case 0x05, 0x35:
		// Read status registers. Keep WIP/WEL clear so polls complete.
		return []byte{0x00}
	case 0x90:
		// Read manufacturer/device ID.
		return []byte{0xef, 0x16}
	case 0x9f:
		// JEDEC RDID for a Winbond W25Q64-class 8 MiB SPI NOR.
		return []byte{0xef, 0x40, 0x17}
	default:
		return nil
	}
}

func min32(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

var _ Device = (*SFC)(nil)
