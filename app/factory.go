package main

import (
	"time"

	"github.com/magnesium38/balancer"
)

func NewStatusFactory() balancer.StatusFactory {
	return &Factory{}
}

type Factory struct {
}

func (factory *Factory) Create() balancer.Status {
	status := Status{}
	return &status
}

// TO DO: THIS ENTIRE FILE

type Status struct {
}

func (status *Status) GetIdleTime() time.Duration {
	return time.Minute
}

func (status *Status) String() string {
	return time.Now().String()
}

func (status *Status) Update(z string) {

}
