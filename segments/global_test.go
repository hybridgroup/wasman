package segments_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/hybridgroup/wasman/expr"
	"github.com/hybridgroup/wasman/segments"
	"github.com/hybridgroup/wasman/types"
)

func TestReadGlobalSegment(t *testing.T) {
	exp := &segments.GlobalSegment{
		Type: &types.GlobalType{ValType: types.ValueTypeI64, Mutable: false},
		Init: &expr.Expression{
			OpCode: expr.OpCodeI64Const,
			Data:   []byte{0x01},
		},
	}

	buf := []byte{0x7e, 0x00, 0x42, 0x01, 0x0b}
	actual, err := segments.ReadGlobalSegment(bytes.NewReader(buf))
	if err != nil {
		t.Fail()
	}
	if !reflect.DeepEqual(exp, actual) {
		t.Fail()
	}
}
