package goscrape

import (
	"time"
)

type Bulk struct {
	Sess   []Session
	Expire time.Time
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

	return Bulk{Sess: sessions, Expire: time.Now().Add(1 * time.Minute)}
}

func (bulk *Bulk) ScrapeBulk(btihs []string) []Result {
	if time.Now().After(bulk.Expire) {
		bulk.refreshSessions()
	}
	var results []Result = make([]Result, len(btihs))
	for i := 0; i < len(results); i++ {
		results[i] = Result{btihs[i], 0, 0, 0}
	}

	for i, btih := range btihs {
		for _, sess := range bulk.Sess {
			result, err := sess.Scrape(btih)
			if err == nil {
				if result.Seeders > results[i].Seeders {
					results[i].Seeders = result.Seeders
				}
				if result.Leechers > results[i].Leechers {
					results[i].Leechers = result.Leechers
				}
				if result.Completed > results[i].Completed {
					results[i].Completed = result.Completed
				}
			}
		}
	}

	return results
}

func asyncSession(url string, output chan Session) {
	output <- NewConn(url)
}

func (bulk *Bulk) refreshSessions() {
	// Get the size of the sessions
	size := len(bulk.Sess)
	// Make channels
	var channels = make([]chan Session, size)

	// Make channels and make new sessions asynchronously
	for i := 0; i < size; i++ {
		channels[i] = make(chan Session)
		go asyncSession(bulk.Sess[i].URL, channels[i])
	}

	// Replace old sessions with new ones
	for i := 0; i < size; i++ {
		bulk.Sess[i] = <-channels[i]
	}

	bulk.Expire = time.Now().Add(1 * time.Minute)
}
