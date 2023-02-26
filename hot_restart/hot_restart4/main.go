package main

import (
	"errors"
	"net"
)

func ResolveUnixAddr(name string) *net.UnixAddr {
	return &net.UnixAddr{
		Name: name,
		Net:  "unixgram",
	}
}

func NewUnixDomainSocket(name string) (*net.UnixConn, error) {
	addr, err := net.ResolveUnixAddr("unixgram", name)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUnixgram("unixgram", addr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func SendToDomainSocket(conn *net.UnixConn, data []byte, addr *net.UnixAddr) (int, error) {
	if conn == nil {
		return 0, errors.New("invalid conn")
	}
	return conn.WriteToUnix(data, addr)
}

func ReadFromDomainSocket(conn *net.UnixConn, data []byte) (int, *net.UnixAddr, error) {
	return conn.ReadFromUnix(data)
}
