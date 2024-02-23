package wasm

import (
	"github.com/hybridgroup/wasman/types"
)

// Host-defined function that accepts and returns raw values.
//
// It is up to the function implementation to interpret bits from the raw values
// as the expected Go types.
type RawHostFunc = func([]uint64) []uint64

// HostFunc is an implement of wasm.Fn,
// which represents all the functions defined under host(golang) environment
type HostFunc struct {
	Signature *types.FuncType // the shape of func (defined by inputs and outputs)

	// Generator is a func defined by other dev which acts as a Generator to the function
	// (generate when NewInstance's func initializing
	Generator func(ins *Instance) RawHostFunc

	// function is the generated func from Generator, should be set at the time of wasm instance creation
	function RawHostFunc
}

func (f *HostFunc) getType() *types.FuncType {
	return f.Signature
}

func (f *HostFunc) call(ins *Instance) error {
	args := make([]uint64, len(f.Signature.InputTypes))
	for i := len(args) - 1; i >= 0; i-- {
		args[i] = ins.OperandStack.Pop()
	}
	results := f.function(args)
	for _, val := range results {
		ins.OperandStack.Push(val)
	}
	return nil
}
