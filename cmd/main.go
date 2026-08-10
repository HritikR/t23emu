// Copyright (c) 2026 Hritik R
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"runtime/pprof"
	"syscall"
	"time"

	"golang.org/x/term"

	"github.com/HritikR/t23emu/internal/cpu"
	"github.com/HritikR/t23emu/internal/debug"
	"github.com/HritikR/t23emu/internal/device"
	"github.com/HritikR/t23emu/internal/machine"
)

func main() {
	romPath := flag.String("rom", "", "Path to the ROM binary image file")
	ramSize := flag.Uint("ram", 64*1024*1024, "RAM size in bytes")
	flashSize := flag.Uint("flash-size", 8*1024*1024, "SPI flash size in bytes")
	cycles := flag.Uint64("cycles", 0, "Maximum instruction cycles to run; use 0 for unlimited")
	trace := flag.Bool("trace", false, "Enable instruction tracing to stderr")
	traceMMIO := flag.Bool("trace-mmio", false, "Trace peripheral register accesses to stderr")
	traceFrom := flag.Uint64("trace-from", 0, "Begin instruction tracing at this cycle")
	traceADE := flag.Bool("trace-ade", false, "Halt on the first userspace AdEL/AdES exception and dump instruction, registers, and history")
	liveUART := flag.Bool("live-uart", true, "Echo UART output live while the emulator runs")
	uartLimit := flag.Int("uart-limit", 16384, "Maximum captured UART bytes to print per port; use 0 for unlimited")
	history := flag.Bool("history", false, "Save and print history of the last 40 executed instructions on halt")
	haltPC := flag.Uint("halt-pc", 0, "Halt before executing the specified guest PC address; use 0 to disable")
	watchInterval := flag.Uint("watch", 0, "Print PC, Status, Cause, and the last 40 instructions to stderr every N seconds while running (0=off). Useful for diagnosing hangs without needing to interrupt.")
	gdbAddr := flag.String("gdb", "", "Enable GDB RSP server on port or address (e.g. :1234)")
	gdbWait := flag.Bool("gdb-wait", false, "Pause CPU execution on start until GDB connects")
	sdcardPath := flag.String("sdcard", "", "Path to optional SD card image file to mount")
	noSDCard := flag.Bool("no-sdcard", false, "Disable SD card presence in MSC controller")
	riTrace := flag.Bool("ri-trace", false, "Log Reserved Instruction exceptions to stderr")

	cpuProfile := flag.String("cpuprofile", "", "Write cpu profile to file")

	noRTSync := flag.Bool("no-rt-sync", false, "Disable wall-clock governor (login timeouts will fire too quickly)")

	flag.Parse()

	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating CPU profile: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting CPU profile: %v\n", err)
			os.Exit(1)
		}
		defer pprof.StopCPUProfile()
	}

	if *romPath == "" {
		fmt.Fprintf(os.Stderr, "Error: -rom argument is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	romData, err := os.ReadFile(*romPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to read ROM file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Loading ROM: %s (%d bytes)...\n", *romPath, len(romData))
	maxCyclesStr := fmt.Sprintf("%d", *cycles)
	if *cycles == 0 {
		maxCyclesStr = "unlimited"
	}
	fmt.Printf("Initializing T23 Machine (RAM: %d bytes, Flash: %d bytes, Max Cycles: %s)...\n", *ramSize, *flashSize, maxCyclesStr)

	var machineOpts []machine.Option
	if *noSDCard {
		machineOpts = append(machineOpts, machine.WithDisableSDCard())
		fmt.Println("SD Card: disabled (no card present)")
	} else if *sdcardPath != "" {
		sdData, err := os.ReadFile(*sdcardPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to read SD card image file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Loading SD card image: %s (%d bytes)...\n", *sdcardPath, len(sdData))
		machineOpts = append(machineOpts, machine.WithSDCardImage(sdData))
	} else {
		fmt.Println("SD Card: using default empty FAT32 image")
	}
	if *noRTSync {
		machineOpts = append(machineOpts, machine.WithDisableRTSync())
		fmt.Println("Real-time sync: disabled (wall-clock governor off)")
	}

	m := machine.New(uint32(*ramSize), romData, uint32(*flashSize), machineOpts...)
	m.CPU.RecordHistory = *history

	if *riTrace {
		m.CPU.EnableRITrace(os.Stdout)
	}

	if *haltPC != 0 {
		m.CPU.Breakpoints = map[uint32]bool{uint32(*haltPC): true}
	}
	if !*liveUART {
		for _, port := range m.UARTs {
			port.SetOutput(io.Discard)
		}
	}

	m.CPU.TraceADE = *traceADE

	var gdbServer *debug.Server
	if *gdbAddr != "" {
		gdbServer = debug.NewServer(*gdbAddr, m.CPU)
		if err := gdbServer.Start(*gdbWait); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting GDB server: %v\n", err)
			os.Exit(1)
		}
	}

	blocks := map[string]*device.RegisterBlock{
		"CPM":      m.CPM,
		"SYSCTL":   m.SYSCTL,
		"INTC":     m.INTC.RegisterBlock,
		"TCU":      m.TCU.RegisterBlock,
		"OST":      m.OST.RegisterBlock,
		"GPIO":     m.GPIO,
		"I2C0":     m.I2C0.RegisterBlock,
		"DDRC":     m.DDRC,
		"DDRP":     m.DDRP,
		"SFC":      m.SFC.RegisterBlock,
		"GMAC":     m.GMAC,
		"DWC2":     m.DWC2,
		"EFUSE":    m.EFUSE,
		"MSC":      m.MSC.RegisterBlock,
		"ISP_CORE": m.ISPCore,
		"ISP_IVDC": m.ISPIVDC,
		"ISP_VIC":  m.ISPVIC,
		"ISP_CSI":  m.ISPCSI,
		"PERIPH":   m.Periph,
	}

	if *traceMMIO {
		for _, block := range blocks {
			block.Trace = true
		}
	}

	fmt.Printf("Reset PC: 0x%08X\n", m.CPU.PC)
	fmt.Println("Starting execution...")

	// Enable Raw Mode for live UART right before starting execution
	rawModeActive := false
	var oldState *term.State
	stdinFd := int(os.Stdin.Fd())

	if *liveUART && term.IsTerminal(stdinFd) {
		state, err := term.MakeRaw(stdinFd)
		if err == nil {
			oldState = state
			rawModeActive = true
		}
	}

	var restoreOnce sync.Once
	restoreTerminal := func() {
		restoreOnce.Do(func() {
			if oldState != nil {
				_ = term.Restore(stdinFd, oldState)
			}
		})
	}
	defer restoreTerminal()

	userQuit := false
	quitEmulator := func() {
		userQuit = true
		restoreTerminal()
		if gdbServer != nil {
			gdbServer.Close()
		}
		m.CPU.Stop()
	}

	// Catch OS signals to ensure terminal state is restored on interruption
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		quitEmulator()
	}()

	// Redirect standard input to the UARTs
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := os.Stdin.Read(buf)
			if n > 0 {
				if rawModeActive {
					for i := 0; i < n; i++ {
						if buf[i] == 0x1d { // Ctrl+]
							fmt.Fprintln(os.Stderr, "\r\n[t23emu: terminated by Ctrl+]]")
							quitEmulator()
							return
						}
					}
				}
				for _, port := range m.UARTs {
					port.Feed(buf[:n])
				}
			}
			if err != nil {
				break
			}
		}
	}()

	var executedCycles uint64

	traceStart := *traceFrom
	if *cycles > 0 && traceStart > *cycles {
		traceStart = *cycles
	}

	// Live watchdog: every N seconds while the CPU is running, dump PC,
	// Status, Cause, and instruction history to stderr. Only enables
	// history recording if -history is also set; otherwise just prints
	// the register state (cheaper).
	if *watchInterval > 0 {
		if *history {
			m.CPU.RecordHistory = true
		}
		interval := time.Duration(*watchInterval) * time.Second
		go func() {
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			for range ticker.C {
				if !m.CPU.Running {
					return
				}
				fmt.Fprintf(os.Stderr, "\n[watch @ %d cycles] pc=0x%08X status=0x%08X cause=0x%08X EPC=0x%08X BadVAddr=0x%08X\n",
					m.CPU.Cycles, m.CPU.PC,
					m.CPU.CP0[cpu.CP0_STATUS], m.CPU.CP0[cpu.CP0_CAUSE],
					m.CPU.CP0[cpu.CP0_EPC], m.CPU.CP0[cpu.CP0_BADVADDR])
				fmt.Fprintln(os.Stderr, "--- last 60 instructions ---")
				entries := m.CPU.GetHistory()
				if len(entries) > 60 {
					entries = entries[len(entries)-60:]
				}
				for _, e := range entries {
					marker := " "
					if e.InDelaySlot {
						marker = "+"
					}
					fmt.Fprintf(os.Stderr, "  %s 0x%08X  %08X  %s\n",
						marker, e.PC, e.Instruction,
						cpu.Disassemble(e.Instruction, e.PC))
				}
			}
		}()
	}

	if *trace && traceStart > 0 {
		// Run without tracing up to the requested cycle, then trace the
		// rest. Tracing a long boot from cycle zero buries the
		// interesting part in setup code.
		executedCycles = m.Run(traceStart)
		m.CPU.Trace = true
		if *cycles > 0 {
			executedCycles += m.Run(*cycles - executedCycles)
		} else {
			executedCycles += m.Run(0)
		}
	} else {
		m.CPU.Trace = *trace
		if gdbServer != nil {
			for !userQuit && (gdbServer.IsConnected() || m.CPU.Running) {
				if m.CPU.Running {
					executedCycles += m.Run(*cycles)
				} else {
					time.Sleep(10 * time.Millisecond)
				}
			}
		} else {
			executedCycles = m.Run(*cycles)
		}
	}

	restoreTerminal()

	fmt.Printf("Execution stopped after %d cycles.\n\n", executedCycles)

	fmt.Println("--- CPU State ---")
	fmt.Printf("PC: 0x%08X\n", m.CPU.PC)
	fmt.Printf("Cycles: %d\n", m.CPU.Cycles)
	fmt.Printf("Running: %v\n", m.CPU.Running)
	fmt.Printf("Halt Reason: %v\n", m.CPU.HaltReason)
	if m.CPU.HaltDetail != "" {
		fmt.Printf("Detail: %s\n", m.CPU.HaltDetail)
	}
	fmt.Printf("Status: 0x%08X  Cause: 0x%08X  EPC: 0x%08X  BadVAddr: 0x%08X\n",
		m.CPU.CP0[cpu.CP0_STATUS], m.CPU.CP0[cpu.CP0_CAUSE],
		m.CPU.CP0[cpu.CP0_EPC], m.CPU.CP0[cpu.CP0_BADVADDR])

	fmt.Println("Registers:")
	for i := 0; i < 32; i++ {
		fmt.Printf("  R%-2d ($%-4s): 0x%08X", i, cpu.RegNames[i], m.CPU.ReadRegister(uint8(i)))
		if (i+1)%4 == 0 {
			fmt.Println()
		}
	}
	fmt.Printf("  HI: 0x%08X  LO: 0x%08X\n", m.CPU.HI, m.CPU.LO)

	fmt.Println("\n--- CP0 TLB Registers ---")
	fmt.Printf("  Index: 0x%08X  Random: 0x%08X  Wired: 0x%08X\n",
		m.CPU.CP0[cpu.CP0_INDEX], m.CPU.CP0[cpu.CP0_RANDOM], m.CPU.CP0[cpu.CP0_WIRED])
	fmt.Printf("  EntryHi: 0x%08X  EntryLo0: 0x%08X  EntryLo1: 0x%08X\n",
		m.CPU.CP0[cpu.CP0_ENTRYHI], m.CPU.CP0[cpu.CP0_ENTRYLO0], m.CPU.CP0[cpu.CP0_ENTRYLO1])
	fmt.Printf("  Context: 0x%08X  PageMask: 0x%08X\n",
		m.CPU.CP0[cpu.CP0_CONTEXT], m.CPU.CP0[cpu.CP0_PAGEMASK])
	fmt.Printf("  Hits: %d  Misses: %d  (ratio: %.1f%%)\n",
		m.CPU.TLBHits, m.CPU.TLBMisses,
		func() float64 {
			t := float64(m.CPU.TLBHits + m.CPU.TLBMisses)
			if t == 0 {
				return 0
			}
			return 100 * float64(m.CPU.TLBHits) / t
		}())
	fmt.Printf("  TLBWI: %d  TLBWR: %d\n", m.CPU.TLBWI, m.CPU.TLBWR)

	fmt.Println("\n--- TLB Entries ---")
	for i, entry := range m.CPU.TLB {
		if entry.EntryHi == 0 && entry.EntryLo0 == 0 && entry.EntryLo1 == 0 && entry.PageMask == 0 {
			continue
		}
		v0 := entry.EntryLo0 & 0x2 != 0
		d0 := entry.EntryLo0 & 0x4 != 0
		g0 := entry.EntryLo0 & 0x1 != 0
		v1 := entry.EntryLo1 & 0x2 != 0
		d1 := entry.EntryLo1 & 0x4 != 0
		g1 := entry.EntryLo1 & 0x1 != 0
		vpn2 := entry.EntryHi >> 13
		asid := entry.EntryHi & 0xFF
		pfn0 := (entry.EntryLo0 >> 6) & 0xFFFFF
		pfn1 := (entry.EntryLo1 >> 6) & 0xFFFFF
		flags := ""
		if v0 || v1 || g0 || g1 {
			flags = fmt.Sprintf(" V:%v%v D:%v%v G:%v%v", v0, v1, d0, d1, g0, g1)
		}
		fmt.Printf("  [%2d] VPN2=0x%08X ASID=0x%02X PFN0=0x%05X PFN1=0x%05X%s\n",
			i, vpn2, asid, pfn0, pfn1, flags)
	}

	if *history {
		fmt.Println("\n--- Last 40 Instructions ---")
		historyEntries := m.CPU.GetHistory()
		if len(historyEntries) == 0 {
			fmt.Println("  <no instructions executed>")
		} else {
			fmt.Printf("  %-12s %-12s %-12s %-40s %s\n", "Cycle", "PC", "Instruction", "Disassembly", "Memory Access")
			for _, entry := range historyEntries {
				marker := " "
				if entry.InDelaySlot {
					marker = "+"
				}
				memStr := ""
				if entry.MemAccess == "R" {
					memStr = fmt.Sprintf("Read 0x%08X from 0x%08X", entry.MemVal, entry.MemAddr)
				} else if entry.MemAccess == "W" {
					memStr = fmt.Sprintf("Write 0x%08X to 0x%08X", entry.MemVal, entry.MemAddr)
				}
				fmt.Printf("  [%010d]%s 0x%08X   0x%08X   %-40s %s\n",
					entry.Cycle, marker, entry.PC, entry.Instruction,
					cpu.Disassemble(entry.Instruction, entry.PC), memStr)
			}
		}
	}

	reportPeripherals(blocks)

	fmt.Println("\n--- UART Console Output ---")
	any := false
	for _, port := range m.UARTs {
		output := port.GetCapturedOutput()
		if output == "" {
			continue
		}
		truncated := false
		if *uartLimit > 0 && len(output) > *uartLimit {
			output = output[:*uartLimit]
			truncated = true
		}
		any = true
		fmt.Printf("[%s]\n%s\n", port.Name, output)
		if truncated {
			fmt.Printf("<UART output truncated to %d bytes>\n", *uartLimit)
		}
	}
	if !any {
		fmt.Println("<no output>")
	}
}

// reportPeripherals summarises what the firmware did to each peripheral.
//
// The hot-register list is the important half: a register read thousands
// of times is firmware spinning on a status bit, which is the usual way a
// boot stalls against an incompletely modelled peripheral.
func reportPeripherals(blocks map[string]*device.RegisterBlock) {

	const pollThreshold = 100

	for _, name := range []string{"CPM", "SYSCTL", "INTC", "TCU", "OST", "GPIO", "I2C0", "DDRC", "DDRP", "SFC", "GMAC", "DWC2", "EFUSE", "MSC", "PERIPH"} {
		block := blocks[name]

		written := block.Written()
		hot := block.HotRegisters(pollThreshold)

		if len(written) == 0 && len(hot) == 0 {
			continue
		}

		fmt.Printf("\n--- %s ---\n", name)

		for _, access := range written {
			fmt.Printf("  %-8s = 0x%08X  (%d writes, %d reads)\n",
				access.Name, access.Value, access.Writes, access.Reads)
		}

		for _, access := range hot {
			fmt.Printf("  POLLING %-8s read %d times, reads as 0x%08X\n",
				access.Name, access.Reads, access.Value)
		}
	}
}
