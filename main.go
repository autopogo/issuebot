// Package main provides a service-like program, issuebot, that listens to a slack channel and allows users to create new issues on a github repo
package main

import (
	"github.com/mailgun/log"
	"os"
	"os/signal"
	"sync"
	"time"
)

// running is a flag that when set to false, tells all goroutines to exit nicely- and new callbacks not to start (important) network op
var running = true

func init() {
	// Load a logger- load more later if you want them- some might depend on flags.
	console, _ := log.NewLogger(log.Config{"console", "debug"}) // note: debug, info, warning, error
	log.Init(console)
}

// main is being used here kind of like a forward declaration- it's the outline of the program.
// I don't use init because I can't control execution order and so test can't use env variables
func main() {

	// This is so that we wait for callbacks to finish if we're exiting cleanly
	var waitForCb sync.WaitGroup
	// Note: use waitForCb.Add(1) to count an ongoing op and waitForCb.Done() to finish

	if err := flagInit(); err != nil {
		log.Errorf("Program couldn't start: %v", err)
		os.Exit(1)
	}

	// TODO: Init github and test
	// TODO: Init slack with function and callback

	// run will wait for a signal
	run()

	waitForCb.Wait()
}

// run sets up both apis with the proper definitions, and then it waits for signals
func run() {

	// This is all for catching signals
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	timeNow := time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)
	log.Infof("Issuebot booted for org %v", *flag_org)

	// Now we're going to wait on signals from terminal
	for signalRecvd := range signalChannel {
		newTime := time.Now()
		if newTime.Sub(timeNow) < 1000*time.Millisecond {
			log.Infof("Exiting...")
			running = false
			break
		}
		timeNow = newTime
		log.Infof("Received a signal: %v", signalRecvd)
		log.Infof("Reloading auth'ed users and keys")
		// TODO: reload authed users and keys- don't panic on error
		log.Infof("Send again <1 second to exit cleanly")
	}

}

// TODO: Refactor this todo list
// TODO: Issue log would be cool on custom response to log channel
// TODO: It would be cool if you could use a particular user's github credentials from slack but that's a part of a custom auth feature
// TODO: Could use oauth but oauth reg on github was requiring a larger-scoped registration process (website, etc)- the super benefit of this is that it restrict users to repos they have access to, which would eliminate some of the problems with github's over-scoped token situation (below). This would also allow us to use the suggest log channel as a way to communicate with issues.
// TODO: the above would change the keyword too
// TODO: unfortunately, github scopes are _not_ granular. for issues, you get +rw on code, pull reqs, wikis, settings, webhooks, deploy keys. this is a 2yr mega thread on github.com/dear-githu[M#Èb
// TODO: one way to turn this into an interface would be to create a new bot type that could be initialized with a slack org(s) and github org(s) but I feel like it would just be better to run seperate processes -- although multiple slack orgs and github orgs would be good (although multiple github orgs if users supply their own keys too)
