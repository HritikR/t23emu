package machine

import (
	"fmt"
	"github.com/HritikR/t23emu/internal/bus"
	"github.com/HritikR/t23emu/internal/cpu"
	"github.com/HritikR/t23emu/internal/device"
	"github.com/HritikR/t23emu/internal/memory"
)

const (
	ROMStart uint32 = 0x1fc00000 // Physical ROM start

	// CPMStart covers only the clock and power block itself. A wider
	// window would swallow the interrupt controller and timer unit that sit
	// immediately above it, and misattribute their accesses to CPM.
	CPMStart uint32 = 0x10000000
	CPMEnd   uint32 = 0x10000FFF

	// INTCStart is the interrupt controller.
	INTCStart uint32 = 0x10001000
	INTCEnd   uint32 = 0x10001FFF

	// TCUStart covers the watchdog, timer/counter unit and OS timer, which
	// share one 4 KB window.
	TCUStart uint32 = 0x10002000
	TCUEnd   uint32 = 0x10002FFF

	// OSTStart is the OS timer block Linux uses as its clocksource.
	OSTStart uint32 = 0x12000000
	OSTEnd   uint32 = 0x12000FFF

	// OSTIRQ is the interrupt line the OS timer drives. The kernel
	// confirms it: the only bit it ever touches in the INTC mask
	// registers is bit 3.
	OSTIRQ uint8 = 3

	GPIOStart uint32 = 0x10010000
	GPIOEnd   uint32 = 0x1001FFFF

	// UARTStart is the physical base of UART0. The three ports are spaced
	// 0x1000 apart and are mapped individually, because firmware picks a
	// console port by base address and the SPL in particular uses UART1.
	UARTStart  uint32 = 0x10030000
	UARTStride uint32 = 0x1000
	UARTCount  int    = 3
	UARTEnd    uint32 = 0x1003FFFF

	// I2C0Start is the hardware I2C controller. Bus 1 on this board is
	// bit-banged over GPIO instead and needs no block of its own.
	I2C0Start uint32 = 0x10050000
	I2C0End   uint32 = 0x10050FFF

	// DDRCStart covers the DDR controller and PHY, which the SPL programs
	// before it can relocate anything into external memory.
	DDRCStart uint32 = 0x13010000
	DDRCEnd   uint32 = 0x1301FFFF

	// DDRPStart is the DDR PHY block.
	DDRPStart uint32 = 0x134F0000
	DDRPEnd   uint32 = 0x134FFFFF

	// SFCStart is the serial flash controller.
	SFCStart uint32 = 0x13440000
	SFCEnd   uint32 = 0x1344FFFF

	// GMACStart is the Ethernet MAC and MDIO controller.
	GMACStart uint32 = 0x134B0000
	GMACEnd   uint32 = 0x134BFFFF

	// EFUSEStart is the one-time-programmable fuse controller, which the
	// SPL reads for the chip identifier and DDR calibration values.
	EFUSEStart uint32 = 0x13540000
	EFUSEEnd   uint32 = 0x1354FFFF

	// PeriphStart and PeriphEnd bracket the whole on-chip peripheral
	// window. A catch-all register file is mapped across it, behind every
	// specific device, so that touching a peripheral the emulator does not
	// model yet reads back as a benign register rather than faulting and
	// killing the boot. Everything it absorbs is recorded and reported, so
	// the gap stays visible instead of being silently papered over.
	PeriphStart uint32 = 0x10000000
	PeriphEnd   uint32 = 0x13FFFFFF

	// SPLLoadAddress is the physical address at which the boot ROM places
	// byte zero of an SPL image.
	//
	// The image itself pins this down. A J instruction only replaces the
	// low 28 bits of the program counter, so every call and jump in the
	// image encodes its link address directly. Solving for the offset that
	// makes the image's JAL targets land on function prologues gives a
	// single sharp answer of 0x1000: at that offset 32 of 79 in-range
	// calls hit a prologue, against 13 for the next best candidate. It is
	// confirmed by the entry code's `j 0x1b40`, which then resolves to the
	// SPL's main routine rather than to the middle of a byte-swap loop.
	SPLLoadAddress uint32 = 0x00001000

	// SPLHeaderSize is the size of the Ingenic boot header that precedes
	// the first instruction.
	SPLHeaderSize uint32 = 0x800

	// BootROMReturn is the return address the boot ROM would leave in $ra.
	// It is deliberately not backed by any device: reaching it means the
	// firmware returned to its caller, which the machine reports rather
	// than letting the core execute whatever happens to be there.
	BootROMReturn uint32 = 0xbfc0f000
)

