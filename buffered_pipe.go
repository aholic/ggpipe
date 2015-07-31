package ggpipe

// provide a buffered pipe to connect two thread(goroutine)
// actually, it's a BoundedBlockingQueue

import (
	"bytes"
	"errors"
	"io"
	"sync"
)

type bufferedPipe struct {
	lock     sync.Mutex
	data     bytes.Buffer
	cp       int
	notEmpty sync.Cond
	notFull  sync.Cond
	rerr     error
	werr     error
}

func (bp *bufferedPipe) capacity() int {
	return bp.cp
}

func (bp *bufferedPipe) buffered() int {
	bp.lock.Lock()
	defer bp.lock.Unlock()

	return bp.data.Len()
}

func (bp *bufferedPipe) read(b []byte) (int, error) {
	bp.lock.Lock()
	defer bp.lock.Unlock()

	for {
		if bp.rerr != nil {
			return 0, errors.New("pipe already close by reader")
		}

		if bp.data.Len() != 0 {
			break
		}

		if bp.werr != nil {
			return 0, errors.New("pipe already close by writer")
		}

		bp.notEmpty.Wait()
	}

	n, err := bp.data.Read(b)
	if bp.data.Len() < bp.cp {
		bp.notFull.Signal()
	}

	return n, err
}

func (bp *bufferedPipe) write(b []byte) (int, error) {
	bp.lock.Lock()
	defer bp.lock.Unlock()

	for {
		if bp.werr != nil {
			return 0, errors.New("pipe already close by writer")
		}

		if bp.data.Len() < bp.cp {
			break
		}

		if bp.rerr != nil {
			return 0, errors.New("pipe already close by reader")
		}

		bp.notFull.Wait()
	}

	n := len(b)
	if bp.data.Len()+n > bp.cp {
		n = bp.cp - bp.data.Len()
	}

	n, err := bp.data.Write(b[:n])
	if bp.data.Len() > 0 {
		bp.notEmpty.Signal()
	}

	return n, err
}

func (bp *bufferedPipe) rclose(err error) {
	if err == nil {
		err = errors.New("pipe already close by reader")
	}

	bp.lock.Lock()
	defer bp.lock.Unlock()

	bp.rerr = err
	bp.notEmpty.Signal()
	bp.notFull.Signal()
}

func (bp *bufferedPipe) wclose(err error) {
	if err == nil {
		err = io.EOF
	}

	bp.lock.Lock()
	defer bp.lock.Unlock()

	bp.werr = err
	bp.notFull.Signal()
	bp.notEmpty.Signal()
}

type BufferedPipeReader struct {
	bp *bufferedPipe
}

func (r *BufferedPipeReader) Read(data []byte) (n int, err error) {
	return r.bp.read(data)
}

func (r *BufferedPipeReader) Close() error {
	return r.CloseWithError(nil)
}

func (r *BufferedPipeReader) CloseWithError(err error) error {
	r.bp.rclose(err)
	return nil
}

func (r *BufferedPipeReader) Buffered() int {
	return r.bp.buffered()
}

func (r *BufferedPipeReader) Capacity() int {
	return r.bp.capacity()
}

type BufferedPipeWriter struct {
	bp *bufferedPipe
}

func (w *BufferedPipeWriter) Write(data []byte) (n int, err error) {
	return w.bp.write(data)
}

func (w *BufferedPipeWriter) CloseWithError(err error) error {
	w.bp.wclose(err)
	return nil
}

func (w *BufferedPipeWriter) Close() error {
	return w.CloseWithError(nil)
}

func (w *BufferedPipeWriter) Buffered() int {
	return w.bp.buffered()
}

func (w *BufferedPipeWriter) Capacity() int {
	return w.bp.capacity()
}

func BufferedPipe(cp int) (*BufferedPipeReader, *BufferedPipeWriter) {
	bp := new(bufferedPipe)
	bp.cp = cp
	bp.notFull.L = &bp.lock
	bp.notEmpty.L = &bp.lock
	pr := &BufferedPipeReader{bp}
	pw := &BufferedPipeWriter{bp}

	return pr, pw
}
