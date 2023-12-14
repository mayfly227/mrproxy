package listener

import (
	"context"
	"fmt"
	"log/slog"
	"mrproxy/common"
	"mrproxy/proxy"
	"net"
)

func init() {

}

type TcpListener struct {
	address *common.TargetAddress
	proxy.Listener
	listener net.Listener
}

func (c *TcpListener) Accept() (net.Conn, error) {

	conn, err := c.listener.Accept()
	if err != nil {
		return nil, fmt.Errorf("listen error")

	}
	return conn, nil
}

func (c *TcpListener) Listen(handler func(conn net.Conn)) (proxy.Listener, error) {

	listener, err := net.Listen(c.address.Network(), c.address.String())
	if err != nil {
		return nil, fmt.Errorf("listen error")
	}
	slog.Info(fmt.Sprintf("listen at %s", listener.Addr()))
	c.listener = listener
	go func() {
		for {
			accept, err := listener.Accept()
			if err != nil {
				continue
				//return nil, err
			}
			go handler(accept)
		}

	}()
	return c, err
}

func (c *TcpListener) Network() string {
	return "tcp"
}

func NewTcpListener(address *common.TargetAddress) proxy.Listener {
	l := &TcpListener{
		address:  address,
		Listener: nil,
	}
	return l
}
func ListenRawTCP(ctx context.Context, address *common.TargetAddress, settings *common.Config, handler ConnHandler) (proxy.Listener, error) {
	lis := NewTcpListener(address)
	listen, err := lis.Listen(handler)
	if err != nil {
		return nil, err
	}
	return listen, nil
}

func init() {
	err := RegisterTransportListener("tcp", ListenRawTCP)
	if err != nil {
		panic(err)
		return
	}
}
