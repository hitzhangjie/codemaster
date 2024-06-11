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

func Test_timezone(t *testing.T) {
	// time.Time格式化存储到字符串时注意带上时区信息，因为t.Format(layout)会按Local时区，
	// 但是time.Parse(layout, t)时会按UTC，这样根据时间比较先后时会出问题……更建议用unixnano代替t.Format后的字符串、time.Time
	//
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
