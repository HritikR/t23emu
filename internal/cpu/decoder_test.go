package cpu

import "testing"

func TestDecodeRTypeADD(t *testing.T) {

	// add $t0,$t1,$t2
	//
	// Format:
	//
	// opcode = 000000
	// rs     = 01001 ($t1)
	// rt     = 01010 ($t2)
	// rd     = 01000 ($t0)
	// shamt  = 00000
	// funct  = 100000 (ADD)

	raw := uint32(0x012A4020)

	inst := Decode(raw)

	if inst.Raw != raw {
		t.Fatalf(
			"raw mismatch",
		)
	}

	if inst.Opcode != 0 {
		t.Fatalf(
			"expected opcode 0 got %d",
			inst.Opcode,
		)
	}

	if inst.Rs != 9 {
		t.Fatalf(
			"expected rs=9 got %d",
			inst.Rs,
		)
	}

	if inst.Rt != 10 {
		t.Fatalf(
			"expected rt=10 got %d",
			inst.Rt,
		)
	}

	if inst.Rd != 8 {
		t.Fatalf(
			"expected rd=8 got %d",
			inst.Rd,
		)
	}

	if inst.Shamt != 0 {
		t.Fatalf(
			"expected shamt=0 got %d",
			inst.Shamt,
		)
	}

	if inst.Funct != 32 {
		t.Fatalf(
			"expected funct=32 got %d",
			inst.Funct,
		)
	}
}

func TestDecodeADDI(t *testing.T) {

	// addi $t0,$zero,100
	//
	// opcode = 001000 (8)
	// rs     = 00000
	// rt     = 01000
	// imm    = 100

	raw := uint32(0x20080064)

	inst := Decode(raw)

	if inst.Opcode != 8 {

		t.Fatalf(
			"expected opcode=8 got %d",
			inst.Opcode,
		)
	}

	if inst.Rs != 0 {

		t.Fatalf(
			"expected rs=0 got %d",
			inst.Rs,
		)
	}

	if inst.Rt != 8 {

		t.Fatalf(
			"expected rt=8 got %d",
			inst.Rt,
		)
	}

	if inst.Immediate != 100 {

		t.Fatalf(
			"expected immediate=100 got %d",
			inst.Immediate,
		)
	}
}

func TestDecodeFields(t *testing.T) {

	// Construct a fake instruction:
	//
	// opcode = 0x15
	// rs     = 3
	// rt     = 7
	// rd     = 12
	// shamt  = 4
	// funct  = 0x22

	raw :=
		(uint32(0x15) << 26) |
			(uint32(3) << 21) |
			(uint32(7) << 16) |
			(uint32(12) << 11) |
			(uint32(4) << 6) |
			uint32(0x22)

	inst := Decode(raw)

	tests := []struct {
		name string
		got  uint8
		want uint8
	}{
		{"opcode", inst.Opcode, 0x15},
		{"rs", inst.Rs, 3},
		{"rt", inst.Rt, 7},
		{"rd", inst.Rd, 12},
		{"shamt", inst.Shamt, 4},
		{"funct", inst.Funct, 0x22},
	}

	for _, tt := range tests {

		if tt.got != tt.want {

			t.Fatalf(
				"%s: expected %d got %d",
				tt.name,
				tt.want,
				tt.got,
			)
		}
	}
}

func TestDecodeImmediateExtraction(t *testing.T) {

	raw := uint32(0xFFFF1234)

	inst := Decode(raw)

	if inst.Immediate != 0x1234 {

		t.Fatalf(
			"expected immediate 0x1234 got 0x%04X",
			inst.Immediate,
		)
	}
}
