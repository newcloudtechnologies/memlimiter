package utils

// ClampFloat64 ограничивает переданное значение заданным диапазоном.
// Калька с https://docs.unity3d.com/ScriptReference/Mathf.Clamp.html.
func ClampFloat64(value, min, max float64) float64 {
	if value < min {
		return min
	}

	if value > max {
		return max
	}

	return value
}
