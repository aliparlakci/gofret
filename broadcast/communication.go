package broadcast

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
)

type communication struct {
	listener net.Listener
	address  string
}

type communicator interface {
	Listen() (chan []byte, error)
	CloseConnection()
	Send(address string, message []byte) error
}

func new_communication(address string) communicator {
	return &communication{address: address}
}

func (c *communication) Listen() (chan []byte, error) {
	l, err := net.Listen("tcp", c.address)
	if err != nil {
		return nil, fmt.Errorf("cannot start listening on address %v: %v", c.address, err)
	}
	c.listener = l

	incoming_messages := make(chan []byte)
	handle_request := func(conn net.Conn) {
		buffer := bytes.NewBuffer([]byte{})
		tmp := make([]byte, 8) // using small tmo buffer for demonstrating
		for {
			n, err := conn.Read(tmp)
			if err == io.EOF {
				break
			} else if err != nil {
				fmt.Println("read error:", err)
			}
			buffer.Write(tmp[:n])
		}
		incoming_messages <- buffer.Bytes()
	}

	go func() {
		for {
			connection, err := l.Accept()
			if err != nil {
				log.Fatalf("something happened while trying to accept an incoming connection: %v", err)
			}
			go handle_request(connection)
		}
	}()

	return incoming_messages, nil
}

func (c *communication) CloseConnection() {
	c.listener.Close()
}

func (c *communication) Send(address string, message []byte) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("dial failed: %v", err)
	}
	defer conn.Close()
	_, err = conn.Write(message)
	if err != nil {
		return fmt.Errorf("write failed: %v", err)
	}

	return nil
}
