package main

import (
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/cockroachdb/errors"
)

func main() {
	switch os.Args[1] {
	case "run":
		if err := run(os.Args[2:]); err != nil {
			log.Fatalf("%+v", err)
		}

	default:
		log.Fatalf("unknown command: %s", os.Args[1])
	}
}

func run(command []string) error {
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS,
	}

	if err := cmd.Run(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
