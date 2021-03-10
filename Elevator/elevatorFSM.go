package elevfsm

import (
	"fmt"
	"runtime"
	"time"
	"./elevhandler"
	"./elevio"
)

/*
	#define up_direction 1
	#define down_direction 0
*/

var elevator ElevatorStatus

func elevatorFSM() {
	numFloors := 4

    elevio.Init("localhost:15657", numFloors)
    
    var d elevio.MotorDirection = elevio.MD_Up
    //elevio.SetMotorDirection(d)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	floorCH := make(chan int)
	directionCH := make(chan elevio.MotorDirection)
	clearCH := make(chan bool)
	elevatorCH := make(chan ElevatorStatus)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	go elevio.PollFloorSensor(floorCH)
	go elevhandler.ElevatorStatusUpdateForever(drv_buttons, directionCH, floorCH, clearCH, elevatorCH)
	//go fix lights elns

	state := "idle_state"
	for {
		switch state {
		case "idle_state":
			state = idle(elevatorCH, drv_stop, directionCH)
		case "moving_up_state":
			state = moving(elevatorCH, drv_stop, drv_floors, directionCH, elevio.MD_Up)
		case "moving_down_state":
			state = moving(elevatorCH, drv_stop, drv_floors, directionCH, elevio.MD_Down)
		case "stop_up_state":
			state = stop(elevatorCH, drv_stop, drv_obstr, directionCH, clearCH, elevio.MD_Up)
		case "stop_down_state":
			state = stop(elevatorCH, drv_stop, drv_obstr, directionCH, clearCH, elevio.MD_Down)
		case "emergency_stop_state":
			state = emergency_stop()
		default:
			state = idle()
		}

	}
}

/*
func orders_set_endstation(floor_from int, floor_to int, p_orders *Orders) bool{ //sets endstation to orders, returns true if there is an order
	if floor_from < floor_to{
		for f := floor_from; f <= floor_to; f++{
			if p_orders.inside[f] || p_orders.down[f] || p_orders.up[f]{
				p_orders.endstation = f
				return true
			}
		}
		return false
	}
	else {
		for f := floor_from; f >= floor_to; f--{
			if p_orders.inside[f] || p_orders.down[f] || p_orders.up[f]{
				p_orders.endstation = f
				return true
			}
		}
		return false
	}
}
*/
func idle(elevatorCH <-chan ElevatorStatus, drv_stop <-chan bool, directionCH chan<- elevio.MotorDirection) string {
	directionCH <- elevio.MD_Stop
	elevio.setMotorDirection(elevio.MD_Stop)
	for {
		select {
		case s := <-drv_stop:
			if s == true {
				return "emergency_stop_state"
			}
		case e := <-elevatorCH:
			switch e.endstation {
			case e.endstation < e.floor:
				return "moving_down_state"
			case e.endstation > e.floor:
				return "moving_up_state"
			case e.floor: //fyll ut ifs pga emergency stop
				if e.orders.inside[e.floor] || e.orders.down[e.floor] {
					return "stop_down_state"
				} else if e.orders.up[e.floor] {
					return "stop_up_state"
				}
			}
		}
	}
}

func moving(elevatorCH <-chan ElevatorStatus, drv_stop <-chan bool, drv_floors <-chan int, directionCH chan<- elevio.MotorDirection, direction elevio.MotorDirection) string {
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
			switch elevator.direction {
			case elevio.MD_Up:
				if elevator.order.up[f] || elevator.order.inside[f] || elevator.endstation == f {
					return "stop_up_state"
				}
			case elevio.MD_Down:
				if elevator.order.down[f] || elevator.order.inside[f] || elevator.endstation == f {
					return "stop_down_state"
				}
			}

		}
	}
}

func stop(elevatorCH <-chan ElevatorStatus, drv_stop <-chan bool, drv_obstr <-chan bool, directionCH chan<- elevio.MotorDirection, clearCH chan<- bool, direction elevio.MotorDirection) string {
	directionCH <- direction
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
			if elevator.endstation == elevator.floor{
				return "idle_state"
			}else if directionCH == elevio.MD_Up {
				return "moving_up_state"
			}else {
				return "moving_down_state" 
			}
			}
		}
	}

}

func emergency_stop() string {
	return "idle_state" //fiks senere

}
