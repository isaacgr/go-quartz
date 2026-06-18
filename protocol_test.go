package quartz

import (
	"net"
	"testing"
	"time"
)

func TestSingleLine(t *testing.T) {
	server, client := net.Pipe()

	receiver := NewProtocol(server, nil, nil)
	receiver.Start()

	client.Write([]byte("foo\r"))

	go func() {
		time.Sleep(time.Duration(1 * time.Second))
		receiver.Stop()
	}()
	receiver.WaitUntilClosed()
}

func TestTwoLines(t *testing.T) {
	server, client := net.Pipe()

	receiver := NewProtocol(server, nil, nil)
	receiver.Start()

	client.Write([]byte("1234\r56789\r"))

	go func() {
		time.Sleep(time.Duration(1 * time.Second))
		receiver.Stop()
	}()
	receiver.WaitUntilClosed()
}
