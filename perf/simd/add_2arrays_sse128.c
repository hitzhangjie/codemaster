#include <xmmintrin.h> // Need this in order to be able to use the SSE "intrinsics" (which provide access to instructions without writing assembly)
#include <stdio.h>

int main(int argc, char **argv) {
    float a[4], b[4], result[4]; // a and b: input, result: output
    __m128 va, vb, vresult; // these vars will "point" to SIMD registers

    // initialize arrays (just {0,1,2,3})
    for (int i = 0; i < 4; i++) {
        a[i] = (float)i;
        b[i] = (float)i;
    }
    
    // load arrays into SIMD registers
    va = _mm_loadu_ps(a); // https://software.intel.com/en-us/node/524260
    vb = _mm_loadu_ps(b); // same

    // add them together
    vresult = _mm_add_ps(va, vb);

    // store contents of SIMD register into memory
    _mm_storeu_ps(result, vresult); // https://software.intel.com/en-us/node/524262

    // print out result
    for (int i = 0; i < 4; i++) {
        printf("%f\n", result[i]);
    }
}
