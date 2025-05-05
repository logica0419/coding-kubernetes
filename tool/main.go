package main

import (
	"log"
)

func main() {
	if err := copySampleCodeFiles(); err != nil {
		log.Panic(err)
	}

	if err := generateSampleDiffFiles(); err != nil {
		log.Panic(err)
	}

	if err := copyReadMe(); err != nil {
		log.Panic(err)
	}
}
