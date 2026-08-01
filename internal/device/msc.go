package device

import (
	"encoding/binary"
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
	MSC_STRPCL_CLOCK_CONTROL_STOP  uint32 = 1 << 0
	MSC_STRPCL_CLOCK_CONTROL_START uint32 = 1 << 1
	MSC_STRPCL_START_OP            uint32 = 1 << 2
	MSC_STRPCL_RESET               uint32 = 1 << 3

	MSC_STAT_IS_RESETTING    uint32 = 1 << 15
	MSC_STAT_PRG_DONE        uint32 = 1 << 13
	MSC_STAT_DATA_TRAN_DONE  uint32 = 1 << 12
	MSC_STAT_END_CMD_RES     uint32 = 1 << 11
	MSC_STAT_DATA_FIFO_FULL  uint32 = 1 << 7
	MSC_STAT_DATA_FIFO_EMPTY uint32 = 1 << 6
	MSC_STAT_TIME_OUT_RES    uint32 = 1 << 1
	MSC_STAT_TIME_OUT_READ   uint32 = 1 << 0

	MSC_IREG_END_CMD_RES    uint32 = 1 << 2
	MSC_IREG_DATA_TRAN_DONE uint32 = 1 << 0
	MSC_IREG_PRG_DONE       uint32 = 1 << 1
	MSC_IREG_TIME_OUT_RES   uint32 = 1 << 9
	MSC_IREG_TIME_OUT_READ  uint32 = 1 << 8
)

type MSC struct {
	*RegisterBlock

	cardPresent  bool
	clockRunning bool
	expectAppCmd bool
	rca          uint16
	isSDHC       bool

	diskImage  []byte
	fifoBuffer []byte
	fifoIndex  int

	// MSC_RES is a 16-bit streaming register. Keep the encoded wire response
	// as bytes so byte/halfword/word accesses all consume it correctly.
	resBuffer []byte
	resIndex  int
}

