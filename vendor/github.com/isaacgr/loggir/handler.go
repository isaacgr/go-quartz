package loggir

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"
)

type options struct {
	Level slog.Leveler
}

type moduleHandler struct {
	opts   options
	groups []string
	attrs  []slog.Attr
	mu     *sync.Mutex
	out    io.Writer
}

func newModuleHandler(out io.Writer, opts *options) *moduleHandler {
	h := &moduleHandler{
		out: out,
		mu:  &sync.Mutex{},
	}
	if opts != nil {
		h.opts = *opts
	}
	if h.opts.Level == nil {
		h.opts.Level = slog.LevelInfo
	}
	return h
}

func (h *moduleHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

func (h *moduleHandler) Handle(_ context.Context, r slog.Record) error {
	buf := make([]byte, 0, 1024)

	timestamp := r.Time.Format(time.RFC3339Nano)

	buf = append(
		buf,
		fmt.Sprintf(
			"%s %s:",
			timestamp,
			r.Level.String(),
		)...,
	)

	for n, attr := range h.attrs {
		buf = append(buf, fmt.Sprintf("%v", attr.Value)...)
		if len(h.attrs) > n+1 {
			buf = append(buf, "."...)
		}
	}

	buf = append(buf, fmt.Sprintf(": %s", r.Message)...)
	r.Attrs(
		func(a slog.Attr) bool {
			buf = append(buf, fmt.Sprintf(" %s [%v].", a.Key, a.Value)...)
			return true
		},
	)

	buf = append(buf, "\n"...)
	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.out.Write(buf)
	return err
}

func (h *moduleHandler) WithGroup(name string) slog.Handler {
	ch := *h
	ch.groups = append(ch.groups, name)
	return &ch
}

func (h *moduleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	ch := *h
	ch.attrs = append(ch.attrs, attrs...)
	return &ch
}
