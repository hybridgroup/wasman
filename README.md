# WASMan (WebAssembly Manager)

[![](https://godoc.org/github.com/hybridgroup/wasman?status.svg)](http://godoc.org/github.com/hybridgroup/wasman)
[![Go Report Card](https://goreportcard.com/badge/github.com/hybridgroup/wasman)](https://goreportcard.com/report/github.com/hybridgroup/wasman)
![CI](https://github.com/hybridgroup/wasman/workflows/CI/badge.svg)

Another wasm interpreter engine for gophers.

This is a substantially modified fork of https://github.com/c0mm4nd/wasman by way of https://github.com/orsinium-forks/wasman for the purpose of major bugfixing.

## Usage

### Executable

Install

```bash
go install github.com/hybridgroup/wasman/cmd/wasman
```

```bash
$ wasman -h
Usage of ./wasman:
  -extern-files string
        external modules files
  -func string
        main func (default "main")
  -main string
        main module (default "module.wasm")
  -max-toll uint
        the maximum toll in simple toll station
```

Example: [numeric.wasm](https://github.com/hybridgroup/minimum-wasm-rs/releases/latest)

```bash
$ wasman -main numeric.wasm -func fib 20 # calc the fibonacci number
{
  type: i32
  result: 6765
  toll: 315822
}
```

If we limit the max toll, it will panic when overflow.

```bash
$ wasman -main numeric.wasm -max-toll 300000 -func fib 20
panic: toll overflow

goroutine 1 [running]:
main.main()
        /home/ubuntu/Desktop/wasman/cmd/wasman/main.go:85 +0x87d
```

### Go Embedding

[![PkgGoDev](https://pkg.go.dev/badge/github.com/hybridgroup/wasman)](https://pkg.go.dev/github.com/hybridgroup/wasman)

#### Example

*Look for examples?*

They are in [examples folder](./examples)

## TODOs

- add more complex examples
