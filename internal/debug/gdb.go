// Copyright (c) 2026 Hritik R
package debug

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/HritikR/t23emu/internal/cpu"
)

// Server handles GDB Remote Serial Protocol (RSP) connections.
type Server struct {
	addr      string
	listener  net.Listener
	cpu       *cpu.CPU
	conn      net.Conn
	reader    *bufio.Reader
	mu        sync.Mutex
	connected bool

	// Breakpoints maps virtual PC address to bool
	breakpoints map[uint32]bool
	waitChan    chan struct{}
	unblockOnce sync.Once
}

// NewServer creates a new GDB RSP server attached to a CPU.
func NewServer(addr string, cpuInst *cpu.CPU) *Server {
	s := &Server{
		addr:        addr,
		cpu:         cpuInst,
		breakpoints: make(map[uint32]bool),
		waitChan:    make(chan struct{}),
	}

	if cpuInst.Watchpoints == nil {
		cpuInst.Watchpoints = make(map[string]cpu.Watchpoint)
	}

	cpuInst.Breakpoints = s.breakpoints
	return s
}

// IsConnected returns whether a GDB client is currently connected.
func (s *Server) IsConnected() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.connected
}

// Close terminates the GDB server listener and active client connection.
func (s *Server) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.connected = false
	if s.conn != nil {
		_ = s.conn.Close()
		s.conn = nil
		s.reader = nil
	}
	if s.listener != nil {
		_ = s.listener.Close()
		s.listener = nil
	}
}

// Start opens the TCP listener and accepts incoming GDB connections.
func (s *Server) Start(wait bool) error {
	addr := s.addr
	if !strings.Contains(addr, ":") {
		addr = ":" + addr
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.addr, err)
	}
	s.listener = l

	fmt.Printf("[GDB] Server listening on %s... Waiting for GDB connection.\n", l.Addr().String())

	go s.acceptLoop()

	if wait {
		// Block execution until GDB connects and issues a continue or step command
		<-s.waitChan
		fmt.Println("[GDB] Continuing CPU execution after GDB attach.")
	} else {
		s.unblockWait()
	}

	return nil
}

func (s *Server) unblockWait() {
	s.unblockOnce.Do(func() {
		close(s.waitChan)
	})
}

func (s *Server) acceptLoop() {
	for {
		s.mu.Lock()
		l := s.listener
		s.mu.Unlock()

		if l == nil {
			return
		}

		conn, err := l.Accept()
		if err != nil {
			return
		}

		s.mu.Lock()
		if s.conn != nil {
			_ = s.conn.Close()
		}
		s.conn = conn
		s.reader = bufio.NewReader(conn)
		s.connected = true
		s.mu.Unlock()

		fmt.Printf("[GDB] Client connected from %s\n", conn.RemoteAddr().String())
		s.handleSession()
	}
}

func (s *Server) handleSession() {
	for {
		pkt, err := s.readPacket()
		if err != nil {
			s.mu.Lock()
			s.connected = false
			s.conn = nil
			s.reader = nil
			s.mu.Unlock()
			return
		}

		if pkt == "\x03" {
			s.cpu.Stop()
			s.sendPacket("S05")
			continue
		}

		s.sendACK()
		s.dispatchPacket(pkt)
	}
}

func (s *Server) readPacket() (string, error) {
	s.mu.Lock()
	r := s.reader
	s.mu.Unlock()

	if r == nil {
		return "", io.EOF
	}

	for {
		b, err := r.ReadByte()
		if err != nil {
			return "", err
		}

		// Wait for start of packet '$'
		if b == '$' {
			break
		}
		// Interrupt character 0x03 sent by GDB during continue
		if b == 0x03 {
			return "\x03", nil
		}
	}

	var buf strings.Builder
	for {
		b, err := r.ReadByte()
		if err != nil {
			return "", err
		}
		if b == '#' {
			break
		}
		buf.WriteByte(b)
	}

	// Read 2-byte checksum
	csStr := make([]byte, 2)
	_, err := io.ReadFull(r, csStr)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (s *Server) sendACK() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn != nil {
		_, _ = s.conn.Write([]byte{'+'})
	}
}

