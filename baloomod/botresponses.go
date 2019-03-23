package baloomod

import (
	"strings"

	"github.com/nlopes/slack"
)

// Responder - check for chatty messages that need responses not actions
func Responder(lowerString string, baloo *BalooConf, ev *slack.MessageEvent, rtm *slack.RTM) {
	// -- ALL BUSINESS
	if strings.Contains(lowerString, "your 411") || strings.Contains(lowerString, "version") {
		rtm.SendMessage(rtm.NewOutgoingMessage("Hi! My name is "+baloo.Config.BotName+" and I'm version "+baloo.Config.Version+". My slack ID is "+baloo.Config.BotID+" and I'm part of the "+baloo.Config.TeamName+" team (ID: "+baloo.Config.TeamID+").  This channels ID is "+ev.Msg.Channel+". Your Slack UID is "+ev.Msg.User+". I currently write my logs to "+baloo.Config.LogChannel, ev.Msg.Channel))
	}

	// -- FUN STUFF
	if strings.Contains(lowerString, "hello") || strings.Contains(lowerString, "hey there") || strings.Contains(lowerString, " hi") {
		rtm.SendMessage(rtm.NewOutgoingMessage("Hi there!", ev.Msg.Channel))
	}

	if strings.Contains(lowerString, "who is the prettiest") || strings.Contains(lowerString, "who is the fairest") {
		rtm.SendMessage(rtm.NewOutgoingMessage("Robert Blue, that's who.  :blue_heart:", ev.Msg.Channel))
	}

	if strings.Contains(lowerString, "you dumb") || strings.Contains(lowerString, "you suck") || strings.Contains(lowerString, "you stupid") {
		rtm.SendMessage(rtm.NewOutgoingMessage("All I have to say is:  EBKAC", ev.Msg.Channel))
	}

	if strings.Contains(lowerString, "salute") {
		rtm.SendMessage(rtm.NewOutgoingMessage(":salute:", ev.Msg.Channel))
	}

	if strings.Contains(lowerString, "shit list") {
		rtm.SendMessage(rtm.NewOutgoingMessage("All y'all are on the :poop: list", ev.Msg.Channel))
	}

	if strings.Contains(lowerString, "beat you") || strings.Contains(lowerString, "kill you") || strings.Contains(lowerString, "destroy you") || strings.Contains(lowerString, "punch you") {
		rtm.SendMessage(rtm.NewOutgoingMessage(":challenge_accepted:", ev.Msg.Channel))
	}

	if strings.Contains(lowerString, "you rule") || strings.Contains(lowerString, " rock") {
		rtm.SendMessage(rtm.NewOutgoingMessage("Yes..Yes I do.  Thanks!", ev.Msg.Channel))
	}

	if strings.Contains(lowerString, "bro do you even") || strings.Contains(lowerString, "do you even") {
		rtm.SendMessage(rtm.NewOutgoingMessage("Like a Boss!!", ev.Msg.Channel))
	}

	if strings.Contains(lowerString, "are you back") || strings.Contains(lowerString, "are you here") || strings.Contains(lowerString, "are you there") {
		rtm.SendMessage(rtm.NewOutgoingMessage(":pony_trotting:", ev.Msg.Channel))
	}

	if strings.Contains(lowerString, "nice work") || strings.Contains(lowerString, "good job") || strings.Contains(lowerString, "good work") || strings.Contains(lowerString, "nice job") {
		rtm.SendMessage(rtm.NewOutgoingMessage("Thank you so much! :cheers:", ev.Msg.Channel))
	}

	if strings.Contains(lowerString, "thank you") || strings.Contains(lowerString, "thanks ") {
		rtm.SendMessage(rtm.NewOutgoingMessage("You are most welcome!", ev.Msg.Channel))
	}

	if strings.Contains(lowerString, "mix me a martini") || strings.Contains(lowerString, "make me a martini ") {
		rtm.SendMessage(rtm.NewOutgoingMessage("Gin, not vodka, obviously, stirred for ten seconds while glancing at an unopened bottle of vermouth.  Coming right up. :cocktail:", ev.Msg.Channel))
	}

	if strings.Contains(lowerString, "more ponies") {
		rtm.SendMessage(rtm.NewOutgoingMessage("Coming up!\n:pony_trotting: :pony_trotting: :pony_trotting: :pony_trotting: :pony_trotting: :pony_trotting: :pony_trotting: :pony_trotting: :pony_trotting: :pony_trotting: :pony_trotting: :pony_trotting: :pony_trotting: :pony_trotting: :pony_trotting: ", ev.Msg.Channel))
	}

	if strings.Contains(lowerString, "eat it") {
		rtm.SendMessage(rtm.NewOutgoingMessage(":cookie-monster:", ev.Msg.Channel))
	}
}
