package main

import (
	"context"
	"errors"
	"github.com/mailgun/log"
	"github.com/shomali11/slacker"
	"strings"
	"sync"
)

// processParam processes command parameters manualy because the built-in command processor is extremely weak
func processParam(allParam string) (repo string, title string, body string, ok bool) {
	// look for three sets of quotes
	// loop through allParam
	var threeParams [3]string
	var paramCount int = 0
	var quoteSwitch bool = false
	var escape bool = false
	var start int = 0
	log.Debugf("allParam: %v", allParam)
	for i, c := range allParam {
		log.Debugf("i, c: %d, %c", i, c)
		if (i == 0) && (c != '"') { // first character has to be a quote
			log.Debugf("Bad start to allParam")
			return "", "", "", false
		} else if i == 0 { // first character was a quote
			quoteSwitch = true
			start = i + 1
			log.Debugf("Open the quotes!")
		} else if (!escape) && (c == '\\') { // we'll escape the next character if we weren't escaped
			escape = true
			log.Debugf("Next character literal")
		} else if escape { // turn off escape and move on (we have escaped the current character)
			escape = false
			log.Debugf("Character was literal")
		} else if c == '"' { // we've encountered a non-escaped quote
			if quoteSwitch { // we were in quotes, now we're out
				threeParams[paramCount] = allParam[start:i]
				paramCount += 1
				log.Debugf("Turn off quotes! %v", threeParams[paramCount-1])
			} else { // we are just starting quotes
				start = i + 1
				log.Debugf("Turn on quotes!")
			}
			quoteSwitch = !quoteSwitch
		}
	}
	if paramCount != 3 {
		return "", "", "", false
	}
	return threeParams[0], threeParams[1], threeParams[2], true
}

// OpenBot just starts the bot with the callback. BUG(AJ) Warning- this bot library doesn't like concurrency. This library is written like we're in node.js.
func openBot(token string, authedUsers []string, waitForCb sync.WaitGroup, gBot *GitHubIssueBot) (err error) {
	var descriptionString strings.Builder
	descriptionString.WriteString("Creates a new issue on github for ")
	descriptionString.WriteString(gBot.GetOrg())
	descriptionString.WriteString("/YOUR_REPO")
	_ = authedUsers
	sBot := slacker.NewClient(token)

	// newCommand is built by a callback factory to attach to a certain waitgroup and GitHubIssueBot
	newCommand := func(waitForCb sync.WaitGroup, gBot *GitHubIssueBot) func(slacker.Request, slacker.ResponseWriter) {
		return func(request slacker.Request, response slacker.ResponseWriter) {
			if !running {
				response.ReportError(errors.New("Issuebot is starting up or shutting down, try again in a few seconds."))
				return
			}
			waitForCb.Add(1)
			defer waitForCb.Done()
			// Note: This supports multiple commands but not "", and I didn't want to override/reimplement the interfaces due to time-cost
			allParam := request.StringParam("all", "")
			repo, title, body, ok := processParam(allParam)
			if !ok {
				response.ReportError(errors.New("You must specify repo, title, and body for new issue! All in quotes."))
				return
			}
			var URL string
			URL, err = gBot.NewIssue(repo, title, body)

			if err != nil {
				response.ReportError(errors.New("There was an error with the GitHub interface... Check 1) the repo name 2) the logs"))
				return
			}
			response.Reply(URL)
			return
		}
	}(waitForCb, gBot)

	newIssue := &slacker.CommandDefinition{
		Description: descriptionString.String(),
		Example:     "new \"repo\" \"issue title\" \"issue body\"",
		Handler:     newCommand,
	}
	sBot.Command("new <all>", newIssue)

	// TODO: is this how you turn it off?
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log.Infof("Starting slack bot listen...")
	err = sBot.Listen(ctx) // TODO: This blocks, so how are we going to turn it off?
	log.Infof("bot.Listen(ctx) returned")

	return err
}
