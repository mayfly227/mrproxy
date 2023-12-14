package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"mrproxy/common"
	"mrproxy/dns"
	"mrproxy/listener"
	"mrproxy/proxy"
	_ "mrproxy/proxy/freedom"
	_ "mrproxy/proxy/slow"
	_ "mrproxy/proxy/socks5"
	"mrproxy/transport"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Hub struct {
	listener proxy.Listener
}

// func Start() error {
// ctx := context.Background()
// hub, err := internet.ListenTCP(ctx, w.address, w.port, w.stream, func(conn stat.Connection) {
// go w.callback(conn)
// })
// if err != nil {
// return newError("failed to listen TCP on ", w.port).AtWarning().Base(err)
// }
// w.hub = hub
// return nil
// }
func main() {
	//init work dir
	common.SetWorkDir()
	// init Log
	var programLevel = new(slog.LevelVar)
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: programLevel})
	slog.SetDefault(slog.New(h))
	programLevel.Set(slog.LevelDebug)

	var inNode string
	var outNode string
	var configFile string

	// 给命令行参数绑定变量
	flag.StringVar(&configFile, "c", "", "config file path")

	// 解析命令行参数
	flag.Parse()
	// 初始化配置
	common.InitGlobalConfig(configFile)
	inNode = common.GetConfig().GetNode("inNode").(map[string]any)["protocol"].(string)
	outNode = common.GetConfig().GetNode("outNode").(map[string]any)["protocol"].(string)
	//TODO 配置中心

	listen := common.GetConfig().GetNode("inNode").(map[string]any)["listen"].(string)
	port := common.GetConfig().GetNode("inNode").(map[string]any)["port"].(float64)

	//port =
	listenAddr := common.NewTargetAddress()
	listenAddr.SetNetwork("tcp")
	listenAddr.SetAddress(listen)
	listenAddr.SetPort(int(port))

	server, _ := proxy.ServerMap[inNode]()
	client, _ := proxy.ClientMap[outNode]()
	freedom, _ := proxy.ClientMap["freedom"]()
	dnsQuery := dns.NewLocalQuery()
	cache := common.NewKVCache(time.Second * 3600 * 8)
	_, err := listener.ListenTCP(context.TODO(), listenAddr, nil, func(conn net.Conn) {
		//slog.Debug(fmt.Sprintf("new conn from %s", conn.RemoteAddr()))
		defer func(conn net.Conn) {
			err := conn.Close()
			if err != nil {
				slog.Warn(fmt.Sprintf("close lc conn error,err=%s", err.Error()))
			}
		}(conn)

		lc, remoteAddr, err := server.Handle(conn)

		if err != nil {
			slog.Warn(fmt.Sprintf("server.Handle(conn) error,err=%s", err.Error()))
			return
		}
		//TODO 分流
		ip := net.ParseIP(remoteAddr.GetAddress())

		switch ip {
		//domain
		case nil:
			if val, ok := cache.Get(remoteAddr.GetAddress()); ok {
				slog.Debug(fmt.Sprintf("get cache for %s:", remoteAddr.GetAddress()))
				ip = net.ParseIP(val.(string))
				break
			}
			ips, err := dnsQuery.LookupIP(remoteAddr.GetAddress(), "ip4")
			if err != nil {
				ip = net.ParseIP("8.8.8.8")
				break
			}
			ip = net.ParseIP(ips[0])
			cache.Set(remoteAddr.GetAddress(), ip.String())
		default:
			//ip
		}

		mmdb := common.Instance()
		record, err1 := mmdb.Country(ip)
		if err1 != nil {
			return
		}
		code := record.Country.IsoCode

		var clientSwitch proxy.OutNode
		clientSwitch = freedom

		slog.Info(fmt.Sprintf("[%s] ===> ioscode[%s]", remoteAddr.GetAddress(), code))

		mode := "Rule"
		switch mode {
		case "Rule":
			if code != "CN" {
				clientSwitch = client
			}
			if code == "" {
				clientSwitch = client
			}
		case "Global":
			clientSwitch = client
		default:
			clientSwitch = freedom
		}

		if clientSwitch == client {
			slog.Info(fmt.Sprintf("[%s] ===> choose proxy", remoteAddr.GetAddress()))

		} else {
			slog.Info(fmt.Sprintf("[%s] ===> choose direct", remoteAddr.GetAddress()))
		}
		rc, err := clientSwitch.Handle(conn, remoteAddr, transport.NewTcpDialer())
		if err != nil {
			slog.Info(fmt.Sprintf("client handle error =%s", err.Error()))
			return
		}
		defer func(rc net.Conn) {
			err := rc.Close()
			if err != nil {
				slog.Warn(fmt.Sprintf("close rc conn error,err=%s", err.Error()))
				return
			}
		}(rc)
		//lc是浏览器 rc是要代理访问的远程端
		//common.Relay(lc, rc)
		common.Relay2(lc, rc)
	})
	if err != nil {
		return
	}

	{
		osSignals := make(chan os.Signal, 1)
		signal.Notify(osSignals, os.Interrupt, os.Kill, syscall.SIGTERM)
		<-osSignals
		cache.StopCleanup()
	}

}
