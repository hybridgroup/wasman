package wasm

import "github.com/hybridgroup/wasman/types"

// Table is an instance of the table value
type Table struct {
	types.TableType
	Value []*uint32 // vec of addr to func
}
