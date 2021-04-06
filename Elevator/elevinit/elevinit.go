package elevinit

import (
	"fmt"

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

func InitializeElevator(addr string, numFloors int, floorCH <-chan int, elevPt *elevhandler.ElevatorStatus) {
	//elevio.Init(addr, numFloors)
	ClearAllOrderLights(numFloors)
	elevio.SetDoorOpenLamp(false)
	elevio.SetStopLamp(false)
	elevio.SetFloorIndicator(0)

	elevio.SetMotorDirection(elevio.MD_Down)
	select {
	case f := <-floorCH:
		elevPt.Floor = f
		elevPt.Endstation = f
		elevio.SetFloorIndicator(f)
	}
	elevio.SetMotorDirection(elevio.MD_Stop)
	fmt.Print("Initializing complete\n")
}
