package goodreads

import (
	"encoding/xml"
	"fmt"
	"time"
)

func xmlUnmarshal(b []byte, i interface{}) error {
	return xml.Unmarshal(b, i)
}

func parseDate(s string) (time.Time, error) {
	date, err := time.Parse(time.RFC3339, s)
	if err != nil {
		date, err = time.Parse(time.RubyDate, s)
		if err != nil {
			return time.Time{}, err
		}
	}

	return date, nil
}

func relativeDate(d string) string {
	date, err := parseDate(d)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	s := time.Now().Sub(date)

	days := int(s / (24 * time.Hour))
	if days > 1 {
		return fmt.Sprintf("%v days ago", days)
	} else if days == 1 {
		return fmt.Sprintf("%v day ago", days)
	}

	hours := int(s / time.Hour)
	if hours > 1 {
		return fmt.Sprintf("%v hours ago", hours)
	}

	minutes := int(s / time.Minute)
	if minutes > 2 {
		return fmt.Sprintf("%v minutes ago", minutes)
	} else {
		return "Just now"
	}
}
