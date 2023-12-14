package common

import "io"

type ReadOnlyReader struct {
	io.Reader
}

type WriteOnlyWriter struct {
	io.Writer
}