func (s *Server) sendPacket(data string) {
	checksum := CalculateChecksum(data)
	pkt := fmt.Sprintf("$%s#%02x", data, checksum)

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn != nil {
		_, _ = s.conn.Write([]byte(pkt))
	}
}

func CalculateChecksum(data string) uint8 {
	var sum uint8
	for i := 0; i < len(data); i++ {
		sum += data[i]
	}
	return sum
}

func (s *Server) dispatchPacket(pkt string) {
	if len(pkt) == 0 {
		return
	}

	cmd := pkt[0]
	body := pkt[1:]

	switch cmd {
	case '?':
		// Query halt status
		s.sendPacket("S05") // TRAP signal (SIGTRAP)

	case 'q':
		s.handleQuery(body)

	case 'g':
		// Read registers
		s.sendPacket(s.encodeRegisters())

	case 'G':
		// Write registers
		s.decodeRegisters(body)
		s.sendPacket("OK")

	case 'p':
		// Read single register
		regIdx, err := strconv.ParseUint(body, 16, 32)
		if err != nil {
			s.sendPacket("E01")
		} else {
			s.sendPacket(s.encodeSingleRegister(uint32(regIdx)))
		}

	case 'P':
		// Write single register: P<reg>=<val>
		parts := strings.Split(body, "=")
		if len(parts) == 2 {
			regIdx, err1 := strconv.ParseUint(parts[0], 16, 32)
			val, err2 := decodeHex32LE(parts[1])
			if err1 == nil && err2 == nil {
				s.writeSingleRegister(uint32(regIdx), val)
				s.sendPacket("OK")
				return
			}
		}
		s.sendPacket("E01")

	case 'm':
		// Read memory: m<addr>,<len>
		parts := strings.Split(body, ",")
		if len(parts) == 2 {
			addr, err1 := strconv.ParseUint(parts[0], 16, 32)
			length, err2 := strconv.ParseUint(parts[1], 16, 32)
			if err1 == nil && err2 == nil {
				s.sendPacket(s.readMemory(uint32(addr), uint32(length)))
				return
			}
		}
		s.sendPacket("E01")

	case 'M':
		// Write memory: M<addr>,<len>:<hex_bytes>
		colonIdx := strings.Index(body, ":")
		if colonIdx != -1 {
			addrLen := body[:colonIdx]
			hexBytes := body[colonIdx+1:]
			parts := strings.Split(addrLen, ",")
			if len(parts) == 2 {
				addr, err1 := strconv.ParseUint(parts[0], 16, 32)
				length, err2 := strconv.ParseUint(parts[1], 16, 32)
				if err1 == nil && err2 == nil {
					s.writeMemory(uint32(addr), uint32(length), hexBytes)
					s.sendPacket("OK")
					return
				}
			}
		}
		s.sendPacket("E01")

	case 'Z', 'z':
		// Breakpoint / Watchpoint insert/remove
		// Format: Z<type>,<addr>,<kind> or z<type>,<addr>,<kind>
		isInsert := (cmd == 'Z')
		parts := strings.Split(body, ",")
		if len(parts) >= 3 {
			zType := parts[0]
			addr, err1 := strconv.ParseUint(parts[1], 16, 32)
			length, err2 := strconv.ParseUint(parts[2], 16, 32)
			if err1 == nil && err2 == nil {
				vaddr := uint32(addr)
				kindLen := uint32(length)
				key := fmt.Sprintf("0x%08x", vaddr)

				switch zType {
				case "0", "1": // Software or Hardware Breakpoint
					if isInsert {
						s.breakpoints[vaddr] = true
					} else {
						delete(s.breakpoints, vaddr)
					}
					s.sendPacket("OK")
					return
				case "2": // Write Watchpoint
					if isInsert {
						s.cpu.Watchpoints[key] = cpu.Watchpoint{Addr: vaddr, Len: kindLen, Type: cpu.WatchWrite}
					} else {
						delete(s.cpu.Watchpoints, key)
					}
					s.sendPacket("OK")
					return
				case "3": // Read Watchpoint
					if isInsert {
						s.cpu.Watchpoints[key] = cpu.Watchpoint{Addr: vaddr, Len: kindLen, Type: cpu.WatchRead}
					} else {
						delete(s.cpu.Watchpoints, key)
					}
					s.sendPacket("OK")
					return
				case "4": // Access Watchpoint
					if isInsert {
						s.cpu.Watchpoints[key] = cpu.Watchpoint{Addr: vaddr, Len: kindLen, Type: cpu.WatchAccess}
					} else {
						delete(s.cpu.Watchpoints, key)
					}
					s.sendPacket("OK")
					return
				}
			}
		}
		s.sendPacket("")

	case 's':
		// Single step
		s.unblockWait()
		s.cpu.SingleStep = true
		s.cpu.Running = true
		s.cpu.HitWatchpoint = 0
		s.cpu.Step()
		s.cpu.SingleStep = false
		if s.cpu.HitWatchpoint != 0 {
			s.sendPacket(fmt.Sprintf("T05watch:%x;", s.cpu.HitWatchpoint))
			s.cpu.HitWatchpoint = 0
		} else {
			s.sendPacket("S05")
		}

	case 'c':
		// Continue execution until breakpoint, watchpoint, halt, or Ctrl+C interrupt
		s.unblockWait()
		s.cpu.Running = true
		s.cpu.HitWatchpoint = 0
		go func() {
			for s.cpu.Running {
				time.Sleep(5 * time.Millisecond)
			}
			if s.cpu.HitWatchpoint != 0 {
				s.sendPacket(fmt.Sprintf("T05watch:%x;", s.cpu.HitWatchpoint))
				s.cpu.HitWatchpoint = 0
			} else {
				s.sendPacket("S05")
			}
		}()

	case 'D':
		// Detach from target
		s.sendPacket("OK")
		s.mu.Lock()
		if s.conn != nil {
			_ = s.conn.Close()
			s.conn = nil
		}
		s.connected = false
		s.mu.Unlock()

	case 'k':
		// Kill target
		s.cpu.Stop()
		s.sendPacket("OK")

	default:
		s.sendPacket("")
	}
}

