package bus

import "github.com/HritikR/t23emu/internal/device"

type Mapping struct {
	Start uint32

	End uint32

	Device device.Device
}
