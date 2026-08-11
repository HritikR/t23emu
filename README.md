# t23emu

I built this emulator to run real firmware dumps from Ingenic T23 chips. It is not QEMU. My goal was simple: load a raw firmware image, see where execution hangs, and write just enough hardware emulation to move past that spot.

Right now, my emulator boots all the way through the Linux kernel and launches userspace apps. The Tuya IoT camera software starts up and gets past TX ISP camera driver probing, ISP DMA channel setup, and MMC card initialization.

## what works

Here is what boots successfully:

* SPL startup and DDR memory setup
* UBoot relocation into RAM
* Serial console output over UART
* EFUSE board check showing T23N
* SPI flash detection as P25Q64H
* Linux kernel loading and decompression from flash
* Linux delay calibration at 1185.38 BogoMIPS
* CP0 exceptions, interrupts, TLB, and wait instruction handling
* OST and TCU hardware timers for early Linux timekeeping
* SFC flash interrupts and DMA memory transfers
* JEDEC code readback for P25Q64H flash
* MTD flash partition mapping from kernel arguments
* DWC2 USB OTG setup and USB ethernet gadget
* GMAC network reset and MDIO bus scanning
* Mounting squashfs root filesystem from NOR flash
* I2C bus detection and SC2336 camera sensor probing
* Driver initialization for TX ISP, audio codec, and VPU
* TX ISP core probing and DMA channel setup
* MSC MMC storage driver, SDHC card detection, command interrupts, and short-response reads
* Running linuxrc, building dev files with mdev, and setting up zram swap
* Tuya IoT userspace startup beyond the original TX ISP module-load panic

## hardware model

Here is what I have modeled so far:

* MIPS XBurst CPU core interpreter with branch delay slots
* CP0 register set for exceptions, interrupts, timers, and TLB
* System RAM and boot ROM
* UART serial interface
* Clock and power management unit (CPM)
* Interrupt controller (INTC)
* Timer units (TCU and OST)
* GPIO controller
* DDR memory controllers (DDRC and DDRP)
* I2C controller and SC2336 camera sensor
* SFC flash controller with SPI NOR flash emulation
* TX ISP core, IVDC, VIC, and CSI register windows
* DWC2 USB OTG controller
* MSC MMC controller with SDHC command, response, read-block, and interrupt support
* EFUSE hardware block
* GMAC network stub and MDIO bus
* GDB RSP server for remote debugging and stepping
* squashfs reader support
* Catch-all register stubs for remaining unused peripheral regions

The SFC flash controller is the most complete piece because UBoot and Linux read from flash constantly.

## build and run

Prebuilt binaries, when available, are published on the [GitHub releases page](https://github.com/HritikR/t23emu/releases).

You can build the project with make:

```sh
make build
```

Or use the Go tool directly:

```sh
go build ./cmd/main.go
```

I also wrote a small image scanner tool:

```sh
make imgscan
```

To run a firmware dump:

```sh
make run ROM=firmware_dump.bin
```

You can pass extra parameters to make if needed:

```sh
make run ROM=firmware_dump.bin RAM=67108864
make run ROM=firmware_dump.bin HISTORY=1
```

To enable SFC trace logs while running:

```sh
T23EMU_TRACE_SFC=1 make run ROM=firmware_dump.bin
```

## testing

Run the full test suite:

```sh
make test
```

Or run tests for specific packages:

```sh
go test ./internal/device
go test ./internal/machine
go test ./internal/cpu
```

If Go needs a local build cache directory:

```sh
GOCACHE=$PWD/.gocache go test ./internal/device
```

## how I debug

Most of my fixes start by reading the boot log.

When firmware gets stuck in a loop, I look at what register it reads over and over. When Linux hangs, I check which interrupt bit it expects. I only add hardware features when the boot log proves they are needed.

I also built a GDB RSP server into the emulator. You can attach GDB over a network port like :1234 to inspect registers, set breakpoints, or step through instructions single cycle at a time. You can also pause startup until GDB connects.

When execution stops, my emulator prints a register access summary. Hot registers in that summary usually show what peripheral needs attention next.

## todo

* [x] Boot SPL far enough to initialize DDR RAM
* [x] Run relocated UBoot from RAM
* [x] Capture serial console output
* [x] Report board identification as T23N
* [x] Emulate SFC flash controller for UBoot probe
* [x] Load and decompress Linux kernel image
* [x] Handle CP0 exceptions, interrupts, TLB, and wait instruction
* [x] Provide OST and TCU timers for Linux timekeeping
* [x] Route SFC completion interrupts correctly
* [x] Return correct JEDEC code for P25Q64H flash
* [x] Create Linux MTD partitions from command line
* [x] Mount squashfs root filesystem from NOR flash
* [x] Model GMAC reset and MDIO probe
* [x] Model DWC2 USB controller to prevent boot timeouts
* [x] Boot to linuxrc init process
* [x] Support MSC MMC controller and SDHC card detection
* [x] Route MSC completion interrupts to the kernel-derived INTC line
* [x] Handle MSC short responses when the guest reads MSC_RES as words
* [x] Run Tuya IoT userspace beyond TX ISP module initialization
* [x] Fix TX ISP kernel panic in camera driver module
* [ ] Replace register stubs with real hardware logic
* [ ] Speed up boot time without breaking kernel timers
* [ ] Clean up CPU tests and add boot regression tests

## license

MIT. See LICENSE file for details.
