package elevinit

import (
	"fmt"
	"time"

	"../elevhandler"
	"../elevio"
)

func InitializeElevator(addr string,
						numFloors int,
						floorCH <-chan int,
						elevPt *elevhandler.ElevatorStatus,
						elev <-chan elevhandler.ElevatorStatus) {
	*elevPt = <- elev
	fmt.Println(*elevPt)

	for f := 0; f < len(elevPt.Orders.Inside); f++ {
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
		elevio.SetFloorIndicator(f)
	}
	elevhandler.SetEndstation(elevPt)
	elevPt.Available = true
	elevPt.TimeSinceNewFloor = time.Now()
	switch {
	case elevPt.Endstation < elevPt.Floor:
		elevio.SetMotorDirection(elevio.MD_Down)
		elevPt.Direction = elevio.MD_Down
		elevPt.State = elevhandler.ST_MovingDown
	case elevPt.Endstation > elevPt.Floor:
		elevio.SetMotorDirection(elevio.MD_Up)
		elevPt.Direction = elevio.MD_Up
		elevPt.State = elevhandler.ST_MovingUp
	default:
		elevio.SetMotorDirection(elevio.MD_Stop)
		elevPt.Direction = elevio.MD_Stop
		elevPt.State = elevhandler.ST_Idle
	}
	fmt.Print("Initializing complete\n")
}