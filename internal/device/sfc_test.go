package device

import "testing"

func TestSFCStatusReportsReceiveReadyDuringTransfer(t *testing.T) {
	sfc := NewSFC([]byte{1, 2, 3, 4}, 4)

	sfc.Write32(SFC_TRAN_CONF, 0x03)
	sfc.Write32(SFC_DEV_ADDR, 0)
	sfc.Write32(SFC_TRAN_LEN, 4)
	sfc.Write32(SFC_TRIG, SFC_TRIG_START)

	if got := sfc.Read32(SFC_SR); got&SFC_SR_RECE_REQ == 0 {
		t.Fatalf("expected SFC receive-ready bit set, got 0x%08X", got)
	}
}

func TestSFCDataReadsFromFlashAndCompletes(t *testing.T) {
	sfc := NewSFC([]byte{0x11, 0x22, 0x33, 0x44, 0x55}, 5)

	sfc.Write32(SFC_TRAN_CONF, 0x03)
	sfc.Write32(SFC_DEV_ADDR, 1)
	sfc.Write32(SFC_TRAN_LEN, 4)
	sfc.Write32(SFC_TRIG, SFC_TRIG_START)

	if got := sfc.Read32(SFC_DR); got != 0x55443322 {
		t.Fatalf("expected flash word 0x55443322, got 0x%08X", got)
	}
	if got := sfc.Read32(SFC_SR); got&SFC_SR_RECE_REQ != 0 || got&SFC_SR_END == 0 {
		t.Fatalf("expected completed transfer status, got 0x%08X", got)
	}
}

func TestSFCStartReportsTransferEnd(t *testing.T) {
	sfc := NewSFC([]byte{1, 2, 3, 4}, 4)

	sfc.Write32(SFC_TRAN_CONF, 0x03)
	sfc.Write32(SFC_DEV_ADDR, 0)
	sfc.Write32(SFC_TRAN_LEN, 4)
	sfc.Write32(SFC_TRIG, SFC_TRIG_START)

	if got := sfc.Read32(SFC_SR); got&SFC_SR_END == 0 {
		t.Fatalf("expected transfer-end bit set after start, got 0x%08X", got)
	}
}

func TestSFCTransferDoneAssertsAndClearsInterrupt(t *testing.T) {
	sfc := NewSFC([]byte{1, 2, 3, 4}, 4)
	asserted := false
	sfc.Interrupt = func(assert bool) {
		asserted = assert
	}

	sfc.Write32(SFC_INTC, ^SFC_IRQ_END)
	sfc.Write32(SFC_TRAN_CONF, 0x05)
	sfc.Write32(SFC_TRAN_LEN, 1)
	sfc.Write32(SFC_TRIG, SFC_TRIG_START)

	if !asserted {
		t.Fatalf("expected SFC transfer completion to assert interrupt")
	}

	sfc.Write32(SFC_SCR, SFC_SCR_CLR_END)

	if asserted {
		t.Fatalf("expected SCR end clear to deassert interrupt")
	}
}

func TestSFCInterruptMaskControlsPendingStatus(t *testing.T) {
	sfc := NewSFC([]byte{1, 2, 3, 4}, 4)
	asserted := false
	sfc.Interrupt = func(assert bool) {
		asserted = assert
	}

	sfc.Write32(SFC_TRAN_CONF, 0x9f)
	sfc.Write32(SFC_TRAN_LEN, 3)
	sfc.Write32(SFC_TRIG, SFC_TRIG_START)

	if asserted {
		t.Fatalf("expected reset interrupt mask to suppress pending SFC interrupt")
	}

	sfc.Write32(SFC_INTC, ^SFC_IRQ_END)
	if !asserted {
		t.Fatalf("expected unmasking pending END to assert interrupt")
	}

	sfc.Write32(SFC_INTC, SFC_IRQ_ALL)
	if asserted {
		t.Fatalf("expected masking all SFC causes to deassert interrupt")
	}
}

