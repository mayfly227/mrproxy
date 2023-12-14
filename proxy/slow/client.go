package slow

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"fmt"
	"log/slog"
	"mrproxy/common"
	"mrproxy/proxy"
	"mrproxy/transport"
	"net"
)

func init() {
	proxy.RegClient("slow", NewSlowClient)
}

type SlowClient struct {
	proxy.OutNode
	targetAddr *common.TargetAddress
}

type InfoClient struct {
	key   [16]byte
	iv    [16]byte
	nonce [12]byte
}

func (c *SlowClient) Handle(_ net.Conn, addr *common.TargetAddress, dialer transport.Dialer) (net.Conn, error) {

	//addr是真正要代理请求的地址
	// add是slow服务器的地址
	dial, err := dialer.Dial(context.Background(), c.targetAddr)
	if err != nil {
		return nil, err
	}

	request := make([]byte, 17+255)
	request[0] = 0x01
	//request[1:13] = 0
	request[13] = 0x01
	binary.BigEndian.PutUint16(request[14:16], uint16(addr.Port()))
	// ipv4 v6 domain
	request[16] = 0x01 //ipv4
	// 包装IP
	index := 17
	ip := net.ParseIP(addr.GetAddress())

	switch ip {
	case nil:
		//domain
		domainLen := len(addr.GetAddress())
		request[16] = 0x02 //domain
		request[17] = byte(domainLen)

		if domainLen > 255 {
			panic(fmt.Sprintf("too much domain %s", addr.GetAddress()))
		}
		byteArrayDomain := []byte(addr.GetAddress())
		copy(request[18:18+domainLen], byteArrayDomain)
		index += domainLen + 1

	default:
		//ip
		if ip.To4() != nil {
			request[16] = 0x01
			copy(request[17:21], ip.To4())
			slog.Info(fmt.Sprintf("slow client get target IPv4 address = %s", ip.String()))
			index += 4
		} else if ip.To16() != nil {
			request[16] = 0x03
			copy(request[17:17+16], ip.To16())
			index += 16
			slog.Error("not support ipv6")
			return nil, fmt.Errorf("not support ipv6")
		} else {
			panic("Unknown IP version")
		}
	}

	//加密request
	info := InfoClient{
		key:   [16]byte{},
		iv:    [16]byte{},
		nonce: [12]byte{},
	}
	block, err := aes.NewCipher(info.key[:])
	if err != nil {
		return nil, err
	}
	stream := cipher.NewCFBEncrypter(block, info.iv[:])
	stream.XORKeyStream(request[:index], request[:index])

	_, err = dial.Write(request[:index])
	if err != nil {
		return nil, err
	}
	//return dial, nil
	//地址解析完成，开始真正数据的转发
	wrapConn := NewSlowConn(dial, info.key[:], info.nonce[:])
	return wrapConn, nil
}

func NewSlowClient() (proxy.OutNode, error) {
	targetAddr := common.NewTargetAddress()

	tAddr := common.GetConfig().GetNode("outNode").(map[string]any)["ip"].(string)
	tPort := common.GetConfig().GetNode("outNode").(map[string]any)["port"].(float64)

	targetAddr.SetAddress(tAddr)
	targetAddr.SetPort(int(tPort))
	cc := &SlowClient{targetAddr: targetAddr}
	return cc, nil
}
