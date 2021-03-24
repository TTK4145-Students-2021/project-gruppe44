package orderhandler

import (
	"fmt"
	"math"

	"../elevio"
	"../elevhandler"
)

// TODO:
// Might have to add a condition in CostFunction where: elevStatus.Floor == orderReq.Floor


// We assume that the person waits for the assigned elevator
func CostFunction(orderReq elevio.ButtonEvent, elevStatus elevhandler.ElevatorStatus) int {
	
	distToRequest := int(math.Abs(float64(orderReq.Floor) - float64(elevStatus.Floor)))
	distToEndstation := int(math.Abs(float64(elevStatus.Endstation) - float64(elevStatus.Floor)))
	distFromEndstationToRequest := int(math.Abs(float64(orderReq.Floor) - float64(elevStatus.Endstation)))
	distTotal := distFromEndstationToRequest + distToEndstation

	// Boolean expressions:
	elevFloorUNDER	:= (elevStatus.Floor < orderReq.Floor)
	elevFloorOVER	:= (elevStatus.Floor > orderReq.Floor)
	elevDirUP		:= (elevStatus.Direction == elevio.MD_Up)
	elevDirDOWN		:= (elevStatus.Direction == elevio.MD_Down)
	orderReqUP		:= (orderReq.Button == elevio.BT_HallUp)
	orderReqDOWN	:= (orderReq.Button == elevio.BT_HallDown)
	endStationUNDER	:= (elevStatus.Endstation <= orderReq.Floor)
	endStationOVER	:= (elevStatus.Endstation > orderReq.Floor)

	if (elevFloorUNDER) {
		if (elevDirUP && orderReqUP){

			if (endStationUNDER){return distToRequest
			} else				{return distToRequest - 1}

		} else if (elevDirUP && orderReqDOWN){
			
			if (endStationUNDER){return distToRequest
			} else				{return distTotal}

		} else if (elevDirDOWN)	{return distTotal}
	
	} else if (elevFloorOVER) {
		if (elevDirDOWN && orderReqUP){
			
			if (endStationOVER)	{return distToRequest
			} else				{return distTotal}

		} else if (elevDirDOWN && orderReqDOWN){
			
			if (endStationOVER)	{return distToRequest
			} else				{return distToRequest - 1}

		} else if (elevDirUP)	{return distTotal}
	}
	return 0 // return 0 for now (if conditions over are not met)
}

func Timer(){
	
}

func FileHandler(){
	
}

/************** OrderHandler **************/

func OrderHandlerFSM(){
	
}

func Init(){

}

func Wait(){

}

func ChooseElevator(){

}

func ConfirmOrder(){

}

func ResendOrder(){

}

func UpdateElevators(){

}

func ClearOrder(){

}

func SyncElevators(){

}