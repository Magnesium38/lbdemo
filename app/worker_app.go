package main

import (
	"fmt"
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

// NewApp creates a new app server worker.
func NewApp(config *common.Config) (Worker, error) {
	worker := AppServer{
		config,
		true,
	}

	return &worker, nil
}

type AppServer struct {
	config *common.Config
	doWork bool
}

func (worker *AppServer) Work() error {
	// This isn't real work. Parse to a db maybe?
	for worker.doWork {
		time.Sleep(time.Minute)
	}
	return nil
}

func (worker *AppServer) Shutdown() {

}

func (worker *AppServer) process(msg *irc.Message) string {
	fmt.Println(msg.String())

	if msg.Command == irc.PING {
		r := &irc.Message{
			Command:  irc.PONG,
			Params:   msg.Params,
			Trailing: msg.Trailing,
		}

		return r.String()
	}

	return ""
}

func (worker *AppServer) Do(work string) (string, error) {
	// Parse the message into a workable format.
	message := irc.ParseMessage(work)

	response := worker.process(message)

	return response, nil
}

func (worker *AppServer) Status(requestTime time.Time) balancer.Status {
	// TO DO: revisit statuses.
	return common.NewStatus()
}
