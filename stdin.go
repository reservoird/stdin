package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/reservoird/icd"
)

type stdin struct {
	run       bool
	Name      string
	Timestamp bool
}

// NewIngester is what reservoird to create and start stdin
func NewIngester() (icd.Ingester, error) {
	return new(stdin), nil
}

// Config configures ingester
func (o *stdin) Config(cfg string) error {
	d, err := ioutil.ReadFile(cfg)
	if err != nil {
		return err
	}
	s := stdin{}
	err = json.Unmarshal(d, &s)
	if err != nil {
		return err
	}
	o.Name = s.Name
	o.Timestamp = s.Timestamp
	return nil
}

// Ingest reads data from stdin and writes it to a channel
func (o *stdin) Ingest(queue icd.Queue, done <-chan struct{}, wg *sync.WaitGroup) error {
	defer wg.Done()

	reader := bufio.NewReader(os.Stdin)
	o.run = true
	for o.run == true {
		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		if line != "" {
			if o.Timestamp == true {
				line = fmt.Sprintf("%s %s: ", o.Name, time.Now().Format(time.RFC3339)) + line
			}
			err = queue.Push([]byte(line))
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
	return nil
}
