package pool

import (
	"net"
	"time"
)

// connection 连接池pool直接管理的对象
//
// WARN: 从pool拿到、放回的连接类型是*connection，非net.Conn
type connection struct {
	net.Conn

	used     time.Time // 上次获取的时间
	released time.Time // 用完放回的时间（通过这个可以算出get/put之间的用时，进而监控使用异常)

	err error // 上一次连接上发生的错误，连接池pool根据不同错误类型进行不同处理

	pool *pool // 连接锁归属的池子
}

// Read 读取数据到b，返回读取的字节数、错误
func (c *connection) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	c.err = err
	return
}

// Write 写数据到连接，返回写入的字节数、错误
func (c *connection) Write(b []byte) (n int, err error) {
	n, err = c.Conn.Write(b)
	c.err = err
	return
}

// Close 使用方conn.Close()关闭连接时交给pool处理
func (c *connection) Close() error {
	// 更新下连接最后使用的时间
	c.released = time.Now()
	return c.pool.put(c)
}
