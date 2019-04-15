package tiktokmod

import (
	"flag"
	"fmt"
	"os"
)

// Startup - Startup stuff
func Startup(tiktokOpts *TikTokConf) (*TikTokConf, bool) {

	var attachments Attachment

	tkey := flag.String("tkey", "", "Trello Key")
	ttoken := flag.String("ttoken", "", "Trello Token")
	slackhook := flag.String("slackhook", "", "Slack Webhook")
	slacktoken := flag.String("slacktoken", "", "Slack Bot Token")
	slackoauth := flag.String("slackoauth", "", "Slack OAuth User Token")
	ghtoken := flag.String("git", "", "Github Token")
	dbuser := flag.String("dbuser", "", "CSQL DB User Acct")
	dbpassword := flag.String("dbpassword", "", "CSQL DB User Password")
	nocron := flag.Bool("nocron", false, "Start "+tiktok.Config.BotName+" without loading cron jobs")
	version := flag.Bool("v", false, "Show current version number")
	osenv := flag.Bool("osenv", false, "All tokens are being passed by OS ENV instead of CLI")

	flag.Parse()

	if *osenv {
		fmt.Println("Checking OS ENV for parameters")
		tiktokOpts.Config.Tkey = os.Getenv("tkey")
		tiktokOpts.Config.Ttoken = os.Getenv("ttoken")
		tiktokOpts.Config.SlackHook = os.Getenv("slackhook")
		tiktokOpts.Config.SlackToken = os.Getenv("slacktoken")
		tiktokOpts.Config.SlackOAuth = os.Getenv("slackoauth")
		tiktokOpts.Config.DBUser = os.Getenv("dbuser")
		tiktokOpts.Config.DBPassword = os.Getenv("dbpassword")
		tiktokOpts.Config.GitToken = os.Getenv("git")

		if *version {
			fmt.Println("I'm TikTok Version " + tiktokOpts.Config.Version)
			os.Exit(0)
		}

		if tiktokOpts.Config.Tkey == "" || tiktokOpts.Config.Ttoken == "" || tiktokOpts.Config.SlackHook == "" || tiktokOpts.Config.SlackToken == "" || tiktokOpts.Config.DBUser == "" || tiktokOpts.Config.DBPassword == "" || tiktokOpts.Config.GitToken == "" {
			fmt.Println("\nWarning OS Environment parameters: tkey, ttoken, slacktoken, slackhook, git, dbuser and dbpassword are required!")
			os.Exit(0)
		}
	} else {
		fmt.Println("Checking CLI for parameters")
		tiktokOpts.Config.Tkey = *tkey
		tiktokOpts.Config.Ttoken = *ttoken
		tiktokOpts.Config.SlackHook = *slackhook
		tiktokOpts.Config.SlackToken = *slacktoken
		tiktokOpts.Config.SlackOAuth = *slackoauth
		tiktokOpts.Config.DBUser = *dbuser
		tiktokOpts.Config.DBPassword = *dbpassword
		tiktokOpts.Config.GitToken = *ghtoken

		if *version {
			fmt.Println("I'm TikTok Version " + tiktokOpts.Config.Version)
			os.Exit(0)
		}

		if tiktokOpts.Config.Tkey == "" || tiktokOpts.Config.Ttoken == "" || tiktokOpts.Config.SlackHook == "" || tiktokOpts.Config.SlackToken == "" || tiktokOpts.Config.DBUser == "" || tiktokOpts.Config.DBPassword == "" || tiktokOpts.Config.GitToken == "" {
			fmt.Println("\nWarning CLI parameters: -tkey, -ttoken, -slacktoken, -slackhook, -git, -dbuser and -dbpassword are required!")
			os.Exit(0)
		}
	}

	nocrontab := *nocron

	if tiktokOpts.Config.LogToSlack {
		LogToSlack("*Hi I'm starting up after being stopped!* - Version `"+tiktokOpts.Config.Version+"`", tiktokOpts, attachments)
	}

	// Dump start message to STDOUT for logging purposes - regardless if DEBUG is on
	fmt.Println("---- Hi I'm starting up after being stopped! - Version " + tiktokOpts.Config.Version + " -----")

	return tiktokOpts, nocrontab
}
