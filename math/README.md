# README

数据类型是个好东西，类型定义了一种数据组成以及允许在其上进行的操作。
数据类型是个好东西，它定义了一种最基础的“安全”，类型安全。

我们在进行数值运算时，有可能会“超出”类型本身的值域，但是受限于位宽限制，进而表现为“上溢出”。
以a+b为例：
- 如果a、b都是有符号数，且其符号相同，有可能超过最大值、最小值而在值域空间中轮转；
  两个正数相加，结果却是负数；两个负数相加，结果却是正数.
- 如果a、b都是无符号数，也有可能超过最大值而在值域空间中轮转。

这个很容易理解，今天我们想看下如何解决此类问题。

## 如何解决溢出问题

### 升级32位到64位？

这通常是第一反应，它可能是有效的，也可能无效。
- 有效：如果输入int32 a、b是有明确约束保证的，比如任意一个都必须在[-1*1<<31,(1<<31)-1]，
       a+b可能对int32可能会溢出，但是如果提升成int64则可以解决问题，前提有这样的约束保证；
- 无效：没有任何输入约束做保证，只是简单提升成int64 a、b是没有用的，极端情况，a=b=1<<36-1，
       a+b很明显就溢出了，这种就需要其他方法做保证。

### 设计上应该有上限？

在设计上就要有这方面的“数据”上的“安全”的意识，比如：
- 玩家每赛季的经验应该是有上限的，满经验后就提示玩家满经验，后续就不给加了；
- 比如用uint32表示经验值，那么加之前先测一下是否发生了溢出(v=orig+delta, 如果v小于任意一个则溢出)
  这很好理解，正常情况下，v应该大于orig、delta，就是逻辑反嘛。
  ps：不好理解？把值域想象成一个转盘，delta不可能让v在值域范围内转到orig，反之orig也不能让v转到delta。
  如果发生了溢出，则直接将v=maxUint32完事，多出来的就扔掉，提示玩家满经验。
- 或者，这里的满经验不一定要maxUint32，可以是认为设计好的一个小值，比如99999；
  如果输入有约束，比较小比如int8 a, int8 b，那么至少可以保证 if a+b > 99999 then v=999999 是ok的，
  也不会触及累积量v达到uint32最大值的情况。可能这种情况比较理想化了。

### 检查是否发生溢出？

言归正传，还是要有办法来比较可靠地检查运算结果a+b是否发生了溢出？

- 可以用大数计算来避免溢出，比如golang里面的math/big包。
  比如int32 a,b相加，按int c=a+b的方式，c有可能是个溢出后的错误结果。
  但是如果用大数计算，位宽充足可以算出正确结果，只要将其和maxInt32比较下即可知道是否发生了溢出。
  如果确实发生了溢出，应该如何处理，如fallback到满经验值不再加经验。
- 也可以不用大数计算，通过一些有趣的副作用也可以知道是否发生了溢出。
  比如在x86汇编中，可以通过 `test OF,OF` 来判断是否发生了溢出。
  高级语言中，就没那么直接，比如go，得借助一些其他办法来判断，这就是这个math_test.go要测试的东西。

## 当前测试

math_test.go中定义了两个函数safeSignedAdd、safeUnsignedAdd来对有符号数、无符号数加法进行安全的计算：
- 如果发生了溢出则返回错误，方便调用方处理；
- 如果没发生错误则返回两数之和；

我们想检测下如何更好地发现一些造成溢出的边界条件，我们使用go fuzztest来帮助发现潜在的问题。
我们设置了边界附近的值作为seed scorpus，这样方便go fuzztest引擎使用mutator微调输入参数时能够覆盖到边界条件。

其实也可以使用go fuzztest的随机构造输入的模式，但是这样往往需要执行更多的时间才有助于发现问题。
ps：改天再写篇文章详细介绍下go fuzztest内部是如何工作的。

这里看起来我们是为了使用go fuzztest而使用fuzztest，比如你怎么精心构造这样的seed scorpus的？其实不是为了用而用。
- 当我们设计实现一个函数时，脑海中应该知道输入是啥、输出是啥，过程中的极端case是啥，那你就有了一个输入的值域范围，
  或者说不同的参数组合有几种特殊的情况，可以多次调用 f.Seed(v1,v2,...)，来奖这些参数作为一个seed scorpus，
  以方便后续模糊测试引擎微调这些参数来覆盖特殊分支。
- 你不一定要精心构造出一定能触发异常边界的seed scorpus，你可以设置个大概的值，让后交给模糊测试引擎去做剩下的工作，
  假定一个参数是uint32类型，你设置的seed参数设置的是n，那么这个n最终会在[n-100,n+100]的范围内变化，当然下界、
  上界要在uint32范围内，每个参数都会这样变化。所以你的seed scorpus不一定刚好触发边界。
  模糊测试运行过程中，如果它发现某个输入发生了错误（t.Errorf标记的），或者此输入导致代码覆盖率提升了（给每条语句插桩），
  那么就会将当前输入作为一个新的seed scorpus存起来，在其基础上微调参数执行。
- 最终我们尽可能地覆盖了更多的代码，并尽力去发现可能存在的边界异常。但是确实不能保证一定能找到问题。

随机模式的话，输入参数随机意味着逼近边界异常处需要更多的测试用例，可能耗时很长才能发现，但是也不一定嗯呢发现。

## 模糊测试!=漫无目的的测试

