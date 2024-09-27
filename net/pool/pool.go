package pool

import (
	"context"
	"errors"
	"log"
	"net"
	"sync"
	"time"

	"go.uber.org/atomic"
)

const (
	defaultCheckInterval = time.Second * 5
	defaultIdleDuration  = time.Minute * 5
	defaultDialTimeout   = time.Second
)

var (
	ErrConnLeaked       = errors.New("connection leaked")              // 连接池连接泄露
	ErrConnIdle         = errors.New("connection idle")                // 连接空闲
	ErrConnTooMany      = errors.New("connection too many")            // 连接太多了
	ErrConnInitNumLimit = errors.New("open connection: limit by init") // 初始数量限制
	ErrConnMinNumLimit  = errors.New("open connection: limit by min")  // 最小数量限制
)

// pool 连接池，在caller中会为每个被调node维护一个连接池
type pool struct {
	mux   sync.RWMutex
	total atomic.Int32
	conns chan net.Conn // FIXME 现在是个chan，可能导致空闲连接无法释放掉，考虑stack

	network string
	address string

	cfg config
}

func newPool(network, address string, initNum, minNum, maxNum int) *pool {
	return &pool{
		network: network,
		address: address,
		conns:   make(chan net.Conn, maxNum),
		cfg: config{
			initNum: initNum,
			minNum:  minNum,
			maxNum:  maxNum,
		},
	}
}

// init 初始化连接池，当右一个连接可用时通过close(ch)通知调用方可用
func (p *pool) init() <-chan bool {
	if p.cfg.checkInterval == 0 {
		p.cfg.checkInterval = defaultCheckInterval
	}
	if p.cfg.idleBeforeClose == 0 {
		p.cfg.idleBeforeClose = defaultIdleDuration
	}
	go p.check()

	chok := make(chan bool, 1)
	once := new(sync.Once)

	if p.cfg.initNum == 0 {
		close(chok)
		return chok
	}

	for i := 0; i < p.cfg.initNum; i++ {
		go func() {
			conn, err := p.createConn(defaultDialTimeout, limitByInitNum)
			if err == ErrConnInitNumLimit {
				return
			}
			if err != nil {
				log.Println("pool init idle connection:", err)
				return
			}
			p.conns <- conn
			once.Do(func() {
				close(chok)
			})
		}()
	}

	return chok
}

// check 连接池中连接健康检查（注意取出、放回连接时不要更新使用时间）
func (p *pool) check() {
	b := make([]byte, 1)
	for {
	NextConn:
		for i := 0; i < p.cfg.maxNum; i++ {
			select {
			case c := <-p.conns:
				conn, ok := c.(*connection)
				if !ok {
					panic("shouldn't reach here")
				}
				// 先检查下这个链接是不是空闲很久了
				if time.Since(conn.released) >= p.cfg.idleBeforeClose {
					//log.Println("conn:", c.LocalAddr().String(), "idle")
					conn.err = ErrConnIdle
				} else {
					//log.Println("conn:", c.LocalAddr().String(), "not idle")
					_ = c.SetReadDeadline(time.Now().Add(time.Millisecond))
					_, err := c.Read(b)
					if err == nil {
						conn.err = ErrConnLeaked
					} else {
						conn.err = err
					}
				}
				if err := p.put(c); err != nil {
					log.Println("pool put:", err)
				}
			default:
				break NextConn
			}
		}

		// 如果当前可用连接数，已经比min少了，提前新建几个连接
		if d := p.cfg.minNum - int(p.total.Load()); d > 0 {
			for i := 0; i < d; i++ {
				c, err := p.createConn(defaultDialTimeout, limitByMinNum)
				if err != nil {
					break
				}
				_ = p.put(c)
			}
		}
		time.Sleep(p.cfg.checkInterval)
	}
}

func (p *pool) get(ctx context.Context) (conn net.Conn, err error) {
	defer func() {
		if err == nil {
			conn.(*connection).used = time.Now()
		}
	}()

	// fastpath：spin尝试取一个出来
	conn, err = p.fastget(ctx)
	if err != nil {
		return nil, err
	}
	if conn != nil {
		return conn, nil
	}

	// slowpath: 检查是否达到连接数上限
	t := defaultDialTimeout
	d, ok := ctx.Deadline()
	if ok {
		t = time.Until(d)
	}

	conn, err = p.createConn(t, limitByMaxNum)
	if err != nil {
		return nil, err
	}
	// 转换成我们定义的Conn，该Conn会重写包装下net.Conn接口，将read/write等过程中的错误记录在内部，
	// 后续Put的时候会检查此err，以判断连接是否仍然可以复用
	return conn, nil
}

const (
	limitByInitNum int = iota
	limitByMinNum
	limitByMaxNum
)

func (p *pool) createConn(timeout time.Duration, limitKind int) (net.Conn, error) {
	p.mux.Lock()
	defer p.mux.Unlock()

	switch limitKind {
	case limitByInitNum:
		total := p.total.Load()
		if total >= int32(p.cfg.initNum) {
			return nil, ErrConnInitNumLimit
		}
	case limitByMinNum:
		total := p.total.Load()
		if total >= int32(p.cfg.minNum) {
			return nil, ErrConnMinNumLimit
		}
	case limitByMaxNum:
		total := p.total.Load()
		if total >= int32(p.cfg.maxNum) {
			return nil, ErrConnTooMany
		}
	default:
	}

	conn, err := net.DialTimeout(p.network, p.address, timeout)
	if err != nil {
		return nil, err
	}
	tc := conn.(*net.TCPConn)
	_ = tc.SetKeepAlive(true)
	_ = tc.SetKeepAlivePeriod(time.Second * 10)
	p.total.Add(1)

	now := time.Now()
	return &connection{Conn: conn, pool: p, used: now, released: now}, err
}

func (p *pool) fastget(ctx context.Context) (net.Conn, error) {
	for i := 0; i < 10; i++ {
		select {
		case c := <-p.conns:
			return c, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			// FIXME 有可能会导致scheduler调度非常频繁，不一定是好事，需要看下trace
			time.Sleep(time.Microsecond * 10)
		}
	}
	return nil, nil
}

func (p *pool) put(conn net.Conn) error {
	c, ok := conn.(*connection)
	if !ok {
		// 不是我们管理的连接
		return conn.Close()
	}

	if c.err == nil {
		goto PUT
	}

	// 检查错误并关闭不可复用的连接
	if c.err == ErrConnLeaked {
		//log.Println("conn:", c.LocalAddr().String(), "is released (leaked)")
		p.total.Add(-1)
		return c.Conn.Close()
	}

	if c.err == ErrConnIdle {
		var closed bool
		var err error
		p.mux.Lock()
		if p.total.Load() > int32(p.cfg.initNum) {
			//log.Println("conn:", c.LocalAddr().String(), "is released (idle)")
			p.total.Add(-1)
			err = c.Conn.Close()
			closed = true
		}
		p.mux.Unlock()
		if closed {
			return err
		}
		goto PUT
	}

	if ne, ok := c.err.(net.Error); !ok || !ne.Timeout() {
		p.total.Add(-1)
		return c.Conn.Close()
	}

	// 把可复用的连接放回池子
	c.err = nil
PUT:
	select {
	case p.conns <- conn:
		//log.Println("conn:", c.LocalAddr().String(), "is reused")
		return nil
	default:
		var err error
		//log.Println("conn:", c.LocalAddr().String(), "is released (full)")
		p.mux.Lock()
		p.total.Add(-1)
		err = c.Conn.Close()
		p.mux.Unlock()
		return err
	}
}
