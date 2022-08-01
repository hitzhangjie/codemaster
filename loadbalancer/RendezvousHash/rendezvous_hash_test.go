package RendezvousHash

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"strconv"
	"testing"

	xxhash "github.com/cespare/xxhash/v2"
	rdv "github.com/dgryski/go-rendezvous"
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

// 看上去比较均匀，我感觉跟这里的 key、host 同时用于计算hash有关系，这样算可能相对来说更容易实现平衡
//
// - rendezvous vs go-zero
// 100w请求时，方差387，远小于前面go-zero设计实现中的1w+
// 最大最小节点负载占总负载比例，1/1000，也小于go-zero设计实现中的5/100
/*
2022/08/01 17:18:00 times: 10000 标准方差: 34.50507209092599  max: 1069  min: 953 (max-min)/times: 0.0116
2022/08/01 17:18:00 times: 20000 标准方差: 38.08936859544931  max: 2048  min: 1935 (max-min)/times: 0.00565
2022/08/01 17:18:00 times: 30000 标准方差: 53.01509219080921  max: 3110  min: 2941 (max-min)/times: 0.005633333333333333
2022/08/01 17:18:00 times: 40000 标准方差: 58.47221562417487  max: 4084  min: 3887 (max-min)/times: 0.004925
2022/08/01 17:18:00 times: 50000 标准方差: 65.16287286484535  max: 5079  min: 4872 (max-min)/times: 0.00414
2022/08/01 17:18:00 times: 60000 标准方差: 71.39047555521675  max: 6084  min: 5866 (max-min)/times: 0.0036333333333333335
2022/08/01 17:18:00 times: 70000 标准方差: 72.72826135691682  max: 7098  min: 6883 (max-min)/times: 0.0030714285714285713
...
2022/08/01 17:18:00 times: 110000 标准方差: 89.65712464718015  max: 11201  min: 10861 (max-min)/times: 0.0030909090909090908
2022/08/01 17:18:00 times: 940000 标准方差: 346.9270816756743  max: 94595  min: 93528 (max-min)/times: 0.0011351063829787235
2022/08/01 17:18:00 times: 950000 标准方差: 353.50473829922  max: 95607  min: 94529 (max-min)/times: 0.0011347368421052632
2022/08/01 17:18:00 times: 960000 标准方差: 358.48319346937313  max: 96617  min: 95516 (max-min)/times: 0.001146875
2022/08/01 17:18:00 times: 970000 标准方差: 346.94034069274795  max: 97570  min: 96542 (max-min)/times: 0.0010597938144329896
2022/08/01 17:18:00 times: 980000 标准方差: 362.84266562795506  max: 98646  min: 97493 (max-min)/times: 0.0011765306122448979
2022/08/01 17:18:00 times: 990000 标准方差: 365.7203849937818  max: 99651  min: 98484 (max-min)/times: 0.0011787878787878788
2022/08/01 17:18:00 times: 1000000 标准方差: 387.6954990711138  max: 100689  min: 99455 (max-min)/times: 0.001234
[99738 99631 99943 100216 100474 100355 100689 99735 99455 99764]
*/
func TestXXXX(t *testing.T) {
	rdvHash := rdv.New(hosts, xxhash.Sum64String)

	// 代表hash到buckets中bucket-i的次数
	count := make([]float64, len(hosts))

	var times int
	for times < 1000000 {
		for {
			times++
			// 默认source是均匀分布的伪随机值，所以用来模拟playerid也合适
			playerid := strconv.Itoa(rand.Int())

			h := rdvHash.Lookup(playerid)
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
