package wasm

import (
	"encoding/binary"
	"errors"

	"github.com/hybridgroup/wasman/config"
)

// ErrPtrOutOfBounds will be throw when the pointer visiting a pos out of the range of memory
var ErrPtrOutOfBounds = errors.New("pointer is out of bounds")

func memoryBase(ins *Instance) (uint64, error) {
	ins.Active.PC++
	_, err := ins.fetchUint32() // ignore align
	if err != nil {
		return 0, err
	}
	ins.Active.PC++
	v, err := ins.fetchUint32()
	if err != nil {
		return 0, err
	}

	base := uint64(v) + ins.OperandStack.Pop()
	if !(base < uint64(len(ins.Memory.Value))) {
		println("memory too small", base, len(ins.Memory.Value))
		return 0, ErrPtrOutOfBounds
	}

	return base, nil
}

func i32Load(ins *Instance) error {
	base, err := memoryBase(ins)
	if err != nil {
		return err
	}

	ins.OperandStack.Push(uint64(binary.LittleEndian.Uint32(ins.Memory.Value[base:])))

	return nil
}

func i64Load(ins *Instance) error {
	base, err := memoryBase(ins)
	if err != nil {
		return err
	}

	ins.OperandStack.Push(binary.LittleEndian.Uint64(ins.Memory.Value[base:]))

	return nil
}

func f32Load(ins *Instance) error {
	return i32Load(ins)
}

func f64Load(ins *Instance) error {
	return i64Load(ins)
}

func i32Load8s(ins *Instance) error {
	base, err := memoryBase(ins)
	if err != nil {
		return err
	}

	ins.OperandStack.Push(uint64(ins.Memory.Value[base]))

	return nil
}

func i32Load8u(ins *Instance) error {
	return i32Load8s(ins)
}

func i32Load16s(ins *Instance) error {
	base, err := memoryBase(ins)
	if err != nil {
		return err
	}

	ins.OperandStack.Push(uint64(binary.LittleEndian.Uint16(ins.Memory.Value[base:])))

	return nil
}

func i32Load16u(ins *Instance) error {
	return i32Load16s(ins)
}

func i64Load8s(ins *Instance) error {
	base, err := memoryBase(ins)
	if err != nil {
		return err
	}

	ins.OperandStack.Push(uint64(ins.Memory.Value[base]))

	return nil
}

func i64Load8u(ins *Instance) error {
	return i64Load8s(ins)
}

func i64Load16s(ins *Instance) error {
	base, err := memoryBase(ins)
	if err != nil {
		return err
	}

	ins.OperandStack.Push(uint64(binary.LittleEndian.Uint16(ins.Memory.Value[base:])))

	return nil
}

func i64Load16u(ins *Instance) error {
	return i64Load16s(ins)
}

func i64Load32s(ins *Instance) error {
	base, err := memoryBase(ins)
	if err != nil {
		return err
	}

	ins.OperandStack.Push(uint64(binary.LittleEndian.Uint32(ins.Memory.Value[base:])))

	return nil
}

func i64Load32u(ins *Instance) error {
	return i64Load32s(ins)
}

func i32Store(ins *Instance) error {
	val := ins.OperandStack.Pop()
	base, err := memoryBase(ins)
	if err != nil {
		return err
	}

	binary.LittleEndian.PutUint32(ins.Memory.Value[base:], uint32(val))

	return nil
}

func i64Store(ins *Instance) error {
	val := ins.OperandStack.Pop()
	base, err := memoryBase(ins)
	if err != nil {
		return err
	}

	binary.LittleEndian.PutUint64(ins.Memory.Value[base:], val)

	return nil
}

func f32Store(ins *Instance) error {
	val := ins.OperandStack.Pop()
	base, err := memoryBase(ins)
	if err != nil {
		return err
	}

	binary.LittleEndian.PutUint32(ins.Memory.Value[base:], uint32(val))

	return nil
}

func f64Store(ins *Instance) error {
	v := ins.OperandStack.Pop()
	base, err := memoryBase(ins)
	if err != nil {
		return err
	}

	binary.LittleEndian.PutUint64(ins.Memory.Value[base:], v)

	return nil
}

