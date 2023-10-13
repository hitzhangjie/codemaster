SIMD programming demos

SSE is supported after nearly all kinds of CPUs after 2000.
AVX is supported by most newer CPUs.
While AVX2 is supported by more rencent newer CPUs.

SSE: add 128-bit register files
AVX: add 256-bit register files
AVX2: add 512-bit register files

Here I added serveral demos:
- ./add_2arrays_sse128.c: add two arrays by SSE
- ./add_2arrays_sse128.c: add two arrays by AVX2
    Before using AVX2 instructions, we must align the arrays, or 'Segmentation fault' will occur.
    And compile option '-mavx2' must be provided, because AVX2 is supported on more rencent and newer CPUs and AVX2 isn't enabled by default in gcc.

- ./bench_add_2arrays.c + ./bench_add_2arrays_avx2.c
    Adding two arrays 1KW times, then compare the execution timecost. AVX2 wins by running 2X as fast as ordinary method.
