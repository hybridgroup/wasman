package main

import "C"
import (
	"fmt"
	"os"
	"unsafe"

	"github.com/hybridgroup/wasman"
	"github.com/hybridgroup/wasman/config"
)

// Run me on root folder
// go run ./examples/hoststring
func main() {
	var ins *wasman.Instance
	linker1 := wasman.NewLinker(config.LinkerConfig{})

	//err := linker1.DefineMemory("env", "memory", make([]byte, 10))

	err := wasman.DefineFunc01(linker1, "env", "host_string", func() uint32 {
		message := "WASMan"

		ret, _, err := ins.CallExportedFunc("allocate", uint64(len(message)+1))
		if err != nil {
			panic(err)
		}

		copy(ins.Memory.Value[ret[0]:], append([]byte(message), byte(0))) // act as a string for rust's CStr::from_ptr(ptr)

		return uint32(ret[0])
	})
	if err != nil {
		panic(err)
	}

	// cannot call host func in the host func
	err = wasman.DefineFunc20(linker1, "env", "log_message", func(ptr uint32, l uint32) {
		// string way
		fmt.Println(C.GoString((*C.char)(unsafe.Pointer(&ins.Memory.Value[ptr]))))

		// bytes way
		msg := ins.Memory.Value[ptr : ptr+l]
		fmt.Println(string(msg))
	})
	if err != nil {
		panic(err)
	}

	f, err := os.Open("examples/hoststring.wasm")
	if err != nil {
		panic(err)
	}

	module, err := wasman.NewModule(config.ModuleConfig{}, f)
	if err != nil {
		panic(err)
	}
	ins, err = linker1.Instantiate(module)
	if err != nil {
		panic(err)
	}

	for range make([]byte, 999) {
		_, _, err = ins.CallExportedFunc("greet")
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("mem size", len(ins.Memory.Value))
}