func NewMSC(name string, cardPresent bool, diskImage []byte) *MSC {
	m := &MSC{
		RegisterBlock: NewRegisterBlock(name, 0x10000),
		cardPresent:   cardPresent,
		diskImage:     diskImage,
		resBuffer:     make([]byte, 0),
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

func (m *MSC) setShortResponse(val uint32) {
	// The Ingenic driver reads three 16-bit values and assembles them as:
	//   response = res0<<24 | res1<<8 | (res2 & 0xff)
	// Encode the desired 32-bit SD response in exactly that layout.
	m.resBuffer = []byte{
		byte(val >> 24), 0,
		byte(val >> 8), byte(val >> 16),
		byte(val), 0,
	}
	m.resIndex = 0
}

func (m *MSC) setLongResponse(words ...uint32) {
	m.resBuffer = m.resBuffer[:0]
	m.resIndex = 0

	if len(words) == 0 {
		return
	}

	// The Ingenic driver reads one initial halfword, then two halfwords per
	// response word. The low byte of c becomes the high byte of the next word.
	appendHalfword := func(v uint16) {
		m.resBuffer = append(m.resBuffer, byte(v), byte(v>>8))
	}

	appendHalfword(uint16(words[0] >> 24))
	for i, word := range words {
		appendHalfword(uint16(word >> 8))
		nextTop := byte(0)
		if i+1 < len(words) {
			nextTop = byte(words[i+1] >> 24)
		}
		appendHalfword(uint16(word&0xff)<<8 | uint16(nextTop))
	}
}

func (m *MSC) readResponseByte() byte {
	if m.resIndex >= len(m.resBuffer) {
		return 0
	}
	value := m.resBuffer[m.resIndex]
	m.resIndex++
	return value
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
		for i := 0; i < 4; i++ {
			value |= uint32(m.readResponseByte()) << (uint(i) * 8)
		}

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
	offset := addr &^ uint32(3)
	if offset == MSC_RES {
		m.readCounts[MSC_RES]++
		value := m.readResponseByte()
		if m.Trace {
			fmt.Fprintf(m.Out, "  %s read8 %s => 0x%02x\n", m.Name, m.RegName(offset), value)
		}
		return value
	}

	word := m.Read32(offset)
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
	m.regs[MSC_STRPCL] = value &^ (MSC_STRPCL_RESET | MSC_STRPCL_START_OP)

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
	m.resBuffer = m.resBuffer[:0]
	m.resIndex = 0
	m.rca = 0
	m.isSDHC = false
	m.expectAppCmd = false
	m.clockRunning = false
	m.regs[MSC_STAT] &^= MSC_STAT_IS_RESETTING
}

func (m *MSC) startCommand() {
	m.resBuffer = m.resBuffer[:0] // Clear response stream for new command
	m.resIndex = 0

	if !m.cardPresent {
		m.regs[MSC_STAT] |= MSC_STAT_TIME_OUT_RES
		m.regs[MSC_IREG] |= MSC_IREG_TIME_OUT_RES
		m.expectAppCmd = false
		return
	}

	cmdIndex := m.regs[MSC_CMD] & 0x3F
	arg := m.regs[MSC_ARG]

	m.regs[MSC_STAT] &^= (MSC_STAT_DATA_TRAN_DONE | MSC_STAT_TIME_OUT_RES)
	m.regs[MSC_STAT] |= (MSC_STAT_END_CMD_RES | MSC_STAT_PRG_DONE)

	m.regs[MSC_IREG] &^= (MSC_IREG_DATA_TRAN_DONE | MSC_IREG_TIME_OUT_RES)
	m.regs[MSC_IREG] |= (MSC_IREG_END_CMD_RES | MSC_IREG_PRG_DONE)

	isAppCmd := m.expectAppCmd
	m.expectAppCmd = false

	defaultStatus := uint32(0x00000900)

	switch cmdIndex {
	case 0:
		m.rca = 0
		m.setShortResponse(0)

	case 1:
		m.setShortResponse(0x80FF8000)

	case 2, 10:
		// Standard 136-bit CID
		m.setLongResponse(
			0x0002504d,
			0x53303447,
			0x20000000,
			0x01000000,
		)

	case 3:
		m.rca = 0x0001
		m.setShortResponse((uint32(m.rca) << 16) | 0x0500)

	case 6: // SWITCH_FUNC
		m.setShortResponse(defaultStatus)
		m.fifoBuffer = make([]byte, 64)
		m.fifoIndex = 0
		m.regs[MSC_STAT] |= MSC_STAT_DATA_TRAN_DONE
		m.regs[MSC_IREG] |= MSC_IREG_DATA_TRAN_DONE

	case 7:
		reqRca := uint16(arg >> 16)
		if reqRca == m.rca {
			m.setShortResponse(0x00000900)
		} else {
			m.setShortResponse(0)
		}

	case 8:
		m.setShortResponse(arg & 0xFFF)

	case 9: // SEND_CSD
		m.setSDHCCSD()

	case 13:
		m.setShortResponse(0x00000900)
		if isAppCmd {
			m.fifoBuffer = make([]byte, 64)
			m.fifoBuffer[0] = 0x80
			m.fifoIndex = 0
			m.regs[MSC_STAT] |= MSC_STAT_DATA_TRAN_DONE
			m.regs[MSC_IREG] |= MSC_IREG_DATA_TRAN_DONE
		}

	case 16:
		m.setShortResponse(defaultStatus)

	case 41:
		if isAppCmd {
			resp := uint32(0x80FF8000)
			if (arg & (1 << 30)) != 0 {
				resp |= (1 << 30) // CCS flag
				m.isSDHC = true
			} else {
				m.isSDHC = false
			}
			m.setShortResponse(resp)
		} else {
			m.setShortResponse(0x80FF8000)
		}

	case 51: // SEND_SCR
		m.setShortResponse(defaultStatus)
		if isAppCmd {
			m.fifoBuffer = []byte{
				0x02, 0x35, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
			}
			m.fifoIndex = 0
			m.regs[MSC_STAT] |= MSC_STAT_DATA_TRAN_DONE
			m.regs[MSC_IREG] |= MSC_IREG_DATA_TRAN_DONE
		}

	case 55:
		m.expectAppCmd = true
		m.setShortResponse(defaultStatus | (1 << 5))

	case 17, 18: // READ_BLOCK
		m.setShortResponse(0x00000900)

		// CRITICAL FIX: Always strictly enforce 512 bytes
		// to bypass U-Boot 1-byte read glitch.
		blockLen := uint32(512)
		nob := m.regs[MSC_NOB]
		if nob == 0 {
			nob = 1
		}

		byteOffset := int(arg)
		if m.isSDHC {
			byteOffset = int(arg) * int(blockLen)
		}
		totalBytes := int(blockLen) * int(nob)
		end := byteOffset + totalBytes

		if m.diskImage == nil || byteOffset < 0 || end > len(m.diskImage) {
			m.fifoBuffer = nil
			m.fifoIndex = 0

			m.regs[MSC_STAT] &^= MSC_STAT_DATA_TRAN_DONE
			m.regs[MSC_STAT] |= MSC_STAT_TIME_OUT_READ

			m.regs[MSC_IREG] &^= MSC_IREG_DATA_TRAN_DONE
			m.regs[MSC_IREG] |= MSC_IREG_TIME_OUT_READ
			return
		}

		m.fifoBuffer = m.diskImage[byteOffset:end]
		m.fifoIndex = 0

		m.regs[MSC_STAT] |= MSC_STAT_DATA_TRAN_DONE
		m.regs[MSC_IREG] |= MSC_IREG_DATA_TRAN_DONE

	default:
		m.setShortResponse(defaultStatus)
	}
}

func (m *MSC) setSDHCCSD() {
	sectors := uint64(len(m.diskImage)) / 512

	// SDHC CSD v2.0 represents capacity in units of 1024 sectors,
	// equivalent to 512 KiB.
	capacityUnits := sectors / 1024
	if capacityUnits == 0 {
		capacityUnits = 1
	}

	cSize := capacityUnits - 1

	// C_SIZE is 22 bits in CSD version 2.0.
	if cSize > 0x3FFFFF {
		cSize = 0x3FFFFF
	}

	var csd [16]byte

	// CSD_STRUCTURE = 1, indicating CSD version 2.0.
	csd[0] = 0x40

	// Typical timing and transfer-rate fields.
	csd[1] = 0x0E // TAAC
	csd[2] = 0x00 // NSAC
	csd[3] = 0x32 // TRAN_SPEED: 25 MHz

	// CCC and READ_BL_LEN.
	// READ_BL_LEN is fixed at 512 bytes for SDHC.
	csd[4] = 0x5B
	csd[5] = 0x59

	// C_SIZE occupies bits 69:48.
	csd[7] |= byte((cSize >> 16) & 0x3F)
	csd[8] = byte(cSize >> 8)
	csd[9] = byte(cSize)

	// Erase and write-protection related defaults.
	csd[10] = 0x7F
	csd[11] = 0x80
	csd[12] = 0x0A
	csd[13] = 0x40

	// CRC is normally ignored by the host controller emulator.
	csd[14] = 0x00
	csd[15] = 0x01

	m.setLongResponse(
		binary.BigEndian.Uint32(csd[0:4]),
		binary.BigEndian.Uint32(csd[4:8]),
		binary.BigEndian.Uint32(csd[8:12]),
		binary.BigEndian.Uint32(csd[12:16]),
	)
}

func (m *MSC) status() uint32 {
	value := m.regs[MSC_STAT]
	value &^= MSC_STAT_IS_RESETTING

	value |= MSC_STAT_PRG_DONE

	if m.fifoIndex >= len(m.fifoBuffer) {
		value |= MSC_STAT_DATA_FIFO_EMPTY
	}

	return value
}

var _ Device = (*MSC)(nil)
