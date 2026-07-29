# t23emu

`t23emu` is a small emulator for firmware built for the Ingenic XBurst T23 SoC.

It is not trying to be QEMU. I am using it to run a real T23 firmware dump,
watch where the boot gets stuck, and then add only the hardware behavior needed
to get past that point.

Right now it gets through SPL, U-Boot, kernel entry, early Linux init, SFC flash
detection, and MTD partition setup.

## what works

The emulator currently gets through:

- SPL startup and DDR setup
- U-Boot relocation into RAM
- UART console output
- EFUSE board ID reporting as `T23N`
- SPI flash detection as `P25Q64H`
- kernel load and decompression from flash
- Linux delay calibration
- CP0 exceptions, interrupts, TLB mappings, and `wait`
- OST/TCU timing needed by early Linux
- SFC interrupt completion
- Linux JEDEC readback: `the id code = 856017, the flash name is P25Q64H`
- MTD partition creation from the kernel command line

It does not boot to userspace yet. The next big job is making the later storage
and peripheral paths believable enough for the kernel to mount rootfs and run
`/linuxrc`.

## hardware model

- MIPS/XBurst-style CPU interpreter
- branch delay slots
- CP0 exception, interrupt, timer, and TLB behavior
- RAM and ROM
- UART
- CPM
- INTC
- TCU and OST
- GPIO
- DDRC and DDRP
- I2C enable handshake
- SFC flash controller
- EFUSE
- GMAC/MDIO stub
- MSC/MMC stub
- generic register blocks for the boring parts

The SFC model is the most complete device at the moment because both U-Boot and
Linux lean on it heavily during boot.

## build

```sh
make build
```

Or without the Makefile:

```sh
go build -o bin/t23emu ./cmd/main.go
```

There is also a small image scanner:

```sh
make imgscan
```

## run

```sh
go run ./cmd/main.go -rom firmware_dump.bin
```

Or:

```sh
make run ROM=firmware_dump.bin
```

Useful flags:

```text
-rom <path>          firmware image to load
-ram <bytes>         RAM size, default 64 MiB
-flash-size <bytes>  SPI flash size, default 8 MiB
-cycles <count>      max cycles before stopping
-history             print the last 40 instructions on halt
-trace               print instruction trace
-trace-from <cycle>  start tracing at a specific cycle
-trace-mmio          print MMIO register accesses
-live-uart=false     collect UART output instead of printing it live
-uart-limit <bytes>  limit collected UART output, 0 means unlimited
```

Examples I use a lot:

```sh
go run ./cmd/main.go -rom firmware_dump.bin -history
go run ./cmd/main.go -rom firmware_dump.bin -trace -trace-from 578570690
go run ./cmd/main.go -rom firmware_dump.bin -trace-mmio -live-uart=false -uart-limit 4096
```

SFC debugging:

```sh
T23EMU_TRACE_SFC_IRQ=1 go run ./cmd/main.go -rom firmware_dump.bin
T23EMU_TRACE_SFC=1 T23EMU_TRACE_SFC_LINES=2000 go run ./cmd/main.go -rom firmware_dump.bin
```

## tests

Run everything:

```sh
make test
```

Useful focused runs:

```sh
go test ./internal/device -run 'TestSFC'
go test ./internal/device
go test ./internal/machine
go test ./internal/cpu -run 'TLB|Interrupt|CP0|WAIT'
```

If Go cannot write to the default build cache:

```sh
GOCACHE=$PWD/.gocache go test ./internal/device -run 'TestSFC'
```

## debugging notes

Most fixes in this repo start from the boot log.

When the firmware waits forever, check what it is polling. When Linux times out,
check which interrupt or status bit it expected. The emulator should grow from
those facts, not from guessing a whole SoC up front.

The main loop prints a peripheral summary at the end of a run. Hot registers are
usually the best clue.

## todo

- [x] boot SPL far enough to initialize DDR
- [x] run relocated U-Boot from RAM
- [x] capture UART output
- [x] report board info as `T23N`
- [x] emulate enough SFC for U-Boot flash probe
- [x] load and decompress the Linux kernel
- [x] handle early CP0 exceptions, interrupts, TLB, and `wait`
- [x] provide OST/TCU behavior for Linux timekeeping
- [x] route SFC completion interrupts correctly
- [x] return the right JEDEC ID for `P25Q64H`
- [x] reach Linux MTD partition creation
- [ ] make SFC flash reads solid enough for rootfs mounting
- [ ] improve MSC/MMC backing storage
- [ ] model GMAC reset and MDIO behavior more accurately
- [ ] model enough USB/DWC2 to avoid long timeout paths
- [ ] replace generic register stubs with real device behavior where boot needs it
- [ ] reduce boot time without lying to the kernel timers
- [ ] clean up old CPU tests and add more boot-level regression tests
- [ ] boot to `/linuxrc`

## license

MIT. See [LICENSE](LICENSE).
