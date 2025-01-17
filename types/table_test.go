package types_test

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/hybridgroup/wasman/utils"

	"github.com/hybridgroup/wasman/types"
)

func TestReadTableType(t *testing.T) {
	t.Run("ng", func(t *testing.T) {
		buf := []byte{0x00}
		_, err := types.ReadTableType(bytes.NewReader(buf))
		if !errors.Is(err, types.ErrInvalidTypeByte) {
			t.Log(err)
			t.Fail()
		}
	})

	for i, c := range []struct {
		bytes []byte
		exp   *types.TableType
	}{
		{
			bytes: []byte{0x70, 0x00, 0xa},
			exp: &types.TableType{
				Elem:   0x70,
				Limits: &types.Limits{Min: 10},
			},
		},
		{
			bytes: []byte{0x70, 0x01, 0x01, 0xa},
			exp: &types.TableType{
				Elem:   0x70,
				Limits: &types.Limits{Min: 1, Max: utils.Uint32Ptr(10)},
			},
		},
	} {
		c := c
		t.Run(utils.IntToString(i), func(t *testing.T) {
			actual, err := types.ReadTableType(bytes.NewReader(c.bytes))
			if err != nil {
				t.Fail()
			}
			if !reflect.DeepEqual(c.exp, actual) {
				t.Fail()
			}
		})
	}
}
