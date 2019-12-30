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

// StdinCfg contains config
type StdinCfg struct {
	Name      string
	Timestamp bool
}

// StdinStats contains stats
type StdinStats struct {
	MessagesReceived uint64
	MessagesSent     uint64
	Running          bool
}

// Stdin contains what is needed for ingester
type Stdin struct {
	cfg       StdinCfg
	run       bool
	statsChan chan StdinStats
	clearChan chan struct{}
}

// New is what reservoird uses to create and start stdin
func New(cfg string) (icd.Ingester, error) {
	c := StdinCfg{
		Name:      "com.reservoird.ingest.stdin",
		Timestamp: false,
	}
	if cfg != "" {
		d, err := ioutil.ReadFile(cfg)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(d, &c)
		if err != nil {
			return nil, err
		}
	}
	o := &Stdin{
		cfg:       c,
		run:       true,
		statsChan: make(chan StdinStats),
		clearChan: make(chan struct{}),
	}
	return o, nil
}

// Name return the name of the Ingester
func (o *Stdin) Name() string {
	return o.cfg.Name
}

// Stats returns marshalled stats NOTE: thread safe
func (o *Stdin) Stats() (string, error) {
	select {
	case stats := <-o.statsChan:
		data, err := json.Marshal(stats)
		if err != nil {
			return "", err
		}
		return string(data), nil
	default:
		return "", nil
	}
}

// ClearStats clears stats NOTE: thread safe
func (o *Stdin) ClearStats() {
	select {
	case o.clearChan <- struct{}{}:
	default:
	}
}

// Ingest reads data from stdin and writes it to the queue
func (o *Stdin) Ingest(queue icd.Queue, done <-chan struct{}, wg *sync.WaitGroup) error {
	defer wg.Done() // required

	stats := StdinStats{}

	reader := bufio.NewReader(os.Stdin)
	o.run = true
	for o.run == true {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("%v\n", err)
		} else {
			stats.MessagesReceived = stats.MessagesReceived + 1
			if queue.Closed() == false {
				if len(line) != 0 {
					if o.cfg.Timestamp == true {
						line = fmt.Sprintf("[%s %s] ", o.Name(), time.Now().Format(time.RFC3339)) + line
					}
					err = queue.Put([]byte(line))
					if err != nil {
						fmt.Printf("%v\n", err)
					} else {
						stats.MessagesSent = stats.MessagesSent + 1
					}
				}
			}
		}

		// listens for clear stats
		select {
		case <-o.clearChan:
			stats = StdinStats{}
		default:
		}

		// listens for shutdown
		select {
		case <-done:
			o.run = false
		default:
		}

		// sends latest stats
		stats.Running = o.run
		select {
		case o.statsChan <- stats:
		default:
		}

		time.Sleep(time.Millisecond)
	}
	return nil
}
