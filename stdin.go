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

type stdinCfg struct {
	Name      string
	Timestamp bool
}

type stdinStats struct {
	MsgsRead uint64
	Active   bool
}

type stdin struct {
	cfg       stdinCfg
	stats     stdinStats
	statsChan chan<- string
}

// New is what reservoird uses to create and start stdin
func New(cfg string, statsChan chan<- string) (icd.Ingester, error) {
	c := stdinCfg{
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
	o := &stdin{
		cfg:       c,
		stats:     stdinStats{},
		statsChan: statsChan,
	}
	return o, nil
}

// Name return the name of the Ingester
func (o *stdin) Name() string {
	return o.cfg.Name
}

// Ingest reads data from stdin and writes it to the queue
func (o *stdin) Ingest(queue icd.Queue, done <-chan struct{}, wg *sync.WaitGroup) error {
	defer wg.Done() // required

	reader := bufio.NewReader(os.Stdin)
	o.stats.Active = true
	for o.stats.Active == true {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("%v\n", err)
		} else {
			o.stats.MsgsRead = o.stats.MsgsRead + 1
			if queue.Closed() == false {
				if len(line) != 0 {
					if o.cfg.Timestamp == true {
						line = fmt.Sprintf("[%s %s] ", o.Name(), time.Now().Format(time.RFC3339)) + line
					}
					err = queue.Put([]byte(line))
					if err != nil {
						fmt.Printf("%v\n", err)
					}
				}
			}
		}

		select {
		case <-done:
			o.stats.Active = false
		default:
		}

		stats, err := json.Marshal(o.stats)
		if err != nil {
			fmt.Printf("%v\n", err)
		}
		select {
		case o.statsChan <- string(stats):
		default:
		}

		time.Sleep(time.Millisecond)
	}
	return nil
}
