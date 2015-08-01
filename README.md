# ggpipe

##### 需求 #####

想要在俩goroutine之间传递字节流，或者简单的来说是要实现字节流的生产者消费者

##### 方案 #####

1. channel

    这种情况下，如果用channel的话，每次生产者得一个一个的往里放，消费者也得一个一个的往外取，会不会效率有点低？

2. thread(goroutine) safe buffer, BufferedPipe

    如果俩goroutine之间共享一个线程安全的buffer，每次生产这把字节流往buffer里面写，消费者从buffer里面读出来，是不是效率高点

3. io.Pipe

    这个io.Pipe的阻塞方式有点蛋疼，读写操作必须一一对应，不符合需求

##### 结果 #####

~~呵呵呵，打脸了，方案1的速度远超方案2，快了大概10倍吧。其实也有可能是我写的thread(goroutine) safe buffer写的太渣了。
不过我还是继续channel吧，真tm伤心，又白写了好一会儿。~~

睡觉的时候一直觉得很奇怪，明显一块一块的拷贝比一个一个的拷贝好啊，凭啥慢这么多。
而且go test -bench的输出里面，x ns/op的operation，指的是什么operation啊？

结果呵呵呵，我又tm被打脸了，原来不是operation，是loop，尼玛，也就是说BufferedPipe比channel快10倍。
然而我想起来了，我昨晚看反了，看成了op/ns，擦。智商真拙计。
后来发现bench test里面的代码有一点小问题，BufferedPipe传输的字节稍微少了一点，所以改了下代码，大概快7倍的样子。

##### 结论 #####

~~channel大法好~~

没文化真可怕

##### 代码在哪 #####

    bench test: ggpipe/buffered_pipe_b_test.go
    unit test: ggpipe/buffered_pipe_test.go
    source code: ggpipe/buffered_pipe.go
