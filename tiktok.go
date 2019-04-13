package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/robfig/cron"
	"github.com/scottish-terror/bots-tiktok/tiktokmod"

	"github.com/nlopes/slack"
)

func main() {

	var attachments tiktokmod.Attachment
	var c *cron.Cron
	var cronjobs *tiktokmod.Cronjobs
	var CronState string

	// Load TikTokConf Configuration
	tiktok, err := tiktokmod.LoadTikTokConf()
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	// Set version number
	// Major.Features.Bugs-Tweaks & tomls
	tiktok.Config.Version = "4.0.6-0"

	// Grab CLI parameters at launch
	tiktokOpts, nocron := tiktokmod.Startup(tiktok)

	// Load Cron
	if nocron {

		if tiktok.Config.DEBUG {
			fmt.Println("Not loading CRONs per CLI parameter -nocron")
		}
		if tiktok.Config.LogToSlack {
			tiktokmod.LogToSlack("Not Loading CRON based on CLI parameter -nocron", tiktokOpts, attachments)
		}
		CronState = "Not Loaded"
		c = cron.New()

	} else {

		cronjobs, c, err = tiktokmod.CronLoad(tiktokOpts)
		if err != nil {
			if tiktok.Config.DEBUG {
				fmt.Println("CRON Jobs failed to load due to file error.")
			}
		} else {
			if tiktok.Config.DEBUG {
				fmt.Println("CRON Jobs Loaded!")
			}
		}
		CronState = "Running"

	}

	// initate Slack RTM and get started
	api := slack.New(tiktok.Config.SlackToken)
	rtm := api.NewRTM()
	go rtm.ManageConnection()

	// BOT Listen Loop
	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:

		case *slack.ConnectedEvent:
			tiktok.Config.BotID = ev.Info.User.ID
			tiktok.Config.BotName = strings.ToUpper(ev.Info.User.Name)
			tiktok.Config.TeamID = ev.Info.Team.ID
			tiktok.Config.TeamName = ev.Info.Team.Name

			//update this "C7XVAJVRS " to a channel ID that matches what he joined (how do i grab that from the connect)
			rtm.SendMessage(rtm.NewOutgoingMessage("Hello! I rebooted, if you care.  :unicorn_face:", "C7XVAJVRS"))

		case *slack.MessageEvent:
			if tiktok.Config.DEBUG {
				fmt.Printf("Message: %v\n", ev)
			}

			// Check stream if someone says my name or is DM'ing me
			if strings.Contains(ev.Msg.Text, "@"+tiktok.Config.BotID) || string(ev.Msg.Channel[0]) == "D" {
				// Ignore things that I post so i don't loop myself
				if ev.Msg.User != tiktok.Config.BotID {
					// some bot responses are case sensitive due to Trello being case sensitive, so removing the lower case function
					//   until i think of a better way to handle
					// saidWhat := strings.ToLower(ev.Msg.Text)
					saidWhat := ev.Msg.Text

					// BOT Responses
					tiktokmod.Responder(saidWhat, tiktokOpts, ev, rtm)

					// BOT Actions
					c, cronjobs, CronState = tiktokmod.BotActions(saidWhat, tiktokOpts, ev, rtm, api, c, cronjobs, CronState)

					// HELP INFO
					if strings.Contains(saidWhat, "help") || strings.Contains(saidWhat, "what do you do") || strings.Contains(saidWhat, "what can you do") || strings.Contains(saidWhat, "who are you") {
						rtm.SendMessage(rtm.NewOutgoingMessage("I have DM'd you some help information!", ev.Msg.Channel))
						tiktokmod.Help(tiktokOpts, ev.Msg.User, api)
					}
				}
			}

		case *slack.LatencyReport:
			if tiktok.Config.DEBUG {
				fmt.Printf("Current latency: %v\n", ev.Value)
			}

		case *slack.RTMError:
			if tiktok.Config.DEBUG {
				fmt.Printf("Error: %s\n", ev.Error())
			}
			if tiktok.Config.LogToSlack {
				tiktokmod.LogToSlack("`ERROR`: RTMError, See Console DEBUG", tiktok, attachments)
			}

		case *slack.InvalidAuthEvent:
			if tiktok.Config.DEBUG {
				fmt.Printf("Invalid credentials")
			}
			if tiktok.Config.LogToSlack {
				tiktokmod.LogToSlack("`ERROR`: Invalid Slack API Credentials", tiktok, attachments)
			}
			return

		default:

		}
	}
	if tiktok.Config.DEBUG {
		fmt.Println("Dumped out of RTM Loop!")
	}
}
