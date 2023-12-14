package slow

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
)

type Aead struct {
	reader io.Reader
	writer io.Writer
	cipher cipher.AEAD
	//cipher.AEAD
	key   []byte
	nonce []byte
	buf   []byte //max=2**15=32768
	count uint16
}

func (c *Aead) Read(b []byte) (n int, err error) {
	//return c.reader.Read(b)
	//解密从slow客户端发来的数据，获取长度+加密数据
	var dataLenArr [chunkLen]byte
	_, err = io.ReadFull(c.reader, dataLenArr[:])
	if err != nil {
		return 0, err
	}
	//这个长度包括了数据+overhead的大小
	dataLen := binary.BigEndian.Uint16(dataLenArr[:])
	// if length == 0, then this is the end
	l := binary.BigEndian.Uint16(dataLenArr[:])
	if l == 0 {
		panic("raead 0")
		return 0, nil
	}
	buf := c.buf
	_, err = io.ReadFull(c.reader, buf[:dataLen])

	//解密数据(GCM数据加密后大小为12+实际数据大小+16)
	_, err = c.cipher.Open(buf[:0], c.nonce[:], buf[:dataLen], nil)
	if err != nil {
		slog.Info(fmt.Sprintf("open err %s\n", err.Error()))
		return 0, err
	}

	//实际的数据大小为 dataLen - gcm.overhead
	realDataLen := int(dataLen) - c.cipher.Overhead()
	buf = buf[:realDataLen]
	copy(b, c.buf[:realDataLen])

	return realDataLen, err
}

func (c *Aead) Write(b []byte) (n int, err error) {
	//return c.writer.Write(b)
	//加密从freedom发来的数据，加密(长度+加密数据)，发送到slow客户端
	reader := bytes.NewBuffer(b)
	for {
		//fmt.Printf("write to= %s\n", string(b))
		buf := c.buf
		payloadBuf := buf[chunkLen : chunkLen+maxChunkSize-c.cipher.Overhead()]
		// 最多读32768-2-16
		read, errr := reader.Read(payloadBuf)
		if read > 0 {
			n += read

			//fmt.Println("read", read)
			buf = buf[:2+read+c.cipher.Overhead()]
			payloadBuf = payloadBuf[:read]
			binary.BigEndian.PutUint16(buf[:chunkLen], uint16(read+c.cipher.Overhead()))

			c.cipher.Seal(payloadBuf[:0], c.nonce[:], payloadBuf, nil)

			_, errw := c.writer.Write(buf)
			if errw != nil {
				err = errw
				break
			}

		}
		if errr != nil {
			if errr != io.EOF { // ignore EOF as per io.ReaderFrom contract
				err = errr
			}
			break
		}
	}
	return n, err

}

func NewAeadReadWriter(reader io.Reader, writer io.Writer, aead cipher.AEAD, nonce []byte, key []byte, buf []byte) *Aead {

	return &Aead{
		reader: reader,
		writer: writer,
		cipher: aead,
		key:    make([]byte, 16),
		nonce:  make([]byte, 12),
		buf:    buf,
		count:  0,
	}
}

func NewAeadCipher(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 创建GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return gcm, nil
}
