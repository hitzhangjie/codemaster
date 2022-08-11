package ConsistentHash_trpcgo

import (
	"fmt"
	"time"
)

// Hash is the hash function type.
type Hash func(data []byte) uint32

// Node is the information of a node.
type Node struct {
	ServiceName   string        // 服务名
	ContainerName string        // 容器名
	Address       string        // 目标地址 ip:port
	Network       string        // 网络层协议 tcp/udp
	Protocol      string        // 业务协议 trpc/http
	SetName       string        // 节点Set名
	Weight        int           // 权重
	CostTime      time.Duration // 当次请求耗时
	EnvKey        string        // 透传的环境信息
	Metadata      map[string]interface{}
}

// String returns an abbreviation information of node.
func (n *Node) String() string {
	return fmt.Sprintf("service:%s, addr:%s, cost:%s", n.ServiceName, n.Address, n.CostTime)
}

// isNodeSliceEqualBCE check whether two node list is equal by BCE.
func isNodeSliceEqualBCE(a, b []*Node) bool {
	if len(a) != len(b) {
		return false
	}
	if (a == nil) != (b == nil) {
		return false
	}
	b = b[:len(a)]
	for i, v := range a {
		if (v == nil) != (b[i] == nil) {
			return false
		}
		if v.Address != b[i].Address {
			return false
		}
	}
	return true
}

// UInt32Slice defines uint32 slice.
type UInt32Slice []uint32

// Len returns the length of the slice.
func (s UInt32Slice) Len() int {
	return len(s)
}

// Less returns whether the value at i is less than j.
func (s UInt32Slice) Less(i, j int) bool {
	return s[i] < s[j]
}

// Swap swaps values between i and j.
func (s UInt32Slice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
