package quartz

import (
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
	DotR Cmd = 'R'
	DotA Cmd = 'A'
)

type QuartzProtocol struct {
	transport io.ReadWriteCloser // Generally [net.Conn], but the protocol could parse things like files too
	log       *slog.Logger

	delimiter string
	maxLength int

	readErrors chan error
	lines      chan []byte
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
	}
}

func (p *QuartzProtocol) Start() {
	go p.dispatch()
	go p.readLines()
	go p.commandQueue()
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

// readLines reads and parses the incoming data and passes either a valid line
// or an error to [dispatch]
func (p *QuartzProtocol) readLines() {}

func (p *QuartzProtocol) commandQueue() {}

func (p *QuartzProtocol) handleLine(line []byte) {}

func (p *QuartzProtocol) handleError(err error) {}
