package Orderhandler

import (
	"fmt"
	"math"
	"time"

	"../Elevator/elevhandler"
	"../Elevator/elevio"
)

// TODO:
// Add OrderTimeoutFlag

// NOTE: 
// testbranch vs masterbranch

// We assume that the person waits for the assigned elevator
// Do we need a condition where it is in MD_STOP mode? 
func CostFunction(orderReq elevio.ButtonEvent, elevStatus elevhandler.ElevatorStatus) int {
	
	distToRequest := int(math.Abs(float64(orderReq.Floor) - float64(elevStatus.Floor)))
	distToEndstation := int(math.Abs(float64(elevStatus.Endstation) - float64(elevStatus.Floor)))
	distFromEndstationToRequest := int(math.Abs(float64(orderReq.Floor) - float64(elevStatus.Endstation)))
	distTotal := distFromEndstationToRequest + distToEndstation

	// Boolean expressions:
	elevFloorUNDER	:= (elevStatus.Floor < orderReq.Floor)
	elevFloorOVER	:= (elevStatus.Floor > orderReq.Floor)
	elevDirUP		:= (elevStatus.Direction == elevio.MD_Up)
	elevDirDOWN		:= (elevStatus.Direction == elevio.MD_Down)
	orderReqUP		:= (orderReq.Button == elevio.BT_HallUp)
	orderReqDOWN	:= (orderReq.Button == elevio.BT_HallDown)
	endStationUNDER	:= (elevStatus.Endstation < orderReq.Floor)
	endStationOVER	:= (elevStatus.Endstation > orderReq.Floor)

	if (elevFloorUNDER) {
		if (elevDirUP && orderReqUP){

			if (endStationUNDER){return distToRequest
			} else				{return distToRequest - 1}

		} else if (elevDirUP && orderReqDOWN){
			
			if (endStationUNDER){return distToRequest
			} else				{return distTotal}

		} else					{return distTotal} /*if (elevDirDOWN)*/
	
	} else if (elevFloorOVER) {
		if (elevDirDOWN && orderReqUP){
			
			if (endStationOVER)	{return distToRequest
			} else				{return distTotal}

		} else if (elevDirDOWN && orderReqDOWN){
			
			if (endStationOVER)	{return distToRequest
			} else				{return distToRequest - 1}

		} else					{return distTotal} /*if (elevDirUP)*/
	
	} else {
		if ((elevDirUP && orderReqUP) || (elevDirDOWN && orderReqDOWN)) {return -1
		} else {return distTotal}
	}
}

// Used to keep track of time for each order,
// so that a timeout flag occurs when the order has been active for a long time and not finished.
func OrderTimeoutFlag(elevPt *elevhandler.ElevatorStatus) bool {
	fmt.Printf("This is from OrderTimeoutFlag()!")
	
	timeLimit := 5 * time.Second
	time.Sleep(timeLimit)

	/* 
	// Might not need these:
	startTime := time.Now()
	deadline  := startTime.Add(timeLimit)
	diff := timeLimit.Sub(start).Seconds()
	fmt.Printf("difference = %f seconds\n", diff)
	*/

	
	// Do something...
	// Check if floor has changed?
	// Check if # of orders have changed/decreased?

	/*
	// PSEUDOCODE
	if currentFloor != prevFloor	: return elevPT.Timeout = false
	else 							: return elevPT.Timeout = true

	// OR
	if #elevPT.Orders != #prev elevPT.Orders	: return elevPT.Timeout = false
	else 										: return elevPT.Timeout = true
	*/


	return false // temp return
}

// When the order list is altered we will save the orders to file,
// this way we always have an updated order list in case of a crash.
// This file will be loaded on reboot.
// This module will have the needed functions to save and load the elevator status and order list.
func FileHandler(){
	
}

/************** OrderHandler **************/

// It will receive and keep track of all orders and use a cost function to decide which elevator should take which order.
func OrderHandlerFSM(){
// Inputs:
	// NewOrder ButtonEvent: This is a new order that should be handled.
	// FinishedOrder ButtonEvent: This is a finished order that should be cleared.
	// Elevator struct: Includes ElevatorStatus and ElevatorID. This is used to evaluate the cost of an order on each elevator.
	// IsConnected struct: Contains a Connected bool that says if the elevator is connected and ElevatorID.

// Outputs: 
	// Orders struct: A list of all orders, so that the elevator can turn on/off lights.
	// NewOrder ButtonEvent: The new order, sendt to the elevator who is going to take the order.
}


// When the program turns on, it will load all local data from file.
// If there is nothing to load it will initialize with zero orders.
func Init(){

}


// The state machine for the orderHandler,
// waiting for every other function to do their thing before giving order list to the Network module.
func Wait(){

}


// When OrderHandler receives a new order, 
// this function will choose which elevator gets the order by calculating the CostFunction() on each elevator.
// It then updates its local data, OrdersAll.
// These orders are sent out for the elevator to update it’s lights.
// The elevator who got the order will send the specific order and an order confirmation as well.
// Elevators that are not connected will not be taken into consideration.
func ChooseElevator(){

}


// When an order confirmation is recieved, this function will set that order as confirmed.
func ConfirmOrder(){

}


// When an order times out, this function will resend that order to the network module as a new order.
func ResendOrder(){

}


// When a new ElevatorStatus or Connection bool is received,
// this function will save this as local data for the CostFunction() to use.
// And if this elevator has an order that is not in the OrdersAll list, it will add this order.
func UpdateElevators(){

}


// When an old order is finished, this function will clear/update the order table.
func ClearOrder(){

}


// When an elevator reconnects after having lost connection,
// this function will update the elevator with the most up-to-date order table.
// It will also update the other elevators on what this elevator’s previous orders were,
// and make sure all elevators have the same order table.
func SyncElevators(){

}