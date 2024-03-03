package expr

import (
	"fmt"

	"github.com/hybridgroup/wasman/leb128decode"
	"github.com/hybridgroup/wasman/types"
	"github.com/hybridgroup/wasman/utils"
)

// Expression is sequences of instructions terminated by an end marker.
type Expression struct {
	OpCode OpCode
	Data   []byte
}

// ReadExpression will read an expr.Expression from the io.Reader
func ReadExpression(r utils.Reader) (*Expression, error) {
	var b [1]byte
	_, err := r.Read(b[:])
	if err != nil {
		return nil, fmt.Errorf("read opcode: %v", err)
	}

	n := uint64(0)
	op := OpCode(b[0])

	switch op {
	case OpCodeI32Const:
		_, n, err = leb128decode.DecodeInt32(r)
	case OpCodeI64Const:
		_, n, err = leb128decode.DecodeInt64(r)
	case OpCodeF32Const:
		_, err = utils.ReadFloat32(r)
		n = 4
	case OpCodeF64Const:
		_, err = utils.ReadFloat64(r)
		n = 8
	case OpCodeGlobalGet:
		_, n, err = leb128decode.DecodeUint32(r)
	default:
		return nil, fmt.Errorf("%v for opcodes.OpCode: %#x", types.ErrInvalidTypeByte, b[0])
	}

	if err != nil {
		return nil, fmt.Errorf("read value: %v", err)
	}

	if _, err = r.Read(b[:]); err != nil {
		return nil, fmt.Errorf("look for end opcode: %v", err)
	}

	if b[0] != byte(OpCodeEnd) {
		return nil, fmt.Errorf("constant expression has not terminated")
	}

	// skip back
	if _, err := r.Seek(-1*int64(n+1), 1); err != nil {
		return nil, fmt.Errorf("error seeking back to read Expression Data")
	}

	data := make([]byte, n)
	if _, err := r.Read(data); err != nil {
		return nil, fmt.Errorf("error re-buffering Expression Data")
	}

	// skip past end opcode
	if _, err := r.Read(b[:]); err != nil {
		return nil, fmt.Errorf("error skipping past OpCodeEnd")
	}

	return &Expression{
		OpCode: op,
		Data:   data,
	}, nil
}
