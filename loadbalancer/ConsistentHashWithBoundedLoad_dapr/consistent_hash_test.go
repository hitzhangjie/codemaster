/*
Copyright 2021 The Dapr Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package hashing

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"strconv"
	"testing"

	stat "github.com/montanaflynn/stats"
	"github.com/stretchr/testify/assert"
)

var nodes = []string{"node1", "node2", "node3", "node4", "node5"}

func TestReplicationFactor(t *testing.T) {
	keys := []string{}
	for i := 0; i < 100; i++ {
		keys = append(keys, fmt.Sprint(i))
	}

	t.Run("varying replication factors, no movement", func(t *testing.T) {
		factors := []int{1, 100, 1000, 10000}

		for _, f := range factors {
			SetReplicationFactor(f)

			h := NewConsistentHash()
			for _, n := range nodes {
				s := h.Add(n, n, 1)
				assert.False(t, s)
			}

			k1 := map[string]string{}

			for _, k := range keys {
				h, err := h.Get(k)
				assert.NoError(t, err)

				k1[k] = h
			}

			nodeToRemove := "node3"
			h.Remove(nodeToRemove)

			for _, k := range keys {
				h, err := h.Get(k)
				assert.NoError(t, err)

				orgS := k1[k]
				if orgS != nodeToRemove {
					assert.Equal(t, h, orgS)
				}
			}
		}
	})
}

func TestSetReplicationFactor(t *testing.T) {
	f := 10
	SetReplicationFactor(f)

	assert.Equal(t, f, replicationFactor)
}

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

// 这个还可以，dapr中默认虚节点数是用的1000
// 测试后发现，100w请求，标准差2k，最大、最小负载占总负载比重为5/1000，
/*
2022/08/01 17:59:20 times: 10000 标准方差: 43.289721643826724  max: 1053  min: 928 (max-min)/times: 0.0125
2022/08/01 17:59:20 times: 20000 标准方差: 62.03386172083759  max: 2077  min: 1883 (max-min)/times: 0.0097
2022/08/01 17:59:20 times: 30000 标准方差: 77.66852644411377  max: 3118  min: 2890 (max-min)/times: 0.0076
2022/08/01 17:59:20 times: 40000 标准方差: 109.79617479675693  max: 4161  min: 3836 (max-min)/times: 0.008125
2022/08/01 17:59:20 times: 50000 标准方差: 148.63983315383533  max: 5235  min: 4778 (max-min)/times: 0.00914
2022/08/01 17:59:20 times: 60000 标准方差: 171.41761869772895  max: 6280  min: 5732 (max-min)/times: 0.009133333333333334
2022/08/01 17:59:20 times: 70000 标准方差: 209.25200118517387  max: 7299  min: 6681 (max-min)/times: 0.008828571428571429
...
2022/08/01 17:59:20 times: 930000 标准方差: 2121.71642780085  max: 96068  min: 90723 (max-min)/times: 0.005747311827956989
2022/08/01 17:59:20 times: 940000 标准方差: 2129.4827071380505  max: 97061  min: 91756 (max-min)/times: 0.005643617021276596
2022/08/01 17:59:20 times: 950000 标准方差: 2157.690153845079  max: 98129  min: 92758 (max-min)/times: 0.005653684210526316
2022/08/01 17:59:20 times: 960000 标准方差: 2181.550182782876  max: 99089  min: 93719 (max-min)/times: 0.00559375
2022/08/01 17:59:20 times: 970000 标准方差: 2204.672084460635  max: 100143  min: 94690 (max-min)/times: 0.0056216494845360825
2022/08/01 17:59:20 times: 980000 标准方差: 2232.6501293306123  max: 101279  min: 95640 (max-min)/times: 0.005754081632653061
2022/08/01 17:59:20 times: 990000 标准方差: 2242.8319152357362  max: 102292  min: 96643 (max-min)/times: 0.005706060606060606
2022/08/01 17:59:20 times: 1000000 标准方差: 2258.7251714186036  max: 103348  min: 97587 (max-min)/times: 0.005761
*/
func Test_ConsistentHashWithBoundedLoad(t *testing.T) {
	SetReplicationFactor(1000)

	chash := NewConsistentHash()
	for i := 0; i < len(hosts); i++ {
		chash.Add(hosts[i], "", 0)
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
