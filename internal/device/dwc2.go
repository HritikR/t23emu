package device

const (
	DWC2_GRSTCTL uint32 = 0x10
	DWC2_GRXFSIZ uint32 = 0x24
	DWC2_GSNPSID uint32 = 0x40
	DWC2_GHWCFG1 uint32 = 0x44
	DWC2_GHWCFG2 uint32 = 0x48
	DWC2_GHWCFG3 uint32 = 0x4c
	DWC2_GHWCFG4 uint32 = 0x50
)

const (
	DWC2_GRSTCTL_CSFTRST uint32 = 1 << 0
	DWC2_GRSTCTL_AHBIDLE uint32 = 1 << 31
	DWC2_GSNPSID_300A    uint32 = 0x4f54300a
	DWC2_GRXFSIZ_VALUE   uint32 = 1024
	DWC2_GHWCFG1_VALUE   uint32 = 0
	DWC2_GHWCFG3_VALUE   uint32 = (4096 << 16) | (4 << 4) | 4
	DWC2_GHWCFG4_VALUE   uint32 = (8 << 26) | (8 << 16) | (2 << 14) | 4
)

const DWC2_GHWCFG2_VALUE uint32 = (8 << 26) | (3 << 24) | (2 << 22) |
	(1 << 19) | (1 << 18) | (7 << 14) | (8 << 10) |
	(1 << 8) | (1 << 6) | (2 << 3)

func NewDWC2() *RegisterBlock {
	dwc2 := NewRegisterBlock("DWC2", 0x10000)

	dwc2.SetName(DWC2_GRSTCTL, "GRSTCTL")
	dwc2.SetReadFunc(DWC2_GRSTCTL, func() uint32 {
		return (dwc2.regs[DWC2_GRSTCTL] &^ DWC2_GRSTCTL_CSFTRST) | DWC2_GRSTCTL_AHBIDLE
	})
	dwc2.SetName(DWC2_GRXFSIZ, "GRXFSIZ")
	dwc2.SetReadFunc(DWC2_GRXFSIZ, func() uint32 {
		return DWC2_GRXFSIZ_VALUE
	})
	dwc2.SetName(DWC2_GSNPSID, "GSNPSID")
	dwc2.SetReadFunc(DWC2_GSNPSID, func() uint32 {
		return DWC2_GSNPSID_300A
	})
	dwc2.SetName(DWC2_GHWCFG1, "GHWCFG1")
	dwc2.SetReadFunc(DWC2_GHWCFG1, func() uint32 {
		return DWC2_GHWCFG1_VALUE
	})
	dwc2.SetName(DWC2_GHWCFG2, "GHWCFG2")
	dwc2.SetReadFunc(DWC2_GHWCFG2, func() uint32 {
		return DWC2_GHWCFG2_VALUE
	})
	dwc2.SetName(DWC2_GHWCFG3, "GHWCFG3")
	dwc2.SetReadFunc(DWC2_GHWCFG3, func() uint32 {
		return DWC2_GHWCFG3_VALUE
	})
	dwc2.SetName(DWC2_GHWCFG4, "GHWCFG4")
	dwc2.SetReadFunc(DWC2_GHWCFG4, func() uint32 {
		return DWC2_GHWCFG4_VALUE
	})

	return dwc2
}
