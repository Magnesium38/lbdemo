package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"net/rpc"
	"time"

	"github.com/magnesium38/balancer"
	"github.com/magnesium38/lbdemo/common"
)

// NewWriter creates a new writer worker.
func NewWriter(config *common.Config) (*Writer, error) {
	appServer, err := rpc.DialHTTP("tcp", config.Address.App.String())
	if err != nil {
		return nil, err
	}

	worker := Writer{
		config,
		true,
		nil,
		appServer,
		make(chan writePayload),
	}

	return &worker, nil
}

type writePayload struct {
	msg      string
	doneChan chan error
}

// A Writer is how the node writes to irc.
type Writer struct {
	config    *common.Config
	doWork    bool
	ircConn   *net.Conn
	appServer *rpc.Client
	toWrite   chan writePayload
}

// Work is the main function to write to the irc connection.
func (worker *Writer) Work() error {
	// The work is sending whatever payloads are received. But first
	//   the initial connection must be processed.
	ircConn, err := net.Dial("tcp", worker.config.Irc.ConnInfo)
	if err != nil {
		return err
	}
	defer ircConn.Close()

	// These two lines needs to be writen first. Setup a goroutine to send
	//   them first.
	go func() {
		doneChan := make(chan error)
		worker.toWrite <- writePayload{"PASS " + worker.config.Irc.Password, doneChan}
		<-doneChan
		worker.toWrite <- writePayload{"NICK " + worker.config.Irc.Nickname, doneChan}
		<-doneChan
	}()

	// Start the reader and get the writer.
	go worker.startReader(ircConn)
	writer := bufio.NewWriter(ircConn)

	fmt.Println("Starting `work`.")

	for worker.doWork {
		payload := <-worker.toWrite

		// If the payload is empty, no need to attempt to write it. No error.
		if payload.msg == "" {
			payload.doneChan <- nil
			continue
		}

		// Write the payload plus the new line. Making an assumption that all
		//   bytes will always be written and so the first argument can be
		//   ignored. The error is being deferred to whatever gave the payload.
		//   TO DO: Should I log the error here as well to make it easier
		//   to track down if it is serious?
		_, err := writer.WriteString(payload.msg + "\r\n")
		writer.Flush()

		payload.doneChan <- err
	}

	return errors.New("The worker was instructed to stop.")
}

func (worker *Writer) startReader(conn net.Conn) {
	// Create the reader and the sleep duration.
	reader := bufio.NewReader(conn)
	sleepDuration := time.Duration(worker.config.Irc.ReadFrequency) * time.Millisecond

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// The connection has died. Attempt to recover or die.
				// TO DO:
			}

			// Most likely just not enough to read. Sleep it off.
			// ^^ Not true. reader.ReadString is blocking unless EOF.
			// This should be handled somehow. TO DO:
			time.Sleep(sleepDuration)
			continue
		}

		// Send lines off to be processed.
		go worker.process(string(line))
	}
}

func (worker *Writer) process(line string) {
	// Pass the line onto the app server's load balancer.
	var reply string
	err := worker.appServer.Call("Master.Work", line, &reply)
	if err != nil {
		// If there's an error, it's something the app server returned.
		//   Should be safe to just log and ignore.
		// TO DO: Log this.
	}

	// Create the payload.
	doneChan := make(chan error)
	worker.toWrite <- writePayload{reply, doneChan}

	// Process the error??? Honestly, I don't think I'll care most of the time.
	err = <-doneChan
	if err != nil {
		// TO DO: Log the error and check the logs to see what happened.
	}
}

// Do instructs the worker to copmlete some form of load balanced work.
func (worker *Writer) Do(work string) (string, error) {
	// A writer's work is to take commands as given by the app servers and
	//   write them to the IRC connection.

	// Create the payload.
	done := make(chan error)
	worker.toWrite <- writePayload{work, done}

	// Retrieve the potential error from writing.
	err := <-done

	// If an error is here, it should be logged. TO DO: Actually log.
	//   Probably also check that this I'm not missing something here.
	//   This seems awfully short.

	// Nothing important to return except if an error occured.
	return "", err
}

// Shutdown starts as graceful of a shutdown of the worker as possible.
func (worker *Writer) Shutdown() {
	// Close the RPC connection to the App Server.
	worker.appServer.Close()

	// Send a quit message to terminate the connection.
	worker.toWrite <- writePayload{"QUIT Shutting Down", make(chan error)}

	// Breaking the work loop is fine. This'll cause it to return an error
	//   which in turn will cause the process to exit.
	worker.doWork = false
}

func (worker *Writer) Status(requestTime time.Time) balancer.Status {
	// TO DO: revisit statuses.
	return common.NewStatus()
}
