package main

import (
	"bufio"
	"fmt"
	"os"
	"sync"

	"github.com/reservoird/icd"
)

type stdin struct {
	run bool
}

// NewIngester is what reservoird to create and start stdin
func NewIngester() (icd.Ingester, error) {
	return new(stdin), nil
}

// Config configures ingester
func (o *stdin) Config(cfg string) error {
	return nil
}

// Ingest reads data from stdin and writes it to a channel
func (o *stdin) Ingest(queue icd.Queue, done <-chan struct{}, wg *sync.WaitGroup) error {
	defer wg.Done()
	fmt.Printf("std.ingest: into\n")

	reader := bufio.NewReader(os.Stdin)
	o.run = true
	for o.run == true {
		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		if line != "" {
			fmt.Printf("std.ingest: before push\n")
			err = queue.Push([]byte(line))
			fmt.Printf("std.ingest: after push\n")
			if err != nil {
				return err
			}
		}

		select {
		case <-done:
			o.run = false
		default:
		}
	}
	fmt.Printf("std.ingest: outof\n")
	return nil
}
