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
	"./Reboot"
)

func main() {
	
	var numFloors int 
	flag.IntVar(&numFloors, "numFloors", 4, "Number of floors to the elevator")
	
	var addr string
	flag.StringVar(&addr, "addr", "localhost:15657", "Address of elevator server")
	
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()
	
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}
	reboot.Reboot(addr, id)

	elevFromFSM		 := make(chan elevhandler.Elevator)
	elevFromNet		 := make(chan elevhandler.Elevator)
	elevInit		 := make(chan elevhandler.ElevatorStatus)
	orderFromElev	 := make(chan elevio.ButtonEvent)
	orderFromHandler := make(chan elevio.ButtonEvent)
	orderFromNet	 := make(chan elevio.ButtonEvent)
	orderRemove 	 := make(chan elevio.ButtonEvent)
	orderResend 	 := make(chan elevio.ButtonEvent)
	timeOutToElev 	 := make(chan bool)
	discon 			 := make(chan []string)

	go func() {
		for {
			select {
			case o := <-orderResend:
				orderRemove <- o
				orderFromElev <- o
			}
		}
	}()

	go Orderhandler.OrderHandlerFSM(id, numFloors, orderFromNet, elevFromNet, orderFromHandler, orderResend, elevInit, discon, timeOutToElev)
	go Network.Network(id, orderFromNet, orderFromElev, elevFromFSM, elevFromNet, discon)
	Elevator.ElevatorFSM(id, addr, numFloors, orderFromHandler, orderFromElev, elevFromFSM, orderRemove, elevInit, timeOutToElev)
}