package main

import (
	"log"
	"os"

	"github.com/k1LoW/errors"
)

func copyReadMe() error {
	log.Println("copying README.md from docs/en/index.md")

	src, err := os.ReadFile("docs/en/index.md")
	if err != nil {
		return errors.WithStack(err)
	}

	const permission = 0o644
	if err = os.WriteFile("README.md", src, permission); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
