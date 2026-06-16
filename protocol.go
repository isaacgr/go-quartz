package quartz

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
)

const (
	defaultDelimiter = "\r"
	defaultMaxLength = 1024 * 1024
)

type Cmd byte

const (
	DotE Cmd = 'E'
)

type QuartzProtocol struct {
	transport io.ReadWriteCloser // Generally [net.Conn], but the protocol could parse things like files too
	log       *slog.Logger

	delimiter string
	maxLength int

	readErrors   chan error
	lines        chan []byte
	messageQueue []byte
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

		lines:        make(chan []byte, 1),
		messageQueue: make([]byte, 0),
	}
}

func (p *QuartzProtocol) Start() {
	go p.dispatch()
	go p.readLines()
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
			p.handleError(err)
		}
	}
}

// readLines reads the incoming data from the reader
//
// When readLines encounters an error, its sent to the [readErrors] channel.
// Otherwise, the byte array is passed to the [lines] channel for parsing
//
// EOF is not considered an error as this is meant to read a contiguous stream
// of data.
func (p *QuartzProtocol) readLines() {
	buf := make([]byte, 8)

	// TODO: Maybe only pass a complete line from readLines? Which would require
	// building the buffer here
	//
	// TODO: Should we do something else with an EOF?
	for {
		_, err := p.transport.Read(buf)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				p.readErrors <- err
			}
			continue
		}
		p.lines <- buf
	}
}

// handleLine parses read lines from the [lines] channel. If a line does not
// include a delimiter, then the buffer is built until one is encountered.
// Once a complete line is found, handleLine checks if the line is either
// an error, ".E", a NULL (only a delimiter), or otherwise a valid command.
// handleLine will then pass the line onto [handleError], [handleNullCmd] or
// [handleCmd] respectively.
//
// A buffer exceeding p.maxLength will be dropped, and an error returned to the
// client
func (p *QuartzProtocol) handleLine(line []byte) {
	delimIdx := bytes.Index(line, []byte(p.delimiter))
	if delimIdx == -1 {
		p.messageQueue = append(p.messageQueue, line...)
		return
	}

	line = line[:delimIdx]

	if len(line) == 0 {
		p.handleNullCmd(line)
	}

}

func (p *QuartzProtocol) handleError(err error)     {}
func (p *QuartzProtocol) handleNullCmd(line []byte) {}
