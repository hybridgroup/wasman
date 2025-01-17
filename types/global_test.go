package types_test

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/hybridgroup/wasman/types"
	"github.com/hybridgroup/wasman/utils"
)

func TestReadGlobalType(t *testing.T) {
	t.Run("ng", func(t *testing.T) {
		buf := []byte{0x7e, 0x3}
		_, err := types.ReadGlobalType(bytes.NewReader(buf))
		if !errors.Is(err, types.ErrInvalidTypeByte) {
			t.Log(err)
			t.Fail()
		}
	})

	for i, c := range []struct {
		bytes []byte
		exp   *types.GlobalType
	}{
		{bytes: []byte{0x7e, 0x00}, exp: &types.GlobalType{ValType: types.ValueTypeI64, Mutable: false}},
		{bytes: []byte{0x7e, 0x01}, exp: &types.GlobalType{ValType: types.ValueTypeI64, Mutable: true}},
	} {
		t.Run(utils.IntToString(i), func(t *testing.T) {
			actual, err := types.ReadGlobalType(bytes.NewReader(c.bytes))
			if err != nil {
				t.Fail()
			}
			if !reflect.DeepEqual(c.exp, actual) {
				t.Fail()
			}
		})
	}
}
