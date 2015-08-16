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
	cp, wc1, wc2, wc3 := 10, 3, 6, 1

	_, pw := BufferedPipe(cp)

	// write 3 bytes
	if n, err := pw.Write(data[:wc1], wc1); err != nil {
		t.Errorf("expect: %v, actual: %v", nil, err)
	} else if n != wc1 {
		t.Errorf("expect: %v, actual: %v", wc1, n)
	}

	//write 6 bytes
	if n, err := pw.BlockWrite(data[:wc2]); err != nil {
		t.Errorf("expect: %v, actual: %v", nil, err)
	} else if n != wc2 {
		t.Errorf("expect: %v, actual: %v", wc2, n)
	}

	//write error
	if n, err := pw.Write(data, len(data)); n != 0 {
		t.Errorf("expect: %v, actual: %v", 0, n)
	} else if err != ErrTooLargeProvide {
		t.Errorf("expect: %v, actual: %v", ErrTooLargeProvide, err)
	}

	//write 1 byte
	if n, err := pw.DeadlockFreeBlockWrite(data[:wc3], wc3); err != nil {
		t.Errorf("expect: %v, actual: %v", nil, err)
	} else if n != wc3 {
		t.Errorf("expect: %v, actual %v", wc3, n)
	}

	//write 0 byte
	if n, err := pw.NonBlockWrite(data); err != nil {
		t.Errorf("expect: %v, actual: %v", nil, err)
	} else if n != 0 {
		t.Errorf("expect: %v, actual %v", 0, n)
	}

	pw.Close()
}

func TestReader(t *testing.T) {
	data := []byte("hello world!")
	cp, wc1, rc1, wc2, rc2, wc3, rc3, rc4 := 5, 5, 5, 3, 3, 3, 3, 5
	pr, pw := BufferedPipe(cp)

	pw.BlockWrite(data[:wc1])
	tmp := make([]byte, rc1)
	if n, err := pr.Read(tmp, len(tmp)); err != nil {
		t.Errorf("expect: %v, actual: %v", nil, err)
	} else if n != rc1 {
		t.Errorf("expect: %v, actual: %v", rc1, n)
	} else if bytes.Compare(tmp[:rc1], data[:rc1]) != 0 {
		t.Errorf("expect: %v, actual: %v", string(data[:rc1]), string(tmp[:rc1]))
	}

	pw.BlockWrite(data[:wc2])
	tmp = make([]byte, rc2)
	if n, err := pr.BlockRead(tmp); err != nil {
		t.Errorf("expect: %v, actual: %v", nil, err)
	} else if n != rc2 {
		t.Errorf("expect: %v, actual: %v", rc2, n)
	} else if bytes.Compare(tmp[:rc2], data[:rc2]) != 0 {
		t.Errorf("expect: %v, actual: %v", string(data[:rc2]), string(tmp[:rc2]))
	}

	pw.BlockWrite(data[:wc3])
	tmp = make([]byte, rc3)
	if n, err := pr.DeadlockFreeBlockRead(tmp, rc3); err != nil {
		t.Errorf("expect: %v, actual: %v", nil, err)
	} else if n != rc3 {
		t.Errorf("expect: %v, actual: %v", rc3, n)
	} else if bytes.Compare(tmp[:rc3], data[:rc3]) != 0 {
		t.Errorf("expect: %v, actual: %v", string(data[:rc3]), string(tmp[:rc3]))
	}

	tmp = make([]byte, rc4)
	if n, err := pr.NonBlockRead(tmp); err != nil {
		t.Errorf("expect: %v, actual: %v", nil, err)
	} else if n != 0 {
		t.Errorf("expect: %v, actual: %v", 0, n)
	}

	pw.Close()
	if _, err := pr.Read(nil, 1); err != io.EOF {
		t.Errorf("expect: %v, actual: %v", io.EOF, err)
	}
}

func TestBufferd(t *testing.T) {
	data := []byte("hello")
	cp, rc, wc := 10, 3, 5
	pr, pw := BufferedPipe(cp)

	pw.Write(data[:wc], wc)
	if n := pw.Buffered(); n != wc {
		t.Errorf("expect: %v, actual: %v", wc, n)
	}

	tmp := make([]byte, rc)
	pr.Read(tmp, rc)
	if n := pw.Buffered(); n != wc-rc {
		t.Errorf("expect: %v, actual: %v", wc-rc, n)
	}
}
