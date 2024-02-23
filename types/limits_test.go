package types_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/hybridgroup/wasman/utils"

	"github.com/hybridgroup/wasman/types"
)

func TestReadLimitsType(t *testing.T) {
	for i, c := range []struct {
		bytes []byte
		exp   *types.Limits
	}{
		{bytes: []byte{0x00, 0xa}, exp: &types.Limits{Min: 10}},
		{bytes: []byte{0x01, 0xa, 0xa}, exp: &types.Limits{Min: 10, Max: utils.Uint32Ptr(10)}},
	} {
		t.Run(utils.IntToString(i), func(t *testing.T) {
			actual, err := types.ReadLimits(bytes.NewReader(c.bytes))
			if err != nil {
				t.Fail()
			}
			if !reflect.DeepEqual(c.exp, actual) {
				t.Fail()
			}
		})
	}
}
