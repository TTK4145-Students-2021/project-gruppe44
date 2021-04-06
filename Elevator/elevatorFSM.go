package main

import (
	"fmt"
	"time"

	"./elevhandler"
	"./elevio"
)

/*
	#define up_direction 1
	#define down_direction 0
*/

//var elevator elevhandler.ElevatorStatus

func main() {
	numFloors := 4

	elevio.Init("localhost:15657", numFloors)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	//floorCH := make(chan int)
	//directionCH := make(chan elevio.MotorDirection)
	//clearCH := make(chan int)
	//elevatorCH := make(chan elevhandler.ElevatorStatus)
	ordersCH := make(chan elevhandler.Orders)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	//go elevio.PollFloorSensor(floorCH)

	myOrders := elevhandler.Orders{Inside: []bool{false, false, false, false}, Up: []bool{false, false, false, false}, Down: []bool{false, false, false, false}}
	myElevator := elevhandler.ElevatorStatus{Endstation: 0, Orders: myOrders, Floor: 0, Direction: elevio.MD_Stop}
	elevPt := &myElevator
	//go elevhandler.ElevatorStatusUpdateForever(elevPt, drv_buttons, directionCH, floorCH, clearCH, elevatorCH, ordersCH)
	go func() { //temp, skal få orders liste fra handler/network
		for {
			time.Sleep(50 * time.Millisecond)
			ordersCH <- elevPt.Orders
		}

	}()
	go updateOrderLights(ordersCH)
	//go fix lights elns
	state := "idle_state"
	for {
		switch state {
		case "idle_state":
			fmt.Println("in idle")
			state = idle(elevPt, drv_stop, drv_buttons)
		case "moving_up_state":
			fmt.Println("in moving up")
			state = moving(elevPt, drv_stop, drv_floors, drv_buttons, elevio.MD_Up)
		case "moving_down_state":
			fmt.Println("in moving down")
			state = moving(elevPt, drv_stop, drv_floors, drv_buttons, elevio.MD_Down)
		case "stop_up_state":
			fmt.Println("in stop up")
			state = stop(elevPt, drv_stop, drv_obstr, drv_buttons, elevio.MD_Up)
		case "stop_down_state":
			fmt.Println("in stop down")
			state = stop(elevPt, drv_stop, drv_obstr, drv_buttons, elevio.MD_Down)
		case "emergency_stop_state":
			fmt.Println("in stop")
			state = emergency_stop()
		default:
			state = idle(elevPt, drv_stop, drv_buttons)
		}

	}
}

func idle(elevPt *elevhandler.ElevatorStatus, stopCH <-chan bool, orderCH <-chan elevio.ButtonEvent) string {
	elevio.SetMotorDirection(elevio.MD_Stop)
	elevPt.Direction = elevio.MD_Stop
	for {
		select {
		case s := <-stopCH:
			if s == true {
				return "emergency_stop_state"
			}
		case o := <-orderCH:
			elevhandler.ElevatorAddOrder(elevPt, o)
		default:
		}
		elevhandler.ElevatorSetEndstation(elevPt)
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
		default:
		}
	}
}

func moving(elevPt *elevhandler.ElevatorStatus, stopCH <-chan bool, floorCH <-chan int, orderCH <-chan elevio.ButtonEvent, direction elevio.MotorDirection) string {
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
			case elevio.MD_Up: //fiks denne iffen slik at retningen blir rett test down f 3, så inside 1 og 4 samtidig den skal gå ned
				fmt.Println("Up order check, floor: ", f)
				if elevPt.Orders.Up[f] || elevPt.Orders.Inside[f] || elevPt.Endstation == f { //rar bug i disse if-ene, returnerer selv om det bare er down order i floor f
					return "stop_up_state" // altså stopper på alle etasjer med ordre.
				}
			case elevio.MD_Down:
				fmt.Println("Down order check, floor: ", elevPt.Floor)
				if elevPt.Orders.Down[f] || elevPt.Orders.Inside[f] || elevPt.Endstation == f {
					return "stop_down_state"
				}
			}
		case o := <-orderCH:
			elevhandler.ElevatorAddOrder(elevPt, o)
		default:
		}
		elevhandler.ElevatorSetEndstation(elevPt)
	}
}

func stop(elevPt *elevhandler.ElevatorStatus, drv_stop <-chan bool, drv_obstr <-chan bool, orderCH <-chan elevio.ButtonEvent, direction elevio.MotorDirection) string {
	elevio.SetMotorDirection(elevio.MD_Stop)
	elevio.SetDoorOpenLamp(true)
	elevPt.Direction = direction
	elevhandler.ElevatorClearOrdersAtFloor(elevPt, elevPt.Floor) //fjern andre argument

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
			elevio.SetDoorOpenLamp(false)
			if direction == elevio.MD_Up && elevPt.Endstation > elevPt.Floor {
				return "moving_up_state"
			} else if direction == elevio.MD_Down && elevPt.Endstation < elevPt.Floor {
				return "moving_down_state"
			} else {
				return "idle_state"
			}
		case o := <-orderCH:
			elevhandler.ElevatorAddOrder(elevPt, o)
		default:
		}
		elevhandler.ElevatorSetEndstation(elevPt)
	}
}

func emergency_stop() string {
	return "idle_state" //fiks senere
}

func updateOrderLights(orders <-chan elevhandler.Orders) {
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
