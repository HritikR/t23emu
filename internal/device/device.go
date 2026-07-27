package device

type Device interface {
	Read8(addr uint32) byte

	Write8(addr uint32, value byte)

	Read32(addr uint32) uint32

	Write32(addr uint32, value uint32)
}
