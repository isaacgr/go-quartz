package goquartz

import (
	"net"
	"testing"
	"time"

	"github.com/isaacgr/loggir"
)

var logger = loggir.GetLogger("test", false)

func TestSingleLineInvalid(t *testing.T) {
	server, client := net.Pipe()

	sender := NewProtocol(client, logger, nil)
	receiver := NewProtocol(server, logger, nil)

	sender.Start()
	receiver.Start()

	go func() {
		time.Sleep(time.Duration(1 * time.Second))
		sender.Stop()
		receiver.Stop()
	}()

	var resp *DotEResp
	sender.AddCommandHandler(DotE, func(p *QuartzProtocol, cmd any) {
		r, ok := cmd.(*DotEResp)
		if !ok {
			t.Errorf(
				"Incorrect response received. got=%v, expected=.E",
				cmd,
			)
		}
		resp = r
	})

	client.Write([]byte("foo\r"))

	sender.WaitUntilClosed()
	receiver.WaitUntilClosed()

	if resp == nil {
		t.Errorf(
			"No response received. expected=.E",
		)
	}
}

func TestTwoLinesInvalid(t *testing.T) {
	server, client := net.Pipe()

	sender := NewProtocol(client, logger, nil)
	receiver := NewProtocol(server, logger, nil)

	sender.Start()
	receiver.Start()

	go func() {
		time.Sleep(time.Duration(1 * time.Second))
		sender.Stop()
		receiver.Stop()
	}()

	var resps []*DotEResp
	sender.AddCommandHandler(DotE, func(p *QuartzProtocol, cmd any) {
		r, ok := cmd.(*DotEResp)
		if !ok {
			t.Errorf(
				"Incorrect response received. got=%v, expected=.E",
				cmd,
			)
		}
		resps = append(resps, r)
	})

	client.Write([]byte("foo\rbarbaz\r"))

	sender.WaitUntilClosed()
	receiver.WaitUntilClosed()


	if len(resps) != 2 {
		t.Errorf(
			"Not all responses received. got=%d, expected=2",
			len(resps),
		)
	}
}
