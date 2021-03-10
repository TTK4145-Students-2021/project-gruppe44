#ifndef LIB_H
#define LIB_H
#include <stdio.h>
#include <stdlib.h>
#include <signal.h>
#include "hardware.h"
#include "init.h"
#endif


static void clear_all_order_lights(){
    HardwareOrder order_types[3] = {
        HARDWARE_ORDER_UP,
        HARDWARE_ORDER_INSIDE,
        HARDWARE_ORDER_DOWN
    };

    for(int f = 0; f < HARDWARE_NUMBER_OF_FLOORS; f++){
        for(int i = 0; i < 3; i++){
            HardwareOrder type = order_types[i];
            hardware_command_order_light(f, type, 0);
        }
    }
}

static void sigint_handler(int sig){
    (void)(sig);
    printf("Terminating elevator\n");
    hardware_command_movement(HARDWARE_MOVEMENT_STOP);
    exit(0);
}



void init(){
    int error = hardware_init();
    if(error != 0){
        fprintf(stderr, "Unable to initialize hardware\n");
        exit(1);
    }

    signal(SIGINT, sigint_handler);


    clear_all_order_lights();
    hardware_command_door_open(0);
    hardware_command_stop_light(0);
    hardware_command_floor_indicator_on(0);

    hardware_command_movement(HARDWARE_MOVEMENT_DOWN); //går ned til første etasje
    while(!hardware_read_floor_sensor(0)){}
    hardware_command_movement(HARDWARE_MOVEMENT_STOP);
    printf("initializing complete\n");



}
