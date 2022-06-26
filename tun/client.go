package tun

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/byebyebruce/tun/pkg/proxy"
)

type Client struct {
	conn       *net.TCPConn
	wg         sync.WaitGroup
	serverAddr string
}

func NewClient(serverAddr string) (*Client, error) {
	c, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return nil, err
	}
	fmt.Println("connect ok", serverAddr)
	conn := c.(*net.TCPConn)
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(time.Second * 30)
	return &Client{
		conn:       conn,
		serverAddr: serverAddr,
	}, nil
}

func (cli *Client) Run(localAddr, remoteAddr string) error {
	if err := writeDesc(cli.conn, sessionDesc{Address: remoteAddr}); err != nil {
		return err
	}
	for {
		desc, err := readDesc(cli.conn)
		if err != nil {
			return err
		}
		cli.wg.Add(1)
		go cli.tunnel(desc.UUID, cli.serverAddr, localAddr)
	}
}

func (cli *Client) tunnel(uuid, serverAddr, localAddr string) error {
	cli.wg.Done()
	r, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return err
	}
	defer r.Close()
	l, err := net.Dial("tcp", localAddr)
	if err != nil {
		return err
	}
	defer l.Close()
	rc, lc := r.(*net.TCPConn), l.(*net.TCPConn)
	rc.SetKeepAlive(true)
	rc.SetKeepAlivePeriod(time.Second * 30)
	fmt.Println("connect", rc.RemoteAddr().String(), lc.RemoteAddr().String())
	if err := writeDesc(r, sessionDesc{UUID: uuid}); err != nil {
		return err
	}
	proxy.Proxy(rc, lc)
	return nil
}
