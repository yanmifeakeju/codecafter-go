package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"golang.org/x/term"
)

var errExit = errors.New("exit")

type shell struct {
	term *term.Terminal
	str  stream

	fd         int
	isTerminal bool
	oldState   *term.State

	builtins map[string]builtinFunc
}

type stream struct {
	in  io.Reader
	out io.Writer
	err io.Writer
}

func (s *shell) enterRawMode() error {
	if !s.isTerminal {
		return nil
	}

	oldState, err := term.MakeRaw(s.fd)
	if err != nil {
		return fmt.Errorf("make raw: %w", err)
	}

	s.oldState = oldState
	return nil
}

func (s *shell) exitRawMode() error {
	if s.oldState == nil {
		return nil
	}

	err := term.Restore(s.fd, s.oldState)
	s.oldState = nil
	return err
}

func (s *shell) execute(line string) error {
	cmd, ok, err := parse(line)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	return s.runCommand(cmd, s.str)
}

func (s *shell) runCommand(cmd command, std stream) error {
	if builtin, ok := s.builtins[cmd.name]; ok {
		return builtin(s, cmd.args, std)
	}
	return s.runExternal(cmd, std)
}

func (s *shell) runExternal(cmd command, std stream) error {
	path, err := exec.LookPath(cmd.name)
	if err != nil {
		fmt.Fprintf(std.err, "shell: %s: command not found\n", cmd.name)
		return nil
	}

	proc := exec.Command(path, cmd.args...)
	proc.Stdin = std.in
	proc.Stdout = std.out
	proc.Stderr = std.err
	return proc.Run()
}

func newShell(c config) (*shell, error) {
	c = normalizeConfig(c)

	s := &shell{}
	s.term = term.NewTerminal(&readWriter{c.r, c.w}, c.prompt)
	s.str = stream{in: c.r, out: c.w, err: c.err}
	s.builtins = defaultBuiltins()

	if f, ok := c.r.(*os.File); ok {
		fd := int(f.Fd())
		if term.IsTerminal(fd) {
			s.fd = fd
			s.isTerminal = true
		}
	}

	return s, nil
}

func run(c config) error {
	s, err := newShell(c)
	if err != nil {
		return err
	}

	for {
		if err := s.enterRawMode(); err != nil {
			return err
		}

		line, err := s.term.ReadLine()

		if restoreErr := s.exitRawMode(); restoreErr != nil {
			return restoreErr
		}

		if err == io.EOF {
			fmt.Fprintln(s.str.out)
			return nil
		}

		if err != nil {
			return err
		}

		cmd, ok, err := parse(line)
		if err != nil {
			fmt.Fprintf(s.str.err, "%v\n", err)
			continue
		}

		if !ok {
			continue
		}

		if err := s.execute(line); err != nil {
			if errors.Is(err, errExit) {
				return nil
			}

			fmt.Fprintf(s.str.err, "%s: %v\n", cmd.name, err)
			continue
		}
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
	err    io.Writer
}

func normalizeConfig(c config) config {
	if c.r == nil {
		c.r = os.Stdin
	}

	if c.w == nil {
		c.w = os.Stdout
	}

	if c.err == nil {
		c.err = os.Stderr
	}

	if c.prompt == "" {
		c.prompt = "> "
	}

	return c
}
