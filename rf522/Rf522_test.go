package rf522

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestBitCalc(t *testing.T) {

	ba := BlocksAccess{
		B3: KeyA_RN_WB_BITS_RAB_WN_KeyB_RN_WB, // 100
		B2: RAB_WN_IN_DAB,                     // 001
		B1: RB_WN_IN_DN,                       // 101
		B0: RAB_WB_IB_DAB,                     // 110
	}

	access := CalculateBlockAccess(&ba)

	reader := func(s string) (res byte) {
		d, err := strconv.ParseUint(s, 2, 8)
		assert.NoError(t, err)
		res = byte(d & 0xFF)
		return
	}

	assert.Equal(t, reader("0110"), ba.getBits(1))

	expected := []byte{
		reader("11101001"),
		reader("01100100"),
		reader("10110001"),
		0,
	}

	expected[3] = expected[0] ^ expected[1] ^ expected[2]

	assert.Equal(t, expected, access, "Access is incorrect")

	parsedAccess := ParseBlockAccess(access)

	assert.Equal(t, ba, *parsedAccess, "Parsed access mismatch")

	fmt.Println("Block access", ba, printBytes(access))

	fmt.Println("Access for FF0780 is", *ParseBlockAccess([]byte{0x80, 0x07, 0xff}))

}
