package main

import (
	"bufio"
	"errors"
	"net"
	"net/rpc"
	"strings"
	"time"

	"gopkg.in/sorcix/irc.v1"

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

// NewReader creates a new reader worker.
func NewReader(config *common.Config) (Worker, error) {
	appServer, err := rpc.DialHTTP("tcp", config.Address.App.String())
	if err != nil {
		return nil, err
	}

	worker := Reader{
		config,
		&common.AtomicStringSlice{},
		make(chan string),
		make(chan string),
		true,
		appServer,
	}
	return &worker, nil
}

// A Reader is how the node listens to irc.
type Reader struct {
	config    *common.Config
	channels  *common.AtomicStringSlice
	toJoin    chan string
	toPart    chan string
	doWork    bool
	appServer *rpc.Client
}

// Join accepts the name of a channel and attempts to join it.
func (worker *Reader) join(channel string) (string, error) {
	// TO DO: This assumes that joining a channel cannot fail, which is false.
	//   If a channel doesn't exist, joining would fail. This should be fixed,
	//   but it is a low priority fix for now. The biggest issue this has
	//   currently is that a typo on the channel would give no feedback.
	worker.channels.Add(channel)
	worker.toJoin <- channel
	return "", nil
}

// Part accepts the name of a channel and attempts to leave it.
func (worker *Reader) part(channel string) (string, error) {
	if !worker.channels.Has(channel) {
		// This doesn't have the channel, let the load balancer know.
		return "", &balancer.InvalidWorkError{
			Str: "Not currently listening on the channel: " + channel,
		}
	}

	// TO DO: This shouldn't fail like join, but it is blindly assuming that
	//   it successfully left the channel. Another low priority fix.
	worker.channels.Remove(channel)
	worker.toPart <- channel
	return "", nil
}

// Halt starts the shutdown of the worker.
func (worker *Reader) halt() (string, error) {
	worker.doWork = false
	return "", nil
}

// Work is the main function to actually read the irc connection.
func (worker *Reader) Work() error {
	// Perform initial connection to IRC.
	ircConn, err := net.Dial("tcp", worker.config.Irc.ConnInfo)
	if err != nil {
		return err
	}
	defer ircConn.Close()

	sleepDuration := time.Duration(worker.config.Irc.ReadFrequency) * time.Millisecond

	reader := bufio.NewReader(ircConn)
	toWrite := worker.startWriter(ircConn)
	worker.startChannelManager(toWrite)

	toWrite <- "PASS " + worker.config.Irc.Password
	toWrite <- "NICK " + worker.config.Irc.Nickname

	// Begin maintaining IRC connection.
	for worker.doWork {
		line, err := reader.ReadString('\n')
		if err != nil {
			// Most likely just not enough to read. Sleep it off.
			time.Sleep(sleepDuration)
			continue
		}

		go worker.process(line, toWrite)
	}

	return errors.New("The worker was instructed to stop.")
}

func (worker *Reader) startChannelManager(toWrite chan<- string) {
	go func() {
		for {
			select {
			case channel := <-worker.toJoin:
				msg := irc.Message{
					Command: irc.JOIN,
					Params:  []string{channel},
				}
				toWrite <- msg.String()
			case channel := <-worker.toPart:
				msg := irc.Message{
					Command: irc.PART,
					Params:  []string{channel},
				}
				toWrite <- msg.String()
			}
		}
	}()
}

func (worker *Reader) startWriter(conn net.Conn) chan<- string {
	// Create the channel
	write := make(chan string)

	// Create a concurrent function to write to the connection from the channel.
	go func(conn net.Conn, toWrite chan string) {
		writer := bufio.NewWriter(conn)
		for {
			line, okay := <-toWrite
			if !okay {
				return
			}

			// Don't send empty lines.
			if line == "" {
				continue
			}

			writer.WriteString(line + "\r\n")
			writer.Flush()
		}
	}(conn, write)

	return write
}

func (worker *Reader) process(line string, toWrite chan<- string) {
	// Pass the line onto the app server's load balancer.
	var reply string
	// TO DO: Check what the RPC name to call.
	err := worker.appServer.Call("Master.Work", line, &reply)
	if err != nil {
		// If there's an error, it's something the app server returned.
		//   Should be safe to just log and ignore.
		// TO DO: Log this.
	}
	toWrite <- reply
}

// Shutdown starts as graceful of a shutdown of the worker as possible.
func (worker *Reader) Shutdown() {
	// Stop reading IRC.
	worker.halt()
	worker.appServer.Close()
}

// Do instructs the worker to complete some form of load balanced work.
func (worker *Reader) Do(work string) (string, error) {
	// A reader's `work` is leaving or joining a channel.
	//   If there was a good way for the load balancer to spin up new
	//   nodes as needed, this would also be the way to stop them. There
	//   shouldn't be harm in creating a halt `work`.

	//   TO DO: verify if other work exists.

	// Parse whether the work is a leave or join request.
	parts := strings.Split(work, " ")

	if len(parts) == 0 {
		return "", &balancer.InvalidWorkError{
			Str: "Work given was an unexpected empty string.",
		}
	}

	// If the command was halt, stop as gracefully as possible.
	if parts[0] == "HALT" {
		return worker.halt()
	}

	// There should always be either JOIN or PART as a command followed
	//   by the channel name. Anything else is an error.
	if len(parts) != 2 {
		return "", &balancer.InvalidWorkError{
			Str: "Work given was made up of more than two parts: " + work,
		}
	}

	// Use the command to perform the desired action.
	switch parts[0] {
	case "JOIN":
		return worker.join(parts[1])
	case "PART":
		return worker.part(parts[1])
	default:
		return "", &balancer.InvalidWorkError{
			Str: "Work given did not have an accepted command: " + work,
		}
	}
}

func (worker *Reader) Status(requestTime time.Time) balancer.Status {
	// TO DO: revisit statuses.
	return common.NewStatus()
}
