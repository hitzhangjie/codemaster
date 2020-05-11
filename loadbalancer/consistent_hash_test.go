package loadbalancer

import (
	"math/rand"
	"sync"
	"testing"

	chash "github.com/lafikl/consistent"
)

var (
	hosts = []string{
		"10.10.10.1",
		"10.10.10.2",
		"10.10.10.3",
		"10.10.10.4",
		"10.10.10.5",
		"10.10.10.6",
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