const targetXML = `<?xml version="1.0"?>
<!DOCTYPE target SYSTEM "gdb-target.dtd">
<target>
  <architecture>mips:isa32r2</architecture>
</target>`

func (s *Server) handleQuery(body string) {
	if strings.HasPrefix(body, "Supported") {
		s.sendPacket("PacketSize=1000;swbreak+;hwbreak+;qXfer:features:read+")
		return
	}
	if strings.HasPrefix(body, "Attached") {
		s.sendPacket("1") // Attached to an existing process
		return
	}
	if strings.HasPrefix(body, "Xfer:features:read:target.xml:") {
		params := strings.TrimPrefix(body, "Xfer:features:read:target.xml:")
		parts := strings.Split(params, ",")
		offset := 0
		length := len(targetXML)
		if len(parts) == 2 {
			if o, err := strconv.Atoi(parts[0]); err == nil {
				offset = o
			}
			if l, err := strconv.Atoi(parts[1]); err == nil {
				length = l
			}
		}
		if offset >= len(targetXML) {
			s.sendPacket("l")
			return
		}
		end := offset + length
		if end > len(targetXML) {
			end = len(targetXML)
		}
		s.sendPacket("l" + targetXML[offset:end])
		return
	}
	s.sendPacket("")
}

// encodeRegisters serialises MIPS32 registers for GDB.
// Layout (38 registers total):
// 0-31: R0-R31
// 32: Status, 33: LO, 34: HI, 35: BadVAddr, 36: Cause, 37: PC (EPC/PC)
func (s *Server) encodeRegisters() string {
	var buf strings.Builder
	for i := 0; i < 32; i++ {
		buf.WriteString(encodeHex32LE(s.cpu.ReadRegister(uint8(i))))
	}
	buf.WriteString(encodeHex32LE(s.cpu.CP0[cpu.CP0_STATUS]))
	buf.WriteString(encodeHex32LE(s.cpu.LO))
	buf.WriteString(encodeHex32LE(s.cpu.HI))
	buf.WriteString(encodeHex32LE(s.cpu.CP0[cpu.CP0_BADVADDR]))
	buf.WriteString(encodeHex32LE(s.cpu.CP0[cpu.CP0_CAUSE]))
	buf.WriteString(encodeHex32LE(s.cpu.PC))
	return buf.String()
}

