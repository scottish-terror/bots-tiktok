package baloomod

import (
	"flag"
	"fmt"
	"os"
)

// Startup - Startup stuff
func Startup(balooOpts *BalooConf) (*BalooConf, bool) {

	var attachments Attachment

	tkey := flag.String("tkey", "", "Trello Key")
	ttoken := flag.String("ttoken", "", "Trello Token")
	slackhook := flag.String("slackhook", "", "Slack Webhook")
	slacktoken := flag.String("slacktoken", "", "Slack Bot Token")
	slackoauth := flag.String("slackoauth", "", "Slack OAuth User Token")
	dbuser := flag.String("dbuser", "", "CSQL DB User Acct")
	dbpassword := flag.String("dbpassword", "", "CSQL DB User Password")
	nocron := flag.Bool("nocron", false, "Start "+baloo.Config.BotName+" without loading cron jobs")
	ghtoken := flag.String("git", "", "Github Token")
	version := flag.Bool("v", false, "Show current version number")

	flag.Parse()

	balooOpts.Config.Tkey = *tkey
	balooOpts.Config.Ttoken = *ttoken
	balooOpts.Config.SlackHook = *slackhook
	balooOpts.Config.SlackToken = *slacktoken
	balooOpts.Config.SlackOAuth = *slackoauth
	balooOpts.Config.DBUser = *dbuser
	balooOpts.Config.DBPassword = *dbpassword
	balooOpts.Config.GitToken = *ghtoken
	nocrontab := *nocron

	if *version {
		fmt.Println("I'm Baloo Version " + balooOpts.Config.Version)
		os.Exit(0)
	}

	if balooOpts.Config.Tkey == "" || balooOpts.Config.Ttoken == "" || balooOpts.Config.SlackHook == "" || balooOpts.Config.SlackToken == "" || balooOpts.Config.DBUser == "" || balooOpts.Config.DBPassword == "" || balooOpts.Config.GitToken == "" {
		fmt.Println("\nWarning CLI parameters: -tkey, -ttoken, slacktoken, -slackhook, -git, -dbuser and -dbpassword are required!")
		os.Exit(0)
	}

	// Start up message
	if balooOpts.Config.LogToSlack {
		LogToSlack("*Hi I'm starting up after being stopped!* - Version `"+balooOpts.Config.Version+"`", balooOpts, attachments)
	}

	// Dump start message to STDOUT for logging purposes - regardless if DEBUG is on
	fmt.Println("---- Hi I'm starting up after being stopped! - Version " + balooOpts.Config.Version + " -----")

	return balooOpts, nocrontab
}
