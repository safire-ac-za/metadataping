package main

/*
 * Simple program that run external update code on event.
 * Must only run one instance a time but accept a contiuum of events.
 *
 * Accept "pings" and queue them into one event.
 * When the event queue not are empty then run the runcode function and empty
 * the event queue at the beginning of the run.
 * Under the run collect new pings.
 */

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
)

var run chan string
var runcommand string
var netinterface string = "localhost:9000"
var githash = "No git hash Provided"
var buildstamp = "No build time Provided"

// Wait for event on the 'run' channel and then execute external program/script.
func runcode() {
	var url string
	var matched bool
	for {
		url = <-run
		// Do not create command struct before env_runcommand are validated as a single file
		// Need to create Command struct per run else error message. No idea what I do wronge.
		log.Printf("update event run start: %s %s", runcommand, url)
		cmd := exec.Command(runcommand)
		matched, _ = regexp.MatchString("\\bnojanus\\b", url)
		if matched {
			cmd.Args = append(cmd.Args, "--nojanus")
		}
		matched, _ = regexp.MatchString("\\bsilent\\b", url)
		if matched {
			cmd.Args = append(cmd.Args, "--silent")
		}
		matched, _ = regexp.MatchString("\\bforcerefresh\\b", url)
		if matched {
			cmd.Args = append(cmd.Args, "--forcerefresh")
		}
		err := cmd.Run()
		if err == nil {
			log.Print("update event run end")
		} else {
			log.Printf("update event run FAILED: %s", err.Error())
		}
	}
}

// Accept http requests and trigger execution of external program/script.
func ping(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Pong")
	select {
	case run <- r.URL.Path:
		{
		} // Create one event. Push true in to the 'run' channel.
	default:
		{
		} // If the 'run' channel are full then do nothing. The 'run' channel are a non-blocking channel.
	}
}

// Initialize and validate global variables. Runned only one time at program start.
func initialsetup() {
	// env_runcommand must validated as a single file
	env_runcommand := os.Getenv("METADATA_RUN")
	_, err := os.Stat(env_runcommand)
	if err != nil {
		log.Fatal(err)
	}
	runcommand = env_runcommand
	log.Print(runcommand)

	env_port := os.Getenv("METADATA_INTERFACE")
	if env_port != "" {
		netinterface = env_port
	}
	log.Print(netinterface)
	run = make(chan string, 1)
}

// Setup variables
// Start runcode as non-blocking
// Accept http requests and redirect them to the ping function
func main() {
	// go install -ldflags "-X main.buildstamp=`date -u '+%Y-%m-%d_%H:%M:%S'` -X main.githash=`git rev-parse --short HEAD`" src/metadataping.go
	// static build: CGO_ENABLED=0 go install -ldflags "-X main.buildstamp=`date -u '+%Y-%m-%d_%H:%M:%S'` -X main.githash=`git rev-parse --short HEAD`" src/metadataping.go
	fmt.Printf("Git Commit Hash: %s\n", githash)
	fmt.Printf("UTC Build Time: %s\n", buildstamp)
	initialsetup()
	go runcode()

	http.HandleFunc("/", ping)
	http.ListenAndServe(netinterface, nil)
}
