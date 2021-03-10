/**
 * @file
 * @brief The state machine, and functions for each state.
 */



/**
 * @brief The state machine, switches between states.
 */

void elevator();

/**
 * @brief The idle state, updates the orders struct @p p_orders until a new state is set. When there is a new order, it will change the state according to the first new order. 
 * Order in floor above or below will return the state @c moving_up_state or @c moving_down_state respectively.
 * Order in the same floor will return the state @c stop_up_state or @c stop_down_state, depending on the button direction (up/down, panel_button sets it in @p stop_up_state).
 * If the stop button is pressed, it will return the state @c stop_state.
 * 
 * @param p_orders A pointer to the orders struct in use.
 * @param p_floor A pointer to the floor struct, to determine where the elevator is.
 * 
 * @return The new state.
 */

int idle(Orders * p_orders, Floor * p_floor);

/**
 * @brief The moving up/down state. Sets the motor to move up or down depending on the direction @p direction (up=1, down=0).
 * Continously updates orders and checks if it is on a floor with an order (ignoring orders of oposite direction @p direction) or the endstation is reached.
 * If any of the above conditions are met, the state @c stop_up_state or @c stop_down_state depending on the direction @p direction.
 * If the stop button is pressed, the state @c stop_state is returned.
 * 
 * @param p_orders A pointer to the orders struct in use.
 * @param p_floor A pointer to the Floor struct.
 * @param direction The direction the elevator is moving.
 * 
 *  @return The new state. 
 */

int moving(Orders * p_orders, Floor * p_floor, bool direction);


/**
 * @brief The stop up/down state. Stops the motor.
 * Continously updates orders, while opening the door for three seconds before closing it again. 
 * After the door is closed, it will return the state @c idle_state if the elevator is on floor the endstation,
 * or return either state @c moving_up_state or @c moving_down_state if the floor @p p_orders points to is not the endstation.
 * If the stop button is pressed, the state is set to @c stop.
 * 
 * @param p_orders A pointer to the orders struct in use.
 * @param p_floor A pointer to the Floor struct.
 * @param direction The direction the elevator is moving.
 * 
 * @return The new state.
 */

int stop(Orders * p_orders, Floor * p_floor, bool direction);
/**
 * @brief The stop state. Stops the motor, clears all orders and turns on the stop light for as long the stop button is pressed.
 * It then returns the state @c idle_state .
 * If the elevator is on a floor when the stop button is pulled, the door is also open for as long as the stop button is pressed,
 * and the state @c stop_down_state is returned instead (up or down doesn't really matter here, but we choose down).
 *  
 *
 * @param p_orders A pointer to the orders struct in use.
 * @param p_floor A pointer to the Floor struct.
 * @param direction The direction the elevator is moving.
 *  
 * @return The new state.
 */

int emergency_stop(Orders * p_orders, Floor * p_floor);