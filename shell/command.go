package main

import (
	"fmt"
	"io"
)

type command struct {
	name string
	args []string
}

type execContext struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func (c command) run(_ *shell, ctx execContext) error {
	switch c.name {
	case "exit":
		return errExit
	default:
		_, err := fmt.Fprintf(ctx.stdout, "command: %s args: %v\n", c.name, c.args)
		return err
	}
}
