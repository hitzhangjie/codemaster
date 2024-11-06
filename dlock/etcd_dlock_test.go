package dlock_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

// 假定我们先启动一个etcd实例，其监听地址为127.0.0.1:62379
func TestEtcdDistLock(t *testing.T) {
	t.Skip()

	registry := clientv3.Config{
		Endpoints:   []string{"127.0.0.1:62379"},
		DialTimeout: time.Second * 2,
	}

	t.Run("create lock", func(t *testing.T) {
		_, _, mu, err := NewEtcdMutex(registry, randomLockName())
		require.Nil(t, err)
		require.NotNil(t, mu)
	})

	t.Run("lock&unlock", func(t *testing.T) {
		_, _, mu, err := NewEtcdMutex(registry, randomLockName())
		require.Nil(t, err)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second*2)
		defer cancel()

		err = mu.Lock(ctx)
		require.Nil(t, err)

		// doSomething
		time.Sleep(time.Second)

		err = mu.Unlock(ctx)
		require.Nil(t, err)
	})

	t.Run("lock is reenatable with same leaseID", func(t *testing.T) {
		_, _, mu, err := NewEtcdMutex(registry, randomLockName())
		require.Nil(t, err)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		err = mu.Lock(ctx)
		require.Nil(t, err)

		err = mu.Lock(ctx)
		require.Nil(t, err)

		err = mu.TryLock(ctx)
		require.Nil(t, err)
	})

	t.Run("multi-workers: different leaseID", func(t *testing.T) {
		ch := make(chan int)
		chdone := make(chan int)

		begin := time.Now()
		// lockName := randomLockName()
		lockName := "/whatthefuck"

		go func() {
			_, _, mu, err := NewEtcdMutex(registry, lockName)
			require.Nil(t, err)

			ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
			defer cancel()

			err = mu.Lock(ctx)
			require.Nil(t, err)

			time.Sleep(time.Second * 3)
			ch <- 1
			time.Sleep(time.Second * 3)

			err = mu.Unlock(ctx) // ctx肯定用光了呀
			require.NotNil(t, err)

			err = mu.Unlock(context.TODO()) // 释放掉锁
			require.Nil(t, err)
		}()

		go func() {

			<-ch

			_, _, mu, err := NewEtcdMutex(registry, lockName)
			require.Nil(t, err)

			// g1锁持有5s后，g2开始尝试加锁，由于g1还会额外持有5s才会释放，
			// 所以g2 trylock会理解失败
			// 然后g2 lock也会阻塞到超时返回失败
			ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
			defer cancel()

			err = mu.TryLock(ctx)
			require.NotNil(t, err)

			err = mu.Lock(ctx)
			require.NotNil(t, err)

			// g1锁持有时间还有1s，之后才会释放锁，g2才能拿到锁
			err = mu.Lock(context.TODO())
			require.GreaterOrEqual(t, time.Since(begin), time.Second*6)

			// ctx这里控制的是etcdclient和etcdserver之间的rpc超时
			err = mu.Unlock(context.TODO())
			require.Nil(t, err)

			chdone <- 1
		}()

		<-chdone
	})
}

func randomLockName() string {
	lock := uuid.New().String()
	return fmt.Sprintf("/mylock/%s", lock)
}

func NewEtcdMutex(registry clientv3.Config, lock string) (
	client *clientv3.Client,
	session *concurrency.Session,
	mu *concurrency.Mutex,
	err error) {

	client, err = clientv3.New(registry)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			client.Close()
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	lease, err := client.Grant(ctx, 10)
	if err != nil {
		return
	}

	// keepalive
	session, err = concurrency.NewSession(client, concurrency.WithLease(lease.ID))
	if err != nil {
		return
	}

	// add distributed lock
	mu = concurrency.NewMutex(session, lock)
	return
}
