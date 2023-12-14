package freedom

import (
	"context"
	"fmt"
	"log/slog"
	"mrproxy/common"
	"mrproxy/proxy"
	"mrproxy/transport"
	"net"
)

func init() {
	slog.Debug("init freedom")
	proxy.RegClient("freedom", NewClient)
}

type FreedomClient struct {
	proxy.OutNode
}

func (c *FreedomClient) Handle(_ net.Conn, addr *common.TargetAddress, dialer transport.Dialer) (net.Conn, error) {
	dial, err := dialer.Dial(context.Background(), addr)
	slog.Debug(fmt.Sprintf("FreedomClient try to dial %s", addr.String()))
	if err != nil {
		slog.Error(fmt.Sprintf("dial to %s err errmsg=%s", addr.String(), err))
		return nil, err
	}
	return dial, nil

}

func NewClient() (proxy.OutNode, error) {
	cc := &FreedomClient{nil}
	return cc, nil
}