func TestSFCCombinedStartStopFlushKeepsNewTransferActive(t *testing.T) {
	sfc := NewSFC([]byte{1, 2, 3, 4}, 4)

	sfc.Write32(SFC_TRAN_CONF, 0x03)
	sfc.Write32(SFC_DEV_ADDR, 0)
	sfc.Write32(SFC_TRAN_LEN, 4)
	sfc.Write32(SFC_TRIG, SFC_TRIG_START|SFC_TRIG_STOP|SFC_TRIG_FLUSH)
	sfc.Write32(SFC_SCR, SFC_SCR_CLR_RREQ|SFC_SCR_CLR_END)

	if got := sfc.Read32(SFC_SR); got&SFC_SR_RECE_REQ == 0 {
		t.Fatalf("expected combined trigger to leave receive data ready, got 0x%08X", got)
	}
	if got := sfc.Read32(SFC_DR); got != 0x04030201 {
		t.Fatalf("expected flash word 0x04030201, got 0x%08X", got)
	}
}

func TestSFCSCRClearsEndStatus(t *testing.T) {
	sfc := NewSFC([]byte{1, 2, 3, 4}, 4)

	sfc.Write32(SFC_TRAN_CONF, 0x03)
	sfc.Write32(SFC_DEV_ADDR, 0)
	sfc.Write32(SFC_TRAN_LEN, 4)
	sfc.Write32(SFC_TRIG, SFC_TRIG_START)
	_ = sfc.Read32(SFC_DR)

	sfc.Write32(SFC_SCR, SFC_SCR_CLR_END)

	if got := sfc.Read32(SFC_SR); got&SFC_SR_END != 0 {
		t.Fatalf("expected SCR to clear end status, got 0x%08X", got)
	}
}

func TestSFCReadJEDECID(t *testing.T) {
	sfc := NewSFC([]byte{0x06, 0, 0, 0}, 4)

	sfc.Write32(SFC_TRAN_CONF, 0x9f)
	sfc.Write32(SFC_TRAN_LEN, 4)
	sfc.Write32(SFC_TRIG, SFC_TRIG_START)

	if got := sfc.Read32(SFC_DR); got != 0x00176085 {
		t.Fatalf("expected JEDEC ID word 0x00176085, got 0x%08X", got)
	}
}

func TestSFCDMAReadJEDECID(t *testing.T) {
	sfc := NewSFC([]byte{0x06, 0, 0, 0}, 4)
	var addr uint32
	var data []byte
	sfc.DMAWrite = func(a uint32, d []byte) {
		addr = a
		data = append([]byte(nil), d...)
	}

	sfc.Write32(SFC_MEM_ADDR, 0x100)
	sfc.Write32(SFC_TRAN_CONF, 0x9f)
	sfc.Write32(SFC_TRAN_LEN, 3)
	sfc.Write32(SFC_TRIG, SFC_TRIG_START)

	if addr != 0x100 {
		t.Fatalf("expected DMA write to 0x100, got 0x%08X", addr)
	}
	if len(data) != 3 || data[0] != 0x85 || data[1] != 0x60 || data[2] != 0x17 {
		t.Fatalf("expected JEDEC ID bytes 85 60 17, got % x", data)
	}
	if got := sfc.Read32(SFC_SR); got&SFC_SR_RECE_REQ != 0 || got&SFC_SR_END == 0 {
		t.Fatalf("expected DMA read to complete with END only, got 0x%08X", got)
	}
}

func TestSFCStatusCommandsReturnReady(t *testing.T) {
	sfc := NewSFC([]byte{0xff, 0xff, 0xff, 0xff}, 4)

	sfc.Write32(SFC_TRAN_CONF, 0x05)
	sfc.Write32(SFC_TRAN_LEN, 4)
	sfc.Write32(SFC_TRIG, SFC_TRIG_START)

	if got := sfc.Read32(SFC_DR); got != 0 {
		t.Fatalf("expected ready status word 0, got 0x%08X", got)
	}
}

