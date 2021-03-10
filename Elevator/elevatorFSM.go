import (
	"fmt"
	"runtime"
	"time"
	"elevatorHandler"
	"./elevio"
)

	/*
	#define up_direction 1
	#define down_direction 0
	*/

var elevator ElevatorStatus

func ElevatorFSM() {

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors  := make(chan int)
   	drv_obstr   := make(chan bool)
   	drv_stop    := make(chan bool)
		
	floorCH := make(chan int)
	directionCH := make(chan elevio.MotorDirection)
	stopCH := make(chan bool)
	elevatorCH := make(chan ElevatorStatus)

	go elevio.PollButtons(drv_buttons)
   	go elevio.PollFloorSensor(drv_floors)
   	go elevio.PollObstructionSwitch(drv_obstr)
   	go elevio.PollStopButton(drv_stop)

	go elevio.PollFloorSensor(floorCH)
	go ElevatorStatusUpdateForever(drv_buttons, directionCH, floorCH, stopCH, elevatorCH)
	//go fix lights elns


	state := "idle_state"
	for {
		switch state {
		case "idle_state":
			state = idle()
		case "moving_up_state":
			state = moving()
		case "moving_down_state":
			state = moving()
		case "stop_up_state":
			state = stop()
		case "stop_down_state":
			state = stop()
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
func idle() string {
	directionCH <- elevio.MD_Stop
	elevio.setMotorDirection(elevio.MD_Stop)
	for {
		select{
		case s := <- drv_stop:
			return "emergency_stop_state"
		case e := <-elevatorCH:
			switch e.endstation{
			case e.endstation < e.floor:
				return "moving_down_state"
			case e.endstation > e.floor:
				return "moving_up_state"
			case e.floor:			//fyll ut ifs pga emergency stop
				if e.orders.inside[e.floor]||e.orders.down[e.floor]{
					return "stop_down_state"
				}else if e.orders.up[e.floor]{
					return "stop_up_state"
				}
			}
		}

	}
}

func moving(direction elevio.MotorDirection) string {
	elevio.SetMotorDirection(direction)
	for{
		select{
		case o := <- ordersCH
		}
	}
}

func stop() string {

}

func emergency_stop() string {

}
