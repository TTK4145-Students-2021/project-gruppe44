package main

import (
	"flag"
	"fmt"
	"os"

	"./Elevator"
	"./Elevator/elevhandler"
	"./Elevator/elevio"
	"./Network"
	"./Network/localip"
	"./Orderhandler"
)

//Currently, the program starts the elevator and connects it to the network.
//When a new order is made, a random elevator gets the order.
//to run the elevator use 'go run main.go --id=our_id'

func main() {
	numFloors := 4
	addr := "localhost:15657"

	orderFromNet := make(chan elevio.ButtonEvent)
	orderFromElev := make(chan elevio.ButtonEvent)
	orderFromHandlr := make(chan elevio.ButtonEvent)
	orderFinished := make(chan elevio.ButtonEvent)
	orderLights := make(chan elevhandler.Orders)
	elevFromFSM := make(chan elevhandler.Elevator)
	elevFromNet := make(chan elevhandler.Elevator)
	/*
		go func() { //temp for å tømme ubrukte channels
			for {
				select {
				case <-elevStatusCH:
				}

			}

		}()
	*/
	var id string

	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	// ... or alternatively, we can use the local IP address.
	// (But since we can run multiple programs on the same PC, we also append the
	//  process ID)
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}

	go Network.Network(id, orderFromNet, orderFromElev, elevFromFSM, elevFromNet)
	go Orderhandler.OrderHandlerFSM(id, orderFromNet, orderFinished, elevFromNet, orderFromHandlr, orderLights)
	Elevator.ElevatorFSM(id, addr, numFloors, orderFromHandlr, orderFromElev, elevFromFSM, orderFinished)

}
