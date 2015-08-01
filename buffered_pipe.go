package ggpipe

// provide a buffered pipe to connect two thread(goroutine)

import (
	"errors"
	"io"
	"sync"
)

var (
	// occurs when ReadBlock with a block size larger than capacity
	ErrTooLargeDemand = errors.New("demand is not supposed to be larger than capacity")

	// occurs when WriteBlock with a block size larger than capacity
	ErrTooLargeProvide = errors.New("provide is not supposed to be larger than capacity")

	// default error occurs when any operation after the reader call close
	ErrReaderClosed = errors.New("pipe already close by reader")
)

type bufferedPipe struct {
	lock     sync.Mutex // guard for members
	data     []byte     // data buffer
	cp       int        // the max size of data can be stored in buffer
	sz       int        // the size of data currently stored in buffer
	notEmpty sync.Cond  // tell reader to read
	notFull  sync.Cond  // tell writer to write
	rerr     error      // the ending err the reader want to tell writer
	werr     error      // the ending err the writer want to tell reader
}

// return the capacity of the buffer
// it's thread safe since it will not change once the pipe is created
func (bp *bufferedPipe) capacity() int {
	return bp.cp
}

// tells the size of data currently stored in the pipe
func (bp *bufferedPipe) buffered() int {
	bp.lock.Lock()
	defer bp.lock.Unlock()

	return bp.sz
}

// read data from pipe, block until blockSize bytes has been read
func (bp *bufferedPipe) read(b []byte, blockSize int) (int, error) {
	bp.lock.Lock()
	defer bp.lock.Unlock()

	for {
		if bp.rerr != nil {
			return 0, bp.rerr
		}

		if bp.sz >= blockSize {
			break
		}

		if bp.werr != nil {
			return 0, bp.werr
		}

		if blockSize > bp.cp {
			return 0, ErrTooLargeDemand
		}

		bp.notEmpty.Wait()
	}

	n := copy(b, bp.data)
	bp.sz -= n

	if bp.sz < bp.cp {
		bp.notFull.Signal()
	}

	return n, nil
}

// write data to pipe, block until blockSize bytes has been written
func (bp *bufferedPipe) write(b []byte, blockSize int) (int, error) {
	bp.lock.Lock()
	defer bp.lock.Unlock()

	for {
		if bp.werr != nil {
			return 0, bp.werr
		}

		if bp.sz+blockSize <= bp.cp {
			break
		}

		if bp.rerr != nil {
			return 0, bp.rerr
		}

		if blockSize > bp.cp {
			return 0, ErrTooLargeProvide
		}

		bp.notFull.Wait()
	}

	n := copy(bp.data[bp.sz:], b)
	bp.sz += n
	if bp.sz > 0 {
		bp.notEmpty.Signal()
	}

	return n, nil
}

// reader close the pipe
func (bp *bufferedPipe) rclose(err error) {
	if err == nil {
		err = ErrReaderClosed
	}

	bp.lock.Lock()
	defer bp.lock.Unlock()

	bp.rerr = err
	bp.notEmpty.Signal()
	bp.notFull.Signal()
}

// writer close the pipe
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

// read data from pipe as much as possible, max : len(data)
// if there is no data currently in the pipe, it will block
func (r *BufferedPipeReader) Read(data []byte) (n int, err error) {
	return r.bp.read(data, 1)
}

// block until len(data) bytes has ben read
// warn: a deadlock may happen when ReadBlock and WriteBlock both be used
func (r *BufferedPipeReader) ReadBlock(data []byte) (n int, err error) {
	return r.bp.read(data, len(data))
}

// normal close, without error
func (r *BufferedPipeReader) Close() error {
	return r.CloseWithError(nil)
}

// close with a error that will be got by the writer
func (r *BufferedPipeReader) CloseWithError(err error) error {
	r.bp.rclose(err)
	return nil
}

// tells the size of data currently stored in the pipe
func (r *BufferedPipeReader) Buffered() int {
	return r.bp.buffered()
}

// the max size of data that can be stored in the pipe
func (r *BufferedPipeReader) Capacity() int {
	return r.bp.capacity()
}

type BufferedPipeWriter struct {
	bp *bufferedPipe
}

// write data to pipe as much as possible, max : len(data)
// if there is no data currently can be written to the pipe, it will block
func (w *BufferedPipeWriter) Write(data []byte) (n int, err error) {
	return w.bp.write(data, 1)
}

// block until len(data) bytes has ben written
// warn: a deadlock may happen when ReadBlock and WriteBlock both be used
func (w *BufferedPipeWriter) WriteBlock(data []byte) (n int, err error) {
	return w.bp.write(data, len(data))
}

// close with a error that will be got by the reader
func (w *BufferedPipeWriter) CloseWithError(err error) error {
	w.bp.wclose(err)
	return nil
}

// normal close, without error
func (w *BufferedPipeWriter) Close() error {
	return w.CloseWithError(nil)
}

// tells the size of data currently stored in the pipe
func (w *BufferedPipeWriter) Buffered() int {
	return w.bp.buffered()
}

// the max size of data that can be stored in the pipe
func (w *BufferedPipeWriter) Capacity() int {
	return w.bp.capacity()
}

// new a pipe with capacity
func BufferedPipe(cp int) (*BufferedPipeReader, *BufferedPipeWriter) {
	bp := new(bufferedPipe)
	bp.cp = cp
	bp.data = make([]byte, cp)
	bp.notFull.L = &bp.lock
	bp.notEmpty.L = &bp.lock
	pr := &BufferedPipeReader{bp}
	pw := &BufferedPipeWriter{bp}

	return pr, pw
}
