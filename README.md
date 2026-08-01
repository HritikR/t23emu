# t23emu

A small emulator for Ingenic T23 (XBurst MIPS) firmware. It's not QEMU. I built it to run a real T23 firmware dump, see where boot gets stuck, and add just enough hardware to get past that point.

Right now it boots all the way through the kernel into userspace. The Tuya IoT camera app starts, connects to WiFi, and initializes video/audio pipelines. It eventually hits a kernel panic in the TX-ISP camera driver module, but that's deep into normal operation.

## what works

Here's what the emulator gets through so far:

- SPL startup and DDR init
- U-Boot relocation into RAM
- UART console output
- EFUSE board ID shows up as `T23N`
- SPI flash detected as `P25Q64H`
- kernel loads and decompresses from flash
- Linux delay calibration (1185.38 BogoMIPS)
- CP0 exceptions, interrupts, TLB, and `wait`
- OST/TCU timing for early Linux
- SFC interrupt completion
- JEDEC readback: `the id code = 856017, the flash name is P25Q64H`
- MTD partitions from the kernel command line
- DWC2 USB OTG init and USB ethernet gadget
- GMAC reset and MDIO bus probe (no PHY, as expected)
- squashfs rootfs mounted from NOR flash
- I2C and SC2336 camera sensor detection
- Core driver probes finish (TX-ISP, audio codec, VPU, RTC)
- MSC/MMC card detection (SDHC, 3.69 GiB)
- `/linuxrc` runs, mdev populates `/dev`, zram swap set up
- Tuya IoT SDK initializes, WiFi connects, BLE starts
- DHCP lease acquired, video/audio ring buffers allocated

The emulator eventually hits a kernel panic inside the TX-ISP camera driver (`insmod tx_isp_t23`) — a null pointer dereference at a virtual address. That's the current wall.

## hardware model

- MIPS/XBurst CPU interpreter (UserLocal, COP1, unaligned/rotate instructions)
- branch delay slots
- CP0 (exceptions, interrupts, timer, TLB, UserLocal, RDHWR)
- RAM and ROM
- UART
- CPM
- INTC
- TCU and OST
- GPIO
- DDRC and DDRP
- I2C controller and SC2336 camera sensor
- SFC flash controller
- SPI NOR flash (P25Q64H, JEDEC readback, DMA transfers)
- DWC2 USB OTG controller
- MSC/MMC controller with SDHC card emulation
- EFUSE
- GMAC/MDIO stub
- squashfs filesystem reader
- generic register blocks for the boring stuff

The SFC model is the most fleshed out because both U-Boot and Linux hit it hard during boot.

## build

```sh
make build
```

Or just:

```sh
go build -o bin/t23emu ./cmd/main.go
```

There's also a small image scanner:

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

Flags:

```text
-rom <path>          firmware image to load
-ram <bytes>         RAM size, default 64 MiB
-flash-size <bytes>  SPI flash size, default 8 MiB
-cycles <count>      max cycles before stopping, 0 = unlimited (default 0)
-history             print the last 40 instructions on halt
-trace               print instruction trace
-trace-from <cycle>  start tracing at a specific cycle
-trace-mmio          print MMIO register accesses
-live-uart=false     collect UART output instead of printing it live
-uart-limit <bytes>  limit collected UART output, 0 = unlimited
-gdb <port>          enable GDB RSP server (e.g. :1234)
-gdb-wait            pause on start until GDB connects
```

Some examples I use a lot:

```sh
go run ./cmd/main.go -rom firmware_dump.bin -history
go run ./cmd/main.go -rom firmware_dump.bin -gdb :1234 -gdb-wait
go run ./cmd/main.go -rom firmware_dump.bin -trace -trace-from 578570690
go run ./cmd/main.go -rom firmware_dump.bin -trace-mmio -live-uart=false -uart-limit 4096
go run ./cmd/main.go -rom openipc-t23n-nor-lite.bin
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

Some focused runs:

```sh
go test ./internal/device -run 'TestSFC'
go test ./internal/device
go test ./internal/machine
go test ./internal/cpu -run 'TLB|Interrupt|CP0|WAIT'
```

If Go can't write to the default build cache:

```sh
GOCACHE=$PWD/.gocache go test ./internal/device -run 'TestSFC'
```

## debugging notes

Most fixes start from the boot log.

When the firmware hangs, check what it's polling. When Linux times out, check which interrupt or status bit it was waiting for. The emulator should grow from those facts — not from guessing a whole SoC up front.

The main loop prints a peripheral summary at the end of a run. Hot registers are usually the best clue.

## todo

- [x] boot SPL far enough to init DDR
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
- [x] make SFC flash reads solid enough for rootfs mounting
- [x] model GMAC reset and MDIO behavior more accurately
- [x] model enough USB/DWC2 to avoid long timeout paths
- [x] boot to `/linuxrc`
- [x] MSC/MMC with SDHC card detection
- [x] Tuya IoT userspace init and WiFi connection
- [ ] fix TX-ISP kernel panic (null deref in camera driver module)
- [ ] replace generic register stubs with real device behavior where boot needs it
- [ ] reduce boot time without lying to the kernel timers
- [ ] clean up old CPU tests and add more boot-level regression tests

## license

MIT. See [LICENSE](LICENSE).
