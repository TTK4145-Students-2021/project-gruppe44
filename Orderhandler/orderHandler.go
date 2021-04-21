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

// We assume that the person waits for the assigned elevator
// The distances are complicated expressions
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


func timeoutCheck(elevMap map[string]elevhandler.ElevatorStatus, ordersPt *AllOrders, myID string, timeout chan<- string){
	timeLimit := 20 * time.Second //FIX random number

	for{
		time.Sleep(time.Second) //FIX random number
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
				} else{
					timeout <- ordersPt.Down[f].ID
				}
			}
			if (ordersPt.Up[f].ID != "") && time.Now().After(ordersPt.Up[f].TimeStarted.Add(timeLimit)){
				if elev, ok := elevMap[ordersPt.Up[f].ID]; ok{
					if time.Now().After(elev.TimeSinceNewFloor.Add(timeLimit)){
						timeout <- ordersPt.Up[f].ID
					}
				} else{
					timeout <- ordersPt.Up[f].ID
				}
			}



			/*
			if (ordersPt.Down[f].ID != "") && time.Now().After(ordersPt.Down[f].TimeStarted.Add(timeLimit)){
				timeout <- ordersPt.Down[f].ID
			}
			if (ordersPt.Up[f].ID != "") && time.Now().After(ordersPt.Up[f].TimeStarted.Add(timeLimit)){
				timeout <- ordersPt.Up[f].ID
			}
			*/
		}
	}
}

func OnTimeout(elevMap map[string]elevhandler.ElevatorStatus,
			   ordersPt *AllOrders,
			   myID string,
			   timedOut string,
			   timeoutElev chan<- bool,
			   orderResend chan<- elevio.ButtonEvent) {
	fmt.Println("in timeout")
	if timedOut == myID{
		timeoutElev <- false
	}
	OnDisconnect(elevMap, ordersPt, []string{timedOut}, orderResend)

	

	/*
	for i :=0; i < len(disconnected); i++{
		elev := elevMap[disconnected[i]]
		elev.IsConnected = false
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
	*/					
}

