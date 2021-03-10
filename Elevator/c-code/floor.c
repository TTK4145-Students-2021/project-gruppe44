#ifndef LIB_H
#define LIB_H
#include <stdio.h>
#include <stdlib.h>
#include <stdbool.h>
#include "hardware.h"
#include "floor.h"
#endif

bool floor_update(Floor * p_floor, bool direction){
    for(int f = 0; f < HARDWARE_NUMBER_OF_FLOORS; f++){
        if(hardware_read_floor_sensor(f)){
            p_floor->current = f;
            p_floor->above = direction; 
            hardware_command_floor_indicator_on(f);
            return 1;
        }
    }
    return 0;
}
