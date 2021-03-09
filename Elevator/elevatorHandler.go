package orders

import "./elevio"

var numFloors int = 4

//Orders comment (gjør evt omm til 2d array)
type Orders struct {
	inside []bool /** < The inside panel orders*/
	up     []bool /** < The upwards orders from outside */
	down   []bool /** < The downwards orders from outside */
}

type ElevatorStatus struct {
	endstation int
	orders     Orders
	floor      int
	direction  elevio.MotorDirection
}

func ElevatorStatusUpdateForever(order <-chan elevio.ButtonEvent, direction <-chan elevio.MotorDirection, floor <-chan int, stop <-chan bool, elevatorCH chan<- ElevatorStatus) {
	myOrders := Orders{inside: []bool{false, false, false, false}, up: []bool{false, false, false, false}, down: []bool{false, false, false, false}}
	//load orders from fil
	//var currentDir elevio.MotorDirection = elevio.MD_Stop
	elevator := ElevatorStatus{endstation: 0, orders: myOrders, floor: 0, direction: elevio.MD_Stop}
	for {
		select {
		case d := <-direction:
			elevator.direction = d
		case f := <-floor:
			elevator.floor = f
		case o := <-order:
			elevator = ElevatorAddOrder(o, elevator)
			elevator.endstation = ElevatorGetEndstation(elevator)
		case <-stop:
			ElevatorClearOrdersAtFloor(elevator)
		}
	}

}

func ElevatorAddOrder(order elevio.ButtonEvent, elevator ElevatorStatus) ElevatorStatus {
	switch order.Button {
	case elevio.BT_Cab:
		elevator.orders.inside[order.Floor] = true
	case elevio.BT_HallUp:
		elevator.orders.up[order.Floor] = true
	case elevio.BT_HallDown:
		elevator.orders.down[order.Floor] = true
	}
	return elevator

}

//ElevatorGetEndstation returns endstation
func ElevatorGetEndstation(elevator ElevatorStatus) int {
	switch elevator.direction {
	case elevio.MD_Up:
		for f := numFloors - 1; f >= 0; f-- {
			if elevator.orders.inside[f] || elevator.orders.down[f] || elevator.orders.up[f] {
				return f
			}
		}
	case elevio.MD_Down, elevio.MD_Stop: //bias til å gå nedover
		for f := 0; f < numFloors; f++ {
			if elevator.orders.inside[f] || elevator.orders.down[f] || elevator.orders.up[f] {
				return f
			}
		}

	}
	return elevator.floor
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
	elevator.orders.inside[elevator.floor] = false
	elevator.orders.up[elevator.floor] = false
	elevator.orders.down[elevator.floor] = false

}
