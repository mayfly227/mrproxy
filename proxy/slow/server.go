package slow

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"mrproxy/common"
	"mrproxy/dns"
	"mrproxy/proxy"
	"net"
	"time"
)

func init() {
	proxy.RegServer("slow", NewSlowServer)
}

type SlowServer struct {
	proxy.InNode
	query *dns.LocalQuery
}

type Info struct {
	key   [16]byte
	iv    [16]byte
	nonce [12]byte
}

func (c *SlowServer) Handle(conn net.Conn) (net.Conn, *common.TargetAddress, error) {
	//进行slow握手
	if err := conn.SetReadDeadline(time.Now().Add(time.Second * 30)); err != nil {
		return nil, nil, err
	}

	//解密指定请求部分
	buf := make([]byte, 17+255+16)
	_, err := io.ReadFull(conn, buf[:17])
	if err != nil {
		return nil, nil, err
	}

	info := &Info{
		key:   [16]byte{0},
		iv:    [16]byte{0},
		nonce: [12]byte{0},
	}
	block, err := aes.NewCipher(info.key[:])
	if err != nil {
		return nil, nil, err
	}
	stream := cipher.NewCFBDecrypter(block, info.iv[:])
	stream.XORKeyStream(buf[:17], buf[:17])

	if buf[0] != 0x01 {
		return nil, nil, fmt.Errorf("version is 1")
	}
	var nonce [12]byte
	copy(nonce[:], buf[1:13])

	//only support tcp
	if buf[13] != 0x01 {
		return nil, nil, fmt.Errorf("only support tcp")
	}
	addr := common.NewTargetAddress()
	addr.SetNetwork("tcp")
	//获取端口号
	port := int(buf[14])<<8 | int(buf[15])
	port2 := binary.BigEndian.Uint16(buf[14:16])
	if port2 != port2 {
		panic("!")
	}
	addr.SetPort(port)

	// 获取地址类型T
	switch buf[16] {
	//ipv4
	case 0x01:
		//读取4个字节
		rn, err := io.ReadFull(conn, buf[17:21])
		if rn != 4 {
			panic(fmt.Sprintf("rn!=4 rn=%d", rn))
		}
		if err != nil {
			return nil, nil, err
		}
		stream.XORKeyStream(buf[17:17+4], buf[17:17+4])
		ip4 := make(net.IP, net.IPv4len)
		copy(ip4, buf[17:17+4])
		addr.SetAddress(ip4.String())
	//domain
	case 0x02:
		_, err := io.ReadFull(conn, buf[17:18])
		if err != nil {
			errMsg := fmt.Sprintf("read domain size error")
			slog.Error(errMsg)
			return nil, nil, err
		}
		stream.XORKeyStream(buf[17:18], buf[17:18])
		domainLen := int(buf[17])
		_, err = io.ReadFull(conn, buf[18:18+domainLen])
		if err != nil {
			errMsg := fmt.Sprintf("read domain error")
			slog.Error(errMsg)
			return nil, nil, err
		}
		//dns resolve
		stream.XORKeyStream(buf[18:18+domainLen], buf[18:18+domainLen])

		domain := string(buf[18 : 18+domainLen])
		slog.Debug(fmt.Sprintf("try to resolve domain:%s", domain))
		ip, err := c.query.LookupIP(domain, "ip4")
		if err != nil {
			return nil, nil, err
		}
		addr.SetAddress(ip[0])
		//return nil, nil, fmt.Errorf("not support domain")
	//ipv6
	case 0x03:
		stream.XORKeyStream(buf[17:17+16], buf[17:17+16])
		ip6 := make(net.IP, net.IPv6len)
		copy(ip6, buf[17:17+16])
		//fmt.Println(ip6.String())
		addr.SetAddress(ip6.String())
	default:
		return nil, nil, fmt.Errorf("not support T")
	}
	//return conn, addr, nil
	//地址解析完成，开始真正数据的转发
	wrapConn := NewSlowConn(conn, info.key[:], info.nonce[:])

	return wrapConn, addr, nil
}

func NewSlowServer() (proxy.InNode, error) {
	r := dns.NewLocalQuery()
	s := &SlowServer{query: r}
	return s, nil
}
