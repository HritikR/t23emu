# t23emu

`t23emu` is a Go emulator for firmware targeting the Ingenic XBurst T23 SoC.

The project is focused on booting and debugging real firmware images while keeping
the hardware model small enough to understand and extend. It currently reaches
U-Boot, loads the Linux kernel from emulated SPI flash, and runs into early Linux
init.

## current boot status

The emulator currently gets through:

- Ingenic SPL startup
- PLL and DDR setup
- U-Boot relocation to RAM
- board info reporting as `T23N`
- SPI flash probing as `W25Q64`
- kernel image load from flash
- Linux decompression and entry
- Linux delay calibration
- early timer interrupts
- early CP0 TLB mapped-space handling
- CPU `wait` wakeup on interrupt
- 8-byte MIPS ABI stack alignment for Ingenic SPLs

The current known boot limit is after early Linux device init. It has not yet
booted fully to a userspace shell.

## features

- MIPS/XBurst-style CPU interpreter
- branch delay slot handling
- CP0 exception, interrupt, and TLB support
- RAM, ROM, UART, CPM, INTC, TCU, OST, GPIO, DDR, SFC, GMAC, and EFUSE models
- SPI flash reads from a firmware image
- UART console output
- MMIO reporting for device bring-up
- helper image scanner at `cmd/imgscan`

## build

```sh
go build -o bin/t23emu ./cmd/main.go
go build -o bin/imgscan ./cmd/imgscan
```

Or:

```sh
make build
```

## run

```sh
go run ./cmd/main.go -rom firmware_dump.bin
```

With the Makefile:

```sh
make run
```

Useful flags:

```sh
-rom <path>          firmware image to load
-ram <bytes>         RAM size
-cycles <count>      maximum cycles to execute
-trace-mmio          print MMIO accesses
-live-uart=false     collect UART output instead of printing live
-uart-limit <bytes>  limit collected UART output
```

Example MMIO run:

```sh
go run ./cmd/main.go -rom firmware_dump.bin -cycles 2000000000 -trace-mmio -live-uart=false -uart-limit 4096
```

## tests

Focused tests are useful while CPU behavior is still being brought up:

```sh
go test ./internal/cpu -run "TLB|ExternalInterrupt|CP0|COP0|Kseg2|WAIT"
go test ./internal/machine -run "TimerInterruptWakesWAIT"
go test ./internal/device
```

Some legacy full CPU tests still need cleanup around older direct-execute
expectations.

## releases

Tagged pushes matching `v*` build release archives for Linux, macOS, and
Windows through GitHub Actions. Release notes are generated automatically from
the commit history.

## todo

- [x] boot Ingenic SPL
- [x] run U-Boot from RAM
- [x] report existing board info as `T23N`
- [x] emulate enough SPI flash for U-Boot to load the kernel
- [x] boot the Linux kernel image
- [x] provide OST/TCU timers for delay calibration
- [x] handle early timer interrupts
- [x] support CP0 TLB mappings used by early Linux
- [x] wake the CPU from `wait` on interrupts
- [ ] boot fully to userspace shell
- [ ] model real TCU/OST compare interrupt timing
- [ ] model more peripherals instead of generic register stubs
- [ ] clean up legacy CPU tests

## license

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
