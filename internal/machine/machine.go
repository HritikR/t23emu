package machine

import (
	"fmt"
	"os"

	"github.com/HritikR/t23emu/internal/bus"
	"github.com/HritikR/t23emu/internal/cpu"
	"github.com/HritikR/t23emu/internal/device"
	"github.com/HritikR/t23emu/internal/fs"
	"github.com/HritikR/t23emu/internal/memory"
	"github.com/HritikR/t23emu/internal/sensor"
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

	// SFCIRQ is the INTC hardware bit that maps to Linux IRQ 15. The
	// decompressed jz-sfc platform resource lists IRQ 15, and the Ingenic
	// bank-0 dispatcher adds 8 to the pending bit index before do_IRQ().
	SFCIRQ uint8 = 7

	// MSC0IRQ is the INTC hardware bit for the jzmmc_v1.2 controller. The
	// decompressed kernel's platform resource lists Linux IRQ 45, and the
	// Ingenic bank dispatcher adds 8 to the hardware bit before do_IRQ().
	MSC0IRQ uint8 = 37

	// UART IRQs. UART1 (ttyS1) maps to Linux IRQ 58.
	// Linux IRQ = INTC HW Bit + 8 => UART1IRQ = 58 - 8 = 50.
	UART0IRQ uint8 = 49 // Linux IRQ 57
	UART1IRQ uint8 = 50 // Linux IRQ 58
	UART2IRQ uint8 = 51 // Linux IRQ 59

	// I2C0IRQ is the INTC bit the vendor kernel unmasks for the hardware
	// I2C0 controller at 0x10050000.
	I2C0IRQ uint8 = 60

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

	// ISP blocks used by tx-isp-t23.ko. These are mapped explicitly so
	// camera-driver register access is visible separately from the broad
	// peripheral catch-all.
	ISPCoreStart uint32 = 0x13300000
	ISPCoreEnd   uint32 = 0x1330FFFF
	ISPIVDCStart uint32 = 0x13200000
	ISPIVDCEnd   uint32 = 0x1320FFFF
	ISPVICStart  uint32 = 0x133E0000
	ISPVICEnd    uint32 = 0x133EFFFF
	ISPCSIStart  uint32 = 0x10023000
	ISPCSIEnd    uint32 = 0x10023FFF

	// SFCStart is the serial flash controller.
	SFCStart uint32 = 0x13440000
	SFCEnd   uint32 = 0x1344FFFF

	// GMACStart is the Ethernet MAC and MDIO controller.
	GMACStart uint32 = 0x134B0000
	GMACEnd   uint32 = 0x134BFFFF

	DWC2Start uint32 = 0x13500000
	DWC2End   uint32 = 0x1350FFFF

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

	MSC0Start uint32 = 0x13450000
	MSC0End   uint32 = 0x1345FFFF
)

type Machine struct {
	CPU *cpu.CPU

	RAM *memory.RAM

	ROM *device.ROM

	// romData holds the original firmware image for re-loading on reboot.
	romData []byte

	// splBoot indicates the firmware is an Ingenic SPL image, which
	// needs stack/return-address setup on every boot.
	splBoot bool

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
	TCU *device.TCU

	// OST is the Linux OS timer clocksource and tick block.
	OST *device.OST

	// GPIO is the pin multiplexing and direction block.
	GPIO *device.RegisterBlock

	// I2C0 is the hardware I2C controller.
	I2C0 *device.I2C

	// DDRC is the DDR memory controller block.
	DDRC *device.RegisterBlock

	// DDRP is the DDR PHY block.
	DDRP *device.RegisterBlock

	// SFC is the serial flash controller.
	SFC *device.SFC

	// GMAC is the Ethernet MAC and MDIO controller.
	GMAC *device.RegisterBlock

	DWC2 *device.RegisterBlock

	// EFUSE is the one-time-programmable fuse controller.
	EFUSE *device.RegisterBlock

	MSC *device.MSC

	// ISP register windows used by the vendor tx_isp_t23 module.
	ISPCore *device.RegisterBlock
	ISPIVDC *device.RegisterBlock
	ISPVIC  *device.RegisterBlock
	ISPCSI  *device.RegisterBlock

	// Periph is the catch-all covering peripherals with no model yet.
	// Anything it records is a peripheral the emulator still needs.
	Periph *device.RegisterBlock

	Bus *bus.Bus

	// BootROMReturn is the address that signals a return to the boot ROM.
	BootROMReturn uint32
}

