package main

import (
	"log"

	"github.com/k1LoW/errors"
)

func main() {
	if err := copySampleCodeFiles(); err != nil {
		log.Fatalln(errors.StackTraces(err))
	}

	if err := generateSampleDiffFiles(); err != nil {
		log.Fatalln(errors.StackTraces(err))
	}

	if err := copyReadMe(); err != nil {
		log.Fatalln(errors.StackTraces(err))
	}
}
