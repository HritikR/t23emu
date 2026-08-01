package device

import (
	"bytes"
	"testing"
)

// Helper to decode a short 32-bit SD response from MSC_RES matching Ingenic driver logic:
// response = res0<<24 | res1<<8 | (res2 & 0xff)
func readShortResponse(msc *MSC) uint32 {
	res0 := uint32(msc.Read8(MSC_RES)) | (uint32(msc.Read8(MSC_RES)) << 8)
	res1 := uint32(msc.Read8(MSC_RES)) | (uint32(msc.Read8(MSC_RES)) << 8)
	res2 := uint32(msc.Read8(MSC_RES)) | (uint32(msc.Read8(MSC_RES)) << 8)
	return (res0 << 24) | (res1 << 8) | (res2 & 0xff)
}

func TestMSCNoCardCommandTimesOut(t *testing.T) {
	msc := NewMSC("MSC0", false, nil)

	// Initial status has no timeout
	if got := msc.Read32(MSC_STAT); got&MSC_STAT_TIME_OUT_RES != 0 {
		t.Fatalf("expected no timeout bit initially, got 0x%08X", got)
	}

	msc.Write32(MSC_CMD, 41)
	msc.Write32(MSC_STRPCL, MSC_STRPCL_START_OP)

	if got := msc.Read32(MSC_IREG); got&MSC_IREG_TIME_OUT_RES == 0 {
		t.Fatalf("expected no-card command to time out in IREG, got 0x%08X", got)
	}
	if got := msc.Read32(MSC_STAT); got&MSC_STAT_TIME_OUT_RES == 0 {
		t.Fatalf("expected no-card command to time out in STAT, got 0x%08X", got)
	}
}

func TestMSCControllerReset(t *testing.T) {
	msc := NewMSC("MSC0", true, nil)

	// Issue CMD55 and then reset
	msc.Write32(MSC_CMD, 55)
	msc.Write32(MSC_STRPCL, MSC_STRPCL_START_OP)

	msc.Write32(MSC_STRPCL, MSC_STRPCL_RESET)

	if got := msc.Read32(MSC_IREG); got != 0 {
		t.Fatalf("expected IREG to be cleared after reset, got 0x%08X", got)
	}

	// ACMD41 without preceding CMD55 (since reset cleared expectAppCmd)
	msc.Write32(MSC_CMD, 41)
	msc.Write32(MSC_ARG, 0x40FF8000)
	msc.Write32(MSC_STRPCL, MSC_STRPCL_START_OP)

	got := readShortResponse(msc)
	if got != 0x80FF8000 {
		t.Fatalf("expected fallback response 0x80FF8000 after reset cleared expectAppCmd, got 0x%08X", got)
	}
}

func TestMSCClockControl(t *testing.T) {
	msc := NewMSC("MSC0", true, nil)

	msc.Write32(MSC_STRPCL, MSC_STRPCL_CLOCK_CONTROL_START)
	if !msc.clockRunning {
		t.Fatalf("expected clockRunning to be true")
	}

	msc.Write32(MSC_STRPCL, MSC_STRPCL_CLOCK_CONTROL_STOP)
	if msc.clockRunning {
		t.Fatalf("expected clockRunning to be false")
	}
}

func TestMSCCMD0AndCMD1(t *testing.T) {
	msc := NewMSC("MSC0", true, nil)

	// CMD0: GO_IDLE_STATE
	msc.Write32(MSC_CMD, 0)
	msc.Write32(MSC_STRPCL, MSC_STRPCL_START_OP)
	if got := readShortResponse(msc); got != 0 {
		t.Fatalf("CMD0: expected 0, got 0x%08X", got)
	}

	// CMD1: SEND_OP_COND
	msc.Write32(MSC_CMD, 1)
	msc.Write32(MSC_STRPCL, MSC_STRPCL_START_OP)
	if got := readShortResponse(msc); got != 0x80FF8000 {
		t.Fatalf("CMD1: expected 0x80FF8000, got 0x%08X", got)
	}
}

