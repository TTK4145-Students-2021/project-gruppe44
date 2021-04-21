package Orderhandler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"sort"
	"time"

	"../Elevator/elevhandler"
	"../Elevator/elevio"
)

// Prioritize elevators already moving
func CostFunction(orderReq elevio.ButtonEvent, elevStatus elevhandler.ElevatorStatus) int {
	distToRequest				:= DistanceBetweenFloors(elevStatus.Floor, orderReq.Floor)
	distToEndstation			:= DistanceBetweenFloors(elevStatus.Floor, elevStatus.Endstation)
	distFromEndstationToRequest	:= DistanceBetweenFloors(elevStatus.Endstation, orderReq.Floor)
	distTotal					:= distFromEndstationToRequest + distToEndstation

	// Boolean expressions:
	elevFloorUNDER	:= (elevStatus.Floor < orderReq.Floor)
	elevFloorOVER	:= (elevStatus.Floor > orderReq.Floor)
	elevDirUP		:= (elevStatus.Direction == elevio.MD_Up)
	elevDirDOWN		:= (elevStatus.Direction == elevio.MD_Down)
	orderReqUP		:= (orderReq.Button == elevio.BT_HallUp)
	orderReqDOWN	:= (orderReq.Button == elevio.BT_HallDown)
	endStationUNDER	:= (elevStatus.Endstation < orderReq.Floor)
	endStationOVER	:= (elevStatus.Endstation > orderReq.Floor)

	if elevFloorUNDER {
		if elevDirUP && orderReqUP {

			if endStationUNDER {
				return distToRequest
			} else {
				return distToRequest - 1
			}

		} else if elevDirUP && orderReqDOWN {

			if endStationUNDER {
				return distToRequest
			} else {
				return distTotal
			}

		} else {
			return distTotal
		}

	} else if elevFloorOVER {
		if elevDirDOWN && orderReqUP {

			if endStationOVER {
				return distToRequest
			} else {
				return distTotal
			}

		} else if elevDirDOWN && orderReqDOWN {

			if endStationOVER {
				return distToRequest
			} else {
				return distToRequest - 1
			}

		} else {
			return distTotal
		}

	} else {
		if (elevDirUP && orderReqUP) || (elevDirDOWN && orderReqDOWN) {
			return -1
		} else {
			return distTotal
		}
	}
}

func DistanceBetweenFloors(floor1, floor2 int) int {
	return int(math.Abs(float64(floor1) - float64(floor2)))
}

func TimeoutCheck(elevMap map[string]elevhandler.ElevatorStatus,
				  ordersPt *HallOrders,
				  myID string,
				  timeout chan<- string){
	
	timeLimit := 20 * time.Second

	for{
		time.Sleep(time.Second)
		for f := 0; f < len(ordersPt.Down); f++{
			
			myElev := elevMap[myID]
			
			if myElev.Orders.Inside[f]{
				if time.Now().After(myElev.TimeSinceNewFloor.Add(timeLimit)){
					timeout <- myID
				}
			}

			if (ordersPt.Down[f].ID != "") && time.Now().After(ordersPt.Down[f].TimeStarted.Add(timeLimit)){
				if elev, ok := elevMap[ordersPt.Down[f].ID]; ok{
					if time.Now().After(elev.TimeSinceNewFloor.Add(timeLimit)){
						timeout <- ordersPt.Down[f].ID
					}
				} else {
					timeout <- ordersPt.Down[f].ID
				}
			}
			
			if (ordersPt.Up[f].ID != "") && time.Now().After(ordersPt.Up[f].TimeStarted.Add(timeLimit)){
				if elev, ok := elevMap[ordersPt.Up[f].ID]; ok{
					if time.Now().After(elev.TimeSinceNewFloor.Add(timeLimit)){
						timeout <- ordersPt.Up[f].ID
					}
				} else {
					timeout <- ordersPt.Up[f].ID
				}
			}
		}
	}
}

func OnTimeout(elevMap map[string]elevhandler.ElevatorStatus,
			   ordersPt *HallOrders,
			   myID string,
			   timedOut string,
			   timeoutElev chan<- bool,
			   orderResend chan<- elevio.ButtonEvent) {
	
	fmt.Println("In timeout")
	if timedOut == myID { timeoutElev <- false }

	OnDisconnect(elevMap, ordersPt, []string{timedOut}, orderResend)	
}

