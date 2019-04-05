package tiktokmod

import "github.com/nlopes/slack"

//Help - Help return message
func Help(tiktok *TikTokConf, user string, api *slack.Client) {

	var attachments Attachment
	var hmessage string
	var message string
	var emessage string
	var testPayload BotDMPayload

	userInfo, _ := api.GetUserInfo(user)

	message = message + "*Hi, I heard you need some help!*\n"
	message = message + "I can do many things based on specific keywords and permissions.\n\n"
	message = message + "For *more detailed easier to read* help go here <https://github.com/srv1054/bots-tiktok/wiki/TikTokConf-Help|to my Wiki Page>\n\n"
	message = message + "Here's some more common commands I know though:\n\n"

	hmessage = hmessage + "* what's your 411 (or version)\n"
	hmessage = hmessage + "* start a new sprint [<board>] - I'll setup a new sprint for your board - `perms required`\n"
	hmessage = hmessage + "* shutdown please - I will shut all services down and log out of slack - `perms required`\n"
	hmessage = hmessage + "* stop/shutdown/halt all cron - I will disable all running cronjobs until told otherwise - `perms required`\n"
	hmessage = hmessage + "* reload/re-load cron - I will re-read cron.toml and reload all the cron jobs in it\n"
	hmessage = hmessage + "* build a configuration file [<trello board ID>] - I will run through any trello board and find the Unique ID's you need to build a .toml file for your board!\n"
	hmessage = hmessage + "* list all cronjobs - I will list all programmed cron jobs that I know about\n"
	hmessage = hmessage + "* list available boards - I will list all of the Trello Team/Boards I have TOML configurations for and their access name\n"
	hmessage = hmessage + "* dupe trello board [<board>] - I will make a copy of this board and name it DUPE-M-D-Y<board name> and assign it to the `Board Copies` Collection\n"
	hmessage = hmessage + "* retro board [<board>] - will return the URL to the current sprint retro board for that team\n"
	hmessage = hmessage + "* <well|wrong|vent> retro card [<board>] <my card info> - will create a new card on the current sprint retro board for that team in the Well or Wrong list\n"
	hmessage = hmessage + "* add me [email,trello id,github id] - register yourself with Tik-Tok so he knows your ID's. No quotes needed around items with spaces or special characters\n"
	hmessage = hmessage + "* description history `cardID` - Well return the historical card description data for a given card ID.  Look in a card URL to get its ID #\n"
	hmessage = hmessage + "* company holidays - I will return a list of company Holidays that I know about.\n"
	hmessage = hmessage + "* previous sprint points [<board>] `SprintName` - will return points by squad for previous sprint named `SprintName`\n"
	hmessage = hmessage + "* list github <users|repos> - Will DM user all github `users` or `repos` depending on which you asked for.\n"

	emessage = emessage + "If you are DM'ing me you do not need to say @" + tiktok.Config.BotName + " first\n\n"
	emessage = emessage + "@" + tiktok.Config.BotName + " whats your 411\n"
	emessage = emessage + "@" + tiktok.Config.BotName + " dupe trello board [mcboard]\n"
	emessage = emessage + "@" + tiktok.Config.BotName + " well retro card [mcboard] this sprint went awesome!\n"
	emessage = emessage + "@" + tiktok.Config.BotName + " description history pBxxmKI6\n"
	emessage = emessage + "@" + tiktok.Config.BotName + " previous sprint points [mcboard] mcboard-08-25-2018\n"

	testPayload.Text = message
	testPayload.Channel = userInfo.ID
	attachments.Color = "#5EA7B1"
	attachments.Text = hmessage
	testPayload.Attachments = append(testPayload.Attachments, attachments)
	attachments.Color = "#AAA999"
	attachments.Text = emessage
	testPayload.Attachments = append(testPayload.Attachments, attachments)

	_ = WranglerDM(tiktok, testPayload)

	return
}
