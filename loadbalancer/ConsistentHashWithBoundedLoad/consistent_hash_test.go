package ConsistentHashWithBoundedLoad

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"strconv"
	"sync"
	"testing"

	chash "github.com/lafikl/consistent"
	stat "github.com/montanaflynn/stats"
)

var (
	hosts = []string{
		"1.1.1.1",
		"1.1.1.2",
		"1.1.1.3",
		"1.1.1.4",
		"1.1.1.5",
		"1.1.1.6",
		"1.1.1.7",
		"1.1.1.8",
		"1.1.1.9",
		"1.1.1.10",
	}

	hostsIndex = map[string]int{
		"1.1.1.1":  0,
		"1.1.1.2":  1,
		"1.1.1.3":  2,
		"1.1.1.4":  3,
		"1.1.1.5":  4,
		"1.1.1.6":  5,
		"1.1.1.7":  6,
		"1.1.1.8":  7,
		"1.1.1.9":  8,
		"1.1.1.10": 9,
	}
)

// 不考虑节点负载的情况, h.Get(key)
//
// 按照一致性hash环上的顺序，查找hash(key)对应的位置的下一个节点，返回该节点
func TestConsistentHash(t *testing.T) {
	h := chash.New()

	for _, host := range hosts {
		h.Add(host)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		uin := "10000"
		for i := 0; i < 10; i++ {
			host, err := h.Get(uin)
			if err != nil {
				t.Fatalf("uin:%s, consistent hash get host error: %v", uin, err)
			}
			if i%5 == 0 {
				h.Remove(host)
			}
			t.Logf("uin:%s, consistent hash get host ok, host: %s", uin, host)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		uin := "20000"
		for i := 0; i < 10; i++ {
			host, err := h.Get(uin)
			if err != nil {
				t.Fatalf("uin:%s, consistent hash get host error: %v", uin, err)
			}
			if i%5 == 0 {
				h.Remove(host)
			}
			t.Logf("uin:%s, consistent hash get host ok, host: %s", uin, host)
		}
	}()

	wg.Wait()
}

// 考虑节点负载的情况，h.GetLeast(key)
//
// 如果按照一致性hash环上顺序查找hash(key)对应位置的下一个节点，该节点的负载不超过所有节点平均负载的情况下，则返回该节点；
// 反之继续查找下一个节点
func TestConsistentHashWithLoad(t *testing.T) {
	h := chash.New()

	for _, host := range hosts {
		h.Add(host)
	}

	uin := "10000"
	for i := 0; i < 10; i++ {
		host, err := h.GetLeast(uin)
		if err != nil {
			t.Fatalf("uin:%s, consistent hash get host error: %v", uin, err)
		}
		// simulate load (for example timecost)
		load := rand.Int() % 10
		h.UpdateLoad(host, int64(load))

		t.Logf("uin:%s, consistent hash get host ok, host: %s, update load: %d", uin, host, load)
	}
}

func Test_ConsistentHashWithBoundedLoad(t *testing.T) {
	chash := chash.New()
	for i := 0; i < len(hosts); i++ {
		chash.Add(hosts[i])
	}

	var numOfNodes = len(hosts)

	// 代表hash到buckets中bucket-i的次数
	count := make([]float64, numOfNodes)

	var times int
	for times < 1000000 {
		for {
			times++
			// 默认source是均匀分布的伪随机值，所以用来模拟playerid也合适
			playerid := strconv.Itoa(rand.Int())

			h, err := chash.Get(playerid)
			if err != nil {
				log.Println("WARN: no host")
				continue
			}
			idx, ok := hostsIndex[h]
			if !ok {
				log.Println("WARN: invalid data")
				continue
			}
			count[idx]++

			if times%10000 == 0 {
				break
			}
		}

		// 计算hash到的次数的方差
		data := count[:]
		sdev, _ := stat.Variance(data)
		sdev = math.Pow(sdev, 0.5)
		max, _ := stat.Max(data)
		min, _ := stat.Min(data)
		log.Println("times:", times, "标准方差:", sdev, " max:", max, " min:", min, "(max-min)/times:", float64(max-min)/float64(times))
	}
	fmt.Println(count)
}
