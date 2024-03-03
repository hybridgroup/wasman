package wasman

import (
	"bytes"

	"github.com/hybridgroup/wasman/config"
	"github.com/hybridgroup/wasman/utils"
	"github.com/hybridgroup/wasman/wasm"
)

// Module is same to wasm.Module
type Module = wasm.Module

// NewModule is a wrapper to the wasm.NewModule
func NewModule(config config.ModuleConfig, r utils.Reader) (*Module, error) {
	return wasm.NewModule(config, r)
}

// NewModuleFromBytes is a wrapper to the wasm.NewModule that avoids having to
// make a copy of bytes that are already in memory.
func NewModuleFromBytes(config config.ModuleConfig, b []byte) (*Module, error) {
	return wasm.NewModule(config, bytes.NewReader(b))
}
