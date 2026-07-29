package device

import "testing"

func TestMSCStatusReportsNoCardWhenAbsent(t *testing.T) {
	msc := NewMSC("MSC0", false, nil)

	if got := msc.Read32(MSC_STAT); got&MSC_STAT_CARD_DETECTED != 0 {
		t.Fatalf("expected no card-detect bit, got 0x%08X", got)
	}
}

func TestMSCNoCardCommandTimesOut(t *testing.T) {
	msc := NewMSC("MSC0", false, nil)

	msc.Write32(MSC_CMD, 41)
	msc.Write32(MSC_STRPCL, MSC_STRPCL_START_OP)

	if got := msc.Read32(MSC_IREG); got&MSC_IREG_TIME_OUT_RES == 0 {
		t.Fatalf("expected no-card command to time out, got IREG 0x%08X", got)
	}
	if got := msc.Read32(MSC_STAT); got&MSC_STAT_CARD_DETECTED != 0 {
		t.Fatalf("expected no card-detect bit after command, got 0x%08X", got)
	}
}

func TestMSCACMD41ReportsReadyOCR(t *testing.T) {
	msc := NewMSC("MSC0", true, nil)

	msc.Write32(MSC_CMD, 41)
	msc.Write32(MSC_ARG, 0x00FF8000)
	msc.Write32(MSC_STRPCL, MSC_STRPCL_START_OP)

	if got := msc.Read32(MSC_RES); got != 0x80FF8000 {
		t.Fatalf("expected ready OCR response, got 0x%08X", got)
	}
}
