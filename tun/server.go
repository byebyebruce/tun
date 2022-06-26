package tun

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/byebyebruce/tun/pkg/proxy"
	"github.com/google/uuid"
)

// Server server
type Server struct {
	addr   string
	sess   sync.Map
	remote sync.Map
}

func New(addr string) *Server {
	return &Server{
		addr: addr,
	}
}

func (s *Server) Run() error {
	handle := func(conn *net.TCPConn, l net.Listener) {
		defer func() {
			conn.Close()
			al := make([]net.Listener, 0)
			s.remote.Range(func(key, value any) bool {
				l, c := key.(net.Listener), value.(*net.TCPConn)
				if c == conn {
					al = append(al, l)
					l.Close()
				}
				return true
			})
			for _, v := range al {
				s.remote.Delete(v)
			}
		}()

		for {
			desc, err := readDesc(conn)
			if err != nil {
				fmt.Println("read err", conn.RemoteAddr().String(), err)
				return
			}

			// new conn
			switch {
			case len(desc.Address) > 0:
				go s.tunnelServer(conn, desc.Address)
			case len(desc.UUID) > 0:
				s.tunnel(conn, desc.UUID)
			default:
				conn.Close()
				return
			}
		}
	}
	return serve(s.addr, nil, handle)
}

func (s *Server) tunnel(conn *net.TCPConn, uuid string) {
	r, ok := s.sess.LoadAndDelete(uuid)
	if ok {
		client := r.(*net.TCPConn)
		proxy.Proxy(conn, client)
		fmt.Println("create tunnel", conn.RemoteAddr().String(), client.RemoteAddr().String())
	} else {
		conn.Close()
	}
}
func (s *Server) tunnelServer(client *net.TCPConn, addr string) error {
	var l net.Listener = nil
	defer func() {
		if l != nil {
			s.remote.Delete(l)
		}
		client.Close()
	}()

	onListen := func(lis net.Listener) {
		l = lis
		s.remote.Store(l, client)
	}
	handle := func(conn *net.TCPConn, l net.Listener) {
		uuid := uuid.New().String()
		s.sess.Store(uuid, conn)
		err := writeDesc(client, sessionDesc{
			UUID: uuid,
		})
		if err != nil {
			conn.Close()
			s.sess.Delete(uuid)
		}
	}
	return serve(addr, onListen, handle)
}

func serve(addr string, onListen func(l net.Listener), handle func(conn *net.TCPConn, l net.Listener)) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	if onListen != nil {
		onListen(lis)
	}
	defer lis.Close()
	fmt.Println("listening", lis.Addr().String())
	defer fmt.Println("listen stop", lis.Addr().String())
	for {
		c, err := lis.Accept()
		if err != nil {
			return err
		}
		fmt.Println("incoming", c.RemoteAddr().String())
		tcpCon := c.(*net.TCPConn)
		tcpCon.SetKeepAlive(true)
		tcpCon.SetKeepAlivePeriod(time.Second * 30)

		go handle(tcpCon, lis)
	}
}
