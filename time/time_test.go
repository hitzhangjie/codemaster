package main

import (
	"fmt"
	"testing"
	"time"
)

// 获取指定时间t对应着星期几
func Test_weekday(t *testing.T) {
	now := time.Now()
	println(now.Format("2006-01-02"))
	println(now.Weekday().String())
	println("--------")

	date := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	monday := date.AddDate(0, 0, -1*int(now.Weekday()-1))
	println(monday.Weekday().String())
	println(monday.Format("2006-01-02"))
}

func Test_truncate(t *testing.T) {
	now := time.Now()
	fmt.Println(now.Format("2006-01-02 15:04:05.999"))

	// truncate `now` to day
	fmt.Println(now.Truncate(time.Hour))

	// 跟时区有关系，hour不一定是0
	fmt.Println(now.Truncate(time.Hour * 24))
}
