package transport

import (
	"context"
	"net"
)

type Dialer interface {
	Dial(ctx context.Context, destination net.Addr) (net.Conn, error)

	Address() net.Addr
}