func (s *Server) decodeRegisters(hexData string) {
	if len(hexData) < 304 {
		return
	}
	for i := 0; i < 32; i++ {
		val, err := decodeHex32LE(hexData[i*8 : (i+1)*8])
		if err == nil {
			s.cpu.WriteRegister(uint8(i), val)
		}
	}
	if val, err := decodeHex32LE(hexData[32*8 : 33*8]); err == nil {
		s.cpu.CP0[cpu.CP0_STATUS] = val
	}
	if val, err := decodeHex32LE(hexData[33*8 : 34*8]); err == nil {
		s.cpu.LO = val
	}
	if val, err := decodeHex32LE(hexData[34*8 : 35*8]); err == nil {
		s.cpu.HI = val
	}
	if val, err := decodeHex32LE(hexData[35*8 : 36*8]); err == nil {
		s.cpu.CP0[cpu.CP0_BADVADDR] = val
	}
	if val, err := decodeHex32LE(hexData[36*8 : 37*8]); err == nil {
		s.cpu.CP0[cpu.CP0_CAUSE] = val
	}
	if val, err := decodeHex32LE(hexData[37*8 : 38*8]); err == nil {
		s.cpu.PC = val
		s.cpu.NextPC = val + 4
	}
}

func (s *Server) encodeSingleRegister(regIdx uint32) string {
	if regIdx < 32 {
		return encodeHex32LE(s.cpu.ReadRegister(uint8(regIdx)))
	}
	switch regIdx {
	case 32:
		return encodeHex32LE(s.cpu.CP0[cpu.CP0_STATUS])
	case 33:
		return encodeHex32LE(s.cpu.LO)
	case 34:
		return encodeHex32LE(s.cpu.HI)
	case 35:
		return encodeHex32LE(s.cpu.CP0[cpu.CP0_BADVADDR])
	case 36:
		return encodeHex32LE(s.cpu.CP0[cpu.CP0_CAUSE])
	case 37:
		return encodeHex32LE(s.cpu.PC)
	}
	return "00000000"
}

func (s *Server) writeSingleRegister(regIdx uint32, val uint32) {
	if regIdx < 32 {
		s.cpu.WriteRegister(uint8(regIdx), val)
		return
	}
	switch regIdx {
	case 32:
		s.cpu.CP0[cpu.CP0_STATUS] = val
	case 33:
		s.cpu.LO = val
	case 34:
		s.cpu.HI = val
	case 35:
		s.cpu.CP0[cpu.CP0_BADVADDR] = val
	case 36:
		s.cpu.CP0[cpu.CP0_CAUSE] = val
	case 37:
		s.cpu.PC = val
		s.cpu.NextPC = val + 4
	}
}

func (s *Server) readMemory(addr uint32, length uint32) string {
	var buf strings.Builder
	for i := uint32(0); i < length; i++ {
		vaddr := addr + i
		val := s.cpu.Bus.Read8(vaddr)
		buf.WriteString(fmt.Sprintf("%02x", val))
	}
	return buf.String()
}

func (s *Server) writeMemory(addr uint32, length uint32, hexData string) {
	if uint32(len(hexData)) < length*2 {
		return
	}
	for i := uint32(0); i < length; i++ {
		bVal, err := strconv.ParseUint(hexData[i*2:(i+1)*2], 16, 8)
		if err == nil {
			vaddr := addr + i
			s.cpu.Bus.Write8(vaddr, byte(bVal))
		}
	}
}

func encodeHex32LE(val uint32) string {
	return fmt.Sprintf("%02x%02x%02x%02x", byte(val), byte(val>>8), byte(val>>16), byte(val>>24))
}

func decodeHex32LE(s string) (uint32, error) {
	if len(s) != 8 {
		return 0, fmt.Errorf("invalid hex length")
	}
	b0, err0 := strconv.ParseUint(s[0:2], 16, 8)
	b1, err1 := strconv.ParseUint(s[2:4], 16, 8)
	b2, err2 := strconv.ParseUint(s[4:6], 16, 8)
	b3, err3 := strconv.ParseUint(s[6:8], 16, 8)
	if err0 != nil || err1 != nil || err2 != nil || err3 != nil {
		return 0, fmt.Errorf("invalid hex format")
	}
	return uint32(b0) | (uint32(b1) << 8) | (uint32(b2) << 16) | (uint32(b3) << 24), nil
}
