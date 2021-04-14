package elevhandler

import (
	"../elevio"
	"../../Orderhandler"
	"time"
)

var numFloors int = 4

//Orders comment (gjør evt omm til 2d array)
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
	case elevio.BT_HallUp:
		elevPt.Orders.Up[order.Floor] = true
	case elevio.BT_HallDown:
		elevPt.Orders.Down[order.Floor] = true
	}
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
}

func ClearOrdersAtFloor(elevPt *ElevatorStatus, finishedOrder chan<- elevio.ButtonEvent) {
	/*
		elevPt.Orders.Inside[elevPt.Floor] = false
		if elevPt.Endstation == elevPt.Floor {
			elevPt.Orders.Up[elevPt.Floor] = false
			elevPt.Orders.Down[elevPt.Floor] = false
		} else if elevPt.Direction == elevio.MD_Up {
			elevPt.Orders.Up[elevPt.Floor] = false
		} else {
			elevPt.Orders.Down[elevPt.Floor] = false
		}
	*/
	if elevPt.Orders.Inside[elevPt.Floor] {
		elevPt.Orders.Inside[elevPt.Floor] = false
		finishedOrder <- elevio.ButtonEvent{Floor: elevPt.Floor, Button: elevio.BT_Cab}
	}
	if (elevPt.Orders.Up[elevPt.Floor]) && ((elevPt.Direction == elevio.MD_Up) || (elevPt.Endstation == elevPt.Floor)) {
		elevPt.Orders.Up[elevPt.Floor] = false
		finishedOrder <- elevio.ButtonEvent{Floor: elevPt.Floor, Button: elevio.BT_HallUp}
	}
	if (elevPt.Orders.Down[elevPt.Floor]) && ((elevPt.Direction == elevio.MD_Down) || (elevPt.Endstation == elevPt.Floor)) {
		elevPt.Orders.Down[elevPt.Floor] = false
		finishedOrder <- elevio.ButtonEvent{Floor: elevPt.Floor, Button: elevio.BT_HallDown}
	}

}

// Used to keep track of time for each order,
// so that a timeout flag occurs when the order has been active for a long time and not finished.
func OrderTimeoutFlag(elevPt *ElevatorStatus, order elevio.ButtonEvent) {

	// Calculate expected completion time for order
	timeLimitPerFloor := 3 * time.Second // Might have to adjust this time...
	numOfFloorsToMove := Orderhandler.DistanceBetweenFloors(elevPt.Floor, order.Floor)
	totalTimeForOrder := timeLimitPerFloor * time.Duration(numOfFloorsToMove)

	timer := time.NewTimer(totalTimeForOrder)
	// time.Sleep(totalTimeForOrder)

	// If elevPT.order == true -> order has not completed, meaning something is wrong. Set TimeoutFlag.
	for {
		select {
		case <-timer.C:
			switch order.Button {
				case elevio.BT_Cab:
			
					if elevPt.Orders.Inside[order.Floor] == true {
						elevPt.Timeout = true
					} else {
						elevPt.Timeout = false
					}
					return
	
				case elevio.BT_HallUp:

					if elevPt.Orders.Up[order.Floor] == true {
						elevPt.Timeout = true
					} else {
						elevPt.Timeout = false
					}
					return

				case elevio.BT_HallDown:

					if elevPt.Orders.Down[order.Floor] == true {
						elevPt.Timeout = true
					} else {
						elevPt.Timeout = false
					}
					return
			}
		default:
		}
	}

	// If elevPT.order == true -> order has not completed, meaning something is wrong. Set TimeoutFlag.
	// switch order.Button {
	// case elevio.BT_Cab:

	// 	if elevPt.Orders.Inside[order.Floor] == true {
	// 		elevPt.Timeout = true
	// 	} else {
	// 		elevPt.Timeout = false
	// 	}

	// case elevio.BT_HallUp:

	// 	if elevPt.Orders.Up[order.Floor] == true {
	// 		elevPt.Timeout = true
	// 	} else {
	// 		elevPt.Timeout = false
	// 	}

	// case elevio.BT_HallDown:

	// 	if elevPt.Orders.Down[order.Floor] == true {
	// 		elevPt.Timeout = true
	// 	} else {
	// 		elevPt.Timeout = false
	// 	}
	// }
}