type Machine struct {
	CPU *cpu.CPU

	RAM *memory.RAM

	ROM *device.ROM

	// UART is UART0, retained as the conventional console handle.
	UART *device.UART

	// UARTs holds every serial port, since firmware may pick any of them
	// as its console.
	UARTs []*device.UART

	// CPM is the clock and power management block. The SPL programs the
	// PLLs through it and polls them for lock.
	CPM *device.RegisterBlock

	// INTC is the interrupt controller.
	INTC *device.INTC

	// TCU is the watchdog, timer/counter and OS timer block.
	TCU *device.RegisterBlock

	// OST is the Linux OS timer clocksource and tick block.
	OST *device.OST

	// GPIO is the pin multiplexing and direction block.
	GPIO *device.RegisterBlock

	// I2C0 is the hardware I2C controller.
	I2C0 *device.RegisterBlock

	// DDRC is the DDR memory controller block.
	DDRC *device.RegisterBlock

	// DDRP is the DDR PHY block.
	DDRP *device.RegisterBlock

	// SFC is the serial flash controller.
	SFC *device.SFC

	// GMAC is the Ethernet MAC and MDIO controller.
	GMAC *device.RegisterBlock

	// EFUSE is the one-time-programmable fuse controller.
	EFUSE *device.RegisterBlock

	// Periph is the catch-all covering peripherals with no model yet.
	// Anything it records is a peripheral the emulator still needs.
	Periph *device.RegisterBlock

	Bus *bus.Bus

	// BootROMReturn is the address that signals a return to the boot ROM.
	BootROMReturn uint32
}

