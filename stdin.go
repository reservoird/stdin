package main

import (
	"bufio"
	"os"
)

type stdin struct {
}

// Config configures ingester
func (o *stdin) Config(cfg string) error {
	return nil
}

// Ingest reads data from stdin and writes it to a channel
func (o *stdin) Ingest(channel chan<- []byte) error {
	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		channel <- []byte(line)
	}
}

// Ingester for stdin
var Ingester stdin
