package dns

import (
	"context"
	"fmt"
	"net"
	"time"
)

var defaultNS []string

func init() {
	defaultNS = []string{"114.114.114.114:53", "1.1.1.1:53"}
	//defaultNS = []string{"1.1.1.1:53", "8.8.8.8:53", "114.114.114.114:53"}
}

type LocalQuery struct {
	resolver *net.Resolver
}

func (c *LocalQuery) LookupIP(hostname string, network string) ([]string, error) {
	//case
	var ips []string

	ipAddresses, err := c.resolver.LookupIP(context.Background(), "ip4", hostname)
	if err != nil {
		return nil, err
	}
	for _, ipAddress := range ipAddresses {
		ips = append(ips, ipAddress.String())
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("can not query[%s] dns", hostname)
	}
	return ips, err
}

func NewLocalQuery() *LocalQuery {
	r := NewResolver()
	return &LocalQuery{resolver: r}
}

func NewResolver() *net.Resolver {
	resolver := &net.Resolver{
		PreferGo: true, // 使用 Go 的 DNS 解析实现
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			// 在这里使用多个 DNS 服务器
			dnsServers := defaultNS
			for _, server := range dnsServers {
				dialer := net.Dialer{
					Timeout: time.Second * 3,
				}
				conn, err := dialer.DialContext(ctx, "udp", server)
				if err == nil {
					return conn, nil
				}
			}
			return nil, fmt.Errorf("failed to dial any DNS server")
		},
	}
	return resolver

}
