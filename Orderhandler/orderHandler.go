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
// The distances are complicated expressions
func CostFunction(orderReq elevio.ButtonEvent, elevStatus elevhandler.ElevatorStatus) int {
	distToRequest := DistanceBetweenFloors(elevStatus.Floor, orderReq.Floor)
	distToEndstation := DistanceBetweenFloors(elevStatus.Floor, elevStatus.Endstation)
	distFromEndstationToRequest := DistanceBetweenFloors(elevStatus.Endstation, orderReq.Floor)
	distTotal := distFromEndstationToRequest + distToEndstation

	// Boolean expressions:
	elevFloorUNDER := (elevStatus.Floor < orderReq.Floor)
	elevFloorOVER := (elevStatus.Floor > orderReq.Floor)
	elevDirUP := (elevStatus.Direction == elevio.MD_Up)
	elevDirDOWN := (elevStatus.Direction == elevio.MD_Down)
	orderReqUP := (orderReq.Button == elevio.BT_HallUp)
	orderReqDOWN := (orderReq.Button == elevio.BT_HallDown)
	endStationUNDER := (elevStatus.Endstation < orderReq.Floor)
	endStationOVER := (elevStatus.Endstation > orderReq.Floor)

	if elevFloorUNDER {
		if elevDirUP && orderReqUP {

			if endStationUNDER {
				return distToRequest
			} else {
				return distToRequest - 1
			}

		} else if elevDirUP && orderReqDOWN {

			if endStationUNDER {
				return distToRequest
			} else {
				return distTotal
			}

		} else {
			return distTotal
		} /*if (elevDirDOWN)*/

	} else if elevFloorOVER {
		if elevDirDOWN && orderReqUP {

			if endStationOVER {
				return distToRequest
			} else {
				return distTotal
			}

		} else if elevDirDOWN && orderReqDOWN {

			if endStationOVER {
				return distToRequest
			} else {
				return distToRequest - 1
			}

		} else {
			return distTotal
		} /*if (elevDirUP)*/

	} else {
		if (elevDirUP && orderReqUP) || (elevDirDOWN && orderReqDOWN) {
			return -1
		} else {
			return distTotal
		}
	}
}

func DistanceBetweenFloors(floor1, floor2 int) int {
	return int(math.Abs(float64(floor1) - float64(floor2)))
}

// Used to keep track of time for each order,
// so that a timeout flag occurs when the order has been active for a long time and not finished.
func OrderTimeoutFlag(elevPt *elevhandler.ElevatorStatus, order elevio.ButtonEvent) {

	// Calculate expected completion time for order
	timeLimitPerFloor := 5 * time.Second // Might have to adjust this time...
	numOfFloorsToMove := DistanceBetweenFloors(elevPt.Floor, order.Floor)
	totalTimeForOrder := timeLimitPerFloor * time.Duration(numOfFloorsToMove)

	time.Sleep(totalTimeForOrder)

	// If elevPT.order == true -> order has not completed, meaning something is wrong. Set TimeoutFlag.
	switch order.Button {
	case elevio.BT_Cab:

		if elevPt.Orders.Inside[order.Floor] == true {
			elevPt.Timeout = true
		} else {
			elevPt.Timeout = false
		}

	case elevio.BT_HallUp:

		if elevPt.Orders.Up[order.Floor] == true {
			elevPt.Timeout = true
		} else {
			elevPt.Timeout = false
		}

	case elevio.BT_HallDown:

		if elevPt.Orders.Down[order.Floor] == true {
			elevPt.Timeout = true
		} else {
			elevPt.Timeout = false
		}
	}
}

// When the order list is altered we will save the orders to file,
// this way we always have an updated order list in case of a crash.
// This file will be loaded on reboot.
// This module will have the needed functions to save and load the elevator status and order list.
func FileHandler() {

}

/************** OrderHandler **************/

var elevMap map[string]elevhandler.ElevatorStatus //map to store all the elevator statuses

// It will receive and keep track of all orders and use a cost function to decide which elevator should take which order.
func OrderHandlerFSM(myID string, newOrder <-chan elevio.ButtonEvent, finishedOrder <-chan elevio.ButtonEvent, elev <-chan elevhandler.Elevator, orderOut chan<- elevio.ButtonEvent, allOrders chan<- elevhandler.Orders) {
	// Inputs:
	// NewOrder ButtonEvent: This is a new order that should be handled.
	// FinishedOrder ButtonEvent: This is a finished order that should be cleared.
	// Elevator struct: Includes ElevatorStatus and ElevatorID. This is used to evaluate the cost of an order on each elevator.
	// IsConnected struct: Contains a Connected bool that says if the elevator is connected and ElevatorID.

	// Outputs:
	// Orders struct: A list of all orders, so that the elevator can turn on/off lights.
	// NewOrder ButtonEvent: The new order, sendt to the elevator who is going to take the order.
	elevMap = make(map[string]elevhandler.ElevatorStatus)

	for {
		select {
		case o := <-newOrder:
			ChooseElevator(elevMap, myID, o, orderOut)
		case c := <-finishedOrder:
			ClearOrder(c)
		case e := <-elev:
			UpdateElevators(elevMap, e)
		}

	}

}

// When the program turns on, it will load all local data from file.
// If there is nothing to load it will initialize with zero orders.
func Init() {

}

// The state machine for the orderHandler,
// waiting for every other function to do their thing before giving order list to the Network module.
func Wait() {

}

// When OrderHandler receives a new order,
// this function will choose which elevator gets the order by calculating the CostFunction() on each elevator.
// It then updates its local data, OrdersAll.
// These orders are sent out for the elevator to update it’s lights.
// The elevator who got the order will send the specific order and an order confirmation as well.
// Elevators that are not connected will not be taken into consideration.
func ChooseElevator(elevMap map[string]elevhandler.ElevatorStatus, myID string, order elevio.ButtonEvent, orderOut chan<- elevio.ButtonEvent) {
	//TODO: Add to OrdersAll struct, and save to file
	//TODO: Add small offset based on id, so that no elevator will have the same cost
	fmt.Println("Got order request")
	minCost := 1000000000000000000 //Big number so that the first cost is lower, couldn't use math.Inf(1) because of different types. Fix this
	var chosenElev string
	for id, elevStatus := range elevMap {
		cost := CostFunction(order, elevStatus)
		if (cost < minCost) && elevStatus.IsConnected {
			minCost = cost
			chosenElev = id
		}
	}
	if chosenElev == myID {
		orderOut <- order
		fmt.Println("Took the order")
	}
}

// When an order confirmation is recieved, this function will set that order as confirmed.
func ConfirmOrder() {

}

// When an order times out, this function will resend that order to the network module as a new order.
func ResendOrder() {

}

// When a new ElevatorStatus or Connection bool is received,
// this function will save this as local data for the CostFunction() to use.
// And if this elevator has an order that is not in the OrdersAll list, it will add this order.
func UpdateElevators(elevMap map[string]elevhandler.ElevatorStatus, elev elevhandler.Elevator) {
	//TODO: save map to file
	//TODO: check if the elevator has order not in list, if yes add order.
	elevMap[elev.ID] = elev.Status
	fmt.Print("Updated elevator Map: ")
	fmt.Println(elevMap)

}

// When an old order is finished, this function will clear/update the order table.
func ClearOrder(order elevio.ButtonEvent) {
	fmt.Println("Cleared order")
}

// When an elevator reconnects after having lost connection,
// this function will update the elevator with the most up-to-date order table.
// It will also update the other elevators on what this elevator’s previous orders were,
// and make sure all elevators have the same order table.
func SyncElevators() {

}
