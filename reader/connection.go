package main

import (
	"fmt"
	"net/rpc"
	"strconv"
	"time"

	"github.com/magnesium38/balancer"
)

type Connection struct {
	host     string
	port     int
	jobCount int
	status   balancer.Status
	client   *rpc.Client
}

// GetHost returns the hostname that the node is listening on.
func (conn *Connection) GetHost() string {
	return conn.host
}

// GetPort returns the port that the node is listening on.
func (conn *Connection) GetPort() int {
	return conn.port
}

// AddJob increments the internal counter of jobs by 1.
func (conn *Connection) AddJob() {
	conn.jobCount++
}

// FinishJob decrements the internal counter of jobs by 1.
func (conn *Connection) FinishJob() {
	conn.jobCount--
}

// GetWorkLoad returns the current estimated work load that
//   the node has. It is made possible through Add/FinishJob
func (conn *Connection) GetWorkLoad() int {
	return conn.jobCount
}

// Connect initiates the connection between the balancer
//   and the node.
func (conn *Connection) Connect() error {
	// Construct the address and dial.
	addr := conn.host + ":" + strconv.Itoa(conn.port)
	client, err := rpc.DialHTTP("tcp", addr)

	// Check if there is an error before storing the connection.
	if err != nil {
		return err
	}

	conn.client = client

	return nil
}

// GetStatus returns the current Status of the node, or an error
//   if it cannot retrieve the status.
func (conn *Connection) GetStatus() balancer.Status {
	return conn.status
}

// UpdateStatus requests the status from the node and stores it.
func (conn *Connection) UpdateStatus() error {
	// Request the status from the node.
	requestTime := time.Now()
	var response string
	err := conn.client.Call("Status", requestTime, &response)

	if err != nil {
		return err
	}

	conn.status.Update(response)

	return nil
}

// Send is how a balancer can send work to the nodes. This
//   implementation is using RPC.
func (conn *Connection) Send(work string) (string, error) {
	var response string

	err := conn.client.Call("Server.Do", work, &response)

	if err != nil {
		fmt.Println("Send error: ", err)
		return "", err
	}

	return response, nil
}
