package main

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func Test_etcd_register(t *testing.T) {
	// init etcdclient
	client, err := clientv3.New(clientv3.Config{
		Endpoints:         []string{"http://localhost:2379"},
		DialTimeout:       time.Second,
		DialKeepAliveTime: time.Second,
	})
	require.Nil(t, err)

	// create a lease
	lease, err := client.Grant(context.Background(), 10)
	if err != nil {
		panic(err)
	}
	t.Logf("lease created: lease=%v, ttl=%d, revision=%d", lease.ID, lease.TTL, lease.Revision)

	// start keepalive
	go keepalive(t, client, lease.ID)

	// provided some service instances started and register into etcd registry
	prepareServiceInstances(t, client, lease.ID, 10000)

	// provided every service to fetch the instances list from registry
	startServiceInstances(t, client, lease.ID, 1)

	mu.Lock()
	num := len(set)
	mu.Unlock()

	t.Logf(fmt.Sprintf("found %d nodes", num))
}

func keepalive(t *testing.T, client *clientv3.Client, lease clientv3.LeaseID) {
	ch, err := client.KeepAlive(context.Background(), lease)
	if err != nil {
		panic(err)
	}

	for {
		_, sentBeforeClosed := <-ch
		if !sentBeforeClosed {
			break
		}
	}

	t.Logf("etcd keep alive failed, lease:%v", lease)
}

func prepareServiceInstances(t *testing.T, client *clientv3.Client, lease clientv3.LeaseID, total int) {
	opts := []clientv3.OpOption{
		clientv3.WithLease(lease),
	}

	wg := sync.WaitGroup{}
	wg.Add(total)
	begin := time.Now()
	for i := 1; i <= total; i++ {
		go func() {
			defer wg.Done()
			key := fmt.Sprintf("/myname/%d", i)
			val := fmt.Sprintf("/myvalue/%d", i)
			rsp, err := client.Put(context.Background(), key, val, opts...)
			if err != nil {
				t.Errorf("put kvpair failed: err=%v", err)
			}
			t.Logf("put kvpair ok: index=%d, key=%s, revision=%+v", i, key, rsp.Header.Revision)
		}()
	}
	wg.Wait()
	t.Logf("put %d kvpairs ... %v", total, time.Since(begin))
}

var mu sync.Mutex
var set = map[string]struct{}{}

func startServiceInstances(t *testing.T, client *clientv3.Client, lease clientv3.LeaseID, total int) {
	wg := sync.WaitGroup{}
	for dsaInstanceIdx := 1; dsaInstanceIdx <= total; dsaInstanceIdx++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// get others
			revision, err := getbyprefix(t, client, "/")
			if err != nil {
				t.Logf("getbyprefix failed: err=%v, revision=%d", err, revision)
				return
			}

			// watch others since revision
			go func() {
				ch := client.Watch(context.TODO(), "/myname",
					clientv3.WithPrefix(),
					clientv3.WithRev(revision+1))
				for msg := range ch {
					for _, evt := range msg.Events {
						t.Logf("watchevent: type=%v key=%s val=%s", evt.Type, evt.Kv.Key, evt.Kv.Value)

						mu.Lock()
						set[string(evt.Kv.Key)] = struct{}{}
						mu.Unlock()
					}
				}
			}()

			// register itself
			myself := fmt.Sprintf("/myself-%d", dsaInstanceIdx)
			_, err = client.Put(context.Background(), myself, "value", clientv3.WithLease(lease))
			if err != nil {
				t.Logf("register myself fail: key=%s, err=%v", myself, err)
				time.Sleep(time.Second / 10000)
			}
		}()
	}
	wg.Wait()
}

func getbyprefix(t *testing.T, client *clientv3.Client, key string) (revision int64, err error) {
	var pageno = 1
	var pagesize = 100
	var fetchedKeys = 0

	opts := []clientv3.OpOption{
		// we specify [key, getprefixrangeend(key)), so comment this line
		// clientv3.WithPrefix(),
		clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend),
		clientv3.WithRange(clientv3.GetPrefixRangeEnd(key)),
		clientv3.WithLimit(int64(pagesize)),
		// we need current revision, so comment this line, when fetching
		// following pages, we need to specify the revision there
		// clientv3.WithRev(x)
	}

	for {
		rsp, err := client.Get(context.Background(), key, opts...)
		if err != nil {
			t.Logf("Failed to get key-value pair in etcd, error: %v", err)
			continue
		}

		// record current revision
		if revision == 0 {
			revision = rsp.Header.Revision
			opts = append(opts, clientv3.WithRev(rsp.Header.Revision))
		}

		// record next range start
		if len(rsp.Kvs) == 0 {
			break
		}
		last := string(rsp.Kvs[len(rsp.Kvs)-1].Key)

		// record fetched kvpairs
		for i, kv := range rsp.Kvs {
			if pageno != 1 && i == 0 {
				continue
			}

			fetchedKeys++
			t.Logf("getbyprefix: range=[%s ~ %s] page=%d key=%s fetchedKeys=%d", key, last, pageno, kv.Key, fetchedKeys)

			mu.Lock()
			set[string(kv.Key)] = struct{}{}
			mu.Unlock()
		}

		// if there's no more pages
		if !rsp.More {
			break
		}

		// prepare query param for next page
		key = last

		pageno++
		time.Sleep(time.Millisecond * 100)
	}
	return revision, nil
}
