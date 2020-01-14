package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
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
	cfg StdinCfg
	run bool
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
		cfg: c,
		run: false,
	}
	return o, nil
}

// Name return the name of the Ingester
func (o *Stdin) Name() string {
	return o.cfg.Name
}

// Running states wheter or not ingest is running
func (o *Stdin) Running() bool {
	return o.run
}

// Ingest reads data from stdin and writes it to the queue
func (o *Stdin) Ingest(queue icd.Queue, mc *icd.MonitorControl) {
	defer mc.WaitGroup.Done() // required

	reader := bufio.NewReader(os.Stdin)

	o.run = true

	stats := StdinStats{}
	stats.Name = o.cfg.Name
	stats.Running = o.run

	// stdin blocks so send message early
	select {
	case mc.StatsChan <- stats:
	default:
	}

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

		// clear stats
		select {
		case <-mc.ClearChan:
			stats = StdinStats{
				Name:    o.cfg.Name,
				Running: o.run,
			}
		default:
		}

		// send stats
		select {
		case mc.StatsChan <- stats:
		default:
		}

		// listens for shutdown
		select {
		case <-mc.DoneChan:
			fmt.Printf("stdin.done\n")
			o.run = false
			stats.Running = o.run
		default:
		}

		if o.run == true {
			time.Sleep(time.Millisecond)
		}
	}

	// send final stats
	select {
	case mc.StatsChan <- stats:
	default:
	}
}
