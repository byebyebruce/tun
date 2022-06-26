package main

import (
	"flag"
	"fmt"
	"net"
	"time"

	"github.com/byebyebruce/tun/pkg/proxy"
)

var (
	s = flag.String("listen", ":9999", "listen address")
	c = flag.String("server", "127.0.0.1:9999", "connect server")
	l = flag.String("local", "", "local address")
)

func main() {
	flag.Parse()
	if len(*l) > 0 {
		errTimes := 0
		for {
			if err := connect(*c, *l); err != nil {
				errTimes++
				if errTimes >= 100 {
					return
				}
				fmt.Println("connect", err)
				time.Sleep(time.Second)
			} else {
				errTimes = 0
			}
		}
	} else {
		serve(*s)
	}
}

func connect(addr string, local string) error {
	r, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	l, err := net.Dial("tcp", local)
	if err != nil {
		return err
	}
	rc, lc := r.(*net.TCPConn), l.(*net.TCPConn)
	rc.SetKeepAlive(true)
	rc.SetKeepAlivePeriod(time.Second * 30)
	lc.SetKeepAlive(true)
	lc.SetKeepAlivePeriod(time.Second * 30)
	fmt.Println("connect", rc.RemoteAddr().String(), lc.RemoteAddr().String())
	proxy.Proxy(rc, lc)
	return nil
}

func serve(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	var c1 *net.TCPConn
	for {
		c, err := lis.Accept()
		if err != nil {
			return err
		}
		fmt.Println("incoming", c.RemoteAddr().String())
		tcpCon := c.(*net.TCPConn)
		tcpCon.SetKeepAlive(true)
		tcpCon.SetKeepAlivePeriod(time.Second * 30)

		if c1 == nil {
			c1 = tcpCon
		} else {
			temp := c1
			c1 = nil
			fmt.Println("Proxy Begin", temp.RemoteAddr().String(), "---", tcpCon.RemoteAddr().String())
			proxy.Proxy(temp, tcpCon)
			fmt.Println("Proxy End", temp.RemoteAddr().String(), "---", tcpCon.RemoteAddr().String())
		}
	}
}
