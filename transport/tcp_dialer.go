package transport

import (
	"context"
	"fmt"
	"mrproxy/common"
	"net"
	"time"
)

type TcpDialer struct {
	Dialer
}

func (c *TcpDialer) Dial(ctx context.Context, destination net.Addr) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", destination.String(), time.Second*15)
	//conn, err := net.Dial("tcp", "39.156.66.10:80")

	if err != nil {
		return nil, fmt.Errorf("cannot dial %s %s", destination.Network(), destination.String())
	}
	return conn, nil

}

func (c *TcpDialer) Address() net.Addr {
	return common.NewTargetAddress()
}

func NewTcpDialer() Dialer {
	return &TcpDialer{}
}
