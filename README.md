TTK 4145 - Elevator project Spring 2021
=======================================


Summary
-------
This document describes the overall architecture of the code, which is written in Go. The code consists of a main module, an Elevator FSM module, Order handler FSM module, Network FSM module, as well as a reboot file. The Network module is written by and borrowed from [klasbo](https://github.com/klasbo).

### Main
This function called in the terminal and spawns three threads which starts the FSM from the other modules. Main also declares go channels used for message passing in the different FSMs. Finally main is also launching a reboot file after the program crashes as a anti-fault tolerance system.

### Order handler FSM
The Order handler module receives orders and elevator statuses from Network and decides what to do with them. Which elevator should get the order is decided here, via a cost function. This information is then sent to the Network to communicate the information with all the elevators.

### Elevator FSM
The Elevator module will handle the physical behavior of the elevator. It will use different submodules to control and read inputs from the elevator.

### Network FSM
The Network module will transmit and receive messages from other elevators via UDP. The module is made to be used in a peer to peer configuration between elevators where the peers use broadcasting as the way of communicating.
