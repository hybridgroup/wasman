package utils

import "io"

type Reader interface {
	io.Reader
	io.ReaderAt
	io.Seeker
}
