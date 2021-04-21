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
				 elevInit <-chan elevhandler.ElevatorStatus,
				 timeOutToElev <-chan bool){

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
											 Available:  true,
											 Direction:	 elevio.MD_Stop}
	elevPt := &myElevator

	elevinit.InitializeElevator(addr, numFloors, drv_floors, elevPt, elevInit)

	/*
	ordersCH := make(chan elevhandler.Orders)
	go func() { //temp, skal få hallOrders liste fra handler/network FIX
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
			elevCH <- elevhandler.Elevator{ID: id, Status: *elevPt}
			time.Sleep(sendRate)
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
	doorOpen := make(chan bool)
	doorTimeout := make(chan bool)
	for {
		select {
		case <- doorOpen:
			fmt.Println("State: DoorOpen")
			elevPt.State = elevhandler.ST_DoorOpen
		case <- timeOutToElev:
			go onTimeout(elevPt)
		case f := <- drv_floors:
			onFloorSensor(elevPt, f)
			go doorTimer(elevPt.State, drv_obstr, doorTimeout, doorOpen)
		case <-doorTimeout:
			onDoorTimeout(elevPt)
		case o := <-orderRecieved:
			onNewOrder(elevPt, o)
			go doorTimer(elevPt.State, drv_obstr, doorTimeout, doorOpen)
		case o := <-orderRemove:
			onRemoveOrder(elevPt, o)
		}
	}
}

func onNewOrder(elevPt *elevhandler.ElevatorStatus, order elevio.ButtonEvent){
	fmt.Println("onNewOrder")
	elevhandler.AddOrder(elevPt, order)
	switch elevPt.State {
	case elevhandler.ST_Idle:
		switch {
		case elevPt.Endstation < elevPt.Floor:
			elevio.SetMotorDirection(elevio.MD_Down)
			elevPt.Direction = elevio.MD_Down
			elevPt.State = elevhandler.ST_MovingDown
			fmt.Println("State: MovingDown")
		case elevPt.Endstation > elevPt.Floor:
			elevio.SetMotorDirection(elevio.MD_Up)
			elevPt.Direction = elevio.MD_Up
			elevPt.State = elevhandler.ST_MovingUp
			fmt.Println("State: MovingUp")
		case elevPt.Endstation == elevPt.Floor: //fyll ut ifs pga emergency stop
			if elevPt.Orders.Inside[elevPt.Floor] || elevPt.Orders.Down[elevPt.Floor] {
				elevPt.Direction = elevio.MD_Down
				elevPt.State = elevhandler.ST_StopDown
				fmt.Println("State: StopDown")
			} else if elevPt.Orders.Up[elevPt.Floor] {
				elevPt.Direction = elevio.MD_Up
				elevPt.State = elevhandler.ST_StopUp
				fmt.Println("State: StopUp")
			}
		}
	}
}

func onRemoveOrder(elevPt *elevhandler.ElevatorStatus, order elevio.ButtonEvent){
	fmt.Println("onRemoveOrder")
	elevhandler.RemoveOrder(elevPt, order)
}

func onFloorSensor(elevPt *elevhandler.ElevatorStatus, floor int){
	fmt.Println("onFloorSensor")
	elevPt.Floor = floor
	elevPt.TimeSinceNewFloor = time.Now()
	elevPt.Available = true
	/*
	if !elevPt.Available {
		elevPt.Available = true
		elevio.SetMotorDirection(elevio.MD_Stop)
		elevPt.Direction = elevio.MD_Stop
		elevPt.State = elevhandler.ST_Idle
	}
	*/
	elevio.SetFloorIndicator(floor)
	switch elevPt.State {
	case elevhandler.ST_MovingUp:
		//fmt.Println("Up order check, floor: ", f)
		if elevPt.Orders.Up[floor] || elevPt.Orders.Inside[floor] || elevPt.Endstation <= floor {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevPt.Direction = elevio.MD_Up
			elevPt.State = elevhandler.ST_StopUp
			fmt.Println("State: StopUp")
		}
	case elevhandler.ST_MovingDown:
		//fmt.Println("Down order check, floor: ", elevPt.Floor)
		if elevPt.Orders.Down[floor] || elevPt.Orders.Inside[floor] || elevPt.Endstation >= floor {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevPt.Direction = elevio.MD_Down
			elevPt.State = elevhandler.ST_StopDown
			fmt.Println("State: StopDown")
		}
	case elevhandler.ST_Idle:
		elevio.SetMotorDirection(elevio.MD_Stop)
	}
}

func doorTimer(elevState elevhandler.ElevatorState, obstr <-chan bool, finished chan<- bool, doorOpen chan<- bool){
	if (elevState == elevhandler.ST_StopUp) || (elevState == elevhandler.ST_StopDown){
		doorOpen <- true
		elevio.SetDoorOpenLamp(true)
		fmt.Println("onDoorTimer")
		timer := time.NewTimer(3 * time.Second)
			for {
				select{
				case o := <-obstr:
					if o {
						timer.Stop()
					} else {
						timer = time.NewTimer(3 * time.Second)
					}
				case <-timer.C:
					finished <- true
					return
			}
		}
	}
}

func onDoorTimeout(elevPt *elevhandler.ElevatorStatus){
	fmt.Println("onDoorTimeout")
	elevhandler.ClearOrdersAtFloor(elevPt) //Quickfix, clear orders after door, else order doesn't have time to confirm first when order on same floor. FIX
	elevPt.Available = true
	elevio.SetDoorOpenLamp(false)
	switch{
	case elevPt.Endstation == elevPt.Floor:
		fmt.Println("Endstation on floor")
		elevio.SetMotorDirection(elevio.MD_Stop)
		elevPt.Direction = elevio.MD_Stop
		elevPt.State = elevhandler.ST_Idle
		fmt.Println("State: Idle")
	case elevPt.Endstation > elevPt.Floor: //elevPt.Direction == elevio.MD_Up &&
		fmt.Println("Endstation above floor")
 		elevio.SetMotorDirection(elevio.MD_Up)
		elevPt.Direction = elevio.MD_Up
		elevPt.State = elevhandler.ST_MovingUp
		fmt.Println("State: MovingUp")
	case elevPt.Endstation < elevPt.Floor: //elevPt.Direction == elevio.MD_Down &&
		fmt.Println("Endstation below floor")
		elevio.SetMotorDirection(elevio.MD_Down)
		elevPt.Direction = elevio.MD_Down
		elevPt.State = elevhandler.ST_MovingDown
		fmt.Println("State: MovingDown")
	}
}

func onTimeout(elevPt *elevhandler.ElevatorStatus){
	fmt.Println("onTimeout")
	elevPt.Available = false
	//elevPt.State = elevhandler.ST_TimedOut
	checkRate := 100*time.Millisecond
	for !elevPt.Available{
		time.Sleep(checkRate)
		switch elevPt.State{
		case elevhandler.ST_Idle:
			if elevPt.Floor == 0 {
				elevio.SetMotorDirection(elevio.MD_Up)
			} else {
				elevio.SetMotorDirection(elevio.MD_Down)
			}
		case elevhandler.ST_MovingUp:
			elevio.SetMotorDirection(elevio.MD_Up)

		case elevhandler.ST_MovingDown:
			elevio.SetMotorDirection(elevio.MD_Down)
		case elevhandler.ST_StopDown, elevhandler.ST_StopUp:
			return
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
