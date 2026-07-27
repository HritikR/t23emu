package device

import "testing"

func TestSFCStatusReportsReceiveReadyDuringTransfer(t *testing.T) {
	sfc := NewSFC([]byte{1, 2, 3, 4})

	sfc.Write32(SFC_DEV_ADDR, 0)
	sfc.Write32(SFC_TRAN_LEN, 4)
	sfc.Write32(SFC_TRIG, SFC_TRIG_START)

	if got := sfc.Read32(SFC_SR); got&SFC_SR_RECE_REQ == 0 {
		t.Fatalf("expected SFC receive-ready bit set, got 0x%08X", got)
	}
}

func TestSFCDataReadsFromFlashAndCompletes(t *testing.T) {
	sfc := NewSFC([]byte{0x11, 0x22, 0x33, 0x44, 0x55})

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
	sfc := NewSFC([]byte{1, 2, 3, 4})

	sfc.Write32(SFC_DEV_ADDR, 0)
	sfc.Write32(SFC_TRAN_LEN, 4)
	sfc.Write32(SFC_TRIG, SFC_TRIG_START)

	if got := sfc.Read32(SFC_SR); got&SFC_SR_END == 0 {
		t.Fatalf("expected transfer-end bit set after start, got 0x%08X", got)
	}
}

func TestSFCSCRClearsEndStatus(t *testing.T) {
	sfc := NewSFC([]byte{1, 2, 3, 4})

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
	sfc := NewSFC([]byte{0x06, 0, 0, 0})

	sfc.Write32(SFC_TRAN_CONF, 0x9f)
	sfc.Write32(SFC_TRAN_LEN, 4)
	sfc.Write32(SFC_TRIG, SFC_TRIG_START)

	if got := sfc.Read32(SFC_DR); got != 0x001740ef {
		t.Fatalf("expected JEDEC ID word 0x001740ef, got 0x%08X", got)
	}
}

func TestSFCStatusCommandsReturnReady(t *testing.T) {
	sfc := NewSFC([]byte{0xff, 0xff, 0xff, 0xff})

	sfc.Write32(SFC_TRAN_CONF, 0x05)
	sfc.Write32(SFC_TRAN_LEN, 4)
	sfc.Write32(SFC_TRIG, SFC_TRIG_START)

	if got := sfc.Read32(SFC_DR); got != 0 {
		t.Fatalf("expected ready status word 0, got 0x%08X", got)
	}
}
