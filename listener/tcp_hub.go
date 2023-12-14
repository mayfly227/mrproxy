package listener

import (
	"context"
	"fmt"
	"mrproxy/common"
	"mrproxy/proxy"
	"net"
)

type ConnHandler func(conn net.Conn)

type ListenFunc func(ctx context.Context, address *common.TargetAddress, settings *common.Config, handler ConnHandler) (proxy.Listener, error)

var transportListenerCache = make(map[string]ListenFunc)

func RegisterTransportListener(protocol string, listener ListenFunc) error {
	if _, found := transportListenerCache[protocol]; found {
		return fmt.Errorf(protocol, " listener already registered.")
	}
	transportListenerCache[protocol] = listener
	return nil
}

func GetTransportListener(name string) ListenFunc {
	return transportListenerCache[name]
}

func ListenTCP(ctx context.Context, address *common.TargetAddress, settings *common.Config, handler ConnHandler) (proxy.Listener, error) {

	//protocol := settings.ProtocolName
	listenFunc := transportListenerCache["tcp"] //TODO support ws

	listener, _ := listenFunc(ctx, address, settings, handler)

	return listener, nil
}
