package socks5

import (
	"mrproxy/common"
	"mrproxy/proxy"
	"mrproxy/transport"
	"net"
)

func init() {
	proxy.RegClient("socks", NewClient)
}

type SocksClient struct {
	proxy.OutNode
}

func (c *SocksClient) Handle(conn net.Conn, addr *common.TargetAddress, dialer transport.Dialer) (net.Conn, error) {
	return conn, nil
}

func NewClient() (proxy.OutNode, error) {
	cc := &SocksClient{nil}
	return cc, nil
}
