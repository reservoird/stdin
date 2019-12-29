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
	stats     chan<- string
	run       bool
	Tag       string
	Timestamp bool
}

// New is what reservoird to create and start stdin
func New(cfg string, stats chan<- string) (icd.Ingester, error) {
	o := &stdin{
		stats:     stats,
		run:       true,
		Tag:       "stdin",
		Timestamp: false,
	}
	if cfg != "" {
		d, err := ioutil.ReadFile(cfg)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(d, o)
		if err != nil {
			return nil, err
		}
	}
	return o, nil
}

// Name return the name of the Ingester
func (o *stdin) Name() string {
	return o.Tag
}

// Ingest reads data from stdin and writes it to a channel
func (o *stdin) Ingest(queue icd.Queue, done <-chan struct{}, wg *sync.WaitGroup) error {
	defer wg.Done()

	reader := bufio.NewReader(os.Stdin)
	o.run = true
	for o.run == true {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("%v\n", err)
		} else {
			if line != "" {
				if o.Timestamp == true {
					line = fmt.Sprintf("[%s %s] ", o.Name(), time.Now().Format(time.RFC3339)) + line
				}
				err = queue.Put([]byte(line))
				if err != nil {
					fmt.Printf("%v\n", err)
				}
			}
		}

		select {
		case <-done:
			o.run = false
		default:
			time.Sleep(time.Millisecond)
		}
	}
	return nil
}
