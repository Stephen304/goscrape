package goscrape

import (
	"bytes"
	"encoding/binary"
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
	defer conn.Close()

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
