package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/HritikR/t23emu/internal/machine"
)

func main() {
	romPath := flag.String("rom", "", "Path to the ROM binary image file")
	ramSize := flag.Uint("ram", 1048576, "RAM size in bytes")
	cycles := flag.Uint64("cycles", 100000, "Maximum instruction cycles to run")
	trace := flag.Bool("trace", false, "Enable instruction tracing")

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
	m.CPU.Trace = *trace

	fmt.Println("Starting execution...")
	executedCycles := m.Run(*cycles)
	fmt.Printf("Execution stopped after %d cycles.\n\n", executedCycles)

	fmt.Println("--- CPU State ---")
	fmt.Printf("PC: 0x%08X\n", m.CPU.PC)
	fmt.Printf("Cycles: %d\n", m.CPU.Cycles)
	fmt.Printf("Running: %v\n", m.CPU.Running)
	fmt.Printf("Halt Reason: %v\n", m.CPU.HaltReason)
	fmt.Println("Registers:")
	for i := 0; i < 32; i++ {
		fmt.Printf("  R%-2d ($%-4s): 0x%08X", i, getRegName(i), m.CPU.ReadRegister(uint8(i)))
		if (i+1)%4 == 0 {
			fmt.Println()
		}
	}

	fmt.Println("\n--- UART Console Output ---")
	output := m.UART.GetCapturedOutput()
	if output != "" {
		fmt.Println(output)
	} else {
		fmt.Println("<no output>")
	}
}

func getRegName(index int) string {
	names := []string{
		"zero", "at", "v0", "v1", "a0", "a1", "a2", "a3",
		"t0", "t1", "t2", "t3", "t4", "t5", "t6", "t7",
		"s0", "s1", "s2", "s3", "s4", "s5", "s6", "s7",
		"t8", "t9", "k0", "k1", "gp", "sp", "fp", "ra",
	}
	if index >= 0 && index < len(names) {
		return names[index]
	}
	return "unknown"
}
