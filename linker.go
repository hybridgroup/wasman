package wasman

import (
	"errors"
	"fmt"
	"math"

	"github.com/hybridgroup/wasman/config"
	"github.com/hybridgroup/wasman/utils"
	"github.com/hybridgroup/wasman/wasm"

	"github.com/hybridgroup/wasman/segments"
	"github.com/hybridgroup/wasman/types"
)

// errors on linking modules
var (
	ErrInvalidSign = errors.New("invalid signature")
)

// Primitive is a type constraint for arguments and results of host-defined functions
// that can be used with the linker.
type Primitive interface {
	int | int8 | int16 | int32 | int64 |
		uint | uint8 | uint16 | uint32 | uint64 | uintptr |
		float32 | float64
}

// Linker is a helper to instantiate new modules
type Linker struct {
	config.LinkerConfig

	Modules map[string]*Module // the built-in modules which acts as externs when instantiating coming main module
}

// NewLinker creates a new Linker
func NewLinker(config config.LinkerConfig) *Linker {
	return &Linker{
		LinkerConfig: config,
		Modules:      map[string]*Module{},
	}
}

// NewLinkerWithModuleMap creates a new Linker with the built-in modules
func NewLinkerWithModuleMap(config config.LinkerConfig, in map[string]*Module) *Linker {
	return &Linker{
		LinkerConfig: config,
		Modules:      in,
	}
}

// Define put the module on its namespace
func (l *Linker) Define(modName string, mod *Module) {
	l.Modules[modName] = mod
}

func DefineFunc(l *Linker, modName, funcName string, f func()) error {
	return l.defineFunc(modName, funcName, wrapFunc00(f), []any{}, []any{})
}

func DefineFunc01[Z Primitive](l *Linker, modName, funcName string, f func() Z) error {
	return l.defineFunc(modName, funcName, wrapFunc01(f), []any{}, []any{*new(Z)})
}

func DefineFunc10[A Primitive](l *Linker, modName, funcName string, f func(A)) error {
	return l.defineFunc(modName, funcName, wrapFunc10(f), []any{*new(A)}, []any{})
}

func DefineFunc11[A, Z Primitive](l *Linker, modName, funcName string, f func(A) Z) error {
	return l.defineFunc(modName, funcName, wrapFunc11(f), []any{*new(A)}, []any{*new(Z)})
}

func DefineFunc20[A, B Primitive](l *Linker, modName, funcName string, f func(A, B)) error {
	return l.defineFunc(modName, funcName, wrapFunc20(f), []any{*new(A), *new(B)}, []any{})
}

func DefineFunc21[A, B, Z Primitive](l *Linker, modName, funcName string, f func(A, B) Z) error {
	return l.defineFunc(modName, funcName, wrapFunc21(f), []any{*new(A), *new(B)}, []any{*new(Z)})
}

// DefineFunc puts a simple go style func into Linker's modules.
// This f should be a simply func which doesnt handle ins's fields.
func (l *Linker) defineFunc(modName, funcName string, f wasm.RawHostFunc, ins []any, outs []any) error {
	var err error
	sig := &types.FuncType{}
	sig.InputTypes, err = getTypesOf(ins)
	if err != nil {
		return err
	}
	sig.ReturnTypes, err = getTypesOf(outs)
	if err != nil {
		return err
	}

	mod, exists := l.Modules[modName]
	if !exists {
		mod = &Module{IndexSpace: new(wasm.IndexSpace), ExportSection: map[string]*segments.ExportSegment{}}
		l.Modules[modName] = mod
	}

	if l.DisableShadowing && mod.ExportSection[funcName] != nil {
		return config.ErrShadowing
	}

	mod.ExportSection[funcName] = &segments.ExportSegment{
		Name: funcName,
		Desc: &segments.ExportDesc{
			Kind:  segments.KindFunction,
			Index: uint32(len(mod.IndexSpace.Functions)),
		},
	}

	mod.IndexSpace.Functions = append(mod.IndexSpace.Functions, &wasm.HostFunc{
		Generator: func(_ *Instance) wasm.RawHostFunc {
			return f
		},
		Signature: sig,
	})

	return nil
}

func DefineGlobal[T any](l *Linker, modName, globalName string, global T) error {
	ty, err := getTypeOf(*new(T))
	if err != nil {
		return err
	}
	return l.defineGlobal(modName, globalName, ty, global)
}

// DefineGlobal will defined an external global for the main module
func (l *Linker) defineGlobal(modName, globalName string, ty types.ValueType, global any) error {
	mod, exists := l.Modules[modName]
	if !exists {
		mod = &Module{IndexSpace: new(wasm.IndexSpace), ExportSection: map[string]*segments.ExportSegment{}}
		l.Modules[modName] = mod
	}

	if l.DisableShadowing && mod.ExportSection[globalName] != nil {
		return config.ErrShadowing
	}

	mod.ExportSection[globalName] = &segments.ExportSegment{
		Name: globalName,
		Desc: &segments.ExportDesc{
			Kind:  segments.KindGlobal,
			Index: uint32(len(mod.IndexSpace.Globals)),
		},
	}

	mod.IndexSpace.Globals = append(mod.IndexSpace.Globals, &wasm.Global{
		GlobalType: &types.GlobalType{
			ValType: ty,
			Mutable: true,
		},
		Val: global,
	})

	return nil
}

