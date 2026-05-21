package dashboard

import (
	"fmt"
	"strconv"
)

func formatInt(value int) string {
	raw := strconv.Itoa(value)
	out := make([]byte, 0, len(raw)+len(raw)/3)
	for index, char := range raw {
		if index > 0 && (len(raw)-index)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, byte(char))
	}
	return string(out)
}

func formatMoney(value float64) string {
	if value >= 1_000_000 {
		return fmt.Sprintf("%.1fM", value/1_000_000)
	}
	if value >= 1_000 {
		return fmt.Sprintf("%.1fK", value/1_000)
	}
	return fmt.Sprintf("%.0f", value)
}

func formatPercent(value float64) string {
	return fmt.Sprintf("%.1f", value)
}

func formatSeconds(seconds int) string {
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	minutes := seconds / 60
	remaining := seconds % 60
	if remaining == 0 {
		return fmt.Sprintf("%dm", minutes)
	}
	return fmt.Sprintf("%dm %ds", minutes, remaining)
}
