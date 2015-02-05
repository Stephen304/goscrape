package goscrape

import (
	"encoding/hex"
	"time"
)

/*
Stores multiple sessions and keeps track
of when to refresh the connection ID.
*/
type Bulk struct {
	Sess   []Session
	Expire time.Time
}

/*
Scrapes 1 or more trackers for 1 or more
torrents. It doesn't save the connection.
*/
func Single(urls []string, btihs []string) []Result {
	bulk := NewBulk(urls)
	return bulk.ScrapeBulk(btihs)
}

/*
Creates a new bulk object for running multiple
scrapes on the same set of trackers without recreating
a new connection each time.
*/
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

/*
Scrapes a set of info hashes from the
connections the bulk was initialized with.
*/
func (bulk *Bulk) ScrapeBulk(btihs []string) []Result {
	// Refresh sessions if it's been over a minute
	if time.Now().After(bulk.Expire) {
		bulk.refreshSessions()
	}

	// Validate the btihs and get size
	var cleanBtihs []string = make([]string, 0)
	for _, btih := range btihs {
		// Take the BTIH and convert it into bytes
		infohash, err := hex.DecodeString(btih)
		// Check errors
		if err == nil {
			if len(infohash) == 20 {
				cleanBtihs = append(cleanBtihs, btih)
			}
		}
	}

	// Make a result variable
	var results []Result = make([]Result, len(cleanBtihs))
	for i := 0; i < len(results); i++ {
		results[i] = Result{cleanBtihs[i], 0, 0, 0}
	}

	// Loop through the sessions
	for _, sess := range bulk.Sess {
		// Perform a multi scrape with all btihs on the single session
		scrape, err := sess.scrape(cleanBtihs)
		if err == nil {
			// Merge result array into results
			for i, result := range scrape {
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
	output <- newConn(url)
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

	// Update the expire time.
	bulk.Expire = time.Now().Add(1 * time.Minute)
}
