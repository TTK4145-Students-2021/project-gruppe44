
/**
 * @file
 * @brief Floor library
 */

/**
 * @brief A structure to represent the floor of the elevator.
 */

typedef struct{
    int current;            /** < The current floor of the elevator */
    bool above;             /** < The elevators position relative to the current floor, 1 for above, 0 for on the floor or below */
}Floor;


/**
 * @brief Updates the current floor and the relative position to that floor on floor @p floor. Sets floorlights, if it is on a floor.
 * 
 * @param floor The floor struct that is going to be updated.
 * @param direction The direction of the elevator.
 * 
 * @return 1 if the floor was changed, 0 if not.
 */

bool floor_update(Floor * floor, bool direction);
