package weired

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"
)

func Test_MonotonicClock(t *testing.T) {
	// =================================================================================================================
	// 背景知识: Wall Clock vs Monotonic Clock
	//
	// Go 的 time.Time 结构内部同时包含了两种时间:
	// 1. Wall Clock (墙上时钟): 这是我们通常意义上的时间，比如 "2023-10-27 10:00:00"。它会受到系统时间设置的影响，
	//    可能会因为手动修改或NTP同步而发生跳变 (向前或向后)。
	// 2. Monotonic Clock (单调时钟): 这是一个从某个不确定的过去时间点开始，只会单调递增、永不回退的时钟。它不受系统
	//    Wall Clock 的影响，像一个秒表，只用来精确测量时间间隔。
	//
	// time.Now() 会同时获取这两种时间并存放在 time.Time 对象中 (如果操作系统支持单调时钟)。
	//
	// 关键点:
	// - 当你打印一个 time.Time 对象时 (e.g., fmt.Println(t1))，它显示的是 Wall Clock 的时间。
	// - 当你对两个 time.Time 对象进行操作时 (e.g., t2.After(t1), t2.Sub(t1))，Go 会优先使用 Monotonic Clock
	//   来进行计算。这样可以保证即使系统时间被修改，你测量到的时间差依然是准确的物理流逝时间。
	// =================================================================================================================

	var t1 time.Time
	var t2 time.Time

	// 1. 获取当前时间 t1
	// 此时，t1 内部包含了当前的 Wall Clock 时间和 Monotonic Clock 读数。
	{
		t1 = time.Now()
		fmt.Println("t1 (Wall Clock):", t1)
	}

	// 2. 手动将系统时间向前拨 5 秒
	// 这个操作只会修改系统的 Wall Clock，不会影响 Monotonic Clock。
	{
		// 注意: 这个测试需要 sudo 权限来修改系统时间，执行时可能会失败。
		// ps: 如果执行失败，后续的结果将不符合预期。
		tt := t1.Add(-5 * time.Second)
		fmt.Println("change system time to:", tt)

		cmd := exec.Command("sudo", "date", "-s", "@"+fmt.Sprintf("%d", tt.Unix()))
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("error: failed to change system time. This test requires sudo privileges.\n")
			fmt.Printf("error detail: %v\n", err)
			fmt.Printf("output: %s\n", string(output))
			// 如果无法修改时间，后续测试无意义，直接跳过
			t.Skip("Skipping test: failed to set system time.")
			os.Exit(1) // 在非 testing 环境下，应选择 os.Exit(1)
		}
	}

	// 3. 再次获取当前时间 t2
	{
		t2 = time.Now()

		// 打印 t2 时，显示的是修改后的 Wall Clock 时间。
		// 所以 t2 的 Wall Clock 确实约等于 t1 的 Wall Clock 减去 5 秒。
		fmt.Println("t2 (Wall Clock):", t2)

		// 你的问题: for t2 = t1-5s, so t2>t1, right?
		// 答案: 不对。这个比较操作 (`After`) 使用的是 Monotonic Clock。
		//
		// 解释:
		// 尽管 t2 的 Wall Clock 时间比 t1 要早，但在物理世界中，获取 t2 的这个动作确实发生在获取 t1 之后。
		// Go 在执行 t2.After(t1) 时，会比较两者内部的 Monotonic Clock 读数。
		// 因为 Monotonic Clock 不受系统时间修改的影响，它忠实地记录了时间的流逝，所以 t2 的 Monotonic Clock
		// 读数必然大于 t1 的读数。
		//
		// 结论: t2.After(t1) 的结果是 true，代表从 t1 到 t2，物理时间确实是向前流逝了。
		// 这正是 Monotonic Clock 的价值所在：提供一个可靠的、不受外界干扰的时间测量基准。
		fmt.Println("t2.After(t1)? (t2.After(t1), using Monotonic Clock):", t2.After(t1))

		// 我们可以通过 Sub 操作来更清晰地看到这一点。
		// t2.Sub(t1) 计算的是两个时间点之间物理流逝的时间，结果应该是一个很小的正值 (几毫秒或微秒)。
		// 这个值代表了执行 "修改系统时间" 和 "第二次调用 time.Now()" 所花费的时间。
		elapsed := t2.Sub(t1)
		fmt.Printf("Elapsed time (t2.Sub(t1), using Monotonic clock): %v\n", elapsed)
	}
}
