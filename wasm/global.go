package wasm

import "github.com/hybridgroup/wasman/types"

// Global is an instance of the global value
type Global struct {
	*types.GlobalType
	Val interface{}
}
