package meshtastic

import (
	"fmt"
	"math"
	"time"
)

func timeAgo(timestamp time.Time) string {
	seconds := int(time.Since(timestamp).Seconds())

	if seconds == 1 {
		return "one second"
	}
	if seconds < 60 {
		return fmt.Sprintf("%d seconds", seconds)
	}

	minutes := int(math.Floor(float64(seconds) / 60))
	if minutes == 1 {
		return "one minute"
	}
	if minutes < 60 {
		return fmt.Sprintf("%d minutes", minutes)
	}

	hours := int(math.Floor(float64(minutes) / 60))
	if hours == 1 {
		return "one hour"
	}
	if hours < 24 {
		return fmt.Sprintf("%d hours", hours)
	}

	days := int(math.Floor(float64(hours) / 24))
	if days == 1 {
		return "one day"
	}
	return fmt.Sprintf("%d days", days)
}
