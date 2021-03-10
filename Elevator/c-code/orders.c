#ifndef LIB_H
#define LIB_H
#include <stdio.h>
#include <stdlib.h>
#include "hardware.h"
#include "orders.h"
#endif

void orders_update(Orders * p_orders,bool direction){
    for(int f = 0; f < HARDWARE_NUMBER_OF_FLOORS; f++){
            /* Internal orders */
            if(hardware_read_order(f, HARDWARE_ORDER_INSIDE)){
                hardware_command_order_light(f, HARDWARE_ORDER_INSIDE, 1);
                p_orders->inside[f] = 1;
            }

            /* Orders going up */
            if(hardware_read_order(f, HARDWARE_ORDER_UP)){
                hardware_command_order_light(f, HARDWARE_ORDER_UP, 1);
                p_orders->up[f] = 1;
            }

            /* Orders going down */
            if(hardware_read_order(f, HARDWARE_ORDER_DOWN)){
                hardware_command_order_light(f, HARDWARE_ORDER_DOWN, 1);
                p_orders->down[f] = 1;
            }
        }
    if (direction){
        for(int f = HARDWARE_NUMBER_OF_FLOORS - 1; f >= 0; f--){        //setting the first floor that has an order to current destination (counting from the top).
             if(p_orders->inside[f] || p_orders->down[f] || p_orders->up[f]){
                p_orders->endstation = f;
                break;
            }
            
        }
        
    }
    else{
        for(int f = 0; f < HARDWARE_NUMBER_OF_FLOORS; f++){             //setting the first floor that has an order to current destination (counting from the bottom).
            if(p_orders->inside[f] || p_orders->down[f] || p_orders->up[f]){
                p_orders->endstation = f;
                break;
            }
            
        }
    }
}

void orders_clear(Orders * p_orders,int floor, HardwareOrder order_type){
    hardware_command_order_light(floor, order_type, 0);
    switch(order_type){
        case HARDWARE_ORDER_INSIDE:{
            p_orders->inside[floor] = 0;
            break;
         }
        case HARDWARE_ORDER_UP:{
            p_orders->up[floor] = 0;
            break;
         }
        case HARDWARE_ORDER_DOWN:{
            p_orders->down[floor] = 0;
            break;
         }
    }
}

void orders_clear_all(Orders * p_orders){
    HardwareOrder order_types[3] = {
        HARDWARE_ORDER_UP,
        HARDWARE_ORDER_INSIDE,
        HARDWARE_ORDER_DOWN
    };

    for(int f = 0; f < HARDWARE_NUMBER_OF_FLOORS; f++){
        for(int i = 0; i < 3; i++){
            HardwareOrder order_type = order_types[i];
            orders_clear(p_orders ,f, order_type);
        }
    }
}