func TestSFCWriteEnableAndPageProgram(t *testing.T) {
	sfc := NewSFC([]byte{0xff, 0xff, 0xff, 0xff}, 8*1024*1024)

	// 1. Send Write Enable (0x06)
	sfc.Write32(SFC_TRAN_CONF, 0x06)
	sfc.Write32(SFC_TRAN_LEN, 0)
	sfc.Write32(SFC_TRIG, SFC_TRIG_START)

	// Verify Status Register 1 (0x05) has WEL bit (0x02) set
	sfc.Write32(SFC_TRAN_CONF, 0x05)
	sfc.Write32(SFC_TRAN_LEN, 1)
	sfc.Write32(SFC_TRIG, SFC_TRIG_START)
	if got := sfc.Read32(SFC_DR) & 0xFF; got != 0x02 {
		t.Fatalf("expected WEL status bit 0x02 set, got 0x%02X", got)
	}

	// 2. Page Program (0x02) at address 0x100
	sfc.Write32(SFC_TRAN_CONF, 0x02)
	sfc.Write32(SFC_DEV_ADDR, 0x100)
	sfc.Write32(SFC_TRAN_LEN, 4)
	sfc.Write32(SFC_TRIG, SFC_TRIG_START)

	// Write data word to SFC_DR
	sfc.Write32(SFC_DR, 0xAABBCCDD)

	// 3. Read Data (0x03) back from 0x100
	sfc.Write32(SFC_TRAN_CONF, 0x03)
	sfc.Write32(SFC_DEV_ADDR, 0x100)
	sfc.Write32(SFC_TRAN_LEN, 4)
	sfc.Write32(SFC_TRIG, SFC_TRIG_START)

	if got := sfc.Read32(SFC_DR); got != 0xAABBCCDD {
		t.Fatalf("expected flash data 0xAABBCCDD, got 0x%08X", got)
	}
}

func TestSFCSectorErase(t *testing.T) {
	sfc := NewSFC([]byte{}, 8*1024*1024)

	// 1. Program 0x12345678 at 0x1000
	sfc.Write32(SFC_TRAN_CONF, 0x06)
	sfc.Write32(SFC_TRAN_LEN, 0)
	sfc.Write32(SFC_TRIG, SFC_TRIG_START)

	sfc.Write32(SFC_TRAN_CONF, 0x02)
	sfc.Write32(SFC_DEV_ADDR, 0x1000)
	sfc.Write32(SFC_TRAN_LEN, 4)
	sfc.Write32(SFC_TRIG, SFC_TRIG_START)
	sfc.Write32(SFC_DR, 0x12345678)

	// 2. Erase sector (0x20) at 0x1000
	sfc.Write32(SFC_TRAN_CONF, 0x06)
	sfc.Write32(SFC_TRAN_LEN, 0)
	sfc.Write32(SFC_TRIG, SFC_TRIG_START)

	sfc.Write32(SFC_TRAN_CONF, 0x20)
	sfc.Write32(SFC_DEV_ADDR, 0x1000)
	sfc.Write32(SFC_TRAN_LEN, 0)
	sfc.Write32(SFC_TRIG, SFC_TRIG_START)

	// 3. Read back from 0x1000 -> should be 0xFFFFFFFF (erased)
	sfc.Write32(SFC_TRAN_CONF, 0x03)
	sfc.Write32(SFC_DEV_ADDR, 0x1000)
	sfc.Write32(SFC_TRAN_LEN, 4)
	sfc.Write32(SFC_TRIG, SFC_TRIG_START)

	if got := sfc.Read32(SFC_DR); got != 0xFFFFFFFF {
		t.Fatalf("expected erased flash word 0xFFFFFFFF, got 0x%08X", got)
	}
}
