go-zero中一致性hash实现源码阅读

## 定义

```go
type ConsistentHash struct {
hashFunc Func
replicas int
keys     []uint64
ring     map[uint64][]interface{}
nodes    map[string]lang.PlaceholderType
lock     sync.RWMutex
}
```

- hashFunc是自定义的hash函数
- replicas表示每个物理节点对应的虚节点数量
- keys其实就是一致性hash环的表示，记录了虚节点对应的hash值，有序
- ring其实是hash值到一组虚节点的映射，它其实是为了解决hash冲突来的
    - 准确地说，keys+ring构成了一致性hash环，查找hash(key)对应的虚节点时，先在keys中找到>=hash(key)的虚节点对应的hash值，
    - 然后，通过hash值到ring中找对应的虚节点
    - 因为可能有冲突，所以map[k]v这里的v是一个slice
- nodes中记录了当前添加了哪些物理节点，但是这里的map[k]v，k是节点描述信息，可以简单理解成node.String()，v是struct{}

这个一致性hash的设计还是不错的。

## Add/AddWithReplicas/AddWithWeight

添加新节点的时候，大致就是这几个函数，逻辑是什么呢？

- 先获取待添加节点的一个描述信息，如String()，然后记录到nodes中，表示记录了这个节点
- 然后呢，根据设置的虚节点数量，for循环，每次在描述信息后面添加数字后缀，计算hash，然后记录到keys、ring里面
  前面提过了，keys+ring共同构成了一致性hash环

## Get

根据key获取节点的时候呢？

- 先计算key的hash，然后从keys中找第一个>=hash(key)的位置，这个位置对应的hash即为候选虚节点的hash，
- 然后通过虚节点的hash去ring里面找，ring里面是个slice，是为了解决散列冲突的，
- 怎么从这个slice中取呢，重新hash一次，从这个slice里面选一个

## 继续优化

ring-based一致性hash的均匀性还可以继续优化，比如从hash函数的选择方面，虚节点数量的选择方面。

以下是10个物理节点，均匀的100w playerid，在不同虚节点数量、hash函数选择情况下的测试情况：

```bash
case: replicas:50+hash:murmur3.Sum64 标准方差: 14628.560790453721  max: 126737  min: 76395 (max-min)/times: 0.050342 peak/mean: 1.26737
case: replicas:100+hash:murmur3.Sum64 标准方差: 14555.295022774357  max: 127129  min: 76438 (max-min)/times: 0.050691 peak/mean: 1.27129 *
case: replicas:200+hash:murmur3.Sum64 标准方差: 6902.00454940447  max: 110178  min: 85121 (max-min)/times: 0.025057 peak/mean: 1.10178
case: replicas:500+hash:murmur3.Sum64 标准方差: 2285.3205902017335  max: 105277  min: 97136 (max-min)/times: 0.008141 peak/mean: 1.05277
case: replicas:1000+hash:murmur3.Sum64 标准方差: 2069.765928794848  max: 104603  min: 97606 (max-min)/times: 0.006997 peak/mean: 1.04603
case: replicas:2000+hash:murmur3.Sum64 标准方差: 2618.900303562547  max: 104628  min: 94870 (max-min)/times: 0.009758 peak/mean: 1.04628

case: replicas:50+hash:xxhash.Sum64 标准方差: 8627.559643375409  max: 119229  min: 91110 (max-min)/times: 0.028119 peak/mean: 1.19229
case: replicas:100+hash:xxhash.Sum64 标准方差: 8918.29840272235  max: 120236  min: 90692 (max-min)/times: 0.029544 peak/mean: 1.20236 *
case: replicas:200+hash:xxhash.Sum64 标准方差: 5913.828556865679  max: 111947  min: 89811 (max-min)/times: 0.022136 peak/mean: 1.11947
case: replicas:500+hash:xxhash.Sum64 标准方差: 4256.551350565384  max: 107631  min: 93326 (max-min)/times: 0.014305 peak/mean: 1.07631
case: replicas:1000+hash:xxhash.Sum64 标准方差: 3148.5766943176086  max: 106134  min: 95150 (max-min)/times: 0.010984 peak/mean: 1.06134
case: replicas:2000+hash:xxhash.Sum64 标准方差: 1664.1786562746202  max: 103375  min: 96885 (max-min)/times: 0.00649 peak/mean: 1.03375

case: replicas:100+hash:crc32.ChecksumIEEE 标准方差: 16188.201024202783  max: 121890  min: 69629 (max-min)/times: 0.052261 peak/mean: 1.2189
case: replicas:200+hash:crc32.ChecksumIEEE 标准方差: 11440.727826497754  max: 126050  min: 82970 (max-min)/times: 0.04308 peak/mean: 1.2605 *
case: replicas:500+hash:crc32.ChecksumIEEE 标准方差: 17259.726985094523  max: 130659  min: 69507 (max-min)/times: 0.061152 peak/mean: 1.30659
case: replicas:1000+hash:crc32.ChecksumIEEE 标准方差: 21791.261533009052  max: 137256  min: 72892 (max-min)/times: 0.064364 peak/mean: 1.37256
case: replicas:2000+hash:crc32.ChecksumIEEE 标准方差: 12953.256825987819  max: 120299  min: 73664 (max-min)/times: 0.046635 peak/mean: 1.20299
```

不难看出： xxhash的均匀性首先比较好，在100个虚节点（这个一般是比较常用的经验值）时，最大最小负载偏差2.9%，peak/mean比为1.20
