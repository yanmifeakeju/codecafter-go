package main

import (
	"strings"
)

type command struct {
	name string
	args []string
}

func parse(line string) (command, bool, error) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return command{}, false, nil
	}

	return command{
		name: fields[0],
		args: fields[1:],
	}, true, nil
}
