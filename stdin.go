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
	stats     StdinStats
	statsLock sync.Mutex
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
		stats:     StdinStats{},
		statsLock: sync.Mutex{},
	}
	return o, nil
}

// Name return the name of the Ingester
func (o *Stdin) Name() string {
	return o.cfg.Name
}

// Stats returns marshalled stats NOTE: thread safe
func (o *Stdin) Stats() (string, error) {
	o.statsLock.Lock()
	defer o.statsLock.Unlock()
	data, err := json.Marshal(o.stats)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ClearStats clears stats NOTE: thread safe
func (o *Stdin) ClearStats() {
	o.statsLock.Lock()
	defer o.statsLock.Unlock()
	o.stats = StdinStats{}
}

func (o *Stdin) incMessagesReceived() {
	o.statsLock.Lock()
	defer o.statsLock.Unlock()
	o.stats.MessagesReceived = o.stats.MessagesReceived + 1
}

func (o *Stdin) incMessagesSent() {
	o.statsLock.Lock()
	defer o.statsLock.Unlock()
	o.stats.MessagesSent = o.stats.MessagesSent + 1
}

func (o *Stdin) setRunning(run bool) {
	o.statsLock.Lock()
	defer o.statsLock.Unlock()
	o.stats.Running = run
}

// Ingest reads data from stdin and writes it to the queue
func (o *Stdin) Ingest(queue icd.Queue, done <-chan struct{}, wg *sync.WaitGroup) error {
	defer wg.Done() // required

	reader := bufio.NewReader(os.Stdin)
	o.run = true
	o.setRunning(o.run)
	for o.run == true {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("%v\n", err)
		} else {
			o.incMessagesReceived()
			if queue.Closed() == false {
				if len(line) != 0 {
					if o.cfg.Timestamp == true {
						line = fmt.Sprintf("[%s %s] ", o.Name(), time.Now().Format(time.RFC3339)) + line
					}
					err = queue.Put([]byte(line))
					if err != nil {
						fmt.Printf("%v\n", err)
					} else {
						o.incMessagesSent()
					}
				}
			}
		}

		// listens for shutdown
		select {
		case <-done:
			o.run = false
			o.setRunning(o.run)
		default:
		}

		time.Sleep(time.Millisecond)
	}
	return nil
}
