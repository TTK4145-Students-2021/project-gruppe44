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
		Remove magic numbers like NumFloors in struct inits
		Remove unnessecary comments in main and other files
		OrderHandler:
			Filehandling:
			- Must receive networkdisconnet from somewhere
			OnTimeout:
			- Integrate with rest of code
			Send orderlights to elevatorFSM
			Init:
			- Add check to see if file is empty
			FIX lights
		Elevator:
			Refactoring (remove uneccesary while loops)
			Emergency stop
		FIX README files and similar stuff that we need/dont need

		TIPS: ctrl+f: type in FIX
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
	orderRemove 	:= make(chan elevio.ButtonEvent)
<<<<<<< HEAD
	timeOutToElev := make(chan bool)
	// timeout			:= make(chan bool)
=======
	orderResend 	:= make(chan elevio.ButtonEvent)
	discon 			:= make(chan []string)
>>>>>>> 918b2257b70ba4d6e3d7aa9e3d2e1fe81fa7d587

	go func() { //temp for å tømme ubrukte channels
		for {
			select {
			case o := <-orderResend:
				orderRemove <- o
				orderFromElev <- o
			}
		}
	}()
<<<<<<< HEAD
	// Happens only at boot

	go Orderhandler.OrderHandlerFSM(id, orderFromNet, elevFromNet, orderFromHandler, orderResend, elevInit, discon, timeOutToElev) //timeout)
	go Network.Network(id, orderFromNet, orderFromElev, elevFromFSM, elevFromNet, discon)
	Elevator.ElevatorFSM(id, addr, numFloors, orderFromHandler, orderFromElev, elevFromFSM, orderRemove, elevInit, timeOutToElev) //timeout)
=======

	go Orderhandler.OrderHandlerFSM(id, orderFromNet, elevFromNet, orderFromHandler, orderResend, elevInit, discon)
	go Network.Network(id, orderFromNet, orderFromElev, elevFromFSM, elevFromNet, discon)
	Elevator.ElevatorFSM(id, addr, numFloors, orderFromHandler, orderFromElev, elevFromFSM, orderRemove, elevInit, orderResend)
>>>>>>> 918b2257b70ba4d6e3d7aa9e3d2e1fe81fa7d587
}
