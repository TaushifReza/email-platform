package dialer

import "errors"

var ErrNotAllowed = errors.New("host not allowed")

// AllowedPorts are the only ports a provider may use (PRV-04). Fixed per the
// SRS; not configurable.
//
//nolint:gochecknoglobals
var AllowedPorts = map[int]struct{}{
	25:   {},
	465:  {},
	587:  {},
	2525: {},
}

func portAllowed(port int) bool {
	_, ok := AllowedPorts[port]
	return ok
}
