package common

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"time"
)

// Relay copies between left and right bidirectionally.
func Relay(leftConn, rightConn net.Conn) {
	ch := make(chan error)
	//lc是浏览器 rc是要代理访问的远程端
	go func() {
		// Wrapping to avoid using *net.TCPConn.(ReadFrom)
		// See also https://github.com/Dreamacro/clash/pull/1209
		wr, err := io.Copy(WriteOnlyWriter{Writer: leftConn}, ReadOnlyReader{Reader: rightConn})
		if err != nil {
			slog.Debug(fmt.Sprintf("one io copy err %s write=%d", err, wr))
		}
		leftConn.SetReadDeadline(time.Now().Add(time.Second * 0))
		//err = leftConn.SetReadDeadline(time.Now())
		//if err != nil {
		//	ch <- err
		//}
		//fmt.Printf("error %s\n", err)
		ch <- err
		slog.Debug("one relay coro done")
	}()

	wr, err := io.Copy(WriteOnlyWriter{Writer: rightConn}, ReadOnlyReader{Reader: leftConn})
	if err != nil {
		slog.Debug(fmt.Sprintf("second io copy err %s write=%d", err, wr))
	}
	rightConn.SetReadDeadline(time.Now().Add(time.Second * 0))
	<-ch
	slog.Debug(fmt.Sprintf("relay coro done"))
}
func Relay2(leftConn, rightConn net.Conn) {
	ch := make(chan error)

	go func() {
		// Wrapping to avoid using *net.TCPConn.(ReadFrom)
		// See also https://github.com/Dreamacro/clash/pull/1209
		_, err := io.Copy(WriteOnlyWriter{Writer: leftConn}, ReadOnlyReader{Reader: rightConn})
		leftConn.SetReadDeadline(time.Now())
		ch <- err
	}()

	io.Copy(WriteOnlyWriter{Writer: rightConn}, ReadOnlyReader{Reader: leftConn})
	rightConn.SetReadDeadline(time.Now())
	<-ch
}
