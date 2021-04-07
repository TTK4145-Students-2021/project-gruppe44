package main

import (
	"./Elevator"
	"./Elevator/elevio"
)

func main() {
	numFloors := 4
	addr := "localhost:15657"

	orderRx := make(chan elevio.ButtonEvent)
	orderTx := make(chan elevio.ButtonEvent)

	go func() {
		for {
			select {
			case o := <-orderTx:
				orderRx <- o
			}
		}
	}()

	//var id string
	//go Network.NetworkMain(id, orderRx, orderTx)
	Elevator.ElevatorFSM(addr, numFloors, orderRx, orderTx)
}
