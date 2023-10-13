#include <stdio.h>
#include <immintrin.h>

#define TIMES 10000000
#define LENGTH 8

#define i256 __m256i
#define avx256_set32 _mm256_set_epi32
#define avx256_add _mm256_add_epi32

int main(int argc, char *argv[])
{
    for (int k=0; k<TIMES; k++) {
        i256 first = avx256_set32(1, 2, 3, 4, 5, 6, 7, 8);
        i256 second = avx256_set32(1, 1, 1, 1, 1, 1, 1, 1);
        i256 result = avx256_add(first ,second);
        
        /*
        int *value = (int*)&result;
        for (int i=0; i<LENGTH; i++) {
            printf("%d ", value[i]);
        }
        printf("\n");
        */
    }

    return 0;
}