func TestMSCACMD41AndSDHC(t *testing.T) {
	msc := NewMSC("MSC0", true, nil)

	// CMD55 (APP_CMD)
	msc.Write32(MSC_CMD, 55)
	msc.Write32(MSC_STRPCL, MSC_STRPCL_START_OP)
	res55 := readShortResponse(msc)
	if res55&(1<<5) == 0 {
		t.Fatalf("expected APP_CMD bit in CMD55 response, got 0x%08X", res55)
	}

	// ACMD41 with CCS flag (bit 30) -> High Capacity SDHC card
	msc.Write32(MSC_CMD, 41)
	msc.Write32(MSC_ARG, 0x40FF8000)
	msc.Write32(MSC_STRPCL, MSC_STRPCL_START_OP)

	got := readShortResponse(msc)
	if got != 0xC0FF8000 {
		t.Fatalf("expected 0xC0FF8000 (ready + CCS), got 0x%08X", got)
	}
	if !msc.isSDHC {
		t.Fatalf("expected isSDHC to be true")
	}
}

func TestMSCCMD2AndCMD10LongResponse(t *testing.T) {
	msc := NewMSC("MSC0", true, nil)

	// CMD2: ALL_SEND_CID
	msc.Write32(MSC_CMD, 2)
	msc.Write32(MSC_STRPCL, MSC_STRPCL_START_OP)

	// Long response returns 16 bytes (or 8 halfwords)
	respBytes := make([]byte, 16)
	for i := 0; i < 16; i++ {
		respBytes[i] = msc.Read8(MSC_RES)
	}

	// Verify long response is non-empty
	if bytes.Equal(respBytes, make([]byte, 16)) {
		t.Fatalf("expected non-zero CID long response")
	}
}

func TestMSCCMD3AndCMD7RCA(t *testing.T) {
	msc := NewMSC("MSC0", true, nil)

	// CMD3: SEND_RELATIVE_ADDR
	msc.Write32(MSC_CMD, 3)
	msc.Write32(MSC_STRPCL, MSC_STRPCL_START_OP)
	res3 := readShortResponse(msc)
	if res3>>16 != 1 {
		t.Fatalf("expected RCA 1 in CMD3 response, got RCA 0x%04X", res3>>16)
	}

	// CMD7: SELECT_CARD with RCA 1
	msc.Write32(MSC_CMD, 7)
	msc.Write32(MSC_ARG, 1<<16)
	msc.Write32(MSC_STRPCL, MSC_STRPCL_START_OP)
	res7 := readShortResponse(msc)
	if res7 != 0x00000900 {
		t.Fatalf("expected 0x00000900 for matching RCA CMD7, got 0x%08X", res7)
	}

	// CMD7 with mismatched RCA
	msc.Write32(MSC_CMD, 7)
	msc.Write32(MSC_ARG, 2<<16)
	msc.Write32(MSC_STRPCL, MSC_STRPCL_START_OP)
	if got := readShortResponse(msc); got != 0 {
		t.Fatalf("expected 0 for mismatched RCA CMD7, got 0x%08X", got)
	}
}

func TestMSCCMD8(t *testing.T) {
	msc := NewMSC("MSC0", true, nil)

	msc.Write32(MSC_CMD, 8)
	msc.Write32(MSC_ARG, 0x000001AA)
	msc.Write32(MSC_STRPCL, MSC_STRPCL_START_OP)

	if got := readShortResponse(msc); got != 0x1AA {
		t.Fatalf("expected 0x1AA check pattern echo, got 0x%08X", got)
	}
}

func TestMSCSDHCCSD(t *testing.T) {
	dummyDisk := make([]byte, 1024*1024) // 1MB disk
	msc := NewMSC("MSC0", true, dummyDisk)

	// CMD9: SEND_CSD
	msc.Write32(MSC_CMD, 9)
	msc.Write32(MSC_STRPCL, MSC_STRPCL_START_OP)

	csdBytes := make([]byte, 16)
	for i := 0; i < 16; i++ {
		csdBytes[i] = msc.Read8(MSC_RES)
	}

	// First halfword low byte contains CSD structure (0x40 for v2.0)
	if csdBytes[0] != 0x40 {
		t.Fatalf("expected CSD v2.0 header byte 0x40, got 0x%02X", csdBytes[0])
	}
}

