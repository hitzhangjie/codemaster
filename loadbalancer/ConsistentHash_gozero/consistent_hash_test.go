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

	"github.com/zhenjl/cityhash"

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
		//{50, hash.Hash, "murmur3.Sum64"},
		{100, hash.Hash, "murmur3.Sum64"},
		//{200, hash.Hash, "murmur3.Sum64"},
		//{500, hash.Hash, "murmur3.Sum64"},
		//{1000, hash.Hash, "murmur3.Sum64"},
		//{2000, hash.Hash, "murmur3.Sum64"},

		// xxhash
		//{50, xxHashFunc, "xxhash.Sum64"},
		{100, xxHashFunc, "xxhash.Sum64"},
		//{200, xxHashFunc, "xxhash.Sum64"},
		//{500, xxHashFunc, "xxhash.Sum64"},
		//{1000, xxHashFunc, "xxhash.Sum64"},
		//{2000, xxHashFunc, "xxhash.Sum64"},

		// xxhash3
		//{50, xxHash3Func, "xxhash3.Hash64"},
		{100, xxHash3Func, "xxhash3.Hash64"},
		//{200, xxHash3Func, "xxhash3.Hash64"},
		//{500, xxHash3Func, "xxhash3.Hash64"},
		//{1000, xxHash3Func, "xxhash3.Hash64"},
		//{2000, xxHash3Func, "xxhash3.Hash64"},

		// crc32 (trpcgo用的），
		//{50, crc32HashFunc, "crc32.ChecksumIEEE"},
		{100, crc32HashFunc, "crc32.ChecksumIEEE"},
		//{200, crc32HashFunc, "crc32.ChecksumIEEE"},
		//{500, crc32HashFunc, "crc32.ChecksumIEEE"},
		//{1000, crc32HashFunc, "crc32.ChecksumIEEE"},
		//{2000, crc32HashFunc, "crc32.ChecksumIEEE"},
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

func crc32HashFunc(data []byte) uint64 {
	return uint64(crc32.ChecksumIEEE(data))
}

func cityHashFunc(data []byte) uint64 {
	return cityhash.CityHash64(data, uint32(len(data)))
}

type el struct {
	minHash uint64
	maxHash uint64
	count   uint64
}

type els []el

func (e els) Len() int {
	return len(e)
}

func (e els) Less(i, j int) bool {
	return e[i].minHash < e[i].maxHash
}

func (e els) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func (e els) Init() {
	for i := range e {
		e[i].minHash = uint64(i)
		e[i].maxHash = uint64(i + 1)
	}
}

func newEls(length int) els {
	v := els(make([]el, length))
	v.Init()
	return v
}

var hashes [80 * 10000]uint64

func init() {
	for i := 0; i < len(hashes); i++ {
		//hashes[i] = crc32HashFunc([]byte(strconv.Itoa(i)))
		hashes[i] = cityHashFunc([]byte(strconv.Itoa(i)))
		//hashes[i] = cityHashFunc(generateRandomBytes(18))
	}
}

func generateRandomBytes(n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}

// max: 119 min: 50 sdev: 8.835655040799182 peak/mean: 1.4875
//
// 对hash值hash，max: 122 min: 48 sdev: 8.952306965246445 peak/mean: 1.525
// 直接用hash值, max: 114 min: 51 sdev: 8.658637306181614 peak/mean: 1.425
//
// 对cityhash值再hash：max: 123 min: 47 sdev: 8.982126696946553 peak/mean: 1.5375
func Test_XXH3_数值转字符串再hash_均匀性(t *testing.T) {

	eee := newEls(10000)

	for i := 0; i < 80*10000; i++ {
		//v := xxh3.HashString(strconv.Itoa(i)) % uint64(len(eee))
		v := xxh3.HashString(strconv.Itoa(int(hashes[i]))) % uint64(len(eee))
		eee[v].count++
	}

	v := eee.toFloat64()
	max, _ := stat.Max(v)
	min, _ := stat.Min(v)
	avg, _ := stat.Mean(v)
	sdev, _ := stat.StdDevP(v)
	peakToMean := max / avg

	fmt.Println("max:", max, "min:", min, "sdev:", sdev, "peak/mean:", peakToMean)
	fmt.Println("------------------")
	//eee.prettyprint()
}

//max: 117 min: 51 sdev: 8.978340603920081 peak/mean: 1.4625
//
//max: 118 min: 50 sdev: 9.029075257189962 peak/mean: 1.475
// 对cityhash值再hash：max: 113 min: 49 sdev: 9.043926138574994 peak/mean: 1.4125
func Test_XXH3_数值二次hash_均匀性_v2(t *testing.T) {
	eee := newEls(10000)

	for i := 0; i < 80*10000; i++ {
		//v := xxh3_v2(uint64(i)) % uint64(len(eee))
		v := xxh3_v2(hashes[i]) % uint64(len(eee))
		eee[v].count++
	}

	v := eee.toFloat64()
	max, _ := stat.Max(v)
	min, _ := stat.Min(v)
	avg, _ := stat.Mean(v)
	sdev, _ := stat.StdDevP(v)
	peakToMean := max / avg

	fmt.Println("max:", max, "min:", min, "sdev:", sdev, "peak/mean:", peakToMean)
	fmt.Println("------------------")
	//eee.prettyprint()
}

