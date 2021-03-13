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

var elevator elevhandler.ElevatorStatus

func main() {
	numFloors := 4

	elevio.Init("localhost:15657", numFloors)

	//var d elevio.MotorDirection = elevio.MD_Up
	//elevio.SetMotorDirection(d)

	drv_buttons := make(chan elevio.ButtonEvent, 10)
	drv_floors := make(chan int, 10)
	drv_obstr := make(chan bool, 10)
	drv_stop := make(chan bool, 10)

	floorCH := make(chan int, 10)
	directionCH := make(chan elevio.MotorDirection, 10)
	clearCH := make(chan bool, 10)
	elevatorCH := make(chan elevhandler.ElevatorStatus, 10)
	ordersCH := make(chan elevhandler.Orders, 10)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	go elevio.PollFloorSensor(floorCH)
	go elevhandler.ElevatorStatusUpdateForever(drv_buttons, directionCH, floorCH, clearCH, elevatorCH, ordersCH)
	go updateOrderLights(ordersCH)
	//go fix lights elns
	elevator = <-elevatorCH
	state := "idle_state"
	for {
		switch state {
		case "idle_state":
			fmt.Println("in idle")
			state = idle(elevatorCH, drv_stop, directionCH)
		case "moving_up_state":
			fmt.Println("in moving up")
			state = moving(elevatorCH, drv_stop, drv_floors, directionCH, elevio.MD_Up)
		case "moving_down_state":
			fmt.Println("in moving down")
			state = moving(elevatorCH, drv_stop, drv_floors, directionCH, elevio.MD_Down)
		case "stop_up_state":
			fmt.Println("in stop up")
			state = stop(elevatorCH, drv_stop, drv_obstr, directionCH, clearCH, elevio.MD_Up)
		case "stop_down_state":
			fmt.Println("in stop down")
			state = stop(elevatorCH, drv_stop, drv_obstr, directionCH, clearCH, elevio.MD_Down)
		case "emergency_stop_state":
			fmt.Println("in stop")
			state = emergency_stop()
		default:
			state = idle(elevatorCH, drv_stop, directionCH)
		}

	}
}

func idle(elevatorCH <-chan elevhandler.ElevatorStatus, drv_stop <-chan bool, directionCH chan<- elevio.MotorDirection) string {
	fmt.Println("idle before direction")
	directionCH <- elevio.MD_Stop
	fmt.Println("idle after direction")
	elevio.SetMotorDirection(elevio.MD_Stop)
	for {
		fmt.Println("in idle loop")
		select {
		case s := <-drv_stop:
			if s == true {
				return "emergency_stop_state"
			}
		case e := <-elevatorCH:
			switch {
			case e.Endstation < e.Floor:
				return "moving_down_state"
			case e.Endstation > e.Floor:
				return "moving_up_state"
			case e.Endstation == e.Floor: //fyll ut ifs pga emergency stop
				if e.Orders.Inside[e.Floor] || e.Orders.Down[e.Floor] {
					return "stop_down_state"
				} else if e.Orders.Up[e.Floor] {
					return "stop_up_state"
				}
			}
		}
	}
}

func moving(elevatorCH <-chan elevhandler.ElevatorStatus, drv_stop <-chan bool, drv_floors <-chan int, directionCH chan<- elevio.MotorDirection, direction elevio.MotorDirection) string {
	directionCH <- direction
	elevio.SetMotorDirection(direction)
	for {
		select {
		case s := <-drv_stop:
			if s == true {
				return "emergency_stop_state"
			}
		case e := <-elevatorCH:
			elevator = e
		case f := <-drv_floors:
			switch elevator.Direction {
			case elevio.MD_Up: //fiks denne iffen slik at retningen blir rett test down f 3, så inside 1 og 4 samtidig den skal gå ned
				if elevator.Orders.Up[f] || elevator.Orders.Inside[f] || elevator.Endstation == f {
					return "stop_up_state"
				}
			case elevio.MD_Down:
				if elevator.Orders.Down[f] || elevator.Orders.Inside[f] || elevator.Endstation == f {
					return "stop_down_state"
				}
			}
		}
	}
}

func stop(elevatorCH <-chan elevhandler.ElevatorStatus, drv_stop <-chan bool, drv_obstr <-chan bool, directionCH chan<- elevio.MotorDirection, clearCH chan<- bool, direction elevio.MotorDirection) string {
	directionCH <- direction
	elevator = <-elevatorCH
	clearCH <- true
	elevio.SetMotorDirection(elevio.MD_Stop)
	elevio.SetDoorOpenLamp(true)

	timer := time.NewTimer(3 * time.Second)
	for {
		select {
		case s := <-drv_stop:
			if s == true {
				timer.Stop()
				return "emergency_stop_state"
			}
		case e := <-elevatorCH:
			elevator = e
		case o := <-drv_obstr:
			if o {
				timer.Stop()
			} else {
				timer = time.NewTimer(3 * time.Second)
			}
		case <-timer.C:
			elevio.SetDoorOpenLamp(false)
			if direction == elevio.MD_Up && elevator.Endstation > elevator.Floor {
				return "moving_up_state"
			} else if direction == elevio.MD_Down && elevator.Endstation < elevator.Floor {
				return "moving_down_state"
			} else {
				return "idle_state"
			}
		}
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
