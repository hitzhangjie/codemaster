#include <stdio.h>

#define TIMES 10000000
#define LENGTH 8

int main(int argc, char *argv[])
{
    for (int k=0; k<TIMES; k++) {
        int nums1[LENGTH] = {1, 2, 3, 4, 5, 6, 7, 8};
        int nums2[LENGTH] = {1, 1, 1, 1, 1, 1, 1, 1};
        int result[LENGTH] = {0};
    
        for (int i=0; i<LENGTH; i++) {
            result[i] = nums1[i] + nums2[i];
        }
    }

    return 0;
}