func TestMSCACMD51SendSCR(t *testing.T) {
	msc := NewMSC("MSC0", true, nil)

	msc.Write32(MSC_CMD, 55)
	msc.Write32(MSC_STRPCL, MSC_STRPCL_START_OP)

	msc.Write32(MSC_CMD, 51)
	msc.Write32(MSC_STRPCL, MSC_STRPCL_START_OP)

	if got := msc.Read32(MSC_STAT); got&MSC_STAT_DATA_TRAN_DONE == 0 {
		t.Fatalf("expected DATA_TRAN_DONE bit set in STAT for ACMD51")
	}

	w0 := msc.Read32(MSC_RXFIFO)
	w1 := msc.Read32(MSC_RXFIFO)
	if w0 != 0x00003502 || w1 != 0 {
		t.Fatalf("unexpected SCR payload: w0=0x%08X, w1=0x%08X", w0, w1)
	}
}

func TestMSCReadBlockSDHC(t *testing.T) {
	disk := make([]byte, 2048)
	for i := range disk {
		disk[i] = byte(i & 0xFF)
	}

	msc := NewMSC("MSC0", true, disk)

	// Set SDHC mode via ACMD41
	msc.Write32(MSC_CMD, 55)
	msc.Write32(MSC_STRPCL, MSC_STRPCL_START_OP)
	msc.Write32(MSC_CMD, 41)
	msc.Write32(MSC_ARG, 1<<30)
	msc.Write32(MSC_STRPCL, MSC_STRPCL_START_OP)

	// Read block 1 (sector 1 -> byte offset 512)
	msc.Write32(MSC_NOB, 1)
	msc.Write32(MSC_CMD, 17)
	msc.Write32(MSC_ARG, 1) // sector 1
	msc.Write32(MSC_STRPCL, MSC_STRPCL_START_OP)

	if got := msc.Read32(MSC_STAT); got&MSC_STAT_DATA_TRAN_DONE == 0 {
		t.Fatalf("expected DATA_TRAN_DONE bit set for CMD17")
	}

	firstWord := msc.Read32(MSC_RXFIFO)
	expectedWord := uint32(0x00) | uint32(0x01)<<8 | uint32(0x02)<<16 | uint32(0x03)<<24
	if firstWord != expectedWord {
		t.Fatalf("expected first word 0x%08X, got 0x%08X", expectedWord, firstWord)
	}
}

func TestMSCReadBlockOutOfBoundsTimesOut(t *testing.T) {
	disk := make([]byte, 512)
	msc := NewMSC("MSC0", true, disk)

	// Attempt to read sector past disk end
	msc.Write32(MSC_NOB, 1)
	msc.Write32(MSC_CMD, 17)
	msc.Write32(MSC_ARG, 100)
	msc.Write32(MSC_STRPCL, MSC_STRPCL_START_OP)

	if got := msc.Read32(MSC_STAT); got&MSC_STAT_TIME_OUT_READ == 0 {
		t.Fatalf("expected MSC_STAT_TIME_OUT_READ for out-of-bounds read")
	}
	if got := msc.Read32(MSC_IREG); got&MSC_IREG_TIME_OUT_READ == 0 {
		t.Fatalf("expected MSC_IREG_TIME_OUT_READ for out-of-bounds read")
	}
}

func TestMSCIREGWriteToClear(t *testing.T) {
	msc := NewMSC("MSC0", true, nil)

	msc.Write32(MSC_CMD, 1)
	msc.Write32(MSC_STRPCL, MSC_STRPCL_START_OP)

	ireg := msc.Read32(MSC_IREG)
	if ireg == 0 {
		t.Fatalf("expected non-zero IREG after command")
	}

	// Write 1s to clear IREG
	msc.Write32(MSC_IREG, ireg)
	if got := msc.Read32(MSC_IREG); got != 0 {
		t.Fatalf("expected IREG to be cleared, got 0x%08X", got)
	}
}
