package Elevator

import (
	"fmt"
	"time"

	"./elevhandler"
	"./elevinit"
	"./elevio"
)

func ElevatorFSM(id string,
				 addr string,
				 numFloors int,
				 orderRecieved chan elevio.ButtonEvent,
				 orderOut chan<- elevio.ButtonEvent,
				 elevCH chan<- elevhandler.Elevator,
				 orderRemove <-chan elevio.ButtonEvent,
				 elevInit <-chan elevhandler.ElevatorStatus){
				//  timeout <-chan bool) { // uimplementert

	// "localhost:15657"
	// numFloors := 4

	elevio.Init(addr, numFloors)

	drv_floors	:= make(chan int)
	drv_obstr	:= make(chan bool)
	drv_stop	:= make(chan bool)
	drv_btn		:= make(chan elevio.ButtonEvent)

	go elevio.PollButtons(drv_btn)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	myOrders := elevhandler.Orders{Inside:	[]bool{false, false, false, false},
								   Up:		[]bool{false, false, false, false},
								   Down:	[]bool{false, false, false, false}} //FIX

	myElevator := elevhandler.ElevatorStatus{Endstation: 0,
											 Orders:	 myOrders,
											 Floor:		 0,
											 IsConnected:true,
											 Direction:	 elevio.MD_Stop}
	elevPt := &myElevator

	elevinit.InitializeElevator(addr, numFloors, drv_floors, elevPt, elevInit)

	/*
	ordersCH := make(chan elevhandler.Orders)
	go func() { //temp, skal få allOrders liste fra handler/network FIX
		for {
			time.Sleep(100 * time.Millisecond)
			ordersCH <- elevPt.Orders
		}
	}()
	go updateOrderLights(ordersCH)
	*/
	
	go func() { //only send hall orders to network
		for {
			o := <-drv_btn
			if o.Button == elevio.BT_Cab {
				orderRecieved <- o //send cab orders directly to this elevator
			} else {
				orderOut <- o
			}
		}
	}()

	go func() { //send elevator status to network
		sendRate := 50 * time.Millisecond
		for {
			time.Sleep(sendRate)
			elevCH <- elevhandler.Elevator{ID: id, Status: *elevPt}
		}

		/*
			prevElev := *elevPt
			elevCH <- elevhandler.Elevator{ID: id, Status: prevElev}
			for {
				time.Sleep(sendRate)

				if !(reflect.DeepEqual(prevElev, *elevPt)) { //burde ikke bare sende en gang, pga packet loss FIX
					prevElev = *elevPt
					elevCH <- elevhandler.Elevator{ID: id, Status: prevElev}
				}
			}
		*/
	}()
	state := "idle_state"
	fmt.Println(*elevPt)
	for {
		//time.Sleep(50 * time.Millisecond)
		switch state {
		case "idle_state":
			fmt.Println("in idle")
			state = idle(elevPt, drv_stop, orderRecieved, orderRemove)
		case "moving_up_state":
			fmt.Println("in moving up")
			state = moving(elevPt, drv_stop, drv_floors, orderRecieved, elevio.MD_Up, orderRemove)
		case "moving_down_state":
			fmt.Println("in moving down")
			state = moving(elevPt, drv_stop, drv_floors, orderRecieved, elevio.MD_Down, orderRemove)
		case "stop_up_state":
			fmt.Println("in stop up")
			state = stop(elevPt, drv_stop, drv_obstr, orderRecieved, elevio.MD_Up, orderRemove)
		case "stop_down_state":
			fmt.Println("in stop down")
			state = stop(elevPt, drv_stop, drv_obstr, orderRecieved, elevio.MD_Down, orderRemove)
		case "emergency_stop_state":
			fmt.Println("in stop")
			state = emergency_stop()
		// case "order_timeout": // FIX
		// 	fmt.Println("timeout") 
		// 	state = timeout(elevPt, timeout) // FIX
		default:
			state = idle(elevPt, drv_stop, orderRecieved, orderRemove)
		}
	}
}

