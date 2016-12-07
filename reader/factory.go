package main

import (
	"strconv"
	"strings"
	"time"

	"github.com/magnesium38/balancer"
)

// NewConnectionFactory returns an implementation of NodeFactory
func NewConnectionFactory(factory balancer.StatusFactory) balancer.NodeFactory {
	return &ConnectionFactory{factory}
}

func NewStatusFactory() balancer.StatusFactory {
	return &Factory{}
}

// The ConnectionFactory specifically required to do the Reader load balancing.
type ConnectionFactory struct {
	status balancer.StatusFactory
}

// Create takes the connection info and creates the connection struct.
func (factory *ConnectionFactory) Create(connInfo string) (balancer.NodeConnection, error) {
	parts := strings.Split(connInfo, ":")

	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, err
	}

	conn := Connection{}
	conn.host = parts[0]
	conn.port = port
	conn.jobCount = 0
	conn.status = factory.status.Create()
	conn.client = nil

	return &conn, nil
}

type Factory struct {
}

func (factory *Factory) Create() balancer.Status {
	status := Status{}
	return &status
}

type Status struct {
	idleTime time.Duration
	str      string
}

func (status *Status) GetIdleTime() time.Duration {
	return time.Minute
}

func (status *Status) String() string {
	return time.Now().String()
}

func (status *Status) Update(z string) {

}
