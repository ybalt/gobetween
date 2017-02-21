/**
 * sni.go - sni sniffer implementation
 * @author Illarion Kovalchuk <illarion.kovalchuk@gmail.com>
 *
 * Package sni provides transparent access to hostname provided by ClientHello
 * message during TLS handshake.
 */

package sni

import (
	"bytes"
	"io"
	"net"
	"sync"
	"time"
)

const MAX_HEADER_SIZE = 16385

var pool = sync.Pool{
	New: func() interface{} {
		return make([]byte, MAX_HEADER_SIZE)
	},
}

// delegatedConn delegates all calls to net.Conn, but Read to reader
type Conn struct {
	hostname string
	reader   io.Reader
	net.Conn //delegate
}

func (c Conn) Hostname() string {
	return c.hostname
}

func (c Conn) Read(b []byte) (n int, err error) {
	return c.reader.Read(b)
}

// Sniff sniffs hostname from ClientHello message (if any),
// returns sni.Conn, filling it's Hostname field
func Sniff(conn net.Conn, readTimeout time.Duration) (net.Conn, error) {
	conn.SetReadDeadline(time.Now().Add(readTimeout))

	buf := pool.Get().([]byte)
	defer pool.Put(buf)

	i, err := conn.Read(buf)

	if err != nil {

		if nerr, ok := err.(net.Error); ok {
			//in case of timed out read from client - do not try to extract sni
			if nerr.Timeout() {
				return conn, nil
			}

			return nil, err
		}
	}

	if err != nil {
		return nil, err
	}

	hostname := extractHostname(buf[0:i])

	data := make([]byte, i)
	copy(data, buf) // Since we reuse buf between invocations, we have to make copy of data
	mreader := io.MultiReader(bytes.NewBuffer(data), conn)

	// Wrap connection so that it will Read from buffer first and remaining data
	// from initial conn
	return Conn{hostname, mreader, conn}, nil
}
