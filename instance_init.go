package wasman

import (
	"bytes"
	"fmt"
	"io"

	"github.com/c0mm4nd/wasman/instr"
	"github.com/c0mm4nd/wasman/leb128"
	"github.com/c0mm4nd/wasman/segments"
	"github.com/c0mm4nd/wasman/types"
)

// buildIndexSpaces build index spaces of the module with the given external modules
func (ins *Instance) buildIndexSpaces(externModules map[string]*Module) error {
	ins.indexSpace = &indexSpace{}

	// resolve imports
	if err := ins.resolveImports(externModules); err != nil {
		return fmt.Errorf("resolve imports: %w", err)
	}

	// fill in the gap between the definition and imported ones in index spaces
	// note: MVP restricts the size of memory index spaces to 1
	if diff := len(ins.TablesSection) - len(ins.indexSpace.Tables); diff > 0 {
		for i := 0; i < diff; i++ {
			ins.indexSpace.Tables = append(ins.indexSpace.Tables, []*uint32{})
		}
	}

	// fill in the gap between the definition and imported ones in index spaces
	// note: MVP restricts the size of memory index spaces to 1
	if diff := len(ins.MemorySection) - len(ins.indexSpace.Memories); diff > 0 {
		for i := 0; i < diff; i++ {
			ins.indexSpace.Memories = append(ins.indexSpace.Memories, []byte{})
		}
	}

	if err := ins.buildGlobalIndexSpace(); err != nil {
		return fmt.Errorf("build global index space: %w", err)
	}
	if err := ins.buildFunctionIndexSpace(); err != nil {
		return fmt.Errorf("build function index space: %w", err)
	}
	if err := ins.buildTableIndexSpace(); err != nil {
		return fmt.Errorf("build table index space: %w", err)
	}
	if err := ins.buildMemoryIndexSpace(); err != nil {
		return fmt.Errorf("build memory index space: %w", err)
	}

	return nil
}

func (ins *Instance) resolveImports(externModules map[string]*Module) error {
	for _, is := range ins.ImportsSection {
		em, ok := externModules[is.Module]
		if !ok {
			return fmt.Errorf("failed to resolve import of module name %s", is.Module)
		}

		es, ok := em.ExportsSection[is.Name]
		if !ok {
			return fmt.Errorf("%s not exported in module %s", is.Name, is.Module)
		}

		if is.Desc.Kind != es.Desc.Kind {
			return fmt.Errorf("type mismatch on export: got %#x but want %#x", es.Desc.Kind, is.Desc.Kind)
		}
		switch is.Desc.Kind {
		case 0x00: // function
			if err := ins.applyFunctionImport(is, em, es); err != nil {
				return fmt.Errorf("applyFunctionImport failed: %w", err)
			}
		case 0x01: // table
			if err := ins.applyTableImport(em, es); err != nil {
				return fmt.Errorf("applyTableImport failed: %w", err)
			}
		case 0x02: // memory
			if err := ins.applyMemoryImport(em, es); err != nil {
				return fmt.Errorf("applyMemoryImport: %w", err)
			}
		case 0x03: // global
			if err := ins.applyGlobalImport(em, es); err != nil {
				return fmt.Errorf("applyGlobalImport: %w", err)
			}
		default:
			return fmt.Errorf("invalid kind of import: %#x", is.Desc.Kind)
		}
	}
	return nil
}

func (ins *Instance) applyFunctionImport(importSeg *segments.ImportSegment, externModule *Module, exportSeg *segments.ExportSegment) error {
	if exportSeg.Desc.Index >= uint32(len(externModule.indexSpace.Functions)) {
		return fmt.Errorf("exported index out of range")
	}

	if importSeg.Desc.TypeIndexPtr == nil {
		return fmt.Errorf("is.Desc.TypeIndexPtr is nill")
	}

	iSig := ins.TypesSection[*importSeg.Desc.TypeIndexPtr]
	f := externModule.indexSpace.Functions[exportSeg.Desc.Index]
	if !types.HasSameSignature(iSig.ReturnTypes, f.getType().ReturnTypes) {
		return fmt.Errorf("return signature mimatch: %#v != %#v", iSig.ReturnTypes, f.getType().ReturnTypes)
	} else if !types.HasSameSignature(iSig.InputTypes, f.getType().InputTypes) {
		return fmt.Errorf("input signature mimatch: %#v != %#v", iSig.InputTypes, f.getType().InputTypes)
	}
	ins.indexSpace.Functions = append(ins.indexSpace.Functions, f)
	return nil
}

func (ins *Instance) applyTableImport(externModule *Module, exportSeg *segments.ExportSegment) error {
	if exportSeg.Desc.Index >= uint32(len(externModule.indexSpace.Tables)) {
		return fmt.Errorf("exported index out of range")
	}

	// note: MVP restricts the size of table index spaces to 1
	ins.indexSpace.Tables = append(ins.indexSpace.Tables, externModule.indexSpace.Tables[exportSeg.Desc.Index])
	return nil
}

