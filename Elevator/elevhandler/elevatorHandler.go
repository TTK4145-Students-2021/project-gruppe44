package elevhandler

import (
	"fmt"

	"../elevio"
)

var numFloors int = 4

//Orders comment (gjør evt omm til 2d array)
type Orders struct {
	Inside []bool /** < The inside panel orders*/
	Up     []bool /** < The upwards orders from outside */
	Down   []bool /** < The downwards orders from outside */
}

/*
type Order struct {
	id string // empty if no one has taken it
	timeStarted int


}
type OrdersAll struct {
	Inside []Order // < The inside panel orders
	Up     []Order // < The upwards orders from outside
	Down   []Order // < The downwards orders from outside
}

ordersAll.Up = [0, "heis1", "heis2"]
*/
type ElevatorStatus struct {
	Endstation int
	Orders     Orders
	Floor      int
	Direction  elevio.MotorDirection
}

func ElevatorStatusUpdateForever(order <-chan elevio.ButtonEvent, direction <-chan elevio.MotorDirection, floor <-chan int, clear <-chan bool, elevatorCH chan<- ElevatorStatus, ordersCH chan<- Orders) {
	myOrders := Orders{Inside: []bool{false, false, false, false}, Up: []bool{false, false, false, false}, Down: []bool{false, false, false, false}}
	//load orders from fil
	//var currentDir elevio.MotorDirection = elevio.MD_Stop
	elevator := ElevatorStatus{Endstation: 0, Orders: myOrders, Floor: 0, Direction: elevio.MD_Stop}
	elevatorCH <- elevator
	for {
		select {
		case d := <-direction:
			fmt.Println("Updated direction")
			elevator.Direction = d
			elevatorCH <- elevator
			fmt.Println(elevator)
		case f := <-floor:
			fmt.Println("Updated floor")
			elevator.Floor = f
			elevatorCH <- elevator
			fmt.Println(elevator)
		case o := <-order:
			fmt.Println("Updated order")
			elevator = ElevatorAddOrder(o, elevator)
			elevator.Endstation = ElevatorGetEndstation(elevator)
			elevatorCH <- elevator
			ordersCH <- elevator.Orders
			fmt.Println(elevator)
		case <-clear:
			fmt.Println("removed orders")
			ElevatorClearOrdersAtFloor(elevator)
			elevator.Endstation = ElevatorGetEndstation(elevator)
			elevatorCH <- elevator
			ordersCH <- elevator.Orders
			fmt.Println(elevator)
		}
	}

}

func ElevatorAddOrder(order elevio.ButtonEvent, elevator ElevatorStatus) ElevatorStatus {
	switch order.Button {
	case elevio.BT_Cab:
		elevator.Orders.Inside[order.Floor] = true
	case elevio.BT_HallUp:
		elevator.Orders.Up[order.Floor] = true
	case elevio.BT_HallDown:
		elevator.Orders.Down[order.Floor] = true
	}
	return elevator

}

//ElevatorGetEndstation returns endstation
func ElevatorGetEndstation(elevator ElevatorStatus) int {
	switch elevator.Direction {
	case elevio.MD_Up:
		for f := numFloors - 1; f >= 0; f-- {
			if elevator.Orders.Inside[f] || elevator.Orders.Down[f] || elevator.Orders.Up[f] {
				return f
			}
		}
	case elevio.MD_Down, elevio.MD_Stop: //bias til å gå nedover
		for f := 0; f < numFloors; f++ {
			if elevator.Orders.Inside[f] || elevator.Orders.Down[f] || elevator.Orders.Up[f] {
				return f
			}
		}

	}
	return elevator.Floor
}

/*
func ElevatorGetEndstation(floor_from int, floor_to int, elevator ElevatorStatus) int {
	if floor_from < floor_to {
		for f := floor_from; f <= floor_to; f++ {
			if elevator.orders.inside[f] || elevator.orders.down[f] || elevator.orders.up[f] {
				return f
			}
		}
	} else {
		for f := floor_from; f >= floor_to; f-- {
			if elevator.orders.inside[f] || elevator.orders.down[f] || elevator.orders.up[f] {
				return f
			}
		}
	}
	return elevator.floor
}
*/

func ElevatorClearOrdersAtFloor(elevator ElevatorStatus) { // evt legg til håndtering for å ikke fjerne ordre motsatt retning
	elevator.Orders.Inside[elevator.Floor] = false
	elevator.Orders.Up[elevator.Floor] = false
	elevator.Orders.Down[elevator.Floor] = false

}
