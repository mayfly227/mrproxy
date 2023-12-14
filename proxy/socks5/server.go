package socks5

import (
	"encoding/binary"
	"errors"
	"fmt"
	"mrproxy/common"
	"mrproxy/dns"
	"mrproxy/proxy"
	"net"
	"time"
)

func init() {
	proxy.RegServer("socks", NewSocksServer)
}

type SocksServer struct {
	proxy.InNode
	query *dns.LocalQuery
}

func (c *SocksServer) Handle(conn net.Conn) (net.Conn, *common.TargetAddress, error) {
	//进行socks5握手
	if err := conn.SetReadDeadline(time.Now().Add(time.Second * 30)); err != nil {
		return nil, nil, err
	}
	buf := make([]byte, 256)

	// Read hello message
	n, err := conn.Read(buf[:])
	if err != nil || n == 0 {
		return nil, nil, fmt.Errorf("failed to read hello: %w", err)
	}
	version := buf[0]
	if version != Version5 {
		return nil, nil, fmt.Errorf("unsupported socks version %v", version)
	}

	// Write hello response
	// TODO: Support Auth
	_, err = conn.Write([]byte{Version5, AuthNone})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to write hello response: %w", err)
	}

	// Read command message
	n, err = conn.Read(buf)
	if err != nil || n < 7 { // Shortest length is 7
		return nil, nil, fmt.Errorf("failed to read command: %w", err)
	}
	cmd := buf[1]
	if cmd != CmdConnect {
		return nil, nil, fmt.Errorf("unsuppoted command %v", cmd)
	}
	addr := common.NewTargetAddress()
	addr.SetNetwork("tcp")
	l := 2
	off := 4
	switch buf[3] {
	case ATypIP4:
		l += net.IPv4len
		ip4 := make(net.IP, net.IPv4len)
		copy(ip4, buf[off:])
		//fmt.Println(ip4.String())
		addr.SetAddress(ip4.String())
	case ATypIP6:
		l += net.IPv6len
		ip6 := make(net.IP, net.IPv6len)
		copy(ip6, buf[off:])
		//fmt.Println(ip6.String())
		addr.SetAddress(ip6.String())
		return nil, nil, fmt.Errorf("not support ipv6")
	case ATypDomain:
		l += int(buf[4])
		off = 5
		domain := string(buf[off : off+l-2])
		//fmt.Println(domain)
		//FIXME
		addr.SetAddress(domain) //domain
		//ips, err := c.query.LookupIP(domain, "ip4")
		//if err != nil {
		//	slog.Error(fmt.Sprintf("error dns query:%s", err.Error()))
		//	return nil, nil, err
		//}
		//addr.SetAddress(ips[0])

	default:
		return nil, nil, fmt.Errorf("unknown address type %v", buf[3])
	}

	if len(buf[off:]) < l {
		return nil, nil, errors.New("short command request")
	}

	port2 := int(buf[off+l-2])<<8 | int(buf[off+l-1])
	port := binary.BigEndian.Uint16(buf[off+l-2 : off+l])
	if port2 != int(port) {
		panic("not equ")
	}
	addr.SetPort(int(port))

	// Write command response
	_, err = conn.Write([]byte{Version5, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to write command response: %w", err)
	}

	if addr.GetAddress() == "127.0.0.1" {
		panic("get 127")
	}
	return conn, addr, err
}

func NewSocksServer() (proxy.InNode, error) {
	r := dns.NewLocalQuery()
	s := &SocksServer{query: r}
	return s, nil
}
