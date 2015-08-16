package ggpipe

//provide a buffered pipe to connect two thread(goroutine)

import (
	"errors"
	"io"
	"sync"
)

var (
	//occurs when BlockRead with a block size larger than capacity
	ErrTooLargeDemand = errors.New("read is not supposed to be larger than capacity")

	//occurs when BlockWrite with a block size larger than capacity
	ErrTooLargeProvide = errors.New("write is not supposed to be larger than capacity")

	//default error occurs when any operation after the reader call close
	ErrReaderClosed = errors.New("pipe already close by reader")
)

type bufferedPipe struct {
	lock     sync.Mutex //guard for members
	data     []byte     //data buffer
	cp       int        //the max size of data can be stored in buffer
	sz       int        //the size of data currently stored in buffer
	canRead  sync.Cond  //tell reader to read
	canWrite sync.Cond  //tell writer to write
	rerr     error      //the ending err the reader want to tell writer
	werr     error      //the ending err the writer want to tell reader
}

//return the capacity of the buffer.
//it's thread safe since it will not change once the pipe is created.
func (bp *bufferedPipe) capacity() int {
	return bp.cp
}

//tells the size of data currently stored in the pipe.
func (bp *bufferedPipe) buffered() int {
	bp.lock.Lock()
	defer bp.lock.Unlock()

	return bp.sz
}

//read data from pipe, block until blockSize bytes has been read.
func (bp *bufferedPipe) read(b []byte, blockSize int) (int, error) {
	if blockSize > bp.cp {
		return 0, ErrTooLargeDemand
	}

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

		bp.canRead.Wait()
	}

	n := copy(b, bp.data[:bp.sz])
	bp.sz -= n
	copy(bp.data[:bp.sz], bp.data[n:n+bp.sz])

	if bp.sz < bp.cp {
		bp.canWrite.Signal()
	}

	return n, nil
}

//write data to pipe, block until blockSize bytes has been written.
func (bp *bufferedPipe) write(b []byte, blockSize int) (int, error) {
	if blockSize > bp.cp {
		return 0, ErrTooLargeProvide
	}

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

		bp.canWrite.Wait()
	}

	n := copy(bp.data[bp.sz:], b)
	bp.sz += n
	if bp.sz > 0 {
		bp.canRead.Signal()
	}

	return n, nil
}

//reader close the pipe.
func (bp *bufferedPipe) rclose(err error) {
	if err == nil {
		err = ErrReaderClosed
	}

	bp.lock.Lock()
	defer bp.lock.Unlock()

	bp.rerr = err
	bp.canRead.Signal()
	bp.canWrite.Signal()
}

//writer close the pipe.
func (bp *bufferedPipe) wclose(err error) {
	if err == nil {
		err = io.EOF
	}

	bp.lock.Lock()
	defer bp.lock.Unlock()

	bp.werr = err
	bp.canWrite.Signal()
	bp.canRead.Signal()
}

type BufferedPipeReader struct {
	bp *bufferedPipe
}

//read from a buffer, block until [least] bytes can be read at the same time
//warn: a deadlock may happen, see http://
func (r *BufferedPipeReader) Read(data []byte, least int) (n int, err error) {
	return r.bp.read(data, least)
}

//block until [len(data)] bytes can be read at the same time
//warn: a deadlock may happen, see http://
func (r *BufferedPipeReader) BlockRead(data []byte) (n int, err error) {
	return r.bp.read(data, len(data))
}

//read from a buffer, block until [least] bytes has been read
//deadlock free, but in the worst case, it will read byte one by one
func (w *BufferedPipeReader) DeadlockFreeBlockRead(data []byte, least int) (n int, err error) {
	tmp := 0
	for n < least && err == nil {
		tmp, err = w.bp.read(data, 1)
		n += tmp
		data = data[tmp:]
	}
	return
}

//read at most [len(data)] bytes, return immediately if no data available
func (w *BufferedPipeReader) NonBlockRead(data []byte) (n int, err error) {
	return w.bp.read(data, 0)
}

//normal close, without error.
func (r *BufferedPipeReader) Close() error {
	return r.CloseWithError(nil)
}

//close with a error that will be got by the writer.
func (r *BufferedPipeReader) CloseWithError(err error) error {
	r.bp.rclose(err)
	return nil
}

//tells the size of data currently stored in the pipe.
func (r *BufferedPipeReader) Buffered() int {
	return r.bp.buffered()
}

//the max size of data that can be stored in the pipe.
func (r *BufferedPipeReader) Capacity() int {
	return r.bp.capacity()
}

type BufferedPipeWriter struct {
	bp *bufferedPipe
}

//write to a buffer, block until [least] bytes has been written
//deadlock free, but in the worst case, it will write byte one by one
func (w *BufferedPipeWriter) DeadlockFreeBlockWrite(data []byte, least int) (n int, err error) {
	tmp := 0
	for n < least && err == nil {
		tmp, err = w.bp.write(data, 1)
		n += tmp
		data = data[tmp:]
	}
	return
}

//write to a buffer, block until [least] bytes can be written at the same time
//warn: a deadlock may happen, see http://
func (w *BufferedPipeWriter) Write(data []byte, least int) (n int, err error) {
	return w.bp.write(data, least)
}

//block until all bytes in [data] can be written at the same time
//warn: a deadlock may happen, see http://
func (w *BufferedPipeWriter) BlockWrite(data []byte) (n int, err error) {
	return w.bp.write(data, len(data))
}

//write at most len(data) bytes, return immediately if no buffer space available
func (w *BufferedPipeWriter) NonBlockWrite(data []byte) (n int, err error) {
	return w.bp.write(data, 0)
}

//close with a error that will be got by the reader.
func (w *BufferedPipeWriter) CloseWithError(err error) error {
	w.bp.wclose(err)
	return nil
}

//normal close, without error.
func (w *BufferedPipeWriter) Close() error {
	return w.CloseWithError(nil)
}

//tells the size of data currently stored in the pipe.
func (w *BufferedPipeWriter) Buffered() int {
	return w.bp.buffered()
}

//the max size of data that can be stored in the pipe.
func (w *BufferedPipeWriter) Capacity() int {
	return w.bp.capacity()
}

//new a pipe with capacity.
func BufferedPipe(cp int) (*BufferedPipeReader, *BufferedPipeWriter) {
	if cp <= 0 {
		panic("no sense for capacity lower then 1")
	}

	bp := new(bufferedPipe)
	bp.cp = cp
	bp.data = make([]byte, cp)
	bp.canWrite.L = &bp.lock
	bp.canRead.L = &bp.lock
	pr := &BufferedPipeReader{bp}
	pw := &BufferedPipeWriter{bp}

	return pr, pw
}
