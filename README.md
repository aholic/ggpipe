# ggpipe

##### 需求 #####

想要在俩goroutine之间传递字节流，或者简单的来说是要实现字节流的生产者消费者

##### 方案 #####

1. channel

    这种情况下，如果用channel的话，每次生产者得一个一个的往里放，消费者也得一个一个的往外取，会不会效率有点低？

2. thread(goroutine) safe buffer

    如果俩goroutine之间共享一个线程安全的buffer，每次生产这把字节流往buffer里面写，消费者从buffer里面读出来，是不是效率高点

3. io.Pipe

    这个io.Pipe的阻塞方式有点蛋疼，读写操作必须一一对应，不符合需求

##### 结果 #####

呵呵呵，打脸了，方案1的速度远超方案2，快了大概10倍吧。其实也有可能是我写的thread(goroutine) safe buffer写的太渣了。
不过我还是继续channel吧，真tm伤心，又白写了好一会儿。

##### 结论 #####

channel大法好

##### 代码在哪 #####

    ggpipe/buffered_pipe_b_test.go
