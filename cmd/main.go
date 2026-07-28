package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/HritikR/t23emu/internal/cpu"
	"github.com/HritikR/t23emu/internal/device"
	"github.com/HritikR/t23emu/internal/machine"
)

func main() {
	romPath := flag.String("rom", "", "Path to the ROM binary image file")
	ramSize := flag.Uint("ram", 64*1024*1024, "RAM size in bytes")
	cycles := flag.Uint64("cycles", 1500000000, "Maximum instruction cycles to run")
	trace := flag.Bool("trace", false, "Enable instruction tracing to stderr")
	traceMMIO := flag.Bool("trace-mmio", false, "Trace peripheral register accesses to stderr")
	traceFrom := flag.Uint64("trace-from", 0, "Begin instruction tracing at this cycle")
	liveUART := flag.Bool("live-uart", true, "Echo UART output live while the emulator runs")
	uartLimit := flag.Int("uart-limit", 16384, "Maximum captured UART bytes to print per port; use 0 for unlimited")

	flag.Parse()

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
	fmt.Printf("Initializing T23 Machine (RAM: %d bytes, Max Cycles: %d)...\n", *ramSize, *cycles)

	m := machine.New(uint32(*ramSize), romData)
	if !*liveUART {
		for _, port := range m.UARTs {
			port.SetOutput(io.Discard)
		}
	}

	// Redirect standard input to the UARTs
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := os.Stdin.Read(buf)
			if n > 0 {
				for _, port := range m.UARTs {
					port.Feed(buf[:n])
				}
			}
			if err != nil {
				break
			}
		}
	}()

	blocks := map[string]*device.RegisterBlock{
		"CPM":    m.CPM,
		"INTC":   m.INTC.RegisterBlock,
		"TCU":    m.TCU,
		"OST":    m.OST.RegisterBlock,
		"GPIO":   m.GPIO,
		"I2C0":   m.I2C0,
		"DDRC":   m.DDRC,
		"DDRP":   m.DDRP,
		"SFC":    m.SFC.RegisterBlock,
		"GMAC":   m.GMAC,
		"EFUSE":  m.EFUSE,
		"PERIPH": m.Periph,
	}

	if *traceMMIO {
		for _, block := range blocks {
			block.Trace = true
		}
	}

	fmt.Printf("Reset PC: 0x%08X\n", m.CPU.PC)
	fmt.Println("Starting execution...")

	var executedCycles uint64

	if *trace && *traceFrom > 0 {
		// Run without tracing up to the requested cycle, then trace the
		// rest. Tracing a long boot from cycle zero buries the
		// interesting part in setup code.
		executedCycles = m.Run(*traceFrom)
		m.CPU.Trace = true
		executedCycles += m.Run(*cycles - executedCycles)
	} else {
		m.CPU.Trace = *trace
		executedCycles = m.Run(*cycles)
	}

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

	for _, name := range []string{"CPM", "INTC", "TCU", "OST", "GPIO", "I2C0", "DDRC", "DDRP", "SFC", "GMAC", "EFUSE", "PERIPH"} {
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
