package main

import "C"
import (
	"fmt"
	"os"

	"github.com/hybridgroup/wasman"
	"github.com/hybridgroup/wasman/config"
)

// Run me on root folder
// go run ./examples/hostbytes
func main() {
	var ins *wasman.Instance
	linker1 := wasman.NewLinker(config.LinkerConfig{})

	message1 := []byte{0xDE, 0xAD, 0x00, 0xBE, 0xEF, 0x00, 0xBA, 0xAD, 0x00, 0xF0, 0x0D}

	err := wasman.DefineFunc01(linker1, "env", "get_host_bytes_size", func() uint32 {
		return uint32(len(message1))
	})
	if err != nil {
		panic(err)
	}

	err = wasman.DefineFunc10(linker1, "env", "get_host_bytes", func(ptr uint32) {
		copy(ins.Memory.Value[ptr:], message1)
	})
	if err != nil {
		panic(err)
	}

	message2 := append(message1, message1...)

	err = wasman.DefineFunc21(linker1, "env", "get_host_bytes_with_buffer", func(index uint32, ptr uint32) uint32 {
		if index == 0 {
			message2 = append(message1, message1...) // reset the value
		}

		length := copy(ins.Memory.Value[ptr:], message2)
		message2 = message2[length:]

		return uint32(length)
	})
	if err != nil {
		panic(err)
	}

	// cannot call host func in the host func
	err = wasman.DefineFunc20(linker1, "env", "log_message", func(ptr uint32, l uint32) {
		// string way
		// fmt.Println(C.GoString((*C.char)(unsafe.Pointer(&ins.Memory.Value[ptr])))) // not good for bytes

		// bytes way
		msg := ins.Memory.Value[ptr : ptr+l]
		fmt.Printf("%x\n", msg)
	})
	if err != nil {
		panic(err)
	}

	f, err := os.Open("examples/hostbytes.wasm")
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
		_, _, err = ins.CallExportedFunc("greet_with_size")
		if err != nil {
			panic(err)
		}

		_, _, err = ins.CallExportedFunc("greet_with_buffer")
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("mem size", len(ins.Memory.Value))
}
