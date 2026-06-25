package goquartz

import (
	"bytes"
	"container/list"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
)

const (
	defaultDelimiter     = "\r"
	defaultMaxLength     = 1024 * 1024
	defaultMaxBufferSize = 8
)

type Cmd byte

const (
	DotE Cmd = 'E'
	DotA Cmd = 'A'
	DotS Cmd = 'S'
)

type CommandHandler func(p *QuartzProtocol, cmd any)
type ConnectionHandler func(p *QuartzProtocol, conn net.Conn)

type QuartzProtocol struct {
	transport net.Conn
	log       *slog.Logger

	delimiter string
	maxLength int

	readErrors chan error
	lines      chan []byte
	writes     chan []byte

	closeSig              chan struct{}
	commandHandlers       map[Cmd]CommandHandler
	connectionMadeHandler ConnectionHandler
	connectionLostHandler ConnectionHandler

	commandQueue list.List
	current      string // The current command awaiting processing
}

type PcolOpts struct {
	delimiter *string
	maxLength *int
}

type ErrResp struct {
	// TODO: We can maybe add a message to an error? Its unclear
	Msg string
}

func (e *ErrResp) Error() string {
	return fmt.Sprintf(
		".E%s",
		defaultDelimiter,
	)
}

func NewProtocol(
	transport net.Conn,
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

		lines:      make(chan []byte, 1),
		writes:     make(chan []byte, 1),
		readErrors: make(chan error, 1),

		commandHandlers: make(map[Cmd]CommandHandler),

		closeSig: make(chan struct{}),

		commandQueue: *list.New(),
	}
}

func (p *QuartzProtocol) Start() {
	go p.dispatch()
	go p.readLines()
	go p.writeLines()
}

func (p *QuartzProtocol) Stop() {
	close(p.closeSig)
}

func (p *QuartzProtocol) WaitUntilClosed() {
	<-p.closeSig
}

func (p *QuartzProtocol) AddConnectionMadeHandler(handler ConnectionHandler) {
	if p.connectionMadeHandler != nil {
		p.log.Error("Connection made handler already registered")
	} else {
		p.connectionMadeHandler = handler
	}
}

func (p *QuartzProtocol) AddConnectionLostHandler(handler ConnectionHandler) {
	if p.connectionMadeHandler != nil {
		p.log.Error("Connection lost handler already registered")
	} else {
		p.connectionMadeHandler = handler
	}
}

func (p *QuartzProtocol) AddCommandHandler(cmd Cmd, handler CommandHandler) {
	if _, ok := p.commandHandlers[cmd]; ok {
		p.log.Error("Handler already registered for command", "Cmd", cmd)
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
		case <-p.closeSig:
			return
		}
	}
}

func (p *QuartzProtocol) writeLines() {
	for line := range p.writes {
		_, err := p.transport.Write(line)
		if err != nil {
			p.log.Error(
				"Unable to write to peer",
				"Error",
				err,
			)
			// TODO: What should happen? Close the connection?
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
	buf := make([]byte, defaultMaxBufferSize)        // buffer for incomming data
	lineBuffer := make([]byte, defaultMaxBufferSize) // buffer to pop off lines

	readPos := 0 // how much valid data has been read

	// TODO: Should we do something else with an EOF? Maybe read errors just shutdown the pcol
	for {
		readBytes, err := p.transport.Read(buf)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				p.log.Error("Error reading data", "Error", err)
				p.readErrors <- err
			} else {
				p.log.Warn("Connection closed by peer")
			}
			continue
		}

		// we have data
		if readBytes > 0 {
			// TODO: Need to make sure we dont exceed the max length
			// TODO: Need to make sure we dont exceed the max buffer size

			if len(lineBuffer) == cap(lineBuffer) {
				// grow the line buffer
				// https://go.dev/wiki/SliceTricks#extend-capacity
				lineBuffer = append(
					make([]byte, 0, len(lineBuffer)+readBytes),
					lineBuffer...,
				)
			}

			copy(lineBuffer[readPos:], buf[:readBytes])
			// our new reader position should increment how much data we sucessfully read
			readPos += readBytes

			linePos := 0 // position of the current line search
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
				currLine := make([]byte, 0, len(lineBuffer))
				currLine = append(currLine, lineBuffer[:linePos]...)
				linePos += len(p.delimiter)

				p.lines <- currLine
			}
			// shift remaining partial line to beginning of buffer
			if linePos > 0 {
				remainingBytes := readPos - linePos
				if remainingBytes > 0 {
					copy(lineBuffer, lineBuffer[linePos:readPos])
				}
				readPos = remainingBytes
			}
		}
	}
}