// New creates a new T23 emulator machine.
func New(ramSize uint32, romData []byte) *Machine {

	// An Ingenic boot image carries a 0x800-byte header whose signature
	// sits at offset 4. The boot ROM strips it and runs the code that
	// follows, so the emulator has to load such an image differently from
	// a raw ROM.
	isIngenicSPL := len(romData) > int(SPLHeaderSize) &&
		romData[4] == 0x02 &&
		romData[5] == 0x55 &&
		romData[6] == 0xAA &&
		romData[7] == 0x55 &&
		romData[8] == 0xAA

	// An SPL executes in place, so RAM has to be large enough to hold it.
	if isIngenicSPL && uint32(len(romData)) > ramSize {
		ramSize = uint32(len(romData))
	}

	ram := memory.NewRAM(ramSize)

	b := bus.New()

	b.Map(0x00000000, ramSize-1, ram)

	// Peripherals. These are register files rather than zero stubs: a
	// driver that configures a peripheral and reads back its own settings
	// gets the value it wrote, which a stub returning zero would break.
	cpm := device.NewCPM()
	b.Map(CPMStart, CPMEnd, cpm)

	intc := device.NewINTC()
	b.Map(INTCStart, INTCEnd, intc)

	// The timer block needs a tick source, but the CPU that provides it
	// does not exist until the bus is fully populated. The indirection
	// through cpuCycles closes that loop.
	var cpuCycles func() uint64
	tcu := device.NewTCU(func() uint64 {
		if cpuCycles == nil {
			return 0
		}
		return cpuCycles()
	})
	b.Map(TCUStart, TCUEnd, tcu)

	ost := device.NewOST(func() uint64 {
		if cpuCycles == nil {
			return 0
		}
		return cpuCycles()
	})
	b.Map(OSTStart, OSTEnd, ost)

	gpio := device.NewRegisterBlock("GPIO", GPIOEnd-GPIOStart+1)
	b.Map(GPIOStart, GPIOEnd, gpio)

	uarts := make([]*device.UART, UARTCount)
	for i := range uarts {
		base := UARTStart + uint32(i)*UARTStride
		uarts[i] = device.NewNamedUART(fmt.Sprintf("UART%d", i), nil)
		b.Map(base, base+UARTStride-1, uarts[i])
	}
	uart := uarts[0]

	i2c0 := device.NewI2C("I2C0")
	b.Map(I2C0Start, I2C0End, i2c0)

	ddrc := device.NewDDRC()
	b.Map(DDRCStart, DDRCEnd, ddrc)

	ddrp := device.NewDDRP()
	b.Map(DDRPStart, DDRPEnd, ddrp)

	sfc := device.NewSFC(romData)
	b.Map(SFCStart, SFCEnd, sfc)

	gmac := device.NewGMAC()
	b.Map(GMACStart, GMACEnd, gmac)

	efuse := device.NewEFUSE()
	b.Map(EFUSEStart, EFUSEEnd, efuse)

	// Mapped last so that every specific device above takes precedence.
	periph := device.NewRegisterBlock("PERIPH", PeriphEnd-PeriphStart+1)
	b.Map(PeriphStart, PeriphEnd, periph)

	var rom *device.ROM

	resetPC := uint32(0xbfc00000)

	if len(romData) > 0 {
		if isIngenicSPL {
			// Copy the whole image, header included, to the load address
			// and enter just past the header. Keeping the header mapped
			// matches the hardware, where the boot ROM has already
			// written it to memory.
			for i, val := range romData {
				ram.Write8(SPLLoadAddress+uint32(i), val)
			}

			// Execute through kseg0, which maps to physical zero.
			resetPC = 0x80000000 + SPLLoadAddress + SPLHeaderSize
		} else {
			rom = device.NewROM(romData)
			b.Map(ROMStart, ROMStart+uint32(len(romData))-1, rom)
		}
	}

	c := cpu.New(b)

	cpuCycles = func() uint64 { return c.Cycles }

	// The periodic tick comes from the OST compare the kernel programmed,
	// not from a fixed cycle count. Driving it from the device keeps the
	// interrupt and the OSTFR flag the handler dispatches on in step: a
	// tick raised here that the handler cannot then see in OSTFR advances
	// nothing, and jiffies stay frozen.
	assertTimer := func() {
		if ost.OST1Expired() {
			intc.Assert(OSTIRQ)
		}
	}
	c.InterruptPending = func() uint32 {
		assertTimer()
		if intc.Pending() != 0 {
			return cpu.CAUSE_IP2
		}
		return 0
	}
	c.WakePending = func() bool {
		assertTimer()
		return intc.RawPending() != 0
	}

	if len(romData) > 0 {
		c.ResetPC = resetPC
		c.Reset()

		if isIngenicSPL {
			// The boot ROM hands control over with a stack below the
			// image and a return address pointing back into itself.
			c.WriteRegister(29, 0x80000000+SPLLoadAddress+uint32(len(romData))+0x4000)
			c.WriteRegister(31, BootROMReturn)
		}
	}

	return &Machine{
		CPU:           c,
		RAM:           ram,
		ROM:           rom,
		UART:          uart,
		UARTs:         uarts,
		CPM:           cpm,
		INTC:          intc,
		TCU:           tcu,
		OST:           ost,
		GPIO:          gpio,
		I2C0:          i2c0,
		DDRC:          ddrc,
		DDRP:          ddrp,
		SFC:           sfc,
		GMAC:          gmac,
		EFUSE:         efuse,
		Periph:        periph,
		Bus:           b,
		BootROMReturn: BootROMReturn,
	}
}

// Reset resets the complete machine state.
func (m *Machine) Reset() {

	m.CPU.Reset()

}

// LoadProgram copies a program into RAM.
//
// address:
//
//	Starting memory address
//
// program:
//
//	Slice of 32-bit MIPS instructions
func (m *Machine) LoadProgram(
	address uint32,
	program []uint32,
) {

	for i, instruction := range program {

		offset := address + uint32(i*4)

		m.RAM.Write32(
			offset,
			instruction,
		)
	}
}

// Run executes the CPU for a number of cycles.
//
// This prevents tests from accidentally creating
// infinite loops.
func (m *Machine) Run(maxCycles uint64) uint64 {

	m.CPU.Running = true

	start := m.CPU.Cycles

	for m.CPU.Running {

		if m.CPU.Cycles-start >= maxCycles {
			break
		}

		// Catch a return to the boot ROM before the fetch, so it is
		// reported as a handoff rather than as a fault on an unmapped
		// address.
		if m.BootROMReturn != 0 && m.CPU.PC == m.BootROMReturn {
			m.CPU.HaltWith(cpu.HaltBootROMReturn,
				"firmware returned to boot ROM at 0x%08X", m.BootROMReturn)
			break
		}

		m.CPU.Step()
	}

	m.CPU.Stop()

	return m.CPU.Cycles - start
}