func (ins *Instance) applyMemoryImport(externModule *Module, exportSegment *segments.ExportSegment) error {
	if exportSegment.Desc.Index >= uint32(len(externModule.indexSpace.Memories)) {
		return fmt.Errorf("exported index out of range")
	}

	// note: MVP restricts the size of memory index spaces to 1
	ins.indexSpace.Memories = append(ins.indexSpace.Memories, externModule.indexSpace.Memories[exportSegment.Desc.Index])
	return nil
}

func (ins *Instance) applyGlobalImport(externModule *Module, exportSegment *segments.ExportSegment) error {
	if exportSegment.Desc.Index >= uint32(len(externModule.indexSpace.Globals)) {
		return fmt.Errorf("exported index out of range")
	}

	gb := externModule.indexSpace.Globals[exportSegment.Desc.Index]
	if gb.Type.Mutable {
		return fmt.Errorf("cannot import mutable global")
	}

	ins.indexSpace.Globals = append(externModule.indexSpace.Globals, gb)
	return nil
}

func (ins *Instance) buildGlobalIndexSpace() error {
	for _, gs := range ins.GlobalsSection {
		v, err := ins.execExpr(gs.Init)
		if err != nil {
			return fmt.Errorf("execution failed: %w", err)
		}
		ins.indexSpace.Globals = append(ins.indexSpace.Globals, &global{
			Type: gs.Type,
			Val:  v,
		})
	}
	return nil
}

func (ins *Instance) buildFunctionIndexSpace() error {
	for codeIndex, typeIndex := range ins.FunctionsSection {
		if typeIndex >= uint32(len(ins.TypesSection)) {
			return fmt.Errorf("function type index out of range")
		} else if codeIndex >= len(ins.CodesSection) {
			return fmt.Errorf("code index out of range")
		}

		f := &wasmFunc{
			Signature: ins.TypesSection[typeIndex],
			Body:      ins.CodesSection[codeIndex].Body,
			NumLocal:  ins.CodesSection[codeIndex].NumLocals,
		}

		brs, err := ins.parseBlocks(f.Body)
		if err != nil {
			return fmt.Errorf("parse blocks: %w", err)
		}

		f.Blocks = brs
		ins.indexSpace.Functions = append(ins.indexSpace.Functions, f)
	}

	return nil
}

func (ins *Instance) buildMemoryIndexSpace() error {
	for _, d := range ins.Module.DataSection {
		// note: MVP restricts the size of memory index spaces to 1
		if d.MemoryIndex >= uint32(len(ins.indexSpace.Memories)) {
			return fmt.Errorf("index out of range of index space")
		} else if d.MemoryIndex >= uint32(len(ins.MemorySection)) {
			return fmt.Errorf("index out of range of memory section")
		}

		rawOffset, err := ins.execExpr(d.OffsetExpression)
		if err != nil {
			return fmt.Errorf("calculate offset: %w", err)
		}

		offset, ok := rawOffset.(int32)
		if !ok {
			return fmt.Errorf("type assertion failed")
		}

		size := int(offset) + len(d.Init)
		if ins.MemorySection[d.MemoryIndex].Max != nil && uint32(size) > *(ins.MemorySection[d.MemoryIndex].Max)*defaultPageSize {
			return fmt.Errorf("memory size out of limit %d * 64Ki", int(*(ins.MemorySection[d.MemoryIndex].Max)))
		}

		memory := ins.indexSpace.Memories[d.MemoryIndex]
		if size > len(memory) {
			next := make([]byte, size)
			copy(next, memory)
			copy(next[offset:], d.Init)
			ins.indexSpace.Memories[d.MemoryIndex] = next
		} else {
			copy(memory[offset:], d.Init)
		}
	}
	return nil
}

func (ins *Instance) buildTableIndexSpace() error {
	for _, elem := range ins.ElementsSection {
		// note: MVP restricts the size of memory index spaces to 1
		if elem.TableIndex >= uint32(len(ins.indexSpace.Tables)) {
			return fmt.Errorf("index out of range of index space")
		} else if elem.TableIndex >= uint32(len(ins.TablesSection)) {
			// this is just in case since we could assume len(SecTables) == len(indexSpace.Table)
			return fmt.Errorf("index out of range of table section")
		}

		rawOffset, err := ins.execExpr(elem.OffsetExpr)
		if err != nil {
			return fmt.Errorf("calculate offset: %w", err)
		}

		offset32, ok := rawOffset.(int32)
		if !ok {
			return fmt.Errorf("type assertion failed")
		}

		offset := int(offset32)
		size := offset + len(elem.Init)
		if ins.TablesSection[elem.TableIndex].Limits.Max != nil &&
			size > int(*(ins.TablesSection[elem.TableIndex].Limits.Max)) {
			return fmt.Errorf("table size out of limit of %d", int(*(ins.TablesSection[elem.TableIndex].Limits.Max)))
		}

		table := ins.indexSpace.Tables[elem.TableIndex]
		if size > len(table) {
			next := make([]*uint32, size)
			copy(next, table)
			for i, b := range elem.Init {
				next[i+offset] = &b
			}
			ins.indexSpace.Tables[elem.TableIndex] = next
		} else {
			for i, b := range elem.Init {
				table[i+offset] = &b
			}
		}
	}
	return nil
}

