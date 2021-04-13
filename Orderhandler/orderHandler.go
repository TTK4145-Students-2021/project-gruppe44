package Orderhandler

import (
	"fmt"
	"math"
	"sort"
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

type Order struct {
	ID        string //ID of elevator who has the order, empty string if no elevator
	Confirmed bool   //true if confirmed, false if else
	//TimeStarted time.Time //currently unused, but might be used for timeout flag
}

type Confirmation struct {
	ID    string             //ID of elevator who confirmed order
	Order elevio.ButtonEvent //The order to be confirmed
}

type HallOrders struct {
	//Inside []Order /** < The inside panel orders*/ //we ignore inside orders as this is handled directly by the elevator
	Up   []Order /** < The upwards orders from outside */
	Down []Order /** < The downwards orders from outside */
}

var elevMap map[string]elevhandler.ElevatorStatus //map to store all the elevator statuses

// It will receive and keep track of all orders and use a cost function to decide which elevator should take which order.
func OrderHandlerFSM(myID string, newOrder <-chan elevio.ButtonEvent, finishedOrder <-chan elevio.ButtonEvent, confirmationIn <-chan Confirmation, elev <-chan elevhandler.Elevator, orderOut chan<- elevio.ButtonEvent, confirmationOut chan<- Confirmation, allOrders chan<- elevhandler.Orders) {
	// Inputs:
	// NewOrder ButtonEvent: This is a new order that should be handled.
	// FinishedOrder ButtonEvent: This is a finished order that should be cleared.
	// confirmedOrder ButtonEvent: This is and order to be confirmed
	// Elevator struct: Includes ElevatorStatus and ElevatorID. This is used to evaluate the cost of an order on each elevator.
	// IsConnected struct: Contains a Connected bool that says if the elevator is connected and ElevatorID.

	// Outputs:
	// Orders struct: A list of all orders, so that the elevator can turn on/off lights.
	// NewOrder ButtonEvent: The new order, sendt to the elevator who is going to take the order.
	o := Order{ID: "", Confirmed: false}
	hallOrders := HallOrders{Up: []Order{o, o, o, o}, Down: []Order{o, o, o, o}} //FIX: initialize in init, remove set order count
	//var hallOrders HallOrders
	ordersPt := &hallOrders
	elevMap = make(map[string]elevhandler.ElevatorStatus)
	for {
		select {
		case conf := <-confirmationIn:
			ConfirmOrder(ordersPt, conf)
		case o := <-newOrder:
			ChooseElevator(elevMap, ordersPt, myID, o, orderOut, confirmationOut)
		case c := <-finishedOrder:
			ClearOrder(ordersPt, c)
		case e := <-elev:
			UpdateElevators(elevMap, ordersPt, e)
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
func ChooseElevator(elevMap map[string]elevhandler.ElevatorStatus, ordersPt *HallOrders, myID string, order elevio.ButtonEvent, orderOut chan<- elevio.ButtonEvent, conf chan<- Confirmation) {
	//TODO: save to file
	fmt.Println("Got order request")
	minCost := 1000000000000000000 //Big number so that the first cost is lower, couldn't use math.Inf(1) because of different types. Fix this
	var chosenElev string
	//sorted ids to make sure every elevator chooses the same elev when cost is the same.
	ids := make([]string, 0, len(elevMap))
	for id := range elevMap {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	//Calculate costs and choose elevator
	for i := 0; i < len(elevMap); i++ {
		id := ids[i]
		elevStatus := elevMap[ids[i]]
		cost := CostFunction(order, elevStatus)
		if (cost < minCost) && elevStatus.IsConnected {
			minCost = cost
			chosenElev = id
		}
	}
	//add order to list
	switch order.Button {
	case elevio.BT_HallUp:
		ordersPt.Up[order.Floor].ID = chosenElev
	case elevio.BT_HallDown:
		ordersPt.Down[order.Floor].ID = chosenElev
	}
	if chosenElev == myID {
		orderOut <- order
		conf <- Confirmation{ID: myID, Order: order}
		fmt.Println("Took the order")
	} else {
		fmt.Println("Didn't take order")
	}
	fmt.Print("Current order list: ")
	fmt.Println(*ordersPt)
}

// When an order confirmation is recieved, this function will set that order as confirmed.
func ConfirmOrder(ordersPt *HallOrders, conf Confirmation) {
	switch conf.Order.Button {
	case elevio.BT_HallUp:
		if ordersPt.Up[conf.Order.Floor].ID == conf.ID {
			ordersPt.Up[conf.Order.Floor].Confirmed = true
			fmt.Println("Confirmed order")
		}
	case elevio.BT_HallDown:
		if ordersPt.Down[conf.Order.Floor].ID == conf.ID {
			ordersPt.Down[conf.Order.Floor].Confirmed = true
			fmt.Println("Confirmed order")
		}
	}
}

// When an order times out, this function will resend that order to the network module as a new order.
func ResendOrder() {

}

// When a new ElevatorStatus or Connection bool is received,
// this function will save this as local data for the CostFunction() to use.
// And if this elevator has an order that is not in the OrdersAll list, it will add this order.
func UpdateElevators(elevMap map[string]elevhandler.ElevatorStatus, ordersPt *HallOrders, elev elevhandler.Elevator) {
	//TODO: save map to file
	//TODO: check if the elevator has order not in list, if yes add order.
	elevMap[elev.ID] = elev.Status
	fmt.Print("Updated elevator Map: ")
	fmt.Println(elevMap)

}

// When an old order is finished, this function will clear/update the order table.
func ClearOrder(ordersPt *HallOrders, order elevio.ButtonEvent) {
	//TODO: save order list to file
	switch order.Button {
	case elevio.BT_HallUp:
		ordersPt.Up[order.Floor].ID = ""
		ordersPt.Up[order.Floor].Confirmed = false
	case elevio.BT_HallDown:
		ordersPt.Down[order.Floor].ID = ""
		ordersPt.Down[order.Floor].Confirmed = false
	}
	fmt.Println("Cleared order")
	fmt.Print("Current order list: ")
	fmt.Println(*ordersPt)
}

// When an elevator reconnects after having lost connection,
// this function will update the elevator with the most up-to-date order table.
// It will also update the other elevators on what this elevator’s previous orders were,
// and make sure all elevators have the same order table.
func SyncElevators() {

}
