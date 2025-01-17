package segments_test

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/hybridgroup/wasman/segments"
	"github.com/hybridgroup/wasman/types"
	"github.com/hybridgroup/wasman/utils"
)

func TestReadImportDesc(t *testing.T) {
	t.Run("ng", func(t *testing.T) {
		buf := []byte{0x04}
		_, err := segments.ReadImportDesc(bytes.NewReader(buf))
		if !errors.Is(err, types.ErrInvalidTypeByte) {
			t.Log(err)
			t.Fail()
		}
	})

	for i, c := range []struct {
		bytes []byte
		exp   *segments.ImportDesc
	}{
		{
			bytes: []byte{0x00, 0x0a},
			exp: &segments.ImportDesc{
				Kind:         0,
				TypeIndexPtr: utils.Uint32Ptr(10),
			},
		},
		{
			bytes: []byte{0x01, 0x70, 0x0, 0x0a},
			exp: &segments.ImportDesc{
				Kind: 1,
				TableTypePtr: &types.TableType{
					Elem:   0x70,
					Limits: &types.Limits{Min: 10},
				},
			},
		},
		{
			bytes: []byte{0x02, 0x0, 0x0a},
			exp: &segments.ImportDesc{
				Kind:       2,
				MemTypePtr: &types.MemoryType{Min: 10},
			},
		},
		{
			bytes: []byte{0x03, 0x7e, 0x01},
			exp: &segments.ImportDesc{
				Kind:          3,
				GlobalTypePtr: &types.GlobalType{ValType: types.ValueTypeI64, Mutable: true},
			},
		},
	} {
		t.Run(utils.IntToString(i), func(t *testing.T) {
			actual, err := segments.ReadImportDesc(bytes.NewReader(c.bytes))
			if err != nil {
				t.Fail()
			}
			if !reflect.DeepEqual(c.exp, actual) {
				t.Fail()
			}
		})

	}
}

func TestReadImportSegment(t *testing.T) {
	exp := &segments.ImportSegment{
		Module: "abc",
		Name:   "ABC",
		Desc:   &segments.ImportDesc{Kind: 0, TypeIndexPtr: utils.Uint32Ptr(10)},
	}

	buf := []byte{byte(len(exp.Module))}
	buf = append(buf, exp.Module...)
	buf = append(buf, byte(len(exp.Name)))
	buf = append(buf, exp.Name...)
	buf = append(buf, 0x00, 0x0a)

	actual, err := segments.ReadImportSegment(bytes.NewReader(buf))
	if err != nil {
		t.Fail()
	}
	if !reflect.DeepEqual(exp, actual) {
		t.Fail()
	}
}
