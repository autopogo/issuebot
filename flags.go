package main

import (
	"errors"
	"flag"
	"github.com/mailgun/log"
	"io/ioutil"
	"strings"
)

// these vars make checking if flags were set easier than using the flag pkg default
// they can't be const because golang wants to optimize them away but we use their addresses
var (
	DEFAULT_SLACK_TOKEN_FILE  = "./slack_token"
	DEFAULT_GITHUB_TOKEN_FILE = "./github_token"
)
var (
	ErrBadFlag = errors.New("command was run improperly, check --help")
)
var (
	// flag_org is the name of the org the bot has access to
	flag_org = flag.String("org",
		"",
		"Organization bot has access to")

	// flag_auth provides a file by which to load authorized users
	flag_auth = flag.String("auth",
		"./userlist",
		"What file contains a list of authorized users")

	// flag_slack_token will provide a slack token manually
	flag_slack_token = flag.String("slack_token",
		"",
		"Specify the slack token")

	// github_token will provide a github token manually
	flag_github_token = flag.String("github_token",
		"",
		"Specify the github oauth token")

	// flag_slack_token_file will provide a filename for a slack token
	flag_slack_token_file = flag.String("slack_token_file",
		"",
		"Specify the slack token file")

	// github_token_file will provide a filename for a github token
	flag_github_token_file = flag.String("github_token_file",
		"",
		"Specify the github oauth token file")
)

func flagInit() (err error) {
	// Read the flags in
	flag.Parse()

	// Now do a basic sanity test on flags.
	err = verifyFlagsSanity()
	return err
}

// verifyFlagsSanity just does a basic check on provided flags- are the ones that need to be there, there?
// This function reads from globals (flags)
func verifyFlagsSanity() (err error) {
	if len(*flag_org) == 0 {
		log.Errorf("You must specify an organization, see --help")
		err = ErrBadFlag
	}
	if *flag_slack_token_file == "" {
		if *flag_slack_token == "" {
			flag_slack_token_file = &DEFAULT_SLACK_TOKEN_FILE
		}
	} else {
		if *flag_slack_token != "" {
			log.Errorf("You must not specify both --flag_slack_token_file AND --flag_slack_token, see --help")
			err = ErrBadFlag
		}
	}
	if *flag_github_token_file == "" {
		if *flag_github_token == "" {
			flag_github_token_file = &DEFAULT_GITHUB_TOKEN_FILE
		}
	} else {
		if *flag_github_token != "" {
			log.Errorf("You must not specify both --flag_github_token_file AND --flag_github_token, see --help")
			err = ErrBadFlag
		}
	}
	return err
}

// loadAuthedUsers reads the file specified by flag_auth to create a list of authorized slack users. The caller can decide whether or not to exit on error.
// This function reads from globals (flags)
func loadAuthedUsers() (ret []string, err error) {
	var authFile []byte
	authFile, err = ioutil.ReadFile(*flag_auth)
	if err != nil {
		return nil, err
	}
	ret = strings.Split(string(authFile), "\r\n")
	if len(ret) == 1 { // it's possible that different OSes have different newline conventions- don't check \n first
		ret = strings.Split(string(authFile), "\n")
	}
	return ret, nil
}

// loadSlackTokentries to return a slack key (from flag or file)
// This function reads from globals (flags)
func loadSlackToken() (slack string, err error) {
	if *flag_slack_token == "" {
		slackFileContents, err := ioutil.ReadFile(*flag_slack_token_file)
		if err != nil {
			return "", err
		}
		slack = string(slackFileContents)

		// BUG(AJ) I just don't like this
		if strings.HasSuffix(slack, "\r\n") {
			log.Warningf("Slack Token ends in whitespace, eliminating two characters (\\r\\n)...")
			slack = strings.TrimSuffix(slack, "\r\n")
		} else if strings.HasSuffix(slack, "\n") {
			log.Warningf("Slack Token ends in whitespace, eliminating one character (\\n)...")
			slack = strings.TrimSuffix(slack, "\n")
		}
	} else {
		slack = *flag_slack_token
	}
	return slack, nil
}

// loadGitHubToken tries to return a github key (from flag or file)
// This function reads from globals (flags)
func loadGitHubToken() (github string, err error) {
	if *flag_github_token == "" {
		githubFileContents, err := ioutil.ReadFile(*flag_github_token_file)
		if err != nil {
			return "", err
		}
		github = string(githubFileContents)
		if strings.HasSuffix(github, "\r\n") {
			log.Warningf("GitHub Token ends in whitespace, eliminating two characters (\\r\\n)...")
			github = strings.TrimSuffix(github, "\r\n")
		} else if strings.HasSuffix(github, "\n") {
			log.Warningf("GitHub Token ends in whitespace, eliminating one character (\\n)...")
			github = strings.TrimSuffix(github, "\n")
		}
	} else {
		github = *flag_slack_token
	}
	return github, nil
}
