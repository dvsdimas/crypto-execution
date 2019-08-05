package math

import (
	"fmt"
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
	return fmt.Sprintf("%.10f", val)
}

func Int64ToString(val int64) string {
	return strconv.FormatInt(val, 10)
}
