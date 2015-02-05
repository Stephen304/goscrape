package goscrape

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"math/rand"
	"net"
	"strings"
	"time"
)

func UDPConnect(url string) (*net.UDPConn, uint64, error) {
	// Remove udp:// and trailing / if it's there
	if strings.HasPrefix(url, "udp://") || strings.HasPrefix(url, "UDP://") {
		url = url[6:]
	}
	if strings.HasSuffix(url, "/") {
		url = url[:len(url)-1]
	}

	// Get server address
	serverAddr, err := net.ResolveUDPAddr("udp", url)
	if err != nil {
		return nil, 0, errors.New("Couldn't resolve address.")
	}

	// Dial the server
	conn, err := net.DialUDP("udp", nil, serverAddr)

	// Set a timeout
	err = conn.SetDeadline(time.Now().Add(1 * time.Second))
	if err != nil {
		return nil, 0, errors.New("Couldn't set timeout")
	}

	/**
	 * Here we send the first UDP packet.
	 * We send a request to connect.
	 * The server should respond with a connection ID
	 * which we will use for subsequent requests.
	 */

	// Set connection ID
	var connID uint64 = 0x41727101980
	// This means connect
	var action uint32 = 0

	// Generate a random transaction ID
	transID := rand.Uint32()

	// Make a request buffer
	reqBuffer := new(bytes.Buffer)

	// Write out the conn ID
	err = binary.Write(reqBuffer, binary.BigEndian, connID)
	if err != nil {
		return nil, 0, errors.New("Couldn't write connection ID.")
	}
	// Write out the action
	err = binary.Write(reqBuffer, binary.BigEndian, action)
	if err != nil {
		return nil, 0, errors.New("Couldn't write action.")
	}
	// Write out the transaction id
	err = binary.Write(reqBuffer, binary.BigEndian, transID)
	if err != nil {
		return nil, 0, errors.New("Couldn't write transaction ID.")
	}

	// Write the request buffer to the connection
	_, err = conn.Write(reqBuffer.Bytes())
	if err != nil {
		return nil, 0, errors.New("Couldn't write to connection.")
	}

	// Make a response slice
	respSlice := make([]byte, 16)

	// Get the response from the connection
	var respLen int
	respLen, err = conn.Read(respSlice)
	if err != nil {
		return nil, 0, errors.New("Couldn't get response.")
	}
	// Verify that the response is 16 bytes
	if respLen != 16 {
		return nil, 0, errors.New("Unexpected response size.")
	}

	// Make a response buffer
	respBuffer := bytes.NewBuffer(respSlice)

	// Get the action id response
	var respAction uint32
	err = binary.Read(respBuffer, binary.BigEndian, &respAction)
	if err != nil {
		return nil, 0, errors.New("Couldn't read action response.")
	}
	// Response action must match requested action, 0 for connect
	if respAction != 0 {
		return nil, 0, errors.New("Unexpected response action.")
	}

	// Get the response transaction ID
	var respTransID uint32
	err = binary.Read(respBuffer, binary.BigEndian, &respTransID)
	if err != nil {
		return nil, 0, errors.New("Couldn't read transaction response.")
	}
	// Response transaction ID must match what we sent
	if respTransID != transID {
		return nil, 0, errors.New("Unexpected response transactionID.")
	}

	// Get the connection ID we need
	err = binary.Read(respBuffer, binary.BigEndian, &connID)
	if err != nil {
		return nil, 0, errors.New("Couldn't read connection ID.")
	}

	return conn, connID, nil
}

func UDPScrape(conn *net.UDPConn, connID uint64, btihs []string) ([]Result, error) {
	/**
	 * Here we send the scrape request.
	 * We attach the connection ID from before.
	 * The server responds with seed/leech counts.
	 */

	var empty []Result = []Result{}
	var results []Result = make([]Result, len(btihs))

	// Set a timeout
	err := conn.SetDeadline(time.Now().Add(1 * time.Second))
	if err != nil {
		return empty, errors.New("Couldn't set timeout.")
	}

	// Make a new random transaction ID
	transactionID := rand.Uint32()

	// Make a request buffer
	scrapeReq := new(bytes.Buffer)

	// Write the connection ID from earlier to the buffer
	err = binary.Write(scrapeReq, binary.BigEndian, connID)
	if err != nil {
		return empty, errors.New("Scrape Failed.")
	}

	// Write action 2 for scrape
	err = binary.Write(scrapeReq, binary.BigEndian, uint32(2))
	if err != nil {
		return empty, errors.New("Scrape Failed.")
	}

	// Write the new transaction ID
	err = binary.Write(scrapeReq, binary.BigEndian, transactionID)
	if err != nil {
		return empty, errors.New("Scrape Failed.")
	}

	// Loop through info hashes
	for _, btih := range btihs {
		infohash, err := hex.DecodeString(btih)
		if err != nil {
			return empty, errors.New("Couldn't decode base64.")
		}
		// Write the 20 byte info hash
		err = binary.Write(scrapeReq, binary.BigEndian, infohash)
		if err != nil {
			return empty, errors.New("Scrape Failed.")
		}
	}

	// Write the packet to the server
	_, err = conn.Write(scrapeReq.Bytes())
	if err != nil {
		return empty, errors.New("Coudn't write packet.")
	}

	// Calculate how big the response packet should be
	const minimumResponseLen = 8
	const peerDataSize = 12
	expectedResponseLen := minimumResponseLen + (peerDataSize * len(btihs))

	// Make a response byte slice
	responseBytes := make([]byte, expectedResponseLen)

	// Read the response into the byte slice and get its length
	var responseLen int
	responseLen, err = conn.Read(responseBytes)
	if err != nil {
		return empty, errors.New("Scrape Failed.")
	}

	// Validate the response length
	if responseLen < minimumResponseLen {
		return empty, errors.New("Unexpected response size.")
	}

	// Write the response to a buffer
	response := bytes.NewBuffer(responseBytes)

	// Get the action code from the response
	var responseAction uint32
	err = binary.Read(response, binary.BigEndian, &responseAction)
	if err != nil {
		return empty, errors.New("Scrape Failed.")
	}
	// Response action should be 2 for scrape
	if responseAction != 2 {
		return empty, errors.New("Unexpected response action.")
	}

	// Get the transaction ID from the response
	var responseTransactionID uint32
	err = binary.Read(response, binary.BigEndian, &responseTransactionID)
	if err != nil {
		return empty, errors.New("Scrape Failed.")
	}
	// Transaction ID should match what we sent
	if transactionID != responseTransactionID {
		return empty, errors.New("Unexpected response transactionID.")
	}

	for i, _ := range results {
		// Get the seeder count from the response
		var seeders uint32
		err = binary.Read(response, binary.BigEndian, &seeders)
		if err != nil {
			return empty, errors.New("Scrape Failed.")
		}

		// Get the completed count from the response
		var completed uint32
		err = binary.Read(response, binary.BigEndian, &completed)
		if err != nil {
			return empty, errors.New("Scrape Failed.")
		}

		// Get the leecher count from the response
		var leechers uint32
		err = binary.Read(response, binary.BigEndian, &leechers)
		if err != nil {
			return empty, errors.New("Scrape Failed.")
		}

		results[i] = Result{Btih: btihs[i], Seeders: int(seeders), Leechers: int(leechers), Completed: int(completed)}
	}

	// Return seeds, leeches, and completed
	return results, nil
}