func idle(elevPt *elevhandler.ElevatorStatus, stopCH <-chan bool, orderCH <-chan elevio.ButtonEvent, orderRemove <-chan elevio.ButtonEvent) string {
	elevio.SetMotorDirection(elevio.MD_Stop)
	elevPt.Direction = elevio.MD_Stop
	switch { // in case of already order
	case elevPt.Endstation < elevPt.Floor:
		return "moving_down_state"
	case elevPt.Endstation > elevPt.Floor:
		return "moving_up_state"
	case elevPt.Endstation == elevPt.Floor: //fyll ut ifs pga emergency stop
		if elevPt.Orders.Inside[elevPt.Floor] || elevPt.Orders.Down[elevPt.Floor] {
			return "stop_down_state"
		} else if elevPt.Orders.Up[elevPt.Floor] {
			return "stop_up_state"
		}
	}

	for {
		select {
		case s := <-stopCH:
			if s == true {
				return "emergency_stop_state"
			}
		case o := <-orderCH:
			elevhandler.AddOrder(elevPt, o)
			switch {
			case elevPt.Endstation < elevPt.Floor:
				return "moving_down_state"
			case elevPt.Endstation > elevPt.Floor:
				return "moving_up_state"
			case elevPt.Endstation == elevPt.Floor: //fyll ut ifs pga emergency stop
				if elevPt.Orders.Inside[elevPt.Floor] || elevPt.Orders.Down[elevPt.Floor] {
					return "stop_down_state"
				} else if elevPt.Orders.Up[elevPt.Floor] {
					return "stop_up_state"
				}
			}
		case o := <-orderRemove:
			elevhandler.RemoveOrder(elevPt, o)
		}
	}
}

func moving(elevPt *elevhandler.ElevatorStatus,
			stopCH <-chan bool,
			floorCH <-chan int,
			orderCH <-chan elevio.ButtonEvent,
			direction elevio.MotorDirection,
			orderRemove <-chan elevio.ButtonEvent) string {

	elevio.SetMotorDirection(direction)
	elevPt.Direction = direction

	for {
		select {
		case s := <-stopCH:
			if s == true {
				return "emergency_stop_state"
			}
		case f := <-floorCH:
			elevPt.Floor = f
			elevio.SetFloorIndicator(f)
			switch direction {
			case elevio.MD_Up:
				//fmt.Println("Up order check, floor: ", f)
				if elevPt.Orders.Up[f] || elevPt.Orders.Inside[f] || elevPt.Endstation <= f {
					return "stop_up_state"
				}
			case elevio.MD_Down:
				//fmt.Println("Down order check, floor: ", elevPt.Floor)
				if elevPt.Orders.Down[f] || elevPt.Orders.Inside[f] || elevPt.Endstation >= f {
					return "stop_down_state"
				}
			}
		case o := <-orderCH:
			elevhandler.AddOrder(elevPt, o)
		case o := <-orderRemove:
			elevhandler.RemoveOrder(elevPt, o)
		}
	}
}

func stop(elevPt *elevhandler.ElevatorStatus,
		  drv_stop <-chan bool,
		  drv_obstr <-chan bool,
		  orderCH <-chan elevio.ButtonEvent,
		  direction elevio.MotorDirection,
		  orderRemove <-chan elevio.ButtonEvent) string {

	elevio.SetMotorDirection(elevio.MD_Stop)
	elevio.SetDoorOpenLamp(true)

	//elevhandler.ClearOrdersAtFloor(elevPt, finishedOrder)

	elevPt.Direction = direction
	timer := time.NewTimer(3 * time.Second)

	for {
		select {
		case s := <-drv_stop:
			if s == true {
				timer.Stop()
				return "emergency_stop_state"
			}
		case o := <-drv_obstr:
			if o {
				timer.Stop()
			} else {
				timer = time.NewTimer(3 * time.Second)
			}
		case <-timer.C:
			elevhandler.ClearOrdersAtFloor(elevPt) //Quickfix, clear orders after door, else order doesn't have time to confirm first when order on same floor. FIX
			elevio.SetDoorOpenLamp(false)
			if direction == elevio.MD_Up && elevPt.Endstation > elevPt.Floor {
				return "moving_up_state"
			} else if direction == elevio.MD_Down && elevPt.Endstation < elevPt.Floor {
				return "moving_down_state"
			} else {
				return "idle_state"
			}
		case o := <-orderCH:
			elevhandler.AddOrder(elevPt, o)
		case o := <-orderRemove:
			elevhandler.RemoveOrder(elevPt, o)
		}
	}
}

func emergency_stop() string {
	return "idle_state" //fiks senere FIX
}

// FIX - do more...
// Reset all orders
func timeout(elevPt *elevhandler.ElevatorStatus, timeout <-chan bool) string {

	for{
		select{
		case t := <-timeout:
			elevPt.Timeout = t
			return "timeout"
		}
	}
}

func updateOrderLights(orders <-chan elevhandler.Orders) { // usikker på om denne skal være her FIX
	for {
		select {
		case o := <-orders:
			for f := 0; f < len(o.Inside); f++ { //var lat, gadd ikke å fikse at forskjellige order types har ferre ordre
				elevio.SetButtonLamp(elevio.BT_Cab, f, o.Inside[f])
				elevio.SetButtonLamp(elevio.BT_HallUp, f, o.Up[f])
				elevio.SetButtonLamp(elevio.BT_HallDown, f, o.Down[f])
			}
		}
	}
}
