#ifndef LIB_H
#define LIB_H
#include <stdio.h>
#include <stdlib.h>
#include "hardware.h"
#include "orders.h"
#include "floor.h"
#include "init.h"
#include "elevator.h"
#endif


int main(){
    init();
    elevator();
    return 0;
}
