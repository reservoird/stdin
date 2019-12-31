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
	Name             string
	MessagesReceived uint64
	MessagesSent     uint64
	Running          bool
	Monitoring       bool
}

// Stdin contains what is needed for ingester
type Stdin struct {
	cfg       StdinCfg
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
		statsChan: make(chan StdinStats, 1),
		clearChan: make(chan struct{}, 1),
	}
	return o, nil
}

// Name return the name of the Ingester
func (o *Stdin) Name() string {
	return o.cfg.Name
}

// Monitor provides stats and clear stats
func (o *Stdin) Monitor(statsChan chan<- string, clearChan <-chan struct{}, doneChan <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done() // required

	stats := StdinStats{}
	monrun := true
	for monrun == true {
		// clear
		select {
		case <-clearChan:
			select {
			case o.clearChan <- struct{}{}:
			default:
			}
		default:
		}

		// done
		select {
		case <-doneChan:
			monrun = false
			stats.Monitoring = monrun
		default:
		}

		// get stats from ingest
		select {
		case stats = <-o.statsChan:
			stats.Monitoring = monrun
		default:
		}

		// marshal
		data, err := json.Marshal(stats)
		if err != nil {
			fmt.Printf("%v\n", err)
		} else {
			// send stats to reservoird
			select {
			case statsChan <- string(data):
			default:
			}
		}

		if monrun == true {
			time.Sleep(time.Millisecond)
		}
	}
}

// Ingest reads data from stdin and writes it to the queue
func (o *Stdin) Ingest(queue icd.Queue, done <-chan struct{}, wg *sync.WaitGroup) error {
	defer wg.Done() // required

	stats := StdinStats{}
	reader := bufio.NewReader(os.Stdin)

	run := true
	stats.Name = o.cfg.Name
	stats.Running = run
	for run == true {
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

		// clear stats
		select {
		case <-o.clearChan:
			stats = StdinStats{}
			stats.Name = o.cfg.Name
			stats.Running = run
		default:
		}

		// listens for shutdown
		select {
		case <-done:
			run = false
			stats.Running = run
		default:
		}

		// send to monitor
		select {
		case o.statsChan <- stats:
		default:
		}

		if run == true {
			time.Sleep(time.Millisecond)
		}
	}
	return nil
}
