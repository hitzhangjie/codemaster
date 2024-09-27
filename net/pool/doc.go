/*
Package pool 实现了连接池管理

pool.Manager，连接池管理器，它为每个后端服务维护了一个连接池pool.pool，
pool.pool，连接池，它维护了一个远程地址对应的一组连接pool.connection，
pool.connection，虚拟连接，它嵌入了一个net.Conn来完成真正的Read, Write操作，

对于Close操作我们对齐进行了重写，以让连接的关闭操作由连接池接管

连接池会根据连接上发生的错误类型、连接池当前状态来决定是关闭还是复用连接。

使用时可以通过:

pm := pool.New(...)
c, err := pm.Get(ctx, net, address)

	if err != nil {
	  ...
	}

defer c.Close()

//do something about reading/writing
...
*/
package pool
