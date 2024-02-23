package wasm

import (
	"testing"

	"github.com/hybridgroup/wasman/expr"
	"github.com/hybridgroup/wasman/stacks"
	"github.com/hybridgroup/wasman/types"
)

func TestHostFunction_Call(t *testing.T) {
	var cnt uint64
	f := func(in []uint64) []uint64 {
		cnt += in[0]
		return []uint64{1, 2, 3, 4}
	}
	hf := &HostFunc{
		function: f,
		Signature: &types.FuncType{
			InputTypes:  []types.ValueType{types.ValueTypeI64},
			ReturnTypes: []types.ValueType{types.ValueTypeI32, types.ValueTypeI64, types.ValueTypeF32, types.ValueTypeF64},
		},
	}

	vm := &Instance{OperandStack: stacks.NewOperandStack()}
	vm.OperandStack.Push(10)
	err := hf.call(vm)
	if err != nil {
		t.Logf("call error: %v", err)
		t.Fail()
	}
	if vm.OperandStack.Ptr != 3 {
		t.Logf("pointer is %v", vm.OperandStack.Ptr)
		t.Fail()
	}
	if cnt != 10 {
		t.Logf("cnt %v", vm.OperandStack.Ptr)
		t.Fail()
	}

	// f64
	if vm.OperandStack.Pop() != 4 {
		t.Logf("cnt %v", vm.OperandStack.Ptr)
		t.Fail()
	}
	if vm.OperandStack.Pop() != 3 {
		t.Fail()
	}
	if vm.OperandStack.Pop() != 2 {
		t.Fail()
	}
	if vm.OperandStack.Pop() != 1 {
		t.Fail()
	}
}

func TestNativeFunction_Call(t *testing.T) {
	n := &wasmFunc{
		signature: &types.FuncType{},
		body: []byte{
			byte(expr.OpCodeI64Const), 0x05, byte(expr.OpCodeReturn),
		},
	}
	vm := &Instance{
		Module:       new(Module),
		OperandStack: stacks.NewOperandStack(),
		Active: &Frame{
			PC: 1000,
		},
		FrameStack: &stacks.Stack[*Frame]{
			Ptr:    -1,
			Values: make([]*Frame, stacks.InitialLabelStackHeight),
		},
	}
	if n.call(vm) != nil {
		t.Fail()
	}
	if vm.OperandStack.Pop() != 0x05 {
		t.Fail()
	}
	if vm.Active.PC != 1000 {
		t.Fail()
	}
}

func TestVirtualMachine_execNativeFunction(t *testing.T) {
	n := &wasmFunc{
		signature: &types.FuncType{},
		body: []byte{
			byte(expr.OpCodeI64Const), 0x05,
			byte(expr.OpCodeI64Const), 0x01,
			byte(expr.OpCodeReturn),
		},
	}
	vm := &Instance{
		Module:       new(Module),
		OperandStack: stacks.NewOperandStack(),
		Active: &Frame{
			Func: n,
		},
	}

	if vm.execFunc() != nil {
		t.Fail()
	}
	if vm.Active.PC != 4 {
		t.Fail()
	}
	if vm.OperandStack.Pop() != 0x01 {
		t.Fail()
	}
	if vm.OperandStack.Pop() != 0x05 {
		t.Fail()
	}
}
