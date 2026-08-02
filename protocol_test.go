package goquartz

import (
	"bytes"
	"log/slog"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/isaacgr/loggir"
)

var logger = loggir.GetLogger("test", false)

func TestReadLinesPartialMessage(t *testing.T) {
	server, client := net.Pipe()
	p := NewProtocol(server, logger, nil)
	p.Start()
	defer p.Stop()

	p.AddCommandHandler(DotS, func(p *QuartzProtocol, cmd any) {
		r, ok := cmd.(*XptMsg)
		if !ok {
			t.Errorf(
				"Incorrect response received. got=%v, expected=.S",
				cmd,
			)
		}
		if r.Dst != 1 || r.Src != 2 {
			t.Errorf(
				"incorrect route: got dst=%d src=%d, want dst=1, src=2",
				r.Dst,
				r.Src,
			)
		}

		if len(r.Levels) != 1 && r.Levels[0] != "V" {
			t.Errorf(
				"incrorrect levels received: got %v, want=V", r.Levels,
			)
		}

	})

	client.Write([]byte(".SV1"))
	time.Sleep(time.Duration(1 * time.Second))
	client.Write([]byte(",2\r"))
}

func TestUnknownCmdLogsWarning(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	server, client := net.Pipe()
	p := NewProtocol(server, logger, nil)
	p.Start()
	defer p.Stop()

	client.Write([]byte("foo\r"))
	time.Sleep(time.Duration(1 * time.Second))

	if !strings.Contains(buf.String(), "unknown command") {
		t.Errorf("expected warning log, got: %s", buf.String())
	}
}

func TestQuartzProtocolDotE(t *testing.T) {
	server, client := net.Pipe()

	sender := NewProtocol(client, logger, nil)
	receiver := NewProtocol(server, logger, nil)
	sender.Start()
	receiver.Start()

	var wg sync.WaitGroup
	var respMu sync.Mutex

	t.Cleanup(func() {
		sender.Stop()
		receiver.Stop()
	})

	var resp *DotEResp
	wg.Add(1)
	sender.AddCommandHandler(DotE, func(p *QuartzProtocol, cmd any) {
		r, ok := cmd.(*DotEResp)
		if !ok {
			t.Errorf(
				"Incorrect response received. got=%v, expected=.E",
				cmd,
			)
		}
		respMu.Lock()
		resp = r
		respMu.Unlock()
		wg.Done()
	})

	client.Write([]byte("foo\r"))

	wg.Wait()

	if resp == nil {
		t.Errorf(
			"No response received. expected=.E",
		)
	}
}

func TestQuartzProtocolMultipleDotE(t *testing.T) {
	server, client := net.Pipe()

	sender := NewProtocol(client, logger, nil)
	receiver := NewProtocol(server, logger, nil)

	sender.Start()
	receiver.Start()

	var wg sync.WaitGroup
	var respMu sync.Mutex

	t.Cleanup(func() {
		sender.Stop()
		receiver.Stop()
	})

	var resps []*DotEResp
	wg.Add(2)

	sender.AddCommandHandler(DotE, func(p *QuartzProtocol, cmd any) {
		r, ok := cmd.(*DotEResp)
		if !ok {
			t.Errorf(
				"Incorrect response received. got=%v, expected=.E",
				cmd,
			)
		}

		respMu.Lock()
		resps = append(resps, r)
		respMu.Unlock()
		wg.Done()
	})

	client.Write([]byte("foo\rbarbaz\r"))

	wg.Wait()

	if len(resps) != 2 {
		t.Errorf(
			"Not all responses received. got=%d, expected=2",
			len(resps),
		)
	}
}

func TestQuartzProtocolDotSValidRoute(t *testing.T) {
	server, client := net.Pipe()

	sender := NewProtocol(client, logger, nil)
	receiver := NewProtocol(server, logger, nil)

	sender.Start()
	receiver.Start()

	var wg sync.WaitGroup
	var respMu sync.Mutex

	t.Cleanup(func() {
		sender.Stop()
		receiver.Stop()
	})

	var resp *XptMsg
	wg.Add(1)

	receiver.AddCommandHandler(DotS, func(p *QuartzProtocol, cmd any) {
		r, ok := cmd.(*XptMsg)
		if !ok {
			t.Errorf(
				"Incorrect response received. got=%v, expected=.E",
				cmd,
			)
		}

		if r.Dst != 1 || r.Src != 1 {
			t.Errorf(
				"incorrect route: got dst=%d src=%d, want dst=1, src=1",
				r.Dst,
				r.Src,
			)
		}

		if len(r.Levels) != 1 && r.Levels[0] != "V" {
			t.Errorf(
				"incrorrect levels received: got %v, want=V", r.Levels,
			)
		}

		respMu.Lock()
		resp = r
		respMu.Unlock()
		wg.Done()
	})

	client.Write([]byte(".SV1,1\r"))

	wg.Wait()

	if resp == nil {
		t.Errorf(
			"No response received. expected=.E",
		)
	}
}

func TestQuartzProtocolDotSInvalidArgs(t *testing.T) {
	server, client := net.Pipe()

	sender := NewProtocol(client, logger, nil)
	receiver := NewProtocol(server, logger, nil)

	sender.Start()
	receiver.Start()

	var wg sync.WaitGroup

	t.Cleanup(func() {
		sender.Stop()
		receiver.Stop()
	})

	var respMu sync.Mutex
	var resp *DotEResp
	wg.Add(1)

	sender.AddCommandHandler(DotE, func(p *QuartzProtocol, cmd any) {
		r, ok := cmd.(*DotEResp)
		if !ok {
			t.Errorf(
				"Incorrect response received. got=%v, expected=.E",
				cmd,
			)
		}
		respMu.Lock()
		resp = r
		respMu.Unlock()
		wg.Done()
	})

	receiver.AddCommandHandler(DotS, func(p *QuartzProtocol, cmd any) {
		t.Error(".S handler should not have been called for invalid input")
	})

	client.Write([]byte(".S1,1\r"))

	wg.Wait()

	if resp == nil {
		t.Errorf(
			"No response received. expected=.E",
		)
	}
}
