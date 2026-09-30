package goquartz

import (
	"bytes"
	"log/slog"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/isaacgr/loggir"
)

var logger = loggir.GetLogger("test", false)

func TestReadLines(t *testing.T) {
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

	client.Write([]byte(".SV1,2\r"))
}

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

func TestNullCmdLogsWarning(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	server, client := net.Pipe()
	p := NewProtocol(server, logger, nil)
	p.Start()
	defer p.Stop()

	client.Write([]byte("\r"))
	time.Sleep(time.Duration(1 * time.Second))

	if !strings.Contains(buf.String(), "empty command") {
		t.Errorf("expected warning log, got: %s", buf.String())
	}
}

func TestConnectionMadeHandler(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	server, _ := net.Pipe()
	p := NewProtocol(server, logger, nil)

	done := make(chan struct{})
	connectionMadeCalled := false
	p.AddConnectionMadeHandler(func(p *QuartzProtocol, err error) {
		if err != nil {
			t.Errorf("connection made got error. err = %v", err)
		}
		connectionMadeCalled = true
		close(done)
	})

	p.Start()
	defer p.Stop()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Errorf("connection made handler timeout")
	}

	if !connectionMadeCalled {
		t.Errorf("connection made handler not called")
	}
}

func TestConnectionLostHandler(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	server, _ := net.Pipe()
	p := NewProtocol(server, logger, nil)

	done := make(chan struct{})
	connectionLostCalled := false
	p.AddConnectionLostHandler(func(p *QuartzProtocol, err error) {
		if err == nil {
			t.Errorf("connection lost got no error")
		}
		connectionLostCalled = true
		close(done)
	})

	p.Start()
	defer p.Stop()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Errorf("connection lost handler timeout")
	}

	if !connectionLostCalled {
		t.Errorf("connection lost handler not called")
	}

}
