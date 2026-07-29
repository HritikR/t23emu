package device

import (
	"fmt"
)

const (
	MSC_STRPCL uint32 = 0x000
	MSC_STAT   uint32 = 0x004
	MSC_CLKRT  uint32 = 0x008
	MSC_CMDAT  uint32 = 0x00C
	MSC_RESTO  uint32 = 0x010
	MSC_RDTO   uint32 = 0x014
	MSC_BLKLEN uint32 = 0x018
	MSC_NOB    uint32 = 0x01C
	MSC_SNOB   uint32 = 0x020
	MSC_IMASK  uint32 = 0x024
	MSC_IREG   uint32 = 0x028
	MSC_CMD    uint32 = 0x02C
	MSC_ARG    uint32 = 0x030
	MSC_RES    uint32 = 0x034
	MSC_RXFIFO uint32 = 0x038
	MSC_TXFIFO uint32 = 0x03C
	MSC_LPM    uint32 = 0x040
	MSC_DMAC   uint32 = 0x044
	MSC_DMANDA uint32 = 0x048
	MSC_DMADA  uint32 = 0x04C
	MSC_DMALEN uint32 = 0x050
	MSC_DMACMD uint32 = 0x054
	MSC_CTRL2  uint32 = 0x058
	MSC_RTCNT  uint32 = 0x05C
)

const (
	// STRPCL command bits.
	MSC_STRPCL_CLOCK_CONTROL_STOP  uint32 = 1 << 0
	MSC_STRPCL_CLOCK_CONTROL_START uint32 = 1 << 1
	MSC_STRPCL_START_OP            uint32 = 1 << 2
	MSC_STRPCL_RESET               uint32 = 1 << 3

	// STAT bits commonly checked by old Ingenic U-Boot.
	MSC_STAT_END_CMD_RES    uint32 = 1 << 11
	MSC_STAT_DATA_TRAN_DONE uint32 = 1 << 12
	MSC_STAT_PRG_DONE       uint32 = 1 << 13
	MSC_STAT_TIME_OUT_RES   uint32 = 1 << 8
	MSC_STAT_TIME_OUT_READ  uint32 = 1 << 9
	MSC_STAT_IS_RESETTING   uint32 = 1 << 15
	MSC_STAT_CARD_DETECTED  uint32 = 1 << 7

	// Interrupt/status bits.
	MSC_IREG_END_CMD_RES    uint32 = 1 << 2
	MSC_IREG_DATA_TRAN_DONE uint32 = 1 << 0
	MSC_IREG_PRG_DONE       uint32 = 1 << 1
	MSC_IREG_TIME_OUT_RES   uint32 = 1 << 8
	MSC_IREG_TIME_OUT_READ  uint32 = 1 << 9
)

type MSC struct {
	*RegisterBlock

	cardPresent  bool
	clockRunning bool

	// Disk emulation backing store
	diskImage  []byte
	fifoBuffer []byte
	fifoIndex  int
}

func NewMSC(name string, cardPresent bool, diskImage []byte) *MSC {
	m := &MSC{
		RegisterBlock: NewRegisterBlock(name, 0x10000),
		cardPresent:   cardPresent,
		diskImage:     diskImage,
	}

	names := map[uint32]string{
		MSC_STRPCL: "STRPCL",
		MSC_STAT:   "STAT",
		MSC_CLKRT:  "CLKRT",
		MSC_CMDAT:  "CMDAT",
		MSC_RESTO:  "RESTO",
		MSC_RDTO:   "RDTO",
		MSC_BLKLEN: "BLKLEN",
		MSC_NOB:    "NOB",
		MSC_SNOB:   "SNOB",
		MSC_IMASK:  "IMASK",
		MSC_IREG:   "IREG",
		MSC_CMD:    "CMD",
		MSC_ARG:    "ARG",
		MSC_RES:    "RES",
		MSC_RXFIFO: "RXFIFO",
		MSC_TXFIFO: "TXFIFO",
		MSC_LPM:    "LPM",
		MSC_DMAC:   "DMAC",
		MSC_DMANDA: "DMANDA",
		MSC_DMADA:  "DMADA",
		MSC_DMALEN: "DMALEN",
		MSC_DMACMD: "DMACMD",
		MSC_CTRL2:  "CTRL2",
		MSC_RTCNT:  "RTCNT",
	}

	for offset, name := range names {
		m.SetName(offset, name)
	}

	return m
}

