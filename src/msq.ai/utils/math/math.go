package math

import "math"

const EPSILON = 0.00001

func IsZero(val float64) bool {

	if math.Abs(val) < EPSILON {
		return true
	}

	return false
}
