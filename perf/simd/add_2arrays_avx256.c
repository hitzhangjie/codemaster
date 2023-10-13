#include <immintrin.h> // Need this in order to be able to use the AVX "intrinsics" (which provide access to instructions without writing assembly)
#include <stdio.h>

int main(int argc, char **argv) {
    float a[8] __attribute__ ((aligned (32))); // Intel documentation states that we need 32-byte alignment to use _mm256_load_ps/_mm256_store_ps
    float b[8]  __attribute__ ((aligned (32))); // GCC's syntax makes this look harder than it is: https://gcc.gnu.org/onlinedocs/gcc-6.4.0/gcc/Common-Variable-Attributes.html#Common-Variable-Attributes
    float result[8]  __attribute__ ((aligned (32)));
    __m256 va, vb, vresult; // __m256 is a 256-bit datatype, so it can hold 8 32-bit floats

    // initialize arrays (just {0,1,2,3,4,5,6,7})
    for (int i = 0; i < 8; i++) {
        a[i] = (float)i;
        b[i] = (float)i;
    }

    // load arrays into SIMD registers
    va = _mm256_load_ps(a); // https://software.intel.com/en-us/node/694474
    vb = _mm256_load_ps(b); // same

    // add them together
    vresult = _mm256_add_ps(va, vb); // https://software.intel.com/en-us/node/523406

    // store contents of SIMD register into memory
    _mm256_store_ps(result, vresult); // https://software.intel.com/en-us/node/694665

    // print out result
    for (int i = 0; i < 8; i++) {
        printf("%f\n", result[i]);
    }
    
    return 0;
}
