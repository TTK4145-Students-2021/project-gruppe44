package Network

import (
	"flag"
	"fmt"
	"os"

	"../Elevator/elevhandler"
	"../Elevator/elevio"

	"./bcast"
	"./localip"
	"./peers"
)

// We define some custom struct to send over the network.
// Note that all members we want to transmit must be public.
// Any private members will be received as zero-values.

type Elevator struct {
	ID     string
	Status elevhandler.ElevatorStatus
}

func Network(id string, orderRx chan<- elevio.ButtonEvent, orderTx <-chan elevio.ButtonEvent, elevStatus <-chan elevhandler.ElevatorStatus, elevRx chan<- Elevator) { //change orderToElev with orderRx when orderhandler works
	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`

	//var id string

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

	// We make a channel for receiving updates on the id's of the peers that are
	// alive on the network
	peerUpdateCh := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
	go peers.Transmitter(15647, id, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)

	go bcast.Transmitter(33333, orderTx)
	go bcast.Receiver(33333, orderRx)

	elevTx := make(chan Elevator)
	go bcast.Transmitter(33334, elevTx)
	go bcast.Receiver(33334, elevRx)

	costTx := make(chan int) // temp
	costRx := make(chan int)
	go bcast.Transmitter(33335, costTx)
	go bcast.Receiver(33335, costRx)

	fmt.Println("Started")
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
		case e := <-elevStatus:
			elev := Elevator{ID: id, Status: e}
			elevTx <- elev

		}

	}
}