// When the order list is altered we will save the orders to file,
// this way we always have an updated order list in case of a crash.
func SaveToFile(hall HallOrders, elev elevhandler.ElevatorStatus){

		hallOrdersFile,_		:= os.Create("Orderhandler/HallOrders.JSON")
		hallOrdersJSONcontent,_	:= json.MarshalIndent(hall,"","\t")
		writeHallOrdersToJSON	:= bufio.NewWriter(hallOrdersFile)
		
		writeHallOrdersToJSON.Write(hallOrdersJSONcontent)
		writeHallOrdersToJSON.Flush()
		
		// -------------------------------------------- //

		elevatorFile,_			  := os.Create("Orderhandler/ElevatorStatus.JSON")
		elevatorJSONcontent,_	  := json.MarshalIndent(elev,"","\t")
		writeElevatorStatusToJSON := bufio.NewWriter(elevatorFile)
		
		writeElevatorStatusToJSON.Write(elevatorJSONcontent)
		writeElevatorStatusToJSON.Flush()
}


/************** OrderHandler **************/

type Order struct {
	ID			string		// ID of elevator who has the order, empty string if no elevator
	Confirmed	bool		// True if confirmed, false if else
	TimeStarted	time.Time	// Used for timeout flag
}

type HallOrders struct {
	Up	 []Order	// The upwards orders from outside
	Down []Order	// The downwards orders from outside
}

// OHFSM will receive and keep track of all orders and use a cost function to decide which elevator should take which order.
func OrderHandlerFSM(myID string,
					 numFloors int,
					 newOrder <-chan elevio.ButtonEvent,
					 elev <-chan elevhandler.Elevator,
					 orderOut chan<- elevio.ButtonEvent,
					 orderResend chan<- elevio.ButtonEvent,
					 elevInit chan<- elevhandler.ElevatorStatus,
					 disconCH <-chan []string,
					 timeOutToElev chan<- bool){
	
	var hallOrders  HallOrders 
	ordersPt	 := &hallOrders
	elevMap		 := make(map[string]elevhandler.ElevatorStatus)
	elevTimedOut := make(chan string)

	LoadFromFile(myID, numFloors, ordersPt, elevMap, elevInit)
	go TimeoutCheck(elevMap, ordersPt, myID, elevTimedOut)
	
	for {
		select {
		case t := <-elevTimedOut:
			OnTimeout(elevMap, ordersPt, myID, t, timeOutToElev, orderResend)
		case o := <-newOrder:
			ChooseElevator(elevMap, ordersPt, myID, o, orderOut)
		case e := <-elev:
			UpdateElevators(elevMap, ordersPt, e, orderResend)
		case d := <-disconCH:
			OnDisconnect(elevMap, ordersPt, d, orderResend)
		}

		hallTemp := *ordersPt
		elevTemp := elevMap[myID]
		SaveToFile(hallTemp, elevTemp)
	}
}
	
// When the program turns on, it will load all local data from file.
// If there is nothing to load it will initialize with zero orders.
func LoadFromFile(myID string,
				  numFloors int,
				  ordersPt *HallOrders,
				  elevMap map[string]elevhandler.ElevatorStatus,
				  elevCH chan<- elevhandler.ElevatorStatus){
	
	// Load from JSON files and send values out of Filehandler as channels
	fmt.Println("In orderhandler init")
	
	var elevTemp elevhandler.ElevatorStatus
	elevPt				  := &elevTemp
	hallOrdersContent, err := ioutil.ReadFile("Orderhandler/HallOrders.JSON")
	
	if err != nil{
		
		o := Order{ID: "", Confirmed: false}
		var hallOrders HallOrders
		
		// Create blank orders struct with numFloors floors
		for i := 0; i < numFloors; i++{
			hallOrders.Up	= append(hallOrders.Up, o)
			hallOrders.Down	= append(hallOrders.Down, o)
		}
		*ordersPt = hallOrders
	} else {
		json.Unmarshal(hallOrdersContent, ordersPt)
	}
	
	elevStatusContent, err := ioutil.ReadFile("Orderhandler/ElevatorStatus.JSON")
	if err != nil{
		var myOrders elevhandler.Orders

		// Create blank orders struct with numFloors floors
		for i := 0; i < numFloors; i++{
			myOrders.Inside	= append(myOrders.Inside, false)
			myOrders.Up		= append(myOrders.Up, false)
			myOrders.Down	= append(myOrders.Down, false)
		}
		elevTemp.Orders = myOrders
	} else {
		json.Unmarshal(elevStatusContent, elevPt)
	}

	elevMap[myID] = elevTemp
	elevCH <- elevTemp
}

