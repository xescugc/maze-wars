package utils

// ClosestMultiple finds the closest multiple of 'b' for the number 'a'
func ClosestMultiple(a, b int) int {
	a = a + b/2
	a = a - (a % b)
	return a
}

// PreviousMultiple find the previous multiple of 'b' for the number 'a'
func PreviousMultiple(a, b int) int {
	na := ClosestMultiple(a, b)
	// If the ClosestMultiple is bigger then we just
	// subtract the 'b' to go to the previous one
	if a < na {
		return na - b
	}
	return na
}
