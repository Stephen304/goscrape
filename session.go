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

func newConn(url string) Session {
	conn, id, _ := udpConnect(url)
	return Session{conn, id, url}
}

func (sess Session) scrape(btihs []string) ([]Result, error) {
	if sess.Conn == nil {
		return []Result{}, errors.New("Session uninitialized.")
	}
	return udpScrape(sess.Conn, sess.ConnID, btihs)
}
