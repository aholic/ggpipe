package ggpipe

import "testing"

func BenchmarkBufferedPipe(b *testing.B) {
	cp := 100
	data := []byte("hello, i am a benchmark test, i hope it will be faster than channel")

	pr, pw := BufferedPipe(cp)
	done := make(chan bool)

	go func() {
		for i := 0; i < b.N; i++ {
			for written := 0; written < len(data); {
				//n, _ := pw.NonBlockWrite(data[written:])
				n, _ := pw.DeadlockFreeBlockWrite(data[written:], len(data[written:]))
				written += n
			}
		}
		pw.Close()
		done <- true
	}()
	go func() {
		tmp := make([]byte, cp)
		cnt := 0
		for cnt < b.N*len(data) {
			//n, err := pr.NonBlockRead(tmp)
			n, err := pr.DeadlockFreeBlockRead(tmp, len(tmp))
			if err != nil {
				break
			}
			cnt += n
		}
		done <- true
	}()

	<-done
	<-done
}

func BenchmarkBufferedChannel(b *testing.B) {
	cp := 100
	data := []byte("hello, i am a benchmark test, i hope it will be faster than channel")

	done := make(chan bool)
	pipe := make(chan byte, cp)

	go func() {
		for i := 0; i < b.N; i++ {
			for j := 0; j < len(data); j++ {
				pipe <- data[j]
			}
		}
		done <- true
	}()
	go func() {
		for i := 0; i < b.N*len(data); i++ {
			<-pipe
		}
		done <- true
	}()

	<-done
	<-done
}
