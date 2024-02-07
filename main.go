package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
)
var routingPoint string
func main() {

	// Parse the command-line flags
	var processPort int
	var balancerPort int
	var numberOfProcesses int
	
	flag.IntVar(&processPort, "process-port", 11000, "port for the processes")
	flag.IntVar(&balancerPort, "balancer-port", 9090, "port for the load balancer")
	flag.IntVar(&numberOfProcesses, "number-of-processes", 5, "number of processes to start")
	flag.StringVar(&routingPoint, "routing-point", "index.php", "routing point")
	flag.Parse()

	//./phpmyserver -process-port=11000 -balancer-port=9090 -number-of-processes=5 -routing-point=index.php

	fmt.Printf("process-port: %d\n", processPort)
	fmt.Printf("balancer-port: %d\n", balancerPort)
	fmt.Printf("number-of-processes: %d\n", numberOfProcesses)
	fmt.Printf("routing-point: %s\n", routingPoint)

	// Create a slice of processes
	processes := make([]*Process, numberOfProcesses)
	for i := 0; i < numberOfProcesses; i++ {
		processes[i] = &Process{Port: processPort + i}
	}

	// Use a wait group to wait for all the processes to start
	var wg sync.WaitGroup
	wg.Add(len(processes))

	// Start the processes
	for _, p := range processes {
		go p.Start(&wg)
	}

	// Create a channel to receive signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Create a round-robin load balancer
	rr := 0
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Forward the request to the next process in the slice
		targetURL := fmt.Sprintf("http://localhost:%d", processes[rr].Port)
		rr = (rr + 2) % len(processes)
		target, err := url.Parse(targetURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Printf("Request URL: %s%s\n", targetURL, r.URL.String())

		// Create a proxy
		proxy := httputil.NewSingleHostReverseProxy(target)

		// Update the headers to allow for SSL redirection
		r.URL.Host = target.Host
		r.URL.Scheme = target.Scheme
		r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
		r.Host = target.Host

		// Note that ServeHttp is non blocking and uses a go routine under the hood
		proxy.ServeHTTP(w, r)
	})

	// Start the load balancer
	go func() {
		fmt.Printf("Load balancer listening on port %d\n", balancerPort)
		http.ListenAndServe(fmt.Sprintf(":%v", balancerPort), nil)
	}()

	// Wait for a signal
	<-sigCh

	// Kill the subprocesses
	for _, p := range processes {
		p.Kill()
	}

	// Wait for the subprocesses to finish
	wg.Wait()
}

// Process represents a process that listens on a specific port.
type Process struct {
	Port int
	cmd  *exec.Cmd
}

// Start starts the PHP development server on the port for this process.
func (p *Process) Start(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		p.cmd = exec.Command("php", "-S", fmt.Sprintf("localhost:%d", p.Port), routingPoint)
		err := p.cmd.Start()
		if err == nil {
			fmt.Printf("Process started on port %d\n", p.Port)
			break
		}

		fmt.Printf("Error starting process on port %d: %s\n", p.Port, err)
		p.Port += 10
	}

	// Wait for the process to finish
	err := p.cmd.Wait()
	if err != nil {
		fmt.Printf("Error waiting for process on port %d: %s\n", p.Port, err)
	}
}

// Kill kills the process.
func (p *Process) Kill() {
	err := p.cmd.Process.Kill()
	if err != nil {
		fmt.Printf("Error killing process on port %d: %s\n", p.Port, err)
	}
}
