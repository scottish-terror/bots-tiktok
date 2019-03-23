package baloomod

import (
	"flag"
	"fmt"
	"os"
)

// Startup - WallE Startup stuff
func Startup(walleOpts *WallConf) (*WallConf, bool) {

	var attachments Attachment

	tkey := flag.String("tkey", "", "Trello Key")
	ttoken := flag.String("ttoken", "", "Trello Token")
	slackhook := flag.String("slackhook", "", "Slack Webhook")
	slacktoken := flag.String("slacktoken", "", "Slack Bot Token")
	slackoauth := flag.String("slackoauth", "", "Slack OAuth User Token")
	dbuser := flag.String("dbuser", "", "CSQL DB User Acct")
	dbpassword := flag.String("dbpassword", "", "CSQL DB User Password")
	nocron := flag.Bool("nocron", false, "Start WallE without loading cron jobs")
	ghtoken := flag.String("git", "", "Github Token")
	version := flag.Bool("v", false, "Show current version number")

	flag.Parse()

	walleOpts.Walle.Tkey = *tkey
	walleOpts.Walle.Ttoken = *ttoken
	walleOpts.Walle.SlackHook = *slackhook
	walleOpts.Walle.SlackToken = *slacktoken
	walleOpts.Walle.SlackOAuth = *slackoauth
	walleOpts.Walle.DBUser = *dbuser
	walleOpts.Walle.DBPassword = *dbpassword
	walleOpts.Walle.GitToken = *ghtoken
	nocrontab := *nocron

	if *version {
		fmt.Println("I'm Wall-E Version " + walleOpts.Walle.Version)
		os.Exit(0)
	}

	if walleOpts.Walle.Tkey == "" || walleOpts.Walle.Ttoken == "" || walleOpts.Walle.SlackHook == "" || walleOpts.Walle.SlackToken == "" || walleOpts.Walle.DBUser == "" || walleOpts.Walle.DBPassword == "" || walleOpts.Walle.GitToken == "" {
		fmt.Println("\nWarning CLI parameters: -tkey, -ttoken, slacktoken, -slackhook, -git, -dbuser and -dbpassword are required!")
		os.Exit(0)
	}

	// Start up message
	if walleOpts.Walle.LogToSlack {
		LogToSlack("*Hi I'm starting up after being stopped!* - Version `"+walleOpts.Walle.Version+"`", walleOpts, attachments)
	}

	// Dump start message to STDOUT for logging purposes - regardless if DEBUG is on
	fmt.Println("---- Hi I'm starting up after being stopped! - Version " + walleOpts.Walle.Version + " -----")

	return walleOpts, nocrontab
}
