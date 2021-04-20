package elevinit

import (
	"fmt"

	// "../Elevator/elevhandler"
	"../elevhandler"
	"../elevio"
)

func ClearAllOrderLights(numFloors int) {
	for floor := 0; floor < numFloors; floor++ {
		elevio.SetButtonLamp(elevio.BT_HallUp, floor, false)
		elevio.SetButtonLamp(elevio.BT_Cab, floor, false)
		elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)
	}
}

func updateOrderLights(orders <-chan elevhandler.Orders) { // usikker på om denne skal være her FIX
	for {
		select {
		case o := <-orders:
			for f := 0; f < len(o.Inside); f++ { //var lat, gadd ikke å fikse at forskjellige order types har ferre ordre
				elevio.SetButtonLamp(elevio.BT_Cab, f, o.Inside[f])
				elevio.SetButtonLamp(elevio.BT_HallUp, f, o.Up[f])
				elevio.SetButtonLamp(elevio.BT_HallDown, f, o.Down[f])
			}
		}
	}
}

func InitializeElevator(addr string,
						numFloors int,
						floorCH <-chan int,
						elevPt *elevhandler.ElevatorStatus,
						elev <-chan elevhandler.ElevatorStatus) {
	//elevio.Init(addr, numFloors)
	*elevPt = <- elev
	// ClearAllOrderLights(numFloors)

	for f := 0; f < len(elevPt.Orders.Inside); f++ { //var lat, gadd ikke å fikse at forskjellige order types har ferre ordre
		elevio.SetButtonLamp(elevio.BT_Cab, f, elevPt.Orders.Inside[f])
		elevio.SetButtonLamp(elevio.BT_HallUp, f, elevPt.Orders.Up[f])
		elevio.SetButtonLamp(elevio.BT_HallDown, f, elevPt.Orders.Down[f])
	}

	elevio.SetDoorOpenLamp(false)
	elevio.SetStopLamp(false)
	elevio.SetFloorIndicator(0)

	elevio.SetMotorDirection(elevio.MD_Down)
	select {
	case f := <-floorCH:
		elevPt.Floor = f
		//elevPt.Endstation = f
		elevhandler.SetEndstation(elevPt)
		elevio.SetFloorIndicator(f)
	}
	fmt.Println(*elevPt)

	elevio.SetMotorDirection(elevio.MD_Stop)
	fmt.Print("Initializing complete\n")
}
