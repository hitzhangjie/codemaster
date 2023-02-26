
### build

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
1234
1234
1234
1234
1234
```

then we see the data is read and output by the the parent and foked child process:

```bash
2022/06/08 18:45:01 pid: 43658, conn read 5 bytes, val: [49 50 51 52 10 0 0 0] <= pid 43658
2022/06/08 18:45:03 pid: 43658, conn read 5 bytes, val: [49 50 51 52 10 0 0 0]
2022/06/08 18:45:04 pid: 43658, conn read 5 bytes, val: [49 50 51 52 10 0 0 0]
2022/06/08 18:45:05 pid: 43587, conn read 5 bytes, val: [49 50 51 52 10 0 0 0] <= pid 43587
2022/06/08 18:45:07 pid: 43587, conn read 5 bytes, val: [49 50 51 52 10 0 0 0]
2022/06/08 18:45:08 pid: 43587, conn read 5 bytes, val: [49 50 51 52 10 0 0 0]

```

obviously, the network stack distribute the IO read-ready event to the different processes!

we can also use `lsof` to make sure whether the tcpconn is shared between the parent and forked processes.
Absolutely Yes!

What's more we see is that on darwin/amd64, each process seems create two file descriptors for the same tcpconn,
Is it some optimization Go does?

```bash
$ lsof -Pi tcp:8888

COMMAND     PID     USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME
share_tcp 43587 zhangjie    3u  IPv4 0x70f6484c55859ca5      0t0  TCP *:8888 (LISTEN)
share_tcp 43587 zhangjie    7u  IPv4 0x70f6484c68318f05      0t0  TCP localhost:8888->localhost:62019 (ESTABLISHED)
share_tcp 43587 zhangjie   10u  IPv4 0x70f6484c68318f05      0t0  TCP localhost:8888->localhost:62019 (ESTABLISHED)
nc        43618 zhangjie    5u  IPv4 0x70f6484c4c67b8e5      0t0  TCP localhost:62019->localhost:8888 (ESTABLISHED)
share_tcp 43658 zhangjie    3u  IPv4 0x70f6484c68318f05      0t0  TCP localhost:8888->localhost:62019 (ESTABLISHED)
share_tcp 43658 zhangjie    4u  IPv4 0x70f6484c68318f05      0t0  TCP localhost:8888->localhost:62019 (ESTABLISHED)
```

Actually, No! The Fd will be duplicated when we call `tcpconn.File().Fd()` and `os.NewFile(uintptr(fd), "name").

After add some f.Close() then we test it again and find `lsof -Pi tcp:8888` outputs:

```bash
$ lsof -Pi tcp:8888
COMMAND     PID     USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME
share_tcp 47491 zhangjie    3u  IPv4 0x70f6484c5162d8e5      0t0  TCP *:8888 (LISTEN)
share_tcp 47491 zhangjie    7u  IPv4 0x70f6484c58ad1a45      0t0  TCP localhost:8888->localhost:64529 (ESTABLISHED)
nc        47514 zhangjie    5u  IPv4 0x70f6484c71cba685      0t0  TCP localhost:64529->localhost:8888 (ESTABLISHED)
share_tcp 47543 zhangjie    4u  IPv4 0x70f6484c58ad1a45      0t0  TCP localhost:8888->localhost:64529 (ESTABLISHED)
```

The same tcpconn has only 1 Fd, which is the expected result.


