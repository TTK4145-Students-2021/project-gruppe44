package main

import (
	"./Elevator"
	"./Elevator/elevio"
)

func main() {
	numFloors := 4
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
	Elevator.ElevatorFSM("localhost:15657", numFloors, orderRx, orderTx)

}
