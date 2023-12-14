package listener

import (
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"mrproxy/proxy"
	"net"
	"net/http"
	"time"
)

func init() {

}

type WsListener struct {
	address net.Addr
	proxy.Listener
	listener net.Listener
}

type connection struct {
	conn   *websocket.Conn
	reader io.Reader
}

// Read implements net.Conn.Read()
func (c *connection) Read(b []byte) (int, error) {
	for {
		reader, err := c.getReader()
		if err != nil {
			return 0, err
		}

		nBytes, err := reader.Read(b)

		//if errors.Cause(err) == io.EOF {
		//	c.reader = nil
		//	continue
		//}
		return nBytes, err
	}
}

func (c *connection) getReader() (io.Reader, error) {
	if c.reader != nil {
		return c.reader, nil
	}

	_, reader, err := c.conn.NextReader()
	if err != nil {
		return nil, err
	}
	c.reader = reader
	return reader, nil
}

// Write implements io.Writer.
func (c *connection) Write(b []byte) (int, error) {
	if err := c.conn.WriteMessage(websocket.BinaryMessage, b); err != nil {
		return 0, err
	}
	return len(b), nil
}

func (c *connection) Close() error {
	var errors []interface{}
	if err := c.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Second*5)); err != nil {
		errors = append(errors, err)
	}
	if err := c.conn.Close(); err != nil {
		errors = append(errors, err)
	}
	if len(errors) > 0 {
		return fmt.Errorf("failed to close connection")
	}
	return nil
}

func (c *connection) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *connection) RemoteAddr() net.Addr {
	//return c.remoteAddr
	return nil
}

func (c *connection) SetDeadline(t time.Time) error {
	if err := c.SetReadDeadline(t); err != nil {
		return err
	}
	return c.SetWriteDeadline(t)
}

func (c *connection) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *connection) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

// --------------------------------
func (listen *WsListener) Accept() (net.Conn, error) {

	conn, err := listen.listener.Accept()
	if err != nil {
		return nil, fmt.Errorf("listen error")

	}
	return conn, nil
}

func (listen *WsListener) Listen(handler func(conn net.Conn)) (proxy.Listener, error) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		// 解决跨域问题
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	http.HandleFunc("/echo", func(writer http.ResponseWriter, request *http.Request) {
		c, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		handler(&connection{conn: c, reader: nil})
	})
	err := http.ListenAndServe("127.0.0.1:8901", nil)
	if err != nil {
		return nil, nil
	}
	return nil, nil
}

func (listen *WsListener) Network() string {
	return "tcp"
}

func NewWsListener() proxy.Listener {
	l := &WsListener{
		address:  nil,
		Listener: nil,
	}
	return l
}
