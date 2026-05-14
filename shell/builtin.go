package main

import (
	"fmt"
	"os/exec"
	"strings"
)

type builtinFunc func(s *shell, args []string, std stream) error

func defaultBuiltins() map[string]builtinFunc {
	return map[string]builtinFunc{
		"echo": runEcho,
		"type": runType,
		"exit": runExit,
	}
}

func runExit(s *shell, args []string, str stream) error {
	return errExit
}

func runEcho(s *shell, args []string, str stream) error {
	fmt.Fprintln(str.out, strings.Join(args, " "))
	return nil
}

func runType(s *shell, args []string, str stream) error {
	if len(args) == 0 {
		return nil
	}

	name := args[0]

	if _, ok := s.builtins[name]; ok {
		fmt.Fprintf(str.out, "%s is a shell builtin\n", name)
		return nil
	}

	if path, err := exec.LookPath(name); err == nil {
		fmt.Fprintf(str.out, "%s is %s\n", name, path)
		return nil
	}

	fmt.Fprintf(str.out, "%s: not found\n", name)
	return nil
}
