package elevhandler

import (
	"fmt"

	"../elevio"
)

var numFloors int = 4

type Orders struct {
	Inside []bool /** < The inside panel orders*/
	Up     []bool /** < The upwards orders from outside */
	Down   []bool /** < The downwards orders from outside */
}

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

func AddOrder(elevPt *ElevatorStatus, order elevio.ButtonEvent) {
	switch order.Button {
	case elevio.BT_Cab:
		elevPt.Orders.Inside[order.Floor] = true
		elevio.SetButtonLamp(elevio.BT_Cab, order.Floor, true) //FIX evt sett lys et annet sted
	case elevio.BT_HallUp:
		elevPt.Orders.Up[order.Floor] = true
	case elevio.BT_HallDown:
		elevPt.Orders.Down[order.Floor] = true
	}
	SetEndstation(elevPt)
}

func RemoveOrder(elevPt *ElevatorStatus, order elevio.ButtonEvent){
	switch order.Button {
	case elevio.BT_Cab:
		elevPt.Orders.Inside[order.Floor] = false
		elevio.SetButtonLamp(elevio.BT_Cab, order.Floor, false) //FIX evt sett lys et annet sted
	case elevio.BT_HallUp:
		elevPt.Orders.Up[order.Floor] = false
	case elevio.BT_HallDown:
		elevPt.Orders.Down[order.Floor] = false
	}
	SetEndstation(elevPt)
}

//ElevatorGetEndstation returns endstation
func SetEndstation(elevPt *ElevatorStatus) {
	switch elevPt.Direction {
	case elevio.MD_Down: //skiftet down og up
		for f := numFloors - 1; f >= 0; f-- {
			if elevPt.Orders.Inside[f] || elevPt.Orders.Down[f] || elevPt.Orders.Up[f] {
				elevPt.Endstation = f
			}
		}
	case elevio.MD_Up, elevio.MD_Stop: //bias til å gå oppover
		for f := 0; f < numFloors; f++ {
			if elevPt.Orders.Inside[f] || elevPt.Orders.Down[f] || elevPt.Orders.Up[f] {
				elevPt.Endstation = f
			}
		}
	}
	fmt.Print("set endstation: ")
	fmt.Println(elevPt.Endstation)
}

func ClearOrdersAtFloor(elevPt *ElevatorStatus) {
	elevPt.Orders.Inside[elevPt.Floor] = false
	elevio.SetButtonLamp(elevio.BT_Cab, elevPt.Floor, false) //FIX sett lys et annet sted evt
	
	if (elevPt.Direction == elevio.MD_Up) || (elevPt.Endstation == elevPt.Floor) {
		elevPt.Orders.Up[elevPt.Floor] = false
	}
	if (elevPt.Direction == elevio.MD_Down) || (elevPt.Endstation == elevPt.Floor) {
		elevPt.Orders.Down[elevPt.Floor] = false
	}
	SetEndstation(elevPt)
}