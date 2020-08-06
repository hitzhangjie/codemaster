# Server graceful restart with Go

## Install and run the server

```
$ go get github.com/hitzhangjie/codemaster/hot_restart2
$ go-graceful-restart-example
2014/12/14 20:26:42 [Server - 4301] Listen on [::]:12345
[...]
```

## Connect with the client

```
$ cd $GOPATH/src/github.com/hitzhangjie/codemaster/hot_restart2/client
$ go run pong.go
```

## Graceful restart

```
# The server pid is included in its log, in the example: 4301

$ kill -HUP <server pid>
```

## Stop with timeout

Let 10 seconds for the current requests to finish.

```
$ kill -TERM <server pid>
```

## Gist of output

https://gist.github.com/Soulou/7ca6a2d4f475f8e2345e
