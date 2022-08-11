package ConsistentHash_trpcgo

import (
	"errors"
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
	"time"
)

var (
	defaultReplicas int  = 100                // the default virtual node coefficient.
	defaultHashFunc Hash = crc32.ChecksumIEEE // uses CRC32 as the default.

	ErrNoServerAvailable = errors.New("no server available")
	ErrInvalidKey        = errors.New("invalid key")
	ErrInvalidNodeList   = errors.New("node list contains nil node")
)

// NewConsistentHash creates a new ConsistentHash.
func NewConsistentHash() *ConsistentHash {
	return &ConsistentHash{
		pickers: new(sync.Map),
	}
}

// ConsistentHash defines the consistent hash.
type ConsistentHash struct {
	pickers  *sync.Map
	interval time.Duration
}

// Select implements loadbalance.LoadBalancer.
func (ch *ConsistentHash) Select(serviceName string, list []*Node, key string, opt ...Option) (*Node, error) {
	opts := &Options{}
	for _, o := range opt {
		o(opts)
	}
	p, ok := ch.pickers.Load(serviceName)
	if ok {
		return p.(*chPicker).Pick(list, key, opts)
	}

	newPicker := &chPicker{
		interval: ch.interval,
	}
	v, ok := ch.pickers.LoadOrStore(serviceName, newPicker)
	if !ok {
		return newPicker.Pick(list, key, opts)
	}
	return v.(*chPicker).Pick(list, key, opts)
}

// chPicker is the picker of the consistent hash.
type chPicker struct {
	list     []*Node
	keys     UInt32Slice      // a hash slice of sorted node list, it's length is #(node)*replica，这个其实就是hash环
	hashMap  map[uint32]*Node // a map which keeps hash-node maps，这个用来跟踪虚节点hash到物理节点的映射关系
	mu       sync.Mutex
	interval time.Duration
}

// Pick picks a node.
func (p *chPicker) Pick(list []*Node, key string, opts *Options) (*Node, error) {
	// 最新节点列表为空，无法pick有效节点
	if len(list) == 0 {
		return nil, ErrNoServerAvailable
	}
	// 如果没有指定一致性hash用的key，返回错误
	if key == "" {
		return nil, ErrInvalidKey
	}
	tmpKeys, tmpMap, err := p.updateState(list, opts.Replicas)
	if err != nil {
		return nil, err
	}
	hash := defaultHashFunc([]byte(key))
	// Find the best matched node by binary search. Node A is better than B if A's hash value is
	// greater than B's.
	idx := sort.Search(len(tmpKeys), func(i int) bool { return tmpKeys[i] >= hash })
	if idx == len(tmpKeys) {
		idx = 0
	}
	node, ok := tmpMap[tmpKeys[idx]]
	if !ok {
		return nil, ErrNoServerAvailable
	}
	return node, nil
}

// updateState 如果节点列表变化，更新一致性hash环，返回最新的hash环相关信息
func (p *chPicker) updateState(list []*Node, replicas int) (UInt32Slice, map[uint32]*Node, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// if node list is the same as last update, there is no need to update hash ring.
	if isNodeSliceEqualBCE(p.list, list) {
		return p.keys, p.hashMap, nil
	}
	actualReplicas := replicas
	if actualReplicas <= 0 {
		actualReplicas = defaultReplicas
	}
	// update node list.
	p.list = list
	p.hashMap = make(map[uint32]*Node)
	p.keys = make(UInt32Slice, len(list)*actualReplicas)
	for i, node := range list {
		if node == nil {
			//不允许有空的node
			return nil, nil, ErrInvalidNodeList
		}
		for j := 0; j < actualReplicas; j++ {
			hash := defaultHashFunc([]byte(strconv.Itoa(j) + node.Address))
			p.keys[i*(actualReplicas)+j] = hash
			p.hashMap[hash] = node
		}
	}
	sort.Sort(p.keys)
	return p.keys, p.hashMap, nil
}
