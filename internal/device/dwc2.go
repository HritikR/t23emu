package device

const (
	DWC2_GRSTCTL uint32 = 0x10
	DWC2_GSNPSID uint32 = 0x40
)

const (
	DWC2_GRSTCTL_CSFTRST uint32 = 1 << 0
	DWC2_GRSTCTL_AHBIDLE uint32 = 1 << 31
	DWC2_GSNPSID_300A    uint32 = 0x4f54300a
)

func NewDWC2() *RegisterBlock {
	dwc2 := NewRegisterBlock("DWC2", 0x10000)

	dwc2.SetName(DWC2_GRSTCTL, "GRSTCTL")
	dwc2.SetReadFunc(DWC2_GRSTCTL, func() uint32 {
		return (dwc2.regs[DWC2_GRSTCTL] &^ DWC2_GRSTCTL_CSFTRST) | DWC2_GRSTCTL_AHBIDLE
	})
	dwc2.SetName(DWC2_GSNPSID, "GSNPSID")
	dwc2.SetReadFunc(DWC2_GSNPSID, func() uint32 {
		return DWC2_GSNPSID_300A
	})

	return dwc2
}
