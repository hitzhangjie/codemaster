package pool

import (
	"context"
	"net"
	"sync"
	"time"
)

type config struct {
	initNum int // 初始连接数
	minNum  int // 保持最少min个连接
	maxNum  int // 最大连接数

	checkInterval   time.Duration // 连接池检查间隔
	idleBeforeClose time.Duration // 连接最大空闲时间
}

// Manager 连接池管理器，应用程序初始化一个实例即可，它负责维护所有callee的连接池
type Manager struct {
	pools *sync.Map // k=calleeAddress, v=*Pool
	cfg   config
}

// New 创建一个连接池管理器
func New(init, min, max int, checkInterval, idleBeforeClose time.Duration) *Manager {
	return &Manager{
		pools: new(sync.Map),
		cfg: config{
			initNum:         init,
			minNum:          min,
			maxNum:          max,
			checkInterval:   checkInterval,
			idleBeforeClose: idleBeforeClose,
		},
	}
}

// Get return a tcpconn for use
func (pm *Manager) Get(ctx context.Context, network, address string) (net.Conn, error) {
	v, ok := pm.pools.Load(network + ":" + address)
	if !ok {
		p := &pool{
			network: network,
			address: address,
			conns:   make(chan net.Conn, pm.cfg.maxNum),
			cfg:     pm.cfg,
		}
		v, ok = pm.pools.LoadOrStore(address, p)
		if !ok {
			// 当有一个连接可用时就可以返回
			<-p.init()
		}
	}
	return v.(*pool).get(ctx)
}
