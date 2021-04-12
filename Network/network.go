package Network

import (
	"fmt"

	"../Elevator/elevhandler"
	"../Elevator/elevio"

	"./bcast"
	"./peers"
)

// We define some custom struct to send over the network.
// Note that all members we want to transmit must be public.
// Any private members will be received as zero-values.

func Network(id string, orderRx chan<- elevio.ButtonEvent, orderTx <-chan elevio.ButtonEvent, elevTx <-chan elevhandler.Elevator, elevRx chan<- elevhandler.Elevator) { //change orderToElev with orderRx when orderhandler works
	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`

	//var id string

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

	go bcast.Transmitter(33334, elevTx)
	go bcast.Receiver(33334, elevRx)

	fmt.Println("Started")
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
		}

	}
}
