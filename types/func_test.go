package types_test

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/hybridgroup/wasman/types"
	"github.com/hybridgroup/wasman/utils"
)

func TestReadFunctionType(t *testing.T) {
	t.Run("ng", func(t *testing.T) {
		buf := []byte{0x00}
		_, err := types.ReadFuncType(bytes.NewReader(buf))
		if !errors.Is(err, types.ErrInvalidTypeByte) {
			t.Fail()
			t.Log(err)
		}
	})

	for i, c := range []struct {
		bytes []byte
		exp   *types.FuncType
	}{
		{
			bytes: []byte{0x60, 0x0, 0x0},
			exp: &types.FuncType{
				InputTypes:  []types.ValueType{},
				ReturnTypes: []types.ValueType{},
			},
		},
		{
			bytes: []byte{0x60, 0x2, 0x7f, 0x7e, 0x0},
			exp: &types.FuncType{
				InputTypes:  []types.ValueType{types.ValueTypeI32, types.ValueTypeI64},
				ReturnTypes: []types.ValueType{},
			},
		},
		{
			bytes: []byte{0x60, 0x1, 0x7e, 0x2, 0x7f, 0x7e},
			exp: &types.FuncType{
				InputTypes:  []types.ValueType{types.ValueTypeI64},
				ReturnTypes: []types.ValueType{types.ValueTypeI32, types.ValueTypeI64},
			},
		},
		{
			bytes: []byte{0x60, 0x0, 0x2, 0x7f, 0x7e},
			exp: &types.FuncType{
				InputTypes:  []types.ValueType{},
				ReturnTypes: []types.ValueType{types.ValueTypeI32, types.ValueTypeI64},
			},
		},
	} {
		t.Run(utils.IntToString(i), func(t *testing.T) {
			actual, err := types.ReadFuncType(bytes.NewReader(c.bytes))
			if err != nil {
				t.Fail()
			}
			if !reflect.DeepEqual(c.exp, actual) {
				t.Fail()
			}
		})
	}
}