在执行uint32上溢出模糊测试时，我专门设计了一个seed scorpus，如下所示：

```go
f.Add(uint32(0xffffffff-1000), uint32(0))
```

此输入下，mutator无能为力，执行了几分钟也发现不了问题，如果知道mutator原理的就很容易明白为什么。

```bash
$ go test -v -count=1 -fuzz=Fuzz_overflow_uint32 -run=^$
=== FUZZ  Fuzz_overflow_uint32
fuzz: elapsed: 0s, gathering baseline coverage: 0/1 completed
fuzz: elapsed: 0s, gathering baseline coverage: 1/1 completed, now fuzzing with 16 workers
fuzz: elapsed: 3s, execs: 767120 (255688/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 6s, execs: 1551732 (261545/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 9s, execs: 2352907 (267067/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 12s, execs: 3148542 (265216/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 15s, execs: 3945075 (265509/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 18s, execs: 4751252 (268643/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 21s, execs: 5550572 (266165/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 24s, execs: 6352445 (267419/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 27s, execs: 7149552 (265831/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 30s, execs: 7932675 (261108/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 33s, execs: 8739835 (268820/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 36s, execs: 9535421 (265373/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 39s, execs: 10330064 (264890/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 42s, execs: 11133080 (267670/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 45s, execs: 11920486 (262525/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 48s, execs: 12679807 (253079/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 51s, execs: 13478128 (265715/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 54s, execs: 14250999 (257967/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 57s, execs: 14970992 (239839/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 1m0s, execs: 15746188 (258630/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 1m3s, execs: 16540185 (264651/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 1m6s, execs: 17327573 (262428/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 1m9s, execs: 18118944 (263843/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 1m12s, execs: 18905240 (262118/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 1m15s, execs: 19694177 (262722/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 1m18s, execs: 20484654 (263715/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 1m21s, execs: 21263553 (259660/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 1m24s, execs: 22046809 (261068/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 1m27s, execs: 22844921 (265793/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 1m30s, execs: 23625904 (260572/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 1m33s, execs: 24424248 (266049/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 1m36s, execs: 25200019 (257994/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 1m39s, execs: 25988466 (263455/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 1m42s, execs: 26768530 (260038/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 1m45s, execs: 27549087 (260201/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 1m48s, execs: 28340498 (263787/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 1m51s, execs: 29144044 (267882/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 1m54s, execs: 29939219 (264884/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 1m57s, execs: 30712422 (257889/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 2m0s, execs: 31493491 (260378/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 2m3s, execs: 32265799 (257353/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 2m6s, execs: 33055825 (262993/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 2m9s, execs: 33839959 (261782/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 2m12s, execs: 34617785 (259259/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 2m15s, execs: 35401605 (261261/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 2m18s, execs: 36183561 (260667/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 2m21s, execs: 36968063 (261462/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 2m24s, execs: 37743465 (258485/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 2m27s, execs: 38525393 (260694/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 2m30s, execs: 39309042 (261184/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 2m33s, execs: 40082850 (257913/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 2m36s, execs: 40876146 (264475/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 2m39s, execs: 41660462 (261389/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 2m42s, execs: 42442111 (260618/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 2m45s, execs: 43234687 (264132/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 2m48s, execs: 43994959 (253473/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 2m51s, execs: 44772385 (258989/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 2m54s, execs: 45562731 (263543/sec), new interesting: 1 (total: 2)
fuzz: elapsed: 2m57s, execs: 46340129 (259039/sec), new interesting: 1 (total: 2)
^C

```

然后，进一步可以纠正下错误的测试思想，模糊测试!=漫无目的的测试，seed scorpus matters!

当我们设计上有章法，测试时关注边界，自然知道seed scorpus该如何设置，比如我们改成：

```go
f.Add(uint32(0xffffffff), uint32(0))
```

继续执行测试，很快就发现了边界case，代码中我们加了模糊测试的次数，发现第9轮它便发现了问题。

```bash
go test -v -count=1 -fuzz=Fuzz_overflow_uint32 -run=^$
=== FUZZ  Fuzz_overflow_uint32
fuzz: elapsed: 0s, gathering baseline coverage: 0/1 completed
fuzz: elapsed: 0s, gathering baseline coverage: 1/1 completed, now fuzzing with 16 workers
fuzz: elapsed: 0s, execs: 9 (383/sec), new interesting: 0 (total: 1)
--- FAIL: Fuzz_overflow_uint32 (0.02s)
    --- FAIL: Fuzz_overflow_uint32 (0.00s)
        math_test.go:31: iter-9 4294967198 + 118 = 20, err: overflow
    
    Failing input written to testdata/fuzz/Fuzz_overflow_uint32/a6532fa5f002651bb1003d5aedbea9bb5716a6d2a8fe7afff0b5252599a6d59b
    To re-run:
    go test -run=Fuzz_overflow_uint32/a6532fa5f002651bb1003d5aedbea9bb5716a6d2a8fe7afff0b5252599a6d59b
FAIL
exit status 1
FAIL    github.com/hitzhangjie/codemaster/math  0.026s
```

## 小结

总结了下如何解决数值计算时的溢出问题，从编码上、从策略上，以及介绍了如何使用go fuzztest来更好地发现潜在的问题。
关于模糊测试的一点想法，模糊测试 != 漫无目的的测试，seed scorpus的选择和设置很有价值。