package initlights

import (
	"fmt"

	"../elevio"
)

func ClearAllOrderLights(numFloors int) {
	for floor := 0; floor < numFloors; floor++{
		elevio.SetButtonLamp(elevio.BT_HallUp, floor, false)
		elevio.SetButtonLamp(elevio.BT_Cab, floor, false)
		elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)
	}
}
/*
func SigintHandler(sig int) {
	//(void)(sig)
	fmt.Printf("Terminating elevator\n")
	elevio.SetMotorDirection(elevio.MD_Stop)
}
*/

//
func InitializeLights(addr string, numFloors int) {
	elevio.Init(addr, numFloors)
	ClearAllOrderLights(numFloors)
	elevio.SetDoorOpenLamp(false)
	elevio.SetStopLamp(false)
	elevio.SetFloorIndicator(0)

	// elevio.SetMotorDirection(elevio.MD_Down)
	// elevio.SetMotorDirection(elevio.MD_Stop)
	fmt.Print("Initializing complete\n")
}
