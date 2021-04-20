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
	"./reboot"
)

/*
	TODO:
		remove magic numbers like NumFloors in struct inits
		Network:
	
		OrderHandler:
			Filehandling:
			- Integrate in rest of code (mainly updateOrder)
			- FIX bug where simulation doesnt update
			- Must receive networkdisconnet from somewhere
			OrderTimeoutFlag:
			- Integrate in rest of code (uncertain where it should be called)
			- Find good inputs
			OnTimeout:
			- Integrate with rest of code
			Send orderlights to elevatorFSM
			Init() - Removed
		Elevator:
			Refactoring (remove uneccesary while loops)
			Emergency stop
		FIX README files and similar stuff that we need/dont need
	
*/

func main() {
	numFloors := 4
	//addr := "localhost:15657"

	// choose addr by '-addr=my_address'
	var addr string
	flag.StringVar(&addr, "addr", "localhost:15657", "Address of elevator server")
	// Our id can be anything. Here we pass it on the command line, using
	// `go run main.go -id=our_id`
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()
	
	// ... or alternatively, we can use the local IP address.
	// (But since we can run multiple programs on the same PC, we also append the process ID)
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}
	reboot.Reboot(addr, id)

	elevFromFSM		:= make(chan elevhandler.Elevator)
	elevFromNet		:= make(chan elevhandler.Elevator)
	elevInit		:= make(chan elevhandler.ElevatorStatus)
	orderFromElev	:= make(chan elevio.ButtonEvent)
	orderFromHandler:= make(chan elevio.ButtonEvent)
	orderFromNet	:= make(chan elevio.ButtonEvent)
	discon 			:= make(chan []string)
	orderResend 	:= make(chan elevio.ButtonEvent)
	orderRemove 	:= make(chan elevio.ButtonEvent)
	startup			:= make(chan bool, 1)
	// timeout			:= make(chan bool)

	go func() { //temp for å tømme ubrukte channels
		for {
			select {
			case o := <-orderResend:
				orderRemove <- o
				orderFromElev <- o
			}
		}
	}()
	// Happens only at boot
	startup <- true

	go Orderhandler.OrderHandlerFSM(id, orderFromNet, elevFromNet, orderFromHandler, orderResend, elevInit, discon, startup) //timeout)
	go Network.Network(id, orderFromNet, orderFromElev, elevFromFSM, elevFromNet, discon)
	Elevator.ElevatorFSM(id, addr, numFloors, orderFromHandler, orderFromElev, elevFromFSM, orderRemove, elevInit) //timeout)
}
