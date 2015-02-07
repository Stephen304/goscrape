# GoScrape [![GoDoc](https://godoc.org/github.com/Stephen304/goscrape?status.svg)](https://godoc.org/github.com/Stephen304/goscrape) [![Build Status](https://travis-ci.org/Stephen304/goscrape.svg?branch=master)](https://travis-ci.org/Stephen304/goscrape)
It scrapes udp trackers

## Usage
Get the package

    go get github.com/Stephen304/goscrape

Then import it

    import ("github.com/Stephen304/goscrape")

Then create a bulk scraper

    scraper := NewBulk([]string{
      "udp://tracker.example.com:80",
      "udp://tracker.example2.com:80"})

Then lets scrape some stuff

    result := scraper.ScrapeBulk([]string{
      "40Character-BitTorrent-Info-Hash-Is-Here",
      "40Character-BitTorrent-Info-Hash-Is-Here",
      "40Character-BitTorrent-Info-Hash-Is-Here",
    })

What did we get?

    for _, one := range result {
      fmt.Println(one)
    }

    {40Character-BitTorrent-Info-Hash-Is-Here 17281 5105 304430}
    {40Character-BitTorrent-Info-Hash-Is-Here 11169 7439 70603}
    {40Character-BitTorrent-Info-Hash-Is-Here 7252 2745 93562}