// Option configures a Machine instance.
type Option func(*MachineOptions)

type MachineOptions struct {
	DisableSDCard bool
	SDCardImage   []byte
	DisableRTSync bool
}

// WithSDCardImage sets a custom SD card disk image.
func WithSDCardImage(image []byte) Option {
	return func(o *MachineOptions) {
		o.SDCardImage = image
	}
}

// WithDisableSDCard disables SD card presence in the MSC controller.
func WithDisableSDCard() Option {
	return func(o *MachineOptions) {
		o.DisableSDCard = true
	}
}

// WithDisableRTSync disables real-time sync for the idle-loop
// fast-forward. By default the emulator throttles idle-loop skipping
// so that simulated time matches wall-clock time once the system is
// idle, keeping interactive timeouts (e.g. login prompts) accurate.
// Disabling it gives maximum speed at the cost of those timeouts
// firing too quickly.
func WithDisableRTSync() Option {
	return func(o *MachineOptions) {
		o.DisableRTSync = true
	}
}

// New creates a new T23 emulator machine.
func New(ramSize uint32, romData []byte, sfcSize uint32, opts ...Option) *Machine {
	options := &MachineOptions{}
	for _, opt := range opts {
		opt(options)
	}

	if sfcSize == 0 {
		sfcSize = 8 * 1024 * 1024
	}

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
	sc2336 := sensor.NewSC2336()
	i2c0.AttachDevice(0x30, sc2336)
	i2c0.AttachDevice(0x36, sc2336)
	b.Map(I2C0Start, I2C0End, i2c0)

	ddrc := device.NewDDRC()
	b.Map(DDRCStart, DDRCEnd, ddrc)

	ddrp := device.NewDDRP()
	b.Map(DDRPStart, DDRPEnd, ddrp)

	var cardPresent bool
	var mscDiskImage []byte

	if options.DisableSDCard {
		cardPresent = false
		mscDiskImage = nil
	} else if options.SDCardImage != nil {
		cardPresent = true
		mscDiskImage = options.SDCardImage
	} else {
		cardPresent = true
		mscDiskImage = fs.CreateEmptyFAT32Image(131072)
	}

	msc0 := device.NewMSC("MSC0", cardPresent, mscDiskImage)
	b.Map(MSC0Start, MSC0End, msc0)

	sfc := device.NewSFC(romData, sfcSize)
	b.Map(SFCStart, SFCEnd, sfc)
	traceSFCIRQ := os.Getenv("T23EMU_TRACE_SFC_IRQ") != ""
	traceSFCIRQLines := 0
	traceSFCIRQf := func(format string, args ...any) {
		if !traceSFCIRQ || traceSFCIRQLines >= 1000 {
			return
		}
		traceSFCIRQLines++
		fmt.Fprintf(os.Stderr, "[sfc-irq] "+format+"\n", args...)
	}
	sfcIRQPendingMask := uint32(1) << (SFCIRQ % 32)
	var lastSFCIRQPend uint32
	sfc.DMAWrite = func(addr uint32, data []byte) {
		for i, value := range data {
			target := addr + uint32(i)
			if target >= ram.Size() {
				break
			}
			ram.Write8(target, value)
		}
		traceSFCIRQf("dma-write addr=0x%08x len=%d data=% x", addr, len(data), data)
	}
	sfc.Interrupt = func(assert bool) {
		if assert {
			intc.Assert(SFCIRQ)
		} else {
			intc.Deassert(SFCIRQ)
		}
		traceSFCIRQf("line assert=%v irq=%d raw=0x%08x pending=%d", assert, SFCIRQ, intc.RawPending(), intc.Pending())
	}
	i2c0.Interrupt = func(assert bool) {
		if assert {
			intc.Assert(I2C0IRQ)
		} else {
			intc.Deassert(I2C0IRQ)
		}
	}

	msc0.Interrupt = func(assert bool) {
		if assert {
			intc.Assert(MSC0IRQ)
		} else {
			intc.Deassert(MSC0IRQ)
		}
	}

	uartIRQs := []uint8{UART0IRQ, UART1IRQ, UART2IRQ}
	for i, u := range uarts {
		irq := uartIRQs[i]
		u.Interrupt = func(assert bool) {
			if assert {
				intc.Assert(irq)
			} else {
				intc.Deassert(irq)
			}
		}
	}

	gmac := device.NewGMAC()
	b.Map(GMACStart, GMACEnd, gmac)

	dwc2 := device.NewDWC2()
	b.Map(DWC2Start, DWC2End, dwc2)

	efuse := device.NewEFUSE()
	b.Map(EFUSEStart, EFUSEEnd, efuse)

	ispCore := device.NewRegisterBlock("ISP_CORE", ISPCoreEnd-ISPCoreStart+1)
	b.Map(ISPCoreStart, ISPCoreEnd, ispCore)

	ispIVDC := device.NewRegisterBlock("ISP_IVDC", ISPIVDCEnd-ISPIVDCStart+1)
	b.Map(ISPIVDCStart, ISPIVDCEnd, ispIVDC)

	ispVIC := device.NewRegisterBlock("ISP_VIC", ISPVICEnd-ISPVICStart+1)
	b.Map(ISPVICStart, ISPVICEnd, ispVIC)

	ispCSI := device.NewRegisterBlock("ISP_CSI", ISPCSIEnd-ISPCSIStart+1)
	b.Map(ISPCSIStart, ISPCSIEnd, ispCSI)

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
	sfc.TraceContext = func() (uint32, uint64) {
		return c.PC, c.Cycles
	}

	cpuCycles = func() uint64 { return c.Cycles }

	// The periodic tick comes from the OST compare the kernel programmed,
	// not from a fixed cycle count. Driving it from the device keeps the
	// interrupt and the OSTFR flag the handler dispatches on in step: a
	// tick raised here that the handler cannot then see in OSTFR advances
	// nothing, and jiffies stay frozen.
	assertTimer := func() {
		if ost.OST1Expired() {
			intc.Assert(OSTIRQ)
		} else {
			intc.Deassert(OSTIRQ)
		}
	}
	c.InterruptPending = func() uint32 {
		assertTimer()
		rawPending := intc.RawPending()
		if intc.Pending() != 0 {
			if rawPending&sfcIRQPendingMask != 0 {
				if rawPending != lastSFCIRQPend {
					traceSFCIRQf("cpu-ip2 raw=0x%08x status=0x%08x cause=0x%08x", rawPending, c.CP0[cpu.CP0_STATUS], c.CP0[cpu.CP0_CAUSE])
					lastSFCIRQPend = rawPending
				}
			} else if lastSFCIRQPend != 0 {
				traceSFCIRQf("cpu-ip2-clear raw=0x%08x status=0x%08x cause=0x%08x", rawPending, c.CP0[cpu.CP0_STATUS], c.CP0[cpu.CP0_CAUSE])
				lastSFCIRQPend = 0
			}
			return cpu.CAUSE_IP2
		}
		return 0
	}
	c.WakePending = func() bool {
		assertTimer()
		return intc.RawPending() != 0
	}
	c.NextWakeCycle = func() uint64 {
		ostNext := ost.NextExpiryCycle()
		wdtNext := tcu.WatchdogExpiryCycle()
		// Treat a stale OST expiry (already passed) as unarmed so
		// the watchdog deadline is used instead. This happens during
		// the reboot spin loop when the kernel has stopped acking OST
		// interrupts and nextCompare is stuck in the past.
		if ostNext != 0 && ostNext <= c.Cycles {
			ostNext = 0
		}
		if wdtNext != 0 && (ostNext == 0 || wdtNext < ostNext) {
			return wdtNext
		}
		return ostNext
	}

	// Real-time sync: throttle idle-loop fast-forward once simulated
	// time catches up to wall-clock time. The T23 CCLK is 1188 MHz,
	// matching the OST prescaler (792 cycles/tick * 1.5 MHz = 100 Hz).
	if !options.DisableRTSync {
		c.CyclesPerSec = 1_188_000_000
		c.RTSyncEnabled = true
	}

	// Watchdog: check expiry on every Step(). When the countdown
	// reaches zero, halt for reboot.
	tcu.OnWatchdogReset = func() {
		c.HaltWith(cpu.HaltWatchdogReset,
			"watchdog reset (system reboot)")
	}
	c.WatchdogCheck = func() bool {
		return tcu.WatchdogExpired()
	}

	if len(romData) > 0 {
		c.ResetPC = resetPC
		c.Reset()

		if isIngenicSPL {
			// The boot ROM hands control over with a stack below the
			// image and a return address pointing back into itself.
			// The stack pointer must be 8-byte aligned per the MIPS ABI.
			sp := 0x80000000 + SPLLoadAddress + uint32(len(romData)) + 0x4000
			sp = sp &^ 7
			c.WriteRegister(29, sp)
			c.WriteRegister(31, BootROMReturn)
		}
	}

	return &Machine{
		CPU:           c,
		RAM:           ram,
		ROM:           rom,
		romData:       romData,
		splBoot:       isIngenicSPL,
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
		DWC2:          dwc2,
		EFUSE:         efuse,
		MSC:           msc0,
		ISPCore:       ispCore,
		ISPIVDC:       ispIVDC,
		ISPVIC:        ispVIC,
		ISPCSI:        ispCSI,
		Periph:        periph,
		Bus:           b,
		BootROMReturn: BootROMReturn,
	}
}