func i32Store8(ins *Instance) error {
	v := byte(ins.OperandStack.Pop())
	base, err := memoryBase(ins)
	if err != nil {
		return err
	}

	ins.Memory.Value[base] = v

	return nil
}

func i32Store16(ins *Instance) error {
	v := uint16(ins.OperandStack.Pop())
	base, err := memoryBase(ins)
	if err != nil {
		return err
	}

	binary.LittleEndian.PutUint16(ins.Memory.Value[base:], v)

	return nil
}

func i64Store8(ins *Instance) error {
	v := byte(ins.OperandStack.Pop())
	base, err := memoryBase(ins)
	if err != nil {
		return err
	}

	ins.Memory.Value[base] = v

	return nil
}

func i64Store16(ins *Instance) error {
	v := uint16(ins.OperandStack.Pop())
	base, err := memoryBase(ins)
	if err != nil {
		return err
	}

	binary.LittleEndian.PutUint16(ins.Memory.Value[base:], v)

	return nil
}

func i64Store32(ins *Instance) error {
	v := uint32(ins.OperandStack.Pop())
	base, err := memoryBase(ins)
	if err != nil {
		return err
	}

	binary.LittleEndian.PutUint32(ins.Memory.Value[base:], v)

	return nil
}

func memorySize(ins *Instance) error {
	ins.Active.PC++
	ins.OperandStack.Push(uint64(int32(len(ins.Memory.Value) / config.DefaultMemoryPageSize)))

	return nil
}

func memoryGrow(ins *Instance) error {
	ins.Active.PC++
	n := uint32(ins.OperandStack.Pop())

	if ins.Module.MemorySection[0].Max != nil &&
		uint64(n+uint32(len(ins.Memory.Value)/config.DefaultMemoryPageSize)) > uint64(*(ins.Module.MemorySection[0].Max)) {
		v := int32(-1)
		ins.OperandStack.Push(uint64(v))

		return nil
	}

	ins.OperandStack.Push(uint64(len(ins.Memory.Value)) / config.DefaultMemoryPageSize)
	ins.Memory.Value = append(ins.Memory.Value, make([]byte, n*config.DefaultMemoryPageSize)...)

	return nil
}

var ErrInvalidSubcode = errors.New("invalid bulk memory subcode")

func bulkMemory(ins *Instance) error {
	ins.Active.PC++
	opByte := ins.Active.Func.body[ins.Active.PC]

	switch opByte {
	case 0x08:
		// memory.init
		return memoryInit(ins)
	case 0x09:
		// data.drop
		return dataDrop(ins)
	case 0x0a:
		// memory.copy
		return memoryCopy(ins)
	case 0x0b:
		// memory.fill
		return memoryFill(ins)
	case 0x0c:
		// table.init
		return tableInit(ins)
	case 0x0d:
		// element.drop
		return elementDrop(ins)
	case 0x0e:
		// table.copy
		return tableCopy(ins)
	case 0x0f:
		// table.grow
		return tableGrow(ins)
	case 0x10:
		// table.size
		return tableSize(ins)
	case 0x11:
		// table.fill
		return tableFill(ins)
	default:
		return ErrInvalidSubcode
	}
}

func memoryInit(ins *Instance) error {
	ins.Active.PC++

	idx, err := ins.fetchUint32()
	if err != nil {
		return err
	}

	size := uint32(ins.OperandStack.Pop())
	offset := uint32(ins.OperandStack.Pop())
	dest := uint32(ins.OperandStack.Pop())

	if dest+size > uint32(len(ins.Memory.Value)) {
		return ErrPtrOutOfBounds
	}

	if offset+size > uint32(len(ins.Module.DataSection[idx].OffsetExpression.Data)) {
		return ErrPtrOutOfBounds
	}

	copy(ins.Memory.Value[dest:], ins.Module.DataSection[idx].OffsetExpression.Data[offset:offset+size])
	return nil
}

func dataDrop(ins *Instance) error {
	ins.Active.PC++

	// value returned here is the index of the data segment.
	_, err := ins.fetchUint32()
	if err != nil {
		return err
	}

	// TODO: can remove the memory in data section
	return nil
}