func OnDisconnect(elevMap map[string]elevhandler.ElevatorStatus,
				  ordersPt *HallOrders,
				  disconnected []string,
				  orderResend chan<- elevio.ButtonEvent) {

	for i :=0; i < len(disconnected); i++{
		
		elev					 := elevMap[disconnected[i]]
		elev.Available			 = false
		elevMap[disconnected[i]] = elev
		
		for f := 0; f < len(ordersPt.Down); f++ {
			if disconnected[i] == ordersPt.Down[f].ID {
				o := elevio.ButtonEvent{ Floor: f, Button: elevio.BT_HallDown}
				ResendOrder(ordersPt, o, orderResend)
			}
			if disconnected[i] == ordersPt.Up[f].ID {
				o := elevio.ButtonEvent{ Floor: f, Button: elevio.BT_HallUp}
				ResendOrder(ordersPt, o, orderResend)
			}
		}
	}					
}

func ChooseElevator(elevMap map[string]elevhandler.ElevatorStatus,
					ordersPt *HallOrders,
					myID string,
					order elevio.ButtonEvent,
					orderOut chan<- elevio.ButtonEvent) {
	
	fmt.Println("Got order request")
	
	minCost	   := 1000000000000000000
	chosenElev := myID

	// Sorted ids to make sure every elevator chooses the same elev when cost is the same.
	ids := make([]string, 0, len(elevMap))
	for id := range elevMap {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	//Calculate costs and choose elevator
	for i := 0; i < len(elevMap); i++ {
		
		id		   := ids[i]
		elevStatus := elevMap[ids[i]]
		cost	   := CostFunction(order, elevStatus)
		
		if (cost < minCost) && elevStatus.Available{
			minCost	   = cost
			chosenElev = id
		}
	}

	// Add order to order list
	switch order.Button {
	case elevio.BT_HallUp:
		ordersPt.Up[order.Floor].ID			 = chosenElev
		ordersPt.Up[order.Floor].TimeStarted = time.Now()

	case elevio.BT_HallDown:
		ordersPt.Down[order.Floor].ID		   = chosenElev
		ordersPt.Down[order.Floor].TimeStarted = time.Now()
	}
	if chosenElev == myID {
		orderOut <- order
		fmt.Println("Took the order")
	} else {
		fmt.Println("Didn't take order")
	}
}

// When an order confirmation is recieved, this function will set that order as confirmed.
func ConfirmOrder(ordersPt *HallOrders, id string, order elevio.ButtonEvent) {
	switch order.Button {
	case elevio.BT_HallUp:
		if ordersPt.Up[order.Floor].ID == id {
			ordersPt.Up[order.Floor].Confirmed = true
			fmt.Println("Confirmed order")
			elevio.SetButtonLamp(elevio.BT_HallUp, order.Floor, true)
		}
	case elevio.BT_HallDown:
		if ordersPt.Down[order.Floor].ID == id {
			ordersPt.Down[order.Floor].Confirmed = true
			fmt.Println("Confirmed order")
			elevio.SetButtonLamp(elevio.BT_HallDown, order.Floor, true)
		}
	}
}

// When an order times out, this function will resend that order to the network module as a new order.
func ResendOrder(ordersPt *HallOrders, order elevio.ButtonEvent, orderResend chan<- elevio.ButtonEvent) {
	ClearOrder(ordersPt, order)
	orderResend <-order
	fmt.Println("Resent order")
}

// When a new ElevatorStatus or Connection bool is received,
// this function will save this as local data for the CostFunction() to use.
// And if this elevator has an order that is not in the OrdersAll list, it will add this order.
func UpdateElevators(elevMap map[string]elevhandler.ElevatorStatus,
					 ordersPt *HallOrders,
					 elev elevhandler.Elevator,
					 orderResend chan<- elevio.ButtonEvent) { 

	elevMap[elev.ID] = elev.Status

	for f := 0; f < len(ordersPt.Down); f++ {
		// Down orders
		switch {
		
		// Confirmed, not taken -> order is finished
		case (elev.ID == ordersPt.Down[f].ID) && ordersPt.Down[f].Confirmed && !elev.Status.Orders.Down[f]:
			ClearOrder(ordersPt, elevio.ButtonEvent{Button: elevio.BT_HallDown, Floor: f})

		// Not confirmed and taken -> confirm order
		case (elev.ID == ordersPt.Down[f].ID) && !ordersPt.Down[f].Confirmed && elev.Status.Orders.Down[f]:
			ConfirmOrder(ordersPt, elev.ID, elevio.ButtonEvent{Button: elevio.BT_HallDown, Floor: f})

		// Not confirmed, not taken -> resend if timed out
		case (elev.ID == ordersPt.Down[f].ID) && !ordersPt.Down[f].Confirmed && !elev.Status.Orders.Down[f]:
			fmt.Println("Should resend order")
			threshold := time.Millisecond * 250 // Time given to confirm order
			if time.Now().After(ordersPt.Down[f].TimeStarted.Add(threshold)){
				o := elevio.ButtonEvent{ Floor: f, Button: elevio.BT_HallDown}
				ResendOrder(ordersPt ,o, orderResend)
			}
		
		// Order taken, but not in list
		case (elev.ID != ordersPt.Down[f].ID) && elev.Status.Orders.Down[f]:
			if ordersPt.Down[f].ID == "" {
				fmt.Println("Order taken without me knowing")
				ordersPt.Down[f].ID = elev.ID
				ConfirmOrder(ordersPt, elev.ID, elevio.ButtonEvent{Button: elevio.BT_HallDown, Floor: f})
			} else {
				fmt.Println("Several elevators have the same order")
				o := elevio.ButtonEvent{ Floor: f, Button: elevio.BT_HallDown}
				ResendOrder(ordersPt ,o, orderResend)
			}
		}

		// Up orders
		switch {

		// Confirmed, not taken -> order is finished
		case (elev.ID == ordersPt.Up[f].ID) && ordersPt.Up[f].Confirmed && !elev.Status.Orders.Up[f]:
			ClearOrder(ordersPt, elevio.ButtonEvent{Button: elevio.BT_HallUp, Floor: f})
		
		// Not confirmed and taken -> confirm order
		case (elev.ID == ordersPt.Up[f].ID) && !ordersPt.Up[f].Confirmed && elev.Status.Orders.Up[f]:
			ConfirmOrder(ordersPt, elev.ID, elevio.ButtonEvent{Button: elevio.BT_HallUp, Floor: f})

		// Not confirmed, not taken -> resend if timed out
		case (elev.ID == ordersPt.Up[f].ID) && !ordersPt.Up[f].Confirmed && !elev.Status.Orders.Up[f]:
			fmt.Println("Should resend order")
			threshold := time.Millisecond * 250 // Time before resend order
			if time.Now().After(ordersPt.Up[f].TimeStarted.Add(threshold)){
				o := elevio.ButtonEvent{ Floor: f, Button: elevio.BT_HallUp}
				ResendOrder(ordersPt ,o, orderResend)
			}

		// Order taken, but not in list
		case (elev.ID != ordersPt.Up[f].ID) && elev.Status.Orders.Up[f]:
			if ordersPt.Up[f].ID == "" {
				fmt.Println("Order taken without me knowing")
				ordersPt.Up[f].ID = elev.ID
				ConfirmOrder(ordersPt, elev.ID, elevio.ButtonEvent{Button: elevio.BT_HallUp, Floor: f})
			} else {
				fmt.Println("Several elevators have the same order")
				o := elevio.ButtonEvent{ Floor: f, Button: elevio.BT_HallUp}
				ResendOrder(ordersPt,o,orderResend)
			}
		}
	}
}

// When an old order is finished, this function will clear/update the order table.
func ClearOrder(ordersPt *HallOrders, order elevio.ButtonEvent) {
	elevio.SetButtonLamp(order.Button, order.Floor, false)
	switch order.Button {
	case elevio.BT_HallUp:
		ordersPt.Up[order.Floor].ID		   = ""
		ordersPt.Up[order.Floor].Confirmed = false
	case elevio.BT_HallDown:
		ordersPt.Down[order.Floor].ID		 = ""
		ordersPt.Down[order.Floor].Confirmed = false
	}
	fmt.Println("Cleared order")
}