// xxh3 		Benchmark_xxh3_1-10    	57173592	        19.69 ns/op
//
// murmur		... 35.28 ns/op
// crc32		... 32.29 ns/op
//
// Benchmark_xxh3_数值转字符串后再hash-10                    	41993408	        28.32 ns/op
// 对cityhash值再hash：Benchmark_xxh3_数值转字符串后再hash-10                    	29426978	        39.02 ns/op
func Benchmark_xxh3_数值转字符串后再hash(b *testing.B) {
	for i := 0; i < b.N; i++ {
		//xxh3.Hash([]byte(strconv.Itoa(i)))
		j := i % len(hashes)
		xxh3.Hash([]byte(strconv.Itoa(int(hashes[j]))))
	}
}

//func Benchmark_murmur3hash(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		hash.Hash([]byte(strconv.Itoa(i)))
//	}
//}

// Benchmark_crc32hash-10    	36708412	        32.29 ns/op
//func Benchmark_crc32hash(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		crc32HashFunc([]byte(strconv.Itoa(i)))
//	}
//}

// Benchmark_cityhash-10    	46138761	        22.44 ns/op
//func Benchmark_cityhash(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		cityHashFunc([]byte(strconv.Itoa(i)))
//	}
//}

//Benchmark_xxh3_trick-10    	187119049	         6.324 ns/op
//
// Benchmark_xxh3_数值二次hash_v2-10            	179257455	         6.612 ns/op
// 对cityhash值再hash：         6.613 ns/op
//                    循环展开  3.7ns
func Benchmark_xxh3_数值二次hash_v2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		j := i % len(hashes)
		//xxh3_v2(uint64(i))
		xxh3_v22(hashes[j])
	}
}

func Benchmark_xxh3_数值二次hash_v2_again(b *testing.B) {
	for i := 0; i < b.N; i++ {
		//xxh3_v2(uint64(i))
		xxh3_v22(uint64(i))
	}
}

func xxh3_v2(n uint64) uint64 {
	var vv [8]byte
	for i := 0; i < 8; i++ {
		vv[i] = byte(n & 0xff)
		n = n >> 8
	}
	return xxh3.Hash(vv[:])
}

func xxh3_v22(n uint64) uint64 {
	var vv [8]byte
	vv[0] = byte(n & 0xff)
	vv[1] = byte((n >> 8) & 0xff)
	vv[2] = byte((n >> 16) & 0xff)
	vv[3] = byte((n >> 24) & 0xff)
	vv[4] = byte((n >> 32) & 0xff)
	vv[5] = byte((n >> 40) & 0xff)
	vv[6] = byte((n >> 48) & 0xff)
	vv[7] = byte((n >> 56) & 0xff)
	return xxh3.Hash(vv[:])
}

func (e els) toFloat64() []float64 {
	f := make([]float64, len(e))
	for i, v := range e {
		f[i] = float64(v.count)
	}
	return f
}

func (e els) prettyprint() {
	for _, v := range e {
		fmt.Println(v)
	}
}

// crc32:
// max: 10 min: 4 avg: 9.74122875
// p10: 7 p20: 7 p30: 8 p50: 8 p99: 8 p999: 8
//
// cityhash:
// max: 20 min: 13 avg: 19.3970125
// p10: 17 p20: 17 p30: 17 p50: 17 p99: 18 p999: 18
func Test_hashValue_length(t *testing.T) {
	nnn := make([]float64, len(hashes))
	for i := 0; i < len(nnn); i++ {
		nnn[i] = float64(len(fmt.Sprintf("%d", hashes[i])))
	}
	max, _ := stat.Max(nnn)
	min, _ := stat.Min(nnn)
	avg, _ := stat.Mean(nnn)
	fmt.Println("max:", max, "min:", min, "avg:", avg)

	pp := []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 0.95, 0.99, 0.999, 0.9999, 1.0}
	ppv := make([]float64, len(pp))
	for i := 0; i < len(pp); i++ {
		v, _ := stat.Percentile(nnn, pp[i])
		ppv[i] = v
	}
	for i := 0; i < len(pp); i++ {
		fmt.Printf("p-%f: %v\n", pp[i], ppv[i])
	}
	fmt.Println()
}

func BenchmarkU8(b *testing.B) {
	var bts [8]byte
	var c uint64
	for n := 0; n < b.N; n++ {
		encodeU8(bts[:], c)
		c++
	}
}

func BenchmarkE8(b *testing.B) {
	var bts [8]byte
	var c uint64
	for n := 0; n < b.N; n++ {
		encodeE8(bts[:], c)
		c++
	}
}

func encodeU8(bts []byte, n uint64) {
	for i := 0; i < 8; i++ {
		bts[i] = byte(n & 0xff)
		n = n >> 8
	}
}

func encodeE8(bts []byte, n uint64) {
	bts[0] = byte(n & 0xff)
	bts[1] = byte((n >> 8) & 0xff)
	bts[2] = byte((n >> 16) & 0xff)
	bts[3] = byte((n >> 24) & 0xff)
	bts[4] = byte((n >> 32) & 0xff)
	bts[5] = byte((n >> 40) & 0xff)
	bts[6] = byte((n >> 48) & 0xff)
	bts[7] = byte((n >> 56) & 0xff)
}
