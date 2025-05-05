package main

import (
	"log"
)

func main() {
	if err := copySampleCodeFiles(); err != nil {
		log.Fatalf("%+v", err)
	}

	if err := generateSampleDiffFiles(); err != nil {
		log.Fatalf("%+v", err)
	}

	if err := copyReadMe(); err != nil {
		log.Fatalf("%+v", err)
	}
}
