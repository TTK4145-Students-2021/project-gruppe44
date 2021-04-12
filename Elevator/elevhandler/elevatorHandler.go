package elevhandler

import (
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
	Endstation  int
	Floor       int
	Timeout     bool
	IsConnected bool
	Orders      Orders
	Direction   elevio.MotorDirection
}

type Elevator struct {
	ID     string
	Status ElevatorStatus
}

func AddOrder(elevatorPt *ElevatorStatus, order elevio.ButtonEvent) {
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
func SetEndstation(elevatorPt *ElevatorStatus) {
	switch elevatorPt.Direction {
	case elevio.MD_Down: //skiftet down og up
		for f := numFloors - 1; f >= 0; f-- {
			if elevatorPt.Orders.Inside[f] || elevatorPt.Orders.Down[f] || elevatorPt.Orders.Up[f] {
				elevatorPt.Endstation = f
			}
		}
	case elevio.MD_Up, elevio.MD_Stop: //bias til å gå oppover
		for f := 0; f < numFloors; f++ {
			if elevatorPt.Orders.Inside[f] || elevatorPt.Orders.Down[f] || elevatorPt.Orders.Up[f] {
				elevatorPt.Endstation = f
			}
		}

	}
}

func ClearOrdersAtFloor(elevatorPt *ElevatorStatus) {
	elevatorPt.Orders.Inside[elevatorPt.Floor] = false
	if elevatorPt.Endstation == elevatorPt.Floor {
		elevatorPt.Orders.Up[elevatorPt.Floor] = false
		elevatorPt.Orders.Down[elevatorPt.Floor] = false
	} else if elevatorPt.Direction == elevio.MD_Up {
		elevatorPt.Orders.Up[elevatorPt.Floor] = false
	} else {
		elevatorPt.Orders.Down[elevatorPt.Floor] = false
	}

}
