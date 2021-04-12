package main

import (
	"./Elevator"
	"./Elevator/elevio"
	"./Network"
)

//Currently, the program starts the elevator and connects it to the network.
//When a new order is made, a random elevator gets the order.
//to run the elevator use 'go run main.go --id=our_id'

func main() {
	numFloors := 4
	addr := "localhost:15657"

	orderRx := make(chan elevio.ButtonEvent)
	orderTx := make(chan elevio.ButtonEvent)

	/*
		go func() {
			for {
				select {
				case o := <-orderTx:
					orderRx <- o
				}

			}

		}()
	*/

	var id string
	go Network.Network(id, orderRx, orderTx)
	Elevator.ElevatorFSM(addr, numFloors, orderRx, orderTx)
}
