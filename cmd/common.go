package cmd

import (
	"fmt"
	"time"
)

const (
	colorGood     = 0x00FF00
	colorWarning  = 0xFFAC1C
	colorError    = 0xFF0000
	colorCritical = 0x964B00

	iconGood    = "🟢" // green circle
	iconWarning = "🟡" // yellow circle
	iconError   = "🔴" // red circle
)

func formattedTime(t time.Time) string {
	return fmt.Sprintf("<t:%d:R>", t.Unix())
}

func getColorForAlertLevel(alertLevel AlertLevel) int {
	switch alertLevel {
	case alertLevelNone:
		return colorGood
	case alertLevelWarning:
		return colorWarning
	case alertLevelCritical:
		return colorCritical
	default:
		return colorError
	}
}
