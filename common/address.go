package common

import (
	"net"
	"strconv"
)

type TargetAddress struct {
	network string // tcp,udp
	address string // 127.0.0.1 or domain line baidu.com www.baidu.com
	port    int    // 0-65535
	net.Addr
}

func (c *TargetAddress) Error() string {
	//TODO implement me
	panic("implement me")
}

func (c *TargetAddress) SetAddress(addr string) {
	c.address = addr
}
func (c *TargetAddress) GetAddress() string {
	return c.address
}
func (c *TargetAddress) SetPort(port int) {
	c.port = port
}
func (c *TargetAddress) SetNetwork(network string) {
	c.network = network
}

func (c *TargetAddress) Network() string {
	return c.network
}

func (c *TargetAddress) Port() int {
	return c.port
}
func (c *TargetAddress) String() string {
	return c.address + ":" + strconv.Itoa(c.port)
}

func NewTargetAddress() *TargetAddress {
	return &TargetAddress{}
}
