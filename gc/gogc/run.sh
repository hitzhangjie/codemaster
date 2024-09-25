#!/bin/bash -e

#压力源：ab -c 64 -n 1000000 -s 10 http://127.0.0.1:8899/

# default behavior
#gc 6506 @17.684s 11%: 0.10+0.44+0.070 ms clock, 2.6+0.060/0.66/0+1.6 ms cpu, 3->3->1 MB, 4 MB goal, 0 MB stacks, 0 MB globals, 24 P
#gc 6507 @17.685s 11%: 0.091+0.78+0.056 ms clock, 2.1+0.030/0.59/0+1.3 ms cpu, 3->4->1 MB, 4 MB goal, 0 MB stacks, 0 MB globals, 24 P
#gc 6508 @17.687s 11%: 0.12+0.38+0.048 ms clock, 3.0+0.047/0.66/0+1.1 ms cpu, 3->3->1 MB, 4 MB goal, 0 MB stacks, 0 MB globals, 24 P
#GOTRACEBACK=crash GODEBUG=gctrace=1 GOGC=100 ./gogc --ballast=false --schedgc=false

# GOGC=100 GOMEMLIMIT=1GiB 堆太小，远远到不了1g限制值，跟上面这种情况一样，频繁gc
#gc 6542 @15.427s 12%: 0.53+0.32+0.040 ms clock, 12+0.034/0.68/0+0.97 ms cpu, 3->3->1 MB, 4 MB goal, 0 MB stacks, 0 MB globals, 24 P
#gc 6543 @15.429s 12%: 0.16+0.71+0.18 ms clock, 3.9+0.061/1.1/0+4.4 ms cpu, 3->4->1 MB, 4 MB goal, 0 MB stacks, 0 MB globals, 24 P
#gc 6544 @15.432s 12%: 0.22+0.35+0.004 ms clock, 5.2+0.028/0.71/0+0.10 ms cpu, 3->3->1 MB, 4 MB goal, 0 MB stacks, 0 MB globals, 24 P
#GOTRACEBACK=crash GODEBUG=gctrace=1 GOGC=100 GOMEMLIMIT=1GiB ./gogc --ballast=false --schedgc=false

# GOGC=2000 GOMEMLIMIT=1GiB,
#gc 33 @15.124s 0%: 0.15+0.48+0.067 ms clock, 3.8+0.059/0.83/0+1.6 ms cpu, 76->76->1 MB, 80 MB goal, 0 MB stacks, 0 MB globals, 24 P
#gc 34 @15.470s 0%: 0.20+0.41+0.043 ms clock, 4.9+0.033/0.75/0+1.0 ms cpu, 76->76->1 MB, 80 MB goal, 0 MB stacks, 0 MB globals, 24 P
#gc 35 @15.826s 0%: 0.084+0.80+0.082 ms clock, 2.0+0.21/1.1/0+1.9 ms cpu, 76->76->1 MB, 80 MB goal, 0 MB stacks, 0 MB globals, 24 P
#GOTRACEBACK=crash GODEBUG=gctrace=1 GOGC=2000 GOMEMLIMIT=1GiB ./gogc --ballast=false --schedgc=false

# GOGC=off GOMEMLIMIT=1GiB
#gc 1 @6.366s 0%: 0.022+0.56+0.071 ms clock, 0.55+0.14/1.3/0+1.7 ms cpu, 934->934->1 MB, 937 MB goal, 0 MB stacks, 0 MB globals, 24 P
#gc 2 @10.797s 0%: 0.12+0.73+0.12 ms clock, 3.0+0.076/1.9/0+3.0 ms cpu, 929->929->1 MB, 932 MB goal, 0 MB stacks, 0 MB globals, 24 P
#GOTRACEBACK=crash GODEBUG=gctrace=1 GOGC=off GOMEMLIMIT=1GiB ./gogc --ballast=false --schedgc=false

GOTRACEBACK=crash GODEBUG=gctrace=1 GOGC=100 GOMEMLIMIT=256MiB ./gogc --ballast=false --schedgc=false


#GOTRACEBACK=crash GODEBUG=gctrace=1 GOGC=100 GOMEMLIMIT=2867MiB ./gogc --ballast=true --schedgc=false
#GOTRACEBACK=crash ./gogc --ballast=true
