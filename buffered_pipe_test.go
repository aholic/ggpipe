package ggpipe

import (
	"bytes"
	"io"
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
	data := []byte("hello world!")
	cp, wc1, wc2 := 10, 3, 7

	_, pw := BufferedPipe(cp)

	if n, err := pw.Write(data[:wc1]); err != nil {
		t.Errorf("expect: %v, actual: %v", nil, err)
	} else if n != wc1 {
		t.Errorf("expect: %v, actual: %v", wc1, n)
	}

	if n, err := pw.WriteBlock(data[:wc2]); err != nil {
		t.Errorf("expect: %v, actual: %v", nil, err)
	} else if n != wc2 {
		t.Errorf("expect: %v, actual: %v", wc1, n)
	}

	if n, err := pw.WriteBlock(data); n != 0 {
		t.Errorf("expect: %v, actual: %v", 0, n)
	} else if err != ErrTooLargeProvide {
		t.Errorf("expect: %v, actual: %v", ErrTooLargeProvide, err)
	}
	pw.Close()
}

func TestReader(t *testing.T) {
	data := []byte("hello world!")
	cp, wc1, rc1, wc2, rc2 := 5, 5, 5, 3, 3
	pr, pw := BufferedPipe(cp)

	pw.WriteBlock(data[:wc1])
	tmp := make([]byte, rc1+1)
	if n, err := pr.Read(tmp); err != nil {
		t.Errorf("expect: %v, actual: %v", nil, err)
	} else if n != rc1 {
		t.Errorf("expect: %v, actual: %v", rc1, n)
	} else if bytes.Compare(tmp[:rc1], data[:rc1]) != 0 {
		t.Errorf("expect: %v, actual: %v", string(data[:rc1]), string(tmp[:rc1]))
	}

	pw.WriteBlock(data[:wc2])
	tmp = make([]byte, rc2)
	if n, err := pr.ReadBlock(tmp); err != nil {
		t.Errorf("expect: %v, actual: %v", nil, err)
	} else if n != rc2 {
		t.Errorf("expect: %v, actual: %v", rc2, n)
	} else if bytes.Compare(tmp[:rc2], data[:rc2]) != 0 {
		t.Errorf("expect: %v, actual: %v", string(data[:rc2]), string(tmp[:rc2]))
	}

	pw.Close()
	if _, err := pr.Read(nil); err != io.EOF {
		t.Errorf("expect: %v, actual: %v", io.EOF, err)
	}
}

func TestBufferd(t *testing.T) {
	data := []byte("hello")
	cp, rc, wc := 10, 3, 5
	pr, pw := BufferedPipe(cp)

	pw.Write(data[:wc])
	if n := pw.Buffered(); n != wc {
		t.Errorf("expect: %v, actual: %v", wc, n)
	}

	tmp := make([]byte, rc)
	pr.Read(tmp)
	if n := pw.Buffered(); n != wc-rc {
		t.Errorf("expect: %v, actual: %v", wc-rc, n)
	}
}
