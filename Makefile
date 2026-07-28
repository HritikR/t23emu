APP := t23emu
SHELL := cmd.exe
.SHELLFLAGS := /C

ROM ?= firmware_dump.bin
RAM ?= 67108864
CYCLES ?= 200000000
UART_LIMIT ?= 4096

BIN_DIR := bin
EMU := $(BIN_DIR)/$(APP).exe
IMGSCAN := $(BIN_DIR)/imgscan.exe

.PHONY: all build imgscan test run run-mmio clean

all: build imgscan

$(BIN_DIR):
	powershell -NoProfile -Command "New-Item -ItemType Directory -Force '$(BIN_DIR)' | Out-Null"

build: $(BIN_DIR)
	go build -o $(EMU) ./cmd/main.go

imgscan: $(BIN_DIR)
	go build -o $(IMGSCAN) ./cmd/imgscan

test:
	go test ./...

run: build
	$(EMU) -rom $(ROM) -ram $(RAM) -cycles $(CYCLES) -live-uart=false -uart-limit $(UART_LIMIT)

run-mmio: build
	$(EMU) -rom $(ROM) -ram $(RAM) -cycles $(CYCLES) -trace-mmio -live-uart=false -uart-limit $(UART_LIMIT)

clean:
	powershell -NoProfile -Command "Remove-Item -Recurse -Force '$(BIN_DIR)' -ErrorAction SilentlyContinue"
