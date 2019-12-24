package main

import (
	"bufio"
	"fmt"
	"strings"
	"testing"
)

func TestConfig(t *testing.T) {
	s := stdin{}
	err := s.Config("")
	if err != nil {
		t.Errorf("expecting nil but got error: %v", err)
	}
	if s.run() == false {
		t.Errorf("expecting true but got false")
	}
}

var flag bool = true
var timesCount int = 0

func times() bool {
	val := flag
	if timesCount == 1 {
		flag = false
	}
	timesCount = timesCount + 1
	return val
}

func TestDigest(t *testing.T) {
	s := stdin{}
	err := s.Config("")
	if err != nil {
		t.Errorf("expecting nil but got error: %v", err)
	}
	flag = true
	timesCount = 0
	s.run = times
	s.reader = bufio.NewReader(strings.NewReader("hello\nhello\n"))
	expected := []byte("hello\n")
	dst := make(chan []byte, 2)
	err = s.Ingest(dst)
	if err != nil {
		t.Errorf("expecting nil but got error: %v", err)
	}
	actual := <-dst
	if string(actual) != string(expected) {
		t.Errorf("expecting %s but got %s", string(expected), string(actual))
	}
	actual = <-dst
	if string(actual) != string(expected) {
		t.Errorf("expecting %s but got %s", string(expected), string(actual))
	}
}

type bufrdtst struct {
	count int
	str   string
}

func newbufrdtst(str string) *bufrdtst {
	b := new(bufrdtst)
	b.count = 0
	b.str = str
	return b
}

func (b *bufrdtst) ReadString(delim byte) (string, error) {
	if b.count == 1 {
		return "", fmt.Errorf("test error")
	}
	b.count = b.count + 1
	return b.str, nil
}

func TestDigestError(t *testing.T) {
	s := stdin{}
	err := s.Config("")
	if err != nil {
		t.Errorf("expecting nil but got error: %v", err)
	}
	flag = true
	timesCount = 0
	s.run = times
	s.reader = newbufrdtst("hello\n")
	dst := make(chan []byte, 2)
	err = s.Ingest(dst)
	if err == nil {
		t.Errorf("expecting error but got nil")
	}
}
