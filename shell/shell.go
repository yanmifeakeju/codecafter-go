package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

type shell struct {
	term     *term.Terminal
	fd       int
	oldState *term.State
}

func newShell(c config) (*shell, error) {
	c = normalizeConfig(c)

	s := &shell{
		term: term.NewTerminal(&readWriter{c.r, c.w}, c.prompt),
	}

	if f, ok := c.r.(*os.File); ok {
		fd := int(f.Fd())
		if term.IsTerminal(fd) {
			oldState, err := term.MakeRaw(fd)
			if err != nil {
				return nil, fmt.Errorf("make raw: %w", err)
			}
			s.fd = fd
			s.oldState = oldState
		}
	}

	return s, nil
}

func run(c config) error {
	s, err := newShell(c)
	if err != nil {
		return err
	}

	if s.oldState != nil {
		defer term.Restore(s.fd, s.oldState)
	}

	for {
		line, err := s.term.ReadLine()
		if err == io.EOF {
			fmt.Fprintln(s.term)
			return nil
		}
		if err != nil {
			return err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fmt.Fprintf(s.term, "input: %s\n", line)
	}
}

// readWriter combines an io.Reader and io.Writer into an io.ReadWriter.
// term.NewTerminal needs a single ReadWriter, but we want to read from
// stdin and write to stdout
type readWriter struct {
	r io.Reader
	w io.Writer
}

func (rw *readWriter) Read(p []byte) (int, error)  { return rw.r.Read(p) }
func (rw *readWriter) Write(p []byte) (int, error) { return rw.w.Write(p) }

type config struct {
	prompt string
	w      io.Writer
	r      io.Reader
}

func normalizeConfig(c config) config {
	if c.r == nil {
		c.r = os.Stdin
	}

	if c.w == nil {
		c.w = os.Stdout
	}

	if c.prompt == "" {
		c.prompt = "> "
	}

	return c
}
