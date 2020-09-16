package chesspb

import (
	"bufio"
	"errors"
	"fmt"
	"google.golang.org/protobuf/proto"
	"net"
	"time"
)

const (
	WRITER_MAXWAIT = 30 * time.Second //max time for packet to send to client
)

type Client struct {
	W chan []byte
}

//SendBytes is used to send data over the client's writer. Useful for loops which would
//otherwise be rebuilding the exact same message. It will not block, and returns an error if it
//would have blocked or if c is nil.
func (c *Client) SendBytes(data []byte) (err error) {
	if c == nil {
		err = errors.New("Nil client")
		return
	}
	select {
	case c.W <- data:
	default:
		err = errors.New("Write buffer full")
	}
	return
}

// Send is the standard way to send data over the client's writer.
// It will not block, and returns an error if it would have blocked or if c is nil.
func (c *Client) Send(msg proto.Message) (err error) {
	if c == nil {
		err = errors.New("Nil client")
		return
	}
	var data []byte
	data, err = BuildMessage(msg)
	if err != nil {
		return
	}
	select {
	case c.W <- data:
	default:
		err = errors.New("Write buffer full")
	}
	return
}

// Writer is a method that sits and waits on channel c.writer for data to send over conn.
func (c *Client) Writer(conn net.Conn) {
	writer := bufio.NewWriter(conn)
	defer func() {
		conn.Close()
	}()

	for {
		data := <-c.W
		if data == nil { //closing connection
			return
		}
		conn.SetWriteDeadline(time.Now().Add(WRITER_MAXWAIT))
		_, err := writer.Write(data)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = writer.Flush()
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
