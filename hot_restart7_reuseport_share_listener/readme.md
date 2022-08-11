
```bash
$ go build -v

$ ./share_tcpconn
2022/06/08 18:44:27 pid: 43587
2022/06/08 18:44:27 tcp server started
```

### open a new shell and run `nc` as tcp client

```bash
$ nc -vt localhost 8888 
nc: connectx to localhost port 8888 (tcp) failed: Connection refused
Connection to localhost port 8888 [tcp/ddi-tcp-1] succeeded!

# then we input 1234 and <enter> to send data to tcp server
1234
1234
1234
1234
1234
1234
```

### we also see the data read and output by the tcp server

```bash
2022/06/08 18:44:36 pid: 43587, conn read 5 bytes, val: [49 50 51 52 10 0 0 0]
2022/06/08 18:44:37 pid: 43587, conn read 5 bytes, val: [49 50 51 52 10 0 0 0]
2022/06/08 18:44:39 pid: 43587, conn read 5 bytes, val: [49 50 51 52 10 0 0 0]
2022/06/08 18:44:40 pid: 43587, conn read 5 bytes, val: [49 50 51 52 10 0 0 0]
2022/06/08 18:44:41 pid: 43587, conn read 5 bytes, val: [49 50 51 52 10 0 0 0]
2022/06/08 18:44:42 pid: 43587, conn read 5 bytes, val: [49 50 51 52 10 0 0 0]
```

### open new shell and send SIGUSR2 to trigger tcpconn passed to forked process

```bash
$ kill -SIGUSR2 43587
```

then we see:

```bash
2022/06/08 18:44:55 ready to restart and pass connfd
2022/06/08 18:44:55 forked pid: 43658
2022/06/08 18:44:55 pid: 43658
2022/06/08 18:44:55 recv tcpconn and read
```

### then we go back to the 2nd shell and input data in `nc`

```bash
1234
1234
1234
1234
1234
1234
1234 <= keep input data
1111111111111111111111111122222222222222222222222222 <= 超过了buffer大小，目的是增加两个进程分别loop read的几率
```

then we see the data is read and output by the the parent and foked child process:

```bash
2022/06/08 18:45:01 pid: 43658, conn read 5 bytes, val: [49 50 51 52 10 0 0 0] <= pid 43658
2022/06/08 18:45:03 pid: 43658, conn read 5 bytes, val: [49 50 51 52 10 0 0 0]
2022/06/08 18:45:04 pid: 43658, conn read 5 bytes, val: [49 50 51 52 10 0 0 0]
....
2022/06/08 ........ pid: p1, conn read 8 bytes, val: [49 49 49 49 49 49 49 49] <= 可以看到数据在两个进程循环被读取到, p1表示父进程，p2表示子进程
2022/06/08 ........ pid: p2, conn read 8 bytes, val: [49 49 49 49 49 49 49 49]
2022/06/08 ........ pid: p1, conn read 8 bytes, val: [49 49 49 49 49 49 49 49]
2022/06/08 ........ pid: p2, conn read 8 bytes, val: [49 50 50 50 50 50 50 50]
2022/06/08 ........ pid: p1, conn read 8 bytes, val: [50 50 50 50 50 50 50 50]
2022/06/08 ........ pid: p2, conn read 8 bytes, val: [50 50 50 50 50 50 50 50]
2022/06/08 ........ pid: p1, conn read 6 bytes, val: [50 50 50 50 50 10 0 0]
```

有什么办法让p1不读取，只让p2读取呢？那不是很简单：没读取到应用程序buf中的数据是停留在socket buf中的，p1检测到热重启停止读取就可以了啊，让p2读取就是后到的数据，
如果需要迁移连接数据的话，直接让p1把连接上的数据通过unix套接字传过去给p2，p2把后面读到的连接上的数据追加在p1传过来的数据末尾，就完成数据迁移了。