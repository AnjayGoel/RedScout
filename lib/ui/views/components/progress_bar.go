package components

import "fmt"

func CreateProgressBar(value, max float64, width int) string {
	if value < 0 {
		value = 0
	}
	if value > max {
		value = max
	}

	filled := int((value / max) * float64(width))
	empty := width - filled

	// Choose color based on value
	var color string
	switch {
	case value >= 90:
		color = "red"
	case value >= 70:
		color = "orange"
	case value >= 50:
		color = "yellow"
	case value >= 30:
		color = "lightgreen"
	default:
		color = "green"
	}

	bar := fmt.Sprintf("[%s][", color)
	for i := 0; i < filled; i++ {
		bar += "|"
	}
	for i := 0; i < empty; i++ {
		bar += " "
	}
	bar += fmt.Sprintf("] %.1f%%[-]", value)
	return bar
}