// When the order list is altered we will save the orders to file,
// this way we always have an updated order list in case of a crash.
// This module will have the needed functions to save the elevator status and order list.
func FileHandler(hall AllOrders, elev elevhandler.ElevatorStatus){

		allOrdersFile,_ := os.Create("Orderhandler/AllOrders.JSON")
		allOrdersJSONcontent,_ :=json.MarshalIndent(hall,"","\t")
		writeAllOrdersToJSON := bufio.NewWriter(allOrdersFile)
		writeAllOrdersToJSON.Write(allOrdersJSONcontent)
		writeAllOrdersToJSON.Flush()
		
		// -------------------------------------------- //

		elevatorFile,_ := os.Create("Orderhandler/ElevatorStatus.JSON")
		elevatorJSONcontent,_ :=json.MarshalIndent(elev,"","\t")
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

type AllOrders struct {
	Inside []bool	/** < The inside panel orders*/ //we ignore inside orders as this is handled directly by the elevator
	Up     []Order	/** < The upwards orders from outside */
	Down   []Order	/** < The downwards orders from outside */
}

var elevMap map[string]elevhandler.ElevatorStatus //Map to store all the elevator statuses

// It will receive and keep track of all orders and use a cost function to decide which elevator should take which order.
func OrderHandlerFSM(myID string,
					 newOrder <-chan elevio.ButtonEvent,
					 elev <-chan elevhandler.Elevator,
					 orderOut chan<- elevio.ButtonEvent,
					 orderResend chan<- elevio.ButtonEvent,
					 elevInit chan<- elevhandler.ElevatorStatus,
					 disconCH <-chan []string,
					 timeOutToElev chan<- bool){
	// Inputs:
	// NewOrder ButtonEvent: This is a new order that should be handled.
	// FinishedOrder ButtonEvent: This is a finished order that should be cleared.
	// confirmedOrder ButtonEvent: This is and order to be confirmed
	// Elevator struct: Includes ElevatorStatus and ElevatorID. This is used to evaluate the cost of an order on each elevator.
	// IsConnected struct: Contains a Connected bool that says if the elevator is connected and ElevatorID.
	
	// Outputs:
	// Orders struct: A list of all orders, so that the elevator can turn on/off lights.
	// NewOrder ButtonEvent: The new order, sendt to the elevator who is going to take the order.
	
	o 		  := Order{ID: "", Confirmed: false}
	AllOrders := AllOrders{	Inside: []bool{false,false,false,false},
								Up:		[]Order{o, o, o, o},
								Down:	[]Order{o, o, o, o}} //FIX: initialize in init, remove set order count
	
	ordersPt := &AllOrders
	elevMap  =  make(map[string]elevhandler.ElevatorStatus)
	// elevCH := make(chan elevhandler.ElevatorStatus)
	Init(myID, ordersPt, elevMap, elevInit)
	//go updateOrderLights(ordersPt) FIX
	elevTimedOut := make(chan string)
	go timeoutCheck(elevMap, ordersPt, myID, elevTimedOut) 
	for {
		// ordersTemp := *ordersPt
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
		// if !reflect.DeepEqual(ordersTemp, *ordersPt) {
			hallTemp := *ordersPt
			elevTemp := elevMap[myID]
			FileHandler(hallTemp, elevTemp)
		// }
	}
}
	
// When the program turns on, it will load all local data from file.
// If there is nothing to load it will initialize with zero orders.
func Init(myID string, ordersPt *AllOrders, elevMap map[string]elevhandler.ElevatorStatus, elevCH chan<- elevhandler.ElevatorStatus){// myOrders chan<- elevhandler.Orders) {
	// Load from JSON files and send values out of Filehandler as channels
	fmt.Println("in orderhandler init")
	var elevTemp elevhandler.ElevatorStatus
	elevPt := &elevTemp
	
	allOrdersContent,_ := ioutil.ReadFile("Orderhandler/AllOrders.JSON")
	json.Unmarshal(allOrdersContent, ordersPt)
	elevStatusContent,_ := ioutil.ReadFile("Orderhandler/ElevatorStatus.JSON")
	json.Unmarshal(elevStatusContent, elevPt)
	
	elevMap[myID] = elevTemp
	if elevPt.Orders.Inside == nil{ //FIX
		elevPt.Orders.Inside = []bool{false, false, false, false}
	} 
	if elevPt.Orders.Down == nil{
		elevPt.Orders.Down = []bool{false, false, false, false}
	}
	if elevPt.Orders.Up == nil{
		elevPt.Orders.Down = []bool{false, false, false, false}
	}
	elevCH <- elevTemp

	/*
	ordersTemp := elevhandler.Orders{Inside:	[]bool{false, false, false, false},
									 Up:		[]bool{false, false, false, false},
									 Down:		[]bool{false, false, false, false}} //FIX
	fmt.Println("Boot")
	*/
}


func OnDisconnect(elevMap map[string]elevhandler.ElevatorStatus,
				  ordersPt *AllOrders,
				  disconnected []string,
				  orderResend chan<- elevio.ButtonEvent) {

	for i :=0; i < len(disconnected); i++{
		elev := elevMap[disconnected[i]]
		elev.Available = false
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

// When OrderHandler receives a new order,
// this function will choose which elevator gets the order by calculating the CostFunction() on each elevator.
// It then updates its local data, OrdersAll.
// These orders are sent out for the elevator to update it’s lights.
// The elevator who got the order will send the specific order and an order confirmation as well.
// Elevators that are not connected will not be taken into consideration.
func ChooseElevator(elevMap map[string]elevhandler.ElevatorStatus,
					ordersPt *AllOrders,
					myID string,
					order elevio.ButtonEvent,
					orderOut chan<- elevio.ButtonEvent) {
	
	//TODO: save to file
	fmt.Println("Got order request")
	minCost := 1000000000000000000 //Big number so that the first cost is lower, couldn't use math.Inf(1) because of different types. Fix this
	
	//var chosenElev string
	chosenElev := myID
	//sorted ids to make sure every elevator chooses the same elev when cost is the same.
	ids := make([]string, 0, len(elevMap))
	for id := range elevMap {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	//Calculate costs and choose elevator
	for i := 0; i < len(elevMap); i++ {
		id := ids[i]
		elevStatus := elevMap[ids[i]]
		cost := CostFunction(order, elevStatus)
		if (cost < minCost) && elevStatus.Available{
			minCost = cost
			chosenElev = id
		}
	}
	//add order to list
	switch order.Button {
	case elevio.BT_HallUp:
		ordersPt.Up[order.Floor].ID = chosenElev
		ordersPt.Up[order.Floor].TimeStarted = time.Now()

	case elevio.BT_HallDown:
		ordersPt.Down[order.Floor].ID = chosenElev
		ordersPt.Down[order.Floor].TimeStarted = time.Now()
	}
	if chosenElev == myID {
		orderOut <- order
		fmt.Println("Took the order")
	} else {
		fmt.Println("Didn't take order")
	}
	//fmt.Print("Current order list: ")
	//fmt.Println(*ordersPt)
}

// When an order confirmation is recieved, this function will set that order as confirmed.
func ConfirmOrder(ordersPt *AllOrders, id string, order elevio.ButtonEvent) {
	switch order.Button {
	case elevio.BT_HallUp:
		if ordersPt.Up[order.Floor].ID == id { //unødvending if statement nå, legacy code FIX
			ordersPt.Up[order.Floor].Confirmed = true
			fmt.Println("Confirmed order")
			elevio.SetButtonLamp(elevio.BT_HallUp, order.Floor, true)
		}
	case elevio.BT_HallDown:
		if ordersPt.Down[order.Floor].ID == id {
			ordersPt.Down[order.Floor].Confirmed = true
			fmt.Println("Confirmed order")
			elevio.SetButtonLamp(elevio.BT_HallDown, order.Floor, true) // evt set lights et annet sted FIX
		}
	}
}

// When an order times out, this function will resend that order to the network module as a new order.
func ResendOrder(ordersPt *AllOrders, order elevio.ButtonEvent, orderResend chan<- elevio.ButtonEvent) {
	ClearOrder(ordersPt, order) //does it need to clear? FIX
	orderResend<-order
	fmt.Println("Resent order")
}

// When a new ElevatorStatus or Connection bool is received,
// this function will save this as local data for the CostFunction() to use.
// And if this elevator has an order that is not in the OrdersAll list, it will add this order.
func UpdateElevators(elevMap map[string]elevhandler.ElevatorStatus,
					 ordersPt *AllOrders,
					 elev elevhandler.Elevator,
					 orderResend chan<- elevio.ButtonEvent) { 

	//TODO: check if the elevator has order not in list, if yes add order.
	elevMap[elev.ID] = elev.Status

	for f := 0; f < len(ordersPt.Down); f++ {
		switch { //down orders
		case (elev.ID == ordersPt.Down[f].ID) && ordersPt.Down[f].Confirmed && !elev.Status.Orders.Down[f]: //confirmed, not taken -> order is finished
			ClearOrder(ordersPt, elevio.ButtonEvent{Button: elevio.BT_HallDown, Floor: f})

		case (elev.ID == ordersPt.Down[f].ID) && !ordersPt.Down[f].Confirmed && elev.Status.Orders.Down[f]: //not confirmed and taken -> confirm order
			ConfirmOrder(ordersPt, elev.ID, elevio.ButtonEvent{Button: elevio.BT_HallDown, Floor: f})

		case (elev.ID == ordersPt.Down[f].ID) && !ordersPt.Down[f].Confirmed && !elev.Status.Orders.Down[f]: //not confirmed, not taken -> resend if timed out?
			fmt.Println("Should resend order")
			threshold := time.Millisecond * 250 // time given to confirm order
			if time.Now().After(ordersPt.Down[f].TimeStarted.Add(threshold)){
				o := elevio.ButtonEvent{ Floor: f, Button: elevio.BT_HallDown}
				ResendOrder(ordersPt ,o, orderResend)
			}

		case (elev.ID != ordersPt.Down[f].ID) && elev.Status.Orders.Down[f]: //order taken, but not in list
			if ordersPt.Down[f].ID == "" {
				fmt.Println("Order taken without me knowing")
				ordersPt.Down[f].ID = elev.ID	//assign order maybe not confirm order aswell? FIX
				ConfirmOrder(ordersPt, elev.ID, elevio.ButtonEvent{Button: elevio.BT_HallDown, Floor: f})
			} else {
				fmt.Println("Several elevators have the same order")
				o := elevio.ButtonEvent{ Floor: f, Button: elevio.BT_HallDown}
				ResendOrder(ordersPt ,o, orderResend) //maybe not resend? FIX
			}
		}
		switch { // up orders
		case (elev.ID == ordersPt.Up[f].ID) && ordersPt.Up[f].Confirmed && !elev.Status.Orders.Up[f]: //confirmed, not taken -> order is finished
			ClearOrder(ordersPt, elevio.ButtonEvent{Button: elevio.BT_HallUp, Floor: f})

		case (elev.ID == ordersPt.Up[f].ID) && !ordersPt.Up[f].Confirmed && elev.Status.Orders.Up[f]: //not confirmed and taken -> confirm order
			ConfirmOrder(ordersPt, elev.ID, elevio.ButtonEvent{Button: elevio.BT_HallUp, Floor: f})

		case (elev.ID == ordersPt.Up[f].ID) && !ordersPt.Up[f].Confirmed && !elev.Status.Orders.Up[f]: //not confirmed, not taken -> resend if timed out?
			fmt.Println("Should resend order")
			threshold := time.Millisecond * 250 // time before resend order
			if time.Now().After(ordersPt.Up[f].TimeStarted.Add(threshold)){
				o := elevio.ButtonEvent{ Floor: f, Button: elevio.BT_HallUp}
				ResendOrder(ordersPt ,o, orderResend)
			}

		case (elev.ID != ordersPt.Up[f].ID) && elev.Status.Orders.Up[f]: //order taken, but not in list
			if ordersPt.Up[f].ID == "" {
				fmt.Println("Order taken without me knowing")
				ordersPt.Up[f].ID = elev.ID	//assign order maybe not confirm order aswell? FIX
				ConfirmOrder(ordersPt, elev.ID, elevio.ButtonEvent{Button: elevio.BT_HallUp, Floor: f})
			} else {
				fmt.Println("Several elevators have the same order")
				o := elevio.ButtonEvent{ Floor: f, Button: elevio.BT_HallUp}
				ResendOrder(ordersPt,o,orderResend)
			}
		}
	}
}

/*
func updateHallLights(AllOrders AllOrders) {
	for f := 0; f < len(o.Inside); f++ { //var lat, gadd ikke å fikse at forskjellige order types har ferre ordre
		//elevio.SetButtonLamp(elevio.BT_Cab, f, cabOrders[f])
		elevio.SetButtonLamp(elevio.BT_HallUp, f, AllOrders.Up[f].Confirmed)
		elevio.SetButtonLamp(elevio.BT_HallDown, f, AllOrders.Down[f].Confirmed)
	}
}
*/

// When an old order is finished, this function will clear/update the order table.
func ClearOrder(ordersPt *AllOrders, order elevio.ButtonEvent) {
	//TODO: save order list to file
	elevio.SetButtonLamp(order.Button, order.Floor, false)
	switch order.Button {
	case elevio.BT_HallUp:
		ordersPt.Up[order.Floor].ID = ""
		ordersPt.Up[order.Floor].Confirmed = false
	case elevio.BT_HallDown:
		ordersPt.Down[order.Floor].ID = ""
		ordersPt.Down[order.Floor].Confirmed = false
	}
	fmt.Println("Cleared order")
	//fmt.Print("Current order list: ")
	//fmt.Println(*ordersPt)
}