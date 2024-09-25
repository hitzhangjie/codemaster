package main

import (
	"flag"
	"net/http"
	"runtime"
	"time"
)

// 对下面几种Go GC设置进行分别测试，验证那种方案最合适
//
// 固定压力源：gocannon -d 30s -c 64 -t 1s -m hist http://127.0.0.1:8899/
//
// go1.16.5 gogc ballast               qps        cpu  mem                p90/p99/p999
// -----------------------------------------------------------------------------------------------------
// -        100  no                    339641.0        4->4->1m           1342微秒         => GC次数6677次
// -        100  1gb                   402726.7        1999->1999->1025m  533微秒          => GC次数19次
//
// go1.19   gogc 定时gc ballast  gomemlimit   qps             cpu  mem                p90/p99/p999
// -----------------------------------------------------------------------------------------------------
// -        100  no     no       no                   298438.1        3->3->1            1486微秒         => GC次数7240次，默认配置性能比go1.16.5还低也可以解释
// -        100  no     1gb      no           399216.7        2047->2047->1025m  561微秒          => GC次数23次，因为开了ballast，GC次数差别不大，性能很接近
// -        off  no     no       1gb          400325.2        934->934->1m       567微秒          => GC次数18次，观察到从heap 947m开始触发GC
// -        off  no     no       2gb          398762.4        1885->1885->1m     560微秒          => GC次数9次，单纯增大了GOMEMLIMIT降低了GC频率，物理内存占用也真的可以飙到2GB，此时GOGC=off如果混部可能对其他服务产生影响
// -        off  10s    no       2gb          398978.8        ...                572微秒          => GC除了达到nextGC goal触发，也多了forced GC，forced GC保证了RSS占用不会飚的很高，算是一种兜底
// -        off  15s    no       2gb          396663.1        ...                ...             ...无明显效果...
// -        100  no     1gb      2gb          398565.6        ...                ...             => 比较稳妥，正常如果把GOMEMLIMIT当做资源上的软限制，可以根据机器分配的资源来设置，比如容器分配了4G，我们可以用4G*70%=2.8G (uber的经验值)
// -        100  no     1gb      2.8gb        401366.6        ...                558微秒          => 效率还可以，但是我们不想用压舱石方案了，因为压舱石的初始化依赖于初始化的时机，容易复用dirty pages而变成真正的占用物理内存
//
// 我们想要什么：
// - 内存占用少时，不要频繁GC，所以引入了压舱石
// - 内存占用多时，希望能通过GC回收内存，GOGC=100
// - 内存使用有资源限制时，不要达到或者逼近这个限制，容易被OOM Kill，GOMEMLIMIT (or uber gctuner)
//
// 升级到go1.19之后，
// - GOMEMLIMIT可以起到延缓GC的作用，或者为了避免达到占用上限更激进地去GC、scavenge释放内存，但是这是解决资源占用高的问题
//
// - 单纯地将GOMEMLIMIT+GOGC=off组合来取代ballast的方案不靠谱，意味着你的程序很可能在资源临界点附近才开始GC：
//   - 对于混部的服务不友好（因为极容易影响到其他服务），
//   - 如果不混部貌似可行（将GOMEMLIMIT设置的比资源限制值稍低些，如设置为容器分配内存的70%，uber经验值）
//
// - 对于混部的服务，有什么好办法解决呢？根据前面的测试：
//   - 似乎可以通过定时地手动GC来强制清理一波，性能差不多，需要在南京压测环境真实地去测试一下
//   - 另外结合GOGC=100+ballast+GOMEMLIMIT的方案也可以，混部情况下就通过几种方案的组合，不混部的话则可以只依赖GOMEMLIMIT
//
// 目前的一点想法 :)
var (
	enableBallast = flag.Bool("ballast", false, "enable ballast")
	enableSchedGC = flag.Bool("schedgc", false, "enable GC every 10s")
)

func init() {
	flag.Parse()
}

func main() {
	if *enableBallast {
		ballast := make([]byte, 1<<30)
		runtime.GC()
		defer runtime.KeepAlive(&ballast)
	}

	if *enableSchedGC {
		go func() {
			for {
				time.Sleep(time.Second * 10)
				runtime.GC()
			}
		}()
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		b := make([]byte, 10*1024*1024)
		b[0] = 1
		time.Sleep(time.Millisecond * 10)
		w.Write([]byte("helloworld"))
	})
	http.ListenAndServe(":8899", nil)
}
