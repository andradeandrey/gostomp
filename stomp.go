// Copyright 2010 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*

STOMP client for Go
See: http://stomp.codehaus.org/Protocol

Usage:

	// create a net.Conn, and pass that into Connect
	nc := net.Dial("tcp", "", "127.0.0.1:65000")
	c := stomp.Connect(nc)

	// subscribe to a queue
	c.Subscribe("/queue/foo")

	// read messages from the channel In
	for msg := range c.In {
		// handle msg (a struct of type Msg)
	}

	// to send a message
	c.Send("/topic/bar", "I'm writing Go code!")

	// disconnect
	c.Disconnect()

TODO:
 - support ERROR
 - support RECEIPT
 - support client ACK
 - support transactions (BEGIN/COMMIT/ABORT)
 - support UNSUBSCRIBE
 - more tests

*/

package stomp

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

type Msg struct {
	Header map[string]string
	Data   []byte
}

type Conn struct {
	In        <-chan Msg
	in        chan Msg
	out       chan frame
	connected bool
	session   string
}

type header map[string]string

type frame struct {
	command string
	header  header
	body    string
}

func (f *frame) writeTo(w io.Writer) os.Error {
	if _, err := fmt.Fprintf(w, "%s\n", f.command); err != nil {
		return err
	}
	for name, value := range f.header {
		_, err := fmt.Fprintf(w, "%s: %s\n", name, value)
		if err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(w, "\n%s", f.body); err != nil {
		return err
	}
	return nil
}

func frameFromString(s string) (*frame, os.Error) {
	// TODO(adg): handle \r at line endings?

	f := new(frame)

	cmd := strings.Split(s, "\n", 2)
	if len(cmd) < 2 {
		return nil, os.NewError("Malformed frame")
	}
	f.command = cmd[0]

	f.header = make(header)
	headers := strings.Split(cmd[1], "\n\n", 2)
	for _, line := range strings.Split(headers[0], "\n", 0) {
		parts := strings.Split(line, ":", 2)
		if len(parts) < 2 {
			return nil, os.NewError("Malformed frame")
		}
		f.header[parts[0]] = strings.TrimSpace(parts[1])
	}

	if len(headers) == 2 {
		f.body = headers[1]
	}

	return f, nil
}

func Connect(nc net.Conn, h map[string]string) *Conn {
	c := &Conn{in: make(chan Msg), out: make(chan frame)}
	c.In = c.in

	go c.reader(nc)
	go c.writer(nc)

	c.out <- frame{"CONNECT", h, ""}

	return c
}

func (c *Conn) reader(nc net.Conn) {
	br := bufio.NewReader(nc)
	for {
		// TODO(adg) make frameFromReader and skip the
		// whole read-until-zero thing
		b, err := br.ReadBytes(0)
		if err != nil {
			// TODO(adg) handle error
			break
		}

		f, err := frameFromString(string(b))
		if err != nil {
			// TODO(adg) handle malformed frame (ignore?)
			continue
		}

		if f.command == "CONNECTED" {
			c.connected = true
			if session, ok := f.header["session"]; ok {
				c.session = session
			}
		}

		// TODO(adg) we only handle 'message' so far
		if f.command != "MESSAGE" {
			continue
		}

		c.in <- Msg{f.header, []byte(f.body)}
	}
}

func (c *Conn) writer(nc net.Conn) {
	bw := bufio.NewWriter(nc)
	for f := range c.out {
		if err := f.writeTo(bw); err != nil {
			// handle error
		}
		if err := bw.WriteByte('\x00'); err != nil {
			// handle error
		}
		if err := bw.Flush(); err != nil {
			// handle error
		}
		if f.command == "DISCONNECT" {
			break
		}
	}
	nc.Close()
	c.connected = false
}

func (c *Conn) Send(dest, body string) {
	c.out <- frame{"SEND", header{"destination": dest}, body}
}

func (c *Conn) Subscribe(dest string, clientAck bool) {
	ack := "auto"
	if clientAck {
		ack = "client"
	}
	c.out <- frame{
		"SUBSCRIBE",
		header{"destination": dest, "ack": ack},
		"",
	}
}

func (c *Conn) Disconnect() { c.out <- frame{"DISCONNECT", header{}, ""} }
