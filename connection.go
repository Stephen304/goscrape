package goscrape

import (
	"net"
)

type Connection struct {
	Conn   *net.UDPConn
	ConnID uint64
	URL    string
}

func NewConn(url string) Connection {
	conn, id, _ := UDPConnect(url)
	return Connection{conn, id, url}
}