// DefineTable will defined an external table for the main module
func (l *Linker) DefineTable(modName, tableName string, table []*uint32) error {
	mod, exists := l.Modules[modName]
	if !exists {
		mod = &Module{IndexSpace: new(wasm.IndexSpace), ExportSection: map[string]*segments.ExportSegment{}}
		l.Modules[modName] = mod
	}

	if l.DisableShadowing && mod.ExportSection[tableName] != nil {
		return config.ErrShadowing
	}

	mod.ExportSection[tableName] = &segments.ExportSegment{
		Name: tableName,
		Desc: &segments.ExportDesc{
			Kind:  segments.KindTable,
			Index: uint32(len(mod.IndexSpace.Tables)),
		},
	}

	mod.IndexSpace.Tables = append(mod.IndexSpace.Tables, &wasm.Table{
		TableType: *mod.TableSection[0],
		Value:     table,
	})

	return nil
}

// DefineMemory will defined an external memory for the main module
func (l *Linker) DefineMemory(modName, memName string, mem []byte) error {
	mod, exists := l.Modules[modName]
	if !exists {
		mod = &Module{IndexSpace: new(wasm.IndexSpace), ExportSection: map[string]*segments.ExportSegment{}}
		l.Modules[modName] = mod
	}

	if l.DisableShadowing && mod.ExportSection[memName] != nil {
		return config.ErrShadowing
	}

	mod.ExportSection[memName] = &segments.ExportSegment{
		Name: memName,
		Desc: &segments.ExportDesc{
			Kind:  segments.KindMem,
			Index: uint32(len(mod.IndexSpace.Memories)),
		},
	}

	mod.IndexSpace.Memories = append(mod.IndexSpace.Memories, &wasm.Memory{
		MemoryType: types.MemoryType{Min: 1, Max: utils.Uint32Ptr(config.DefaultMemoryMaxPages)},
		External:   true,
		Value:      mem,
	})

	return nil
}

// Instantiate will instantiate a Module into an runnable Instance
func (l *Linker) Instantiate(mainModule *Module) (*Instance, error) {
	return NewInstance(mainModule, l.Modules)
}

func getTypesOf(defaults []any) ([]types.ValueType, error) {
	var err error
	types := make([]types.ValueType, len(defaults))
	for i, def := range defaults {
		types[i], err = getTypeOf(def)
		if err != nil {
			return nil, err
		}
	}
	return types, nil
}

// getTypeOf converts the go type into wasm val type
func getTypeOf(def any) (types.ValueType, error) {
	switch def.(type) {
	case float64:
		return types.ValueTypeF64, nil
	case float32:
		return types.ValueTypeF32, nil
	case int32, uint32, int, int16, int8, bool:
		return types.ValueTypeI32, nil
	case int64, uint64, uintptr, uint:
		return types.ValueTypeI64, nil
	default:
		return 0x00, fmt.Errorf("invalid type: %T", def)
	}
}

func wrapFunc00(f func()) wasm.RawHostFunc {
	wrapper := func([]uint64) []uint64 {
		f()
		return []uint64{}
	}
	return wrapper
}

func wrapFunc01[Z Primitive](f func() Z) wasm.RawHostFunc {
	wrapper := func(a []uint64) []uint64 {
		r1 := f()
		return []uint64{toU(r1)}
	}
	return wrapper
}

func wrapFunc10[A Primitive](f func(A)) wasm.RawHostFunc {
	wrapper := func(a []uint64) []uint64 {
		a1 := fromU[A](a[0])
		f(a1)
		return []uint64{}
	}
	return wrapper
}

func wrapFunc11[A, Z Primitive](f func(A) Z) wasm.RawHostFunc {
	wrapper := func(a []uint64) []uint64 {
		a1 := fromU[A](a[0])
		r1 := f(a1)
		return []uint64{toU(r1)}
	}
	return wrapper
}

func wrapFunc20[A, B Primitive](f func(A, B)) wasm.RawHostFunc {
	wrapper := func(a []uint64) []uint64 {
		a1 := fromU[A](a[0])
		a2 := fromU[B](a[1])
		f(a1, a2)
		return []uint64{}
	}
	return wrapper
}

func wrapFunc21[A, B, Z Primitive](f func(A, B) Z) wasm.RawHostFunc {
	wrapper := func(a []uint64) []uint64 {
		a1 := fromU[A](a[0])
		a2 := fromU[B](a[1])
		r1 := f(a1, a2)
		return []uint64{toU(r1)}
	}
	return wrapper
}

func fromU[T Primitive](val uint64) T {
	switch any(*new(T)).(type) {
	case float32:
		return T(float32(math.Float64frombits(val)))
	case float64:
		return T(math.Float64frombits(val))
	default:
		return T(val)
	}
}

func toU[T Primitive](val T) uint64 {
	switch v := any(val).(type) {
	case float32:
		return math.Float64bits(float64(v))
	case float64:
		return math.Float64bits(v)
	default:
		return uint64(val)
	}
}
