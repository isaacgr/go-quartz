package quartz

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
	sender.Start()
	receiver.Start()

	client.Write([]byte("foo\r"))
	client.SetDeadline(<-time.After(time.Duration(1 * time.Second)))

	if resp == nil {
		t.Errorf(
			"No response received. expected=.E",
		)
	}

	sender.Stop()
	receiver.Stop()
	receiver.WaitUntilClosed()
}

func TestTwoLinesInvalid(t *testing.T) {
	server, client := net.Pipe()

	sender := NewProtocol(client, logger, nil)
	receiver := NewProtocol(server, logger, nil)
	sender.Start()
	receiver.Start()

	client.Write([]byte("foo\rbarbaz\r"))

	go func() {
		time.Sleep(time.Duration(1 * time.Second))
		receiver.Stop()
	}()
	receiver.WaitUntilClosed()
}
