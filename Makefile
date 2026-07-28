APP := t23emu

ROM ?= firmware_dump.bin
RAM ?= 67108864
CYCLES ?= 200000000000
UART_LIMIT ?= 4096

BIN_DIR := bin

ifeq ($(OS),Windows_NT)
	EXE := .exe
	RM := rmdir /S /Q
	MKDIR := mkdir
else
	EXE :=
	RM := rm -rf
	MKDIR := mkdir -p
endif

EMU := $(BIN_DIR)/$(APP)$(EXE)
IMGSCAN := $(BIN_DIR)/imgscan$(EXE)

.PHONY: all build imgscan test run run-mmio clean

all: build imgscan

$(BIN_DIR):
	$(MKDIR) $(BIN_DIR)

build: $(BIN_DIR)
	go build -o $(EMU) ./cmd/main.go

imgscan: $(BIN_DIR)
	go build -o $(IMGSCAN) ./cmd/imgscan

test:
	go test ./...

run: build
	$(EMU) -rom $(ROM) -ram $(RAM) -cycles $(CYCLES)

run-mmio: build
	$(EMU) -rom $(ROM) -ram $(RAM) -cycles $(CYCLES) -trace-mmio -live-uart=false -uart-limit $(UART_LIMIT)

clean:
ifeq ($(OS),Windows_NT)
	-$(RM) $(BIN_DIR)
else
	$(RM) $(BIN_DIR)
endif
