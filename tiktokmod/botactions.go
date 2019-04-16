package tiktokmod

import (
	"database/sql"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/nlopes/slack"
	"github.com/robfig/cron"
)

// BotActions - TikTokConf Actions based on commands
func BotActions(lowerString string, tiktok *TikTokConf, ev *slack.MessageEvent, rtm *slack.RTM, api *slack.Client, c *cron.Cron, cronjobs *Cronjobs, CronState string) (*cron.Cron, *Cronjobs, string) {

	var attachments Attachment
	var smessage string
	var teamID string
	var labelID string
	var cardTitle string
	var testPayload BotDMPayload

	if strings.Contains(lowerString, "run builtin test") || strings.Contains(lowerString, "run built-in test") {
		//rtm.SendMessage(rtm.NewOutgoingMessage("Running Built-In test that you built! :unicornfart:", ev.Msg.Channel))
		rtm.SendMessage(rtm.NewOutgoingMessage("Currently no built-in tests!", ev.Msg.Channel))
		//opts, _ := LoadConf(tiktok, "mcboard")

		//SendAlert(tiktok, opts, "demo")
	}

	// Request that TikTokConf checks and alerts on non-active retro action item cards (this can also be CRON'd)
	if strings.Contains(lowerString, "check retro action activity") {
		teamID := Between(ev.Msg.Text, "[", "]")
		if teamID == "" {

			message := ListAllTOML(tiktok)
			attachments.Color = "#0000CC"
			attachments.Text = message

			Wrangler(tiktok.Config.SlackHook, "Please specify team in [ ] - Like `@"+tiktok.Config.BotName+" retro [mcboard]`\nHere's a list: ", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)
		} else {
			attachments.Color = ""
			attachments.Text = ""
			Wrangler(tiktok.Config.SlackHook, "Checking Sprint Retro boards for action items with no activity!", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)

			userInfo, _ := api.GetUserInfo(ev.Msg.User)
			LogToSlack(userInfo.Name+" asked me to check sprint retro boards for action items with no activity on `"+teamID+"` trello board.", tiktok, attachments)

			opts, err := LoadConf(tiktok, teamID)
			if err != nil {
				errTrap(tiktok, "Load Conf Error for TeamID "+teamID, err)
				return c, cronjobs, CronState
			}

			CheckActionCards(tiktok, opts, teamID)

			Wrangler(tiktok.Config.SlackHook, "Check process complete.", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)

		}
	}

	// Dump Chapter Cards by column
	if strings.Contains(lowerString, "chapter points") {
		teamID := Between(ev.Msg.Text, "[", "]")
		if teamID == "" {

			message := ListAllTOML(tiktok)
			attachments.Color = "#0000CC"
			attachments.Text = message

			Wrangler(tiktok.Config.SlackHook, "Please specify team in [ ] - Like `@"+tiktok.Config.BotName+" retro [mcboard]`\nHere's a list: ", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)
		} else {

			var columnID string
			var colName string
			var message string

			attachments.Color = ""
			attachments.Text = ""
			Wrangler(tiktok.Config.SlackHook, "One sec, let me add that up!", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)

			opts, err := LoadConf(tiktok, teamID)
			if err != nil {
				errTrap(tiktok, "Load Conf Error for TeamID "+teamID, err)
				return c, cronjobs, CronState
			}

			columnID, colName = GetColumn(opts, lowerString)

			userInfo, _ := api.GetUserInfo(ev.Msg.User)
			LogToSlack(userInfo.Name+" asked me to add up chapter points on `"+teamID+"` trello board in column `"+colName+"`.", tiktok, attachments)

			allChapters, noChapter, err := ChapterPoint(tiktok, opts, columnID)
			if err != nil {
				return c, cronjobs, CronState
			}
			for _, chapter := range allChapters {
				message = message + "Points for " + chapter.ChapterName + " = " + strconv.Itoa(chapter.ChapterPoints) + "\n"
			}

			message = message + "Points not assigned to a chapter: " + strconv.Itoa(noChapter) + "\n"
			attachments.Color = "#00ff00"
			attachments.Text = message
			Wrangler(tiktok.Config.SlackHook, "Chapter point counts", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)

			return c, cronjobs, CronState
		}
	}

	// Check for Critical bugs when asked
	if strings.Contains(lowerString, "are there any critical bugs") || strings.Contains(lowerString, "check for critical bugs") {
		teamID := Between(ev.Msg.Text, "[", "]")
		userInfo, _ := api.GetUserInfo(ev.Msg.User)
		if teamID == "" {
			teamID = "mcboard"
			userInfo, _ := api.GetUserInfo(ev.Msg.User)
			LogToSlack(userInfo.Name+" asked me to check for Critical Bugs and didn't specify a team, so defaulting to `mcboard` trello board.", tiktok, attachments)
			Wrangler(tiktok.Config.SlackHook, "You didn't specify a team/trello board so I'm assuming you mean `mcboard`.", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)

		}
		LogToSlack(userInfo.Name+" asked me to check for Critical Bugs on the "+teamID+" trello board.", tiktok, attachments)

		opts, err := LoadConf(tiktok, teamID)
		if err != nil {
			errTrap(tiktok, "Load Conf Error for TeamID "+teamID, err)
			return c, cronjobs, CronState
		}

		attachments.Color = ""
		attachments.Text = ""
		Wrangler(tiktok.Config.SlackHook, "One sec, I will check and then alert the "+opts.General.ComplaintChannel+" channel if I find any.", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)

		numBug := CheckBugs(opts, tiktok)

		if numBug == 0 {
			Wrangler(tiktok.Config.SlackHook, "I didn't find any critical bugs, Sweet!", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)
		} else {
			Wrangler(tiktok.Config.SlackHook, "I found a critical bug quantity of "+strconv.Itoa(numBug)+"!", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)
		}
		return c, cronjobs, CronState
	}

	// Dump Chapter Cards by column into the Database - defaults to backlog if no column specified
	if strings.Contains(lowerString, "record chapter count ") {
		teamID := Between(ev.Msg.Text, "[", "]")
		if teamID == "" {

			message := ListAllTOML(tiktok)
			attachments.Color = "#0000CC"
			attachments.Text = message

			Wrangler(tiktok.Config.SlackHook, "Please specify team in [ ] - Like `@"+tiktok.Config.BotName+" retro [mcboard]`\nHere's a list: ", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)
		} else {

			var colName string

			attachments.Color = ""
			attachments.Text = ""
			Wrangler(tiktok.Config.SlackHook, "One sec, let me add that up and record it to the database!", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)

			opts, err := LoadConf(tiktok, teamID)
			if err != nil {
				errTrap(tiktok, "Load Conf Error for TeamID "+teamID, err)
				return c, cronjobs, CronState
			}

			_, colName = GetColumn(opts, lowerString)

			userInfo, _ := api.GetUserInfo(ev.Msg.User)
			LogToSlack(userInfo.Name+" asked me to record in the DB the card count on chapter cards on `"+teamID+"` trello board in column `"+colName+"`.", tiktok, attachments)

			err = RecordChapters(tiktok, teamID, colName)

			if err != nil {
				Wrangler(tiktok.Config.SlackHook, "Something went wrong, please check the logs in the log channel #"+tiktok.Config.LogChannel, ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)
				return c, cronjobs, CronState
			}

			Wrangler(tiktok.Config.SlackHook, "Data recorded!", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)
			return c, cronjobs, CronState
		}
	}

	// Dump Chapter Cards by column
	if strings.Contains(lowerString, "check chapters") || strings.Contains(lowerString, "chapter count ") {
		teamID := Between(ev.Msg.Text, "[", "]")
		if teamID == "" {

			message := ListAllTOML(tiktok)
			attachments.Color = "#0000CC"
			attachments.Text = message

			Wrangler(tiktok.Config.SlackHook, "Please specify team in [ ] - Like `@"+tiktok.Config.BotName+" retro [mcboard]`\nHere's a list: ", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)
		} else {

			var columnID string
			var colName string
			var message string

			attachments.Color = ""
			attachments.Text = ""
			Wrangler(tiktok.Config.SlackHook, "One sec, let me add that up!", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)

			opts, err := LoadConf(tiktok, teamID)
			if err != nil {
				errTrap(tiktok, "Load Conf Error for TeamID "+teamID, err)
				return c, cronjobs, CronState
			}

			columnID, colName = GetColumn(opts, lowerString)

			userInfo, _ := api.GetUserInfo(ev.Msg.User)
			LogToSlack(userInfo.Name+" asked me to add report on chapter cards on `"+teamID+"` trello board in column `"+colName+"`.", tiktok, attachments)

			allChapters, totalCards, err := ChapterCount(tiktok, opts, columnID)
			if err != nil {
				return c, cronjobs, CronState
			}
			for _, chapter := range allChapters {
				message = message + "Cards for " + chapter.ChapterName + " = " + strconv.Itoa(chapter.ChapterCount) + "\n"
			}
			message = message + "Total Cards in column " + colName + " = " + strconv.Itoa(totalCards) + "\n"

			attachments.Color = "#00ff00"
			attachments.Text = message
			Wrangler(tiktok.Config.SlackHook, "Chapter card counts", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)

			return c, cronjobs, CronState
		}
	}

	// Grab Card Timing Data on Command
	if strings.Contains(lowerString, "get card data") {
		teamID := Between(ev.Msg.Text, "[", "]")
		if teamID == "" {

			message := ListAllTOML(tiktok)
			attachments.Color = "#0000CC"
			attachments.Text = message

			Wrangler(tiktok.Config.SlackHook, "Please specify team in [ ] - Like `@"+tiktok.Config.BotName+" retro [mcboard]`\nHere's a list: ", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)
		} else {
			userInfo, _ := api.GetUserInfo(ev.Msg.User)
			LogToSlack(userInfo.Name+" asked me to get all the card timing data on `"+teamID+"` trello board", tiktok, attachments)

			opts, err := LoadConf(tiktok, teamID)
			if err != nil {
				rtm.SendMessage(rtm.NewOutgoingMessage("I couldn't find the team config file ("+teamID+".toml) you asked for!.", ev.Msg.Channel))
				return c, cronjobs, CronState
			}

			rtm.SendMessage(rtm.NewOutgoingMessage("Attempting to pull card timing data.\n*Warning* This can take several minutes, please wait patiently. :knuckles_waiting:", ev.Msg.Channel))

			if strings.Contains(lowerString, "DB ONLY") {
				CardPlay(tiktok, opts, ev.Msg.Channel, teamID, false)
			} else {
				CardPlay(tiktok, opts, ev.Msg.Channel, teamID, true)
			}
			LogToSlack("Completed retrieving card timing on `"+teamID+"` trello board for "+userInfo.Name, tiktok, attachments)
		}
	}

	// Check That cards have a Theme
	if strings.Contains(lowerString, "check themes") {
		var amessage string
		var temp string

		teamID := Between(ev.Msg.Text, "[", "]")
		if teamID == "" {

			message := ListAllTOML(tiktok)
			attachments.Color = "#0000CC"
			attachments.Text = message

			Wrangler(tiktok.Config.SlackHook, "Please specify team in [ ] - Like `@"+tiktok.Config.BotName+" retro [mcboard]`\nHere's a list: ", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)
		} else {
			userInfo, _ := api.GetUserInfo(ev.Msg.User)
			LogToSlack(userInfo.Name+" asked me to verify Theme labels on `"+teamID+"` trello board", tiktok, attachments)

			opts, err := LoadConf(tiktok, teamID)
			if err != nil {
				rtm.SendMessage(rtm.NewOutgoingMessage("I couldn't find the team config file ("+teamID+".toml) you asked for!.", ev.Msg.Channel))
				return c, cronjobs, CronState
			}

			temp, _ = CheckThemes(tiktok, opts, opts.General.Upcoming)
			amessage = amessage + temp
			temp, _ = CheckThemes(tiktok, opts, opts.General.Scoped)
			amessage = amessage + temp
			temp, _ = CheckThemes(tiktok, opts, opts.General.ReadyForWork)
			amessage = amessage + temp

			if amessage != "" {
				attachments.Color = "#ff0000"
				attachments.Text = amessage
				Wrangler(tiktok.Config.SlackHook, "*WARNING*! The following cards do *not* have appropriate Theme Labels on them: ", opts.General.ComplaintChannel, tiktok.Config.SlackEmoji, attachments)
			} else {
				rtm.SendMessage(rtm.NewOutgoingMessage("Hurray all cards have theme labels!", ev.Msg.Channel))
			}
		}
	}

	// Retrieve Previous Card Description
	if strings.Contains(lowerString, "description history ") {
		var cardID string
		var msgBreak []string

		amessage := ""

		if strings.Contains(strings.ToLower(lowerString), "@"+strings.ToLower(tiktok.Config.BotID)) {

			msgBreak = strings.SplitAfterN(lowerString, " ", 4)
			if len(msgBreak) != 4 {

				rtm.SendMessage(rtm.NewOutgoingMessage("I'm not sure what you are asking me to do.", ev.Msg.Channel))
				return c, cronjobs, CronState

			}
			cardID = msgBreak[3]

		} else {

			msgBreak = strings.SplitAfterN(lowerString, " ", 3)
			if len(msgBreak) != 3 {

				rtm.SendMessage(rtm.NewOutgoingMessage("I'm not sure what you are asking me to do.", ev.Msg.Channel))
				return c, cronjobs, CronState

			}
			cardID = msgBreak[2]
		}

		descHistory, err := GetDescHistory(tiktok, cardID)
		if err != nil {
			errTrap(tiktok, "Error in retrieve card description history for card "+cardID+" function GetDescHistory in trello.go", err)
			return c, cronjobs, CronState
		}

		// Build message
		if len(descHistory) > 0 {
			cardTitle = descHistory[0].Data.Card.Name

		}
		userInfo, _ := api.GetUserInfo(ev.Msg.User)

		for _, dh := range descHistory {
			amessage = amessage + "*Date:* " + dh.Date.Format("2006-01-02 15:04:05\n")
			amessage = amessage + "*Editor:* " + dh.MemberCreator.FullName + "\n"
			amessage = amessage + "*Desc:* " + dh.Data.Card.Desc + "\n\n"
		}

		testPayload.Text = "Description history for requested card (" + cardTitle + "): "
		testPayload.Channel = userInfo.ID
		attachments.Color = "#00ff00"
		attachments.Text = amessage
		testPayload.Attachments = append(testPayload.Attachments, attachments)

		err = WranglerDM(tiktok, testPayload)
		if err != nil {
			errTrap(tiktok, "Issue sending Direct Slack message to "+userInfo.Name+" when card history was requested.", err)
			rtm.SendMessage(rtm.NewOutgoingMessage("Issue sending Direct Slack message to "+userInfo.Name+" when card history was requested for `"+cardTitle+"`!", ev.Msg.Channel))

			return c, cronjobs, CronState
		}

		rtm.SendMessage(rtm.NewOutgoingMessage("I have DM'd you the history description of the card `"+cardTitle+"`!", ev.Msg.Channel))

	}

	// Retrieve Holiday List
	if strings.Contains(strings.ToLower(lowerString), "company holidays") {
		var holidaymsg string
		var year string

		if strings.Contains(strings.ToLower(lowerString), "company holidays all") {
			year = "0"
		} else {
			t := time.Now()
			year = t.Format("2006")
		}

		holiday, err := GetHoliday(tiktok, year)

		if err != nil {
			rtm.SendMessage(rtm.NewOutgoingMessage("Sorry this data was unavailable, please check my logs.", ev.Msg.Channel))
			if tiktok.Config.LogToSlack {
				LogToSlack("Error retrieving Holiday dates from DB. "+err.Error(), tiktok, attachments)
			}
			return c, cronjobs, CronState
		}

		for _, h := range holiday {
			holidayDate := h.Day.Format("01/02/2006")
			holidaymsg = holidaymsg + holidayDate + " - " + h.Name + "\n"
		}

		attachments.Color = "#0000ff"
		attachments.Text = holidaymsg
		Wrangler(tiktok.Config.SlackHook, "Known Holidays for "+year+":", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)

	}

	// Record todays points
	if strings.Contains(lowerString, "record points for") {
		teamID := Between(ev.Msg.Text, "[", "]")
		if teamID == "" {

			message := ListAllTOML(tiktok)

			attachments.Color = "#0000CC"
			attachments.Text = message
			Wrangler(tiktok.Config.SlackHook, "Please specify team in [ ] - Like `@"+tiktok.Config.BotName+" retro [mcboard]`\nHere's a list: ", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)

		} else {

			userInfo, _ := api.GetUserInfo(ev.Msg.User)

			if Permissions(tiktok, ev.Msg.User, "scrum", api, tiktok.Config.ScrumControlChannel) {

				LogToSlack(userInfo.Name+" asked me to record all board points for "+teamID+".", tiktok, attachments)

				rtm.SendMessage(rtm.NewOutgoingMessage("Permissions accepted. Checking points for ["+teamID+"] this may take a moment.", ev.Msg.Channel))

				opts, _ := LoadConf(tiktok, teamID)
				sOpts, _ := GetDBSprint(tiktok, teamID)
				message, valid := GetAllPoints(tiktok, opts, sOpts)

				if valid {
					hmessage := "Recording today's sprint points for *" + opts.General.TeamName + "*\n"
					rtm.SendMessage(rtm.NewOutgoingMessage(hmessage+message, ev.Msg.Channel))
				}

			} else {

				smessage = "You are not the boss of me! Permission denied."
				rtm.SendMessage(rtm.NewOutgoingMessage(smessage, ev.Msg.Channel))
				LogToSlack(userInfo.Name+" asked me to record todays points for "+teamID+"  but did not have permissions so I ignored them.", tiktok, attachments)

			}
		}
	}

	// Counting Cards - Record Theme Card #'s for reporting
	if strings.Contains(lowerString, "count cards") {
		teamID := Between(ev.Msg.Text, "[", "]")
		if teamID == "" {

			message := ListAllTOML(tiktok)
			attachments.Color = "#0000CC"
			attachments.Text = message

			Wrangler(tiktok.Config.SlackHook, "Please specify team in [ ] - Like `@"+tiktok.Config.BotName+" retro [mcboard]`\nHere's a list: ", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)

		} else {

			userInfo, _ := api.GetUserInfo(ev.Msg.User)
			LogToSlack(userInfo.Name+" asked me to count up the Theme cards on `"+teamID+"` trello board.", tiktok, attachments)

			opts, err := LoadConf(tiktok, teamID)
			if err != nil {
				rtm.SendMessage(rtm.NewOutgoingMessage("I couldn't find the team config file ("+teamID+".toml) you asked for!.", ev.Msg.Channel))
				return c, cronjobs, CronState
			}

			attachments.Color = ""
			attachments.Text = ""
			Wrangler(tiktok.Config.SlackHook, "Hold please while I count some cards!", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)

			allThemes, err := CountCards(opts, tiktok, teamID)
			if err != nil {
				rtm.SendMessage(rtm.NewOutgoingMessage("Hrm, something went a foul, please check the logs.", ev.Msg.Channel))
				return c, cronjobs, CronState
			}

			// Should Sort Array DESC by points
			sort.Slice(allThemes, func(i, j int) bool {
				return allThemes[i].Pts > allThemes[j].Pts
			})

			amessage := ""
			for _, s := range allThemes {
				amessage = amessage + "Total `" + s.Name + "` Cards: " + strconv.Itoa(s.Pts) + "\n"
			}

			attachments.Color = "#0000ff"
			attachments.Text = amessage
			Wrangler(tiktok.Config.SlackHook, "Number of cards per theme (label) in `Un-Scoped` and `Ready for Points` on "+opts.General.TeamName+" board:", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)

		}
	}

	// Add up Theme Points
	if strings.Contains(lowerString, "theme points") {
		teamID := Between(ev.Msg.Text, "[", "]")
		if teamID == "" {

			message := ListAllTOML(tiktok)
			attachments.Color = "#0000CC"
			attachments.Text = message

			Wrangler(tiktok.Config.SlackHook, "Please specify team in [ ] - Like `@"+tiktok.Config.BotName+" retro [mcboard]`\nHere's a list: ", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)
		} else {

			var columnID string
			var colName string
			var myInt int
			var myPerc float64

			attachments.Color = ""
			attachments.Text = ""
			Wrangler(tiktok.Config.SlackHook, "One sec, let me add that up!", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)

			opts, err := LoadConf(tiktok, teamID)
			if err != nil {
				rtm.SendMessage(rtm.NewOutgoingMessage("I couldn't find the team config file ("+teamID+".toml) you asked for!.", ev.Msg.Channel))
				return c, cronjobs, CronState
			}

			// check which column was specified if any
			lowString := strings.ToLower(lowerString)

			if strings.Contains(lowString, "next sprint") {
				columnID = opts.General.NextsprintID
				colName = "Next Sprint"
			} else if strings.Contains(lowString, "ready for points") {
				columnID = opts.General.Scoped
				colName = "Ready for Points"
			} else if strings.Contains(lowString, "ready for work") {
				columnID = opts.General.ReadyForWork
				colName = "Ready for Work"
			} else {
				columnID = opts.General.NextsprintID
				colName = "Next Sprint"
			}

			// check if total points was specified for % calcs
			tP := Between(ev.Msg.Text, "{", "}")
			if tP != "" {
				myInt, err = strconv.Atoi(tP)
				if err != nil {
					errTrap(tiktok, "Integer Conversion Error", err)
					rtm.SendMessage(rtm.NewOutgoingMessage("There's an issue with what you specified in { } as total points.  This must be an integer! You specified: "+tP, ev.Msg.Channel))
					return c, cronjobs, CronState
				}
			}
			userInfo, _ := api.GetUserInfo(ev.Msg.User)
			LogToSlack(userInfo.Name+" asked me to add up the Theme points on `"+teamID+"` trello board in column `"+colName+"`.", tiktok, attachments)

			allThemes, err := ThemePoints(opts, tiktok, columnID)
			if err != nil {
				errTrap(tiktok, "Trello Error", err)
				rtm.SendMessage(rtm.NewOutgoingMessage("There seems to be an issue with this RetroID in Trello, I can't retrieve this information. ("+err.Error()+")", ev.Msg.Channel))
				return c, cronjobs, CronState
			}

			// Should Sort Array DESC by points
			sort.Slice(allThemes, func(i, j int) bool {
				return allThemes[i].Pts > allThemes[j].Pts
			})

			// Grab ignore labels
			ignoreLabels, err := GetIgnoreLabels(tiktok, opts.General.BoardID)
			if err != nil {
				rtm.SendMessage(rtm.NewOutgoingMessage("Gack there was an error! ("+err.Error()+")", ev.Msg.Channel))
				return c, cronjobs, CronState
			}

			// Build Output
			amessage := ""
			for _, s := range allThemes {
				if SliceExists(tiktok, ignoreLabels, s.ID) {
					if tiktok.Config.DEBUG {
						fmt.Println("Skipping " + s.Name)
					}
				} else {
					if tP != "" {
						deci := float64(s.Pts) / float64(myInt)
						myPerc = deci * 100.0
						amessage = amessage + "Total `" + s.Name + "` Points: " + strconv.Itoa(s.Pts) + " - (" + strconv.FormatFloat(myPerc, 'f', 0, 64) + "%)\n"
					} else {
						amessage = amessage + "Total `" + s.Name + "` Points: " + strconv.Itoa(s.Pts) + "\n"
					}
				}
			}
			attachments.Color = "#0000ff"
			attachments.Text = amessage
			Wrangler(tiktok.Config.SlackHook, "Points per label (Theme)  in `"+colName+"` on "+opts.General.TeamName+" board:", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)
		}

	}

	// Check sprint points for previous sprint (squads and totals)
	// @tiktok previous sprint points [mcboard] mcboard-07-25-2018
	if strings.Contains(lowerString, "previous sprint points") {

		teamID := Between(ev.Msg.Text, "[", "]")
		if teamID == "" {

			message := ListAllTOML(tiktok)

			attachments.Color = "#0000CC"
			attachments.Text = message

			Wrangler(tiktok.Config.SlackHook, "Please specify team inside the [ ]\nHere's a list: ", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)

		} else {

			var msgBreak []string
			var locale int
			var amessage string
			var myTotal int

			//break down message
			if strings.Contains(strings.ToLower(lowerString), "@"+strings.ToLower(tiktok.Config.BotID)) {

				msgBreak = strings.SplitAfterN(lowerString, " ", 6)
				if len(msgBreak) != 6 {

					rtm.SendMessage(rtm.NewOutgoingMessage("I'm not sure what you are asking me to do.", ev.Msg.Channel))
					return c, cronjobs, CronState

				}
				locale = 5

			} else {

				msgBreak = strings.SplitAfterN(lowerString, " ", 5)
				if len(msgBreak) != 5 {

					rtm.SendMessage(rtm.NewOutgoingMessage("I'm not sure what you are asking me to do.", ev.Msg.Channel))
					return c, cronjobs, CronState

				}
				locale = 4

			}

			opts, err := LoadConf(tiktok, teamID)
			if err != nil {
				errTrap(tiktok, "Load Conf Error for TeamID in `previous sprint points` command "+teamID, err)
				return c, cronjobs, CronState
			}

			sprintLow := strings.ToLower(msgBreak[locale])
			userInfo, _ := api.GetUserInfo(ev.Msg.User)
			LogToSlack(userInfo.Name+" asked me to retrieve the squad points for previous sprint `"+sprintLow+"` on `"+teamID+"` trello board", tiktok, attachments)

			attachments.Color = ""
			attachments.Text = ""
			Wrangler(tiktok.Config.SlackHook, "Let me grab the point data for `"+sprintLow+"`!", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)

			myTotal = 0
			SprintPoints, err := GetPreviousSprintPoints(tiktok, sprintLow)
			if err != nil {
				rtm.SendMessage(rtm.NewOutgoingMessage("Something went totally wrong, please check the logs.", ev.Msg.Channel))
			} else {
				amessage = ""
				for _, s := range SprintPoints {
					amessage = amessage + "Total `" + s.SquadName + "` Points: " + strconv.Itoa(s.SprintPoints) + "\n"
					myTotal = myTotal + s.SprintPoints
				}
				amessage = amessage + "Total Sprint Points:" + strconv.Itoa(myTotal)
				attachments.Color = "#0000ff"
				attachments.Text = amessage
				Wrangler(tiktok.Config.SlackHook, "Points per squad for sprint `"+sprintLow+"` on "+opts.General.TeamName+" board:", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)
			}
		}

	}

	// Add up squad points
	if strings.Contains(lowerString, "squad points") {
		teamID := Between(ev.Msg.Text, "[", "]")
		if teamID == "" {

			message := ListAllTOML(tiktok)
			attachments.Color = "#0000CC"
			attachments.Text = message

			Wrangler(tiktok.Config.SlackHook, "Please specify team in [ ] - Like `@"+tiktok.Config.BotName+" retro [mcboard]`\nHere's a list: ", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)
		} else {

			var columnID string
			var colName string

			attachments.Color = ""
			attachments.Text = ""
			Wrangler(tiktok.Config.SlackHook, "One sec, let me add that up!", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)

			opts, err := LoadConf(tiktok, teamID)
			if err != nil {
				errTrap(tiktok, "Load Conf Error for TeamID "+teamID, err)
				return c, cronjobs, CronState
			}

			// check which column was specified if any
			lowString := strings.ToLower(lowerString)

			if strings.Contains(lowString, "next sprint") {
				columnID = opts.General.NextsprintID
				colName = "Next Sprint"
			} else if strings.Contains(lowString, "ready for points") {
				columnID = opts.General.Scoped
				colName = "Ready for Points"
			} else if strings.Contains(lowString, "ready for work") {
				columnID = opts.General.ReadyForWork
				colName = "Ready for Work"
			} else if strings.Contains(lowString, "working") {
				columnID = opts.General.Working
				colName = "Working"
			} else {
				columnID = opts.General.NextsprintID
				colName = "Next Sprint"
			}

			userInfo, _ := api.GetUserInfo(ev.Msg.User)
			LogToSlack(userInfo.Name+" asked me to add up the squad points on `"+teamID+"` trello board in column `"+colName+"`.", tiktok, attachments)

			allSquads, nonPoints, err := SquadPoints(columnID, opts, tiktok)

			amessage := ""
			for _, s := range allSquads {
				if opts.General.BoardID == s.BoardID {
					amessage = amessage + "Total `" + s.Squadname + "` Points: " + strconv.Itoa(s.SquadPts) + "\n"
				}
			}
			amessage = amessage + "Total Points not assigned to a squad: " + strconv.Itoa(nonPoints) + "\n"
			attachments.Color = "#0000ff"
			attachments.Text = amessage
			Wrangler(tiktok.Config.SlackHook, "Points per squad in `"+colName+"` on "+opts.General.TeamName+" board:", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)

		}
	}

	// List retro board
	if strings.Contains(lowerString, "retro board") {
		teamID := Between(ev.Msg.Text, "[", "]")
		if teamID == "" {

			message := ListAllTOML(tiktok)

			attachments.Color = "#0000CC"
			attachments.Text = message

			Wrangler(tiktok.Config.SlackHook, "Please specify team in [ ] - Like `@"+tiktok.Config.BotName+" retro [mcboard]`\nHere's a list: ", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)

		} else {
			sOpts, err := GetDBSprint(tiktok, teamID)
			if err != nil {
				rtm.SendMessage(rtm.NewOutgoingMessage("Sorry I couldn't find what you were asking for! - ", ev.Msg.Channel))
				return c, cronjobs, CronState
			}

			allTheThings, err := RetrieveAll(tiktok, sOpts.RetroID, "none")
			if err != nil {
				errTrap(tiktok, "Error from `RetrieveAll` getting board info in `botactions.go` retro board command ", err)
				rtm.SendMessage(rtm.NewOutgoingMessage("There seems to be an issue with this RetroID in Trello, I can't retrieve this information. Please see logs.", ev.Msg.Channel))
				return c, cronjobs, CronState
			}

			message := "The current sprint reto board for `" + sOpts.SprintName + "` is <" + allTheThings.ShortURL + "|" + allTheThings.Name + ">"
			rtm.SendMessage(rtm.NewOutgoingMessage(message, ev.Msg.Channel))
		}
	}

	// List all cards in PR Column
	ls := strings.ToLower(lowerString)
	if strings.Contains(ls, "list pr") || strings.Contains(ls, "open pr") {
		teamID := Between(ev.Msg.Text, "[", "]")
		if teamID == "" {
			rtm.SendMessage(rtm.NewOutgoingMessage("I did not understand which board you want, sorry.", ev.Msg.Channel))
		} else {
			opts, err := LoadConf(tiktok, teamID)
			if err != nil {
				errTrap(tiktok, "Load Conf Error for TeamID "+teamID, err)
			} else {
				userInfo, err := api.GetUserInfo(ev.Msg.User)
				if err != nil {
					errTrap(tiktok, "api.GetUserInfo function error returned in `botactions.go`", err)
				}
				LogToSlack(userInfo.Name+" asked me to list the PR's on `"+teamID+"` trello board.", tiktok, attachments)

				output, err := PRSummary(opts, tiktok)
				if err != nil {
					errTrap(tiktok, "PRSummary function error returned in `botactions.go`", err)
					return c, cronjobs, CronState
				}

				if output == "" {
					rtm.SendMessage(rtm.NewOutgoingMessage("Dumping list of PRs to main Slack channel, per your request.", ev.Msg.Channel))
				} else {
					rtm.SendMessage(rtm.NewOutgoingMessage(output, ev.Msg.Channel))

				}

			}
		}
	}

	// Dupe a board
	if strings.Contains(lowerString, "dupe trello board") || strings.Contains(lowerString, "copy trello board") {
		teamID := Between(ev.Msg.Text, "[", "]")
		if teamID == "" {
			rtm.SendMessage(rtm.NewOutgoingMessage("I did not understand which board you want, sorry.", ev.Msg.Channel))
		} else {
			opts, err := LoadConf(tiktok, teamID)
			if err != nil {
				errTrap(tiktok, "Load Conf Error for TeamID "+teamID, err)
			} else {
				rtm.SendMessage(rtm.NewOutgoingMessage("Hold please, attempting to dupe it up.", ev.Msg.Channel))

				userInfo, _ := api.GetUserInfo(ev.Msg.User)
				LogToSlack(userInfo.Name+" asked me to make a dupe of the `"+teamID+"` trello board, so I'm doing that.", tiktok, attachments)

				allTheThings, err := RetrieveAll(tiktok, opts.General.BoardID, "none")
				if err != nil {
					errTrap(tiktok, "Error from `RetrieveAll` getting board info in `botactions.go` dupe trello board command ", err)
					rtm.SendMessage(rtm.NewOutgoingMessage("There seems to be an issue with this request in Trello, I can't retrieve this information. Please see logs.", ev.Msg.Channel))
					return c, cronjobs, CronState
				}

				rightnow := time.Now().Local()
				nameDate := rightnow.Format("01-02-06")
				dupeName := "DUPE-" + nameDate + ": " + allTheThings.Name
				output, _ := DupeTrelloBoard(allTheThings.ID, dupeName, opts.General.TrelloOrg, tiktok)

				if tiktok.Config.DEBUG {
					fmt.Println(output)
				}
				if tiktok.Config.LogToSlack {
					LogToSlack(output, tiktok, attachments)
				}

				rtm.SendMessage(rtm.NewOutgoingMessage(output, ev.Msg.Channel))

			}
		}
	}

	// Reload Cron Jobs from file
	if strings.Contains(lowerString, "reload cron") || strings.Contains(lowerString, "re-load cron") || strings.Contains(lowerString, "reload all cron") {
		userInfo, err := api.GetUserInfo(ev.Msg.User)
		LogToSlack(userInfo.Name+" asked me to re-load all the CRONJobs, so I'm attempting that.", tiktok, attachments)

		c.Stop()
		CronState = "Halted"

		cronjobs, c, err = CronLoad(tiktok)

		if err != nil {
			errTrap(tiktok, "CRON Jobs failed to load due to file error.", err)
			rtm.SendMessage(rtm.NewOutgoingMessage("CRON Jobs failed to load due to file error.", ev.Msg.Channel))
		} else {
			if tiktok.Config.DEBUG {
				fmt.Println("CRON Jobs were re-loaded from file!")
			}
			rtm.SendMessage(rtm.NewOutgoingMessage("CRON Jobs were re-loaded!", ev.Msg.Channel))
			CronState = "Running"
		}
	}

	// List Cron Jobs from file
	if strings.Contains(lowerString, "list all cronjobs") || strings.Contains(lowerString, "show me all cronjobs") || strings.Contains(lowerString, "list cronjobs") {
		var message string
		var attachments Attachment

		if CronState == "Not Loaded" {
			rtm.SendMessage(rtm.NewOutgoingMessage("The Cron State is currently not loaded.  Please Reload it!", ev.Msg.Channel))
			return c, cronjobs, CronState
		}
		userInfo, _ := api.GetUserInfo(ev.Msg.User)

		message = "Existing Cron State is: `" + CronState + "`\n"
		message = message + "```"
		for _, cr := range cronjobs.Cronjob {
			opts, err := LoadConf(tiktok, cr.Config)
			if err != nil {
				errTrap(tiktok, "Load Conf Error for TeamID "+teamID, err)
			}
			message = message + cr.Timing + " - " + opts.General.TeamName + " - " + cr.Action + "\n"
		}
		message = message + "```"

		testPayload.Text = "Current Running Cron Jobs"
		testPayload.Channel = userInfo.ID
		attachments.Color = "#0c15dd"
		attachments.Text = message
		testPayload.Attachments = append(testPayload.Attachments, attachments)

		err := WranglerDM(tiktok, testPayload)
		if err != nil {
			return c, cronjobs, CronState
		}

		rtm.SendMessage(rtm.NewOutgoingMessage("I have DM'd you the current cron jobs, lucky you!", ev.Msg.Channel))
	}

	// New Sprint Setup
	if strings.Contains(lowerString, "start a new sprint") {

		var rboard bool

		smessage = ""

		boardID := Between(ev.Msg.Text, "[", "]")
		if boardID == "" {
			rtm.SendMessage(rtm.NewOutgoingMessage("I did not understand which team you want, sorry.", ev.Msg.Channel))
		} else {
			userInfo, _ := api.GetUserInfo(ev.Msg.User)

			LogToSlack(userInfo.Name+" asked me to run a new sprint for the "+boardID+" configuration.", tiktok, attachments)

			if Permissions(tiktok, ev.Msg.User, "admin", api, tiktok.Config.ScrumControlChannel) {
				opts, err := LoadConf(tiktok, boardID)
				if err != nil {
					errTrap(tiktok, "Load Conf Error for TeamID "+teamID, err)
					rtm.SendMessage(rtm.NewOutgoingMessage("I couldn't find the team config file ("+boardID+".toml) you asked for!.", ev.Msg.Channel))
				} else {
					if strings.Contains(lowerString, "suppress retro") {
						rboard = true
						smessage = smessage + "I will supress creation of a Retro board\n"
					} else {
						rboard = false
					}
					smessage = smessage + "Permissions accepted, attempting to Sprint it up!"
					rtm.SendMessage(rtm.NewOutgoingMessage(smessage, ev.Msg.Channel))
					returnMsg, _ := Sprint(opts, tiktok, rboard)
					rtm.SendMessage(rtm.NewOutgoingMessage(returnMsg, ev.Msg.Channel))
				}
			} else {
				smessage = "You are not the boss of me! Permission denied."
				rtm.SendMessage(rtm.NewOutgoingMessage(smessage, ev.Msg.Channel))
				if tiktok.Config.LogToSlack {
					LogToSlack(userInfo.Name+" does not have the appropriate permissions and was told to `get bent`.", tiktok, attachments)
				}
			}
		}
	}

	// STOP Cron Jobs
	if strings.Contains(lowerString, "stop all cron") || strings.Contains(lowerString, "shutdown all cron") || strings.Contains(lowerString, "halt all cron") {
		userInfo, _ := api.GetUserInfo(ev.Msg.User)

		if Permissions(tiktok, ev.Msg.User, "scrum", api, tiktok.Config.ScrumControlChannel) {
			smessage = "Permissions accepted. Halting all Cron Jobs!"
			rtm.SendMessage(rtm.NewOutgoingMessage(smessage, ev.Msg.Channel))
			LogToSlack(userInfo.Name+" asked me to halt all Cron Jobs so I did.", tiktok, attachments)
			CronState = "Halted"
			c.Stop()
		} else {
			smessage = "You are not the boss of me! Permission denied."
			rtm.SendMessage(rtm.NewOutgoingMessage(smessage, ev.Msg.Channel))
			LogToSlack(userInfo.Name+" asked me to halt all Cron Jobs but did not have permissions so I ignored them.", tiktok, attachments)
		}

	}

	// Shutdown
	if strings.Contains(lowerString, "shutdown please") {
		userInfo, _ := api.GetUserInfo(ev.Msg.User)

		if Permissions(tiktok, ev.Msg.User, "admin", api, tiktok.Config.AdminSlackChannel) {
			smessage = "Permissions accepted. Okay logging off bye!"
			rtm.SendMessage(rtm.NewOutgoingMessage(smessage, ev.Msg.Channel))
			LogToSlack(userInfo.Name+" asked me to shutdown, so I am.", tiktok, attachments)
			duration := time.Duration(4) * time.Second
			time.Sleep(duration)
			os.Exit(0)
		} else {
			smessage = "You are not the boss of me! Permission denied."
			rtm.SendMessage(rtm.NewOutgoingMessage(smessage, ev.Msg.Channel))
			LogToSlack(userInfo.Name+" asked me to shut down, but did not have permissions so I ignored them.", tiktok, attachments)

		}

	}

	// Build a config file
	if strings.Contains(lowerString, "build a configuration file") {
		boardID := Between(ev.Msg.Text, "[", "]")
		BuildConfig(boardID, tiktok, ev.Msg.User, api)
		rtm.SendMessage(rtm.NewOutgoingMessage("Okay, I Direct Messaged your config to you.", ev.Msg.Channel))
	}

	// TROLL the board
	if strings.Contains(lowerString, "troll team") || strings.Contains(lowerString, "troll board") {

		attachments.Text = ""
		attachments.Color = ""

		teamID = Between(ev.Msg.Text, "[", "]")
		if teamID == "" {
			rtm.SendMessage(rtm.NewOutgoingMessage("I did not understand which team you want, sorry.", ev.Msg.Channel))
		} else {
			opts, err := LoadConf(tiktok, teamID)
			userInfo, _ := api.GetUserInfo(ev.Msg.User)

			if err != nil {
				errTrap(tiktok, "Load Conf Error for TeamID "+teamID, err)
				rtm.SendMessage(rtm.NewOutgoingMessage("I couldn't find the team config file ("+teamID+".toml) you asked for!.", ev.Msg.Channel))
			} else {
				LogToSlack(userInfo.Name+" asked me to Troll the "+teamID+" configuration.", tiktok, attachments)
				rtm.SendMessage(rtm.NewOutgoingMessage("Okay, running alerting on board for team "+teamID+".", ev.Msg.Channel))

				_, _ = AlertRunner(opts, tiktok)
			}
		}
	}

	// Clean BackLog (separate from archiving)
	if strings.Contains(lowerString, "clean the backlog") {

		attachments.Text = ""
		attachments.Color = ""

		teamID = Between(ev.Msg.Text, "[", "]")
		if teamID == "" {
			rtm.SendMessage(rtm.NewOutgoingMessage("I did not understand which team you want, sorry.", ev.Msg.Channel))
		} else {
			opts, err := LoadConf(tiktok, teamID)
			userInfo, _ := api.GetUserInfo(ev.Msg.User)

			if err != nil {
				errTrap(tiktok, "Load Conf Error for TeamID "+teamID, err)
				rtm.SendMessage(rtm.NewOutgoingMessage("I couldn't find the team config file ("+teamID+".toml) you asked for!.", ev.Msg.Channel))
			} else {
				LogToSlack(userInfo.Name+" asked me to clean the BackLog on the "+teamID+" configuration.", tiktok, attachments)
				rtm.SendMessage(rtm.NewOutgoingMessage("Okay, cleaning the backlog for team "+teamID+".", ev.Msg.Channel))

				err = CleanBackLog(opts, tiktok)
				if err != nil {
					errTrap(tiktok, "Error in `CleanBackLog` process run by slack command request.", err)
				}
			}
		}
	}

	// Archive the BackLog
	if strings.Contains(lowerString, "archive the backlog") {

		attachments.Text = ""
		attachments.Color = ""

		teamID = Between(ev.Msg.Text, "[", "]")
		if teamID == "" {
			rtm.SendMessage(rtm.NewOutgoingMessage("I did not understand which team you want, sorry.", ev.Msg.Channel))
		} else {
			opts, err := LoadConf(tiktok, teamID)
			userInfo, _ := api.GetUserInfo(ev.Msg.User)

			if err != nil {
				errTrap(tiktok, "Load Conf Error for TeamID "+teamID, err)
				rtm.SendMessage(rtm.NewOutgoingMessage("I couldn't find the team config file ("+teamID+".toml) you asked for!.", ev.Msg.Channel))
			} else {
				LogToSlack(userInfo.Name+" asked me to archive old cards in the `BackLog` on the "+teamID+" configuration.", tiktok, attachments)
				rtm.SendMessage(rtm.NewOutgoingMessage("Okay, archiving cards older then "+strconv.Itoa(opts.General.BackLogDays)+" days in the `BackLog` for team "+teamID+".", ev.Msg.Channel))

				err = ArchiveBacklog(tiktok, opts)
				if err != nil {
					errTrap(tiktok, "Error in `ArchiveBacklog` process run by slack command request.", err)
				}
			}
		}
	}

	// Run board archiving
	if strings.Contains(lowerString, "archiving on board") {

		attachments.Text = ""
		attachments.Color = ""

		teamID = Between(ev.Msg.Text, "[", "]")
		if teamID == "" {
			rtm.SendMessage(rtm.NewOutgoingMessage("I did not understand which team you want, sorry.", ev.Msg.Channel))
		} else {
			opts, err := LoadConf(tiktok, teamID)
			userInfo, _ := api.GetUserInfo(ev.Msg.User)

			if err != nil {
				errTrap(tiktok, "Load Conf Error for TeamID "+teamID, err)
				rtm.SendMessage(rtm.NewOutgoingMessage("I couldn't find the team config file ("+teamID+".toml) you asked for!.", ev.Msg.Channel))
			} else {
				LogToSlack(userInfo.Name+" asked me to run archiving on the "+teamID+" configuration.", tiktok, attachments)
				rtm.SendMessage(rtm.NewOutgoingMessage("Okay, running archiving on board for team "+teamID+".", ev.Msg.Channel))

				_, _ = CleanDone(opts, tiktok)
			}
		}
	}

	// Return list of bug labels
	if strings.Contains(lowerString, "show bug labels") {
		var bugmessage string

		attachments.Text = ""
		attachments.Color = ""

		teamID = Between(ev.Msg.Text, "[", "]")
		if teamID == "" {
			rtm.SendMessage(rtm.NewOutgoingMessage("I did not understand which team you want, sorry.", ev.Msg.Channel))
		} else {
			opts, err := LoadConf(tiktok, teamID)

			if err != nil {
				errTrap(tiktok, "Load Conf Error for TeamID "+teamID, err)
				rtm.SendMessage(rtm.NewOutgoingMessage("I couldn't find the team config file ("+teamID+".toml) you asked for!.", ev.Msg.Channel))
			} else {
				rtm.SendMessage(rtm.NewOutgoingMessage("Let me grab the bug labels for team "+teamID+".", ev.Msg.Channel))

				bugs, err := GetBugID(tiktok, opts.General.BoardID)
				if err != nil {
					errTrap(tiktok, "SQL Error returned from GetBugID in `botactions.go`", err)
				}
				if len(bugs) == 0 {
					rtm.SendMessage(rtm.NewOutgoingMessage("There are currently no Bug Labels identified for the team "+teamID+".", ev.Msg.Channel))
				} else {
					for _, b := range bugs {
						bugmessage = bugmessage + b.BugLevel + " label has ID " + b.LabelID + "\n"
					}
					attachments.Color = "#999999"
					attachments.Text = bugmessage
					Wrangler(tiktok.Config.SlackHook, "Found the following bug label info: ", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)
				}
			}
		}
	}

	// Scan for lagging PR
	if strings.Contains(lowerString, "scan for pr") {

		attachments.Text = ""
		attachments.Color = ""

		teamID = Between(ev.Msg.Text, "[", "]")
		if teamID == "" {
			rtm.SendMessage(rtm.NewOutgoingMessage("I did not understand which team you want, sorry.", ev.Msg.Channel))
		} else {
			opts, err := LoadConf(tiktok, teamID)
			userInfo, _ := api.GetUserInfo(ev.Msg.User)

			if err != nil {
				errTrap(tiktok, "Load Conf Error for TeamID "+teamID, err)
				rtm.SendMessage(rtm.NewOutgoingMessage("I couldn't find the team config file ("+teamID+".toml) you asked for!.", ev.Msg.Channel))
			} else {
				LogToSlack(userInfo.Name+" asked me to scan for stale PR's on the "+teamID+" configuration.", tiktok, attachments)
				rtm.SendMessage(rtm.NewOutgoingMessage("Okay, scanning for stale PR's on board for team "+teamID+".", ev.Msg.Channel))

				returnMsg, _ := StalePRcards(opts, tiktok)
				rtm.SendMessage(rtm.NewOutgoingMessage(returnMsg, ev.Msg.Channel))
			}
		}
	}

	// Sync board points to custom field and Alert on changing points
	if strings.Contains(lowerString, "sync points") {

		attachments.Text = ""
		attachments.Color = ""

		teamID = Between(ev.Msg.Text, "[", "]")
		if teamID == "" {
			rtm.SendMessage(rtm.NewOutgoingMessage("I did not understand which team you want, sorry.", ev.Msg.Channel))
		} else {
			opts, err := LoadConf(tiktok, teamID)
			userInfo, _ := api.GetUserInfo(ev.Msg.User)

			if err != nil {
				errTrap(tiktok, "Load Conf Error for TeamID "+teamID, err)
				rtm.SendMessage(rtm.NewOutgoingMessage("I couldn't find the team config file ("+teamID+".toml) you asked for!.", ev.Msg.Channel))
			} else {
				LogToSlack(userInfo.Name+" asked me to syncronize points on the "+teamID+" configuration.", tiktok, attachments)
				rtm.SendMessage(rtm.NewOutgoingMessage("Okay, syncronizing points on board for team "+teamID+".", ev.Msg.Channel))

				_ = PointCleanup(opts, tiktok, teamID)
			}
		}
	}

	// List all manageable trello boards and tomls
	if strings.Contains(lowerString, "is your trello board") || strings.Contains(lowerString, "list available boards") || strings.Contains(lowerString, "list all boards") || strings.Contains(lowerString, "show me all boards") || strings.Contains(lowerString, "list available trello boards") || strings.Contains(lowerString, "list all trello boards") {

		var attachments Attachment
		userInfo, _ := api.GetUserInfo(ev.Msg.User)

		message := ListAllTOML(tiktok)

		attachments.Color = "#0000CC"
		attachments.Text = message

		Wrangler(tiktok.Config.SlackHook, "Hey "+userInfo.Name+", I manage the following boards: ", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)

	}

	// Add me to user DB
	if strings.Contains(lowerString, "add me") || strings.Contains(lowerString, "register me") {
		var newUserData UserData
		var tempSID string
		var myPayload BotDMPayload

		userData := Between(ev.Msg.Text, "[", "]")
		if userData == "" {
			rtm.SendMessage(rtm.NewOutgoingMessage("I can help you register here's how:\n```add me [email,trello id,github id]```\n`Do not use quotes anywhere.`\nExample: ```@"+tiktok.Config.BotName+" add me [some.one@mydomain.com,someone12,someone-ea]```", ev.Msg.Channel))
		} else {
			userInfo, _ := api.GetUserInfo(ev.Msg.User)

			LogToSlack(userInfo.Name+" asked me to to register them in the user DB.", tiktok, attachments)

			brokeOut := strings.Split(userData, ",")
			if len(brokeOut) == 3 {
				newUserData.Name = strings.Replace(userInfo.Name, ".", " ", -1)
				newUserData.SlackID = userInfo.ID
				newUserData.Email = strings.ToLower(brokeOut[0])
				newUserData.Trello = strings.ToLower(brokeOut[1])
				newUserData.Github = strings.ToLower(brokeOut[2])

				// check if already registered
				db, status, err := ConnectDB(tiktok, "tiktok")
				if err != nil {
					if tiktok.Config.DEBUG {
						fmt.Println(err.Error())
					}
					if tiktok.Config.LogToSlack {
						LogToSlack("db.QueryRow error: "+err.Error(), tiktok, attachments)
					}
					return c, cronjobs, CronState
				}

				if status {

					err := db.QueryRow("SELECT slackid FROM tiktok_users where slackid=?", userInfo.ID).Scan(&tempSID)

					switch {
					case err == sql.ErrNoRows:
						if newUserData.Name == "" || newUserData.Email == "" || newUserData.SlackID == "" || newUserData.Trello == "" || newUserData.Github == "" {
							rtm.SendMessage(rtm.NewOutgoingMessage("Your data is Bungle in the Jungle, sorry I can't do this.", ev.Msg.Channel))
						} else {
							if AddDBUser(tiktok, newUserData) {
								rtm.SendMessage(rtm.NewOutgoingMessage("Awesome, I've registered your info!", ev.Msg.Channel))
								umessage := "Name: " + newUserData.Name + "\nE-Mail: " + newUserData.Email + "\nSlack: " + newUserData.SlackID + "\nTrello: " + newUserData.Trello + "\nGithub: " + newUserData.Github + "\n"

								myPayload.Text = "I've registered you as follows:"
								myPayload.Channel = userInfo.ID
								attachments.Color = "#00FF55"
								attachments.Text = umessage
								myPayload.Attachments = append(myPayload.Attachments, attachments)
								_ = WranglerDM(tiktok, myPayload)

								attachments.Color = "#00FF55"
								attachments.Text = umessage
								if tiktok.Config.LogToSlack {
									LogToSlack("A new user was registered per their request.", tiktok, attachments)
								}
							} else {
								rtm.SendMessage(rtm.NewOutgoingMessage("Something went horribly wrong I could not add your new user info!", ev.Msg.Channel))
							}
						}
					case err != nil:
						if tiktok.Config.DEBUG {
							fmt.Println(err.Error())
						}
						if tiktok.Config.LogToSlack {
							LogToSlack("db.QueryRow error: "+err.Error(), tiktok, attachments)
						}

						return c, cronjobs, CronState

					default:
						p := strings.Replace(userInfo.Name, ".", " ", -1)
						rtm.SendMessage(rtm.NewOutgoingMessage("Hey "+p+" I already have you registered. :cheers:", ev.Msg.Channel))
						if tiktok.Config.LogToSlack {
							LogToSlack(userInfo.Name+" is already registered!.", tiktok, attachments)
						}
					}
				}
			} else {
				rtm.SendMessage(rtm.NewOutgoingMessage("Your data is Bungle in the Jungle, sorry I can't do this.", ev.Msg.Channel))
			}
		}

		return c, cronjobs, CronState

	}

	// Add user to user DB
	if strings.Contains(lowerString, "add a new user") {
		var newUserData UserData

		userData := Between(ev.Msg.Text, "[", "]")
		if userData == "" {
			rtm.SendMessage(rtm.NewOutgoingMessage("I did not understand what you want me to do, sorry.\nFormat: ```add a new user [name,email,slackID,trello,github]```\nDo not use quotes. SlackID must be UID not username.", ev.Msg.Channel))
		} else {
			userInfo, _ := api.GetUserInfo(ev.Msg.User)

			LogToSlack(userInfo.Name+" asked me to to add data to the user DB.", tiktok, attachments)

			if Permissions(tiktok, ev.Msg.User, "scrum", api, tiktok.Config.ScrumControlChannel) {
				brokeOut := strings.Split(userData, ",")
				if len(brokeOut) == 5 {
					newUserData.Name = strings.ToLower(brokeOut[0])
					newUserData.Email = strings.ToLower(brokeOut[1])
					newUserData.SlackID = brokeOut[2]
					newUserData.Trello = brokeOut[3]
					newUserData.Github = brokeOut[4]

					if newUserData.Name == "" || newUserData.Email == "" || newUserData.SlackID == "" || newUserData.Trello == "" || newUserData.Github == "" {
						rtm.SendMessage(rtm.NewOutgoingMessage("Your data is Bungle in the Jungle, sorry I can't do this.", ev.Msg.Channel))
					} else {
						if AddDBUser(tiktok, newUserData) {
							rtm.SendMessage(rtm.NewOutgoingMessage("Awesome, i've added your new user info!", ev.Msg.Channel))
							if tiktok.Config.LogToSlack {
								umessage := "Name: " + newUserData.Name + "\nE-Mail: " + newUserData.Email + "\nSlack: " + newUserData.SlackID + "\nTrello: " + newUserData.Trello + "\nGithub: " + newUserData.Github + "\n"
								attachments.Color = "#00FF55"
								attachments.Text = umessage
								LogToSlack("Added new user info to json", tiktok, attachments)
							}
						} else {
							rtm.SendMessage(rtm.NewOutgoingMessage("Something went horribly wrong I could not add your new user info!", ev.Msg.Channel))

						}
					}
				} else {
					rtm.SendMessage(rtm.NewOutgoingMessage("Your data is Bungle in the Jungle, sorry I can't do this.", ev.Msg.Channel))
				}

			} else {
				smessage = "You are not the boss of me! Permission denied."
				rtm.SendMessage(rtm.NewOutgoingMessage(smessage, ev.Msg.Channel))
			}
		}
	}

	// Create retro card
	if strings.Contains(lowerString, "retro card") {

		var listName string
		var listID string

		teamID := Between(ev.Msg.Text, "[", "]")
		if teamID == "" {

			message := ListAllTOML(tiktok)

			attachments.Color = "#0000CC"
			attachments.Text = message

			Wrangler(tiktok.Config.SlackHook, "Please specify team in [ ] - Like `@"+tiktok.Config.BotName+" {well|wrong} retro card [mcboard] my card title`\nHere's a list: ", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)

		} else {
			var msgBreak []string
			var locale int

			//break down message
			if strings.Contains(strings.ToLower(lowerString), "@"+strings.ToLower(tiktok.Config.BotID)) {

				msgBreak = strings.SplitAfterN(lowerString, " ", 6)
				if len(msgBreak) != 6 {

					rtm.SendMessage(rtm.NewOutgoingMessage("I'm not sure what you are asking me to do.", ev.Msg.Channel))
					return c, cronjobs, CronState

				}
				locale = 1

			} else {

				msgBreak = strings.SplitAfterN(lowerString, " ", 5)
				if len(msgBreak) != 5 {

					rtm.SendMessage(rtm.NewOutgoingMessage("I'm not sure what you are asking me to do.", ev.Msg.Channel))
					return c, cronjobs, CronState

				}
				locale = 0

			}

			if strings.Contains(msgBreak[locale], "well") || strings.Contains(msgBreak[locale], "wrong") || strings.Contains(msgBreak[locale], "good") || strings.Contains(msgBreak[locale], "bad") || strings.Contains(msgBreak[locale], "improve") || strings.Contains(msgBreak[locale], "improvement") || strings.Contains(msgBreak[locale], "vent") {

				if strings.Contains(msgBreak[locale], "well") || strings.Contains(msgBreak[locale], "good") {
					listName = "What Went Well"
				} else if strings.Contains(msgBreak[locale], "vent") {
					listName = "Vent"
				} else {
					listName = "What Needs Improvement"
				}

				sOpts, err := GetDBSprint(tiktok, teamID)
				if err != nil {
					rtm.SendMessage(rtm.NewOutgoingMessage("Sorry I couldn't find what you were asking for! - ", ev.Msg.Channel))
					return c, cronjobs, CronState
				}

				allTheThings, err := RetrieveAll(tiktok, sOpts.RetroID, "none")
				if err != nil {
					errTrap(tiktok, "Attempting to add card to retro board and received RetrieveAll trello error: ", err)
					rtm.SendMessage(rtm.NewOutgoingMessage("Sorry somethings wrong with that trello board I can't do it!", ev.Msg.Channel))
					return c, cronjobs, CronState
				}

				allLists, err := GetLists(tiktok, allTheThings.ID)
				if err != nil {
					errTrap(tiktok, "Attempting to add card to retro board and received GetLists trello error: ", err)
					return c, cronjobs, CronState
				}

				for _, list := range allLists {
					if list.Name == listName {
						listID = list.ID
					}
				}

				if listID == "" {
					if tiktok.Config.LogToSlack {
						LogToSlack("Retro Board <"+allTheThings.ShortURL+"|"+allTheThings.Name+"> ("+allTheThings.ID+") is missing a column for `"+listName+"`", tiktok, attachments)
					}
					rtm.SendMessage(rtm.NewOutgoingMessage("Sorry somethings wrong with that trello board I can't find the `"+listName+"` column!", ev.Msg.Channel))
					return c, cronjobs, CronState
				}

				locale = locale + 4
				err = CreateCard(msgBreak[locale], listID, tiktok)

				rtm.SendMessage(rtm.NewOutgoingMessage("I created your card `"+msgBreak[locale]+"` on list `"+listName+"` in <"+allTheThings.ShortURL+"|"+allTheThings.Name+">", ev.Msg.Channel))

			} else {
				rtm.SendMessage(rtm.NewOutgoingMessage("Please specify card type first: What Went Well = `well` or What Went Wrong = `wrong`\n```@"+tiktok.Config.BotName+" well retro card [<team>] my card title```", ev.Msg.Channel))
			}

		}
	}

	// Add Ignore Label
	if strings.Contains(lowerString, "ignore label ") {

		teamID := Between(ev.Msg.Text, "[", "]")
		if teamID == "" {

			message := ListAllTOML(tiktok)
			attachments.Color = "#0000CC"
			attachments.Text = message
			Wrangler(tiktok.Config.SlackHook, "Please specify team in [ ] - Like `@"+tiktok.Config.BotName+" ignore label {myLabel} [mcboard]`\nHere's a list: ", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)

		} else {

			labelName := Between(ev.Msg.Text, "{", "}")
			if labelName == "" {
				attachments.Color = ""
				attachments.Text = ""
				Wrangler(tiktok.Config.SlackHook, "Please specify the label you want me to ignore inside curly braces {myLabel}", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)
			} else {
				// we have a label and a team
				opts, err := LoadConf(tiktok, teamID)
				if err != nil {
					errTrap(tiktok, "Can not find team "+teamID+" in `botactions.go` for `ignore label` action", err)
					rtm.SendMessage(rtm.NewOutgoingMessage("I can not find the team called `"+teamID+"` that you requested.", ev.Msg.Channel))
					return c, cronjobs, CronState
				}

				labelData, err := GetLabel(tiktok, opts.General.BoardID)
				if err != nil {
					errTrap(tiktok, "Error retrieving label data for board "+opts.General.BoardID+" in `trello.go` GetLabelData function", err)
					return c, cronjobs, CronState
				}

				for _, l := range labelData {
					if l.Name == labelName {
						labelID = l.ID
					}
				}

				if labelID != "" {
					err = LabelIgnore(opts, tiktok, labelID)
					if tiktok.Config.LogToSlack {
						userInfo, _ := api.GetUserInfo(ev.Msg.User)

						attachments.Color = ""
						attachments.Text = ""
						LogToSlack(userInfo.Name+" asked me to add the label "+labelName+" on board "+opts.General.TeamName+" to the ignore list", tiktok, attachments)
					}

					rtm.SendMessage(rtm.NewOutgoingMessage("I've added the label `"+labelName+"` on the *"+opts.General.TeamName+"* board to the Theme ignore list.", ev.Msg.Channel))

				} else {
					rtm.SendMessage(rtm.NewOutgoingMessage("I can not find the label called `"+labelName+"` that you requested.", ev.Msg.Channel))
					return c, cronjobs, CronState
				}
			}

		}
	}

	// Check Epic Links
	if strings.Contains(lowerString, "check epic links") {
		teamID := Between(ev.Msg.Text, "[", "]")
		if teamID == "" {

			message := ListAllTOML(tiktok)
			attachments.Color = "#0000CC"
			attachments.Text = message

			Wrangler(tiktok.Config.SlackHook, "Please specify team in [ ] - Like `@"+tiktok.Config.BotName+" retro [mcboard]`\nHere's a list: ", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)
		} else {
			opts, err := LoadConf(tiktok, teamID)
			if err != nil {
				errTrap(tiktok, "Can not find team "+teamID+" in `botactions.go` for `check epic links` action.", err)
				rtm.SendMessage(rtm.NewOutgoingMessage("I can not find the team called `"+teamID+"` that you requested.", ev.Msg.Channel))
				return c, cronjobs, CronState
			}

			EpicLink(tiktok, opts)
		}
	}

	// What time is it according to TikTok
	if strings.Contains(lowerString, "what time is it") {
		today := time.Now()
		workingTime := today.Format("2006-01-02 15:04:05")
		rtm.SendMessage(rtm.NewOutgoingMessage("My Time is: "+workingTime, ev.Msg.Channel))
	}

	// List registered users
	if strings.Contains(lowerString, "list registered users") || strings.Contains(lowerString, "get registered users") {

		var thisMessage string

		userInfo, _ := api.GetUserInfo(ev.Msg.User)
		LogToSlack(userInfo.Name+" asked me to for a list of all registered users in my Database.", tiktok, attachments)

		allUsers, err := GetDBUsers(tiktok)
		if err != nil {
			errTrap(tiktok, "Error returning from `GetDBUsers` in `botactions.go` when asked to `list registered users`", err)
			return c, cronjobs, CronState
		}

		for _, u := range allUsers {
			thisMessage = thisMessage + "*" + u.Name + "* : (Slack: `" + u.SlackID + "`) (Trello: `" + u.Trello + "`) (Github: `" + u.Github + "`)\n"
		}

		attachments.Color = "#12ffcc"
		attachments.Text = thisMessage
		Wrangler(tiktok.Config.SlackHook, "List of users registered with me currently: ", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)

	}

	// List all Github REPOS
	if strings.Contains(ev.Msg.Text, "list repo") || strings.Contains(ev.Msg.Text, "list github repo") {

		var numCount = 0
		var message = ""

		rtm.SendMessage(rtm.NewOutgoingMessage("Let me grab the Github Repo List, I will Direct Message you the list.", ev.Msg.Channel))
		userInfo, _ := api.GetUserInfo(ev.Msg.User)
		LogToSlack(userInfo.Name+" asked me to list all the GitHub REPOs in "+tiktok.Config.GithubOrgName, tiktok, attachments)

		repoList := RetrieveOrgRepo(tiktok, tiktok.Config.GithubOrgName)

		for _, r := range repoList {
			message = message + "• " + *r.Name + " - " + *r.HTMLURL + "\n"
			numCount++
		}

		testPayload.Text = "List of all *" + strconv.Itoa(numCount) + "* " + tiktok.Config.GithubOrgName + " REPOs:"
		testPayload.Channel = userInfo.ID
		attachments.Color = "#6600ff"
		attachments.Text = message
		testPayload.Attachments = append(testPayload.Attachments, attachments)
		_ = WranglerDM(tiktok, testPayload)

	}

	// List GitHub Users
	if strings.Contains(ev.Msg.Text, "list users github") {

		var numCount = 0
		var message = ""

		rtm.SendMessage(rtm.NewOutgoingMessage("Let me grab all the Trello users for you, one sec...I will Direct Message you the list.", ev.Msg.Channel))
		userInfo, _ := api.GetUserInfo(ev.Msg.User)
		LogToSlack(userInfo.Name+" asked me to list all the GitHub users in "+tiktok.Config.GithubOrgName, tiktok, attachments)

		userList := RetrieveUsers(tiktok, tiktok.Config.GithubOrgName)

		for _, u := range userList {
			if u.HTMLURL != nil {
				message = message + *u.Login + " - [" + strconv.Itoa(int(*u.ID)) + "] - " + *u.HTMLURL + "\n"
			} else {
				message = message + *u.Login + " - [" + strconv.Itoa(int(*u.ID)) + "]\n"
			}
			numCount++
		}

		testPayload.Text = "List of all *" + strconv.Itoa(numCount) + "* " + tiktok.Config.GithubOrgName + " Github Users: \n"
		testPayload.Channel = userInfo.ID
		attachments.Color = "#00ff00"
		attachments.Text = message
		testPayload.Attachments = append(testPayload.Attachments, attachments)
		_ = WranglerDM(tiktok, testPayload)
	}

	// List Open PRs on a Repo
	if strings.Contains(ev.Msg.Text, "list pull request") {

		var prMessage = ""
		var repoName = ""
		var locale int
		var msgBreak []string

		//break down message
		if strings.Contains(strings.ToLower(lowerString), "@"+strings.ToLower(tiktok.Config.BotID)) {

			msgBreak = strings.SplitAfterN(lowerString, " ", 5)
			if len(msgBreak) != 5 {

				rtm.SendMessage(rtm.NewOutgoingMessage("I'm not sure what you are asking me to do.", ev.Msg.Channel))
				return c, cronjobs, CronState

			}
			locale = 4

		} else {

			msgBreak = strings.SplitAfterN(lowerString, " ", 4)
			if len(msgBreak) != 5 {

				rtm.SendMessage(rtm.NewOutgoingMessage("I'm not sure what you are asking me to do.", ev.Msg.Channel))
				return c, cronjobs, CronState

			}
			locale = 3

		}

		repoName = msgBreak[locale]

		rtm.SendMessage(rtm.NewOutgoingMessage("Grabbing open PR list for repo `"+repoName+"`", ev.Msg.Channel))
		userInfo, _ := api.GetUserInfo(ev.Msg.User)
		LogToSlack(userInfo.Name+" asked me to list all the open PRs in repo `"+repoName+"` in "+tiktok.Config.GithubOrgName, tiktok, attachments)

		pullList, err := GitPRList(tiktok, repoName, tiktok.Config.GithubOrgName)
		if err != nil {
			rtm.SendMessage(rtm.NewOutgoingMessage("I couldn't find the Repo you wanted called `"+repoName+"`", ev.Msg.Channel))
			return c, cronjobs, CronState
		}

		if len(pullList) == 0 {
			Wrangler(tiktok.Config.SlackHook, "The repo requested `"+repoName+"` currently has no open Pull Requests", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)
			return c, cronjobs, CronState
		}

		for _, u := range pullList {
			loc, _ := time.LoadLocation("America/Los_Angeles")
			prUptime := *u.UpdatedAt
			lastUpdate := prUptime.In(loc).Format("2006-01-02 15:04:05")

			prMessage = prMessage + "Pull Request #" + strconv.Itoa(*u.Number) + " - <" + *u.HTMLURL + "|" + *u.Title + "> (Last Updated: `" + lastUpdate + " PT`)\n" // is " + *u.State  + " by <" + *u.User.HTMLURL + "|" + *u.User.Name + ">\n"

		}

		attachments.Text = prMessage
		attachments.Color = "#0000cc"
		Wrangler(tiktok.Config.SlackHook, "List of all *Open PRs* on Repo `"+repoName+"` in "+tiktok.Config.GithubOrgName+" Github: \n", ev.Msg.Channel, tiktok.Config.SlackEmoji, attachments)
	}

	return c, cronjobs, CronState
}
