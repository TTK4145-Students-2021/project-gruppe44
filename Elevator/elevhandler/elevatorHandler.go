package elevhandler

import (
	"fmt"
	"time"

	"../elevio"
)

type ElevatorState int
const (
	ST_Idle		ElevatorState = 0
	ST_MovingUp               = 1
	ST_MovingDown             = 2
	ST_StopUp				  = 3
	ST_StopDown 			  = 4
	ST_DoorOpen 			  = 5
)

type Orders struct {
	Inside []bool // The inside panel orders
	Up     []bool // The upwards orders from outside
	Down   []bool // The downwards orders from outside
}

type ElevatorStatus struct {
	Endstation  	  int
	Floor       	  int
	Available		  bool
	Orders      	  Orders
	TimeSinceNewFloor time.Time
	Direction   	  elevio.MotorDirection
	State 			  ElevatorState
}

type Elevator struct {
	ID     string
	Status ElevatorStatus
}

func AddOrder(elevPt *ElevatorStatus, order elevio.ButtonEvent) {
	switch order.Button {
	case elevio.BT_Cab:
		elevPt.Orders.Inside[order.Floor] = true
		elevio.SetButtonLamp(elevio.BT_Cab, order.Floor, true)
	
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
		elevio.SetButtonLamp(elevio.BT_Cab, order.Floor, false)
	
	case elevio.BT_HallUp:
		elevPt.Orders.Up[order.Floor] = false
	
	case elevio.BT_HallDown:
		elevPt.Orders.Down[order.Floor] = false
	}
	SetEndstation(elevPt)
}

func SetEndstation(elevPt *ElevatorStatus) {
	elevPt.Endstation = elevPt.Floor
	switch elevPt.Direction {
	case elevio.MD_Down:
		for f := len(elevPt.Orders.Inside) - 1; f >= 0; f-- {
			if elevPt.Orders.Inside[f] || elevPt.Orders.Down[f] || elevPt.Orders.Up[f] {
				elevPt.Endstation = f
			}
		}
	// Bias up
	case elevio.MD_Up, elevio.MD_Stop: 
		for f := 0; f < len(elevPt.Orders.Inside); f++ {
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
	elevio.SetButtonLamp(elevio.BT_Cab, elevPt.Floor, false)
	
	if (elevPt.Direction == elevio.MD_Up) || (elevPt.Endstation == elevPt.Floor) {
		elevPt.Orders.Up[elevPt.Floor] = false
	}
	if (elevPt.Direction == elevio.MD_Down) || (elevPt.Endstation == elevPt.Floor) {
		elevPt.Orders.Down[elevPt.Floor] = false
	}
	SetEndstation(elevPt)
}