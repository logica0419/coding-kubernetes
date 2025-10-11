package main

import (
	"log"
	"os"
	"os/exec"

	"github.com/k1LoW/errors"
	"golang.org/x/sys/unix"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		log.Fatalln(errors.StackTraces(err))
	}
}

func run(command []string) error {
	path, err := exec.LookPath(command[0])
	if err != nil {
		return errors.WithStack(err)
	}

	if err := unix.Exec(path, command, os.Environ()); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
