/*
 * @Author: calmwu 
 * @Date: 2019-02-21 14:58:05 
 * @Last Modified by: calmwu
 * @Last Modified time: 2019-02-21 15:04:49
 */

#include <stdint.h>
#include <stdio.h>
#include "export_gofunc.h"

int32_t main(int32_t argc, char* argv[]) {
    printf("invoke golang export func\n");
    GoString name = {"calmwu", 6};
    SayHello(name);
    SayBye();
    return 0;
}

// g++ -o cinvokego main.c export_gofunc.a -lpthread
