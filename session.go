package goscrape

import (
	"errors"
	"net"
)

type Session struct {
	Conn   *net.UDPConn
	ConnID uint64
	URL    string
}

type Result struct {
	Btih      string
	Seeders   int
	Leechers  int
	Completed int
}

func NewConn(url string) Session {
	conn, id, _ := UDPConnect(url)
	return Session{conn, id, url}
}

func (sess Session) Scrape(btih string) (Result, error) {
	if sess.Conn == nil {
		return Result{"", 0, 0, 0}, errors.New("Session uninitialized.")
	}
	return UDPScrape(sess.Conn, sess.ConnID, btih)
}
