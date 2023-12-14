package slow

import (
	"io"
	"mrproxy/common"
	"net"
)

type SlowConn struct {
	net.Conn
	aeadReader io.Reader
	aeadWriter io.Writer

	key      []byte
	nonce    []byte
	bufRead  []byte
	bufWrite []byte
}

func (c *SlowConn) Read(b []byte) (n int, err error) {
	//return c.Conn.Read(b)
	if c.aeadReader != nil {
		return c.aeadReader.Read(b)
	}

	aeadCipher, err := NewAeadCipher(c.key)

	buffer := make([]byte, chunkLen+maxChunkSize)
	rw := NewAeadReadWriter(common.ReadOnlyReader{Reader: c.Conn}, nil, aeadCipher, c.nonce, c.key, buffer)
	c.aeadReader = rw
	return c.aeadReader.Read(b)
}

func (c *SlowConn) Write(b []byte) (n int, err error) {
	//return c.Conn.Write(b)

	if c.aeadWriter != nil {
		return c.aeadWriter.Write(b)
	}
	buffer := make([]byte, chunkLen+maxChunkSize)

	aeadCipher, err := NewAeadCipher(c.key)
	rw := NewAeadReadWriter(nil, common.WriteOnlyWriter{Writer: c.Conn}, aeadCipher, c.nonce, c.key, buffer)
	c.aeadWriter = rw
	return c.aeadWriter.Write(b)
}

func NewSlowConn(conn net.Conn, key []byte, nonce []byte) *SlowConn {
	return &SlowConn{Conn: conn, key: key, nonce: nonce}
}
