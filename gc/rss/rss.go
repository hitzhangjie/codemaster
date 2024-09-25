package rss

import (
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"syscall"
)

var pageSize = syscall.Getpagesize()

// GetRSS returns the resident set size of the current process
func GetRSS() uint64 {
	data, err := ioutil.ReadFile("/proc/self/stat")
	if err != nil {
		log.Fatal(err)
	}
	fs := strings.Fields(string(data))
	rss, err := strconv.ParseUint(fs[23], 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	return uint64(rss) * uint64(pageSize)
}
