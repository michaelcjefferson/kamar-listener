package components

import "time"

func ConvertToLocalTZ(t time.Time) time.Time {
	tz, err := time.LoadLocation("Pacific/Auckland")
	if err != nil {
		return t
	}
	return t.In(tz)
}
