package goscrape

import (
	"errors"
	"net"
)

/*
Session stores the details of a single tracker session.
This includes the connection object, the connection ID,
and the URL to use for reconnecting.
*/
type Session struct {
	Conn   *net.UDPConn
	ConnID uint64
	URL    string
}

/*
Result represents the scrape result for a single torrent.
It includes the 40 character base64 bit torrent info hash,
Seeders, Leechers, and Completed counts.
*/
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
		return []Result{}, errors.New("session uninitialized")
	}
	return udpScrape(sess.Conn, sess.ConnID, btihs)
}
