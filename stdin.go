package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"
)

// needed to aid in unit testing
type ireader interface {
	ReadString(byte) (string, error)
}

type stdin struct {
	Timestamp bool
	reader    ireader
	run       func() bool
}

func forever() bool {
	return true
}

// Config configures ingester
func (o *stdin) Config(cfg string) error {
	// default
	o.Timestamp = false
	o.reader = bufio.NewReader(os.Stdin)
	o.run = forever

	if cfg != "" {
		b, err := ioutil.ReadFile(cfg)
		if err != nil {
			return err
		}
		s := stdin{}
		err = json.Unmarshal(b, &s)
		if err != nil {
			return err
		}
		o.Timestamp = s.Timestamp
	}
	return nil
}

// Ingest reads data from stdin and writes it to a channel
func (o *stdin) Ingest(channel chan<- []byte) error {
	for o.run() == true {
		line, err := o.reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				return err
			}
		}
		if o.Timestamp == true {
			channel <- []byte(fmt.Sprintf("rditime %s %s", time.Now().Format(time.RFC3339), line))
		} else {
			channel <- []byte(line)
		}
	}
	return nil
}

// Ingester for stdin
var Ingester stdin