// handleLine parses a full line and queues the command
func (p *QuartzProtocol) handleLine(line []byte) {
	if !(len(line) > 1) || line[0] != '.' {
		p.handleUnknownCmd(line)
		return
	}

	if len(line) == 0 {
		p.handleNullCmd(line)
		return
	}

	msg := line[1]
	// Undertale mode
	switch msg {
	case 'S':
		handler, ok := p.commandHandlers[DotS]
		if !ok {
			p.log.Error(
				"No matching handler found for command",
				"Cmd",
				string(DotS),
			)
			return
		}
		cmd, err := DecodeDotS(line[2:])
		if err != nil {
			p.log.Error(
				".S command received, but invalid arguments",
				"Cmd",
				string(line),
				"Error",
				err.Error(),
			)
			dotE := ErrResp{}
			err = p.sendLine([]byte(dotE.Error()))
			if err != nil {
				p.log.Error("Unable to send response")
				p.handleShutdown(err.Error())
				return
			}
			return
		}
		handler(p, cmd)
	case 'M':
	case 'B':
		switch p.cmdType(line) {
		case 'L':
		case 'U':
		case 'I':
		case 'A':
			// Response for .B request
			p.handleResponse(line)
		default:
			p.handleUnknownCmd(line)
		}
	case 'F':
	case 'I':
	case 'C':
	case 'L':
	case 'R':
		switch p.cmdType(line) {
		case 'D', 'E':
		case 'S', 'T':
		case 'L', 'M':
		case 'A':
			// Response for .R request, or a .W request
			p.handleResponse(line)
		default:
			p.handleUnknownCmd(line)
		}
	case 'W':
		switch p.cmdType(line) {
		case 'D', 'E':
		case 'S', 'T':
		case 'L', 'M':
		default:
			p.handleUnknownCmd(line)

		}
	case '?':
		// Embedded control system only
	case '!':
		// Embedded control system only
	case 'Q':
		switch p.cmdType(line) {
		case 'C':
		case 'R':
		case 'S':
		case 'F':
		case 'D':
		case 'L':
		default:
			p.handleUnknownCmd(line)
		}
	case 'A':
		p.handleResponse(line)
	case 'E':
		p.handleErrorResponse(line)
	case 'U':
		// Can be a response to a .S command, or unsolicited
		p.handleUpdate(line)
	case '#':
	case '&':
	default:
		p.handleUnknownCmd(line)
	}
}

// cmdType is a helper method to get the 'type' of the dot command, for
// example the type of .BL would be 'L'
func (p *QuartzProtocol) cmdType(line []byte) byte {
	if len(line) > 2 {
		return line[2]
	}
	return 0
}

// sendLine is the low level function to send a single line to a remote peer
// the line is queued in the transports write queue
func (p *QuartzProtocol) sendLine(line []byte) error {
	select {
	case p.writes <- line:
		return nil
	case <-p.closeSig:
		return errors.New("protocol closed: write cancelled")
	default:
		return errors.New("protocol shut down: write failed")
	}
}

func (p *QuartzProtocol) handleUpdate(line []byte)   {}
func (p *QuartzProtocol) handleResponse(line []byte) {}

func (p *QuartzProtocol) handleErrorResponse(line []byte) {
	handler, ok := p.commandHandlers[DotE]
	if !ok {
		p.log.Error(
			"No matching handler found for command",
			"Cmd",
			string(DotE),
		)
		return
	}
	cmd, err := DecodeDotE(line[1:])
	if err != nil {
		p.log.Error(
			".E received, but invalid structure",
			"Line",
			string(line),
		)
		return
	}
	// TODO: This blocks the go routine
	handler(p, cmd)
}

func (p *QuartzProtocol) handleNullCmd(line []byte) {}
func (p *QuartzProtocol) handleUnknownCmd(line []byte) {
	p.log.Warn(
		"Received unknown command",
		"Line",
		string(line),
	)
	dotE := &ErrResp{}
	err := p.sendLine([]byte(dotE.Error()))
	if err != nil {
		p.log.Error("Unable to send .E response", "Error", err)
	}
}

// handleShutdown closes the transport and stops the protocol
func (p *QuartzProtocol) handleShutdown(reason string) {
	p.log.Error("Shutting down", "Reason", reason)
	// TODO: this should trigger a handler for the transport conn closing
	p.Stop()
}
