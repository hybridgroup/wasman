package config

import (
	"errors"

	"github.com/hybridgroup/wasman/tollstation"
)

const (
	// MemoryPageSize is the unit of memory length in WebAssembly,
	// and is defined as 2^16 = 65536.
	// See https://www.w3.org/TR/2019/REC-wasm-core-1-20191205/#memory-instances%E2%91%A0
	DefaultMemoryPageSize = 65536
	//DefaultMemoryPageSize = 16384 // if you need a tiny page size for tiny devices...
	// MemoryMaxPages is maximum number of pages defined (2^16).
	// See https://www.w3.org/TR/2019/REC-wasm-core-1-20191205/#grow-mem
	DefaultMemoryMaxPages = 1
	// MemoryPageSizeInBits satisfies the relation: "1 << MemoryPageSizeInBits == MemoryPageSize".
	DefaultMemoryPageSizeInBits = 16
	//DefaultMemoryPageSizeInBits = 14 // to match the tiny config version
)

var (
	// ErrShadowing wont appear if LinkerConfig.DisableShadowing is default false
	ErrShadowing = errors.New("shadowing is disabled")
)

// ModuleConfig is the config applied to the wasman.Module
type ModuleConfig struct {
	DisableFloatPoint bool
	TollStation       tollstation.TollStation
	CallDepthLimit    *uint64
	Recover           bool // avoid panic inside vm
	Logger            func(text string)
}

// LinkerConfig is the config applied to the wasman.Linker
type LinkerConfig struct {
	DisableShadowing bool // false by default
}
