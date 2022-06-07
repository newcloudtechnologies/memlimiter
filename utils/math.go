package utils

// ClampFloat64 limits the provided value according to the given range.
// Origin: https://docs.unity3d.com/ScriptReference/Mathf.Clamp.html.
func ClampFloat64(value, min, max float64) float64 {
	if value < min {
		return min
	}

	if value > max {
		return max
	}

	return value
}
