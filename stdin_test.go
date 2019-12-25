package main

//go:generate mkdir -p mocks/iboolmocks
//go:generate mockgen -source $GOPATH/pkg/mod/github.com/reservoird/ibool@v1.0.0/ibool.go -package iboolmocks -destination mocks/iboolmocks/ibool_mock.go

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
	if s.keepRunning.Val() == false {
		t.Errorf("expecting true but got false")
	}
}

type boolbridgetest struct {
	count int
	val   bool
}

func newboolbridgetest() *boolbridgetest {
	b := new(boolbridgetest)
	b.count = 0
	b.val = true
	return b
}

func (o *boolbridgetest) Val() bool {
	val := o.val
	if o.count == 1 {
		o.val = false
	}
	o.count = o.count + 1
	return val
}

func TestDigest(t *testing.T) {
	s := stdin{}
	err := s.Config("")
	if err != nil {
		t.Errorf("expecting nil but got error: %v", err)
	}
	s.keepRunning = newboolbridgetest()
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

type bufreadtest struct {
	count int
	str   string
}

func newbufreadtest(str string) *bufreadtest {
	b := new(bufreadtest)
	b.count = 0
	b.str = str
	return b
}

func (b *bufreadtest) ReadString(delim byte) (string, error) {
	if b.count == 1 {
		return "", fmt.Errorf("mock error")
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
	s.keepRunning = newboolbridgetest()
	s.reader = newbufreadtest("hello\n")
	dst := make(chan []byte, 2)
	err = s.Ingest(dst)
	if err == nil {
		t.Errorf("expecting error but got nil")
	}
}
