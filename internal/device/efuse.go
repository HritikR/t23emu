package device

// Ingenic EFUSE register offsets, relative to the EFUSE physical base at
// 0x13540000.
const (
	EFUSE_SOC_INFO  uint32 = 0x238 // subsoctype1
	EFUSE_SOC_TYPE2 uint32 = 0x250 // subsoctype2
)

const (
	// The T23 U-Boot image shifts SOC_INFO right by 16 and maps 0x1111 to
	// its built-in "Board info: T23N" string. A zero value falls through to
	// "Board info: No SOC Info".
	EFUSE_SOC_T23N uint32 = 0x11110000
)

// NewEFUSE creates a minimal one-time-programmable fuse controller model.
func NewEFUSE() *RegisterBlock {
	efuse := NewRegisterBlock("EFUSE", 0x10000)

	efuse.SetName(EFUSE_SOC_INFO, "SOC_INFO")
	efuse.SetInitial(EFUSE_SOC_INFO, EFUSE_SOC_T23N)

	efuse.SetName(EFUSE_SOC_TYPE2, "SOC_TYPE2")
	efuse.SetInitial(EFUSE_SOC_TYPE2, 0x00000000)

	return efuse
}
