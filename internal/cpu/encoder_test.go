package cpu

import "testing"

func TestEncodeRADD(t *testing.T) {

	raw := EncodeR(
		OP_SPECIAL,
		9,  // rs ($t1)
		10, // rt ($t2)
		8,  // rd ($t0)
		0,
		FUNCT_ADD,
	)

	expected := uint32(0x012A4020)

	if raw != expected {

		t.Fatalf(
			"expected 0x%08X got 0x%08X",
			expected,
			raw,
		)
	}

	inst := Decode(raw)

	if inst.Rs != 9 ||
		inst.Rt != 10 ||
		inst.Rd != 8 ||
		inst.Funct != FUNCT_ADD {

		t.Fatalf(
			"encoded instruction decoded incorrectly",
		)
	}
}

func TestEncodeADDI(t *testing.T) {

	raw := EncodeI(
		OP_ADDI,
		0,
		8,
		42,
	)

	expected := uint32(0x2008002A)

	if raw != expected {

		t.Fatalf(
			"expected 0x%08X got 0x%08X",
			expected,
			raw,
		)
	}

	inst := Decode(raw)

	if inst.Opcode != OP_ADDI ||
		inst.Rs != 0 ||
		inst.Rt != 8 ||
		inst.Immediate != 42 {

		t.Fatalf(
			"encoded instruction decoded incorrectly",
		)
	}
}
