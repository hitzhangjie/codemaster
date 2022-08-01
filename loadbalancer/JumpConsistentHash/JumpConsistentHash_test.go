package JumpConsistentHash

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"testing"

	jump "github.com/lithammer/go-jump-consistent-hash"
	stat "github.com/montanaflynn/stats"
)

// jump consistent hash 感觉负载不均衡比较明显
// 标准差，还是最大、最小负载只差占总负载的比例，这个要优于 go-zero 中实现的一致性hash方案，
// 但是jump consistent hash本身的局限性：服务名不能是任意的，要求按序号递增，更推荐应用于 data shards 场景.
/*
2022/08/01 15:40:35 times: 10000 标准方差: 3.053518398401261  max: 20  min: 1 (max-min)/times: 0.0019
2022/08/01 15:40:35 times: 20000 标准方差: 4.36639779308986  max: 35  min: 7 (max-min)/times: 0.0014
2022/08/01 15:40:35 times: 30000 标准方差: 5.286309521715031  max: 50  min: 12 (max-min)/times: 0.0012666666666666666
2022/08/01 15:40:35 times: 40000 标准方差: 6.125159436700403  max: 60  min: 18 (max-min)/times: 0.00105
2022/08/01 15:40:35 times: 50000 标准方差: 6.668617471738426  max: 72  min: 29 (max-min)/times: 0.00086
2022/08/01 15:40:35 times: 60000 标准方差: 7.2218608015870815  max: 81  min: 37 (max-min)/times: 0.0007333333333333333
2022/08/01 15:40:35 times: 70000 标准方差: 7.977859376698426  max: 99  min: 46 (max-min)/times: 0.0007571428571428572
2022/08/01 15:40:35 times: 80000 标准方差: 8.549945175262822  max: 110  min: 56 (max-min)/times: 0.000675
2022/08/01 15:40:35 times: 90000 标准方差: 9.262752505836211  max: 123  min: 61 (max-min)/times: 0.0006888888888888888
2022/08/01 15:40:35 times: 100000 标准方差: 10.039521511879936  max: 132  min: 66 (max-min)/times: 0.00066
2022/08/01 15:40:35 times: 110000 标准方差: 10.366042349150181  max: 140  min: 72 (max-min)/times: 0.0006181818181818182
2022/08/01 15:40:35 times: 120000 标准方差: 10.88361727552012  max: 152  min: 78 (max-min)/times: 0.0006166666666666666
2022/08/01 15:40:35 times: 130000 标准方差: 11.481972537825328  max: 161  min: 91 (max-min)/times: 0.0005384615384615384
2022/08/01 15:40:35 times: 140000 标准方差: 11.724915045001392  max: 172  min: 100 (max-min)/times: 0.0005142857142857143
...
2022/08/01 15:40:36 times: 980000 标准方差: 30.98886203642367  max: 1045  min: 860 (max-min)/times: 0.00018877551020408164
2022/08/01 15:40:36 times: 990000 标准方差: 31.32222513023899  max: 1062  min: 871 (max-min)/times: 0.00019292929292929293
2022/08/01 15:40:36 times: 1000000 标准方差: 31.39525616036601  max: 1070  min: 881 (max-min)/times: 0.000189
*/
func Test_JumpConsistentHash(t *testing.T) {
	const numOfNodes int32 = 1024
	const limitOfSdev float64 = 5

	// 代表hash到buckets中bucket-i的次数
	count := [numOfNodes]float64{}

	var times int
	for times < 1000000 {
		for {
			times++
			// 默认source是均匀分布的伪随机值，所以用来模拟playerid也合适
			playerid := rand.Uint64()
			h := jump.Hash(playerid, numOfNodes)
			count[h]++

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
}

func TestStat(t *testing.T) {
	nums := []float64{1, 1, 1, 1}
	sdev, _ := stat.Variance(nums)
	fmt.Println("sdev:", sdev)
}
