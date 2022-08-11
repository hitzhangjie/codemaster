package ConsistentHash_gozero

import (
	"fmt"
	"hash/crc32"
	"log"
	"math"
	"math/rand"
	"strconv"
	"testing"

	"github.com/cespare/xxhash/v2"
	stat "github.com/montanaflynn/stats"
	"github.com/zeebo/xxh3"

	"github.com/hitzhangjie/codemaster/loadbalancer/ConsistentHash_gozero/hash"
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

// ring-based一致性hash不均匀的问题可以通过优化hash函数、虚节点数来优化，
// peak-to-mean优化到1.05基本就不错了，表示最大负载和平均负载相差不超过5%
//
// 比较之下，xxhash的均匀度是比较好的，而且其性能在其他zero-allocation hash benchmark测评中也比较好。
/*
2022/08/03 15:28:47 case: replicas:50+hash:murmur3.Sum64 times: 1000000 标准方差: 14628.560790453721  max: 126737  min: 76395 (max-min)/times: 0.050342 peak/mean: 1.26737
2022/08/03 15:28:47 case: replicas:100+hash:murmur3.Sum64 times: 1000000 标准方差: 14555.295022774357  max: 127129  min: 76438 (max-min)/times: 0.050691 peak/mean: 1.27129
2022/08/03 15:28:48 case: replicas:200+hash:murmur3.Sum64 times: 1000000 标准方差: 6902.00454940447  max: 110178  min: 85121 (max-min)/times: 0.025057 peak/mean: 1.10178
2022/08/03 15:28:48 case: replicas:500+hash:murmur3.Sum64 times: 1000000 标准方差: 2285.3205902017335  max: 105277  min: 97136 (max-min)/times: 0.008141 peak/mean: 1.05277
2022/08/03 15:28:48 case: replicas:1000+hash:murmur3.Sum64 times: 1000000 标准方差: 2069.765928794848  max: 104603  min: 97606 (max-min)/times: 0.006997 peak/mean: 1.04603
2022/08/03 15:28:48 case: replicas:2000+hash:murmur3.Sum64 times: 1000000 标准方差: 2618.900303562547  max: 104628  min: 94870 (max-min)/times: 0.009758 peak/mean: 1.04628
-------------------------------------------------------------------------------------------------------------------------------------------------------------------------
2022/08/03 15:28:48 case: replicas:50+hash:xxhash.Sum64 times: 1000000 标准方差: 8627.559643375409  max: 119229  min: 91110 (max-min)/times: 0.028119 peak/mean: 1.19229
2022/08/03 15:28:49 case: replicas:100+hash:xxhash.Sum64 times: 1000000 标准方差: 8918.29840272235  max: 120236  min: 90692 (max-min)/times: 0.029544 peak/mean: 1.20236
2022/08/03 15:28:49 case: replicas:200+hash:xxhash.Sum64 times: 1000000 标准方差: 5913.828556865679  max: 111947  min: 89811 (max-min)/times: 0.022136 peak/mean: 1.11947
2022/08/03 15:28:49 case: replicas:500+hash:xxhash.Sum64 times: 1000000 标准方差: 4256.551350565384  max: 107631  min: 93326 (max-min)/times: 0.014305 peak/mean: 1.07631
2022/08/03 15:28:49 case: replicas:1000+hash:xxhash.Sum64 times: 1000000 标准方差: 3148.5766943176086  max: 106134  min: 95150 (max-min)/times: 0.010984 peak/mean: 1.06134
2022/08/03 15:28:50 case: replicas:2000+hash:xxhash.Sum64 times: 1000000 标准方差: 1664.1786562746202  max: 103375  min: 96885 (max-min)/times: 0.00649 peak/mean: 1.03375
-------------------------------------------------------------------------------------------------------------------------------------------------------------------------
2022/08/11 16:40:45 case: replicas:50+hash:xxhash3.Hash64 times: 1000000 标准方差: 7603.68696094204  max: 109469  min: 82863 (max-min)/times: 0.026606 peak/mean: 1.09469 ok
2022/08/11 16:40:45 case: replicas:100+hash:xxhash3.Hash64 times: 1000000 标准方差: 7820.6347440601  max: 110422  min: 82739 (max-min)/times: 0.027683 peak/mean: 1.10422 ok
2022/08/11 16:40:45 case: replicas:200+hash:xxhash3.Hash64 times: 1000000 标准方差: 7966.721772473293  max: 114431  min: 86124 (max-min)/times: 0.028307 peak/mean: 1.14431 worse than xxhash
2022/08/11 16:40:46 case: replicas:500+hash:xxhash3.Hash64 times: 1000000 标准方差: 6829.993455340935  max: 112575  min: 87118 (max-min)/times: 0.025457 peak/mean: 1.12575 worse than xxhash
2022/08/11 16:40:46 case: replicas:1000+hash:xxhash3.Hash64 times: 1000000 标准方差: 2860.680373617437  max: 104301  min: 95025 (max-min)/times: 0.009276 peak/mean: 1.04301 ok
2022/08/11 16:40:46 case: replicas:2000+hash:xxhash3.Hash64 times: 1000000 标准方差: 2263.951059541703  max: 104415  min: 95723 (max-min)/times: 0.008692 peak/mean: 1.04415 worse than xxhash
-------------------------------------------------------------------------------------------------------------------------------------------------------------------------
2022/08/03 15:28:50 case: replicas:100+hash:crc32.ChecksumIEEE times: 1000000 标准方差: 16188.201024202783  max: 121890  min: 69629 (max-min)/times: 0.052261 peak/mean: 1.2189
2022/08/03 15:28:50 case: replicas:200+hash:crc32.ChecksumIEEE times: 1000000 标准方差: 11440.727826497754  max: 126050  min: 82970 (max-min)/times: 0.04308 peak/mean: 1.2605
2022/08/03 15:28:50 case: replicas:500+hash:crc32.ChecksumIEEE times: 1000000 标准方差: 17259.726985094523  max: 130659  min: 69507 (max-min)/times: 0.061152 peak/mean: 1.30659
2022/08/03 15:28:50 case: replicas:1000+hash:crc32.ChecksumIEEE times: 1000000 标准方差: 21791.261533009052  max: 137256  min: 72892 (max-min)/times: 0.064364 peak/mean: 1.37256
2022/08/03 15:28:51 case: replicas:2000+hash:crc32.ChecksumIEEE times: 1000000 标准方差: 12953.256825987819  max: 120299  min: 73664 (max-min)/times: 0.046635 peak/mean: 1.20299
*/
func Test_ConsistentHash(t *testing.T) {
	args := []arg{
		// murmur3
		{50, hash.Hash, "murmur3.Sum64"},
		{100, hash.Hash, "murmur3.Sum64"},
		{200, hash.Hash, "murmur3.Sum64"},
		{500, hash.Hash, "murmur3.Sum64"},
		{1000, hash.Hash, "murmur3.Sum64"},
		{2000, hash.Hash, "murmur3.Sum64"},
		// xxhash
		{50, xxHashFunc, "xxhash.Sum64"},
		{100, xxHashFunc, "xxhash.Sum64"},
		{200, xxHashFunc, "xxhash.Sum64"},
		{500, xxHashFunc, "xxhash.Sum64"},
		{1000, xxHashFunc, "xxhash.Sum64"},
		{2000, xxHashFunc, "xxhash.Sum64"},
		// xxhash3
		{50, xxHash3Func, "xxhash3.Hash64"},
		{100, xxHash3Func, "xxhash3.Hash64"},
		{200, xxHash3Func, "xxhash3.Hash64"},
		{500, xxHash3Func, "xxhash3.Hash64"},
		{1000, xxHash3Func, "xxhash3.Hash64"},
		{2000, xxHash3Func, "xxhash3.Hash64"},
		// crc32 (trpcgo用的），
		{100, crc32HashFunc, "crc32.ChecksumIEEE"},
		{200, crc32HashFunc, "crc32.ChecksumIEEE"},
		{500, crc32HashFunc, "crc32.ChecksumIEEE"},
		{1000, crc32HashFunc, "crc32.ChecksumIEEE"},
		{2000, crc32HashFunc, "crc32.ChecksumIEEE"},
	}
	for i := 0; i < len(args); i++ {
		doTest(args[i])
	}
}

type arg struct {
	replicas int
	hash     hash.Func
	hashName string
}

func (a arg) String() string {
	return fmt.Sprintf("replicas:%d+hash:%s", a.replicas, a.hashName)
}

func doTest(arg arg) {
	chash := hash.NewCustomConsistentHash(arg.replicas, arg.hash)
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

	}
	// 计算hash到的次数的方差
	data := count[:]
	sdev, _ := stat.Variance(data)
	sdev = math.Pow(sdev, 0.5)
	max, _ := stat.Max(data)
	min, _ := stat.Min(data)
	mean, _ := stat.Mean(data)
	log.Println("case:", arg, "times:", times, "标准方差:", sdev, " max:", max, " min:", min, "(max-min)/times:", float64(max-min)/float64(times), "peak/mean:", max/mean)
	//fmt.Println(count)
}

func xxHashFunc(data []byte) uint64 {
	return xxhash.Sum64(data)
}

func xxHash3Func(data []byte) uint64 {
	return xxh3.Hash(data)
}

func metroHashFunc(data []byte) uint64 {
	return 0
}

func crc32HashFunc(data []byte) uint64 {
	return uint64(crc32.ChecksumIEEE(data))
}
