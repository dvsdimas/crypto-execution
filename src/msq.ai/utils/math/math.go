package math

import (
	"math"
	"strconv"
)

const EPSILON = 0.00000001

func IsZero(val float64) bool {

	if math.Abs(val) < EPSILON {
		return true
	}

	return false
}

func Float64ToString(val float64) string {
	return strconv.FormatFloat(val, 'f', -10, 64)
}

func Int64ToString(val int64) string {
	return strconv.FormatInt(val, 10)
}
