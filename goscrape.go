package goscrape

import ()

type Bulk struct {
	Sess []Session
}

func NewBulk(trackers []string) Bulk {
	size := len(trackers)
	var sessions []Session = make([]Session, size)
	var channels = make([]chan Session, size)

	for i := 0; i < size; i++ {
		channels[i] = make(chan Session)
		go asyncSession(trackers[i], channels[i])
	}

	for i := 0; i < size; i++ {
		sessions[i] = <-channels[i]
	}

	return Bulk{sessions}
}

func ScrapeBulk(btihs []string) {

}

func asyncSession(url string, output chan Session) {
	output <- NewConn(url)
}
