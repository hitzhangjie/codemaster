package ConsistentHash_gozero

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"strconv"
	"testing"

	stat "github.com/montanaflynn/stats"
	"github.com/zeromicro/go-zero/core/hash"
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

// 负载也不十分均衡，标准方差很大，ok，我们关注下 (max-min)/times 吧，
// 最大负载和最小负载的差距，占总负载的百分比：这个差不多稳定在 5% 左右。
/*
2022/08/01 16:30:32 times: 10000 标准方差: 139.45823747631403  max: 1263  min: 778 (max-min)/times: 0.0485
2022/08/01 16:30:32 times: 20000 标准方差: 299.2417083228874  max: 2593  min: 1503 (max-min)/times: 0.0545
2022/08/01 16:30:32 times: 30000 标准方差: 445.59174139564124  max: 3836  min: 2290 (max-min)/times: 0.051533333333333334
2022/08/01 16:30:32 times: 40000 标准方差: 588.7150414249666  max: 5090  min: 3046 (max-min)/times: 0.0511
2022/08/01 16:30:32 times: 50000 标准方差: 743.5193339786128  max: 6364  min: 3807 (max-min)/times: 0.05114
2022/08/01 16:30:32 times: 60000 标准方差: 882.5136826134766  max: 7644  min: 4641 (max-min)/times: 0.05005
2022/08/01 16:30:32 times: 70000 标准方差: 1045.6639039385457  max: 8889  min: 5360 (max-min)/times: 0.05041428571428572
2022/08/01 16:30:32 times: 80000 标准方差: 1181.2622062861403  max: 10132  min: 6132 (max-min)/times: 0.05
2022/08/01 16:30:32 times: 90000 标准方差: 1313.245750040715  max: 11381  min: 6930 (max-min)/times: 0.04945555555555556
2022/08/01 16:30:32 times: 100000 标准方差: 1464.6088897722832  max: 12632  min: 7644 (max-min)/times: 0.04988
2022/08/01 16:30:32 times: 110000 标准方差: 1598.7097297508387  max: 13880  min: 8410 (max-min)/times: 0.049727272727272724
...
2022/08/01 16:30:32 times: 950000 标准方差: 13873.532261107839  max: 120543  min: 72623 (max-min)/times: 0.05044210526315789
2022/08/01 16:30:32 times: 960000 标准方差: 14034.086090657987  max: 121817  min: 73363 (max-min)/times: 0.050472916666666666
2022/08/01 16:30:32 times: 970000 标准方差: 14175.134002893941  max: 123039  min: 74104 (max-min)/times: 0.05044845360824742
2022/08/01 16:30:32 times: 980000 标准方差: 14327.665071462272  max: 124256  min: 74861 (max-min)/times: 0.050403061224489794
2022/08/01 16:30:32 times: 990000 标准方差: 14470.17734514681  max: 125500  min: 75630 (max-min)/times: 0.050373737373737376
2022/08/01 16:30:32 times: 1000000 标准方差: 14628.560790453721  max: 126737  min: 76395 (max-min)/times: 0.050342
[88947 76395 104058 99622 93732 109073 126737 94078 120192 87166]
*/
func Test_ConsistentHash(t *testing.T) {
	chash := hash.NewConsistentHash()
	for i := 0; i < len(hosts); i++ {
		chash.Add(hosts[i])
	}

	// 代表hash到buckets中bucket-i的次数
	count := make([]float64, len(hosts))

	var times int
	for times < 1000000 {
		for {
			times++
			// 默认source是均匀分布的伪随机值，所以用来模拟playerid也合适
			playerid := strconv.Itoa(rand.Int())

			h, ok := chash.Get(playerid)
			if !ok {
				log.Println("WARN: no host")
				continue
			}
			idx, ok := hostsIndex[h.(string)]
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