func (m *MSC) Read32(addr uint32) uint32 {
	offset := addr &^ uint32(3)
	m.readCounts[offset]++

	var value uint32

	switch offset {
	case MSC_STAT:
		value = m.status()

	case MSC_IREG:
		value = m.regs[MSC_IREG]

	case MSC_RES:
		value = 0

	case MSC_RXFIFO:
		if m.fifoIndex < len(m.fifoBuffer) {
			var word uint32
			for i := 0; i < 4 && m.fifoIndex < len(m.fifoBuffer); i++ {
				word |= uint32(m.fifoBuffer[m.fifoIndex]) << (uint(i) * 8)
				m.fifoIndex++
			}
			value = word
		} else {
			value = 0
		}

	default:
		value = m.regs[offset] | m.readOnes[offset]
	}

	if m.Trace {
		fmt.Fprintf(
			m.Out,
			"  %s read  %s => 0x%08x\n",
			m.Name,
			m.RegName(offset),
			value,
		)
	}

	return value
}

func (m *MSC) Write32(addr uint32, value uint32) {
	offset := addr &^ uint32(3)
	m.writeCounts[offset]++

	switch offset {
	case MSC_STRPCL:
		m.writeSTRPCL(value)

	case MSC_IREG:
		// IREG is write-one-to-clear.
		m.regs[MSC_IREG] &^= value

	default:
		m.regs[offset] = value
	}

	if m.Trace {
		fmt.Fprintf(
			m.Out,
			"  %s write %s <= 0x%08x\n",
			m.Name,
			m.RegName(offset),
			value,
		)
	}
}

func (m *MSC) Read8(addr uint32) byte {
	word := m.Read32(addr &^ uint32(3))
	shift := (addr & 3) * 8

	return byte(word >> shift)
}

func (m *MSC) Write8(addr uint32, value byte) {
	offset := addr &^ uint32(3)
	shift := (addr & 3) * 8

	word := m.regs[offset]
	word &= ^(uint32(0xFF) << shift)
	word |= uint32(value) << shift

	m.Write32(offset, word)
}

func (m *MSC) writeSTRPCL(value uint32) {
	m.regs[MSC_STRPCL] = value &^ (MSC_STRPCL_RESET |
		MSC_STRPCL_START_OP)

	if value&MSC_STRPCL_RESET != 0 {
		m.resetController()
	}

	if value&MSC_STRPCL_CLOCK_CONTROL_STOP != 0 {
		m.clockRunning = false
	}

	if value&MSC_STRPCL_CLOCK_CONTROL_START != 0 {
		m.clockRunning = true
	}

	if value&MSC_STRPCL_START_OP != 0 {
		m.startCommand()
	}
}

func (m *MSC) resetController() {
	m.regs[MSC_STAT] = 0
	m.regs[MSC_IREG] = 0
	m.clockRunning = false
	m.regs[MSC_STAT] &^= MSC_STAT_IS_RESETTING
}

func (m *MSC) startCommand() {
	if !m.cardPresent {
		m.regs[MSC_STAT] = MSC_STAT_END_CMD_RES
		m.regs[MSC_IREG] = MSC_IREG_END_CMD_RES
		return
	}

	cmdIndex := m.regs[MSC_CMD] & 0x3F
	arg := m.regs[MSC_ARG]

	m.regs[MSC_STAT] = MSC_STAT_END_CMD_RES
	m.regs[MSC_IREG] = MSC_IREG_END_CMD_RES

	switch cmdIndex {
	case 17, 18: // CMD17 (Read Single Block), CMD18 (Read Multiple Block)
		blockLen := m.regs[MSC_BLKLEN]
		if blockLen == 0 {
			blockLen = 512
		}
		nob := m.regs[MSC_NOB]
		if nob == 0 {
			nob = 1
		}

		byteOffset := int(arg) * int(blockLen)
		totalBytes := int(blockLen) * int(nob)

		if m.diskImage != nil && byteOffset < len(m.diskImage) {
			end := byteOffset + totalBytes
			if end > len(m.diskImage) {
				end = len(m.diskImage)
			}
			m.fifoBuffer = m.diskImage[byteOffset:end]
		} else {
			m.fifoBuffer = make([]byte, totalBytes)
		}
		m.fifoIndex = 0

		m.regs[MSC_STAT] |= MSC_STAT_DATA_TRAN_DONE
		m.regs[MSC_IREG] |= MSC_IREG_DATA_TRAN_DONE

	default:
		// Standard command acknowledgment for initialization sequence (CMD0, CMD8, ACMD41, etc.)
	}
}

func (m *MSC) status() uint32 {
	value := m.regs[MSC_STAT]
	value &^= MSC_STAT_IS_RESETTING

	if m.cardPresent {
		value |= MSC_STAT_CARD_DETECTED
	} else {
		value &^= MSC_STAT_CARD_DETECTED
	}

	return value
}

var _ Device = (*MSC)(nil)
