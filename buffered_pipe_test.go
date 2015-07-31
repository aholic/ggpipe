package ggpipe

import (
	"bytes"
	"testing"
)

func TestCapacity(t *testing.T) {
	cp := 3
	pr, pw := BufferedPipe(cp)

	rcp := pr.Capacity()
	if rcp != cp {
		t.Errorf("expect: %v, actual: %v", cp, rcp)
	}

	wcp := pw.Capacity()
	if wcp != cp {
		t.Errorf("expect: %v, actual: %v", cp, wcp)
	}
}

func TestWriter(t *testing.T) {
	data := []byte("hello")
	cp := 3

	_, pw := BufferedPipe(cp)

	n, err := pw.Write(data)
	if err != nil {
		t.Errorf("expect: %v, actual: %v", nil, err)
	}

	if n != cp {
		t.Errorf("expect: %vv, actual: %v", cp, n)
	}
}

func TestReader(t *testing.T) {
	data := []byte("hello")
	cp := 10

	pr, pw := BufferedPipe(cp)
	if n, err := pw.Write(data); err != nil {
		t.Errorf("expect: %v, actual: %v", nil, err)
	} else if n != len(data) {
		t.Errorf("expect: %v, actual: %v", len(data), n)
	}

	tmp := make([]byte, 10)
	if n, err := pr.Read(tmp); err != nil {
		t.Errorf("expect: %v, actual: %v", nil, err)
	} else if n != len(data) {
		t.Errorf("expect: %v, actual: %v", len(data), n)
	} else if bytes.Compare(tmp[:n], data) != 0 {
		t.Errorf("expect: %v, actual: %v", string(data), string(tmp[:n]))
	}
}

func TestBufferd(t *testing.T) {
	data := []byte("hello")
	cp := 10
	pr, pw := BufferedPipe(cp)

	pw.Write(data)
	if n := pw.Buffered(); n != len(data) {
		t.Errorf("expect: %v, actual: %v", len(data), n)
	}

	tmp := make([]byte, 10)
	pr.Read(tmp)
	if n := pw.Buffered(); n != 0 {
		t.Errorf("expect: %v, actual: %v", 0, n)
	}
}
