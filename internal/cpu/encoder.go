package cpu

// EncodeR creates an R-type MIPS instruction.
//
// Format:
//
// opcode | rs | rt | rd | shamt | funct
//
//	6      5    5    5     5       6
func EncodeR(
	opcode uint8,
	rs uint8,
	rt uint8,
	rd uint8,
	shamt uint8,
	funct uint8,
) uint32 {
	return (uint32(opcode) << 26) |
		(uint32(rs) << 21) |
		(uint32(rt) << 16) |
		(uint32(rd) << 11) |
		(uint32(shamt) << 6) |
		uint32(funct)
}

// EncodeI creates an I-type MIPS instruction.
//
// Format:
//
// opcode | rs | rt | immediate
//
//	6      5    5       16
func EncodeI(
	opcode uint8,
	rs uint8,
	rt uint8,
	immediate uint16,
) uint32 {
	return (uint32(opcode) << 26) |
		(uint32(rs) << 21) |
		(uint32(rt) << 16) |
		uint32(immediate)
}
