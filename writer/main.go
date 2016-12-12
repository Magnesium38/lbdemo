package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/magnesium38/balancer"
	"github.com/magnesium38/lbdemo/common"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config", "config.json", "The configuration file.")
}

func main() {
	// Parse the arguments into variables.
	flag.Parse()

	fmt.Println("Arguments parsed.")

	// Load the config.
	config, err := common.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Config loaded.")

	// Determine if this is the master or a node.
	arg := strings.ToLower(flag.Arg(0))
	switch arg {
	case "master":
		StartMaster(config)
	case "node":
		StartNode(config)
	default:
		log.Fatal("An argument of either `Master` or `Node` is required.")
	}
}

func StartMaster(config *common.Config) {
	fmt.Println("Starting the load balancer.")

	// Create the load balancer.
	loadBalancer := balancer.NewLoadBalancer(
		config.Address.Writer.Hostname,
		config.Address.Writer.Port,
		config.Master.NodeRegistryPath,
		time.Duration(config.Master.NodeCheckFrequency)*time.Second,
		balancer.NewConnectionFactory(NewStatusFactory()))

	// Queue up all the concurrent bits as jobs.
	jobs := common.NewWorkGroup()
	jobs.Add(loadBalancer.MaintainNodes)
	jobs.Add(loadBalancer.ListenAndServe)

	fmt.Println("Running.")

	// Start the jobs.
	jobs.Start()

	// Wait for an error.
	fatalErr := jobs.Wait()

	fmt.Println("Failing:", fatalErr)
	log.Fatal(fatalErr)
}

func StartNode(config *common.Config) {
	fmt.Println("Starting up a node.")

	// Check the configuration.
	host := config.Node.Hostname
	port := config.Node.Port
	registryPath := config.Master.NodeRegistryPath

	// A port of 0 means that an open one will be assigned on listen.
	//   The listening happens after registration though, so make
	//   sure a port is defined if this is still 0 at startup.
	var err error
	if port == 0 {
		port, err = common.GetOpenPort()
		if err != nil {
			log.Fatal(err)
		}
	}

	// If the registry path is empty, the node cannot register.
	if registryPath == "" {
		log.Fatal(errors.New("The node registry path must not be empty."))
	}

	// Create the node worker.
	worker, err := NewWriter(config)
	if err != nil {
		log.Fatal(err)
	}

	// Create the node itself.
	node := balancer.NewNode(host, port, worker, registryPath)

	// Define the concurrent bits as a work group.
	jobs := common.NewWorkGroup()
	jobs.Add(worker.Work)
	jobs.Add(node.ListenAndServe)

	fmt.Println("Running.")

	// Start working.
	jobs.Start()

	// Register the node so it can be picked up by the load balancer.
	node.Register()

	// Wait for things to die somehow.
	fatalErr := jobs.Wait()
	// Shutdown things as gracefully as possible.
	worker.Shutdown()

	fmt.Println("Failing:", fatalErr)
	log.Fatal(fatalErr)
}
