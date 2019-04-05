package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/robfig/cron"
	"github.com/srv1054/bots-baloo/tiktokmod"

	"github.com/nlopes/slack"
)

func main() {

	var attachments tiktokmod.Attachment
	var c *cron.Cron
	var cronjobs *tiktokmod.Cronjobs
	var CronState string

	// Load BalooConf Config
	baloo, err := tiktokmod.LoadBalooConf()
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	// Set version number
	// Major.Features.Bugs-Tweaks & tomls
	baloo.Config.Version = "4.0.0-3"

	// Grab CLI parameters at launch
	wOpts, nocron := tiktokmod.Startup(baloo)

	// Load Cron
	if nocron {

		if baloo.Config.DEBUG {
			fmt.Println("Not loading CRONs per CLI parameter -nocron")
		}
		if baloo.Config.LogToSlack {
			tiktokmod.LogToSlack("Not Loading CRON based on CLI parameter -nocron", wOpts, attachments)
		}
		CronState = "Not Loaded"
		c = cron.New()

	} else {

		cronjobs, c, err = tiktokmod.CronLoad(wOpts)
		if err != nil {
			if baloo.Config.DEBUG {
				fmt.Println("CRON Jobs failed to load due to file error.")
			}
		} else {
			if baloo.Config.DEBUG {
				fmt.Println("CRON Jobs Loaded!")
			}
		}
		CronState = "Running"

	}

	// initate Slack RTM and get started
	api := slack.New(baloo.Config.SlackToken)
	rtm := api.NewRTM()
	go rtm.ManageConnection()

	// BOT Listen Loop
	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:

		case *slack.ConnectedEvent:
			baloo.Config.BotID = ev.Info.User.ID
			baloo.Config.BotName = strings.ToUpper(ev.Info.User.Name)
			baloo.Config.TeamID = ev.Info.Team.ID
			baloo.Config.TeamName = ev.Info.Team.Name

			//update this "C7XVAJVRS " to a channel ID that matches what he joined (how do i grab that from the connect)
			rtm.SendMessage(rtm.NewOutgoingMessage("Hello! I rebooted, if you care.  :unicorn_face:", "C7XVAJVRS"))

		case *slack.MessageEvent:
			if baloo.Config.DEBUG {
				fmt.Printf("Message: %v\n", ev)
			}

			// Check stream if someone says my name or is DM'ing me
			//   Ignore things that I post so i don't loop myself
			if strings.Contains(ev.Msg.Text, "@"+baloo.Config.BotID) || string(ev.Msg.Channel[0]) == "D" {
				if ev.Msg.User != baloo.Config.BotID {
					// some bot responses are case sensitive due to Trello being case sensitive, so removing the lower case function
					//   until i think of a better way to handle
					// lowerString := strings.ToLower(ev.Msg.Text)
					lowerString := ev.Msg.Text

					// BOT Responses
					tiktokmod.Responder(lowerString, wOpts, ev, rtm)

					// BOT Actions
					c, cronjobs, CronState = tiktokmod.BotActions(lowerString, wOpts, ev, rtm, api, c, cronjobs, CronState)

					// HELP INFO
					if strings.Contains(lowerString, "help") || strings.Contains(lowerString, "what do you do") || strings.Contains(lowerString, "what can you do") || strings.Contains(lowerString, "who are you") {
						rtm.SendMessage(rtm.NewOutgoingMessage("I have DM'd you some help information!", ev.Msg.Channel))
						tiktokmod.Help(wOpts, ev.Msg.User, api)
					}
				}
			}

		case *slack.LatencyReport:
			if baloo.Config.DEBUG {
				fmt.Printf("Current latency: %v\n", ev.Value)
			}

		case *slack.RTMError:
			if baloo.Config.DEBUG {
				fmt.Printf("Error: %s\n", ev.Error())
			}

		case *slack.InvalidAuthEvent:
			if baloo.Config.DEBUG {
				fmt.Printf("Invalid credentials")
			}
			return

		default:

		}
	}
}
