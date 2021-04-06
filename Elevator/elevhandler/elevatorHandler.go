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

func ElevatorStatusUpdateForever(elevatorPt *ElevatorStatus, order <-chan elevio.ButtonEvent, direction <-chan elevio.MotorDirection, floor <-chan int, clear <-chan int, elevatorCH chan<- ElevatorStatus, ordersCH chan<- Orders) {
	/*
		myOrders := Orders{Inside: []bool{false, false, false, false}, Up: []bool{false, false, false, false}, Down: []bool{false, false, false, false}}
		//load orders from fil
		//var currentDir elevio.MotorDirection = elevio.MD_Stop
		elevator := ElevatorStatus{Endstation: 0, Orders: myOrders, Floor: 0, Direction: elevio.MD_Stop}
		elevatorCH <- elevator
	*/
	for {
		select {
		case c := <-clear:
			fmt.Println("removed orders")
			ElevatorClearOrdersAtFloor(elevatorPt, c)
			ElevatorSetEndstation(elevatorPt)
			//elevatorCH <- elevator
			ordersCH <- elevatorPt.Orders
			fmt.Println(*elevatorPt)
		case d := <-direction:
			fmt.Println("Updated direction")
			elevatorPt.Direction = d
			//elevatorCH <- elevator
			fmt.Println(*elevatorPt)
		case f := <-floor:
			fmt.Println("Updated floor")
			elevatorPt.Floor = f
			elevio.SetFloorIndicator(f)
			//elevatorCH <- elevator
			fmt.Println(*elevatorPt)
		case o := <-order:
			fmt.Println("Updated order")
			ElevatorAddOrder(elevatorPt, o)
			ElevatorSetEndstation(elevatorPt)
			//elevatorCH <- elevator
			ordersCH <- elevatorPt.Orders
			fmt.Println(*elevatorPt)
		}
	}

}

func ElevatorAddOrder(elevatorPt *ElevatorStatus, order elevio.ButtonEvent) {
	switch order.Button {
	case elevio.BT_Cab:
		elevatorPt.Orders.Inside[order.Floor] = true
	case elevio.BT_HallUp:
		elevatorPt.Orders.Up[order.Floor] = true
	case elevio.BT_HallDown:
		elevatorPt.Orders.Down[order.Floor] = true
	}
}

//ElevatorGetEndstation returns endstation
func ElevatorSetEndstation(elevatorPt *ElevatorStatus) { //fant feilen, når heisa står i ro velger den alltid ned, men kan ha en ordre i ro
	switch elevatorPt.Direction {
	case elevio.MD_Down: //skiftet down og up
		for f := numFloors - 1; f >= 0; f-- {
			if elevatorPt.Orders.Inside[f] || elevatorPt.Orders.Down[f] || elevatorPt.Orders.Up[f] {
				elevatorPt.Endstation = f
			}
		}
	case elevio.MD_Up, elevio.MD_Stop: //bias til å gå nedover
		for f := 0; f < numFloors; f++ {
			if elevatorPt.Orders.Inside[f] || elevatorPt.Orders.Down[f] || elevatorPt.Orders.Up[f] {
				elevatorPt.Endstation = f
			}
		}

	}
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

func ElevatorClearOrdersAtFloor(elevatorPt *ElevatorStatus, floor int) { // evt legg til håndtering for å ikke fjerne ordre motsatt retning
	elevatorPt.Orders.Inside[floor] = false
	elevatorPt.Orders.Up[floor] = false
	elevatorPt.Orders.Down[floor] = false

}
