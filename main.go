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

/*
	TODO:
		Network:
			Send disconnected flag to orderhandler
		OrderHandler:
			Filehandling
			What happens on disconnect?
			What happens on reconnect? SyncElevators()
			Send orderlights to elevatorFSM
			Init()
			ResendOrder()
			UpdateElevators <- add confirmation check, and finished check here (instead of sending them)
		Elevator:
			Refactoring (remove uneccesary while loops)



*/

func main() {
	numFloors := 4
	//addr := "localhost:15657"

	var addr string // choose addr by '-addr=my_address'
	flag.StringVar(&addr, "addr", "localhost:15657", "Address of elevator server")
	flag.Parse()

	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`
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

	orderFromNet := make(chan elevio.ButtonEvent)
	orderFromElev := make(chan elevio.ButtonEvent)
	orderFromHandlr := make(chan elevio.ButtonEvent)
	orderLights := make(chan elevhandler.Orders)
	finIn := make(chan elevio.ButtonEvent)
	finOut := make(chan elevio.ButtonEvent)
	confOut := make(chan Orderhandler.Confirmation)
	confIn := make(chan Orderhandler.Confirmation)
	elevFromFSM := make(chan elevhandler.Elevator)
	elevFromNet := make(chan elevhandler.Elevator)
	/*
		go func() { //temp for å tømme ubrukte channels
			for {
				select {
				case c := <-confOut:
					confIn <- c
				}

			}

		}()
	*/

	go Network.Network(id, orderFromNet, orderFromElev, elevFromFSM, elevFromNet, confOut, confIn, finOut, finIn)
	go Orderhandler.OrderHandlerFSM(id, orderFromNet, finIn, confIn, elevFromNet, orderFromHandlr, confOut, orderLights)
	Elevator.ElevatorFSM(id, addr, numFloors, orderFromHandlr, orderFromElev, elevFromFSM, finOut)

}
