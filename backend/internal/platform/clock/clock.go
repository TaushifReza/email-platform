package clock

import "time"

// Clock provides time for services.
type Clock interface {
	Now() time.Time
}

// System is the production clock.
type System struct{}

// Now returns the current UTC time.
func (System) Now() time.Time {
	return time.Now().UTC()
}
