package common

import "time"

// Status s
type Status struct {
}

// NewStatus builds and returns a new status.
func NewStatus() *Status {
	return &Status{}
}

// GetIdleTime s
func (status *Status) GetIdleTime() time.Duration {
	return time.Minute
}

// String s
func (status *Status) String() string {
	return time.Now().String()
}

// Update s
func (status *Status) Update(z string) {

}
