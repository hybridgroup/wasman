package types

import "github.com/hybridgroup/wasman/utils"

// MemoryType classify linear memories and their size range.
// https://www.w3.org/TR/wasm-core-1/#memory-types%E2%91%A0
type MemoryType = Limits

// ReadMemoryType will read a types.MemoryType from the io.Reader
func ReadMemoryType(r utils.Reader) (*MemoryType, error) {
	return ReadLimits(r)
}
