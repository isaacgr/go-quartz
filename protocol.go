package quartz

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
)

const (
	defaultDelimiter     = "\r"
	defaultMaxLength     = 1024 * 1024
	defaultMaxBufferSize = 8
)

type Cmd byte

const (
	DotS Cmd = 'S'
)

type CommandHandler func(p *QuartzProtocol, d Decoder[any])

type QuartzProtocol struct {
	transport io.ReadWriteCloser // Generally [net.Conn], but the protocol could parse things like files too so I thought this might be interesting
	log       *slog.Logger

	delimiter string
	maxLength int

	readErrors chan error
	lines      chan []byte
	writes     chan []byte

	closeSig        chan struct{}
	commandHandlers map[string]CommandHandler
}

type PcolOpts struct {
	delimiter *string
	maxLength *int
}

func NewProtocol(
	transport io.ReadWriteCloser,
	log *slog.Logger,
	opts *PcolOpts,
) *QuartzProtocol {
	delimiter := defaultDelimiter
	maxLength := defaultMaxLength

	if opts != nil {
		if opts.delimiter != nil {
			delimiter = *opts.delimiter
		}
		if opts.maxLength != nil {
			maxLength = *opts.maxLength
		}
	}
	return &QuartzProtocol{
		transport: transport,
		log:       log,
		delimiter: delimiter,
		maxLength: maxLength,

		lines:  make(chan []byte, 1),
		writes: make(chan []byte, 1),

		commandHandlers: make(map[string]CommandHandler),

		closeSig: make(chan struct{}),
	}
}

func (p *QuartzProtocol) Start() {
	go p.dispatch()
	go p.readLines()
}

func (p *QuartzProtocol) Stop() {
	close(p.closeSig)
}

func (p *QuartzProtocol) WaitUntilClosed() {
	<-p.closeSig
}

func (p *QuartzProtocol) AddCommandHandler(cmd string, handler CommandHandler) {
	if _, ok := p.commandHandlers[cmd]; !ok {
		p.log.Error("Handler already registerd for command", "Cmd", cmd)
	} else {
		p.commandHandlers[cmd] = handler
	}
}

func (p *QuartzProtocol) dispatch() {
	for {
		select {
		case line, ok := <-p.lines:
			if !ok {
				// TODO: close some channel that will close the protocol
				return
			}
			p.handleLine(line)
		case err, ok := <-p.readErrors:
			if !ok {
				// TODO: close some channel that will close the protocol
				return
			}
			p.handleShutdown(err.Error())
		}
	}
}

// readLines parses the incoming data from the reader for a delimiter and
// passes a complete line onto a channel for consumption
//
// When readLines encounters an error, its sent to the [readErrors] channel.
// Otherwise, the byte array is passed to the [lines] channel for command handling.
//
// EOF is not considered an error as this is meant to read a contiguous stream
// of data.
//
// A buffer exceeding p.maxLength will be dropped, and an error returned to the
// client
func (p *QuartzProtocol) readLines() {
	buf := make([]byte, defaultMaxBufferSize)
	readPos := 0
	lineBuffer := make([]byte, defaultMaxBufferSize)

	// TODO: Should we do something else with an EOF? Maybe read errors just shutdown the pcol
	for {
		readBytes, err := p.transport.Read(buf)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				p.log.Error("Error reading data", "Error", err)
				p.readErrors <- err
			} else {
				p.log.Warn("Transport ended by peer")

			}
			continue
		}

		// we have data
		if readBytes > 0 {
			// TODO: This is kind of dirty
			// TODO: Need to make sure we dont exceed the max length
			// TODO: Need to make sure we dont exceed the max buffer size
			if len(lineBuffer) == cap(lineBuffer) {
				// grow the line buffer
				newBuf := make([]byte, readPos+len(lineBuffer))
				copy(newBuf, lineBuffer)
				lineBuffer = newBuf
			}
			copy(lineBuffer[readPos:], buf[:readBytes])
			// our new reader position should increment how much data we sucessfully read
			readPos += readBytes

			// send all complete lines
			linePos := 0
			for {
				delimIdx := bytes.Index(
					lineBuffer[linePos:readPos],
					[]byte(p.delimiter),
				)
				// no complete lines
				if delimIdx < 0 {
					break
				}

				linePos += delimIdx
				line := lineBuffer[:linePos]
				linePos += 1

				p.lines <- line
			}
		}
	}
}

// handleLine parses a complete line and passes the command onto the respective
// handler
func (p *QuartzProtocol) handleLine(line []byte) {
	if line[0] != '.' || !(len(line) > 1) {
		p.handleUnknownCmd(line)
		return
	}

	if len(line) == 0 {
		p.handleNullCmd(line)
		return
	}

	p.handleCommand(line)
}

func (p *QuartzProtocol) handleCommand(line []byte) {
	cmd := line[1]
	// Undertale mode
	switch cmd {
	case 'S':
		d := DecodeDotS(line)
	case 'M':
	case 'B':
		peek := line[2]
		switch peek {
		case 'L':
		case 'U':
		case 'I':
		case 'A':
			// Response for .B request
		default:
			p.handleUnknownCmd(line)
		}
	case 'I':
	case 'L':
	case 'R':
		peek := line[2]
		switch peek {
		case 'D', 'E':
		case 'S', 'T':
		case 'L', 'M':
		default:
			p.handleUnknownCmd(line)
		}
	case 'W':
	case 'Q':
		peek := line[2]
		switch peek {
		case 'C':
		case 'S':
		case 'L':
		case 'R':
		default:
			p.handleUnknownCmd(line)
		}
	case 'A', 'U':
	case '#':
	case '&':
	default:
		p.handleUnknownCmd(line)
	}
}

func (p *QuartzProtocol) handleNullCmd(line []byte)    {}
func (p *QuartzProtocol) handleUnknownCmd(line []byte) {}

// handleShutdown closes the transport and stops the protocol
func (p *QuartzProtocol) handleShutdown(reason string) {
	p.log.Error("Shutting down", "Reason", reason)
	p.transport.Close()
	// TODO: this should trigger a handler for the transport conn closing
	p.Stop()
}
