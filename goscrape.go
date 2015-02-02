package goscrape

import ()

type Bulk struct {
	Sess []Session
}

func NewBulk(trackers []string) Bulk {
	size := len(trackers)
	var sessions []Session = make([]Session, size)

	for i := 0; i < size; i++ {
		sessions[i] = NewConn(trackers[i])
	}
	return Bulk{sessions}
}
