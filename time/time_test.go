package main

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type Event struct {
	Time string `json:"time"`
}

// time.Time格式化存储到字符串时注意带上时区信息，如果layout里不带时区，t.Format(layout)会按Local时区，
// 但是time.Parse(layout, t)时会按UTC，这样根据时间比较先后时会出问题……
//
// 更建议用unixnano代替t.Format后的字符串
func Test_format_without_timezone(t *testing.T) {
	//layout := "2006-01-02/15:04:05.999999999"
	layout := "2006-01-02/15:04:05.999999999 -0700 MST"

	t1 := time.Now()
	fmt.Println(t1)
	evt1, _ := json.Marshal(Event{Time: t1.Format(layout)})
	fmt.Println(string(evt1))

	time.Sleep(time.Second)
	t2 := time.Now()
	evt2, _ := json.Marshal(Event{Time: t2.Format(layout)})
	fmt.Println(t2)
	fmt.Println(string(evt2))

	fmt.Println("----------------")

	var e1, e2 Event
	_ = json.Unmarshal(evt1, &e1)
	_ = json.Unmarshal(evt2, &e2)
	fmt.Println(e1.Time)
	fmt.Println(e2.Time)

	tt1, _ := time.Parse(layout, e1.Time)
	fmt.Println(tt1)

	tt2, _ := time.Parse(layout, e2.Time)
	fmt.Println(tt2)

	assert.Equal(t, t1.UnixNano(), tt1.UnixNano())
	assert.Equal(t, t2.UnixNano(), tt2.UnixNano())
}

func Test_truncate(t *testing.T) {
	now := time.Now()
	fmt.Println(now.Format("2006-01-02 15:04:05.999"))

	// truncate `now` to day
	fmt.Println(now.Truncate(time.Hour))

	// 跟时区有关系，hour不一定是0
	fmt.Println(now.Truncate(time.Hour * 24))
}

func Test_format_with_UTC_timezone(t *testing.T) {
	location := time.FixedZone("Asia/Shanghai", 8*3600)

	now := time.Unix(1734583230, 0)
	println(now.Unix())

	layoutGN := "2006-01-02 15:04:05"
	println(now.Format(layoutGN))
	println(now.In(location).Format(layoutGN))

	// 这里的time.Format()的行为特点, see: time.AppendFormat(...)函数处理过程即可得知：
	//
	// 1. Format()方法总是使用时间值本身携带的时区信息来格式化，而不是layout字符串中指定的时区
	// 2. 即使layout中写了"UTC +0000"，它也只是作为输出格式的模板，不会再格式化时单独额外处理hh:mm:ss的偏移量
	//    这个偏移量是从time.Time中直接拿到的，格式化串可以理解为只是一些年月日时分秒时区位置信息的占位符，仅此而已

	// wrong:
	// 这两种预期会转换成按照UTC 0时区格式化的字符串（如layoutGlobal中声明的那样），这种预期都是错误的，
	// now + now.Local()，当前获取now时，服务器时区设置是东八区，这里都是按照东八区进行格式化
	layoutGlobal := "2006-01-02 15:04:05 UTC +0000"
	println(now.Format(layoutGlobal))
	println(now.Local().Format(layoutGlobal))
	println(now.In(location).Format(layoutGlobal))

	// right: 这种才是正常的
	println(now.UTC().Format(layoutGlobal))
}
