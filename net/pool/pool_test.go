package pool

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPool(t *testing.T) {
	t.Run("连接数：初始-1，最少-0，最大-1", func(t *testing.T) {
		init := 1
		min := 0
		max := 1
		p := &pool{
			conns:   make(chan net.Conn, max),
			network: "tcp",
			address: serverAddr,
			cfg: config{
				initNum: init,
				minNum:  min,
				maxNum:  max,
			},
		}
		<-p.init()
		assert.Equal(t, int32(1), p.total.Load())

		// 尝试获取这个连接，不放回池子
		c, err := p.get(context.TODO())
		assert.Nil(t, err)
		assert.NotNil(t, c)
		usedTime := c.(*connection).used

		// 再次获取1个连接，因为上限是1，所以应该报错
		c2, err := p.get(context.TODO())
		assert.Equal(t, ErrConnTooMany, err)
		assert.Nil(t, c2)

		// 放回这个连接，可以重新获取到
		_ = p.put(c)
		c2, err = p.get(context.TODO())
		usedTime2 := c2.(*connection).used
		assert.Nil(t, err)
		assert.Equal(t, c.RemoteAddr(), c2.RemoteAddr()) // 远程地址相同
		assert.Equal(t, c.LocalAddr(), c2.LocalAddr())   // 本地地址相同，因为只有1个连接
		assert.Less(t, usedTime, usedTime2)              // 使用时间应该更新
	})

	t.Run("连接数：初始-1，最少-0，最大-2", func(t *testing.T) {
		init := 1
		min := 0
		max := 2
		p := &pool{
			conns:   make(chan net.Conn, max),
			network: "tcp",
			address: serverAddr,
			cfg: config{
				initNum:         init,
				minNum:          min,
				maxNum:          max,
				checkInterval:   time.Second,
				idleBeforeClose: time.Second * 2,
			},
		}
		<-p.init()
		assert.Equal(t, int32(init), p.total.Load())

		// 应该获取成功
		c, err := p.get(context.TODO())
		assert.Nil(t, err)
		assert.NotNil(t, c)
		assert.Equal(t, int32(init), p.total.Load())

		// 会新建一个并成功
		c2, err := p.get(context.TODO())
		assert.Nil(t, err)
		assert.NotNil(t, c2)
		assert.Equal(t, int32(init+1), p.total.Load())

		// 再获取就超过上限，应该失败
		c3, err := p.get(context.TODO())
		assert.Equal(t, ErrConnTooMany, err)
		assert.Nil(t, c3)

		_ = p.put(c)
		_ = p.put(c2)

		time.Sleep(p.cfg.idleBeforeClose * 2)
		assert.Equal(t, int32(1), p.total.Load())
	})
}

func TestPool_HandleErrors(t *testing.T) {
	min := 2
	max := 4
	p := &pool{
		conns:   make(chan net.Conn, max),
		network: "tcp",
		address: serverAddr,
		cfg: config{
			initNum:       min,
			minNum:        min,
			maxNum:        max,
			checkInterval: time.Second * 2,
		},
	}
	<-p.init()

	// p.init初始化完1个连接后就表示池子有连接可用，return
	// 这里sleep 1s等待所有的2个备用连接初始化完成
	time.Sleep(time.Second)
	total := p.total.Load()
	assert.Equal(t, min, int(total))

	// handle nil
	c, err := p.get(context.TODO())
	assert.Nil(t, err)
	assert.Nil(t, p.put(c))
	assert.Equal(t, total, p.total.Load())

	// handle timeout
	c, err = p.get(context.TODO())
	assert.Nil(t, err)
	c.(*connection).err = context.DeadlineExceeded
	assert.Nil(t, p.put(c))
	assert.Equal(t, total, p.total.Load())

	// handle eof
	c, err = p.get(context.TODO())
	assert.Nil(t, err)
	c.(*connection).err = io.EOF
	assert.Nil(t, p.put(c))
	assert.Equal(t, total-1, p.total.Load())

	// handle leaked
	c, err = p.get(context.TODO())
	assert.Nil(t, err)
	c.(*connection).err = ErrConnLeaked
	assert.Nil(t, p.put(c))
	assert.Equal(t, total-2, p.total.Load())

	// 前面关掉了几个连接，测试check补足可用连接到min
	assert.Less(t, int(p.total.Load()), min)
	time.Sleep(p.cfg.checkInterval * 2) // 确保check至少执行过一次
	assert.Equal(t, total, p.total.Load())
}
