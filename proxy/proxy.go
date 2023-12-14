package proxy

import (
	"mrproxy/common"
	"mrproxy/transport"
	"net"
	"strings"
)

//type func

type Listener interface {
	Listen(handler func(conn net.Conn)) (Listener, error)
	net.Listener
	Network() string
}

type InNode interface {
	Handle(conn net.Conn) (net.Conn, *common.TargetAddress, error)
}

type OutNode interface {
	Handle(conn net.Conn, addr *common.TargetAddress, dialer transport.Dialer) (net.Conn, error)
}

// ClientCreator client=out server=in
// ClientCreator is a function to create client.
type ClientCreator func() (OutNode, error)
type ServerCreator func() (InNode, error)

var (
	ClientMap = make(map[string]ClientCreator)
	ServerMap = make(map[string]ServerCreator)
)

func RegClient(name string, handler ClientCreator) {
	ClientMap[strings.ToLower(name)] = handler

}

func RegServer(name string, handler ServerCreator) {
	ServerMap[strings.ToLower(name)] = handler
}
