package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"./network/bcast"
	"./network/localip"
	"./network/peers"
)

// We define some custom struct to send over the network.
// Note that all members we want to transmit must be public.
// Any private members will be received as zero-values.
type HelloMsg struct {
	Message string
	Iter    int
}

//Order for
type Order struct {
	Floor     int
	Direction int // 0 for directionless ( inside order) 1 for up, 2 for down (outside orders)

}

//ElevatorStatus contains array of orders, current floor and direction
type ElevatorStatus struct {
	CurrentOrders []Order
	Direction     int
	CurrentFloor  int
}

type SendOrder struct {
	OrderInformation Order
	SenderId         string //id of elevator sending the order
	RecieverId       string //use id of elevator that should take the order, or 0 for everyone to compare their orders
}
type SendCost struct {
	cost     int
	SenderId string
}

func costFunction(order Order, elevatorStatus ElevatorStatus) int {
	return rand.Intn(1000) //return random rumber as temp cost function
	//return Abs(order.Floor - elevatorStatus.CurrentFloor)
}

func main() {
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

	// We make a channel for receiving updates on the id's of the peers that are
	// alive on the network
	peerUpdateCh := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
	go peers.Transmitter(15647, id, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)

	// We make channels for sending and receiving our custom data types
	helloTx := make(chan HelloMsg)
	helloRx := make(chan HelloMsg)
	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	// start multiple transmitters/receivers on the same port.
	go bcast.Transmitter(16569, helloTx)
	go bcast.Receiver(16569, helloRx)

	orderTx := make(chan Order)
	orderRx := make(chan Order)
	go bcast.Transmitter(33333, orderTx)
	go bcast.Receiver(33333, orderRx)

	costTx := make(chan int)
	costRx := make(chan int)
	go bcast.Transmitter(33334, costTx)
	go bcast.Receiver(33334, costRx)

	// The example message. We just send one of these every second.
	go func() {
		/*
			helloMsg := HelloMsg{"Hello from " + id, 0}
			for {
				helloMsg.Iter++
				helloTx <- helloMsg
				time.Sleep(1 * time.Second)
			}
		*/
		time.Sleep(3 * time.Second) //make orders
		order := Order{0, 1}
		for {
			order.Floor++
			orderTx <- order
			time.Sleep(10 * time.Second)
		}
		//sendOrder := SendOrder{order, "333"}
	}()
	//order1 := Order{4, 0}
	orders1 := make([]Order, 0) //lager en dummy elevator status
	thisElevator := ElevatorStatus{orders1, 0, 1}
	fmt.Println("Started")
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		case a := <-helloRx:
			fmt.Printf("Received: %#v\n", a)

		case o := <-orderRx:
			fmt.Printf("Recieved request to send cost function from" + id + "\n") //an order happens everyone do this
			myCost := costFunction(o, thisElevator)
			costTx <- myCost
			fmt.Printf("my cost is: %v\n", myCost)
			takeThisOrder := true
			//time.Sleep(3 * time.Second) // give time for everyone to send their cost
		L:
			for {
				select {
				case recievedCost := <-costRx:
					fmt.Printf("recieved cost: %v\n", recievedCost)
					if recievedCost < myCost {
						takeThisOrder = false
					}
				default:
					break L
				}
			}
			/*
				for len(costRx) > 0 {
					recievedCost := <-costRx
					fmt.Printf("recieved cost: %v\n", recievedCost)
					if recievedCost < myCost {
						takeThisOrder = false
					}
				}
			*/
			if takeThisOrder {
				thisElevator.CurrentOrders = append(thisElevator.CurrentOrders, o) //add order if cost is smallest
				fmt.Printf("Took order: %#v\n", o)
				//fmt.Printf("current status:  %#v\n", thisElevator)
				//send confirmation to others?
			} else {
				fmt.Printf("didn't take order\n")
			}
		}
		/*
			case o := <-orderRx:
				if o.RecieverId == id {
					fmt.Printf("Received my order %#v\n", o) // add to order list
				} else if o.RecieverId == "" {
					fmt.Printf("Recieved request to send cost function %#v\n", o)
					myCost := costFunction(o.OrderInformation, thisElevator)
					costTx <- myCost
					takeThisOrder := true
					time.Sleep(1 * time.Second) // give time for everyone to send their cost
					for len(costRx) > 0 {
						recievedCost := <-costRx
						if recievedCost < myCost {
							takeThisOrder = false
						}
					}
					if takeThisOrder {
						thisElevator.CurrentOrders = append(thisElevator.CurrentOrders, o.OrderInformation) //add order if cost is smallest
						fmt.Printf("Took order: %#v\n", o)
						fmt.Printf("current status:  %#v\n", thisElevator)
						//send confirmation to others?

					}

				} else {
					fmt.Printf("Recieved someone elses order %#v\n", o)

				}
			}
		*/
	}
}