// Reset resets the complete machine state.
func (m *Machine) Reset() {

	m.CPU.Reset()

}

// reboot resets the CPU and critical device state so the firmware
// re-executes from the reset vector, simulating a hardware watchdog
// reset.
func (m *Machine) reboot() {
	m.CPU.Reset()

	// Re-do boot setup (stack pointer, return address for SPL).
	if len(m.romData) > 0 && m.splBoot {
		sp := 0x80000000 + SPLLoadAddress + uint32(len(m.romData)) + 0x4000
		sp = sp &^ 7
		m.CPU.WriteRegister(29, sp)
		m.CPU.WriteRegister(31, BootROMReturn)
	}

	// Clear interrupt controller so stale IRQs don't fire during boot.
	m.INTC.Reset()

	// Reset OS timer so the kernel programs it from scratch.
	m.OST.Reset()

	// Clear watchdog state.
	m.TCU.Reset()

	// Clear SFC transfer state (flash contents persist).
	m.SFC.Reset()

	// Reset RT sync so the fast-forward runs at full speed during boot.
	m.CPU.ResetSyncClock()

	m.CPU.Running = true
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
// If maxCycles is 0, execution continues indefinitely until
// the CPU halts or is stopped.
func (m *Machine) Run(maxCycles uint64) uint64 {

	m.CPU.Running = true

	start := m.CPU.Cycles

	for m.CPU.Running {

		if maxCycles > 0 && m.CPU.Cycles-start >= maxCycles {
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

		// If the watchdog triggered a reset, reboot and continue.
		if !m.CPU.Running && m.CPU.HaltReason == cpu.HaltWatchdogReset {
			m.reboot()
			start = m.CPU.Cycles
			maxCycles = 0 // reboot runs indefinitely
		}
	}

	m.CPU.Stop()

	return m.CPU.Cycles - start
}
