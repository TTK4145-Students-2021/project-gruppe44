#ifndef LIB_H
#define LIB_H
#include <stdio.h>
#include <stdlib.h>
#include <time.h>
#include "hardware.h"
#include "orders.h"
#include "floor.h"
#include "elevator.h"

#define idle_state 0
#define moving_up_state 1
#define moving_down_state 2
#define stop_up_state 3
#define stop_down_state 4
#define stop_state 5

#define up_direction 1
#define down_direction 0
#endif

void elevator(){ 
    int state = idle_state;
    Orders orders = {{0,0,0,0},{0,0,0,0},{0,0,0,0},0};
    Orders * p_orders = &orders;
    Floor floor = {0, 0};
    Floor * p_floor = &floor;
    while(1){
        switch (state) {
            case idle_state: {
                state = idle(p_orders, p_floor);
                break;
            }
            case moving_up_state: {
                state = moving(p_orders, p_floor, up_direction);
                break;
            }
            case moving_down_state: {
                state = moving(p_orders, p_floor, down_direction);
                break;
            }
            case stop_up_state: {
                state = stop(p_orders, p_floor, up_direction);
                break;
            }
            case stop_down_state: {
                state = stop(p_orders, p_floor, down_direction);
                break;
            }
            case stop_state: {
                state = emergency_stop(p_orders, p_floor);
                break;
            }
        }

    }
}

int idle(Orders * p_orders, Floor * p_floor){
    hardware_command_movement(HARDWARE_MOVEMENT_STOP);
    while(1){
        orders_update(p_orders,1);

        if(hardware_read_stop_signal()){return stop_state;}

        if(p_orders->inside[p_floor->current]||p_orders->up[p_floor->current]){    //all these extra ifs is because of the stop button
            if(hardware_read_floor_sensor(p_floor->current)){
                p_orders->endstation = p_floor->current; 
                return stop_up_state;
            }
            else if(p_floor->above){return moving_down_state;} 
                else {return moving_up_state;}
        }
        if(p_orders->down[p_floor->current]){
            if(hardware_read_floor_sensor(p_floor->current)){ 
                p_orders->endstation = p_floor->current;  
                return stop_down_state;
            }
            else{ 
                if(p_floor->above){return moving_down_state;}
                else {return moving_up_state;}
            }
        }                                                      


        for(int f = 0; f < p_floor->current; f++){
            if(p_orders->inside[f] || p_orders->down[f] || p_orders->up[f]){
                p_orders->endstation = f;
                return moving_down_state;
            }
        }
        for(int f = HARDWARE_NUMBER_OF_FLOORS -1; f > p_floor->current; f--){
            if(p_orders->inside[f] || p_orders->down[f] || p_orders->up[f]){
                p_orders->endstation = f;
                return moving_up_state;
            }
        }
    }

}


int moving(Orders * p_orders, Floor * p_floor, bool direction){

    if(direction){hardware_command_movement(HARDWARE_MOVEMENT_UP);}
    else{hardware_command_movement(HARDWARE_MOVEMENT_DOWN);}
    while(1){
        orders_update(p_orders, direction);
        if(hardware_read_stop_signal()){return stop_state;}

        if(floor_update(p_floor, direction)){
            if(direction && (p_orders->up[p_floor->current]||p_orders->inside[p_floor->current]||p_orders->endstation == p_floor->current)){return stop_up_state;}
            if(!direction && (p_orders->down[p_floor->current]||p_orders->inside[p_floor->current]||p_orders->endstation == p_floor->current)){return stop_down_state;}

        }
    }
}



int stop(Orders * p_orders, Floor * p_floor, bool direction){
    hardware_command_movement(HARDWARE_MOVEMENT_STOP);
    HardwareOrder order_types[3] = {
        HARDWARE_ORDER_UP,
        HARDWARE_ORDER_INSIDE,
        HARDWARE_ORDER_DOWN
    };
    for(int i = 0; i < 3; i++){
        HardwareOrder order_type = order_types[i];
        orders_clear(p_orders, p_floor->current, order_type);
    }

    hardware_command_door_open(1);  //åpner døra her
    int sec = 0, trigger = 3;       //timer 3 sek
    clock_t before = clock();
    while(sec < trigger){

        if(hardware_read_stop_signal()){return stop_state;}


        orders_update(p_orders, direction);            //tar imot ordre samtidig
        if(hardware_read_obstruction_signal()){  //Resetter timeren når obstruction er på
            before = clock();
        }

        clock_t diff = clock() - before;
        sec = diff/ CLOCKS_PER_SEC;
    }
    hardware_command_door_open(0);  //lukker døra her

    if(direction && (p_orders->endstation > p_floor->current)){return moving_up_state;}          //set above = 1 her, for  sikre emergency
    else if(direction && (p_orders->endstation < p_floor->current)){return moving_down_state;}   //above = 0
    else{return idle_state;}
}


int emergency_stop(Orders * p_orders, Floor * p_floor){
    hardware_command_movement(HARDWARE_MOVEMENT_STOP);
    hardware_command_stop_light(1);
    orders_clear_all(p_orders);
    if(hardware_read_floor_sensor(p_floor->current)){
        hardware_command_door_open(1);
        while(hardware_read_stop_signal()){}
        hardware_command_stop_light(0);
        return stop_down_state;
    }
    else{
        while(hardware_read_stop_signal()){}
        hardware_command_stop_light(0);
        return idle_state;
    }
}