type blockType = types.FuncType

func (ins *Instance) readBlockType(r io.Reader) (*blockType, uint64, error) {
	raw, num, err := leb128.DecodeInt33AsInt64(r)
	if err != nil {
		return nil, 0, fmt.Errorf("decode int33: %w", err)
	}

	var ret *blockType
	switch raw {
	case -64: // 0x40 in original byte = nil
		ret = &blockType{}
	case -1: // 0x7f in original byte = i32
		ret = &blockType{ReturnTypes: []types.ValueType{types.ValueTypeI32}}
	case -2: // 0x7e in original byte = i64
		ret = &blockType{ReturnTypes: []types.ValueType{types.ValueTypeI64}}
	case -3: // 0x7d in original byte = f32
		ret = &blockType{ReturnTypes: []types.ValueType{types.ValueTypeF32}}
	case -4: // 0x7c in original byte = f64
		ret = &blockType{ReturnTypes: []types.ValueType{types.ValueTypeF64}}
	default:
		if raw < 0 || (raw >= int64(len(ins.TypesSection))) {
			return nil, 0, fmt.Errorf("invalid block type: %d", raw)
		}
		ret = ins.TypesSection[raw]
	}
	return ret, num, nil
}

func (ins *Instance) parseBlocks(body []byte) (map[uint64]*funcBlock, error) {
	ret := map[uint64]*funcBlock{}
	stack := make([]*funcBlock, 0)
	for pc := uint64(0); pc < uint64(len(body)); pc++ {
		rawOc := body[pc]
		if 0x28 <= rawOc && rawOc <= 0x3e { // memory load,store
			pc++
			// align
			_, num, err := leb128.DecodeUint32(bytes.NewBuffer(body[pc:]))
			if err != nil {
				return nil, fmt.Errorf("read memory align: %w", err)
			}
			pc += num
			// offset
			_, num, err = leb128.DecodeUint32(bytes.NewBuffer(body[pc:]))
			if err != nil {
				return nil, fmt.Errorf("read memory offset: %w", err)
			}
			pc += num - 1
			continue
		} else if 0x41 <= rawOc && rawOc <= 0x44 { // const instructions
			pc++
			switch instr.OpCode(rawOc) {
			case instr.OpCodeI32Const:
				_, num, err := leb128.DecodeInt32(bytes.NewBuffer(body[pc:]))
				if err != nil {
					return nil, fmt.Errorf("read immediate: %w", err)
				}
				pc += num - 1
			case instr.OpCodeI64Const:
				_, num, err := leb128.DecodeInt64(bytes.NewBuffer(body[pc:]))
				if err != nil {
					return nil, fmt.Errorf("read immediate: %w", err)
				}
				pc += num - 1
			case instr.OpCodeF32Const:
				pc += 3
			case instr.OpCodeF64Const:
				pc += 7
			}
			continue
		} else if (0x3f <= rawOc && rawOc <= 0x40) || // memory grow,size
			(0x20 <= rawOc && rawOc <= 0x24) || // variable instructions
			(0x0c <= rawOc && rawOc <= 0x0d) || // br,br_if instructions
			(0x10 <= rawOc && rawOc <= 0x11) { // call,call_indirect
			pc++
			_, num, err := leb128.DecodeUint32(bytes.NewBuffer(body[pc:]))
			if err != nil {
				return nil, fmt.Errorf("read immediate: %w", err)
			}
			pc += num - 1
			if rawOc == 0x11 { // if call_indirect
				pc++
			}
			continue
		} else if rawOc == 0x0e { // br_table
			pc++
			r := bytes.NewBuffer(body[pc:])
			nl, num, err := leb128.DecodeUint32(r)
			if err != nil {
				return nil, fmt.Errorf("read immediate: %w", err)
			}

			for i := uint32(0); i < nl; i++ {
				_, n, err := leb128.DecodeUint32(r)
				if err != nil {
					return nil, fmt.Errorf("read immediate: %w", err)
				}
				num += n
			}

			_, n, err := leb128.DecodeUint32(r)
			if err != nil {
				return nil, fmt.Errorf("read immediate: %w", err)
			}
			pc += n + num - 1
			continue
		}

		switch instr.OpCode(rawOc) {
		case instr.OpCodeBlock, instr.OpCodeIf, instr.OpCodeLoop:
			bt, num, err := ins.readBlockType(bytes.NewBuffer(body[pc+1:]))
			if err != nil {
				return nil, fmt.Errorf("read block: %w", err)
			}
			stack = append(stack, &funcBlock{
				StartAt:        pc,
				BlockType:      bt,
				BlockTypeBytes: num,
			})
			pc += num
		case instr.OpCodeElse:
			stack[len(stack)-1].ElseAt = pc
		case instr.OpCodeEnd:
			bl := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			bl.EndAt = pc
			ret[bl.StartAt] = bl
		}
	}

	if len(stack) > 0 {
		return nil, fmt.Errorf("ill-nested block exists")
	}

	return ret, nil
}