func memoryCopy(ins *Instance) error {
	ins.Active.PC++
	ins.Active.PC++

	size := uint32(ins.OperandStack.Pop())
	src := uint32(ins.OperandStack.Pop())
	dest := uint32(ins.OperandStack.Pop())

	if src+size > uint32(len(ins.Memory.Value)) || dest+size > uint32(len(ins.Memory.Value)) {
		return ErrPtrOutOfBounds
	}

	copy(ins.Memory.Value[dest:], ins.Memory.Value[src:src+size])
	return nil
}

func memoryFill(ins *Instance) error {
	ins.Active.PC++

	size := uint32(ins.OperandStack.Pop())
	v := uint32(ins.OperandStack.Pop())
	dest := uint32(ins.OperandStack.Pop())

	if dest+size > uint32(len(ins.Memory.Value)) {
		return ErrPtrOutOfBounds
	}

	for i := dest; i < dest+size; i++ {
		ins.Memory.Value[i] = byte(v)
	}

	return nil
}

func tableInit(ins *Instance) error {
	ins.Active.PC++

	eidx, err := ins.fetchUint32()
	if err != nil {
		return err
	}

	tidx, err := ins.fetchUint32()
	if err != nil {
		return err
	}

	size := uint32(ins.OperandStack.Pop())
	offset := uint32(ins.OperandStack.Pop())
	dest := uint32(ins.OperandStack.Pop())

	if dest+size > uint32(len(ins.IndexSpace.Tables[tidx].Value)) {
		return ErrPtrOutOfBounds
	}

	if offset+size > uint32(len(ins.Module.ElementsSection[eidx].OffsetExpr.Data)) {
		return ErrPtrOutOfBounds
	}

	// TODO: some kind of copy here
	//copy(ins.IndexSpace.Tables[tidx].Value[dest:], ins.Module.ElementsSection[eidx].OffsetExpr.Data[offset:offset+size])
	return nil
}

func elementDrop(ins *Instance) error {
	ins.Active.PC++

	// value returned here is the index of the element.
	_, err := ins.fetchUint32()
	if err != nil {
		return err
	}

	// TODO: can remove the memory in data section
	return nil
}

func tableCopy(ins *Instance) error {
	ins.Active.PC++

	// value returned here is the index of the table.
	xidx, err := ins.fetchUint32()
	if err != nil {
		return err
	}

	// value returned here is the index of the table.
	yidx, err := ins.fetchUint32()
	if err != nil {
		return err
	}

	size := uint32(ins.OperandStack.Pop())
	src := uint32(ins.OperandStack.Pop())
	dest := uint32(ins.OperandStack.Pop())

	if src+size > uint32(len(ins.IndexSpace.Tables[xidx].Value)) || dest+size > uint32(len(ins.IndexSpace.Tables[yidx].Value)) {
		return ErrPtrOutOfBounds
	}

	copy(ins.IndexSpace.Tables[yidx].Value[dest:], ins.IndexSpace.Tables[xidx].Value[src:src+size])
	return nil
}

func tableGrow(ins *Instance) error {
	ins.Active.PC++

	// value returned here is the index of the table.
	_, err := ins.fetchUint32()
	if err != nil {
		return err
	}

	// TODO: grow table
	return nil
}

func tableSize(ins *Instance) error {
	ins.Active.PC++

	// value returned here is the index of the table.
	idx, err := ins.fetchUint32()
	if err != nil {
		return err
	}

	if len(ins.IndexSpace.Tables) == 0 {
		v := int32(-1)
		ins.OperandStack.Push(uint64(v))
		return nil
	}

	ins.OperandStack.Push(uint64(int32(len(ins.IndexSpace.Tables[idx].Value) / config.DefaultMemoryPageSize)))

	return nil
}

func tableFill(ins *Instance) error {
	ins.Active.PC++

	// value returned here is the index of the table.
	idx, err := ins.fetchUint32()
	if err != nil {
		return err
	}

	size := uint32(ins.OperandStack.Pop())
	// the value to fill with
	ins.OperandStack.Pop()
	dest := uint32(ins.OperandStack.Pop())

	if dest+size > uint32(len(ins.IndexSpace.Tables[idx].Value)) {
		return ErrPtrOutOfBounds
	}

	// TODO: fill table data
	return nil
}
