package ggpipe

import "testing"

func BenchmarkBufferedPipe(b *testing.B) {
	cp := 100
	data := []byte("hello, i am a benchmark test, i hope it will be faster than channel")

	pr, pw := BufferedPipe(cp)
	done := make(chan bool)

	go func() {
		for i := 0; i < b.N; i++ {
			pw.Write(data)
		}
		pw.Close()
		done <- true
	}()
	go func() {
		tmp := make([]byte, 100)
		for {
			_, err := pr.Read(tmp)
			if err != nil {
				break
			}
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
