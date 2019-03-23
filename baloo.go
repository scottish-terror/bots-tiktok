package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/robfig/cron"
	"github.com/srv1054/bots-baloo/baloomod"

	"github.com/nlopes/slack"
)

func main() {

	var attachments baloomod.Attachment
	var c *cron.Cron
	var cronjobs *baloomod.Cronjobs
	var CronState string

	// Load WallE Config
	wOpts, err := baloomod.LoadWalle()
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	// Set version number
	wOpts.Walle.Version = "4.0"

	// Grab CLI parameters at launch
	wOpts, nocron := baloomod.Startup(wOpts)

	// Load Cron
	if nocron {

		if wOpts.Walle.DEBUG {
			fmt.Println("Not loading CRONs per CLI parameter -nocron")
		}
		if wOpts.Walle.LogToSlack {
			baloomod.LogToSlack("Not Loading CRON based on CLI parameter -nocron", wOpts, attachments)
		}
		CronState = "Not Loaded"
		c = cron.New()

	} else {

		cronjobs, c, err = baloomod.CronLoad(wOpts)
		if err != nil {
			if wOpts.Walle.DEBUG {
				fmt.Println("CRON Jobs failed to load due to file error.")
			}
		} else {
			if wOpts.Walle.DEBUG {
				fmt.Println("CRON Jobs Loaded!")
			}
		}
		CronState = "Running"

	}

	// initate Slack RTM and get started
	api := slack.New(wOpts.Walle.SlackToken)
	rtm := api.NewRTM()
	go rtm.ManageConnection()

	// BOT Listen Loop
	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:

		case *slack.ConnectedEvent:
			wOpts.Walle.BotID = ev.Info.User.ID
			wOpts.Walle.BotName = strings.ToUpper(ev.Info.User.Name)
			wOpts.Walle.TeamID = ev.Info.Team.ID
			wOpts.Walle.TeamName = ev.Info.Team.Name

			//update this "C7XVAJVRS " to a channel ID that matches what he joined (how do i grab that from the connect)
			rtm.SendMessage(rtm.NewOutgoingMessage("Hello! I rebooted, if you care.  :unicorn_face:", "C7XVAJVRS"))

		case *slack.MessageEvent:
			if wOpts.Walle.DEBUG {
				fmt.Printf("Message: %v\n", ev)
			}

			// Check stream if someone says my name or is DM'ing me
			//   Ignore things that I post so i don't loop myself
			if strings.Contains(ev.Msg.Text, "@"+wOpts.Walle.BotID) || string(ev.Msg.Channel[0]) == "D" {
				if ev.Msg.User != wOpts.Walle.BotID {
					// some WallE responses are case sensitive due to Trello being case sensitive, so removing the lower case function
					//   until i think of a better way to handle
					// lowerString := strings.ToLower(ev.Msg.Text)
					lowerString := ev.Msg.Text

					// BOT Responses
					baloomod.Responder(lowerString, wOpts, ev, rtm)

					// BOT Actions
					c, cronjobs, CronState = baloomod.BotActions(lowerString, wOpts, ev, rtm, api, c, cronjobs, CronState)

					// HELP INFO
					if strings.Contains(lowerString, "help") || strings.Contains(lowerString, "what do you do") || strings.Contains(lowerString, "what can you do") || strings.Contains(lowerString, "who are you") {
						rtm.SendMessage(rtm.NewOutgoingMessage("I have DM'd you some help information!", ev.Msg.Channel))
						baloomod.Help(wOpts, ev.Msg.User, api)
					}
				}
			}

		case *slack.LatencyReport:
			if wOpts.Walle.DEBUG {
				fmt.Printf("Current latency: %v\n", ev.Value)
			}

		case *slack.RTMError:
			if wOpts.Walle.DEBUG {
				fmt.Printf("Error: %s\n", ev.Error())
			}

		case *slack.InvalidAuthEvent:
			if wOpts.Walle.DEBUG {
				fmt.Printf("Invalid credentials")
			}
			return

		default:

		}
	}
}
