package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"time"

	"github.com/magnesium38/balancer"
	"github.com/magnesium38/lbdemo/common"
)

// Worker is a common interface to put multiple workers into the same process.
type Worker interface {
	// Extend the worker as defined in balancer.Worker
	balancer.Worker

	// Work does the additional side tasks that the worker requires.
	Work() error

	// Shutdown attempts to stop the worker as gracefully as possible.
	Shutdown()
}

// NewWriter creates a new writer worker.
func NewWriter(config *common.Config) (Worker, error) {
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

// A Writer is how the node writes to irc.
type Writer struct {
	config    *common.Config
	doWork    bool
	ircConn   *net.Conn
	appServer *rpc.Client
	toWrite   chan writePayload
}

// Halt starts the shutdown of the worker.
func (worker *Writer) halt() {
	worker.toWrite <- writePayload{"QUIT Shutting Down\r\n", make(chan error)}
	worker.doWork = false
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

	// Start the reader and get the writer.
	go worker.startReader(ircConn)
	writer := bufio.NewWriter(ircConn)

	// Write the login information first.
	writer.WriteString("PASS " + worker.config.Irc.Password + "\r\n")
	writer.WriteString("NICK " + worker.config.Irc.Nickname + "\r\n")
	writer.Flush()

	for worker.doWork {
		payload := <-worker.toWrite

		if payload.msg == "" {
			payload.doneChan <- nil
			continue
		}

		_, err := writer.WriteString(payload.msg)

		// TO DO: Should I also flush here?
		writer.Flush()

		select {
		case payload.doneChan <- err:
		default:
		}

	}

	return errors.New("The worker was instructed to stop.")
}

func (worker *Writer) startReader(conn net.Conn) {
	// Create the reader and the sleep duration.
	reader := bufio.NewReader(conn)
	sleepDuration := time.Duration(worker.config.Irc.ReadFrequency) * time.Millisecond

	worker.toWrite <- writePayload{"PASS " + worker.config.Irc.Password, make(chan error)}
	worker.toWrite <- writePayload{"NICK " + worker.config.Irc.Nickname, make(chan error)}

	fmt.Println("reading")
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			// Most likely just not enough to read. Sleep it off.
			time.Sleep(sleepDuration)
			continue
		}

		fmt.Println(string(line))

		// Send lines off to be processed.
		go worker.process(string(line))
	}
}

func (worker *Writer) process(line string) {
	// Pass the line onto the app server's load balancer.
	var reply string
	// TO DO: Check what the RPC name to call.
	err := worker.appServer.Call("Master.Work", line, &reply)
	if err != nil {
		// If there's an error, it's something the app server returned.
		//   Should be safe to just log and ignore.
		// TO DO: Log this.
	}
	// Sending a dummy channel, the error for now doesn't matter.
	//   TO DO: Do something with this channel and make the error matter.
	worker.toWrite <- writePayload{reply, make(chan error)}
}

// Do instructs the worker to copmlete some form of load balanced work.
func (worker *Writer) Do(work string) (string, error) {
	// A writer's work is to take commands as given by the app servers and
	//   write them to the IRC connection.

	done := make(chan error)
	worker.toWrite <- writePayload{work, done}

	err := <-done

	// This an error is here, it should be logged. TO DO: Actually log.
	//   Probably also check that this I'm not missing something here.
	//   This seems awfully short.

	return "", err
}

// Shutdown starts as graceful of a shutdown of the worker as possible.
func (worker *Writer) Shutdown() {
	worker.appServer.Close()
	worker.halt()
}

func (worker *Writer) Status(requestTime time.Time) balancer.Status {
	// TO DO: revisit statuses.
	return common.NewStatus()
}

type writePayload struct {
	msg      string
	doneChan chan error
}
