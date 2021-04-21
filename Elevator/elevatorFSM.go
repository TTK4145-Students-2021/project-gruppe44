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

	elevio.Init(addr, numFloors)

	drv_floors := make(chan int)
	drv_obstr  := make(chan bool)
	drv_stop   := make(chan bool)
	drv_btn	   := make(chan elevio.ButtonEvent)

	go elevio.PollButtons(drv_btn)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	var myElevator elevhandler.ElevatorStatus
	elevPt := &myElevator

	elevinit.InitializeElevator(addr, numFloors, drv_floors, elevPt, elevInit)
	
	// Only send hall orders to network
	go func() {
		for {
			o := <-drv_btn
			if o.Button == elevio.BT_Cab {
				orderRecieved <- o // Send cab orders directly to this elevator
			} else {
				orderOut <- o
			}
		}
	}()

	// Send elevator status to network
	go func() {
		sendRate := 50 * time.Millisecond
		for {
			elevCH <- elevhandler.Elevator{ID: id, Status: *elevPt}
			time.Sleep(sendRate)
		}
	}()

	doorOpen	:= make(chan bool)
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
			elevPt.State	 = elevhandler.ST_MovingDown
			fmt.Println("State: MovingDown")
		
		case elevPt.Endstation > elevPt.Floor:
			elevio.SetMotorDirection(elevio.MD_Up)
			elevPt.Direction = elevio.MD_Up
			elevPt.State	 = elevhandler.ST_MovingUp
			fmt.Println("State: MovingUp")
		
			case elevPt.Endstation == elevPt.Floor:
			if elevPt.Orders.Inside[elevPt.Floor] || elevPt.Orders.Down[elevPt.Floor] {
				elevPt.Direction = elevio.MD_Down
				elevPt.State	 = elevhandler.ST_StopDown
				fmt.Println("State: StopDown")
			} else if elevPt.Orders.Up[elevPt.Floor] {
				elevPt.Direction = elevio.MD_Up
				elevPt.State	 = elevhandler.ST_StopUp
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
	
	elevPt.Floor			 = floor
	elevPt.TimeSinceNewFloor = time.Now()
	elevPt.Available		 = true
	
	elevio.SetFloorIndicator(floor)
	switch elevPt.State {
	case elevhandler.ST_MovingUp:
		if elevPt.Orders.Up[floor] || elevPt.Orders.Inside[floor] || elevPt.Endstation <= floor {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevPt.Direction = elevio.MD_Up
			elevPt.State	 = elevhandler.ST_StopUp
			fmt.Println("State: StopUp")
		}
	case elevhandler.ST_MovingDown:
		if elevPt.Orders.Down[floor] || elevPt.Orders.Inside[floor] || elevPt.Endstation >= floor {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevPt.Direction = elevio.MD_Down
			elevPt.State	 = elevhandler.ST_StopDown
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
	elevhandler.ClearOrdersAtFloor(elevPt)
	elevPt.Available = true
	elevio.SetDoorOpenLamp(false)
	switch{
	case elevPt.Endstation == elevPt.Floor:
		fmt.Println("Endstation on floor")
		elevio.SetMotorDirection(elevio.MD_Stop)
		elevPt.Direction = elevio.MD_Stop
		elevPt.State	 = elevhandler.ST_Idle
		fmt.Println("State: Idle")

	case elevPt.Endstation > elevPt.Floor:
		fmt.Println("Endstation above floor")
 		elevio.SetMotorDirection(elevio.MD_Up)
		elevPt.Direction = elevio.MD_Up
		elevPt.State	 = elevhandler.ST_MovingUp
		fmt.Println("State: MovingUp")

	case elevPt.Endstation < elevPt.Floor:
		fmt.Println("Endstation below floor")
		elevio.SetMotorDirection(elevio.MD_Down)
		elevPt.Direction = elevio.MD_Down
		elevPt.State	 = elevhandler.ST_MovingDown
		fmt.Println("State: MovingDown")
	}
}

func onTimeout(elevPt *elevhandler.ElevatorStatus){
	fmt.Println("onTimeout")
	elevPt.Available = false
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