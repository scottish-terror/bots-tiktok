package tiktokmod

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/nlopes/slack"
)

// DownloadFile - download (stream copy) a file from a URL to the local file system
func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

// errTrap - Generic error handling function
func errTrap(tiktok *TikTokConf, message string, err error) {
	var attachments Attachment

	if tiktok.Config.DEBUG {
		fmt.Println(message + "(" + err.Error() + ")")
	}
	if tiktok.Config.LogToSlack {
		attachments.Color = "#ff0000"
		attachments.Text = err.Error()
		LogToSlack(message, tiktok, attachments)
	}
}

// Between - find string between two chars
func Between(value string, a string, b string) string {
	posFirst := strings.Index(value, a)
	if posFirst == -1 {
		return ""
	}
	posLast := strings.Index(value, b)
	if posLast == -1 {
		return ""
	}
	posFirstAdjusted := posFirst + len(a)
	if posFirstAdjusted >= posLast {
		return ""
	}
	return value[posFirstAdjusted:posLast]
}

func amInslice(validDates []time.Time, rightnow time.Time) bool {
	for _, x := range validDates {
		if x.Format("2006-01-02") == rightnow.Format("2006-01-02") {
			return true
		}
	}
	return false
}

// SliceExists - Determine if value is in a slice
func SliceExists(tiktok *TikTokConf, slice interface{}, item interface{}) bool {
	s := reflect.ValueOf(slice)

	if s.Kind() != reflect.Slice {
		if tiktok.Config.DEBUG {
			fmt.Println("SliceExists() given a non-slice type")
		}
		return false
	}

	for i := 0; i < s.Len(); i++ {
		if s.Index(i).Interface() == item {
			return true
		}
	}

	return false
}

// PointCleanup - module to syncronize points between Plugins and Customfields
func PointCleanup(opts Config, tiktok *TikTokConf, teamID string) (rtnMessage string) {
	var attachments Attachment
	var listList []lists
	var apMessage string
	var tMessage string
	var err error

	listList = append(listList, lists{
		channelID:   opts.General.ReadyForWork,
		channelName: "Ready For Work",
	}, lists{
		channelID:   opts.General.Working,
		channelName: "Working",
	}, lists{
		channelID:   opts.General.ReadyForReview,
		channelName: "Ready for Review (PR)",
	}, lists{
		channelID:   opts.General.Done,
		channelName: "Done",
	})

	attachments.Color = ""
	attachments.Text = ""

	for l := range listList {

		if tiktok.Config.LogToSlack {
			LogToSlack("Sync'ing points on cards in `"+listList[l].channelName+"` on the `"+opts.General.TeamName+"` board.", tiktok, attachments)
			if listList[l].channelID == opts.General.ReadyForWork || listList[l].channelID == opts.General.Working || listList[l].channelID == opts.General.ReadyForReview {
				LogToSlack("I'm trolling the `"+listList[l].channelName+"` list cards in the `"+opts.General.TeamName+"` board for Point Changes.", tiktok, attachments)
			}
		}
		rtnMessage, tMessage, err = SyncPoints(teamID, listList[l].channelID, opts, tiktok)
		if err != nil {
			return "Errors, returning `from action.go`"
		}

		apMessage = apMessage + tMessage

	}

	if apMessage != "" {
		attachments.Text = apMessage
		attachments.Color = "#ff0000"
		Wrangler(tiktok.Config.SlackHook, "<!here> Points have been changed on these cards that are in the *current sprint*.", opts.General.ComplaintChannel, tiktok.Config.SlackEmoji, attachments)
	}

	return rtnMessage
}

// CleanBackLog - Clean-up BackLog
func CleanBackLog(opts Config, tiktok *TikTokConf) error {
	var attachments Attachment
	var nmessage string
	var faceCount int
	var customCount int
	var ancientCard int
	var numCards int
	var squadLabel int

	// Trello args maps for custom fields
	var m map[string]string
	m = make(map[string]string)
	m["fields"] = "name"
	m["customFieldItems"] = "true"

	if tiktok.Config.LogToSlack {
		LogToSlack("I'm checking the BackLog in the `"+opts.General.TeamName+"` and cleaning up those cards.", tiktok, attachments)
	}

	allTheThings, err := RetrieveAll(tiktok, opts.General.BoardID, "visible")
	if err != nil {
		errTrap(tiktok, "Trello error in RetrieveAll `actions.go` for `"+opts.General.TeamName+"` board", err)
		return err
	}

	for _, aTt := range allTheThings.Cards {
		if aTt.IDList == opts.General.BacklogID {
			numCards++

			//remove squad labels
			allSquads, err := GetDBSquads(tiktok, opts.General.BoardID)
			if err != nil {
				errTrap(tiktok, "Failed DB Call to get squad information in trello.go func `SquadPoints`", err)
				return err
			}
			for _, L := range allSquads {
				for _, lab := range aTt.Labels {
					if lab.ID == L.LabelID {
						err := removeLabel(aTt.ID, L.LabelID, tiktok)
						if err != nil {
							errTrap(tiktok, "Error from `removeLabel in `CleanBackLog` in `actions.go`", err)
						}
						squadLabel++
					}
				}

			}

			//remove faces
			if len(aTt.IDMembers) > 0 {
				for _, h := range aTt.IDMembers {
					err := RemoveHead(tiktok, aTt.ID, h)
					if err != nil {
						errTrap(tiktok, "Error in `RemoveHeads` called from `CleanBackLog` in `actions.go`", err)
					}
					faceCount++
				}
			}

			//clear custom fields
			for _, c := range aTt.CustomFieldItems {
				if c.IDCustomField == opts.General.CfpointsID {
					if c.Value.Number != "0" {
						err = PutCustomField(aTt.ID, opts.General.CfpointsID, tiktok, "text", "0")
						if err != nil {
							errTrap(tiktok, "Error from `PutCustomField` for *CFPOINTSID* in `CleanBackLog` in `actions.go`", err)
						}
						customCount++
					}
				}
				if c.IDCustomField == opts.General.CfsprintID {
					if c.Value.Text != "" {
						err = PutCustomField(aTt.ID, opts.General.CfsprintID, tiktok, "text", "")
						if err != nil {
							errTrap(tiktok, "Error from `PutCustomField` for *CFSPRINTID* in `CleanBackLog` in `actions.go`", err)
						}
						customCount++
					}
				}
			}

			//remove points
			// AS OF 8/14/2018 the Trello REST API does not support PUT/POST/DELETE methods against Trello Power-Up data.  You can only GET
			//   This means we can't clear/zero Story Points.

			//check card age
			value, cardListTime := GetTimePutList(opts.General.BacklogID, aTt.ID, opts, tiktok)

			if value {
				format := "2006-01-02 15:04:05"
				fmtTime := cardListTime.Format("2006-01-02 15:04:05")
				then, _ := time.Parse(format, fmtTime)

				date := time.Now()
				diff := date.Sub(then)
				days := int(diff.Hours() / 24)

				if days > opts.General.BackLogDays {
					// Currently just logs to logging that card is old.
					//  This is where expansion of what to do with super old cards would happen.  Alerts, etc...
					ancientCard++
					LogToSlack("*Card in BackLog is older then "+strconv.Itoa(opts.General.BackLogDays)+" days old @ "+strconv.Itoa(days)+"days*", tiktok, attachments)
				}
			}

		}

	}

	// message about cleaning up the backlog
	if faceCount > 0 {
		nmessage = "I removed " + strconv.Itoa(faceCount) + " faces of off cards.\n"
	} else {
		nmessage = "I didn't find any faces on cards to remove though!\n"
	}
	if ancientCard > 0 {
		nmessage = nmessage + "I found " + strconv.Itoa(ancientCard) + " ancient old cards and logged them. \n"
	} else {
		nmessage = nmessage + "I did not find any cards older then " + strconv.Itoa(opts.General.BackLogDays) + " days old to complain about.\n"
	}
	if customCount > 0 {
		nmessage = nmessage + "Cleaned up " + strconv.Itoa(customCount) + " custom card fields.\n"

	} else {
		nmessage = nmessage + "I didn't find any custom card fields I had to cleanup!!\n"
	}
	if squadLabel > 0 {
		nmessage = nmessage + "I removed " + strconv.Itoa(squadLabel) + " squad labels from cards.\n"
	} else {
		nmessage = nmessage + "I didn't find any old squad labels to remove.\n"
	}
	nmessage = nmessage + "There is a total of " + strconv.Itoa(numCards) + " cards in the backlog currently.\n"
	attachments.Color = "#00ff00"
	attachments.Text = nmessage
	Wrangler(tiktok.Config.SlackHook, "Team, I just troll'd the backlog for clean up. :sweep:", opts.General.ComplaintChannel, tiktok.Config.SlackEmoji, attachments)

	return nil
}

// ArchiveBacklog - Archive old cards in the backlog
func ArchiveBacklog(tiktok *TikTokConf, opts Config) (err error) {

	var message string
	var attachments Attachment
	var cardCount int
	var hushed bool

	allTheThings, err := RetrieveAll(tiktok, opts.General.BoardID, "visible")
	if err != nil {
		errTrap(tiktok, "Trello error in RetrieveAll `actions.go` for `"+opts.General.TeamName+"` board", err)
		return err
	}

	message = ""
	cardCount = 0

	for _, aTt := range allTheThings.Cards {
		if aTt.IDList == opts.General.BacklogID {

			// Ignore "hush" cards and "template" cards
			hushed = false
			for _, l := range aTt.Labels {
				if l.ID == opts.General.SilenceCardLabel || l.ID == opts.General.TemplateLabelID {
					hushed = true
				}
			}

			if !hushed {
				createDate, err := GetCreateDate(tiktok, aTt.ID)
				if err != nil {
					errTrap(tiktok, "Skipping card <"+aTt.URL+"|"+aTt.Name+"> due to error retrieve creation date in `ArchiveBackLog` `actions.go`", err)
				}

				format := "2006-01-02 15:04:05"
				fmtTime := createDate.Format("2006-01-02 15:04:05")
				then, _ := time.Parse(format, fmtTime)

				date := time.Now()
				diff := date.Sub(then)
				days := int(diff.Hours() / 24)

				if days > opts.General.BackLogDays {
					//archive it
					message = message + "<" + aTt.URL + "|" + aTt.Name + "> is " + strconv.Itoa(days) + " days old.\n"

					url := "https://api.trello.com/1/cards/" + aTt.ID + "?closed=true&key=" + tiktok.Config.Tkey + "&token=" + tiktok.Config.Ttoken

					req, err := http.NewRequest("PUT", url, nil)
					if err != nil {
						errTrap(tiktok, "Error archiving card "+aTt.URL+" during http.NewRequest in `ArchiveBacklog` `actions.go`", err)
					}
					client := &http.Client{}
					resp, err := client.Do(req)
					if err != nil {
						errTrap(tiktok, "Error archiving card "+aTt.URL+" during client.Do API PUT in `ArchiveBacklog` `actions.go`", err)
					}
					defer resp.Body.Close()

					cardCount = cardCount + 1

				}
			} else {
				if tiktok.Config.LogToSlack {
					attachments.Color = ""
					attachments.Text = ""
					LogToSlack("Skipping backlog archival of "+aTt.Name+" card beacuse it is hushed or a template card.", tiktok, attachments)
				}
			}
		}
	}

	attachments.Color = "#00ff00"
	attachments.Text = message
	Wrangler(tiktok.Config.SlackHook, "I archived "+strconv.Itoa(cardCount)+" card(s) in the `BackLog` that were greater then "+strconv.Itoa(opts.General.BackLogDays)+" old.  Here's the list:\n", opts.General.ComplaintChannel, tiktok.Config.SlackEmoji, attachments)

	if tiktok.Config.LogToSlack {
		attachments.Color = ""
		attachments.Text = ""
		LogToSlack("I've archived "+strconv.Itoa(cardCount)+" card(s) in the `BackLog` for team `"+opts.General.TeamName+"` that were greater then "+strconv.Itoa(opts.General.BackLogDays)+" old. See Slack Channel "+opts.General.ComplaintChannel+" for details.", tiktok, attachments)
	}

	return nil
}

// CleanDone - Clean Done column of old cards
func CleanDone(opts Config, tiktok *TikTokConf) (string, error) {

	var attachments Attachment
	var cardCount int
	var message string

	if tiktok.Config.LogToSlack {
		LogToSlack("I'm searching the Done List in the `"+opts.General.TeamName+"` board for cards that are older than "+strconv.Itoa(opts.General.ArchiveDoneDays)+" days and archiving them. ", tiktok, attachments)
	}

	allTheThings, err := RetrieveAll(tiktok, opts.General.BoardID, "visible")
	if err != nil {
		errTrap(tiktok, "Trello error in RetrieveAll `actions.go` for `"+opts.General.TeamName+"` board", err)
		return "Trello error in RetrieveAll `actions.go` for `" + opts.General.TeamName + "` board", err
	}

	cardCount = 0

	for _, aTt := range allTheThings.Cards {
		if aTt.IDList == opts.General.Done {
			value, cardListTime := GetTimePutList(opts.General.Done, aTt.ID, opts, tiktok)

			if value {
				format := "2006-01-02 15:04:05"
				fmtTime := cardListTime.Format("2006-01-02 15:04:05")
				then, _ := time.Parse(format, fmtTime)

				date := time.Now()
				diff := date.Sub(then)
				days := int(diff.Hours() / 24)

				if days > opts.General.ArchiveDoneDays {

					url := "https://api.trello.com/1/cards/" + aTt.ID + "?closed=true&key=" + tiktok.Config.Tkey + "&token=" + tiktok.Config.Ttoken

					req, err := http.NewRequest("PUT", url, nil)
					if err != nil {
						errTrap(tiktok, "", err)
						return "NewRequest Put", err
					}
					client := &http.Client{}
					resp, err := client.Do(req)
					if err != nil {
						errTrap(tiktok, "Error client.DO API Post `actions.go`", err)
						return "client.Do", err
					}
					defer resp.Body.Close()

					cardCount = cardCount + 1

					if tiktok.Config.LogToSlack {
						attachments.Color = ""
						attachments.Text = ""
						LogToSlack("Card _"+aTt.Name+"_ in the Done List on `"+opts.General.TeamName+"` board is more than *"+strconv.Itoa(opts.General.ArchiveDoneDays)+"* days old.  It has been archived. ", tiktok, attachments)
					}
					if tiktok.Config.DEBUG {
						fmt.Printf("Card %s is Days %s old.  Archiving Card!\n", aTt.Name, strconv.Itoa(days))
					}
				} else {
					if tiktok.Config.DEBUG {
						fmt.Printf("Card %s is Days %s old. NOT Archiving\n", aTt.Name, strconv.Itoa(days))
					}
					if tiktok.Config.LogToSlack {
						attachments.Color = ""
						attachments.Text = ""
						LogToSlack("Card _"+aTt.Name+"_ in the Done List on `"+opts.General.TeamName+"` board is only *"+strconv.Itoa(days)+"* days old.  NOT Archiving!. ", tiktok, attachments)
					}
				}

			}
		}

	}

	if cardCount == 0 {
		message = "Hey team, I just checked for archivable cards on " + opts.General.TeamName + " board and found zero older than " + strconv.Itoa(opts.General.ArchiveDoneDays) + " days, so I'm not doing any clean-up today. :beach_with_umbrella:"
	} else {
		message = "Hey team, I just archived " + strconv.Itoa(cardCount) + " cards in the `Done` list because they were more than " + strconv.Itoa(opts.General.ArchiveDoneDays) + " days old."
	}
	attachments.Color = ""
	attachments.Text = ""
	Wrangler(tiktok.Config.SlackHook, message, opts.General.ComplaintChannel, tiktok.Config.SlackEmoji, attachments)

	return "", nil
}

// Permissions - determine if user has permissions to do something
func Permissions(tiktok *TikTokConf, slackID string, action string, api *slack.Client, chkchannel string) bool {

	userInfo, _ := api.GetUserInfo(slackID)

	// Admin Perms
	if action == "admin" {
		ctx := context.Background()
		adminGroup, err := api.GetGroupInfoContext(ctx, chkchannel)
		if err != nil {
			errTrap(tiktok, "Error in Func `permissions` for AdminChannel ID.\n", err)
			return false
		}
		for f := range adminGroup.Members {
			channeluserInfo, err := api.GetUserInfo(adminGroup.Members[f])
			if err != nil {
				errTrap(tiktok, "Error in Func `permissions` for AdminChannel ID.", err)
				return false
			}
			if userInfo.Name == channeluserInfo.Name {
				return true
			}
		}
		return false
	}

	// Scrum Master Perms
	if action == "scrum" {
		ctx := context.Background()
		scrumGroup, err := api.GetGroupInfoContext(ctx, chkchannel)
		if err != nil {
			errTrap(tiktok, "Error in Func `permissions` for ScrumChannel ID", err)
			return false
		}
		for f := range scrumGroup.Members {
			channeluserInfo, _ := api.GetUserInfo(scrumGroup.Members[f])
			if userInfo.Name == channeluserInfo.Name {
				return true
			}
		}
		return false
	}
	return false
}

// PRSummary - Summarize PR Column
func PRSummary(opts Config, tiktok *TikTokConf) (output string, err error) {

	var attachments Attachment
	var message string

	if tiktok.Config.LogToSlack {
		attachments.Text = ""
		attachments.Color = ""
		LogToSlack("Checking for PR cards to return a list of active ones on `"+opts.General.TeamName+"` board", tiktok, attachments)
	}

	allTheThings, err := RetrieveAll(tiktok, opts.General.BoardID, "visible")
	if err != nil {
		errTrap(tiktok, "Trello error in PRSummary in `actions.go` for `"+opts.General.TeamName+"` board", err)
		return "Trello error in PRSummary in `trello.go` for `" + opts.General.TeamName + "` board", err
	}

	for _, aTt := range allTheThings.Cards {
		if aTt.IDList == opts.General.ReadyForReview {
			message = message + "<https://trello.com/c/" + aTt.ID + "|" + aTt.Name + ">\n"
		}
	}

	if message != "" {
		hmessage := "Reminder, here are the current PR's for discussion at Stand-up today:\n"
		attachments.Color = "#006400"
		attachments.Text = message
		Wrangler(tiktok.Config.SlackHook, hmessage, opts.General.ComplaintChannel, tiktok.Config.SlackEmoji, attachments)
		return "", nil
	}

	return "No PR Cards available", nil
}

// CountCards - function to count # of cards per theme in pre-sprint columns for reporting
func CountCards(opts Config, tiktok *TikTokConf, teamID string) (allThemes Themes, err error) {

	sOpts, err := GetDBSprint(tiktok, teamID)
	if err != nil {
		return allThemes, err
	}

	// Load label information from board
	allThemes, err = GetLabel(tiktok, opts.General.BoardID)
	if err != nil {
		errTrap(tiktok, "Failed trello call to get label information in actions.go func `CountCards`", err)
		return allThemes, err
	}

	allTheThings, err := RetrieveAll(tiktok, opts.General.BoardID, "visible")
	if err != nil {
		errTrap(tiktok, "Trello error from RetrieveAll in `CountCards` `trello.go` for `"+opts.General.TeamName+"` board", err)
		return allThemes, err
	}

	for _, aTt := range allTheThings.Cards {
		if aTt.IDList == opts.General.Upcoming || aTt.IDList == opts.General.Scoped {
			for _, labels := range aTt.Labels {
				for s, label := range allThemes {
					if labels.ID == label.ID {
						tPts := allThemes[s].Pts
						allThemes[s].Pts = tPts + 1
					}
				}
			}
		}
	}

	// write to db and output
	err = PutThemeCount(tiktok, allThemes, sOpts, teamID)
	if err != nil {
		return allThemes, err
	}

	return allThemes, nil
}

// SyncPoints - sync points between Agile power-up and custom field in the provided column
func SyncPoints(teamID string, listID string, opts Config, tiktok *TikTokConf) (messasge string, apMessage string, err error) {

	var attachments Attachment
	var existPoints string
	var foundField bool
	var sprintField bool

	sOpts, err := GetDBSprint(tiktok, teamID)
	if err != nil {
		errTrap(tiktok, "Failed DB query, bailing out of syncpoints function in `trello.go`", err)
		return "", "", err
	}

	allTheThings, err := RetrieveAll(tiktok, opts.General.BoardID, "all")

	if err != nil {
		errTrap(tiktok, "Trello error in RetrieveAll `trello.go` for `"+opts.General.TeamName+"` board", err)
		return "", "Trello error in RetrieveAll `trello.go` for `" + opts.General.TeamName + "` board", err
	}

	for _, aTt := range allTheThings.Cards {
		if aTt.IDList == listID {
			var points int

			pluginCard, _ := GetPowerUpField(aTt.ID, tiktok)

			foundField = false
			sprintField = false

			for _, p := range pluginCard {

				if p.IDPlugin == tiktok.Config.PointsPowerUpID {
					var plugins PointsHistory

					pluginJSON := []byte(p.Value)
					json.Unmarshal(pluginJSON, &plugins)
					points = plugins.Points
				}

			}

			for _, cusval := range aTt.CustomFieldItems {
				// sync points to burndown custom field
				if cusval.IDCustomField == opts.General.CfpointsID {
					existPoints = cusval.Value.Number
					foundField = true
				}

				// sync sprintname to custom field in specific lists
				if aTt.IDList == opts.General.ReadyForWork || aTt.IDList == opts.General.Working || aTt.IDList == opts.General.ReadyForReview {
					if cusval.IDCustomField == opts.General.CfsprintID {
						sprintField = true
						if cusval.Value.Text == "" || cusval.Value.Text != sOpts.SprintName {
							// Put custom field
							err := PutCustomField(aTt.ID, opts.General.CfsprintID, tiktok, "text", sOpts.SprintName)
							if err != nil {
								errTrap(tiktok, "Error in PutCustomField in trello.go, updating sprintname field", err)
							}
						}
					}
				}
			}

			// handle cards that have never had customfield SprintName created
			if !sprintField {
				if aTt.IDList == opts.General.ReadyForWork || aTt.IDList == opts.General.Working || aTt.IDList == opts.General.ReadyForReview {
					err := PutCustomField(aTt.ID, opts.General.CfsprintID, tiktok, "text", sOpts.SprintName)
					if err != nil {
						errTrap(tiktok, "Error in PutCustomField in trello.go, updating sprintname field", err)
					}
				}
			}

			// Check specific lists to see if points have been changed and add to alert if they have
			if aTt.IDList == opts.General.ReadyForWork || aTt.IDList == opts.General.Working || aTt.IDList == opts.General.ReadyForReview {
				if existPoints != strconv.Itoa(points) {
					if existPoints != "" && foundField && existPoints != "0" {
						apMessage = apMessage + "Points on card <https://trello.com/c/" + aTt.ID + "|" + aTt.Name + "> have changed from " + existPoints + " to " + strconv.Itoa(points) + "\n"
						if tiktok.Config.LogToSlack {
							LogToSlack("Points on card <https://trello.com/c/"+aTt.ID+"|"+aTt.Name+"> have changed from "+existPoints+" to "+strconv.Itoa(points), tiktok, attachments)
						}
					}
				}
			}

			// Sync points fields
			if existPoints != strconv.Itoa(points) {
				err = PutCustomField(aTt.ID, opts.General.CfpointsID, tiktok, "text", strconv.Itoa(points))
				if err != nil {
					errTrap(tiktok, "Error PutCustomField for Sync Fields `actions.go`", err)
				}
			}
		}
	}
	return "", apMessage, nil
}

// ThemePoints - retrieve all the theme points in a given trello colum (list)
func ThemePoints(opts Config, tiktok *TikTokConf, columnID string) (allThemes Themes, err error) {

	var points int

	// Load label information from board
	allThemes, err = GetLabel(tiktok, opts.General.BoardID)
	if err != nil {
		errTrap(tiktok, "Failed trello call to get label information in trello.go func `ThemePoints`", err)

		return allThemes, err
	}

	allTheThings, err := RetrieveAll(tiktok, opts.General.BoardID, "visible")
	if err != nil {
		errTrap(tiktok, "Trello error in RetrieveAll function `ThemePoints` in `actions.go` for `"+opts.General.TeamName+"` board", err)
		return allThemes, err
	}

	for _, aTt := range allTheThings.Cards {
		if columnID == aTt.IDList {

			// get power-up for story points
			pluginCard, _ := GetPowerUpField(aTt.ID, tiktok)

			for _, p := range pluginCard {

				if p.IDPlugin == tiktok.Config.PointsPowerUpID {

					var plugins PointsHistory

					pluginJSON := []byte(p.Value)
					json.Unmarshal(pluginJSON, &plugins)
					points = plugins.Points
				}
			}

			for _, labels := range aTt.Labels {
				for s, label := range allThemes {
					if labels.ID == label.ID {
						tPts := allThemes[s].Pts
						allThemes[s].Pts = tPts + points
					}
				}
			}
		}
	}

	return allThemes, err
}

// SquadPoints - retrieve all the squad points on a board
func SquadPoints(columnID string, opts Config, tiktok *TikTokConf) (allSquads Squads, nonPoints int, err error) {

	var points int
	var checker bool

	nonPoints = 0

	// Load Squad Information
	allSquads, err = GetDBSquads(tiktok, opts.General.BoardID)
	if err != nil {
		errTrap(tiktok, "Failed DB Call to get squad information in `actions.go` func `SquadPoints`", err)
		return allSquads, nonPoints, err
	}

	allTheThings, err := RetrieveAll(tiktok, opts.General.BoardID, "visible")
	if err != nil {
		errTrap(tiktok, "Trello error in RetrieveAll function `SquadPoints` in `actions.go` for `"+opts.General.TeamName+"` board", err)
		return allSquads, nonPoints, err
	}

	for _, aTt := range allTheThings.Cards {
		if !aTt.Closed {
			if aTt.IDList == columnID {

				// get power-up for story points
				pluginCard, _ := GetPowerUpField(aTt.ID, tiktok)

				for _, p := range pluginCard {

					if p.IDPlugin == tiktok.Config.PointsPowerUpID {

						var plugins PointsHistory

						pluginJSON := []byte(p.Value)
						json.Unmarshal(pluginJSON, &plugins)
						points = plugins.Points
					}
				}

				checker = false

				// update squad points
				for _, labels := range aTt.IDLabels {

					for s, squad := range allSquads {
						if opts.General.BoardID == squad.BoardID && squad.LabelID == labels {
							tPts := squad.SquadPts
							allSquads[s].SquadPts = tPts + points
							checker = true
						}
					}
				}
				if !checker {
					nonPoints = nonPoints + points
				}
			}
		}

	}

	return allSquads, nonPoints, nil
}

// EpicLink - Verify feature cards are linked to Epics
func EpicLink(tiktok *TikTokConf, opts Config) {
	var attachments Attachment
	var featureCard bool
	var linkedCard bool
	var amessage string
	var hush bool

	allTheThings, err := RetrieveAll(tiktok, opts.General.BoardID, "visible")
	if err != nil {
		errTrap(tiktok, "Trello error in EpicLink `actions.go` for `"+opts.General.TeamName+"` board", err)
		return
	}

	amessage = ""
	for _, aTt := range allTheThings.Cards {

		hush = false
		for _, l := range aTt.Labels {
			if l.ID == opts.General.SilenceCardLabel {
				hush = true
			}
		}

		if !hush {
			if aTt.IDList == opts.General.Upcoming || aTt.IDList == opts.General.Scoped || aTt.IDList == opts.General.ReadyForWork || aTt.IDList == opts.General.Working {
				if !aTt.Closed {
					featureCard = false
					linkedCard = false

					for _, l := range aTt.Labels {
						if strings.ToLower(l.Name) == "feature" {
							featureCard = true
						}
					}

					if featureCard {
						// check cards for any attachment back to Epic BoardID
						cardAttachment, err := GetAttachments(tiktok, aTt.ID)
						if err != nil {
							errTrap(tiktok, "Trello error in EpicLink `actions.go` for cardID `"+aTt.ID+"` board", err)

						} else {

							for _, c := range cardAttachment {
								if !c.IsUpload {
									u, _ := url.Parse(c.URL)
									if u.Host == "trello.com" {
										linkedCard = true
									}
								}
							}

							if !linkedCard {
								amessage = amessage + "<https://trello.com/c/" + aTt.ID + "|" + aTt.Name + ">\n"
							}
						}
					}
				}
			}
		}
	}

	if amessage != "" {
		attachments.Color = "#ff0000"
		attachments.Text = amessage
		Wrangler(tiktok.Config.SlackHook, "The following `Feature` cards do not have Epic links!", opts.General.ComplaintChannel, tiktok.Config.SlackEmoji, attachments)
	}

	return
}

// CheckThemes - Check that cards in a specific list have Theme Labels, returns formatted output
func CheckThemes(tiktok *TikTokConf, opts Config, listID string) (amessage string, err error) {

	allTheThings, err := RetrieveAll(tiktok, opts.General.BoardID, "visible")
	if err != nil {
		errTrap(tiktok, "Trello error in CheckThemes `actions.go` for `"+opts.General.TeamName+"` board", err)
		return "", err
	}

	amessage = ""
	for _, aTt := range allTheThings.Cards {
		if !aTt.Closed {
			if aTt.IDList == listID {
				if len(aTt.Labels) == 0 {
					amessage = amessage + "<https://trello.com/c/" + aTt.ID + "|" + aTt.Name + ">\n"
				}
			}
		}
	}

	return amessage, nil
}

// CardPlay - Pull card timing data and dump to CSV
func CardPlay(tiktok *TikTokConf, opts Config, channelResponse string, teamID string, csv bool) {
	var message string
	var wdays string
	var prdays string
	var header string
	var realName string
	var points int
	var WorkingDays int
	var attachments Attachment
	var diff time.Duration
	var today time.Time
	var allCardData CardReportData

	format := "2006-01-02 15:04:05"

	if csv {
		Wrangler(tiktok.Config.SlackHook, "Running card movement routine on `"+teamID+"`, this may take some time", channelResponse, tiktok.Config.SlackEmoji, attachments)
	}
	if tiktok.Config.LogToSlack {
		LogToSlack("Running Card movement routines on "+teamID+".", tiktok, attachments)
	}

	// Trello args maps for custom fields
	var m map[string]string
	m = make(map[string]string)
	m["fields"] = "name"
	m["customFieldItems"] = "true"

	sOpts, err := GetDBSprint(tiktok, teamID)
	if err != nil {
		errTrap(tiktok, "DB Error on `GetDBSprint` in `actions.go` for `CardPlay` func", err)
		return
	}

	allTheThings, err := RetrieveAll(tiktok, opts.General.BoardID, "visible")
	if err != nil {
		errTrap(tiktok, "Trello error in RetrieveAll `cardplay.go` for `"+opts.General.TeamName+"` board", err)
		return
	}

	message = "Card ID,Card Title,Points,Card URL,List,Started in Working,Days,Started in PR,Days,Entered Done,Owners\n"

	if !csv {
		if tiktok.Config.LogToSlack {
			LogToSlack("Truncating tiktok_cardtracker to prepare for new data", tiktok, attachments)
		}
		err := zeroCardDataDB(tiktok)
		if err != nil {
			return
		}
	}

	for _, aTt := range allTheThings.Cards {
		if !aTt.Closed {
			if aTt.IDList == opts.General.ReadyForWork || aTt.IDList == opts.General.Working || aTt.IDList == opts.General.ReadyForReview || aTt.IDList == opts.General.Done {
				for _, cusval := range aTt.CustomFieldItems {
					if cusval.IDCustomField == opts.General.CfsprintID {
						if cusval.Value.Text == sOpts.SprintName {

							powerUp, _ := GetPowerUpField(aTt.ID, tiktok)
							for p := range powerUp {

								var plugins PointsHistory

								pluginJSON := []byte(powerUp[p].Value)
								json.Unmarshal(pluginJSON, &plugins)
								points = plugins.Points
							}

							header = ""
							for _, head := range aTt.IDMembers {
								fullname, _, _ := GetMemberInfo(head, tiktok)
								header = header + fullname + "|"
							}

							// Get Date for each list
							tz, err := time.LoadLocation("America/Los_Angeles")
							if err != nil {
								errTrap(tiktok, "TZ Error", err)
								return
							}
							_, cardListTime := GetTimePutList(opts.General.Working, aTt.ID, opts, tiktok)

							cardTimeW := cardListTime.In(tz)
							workingTime := cardTimeW.Format("2006-01-02 15:04:05")
							if strings.Contains(workingTime, "0000-12-31 ") {
								workingTime = ""
								cardTimeW = time.Date(2000, 01, 01, 00, 00, 0, 0, time.UTC)
							}

							_, cardListTime = GetTimePutList(opts.General.ReadyForReview, aTt.ID, opts, tiktok)
							cardTimePR := cardListTime.In(tz)
							PRTime := cardTimePR.Format("2006-01-02 15:04:05")
							if strings.Contains(PRTime, "0000-12-31 ") {
								PRTime = ""
								cardTimePR = time.Date(2000, 01, 01, 00, 00, 0, 0, time.UTC)

							}
							_, cardListTime = GetTimePutList(opts.General.Done, aTt.ID, opts, tiktok)
							cardTimeD := cardListTime.In(tz)
							DoneTime := cardTimeD.Format("2006-01-02 15:04:05")
							if strings.Contains(DoneTime, "0000-12-31 ") {
								DoneTime = ""
								cardTimeD = time.Date(2000, 01, 01, 00, 00, 0, 0, time.UTC)

							}

							// Calc days in lists
							today = time.Now()

							then, _ := time.Parse(format, workingTime)
							if PRTime == "" {
								diff = today.Sub(then)
							} else {
								diff = cardTimePR.Sub(then)
							}
							WorkingDays = int(diff.Hours() / 24)
							if WorkingDays > 30 {
								wdays = ""
							} else {
								wdays = strconv.Itoa(WorkingDays)
							}

							then, _ = time.Parse(format, PRTime)
							if DoneTime == "" {
								diff = today.Sub(then)
							} else {
								diff = cardTimeD.Sub(then)
							}
							UATDays := int(diff.Hours() / 24)

							if UATDays > 30 {
								prdays = ""
							} else {
								prdays = strconv.Itoa(UATDays)
							}

							list, _ := GetLists(tiktok, opts.General.BoardID)
							for _, listName := range list {
								if listName.ID == aTt.IDList {
									realName = listName.Name
								}
							}

							if csv {
								// dump csv
								message = message + aTt.ID + "," + aTt.Name + "," + strconv.Itoa(points) + "," + aTt.ShortURL + "," + realName + "," + workingTime + "," + wdays + "," + PRTime + "," + prdays + "," + DoneTime + "," + header + "\n"
							} else {
								// write to DB
								allCardData.CardID = aTt.ID
								allCardData.CardTitle = aTt.Name
								allCardData.CardURL = aTt.ShortURL
								allCardData.EnteredDone = cardTimeD
								allCardData.List = realName
								allCardData.Owners = header
								allCardData.Points = points
								allCardData.StartedInPR = cardTimePR
								allCardData.StartedInWorking = cardTimeW

								err = PutCardData(tiktok, allCardData, teamID)
								if err != nil {
									errTrap(tiktok, "PutCardData error in `Cardplay` in `actions.go`", err)
									return
								}

							}
						}
					}
				}
			}
		}
	}

	tz, _ := time.LoadLocation("America/Los_Angeles")
	tnow := time.Now().In(tz)
	now := tnow.Format("01-02-2006-15:04")

	if csv {
		err = PostSnippet(tiktok, "csv", message, channelResponse, "Card-Data-"+now)

		if err != nil {
			Wrangler(tiktok.Config.SlackHook, "There was an error getting your information, please check the logs in #"+tiktok.Config.LogChannel, channelResponse, tiktok.Config.SlackEmoji, attachments)
		}
	} else {
		if channelResponse != "" {
			Wrangler(tiktok.Config.SlackHook, "Card movement data gathering complete, database updated", channelResponse, tiktok.Config.SlackEmoji, attachments)
		}
		LogToSlack("Card movement data gathering complete, database updated", tiktok, attachments)
	}

}

//GetColumn - determine which column was specified in a request (Default: BackLog)
func GetColumn(opts Config, someString string) (columnID string, colName string) {
	// check which column was specified if any
	lowString := strings.ToLower(someString)

	switch {
	case strings.Contains(lowString, "backlog"):
		columnID = opts.General.BacklogID
		colName = "Backlog"

	case strings.Contains(lowString, "upcoming"):
		columnID = opts.General.Upcoming
		colName = "Upcoming/Un-Scoped"
		break

	case strings.Contains(lowString, "un-scoped"):
		columnID = opts.General.Upcoming
		colName = "Upcoming/Un-Scoped"
		break

	case strings.Contains(lowString, "next sprint"):
		columnID = opts.General.NextsprintID
		colName = "Next Sprint"

	case strings.Contains(lowString, "ready for points"):
		columnID = opts.General.Scoped
		colName = "Ready for Points"

	case strings.Contains(lowString, "ready for work"):
		columnID = opts.General.ReadyForWork
		colName = "Ready for Work"

	case strings.Contains(lowString, "working"):
		columnID = opts.General.Working
		colName = "Working"

	case strings.Contains(lowString, "ready for pr"):
		columnID = opts.General.ReadyForReview
		colName = "Ready for Review"

	case strings.Contains(lowString, "ready for review"):
		columnID = opts.General.ReadyForReview
		colName = "Ready for Review"

	case strings.Contains(lowString, "done"):
		columnID = opts.General.Done
		colName = "Done"

	default:
		columnID = opts.General.BacklogID
		colName = "Backlog"
	}

	return columnID, colName
}

// RecordChapters - Record Chapter card info to SQL DB per specified column/list
func RecordChapters(tiktok *TikTokConf, teamID string, listName string) error {
	var columnID string
	var colName string

	opts, err := LoadConf(tiktok, teamID)
	if err != nil {
		errTrap(tiktok, "Load Conf Error for TeamID "+teamID, err)
		return err
	}

	columnID, colName = GetColumn(opts, listName)

	allChapters, _, err := ChapterCount(tiktok, opts, columnID)
	if err != nil {
		return err
	}

	for _, chapter := range allChapters {
		_ = RecordChapterCount(tiktok, chapter.ChapterName, colName, chapter.ChapterCount, teamID)
	}

	return nil
}

//RetroCheck - Check a specified Retro board for un-finished action cards
func RetroCheck(tiktok *TikTokConf, opts Config, boardID string) (err error) {
	var attachments Attachment
	var listID string
	var testPayload BotDMPayload

	users, err := GetDBUsers(tiktok)
	if err != nil {
		errTrap(tiktok, "Error getting user data from `GetDBUsers` in `RetroCheck` in `actions.go`", err)
		return
	}

	allTheThings, err := RetrieveAll(tiktok, boardID, "visible")
	if err != nil {
		errTrap(tiktok, "Trello error in RetroCheck in `trello.go` for `"+allTheThings.Name+"` ("+boardID+") retro board", err)
		return
	}

	if tiktok.Config.LogToSlack {
		LogToSlack("Scanning Retro Board `"+allTheThings.Name+"` for open action items.", tiktok, attachments)
	}

	// Get Actions column ListID from its name
	listData, err := GetLists(tiktok, boardID)
	if err != nil {
		return err
	}
	for _, listD := range listData {
		if strings.ToLower(listD.Name) == "action items" {
			listID = listD.ID
		}
	}

	if listID == "" {
		if tiktok.Config.LogToSlack {
			LogToSlack("No `Action Items` list found in Retro board "+allTheThings.Name+" so skipping it.", tiktok, attachments)
		}
	} else {
		for _, aTt := range allTheThings.Cards {
			if aTt.IDList == listID {
				if !aTt.Closed {
					// check date of last activity
					format := "2006-01-02 15:04:05"
					fmtTime := aTt.DateLastActivity.Format("2006-01-02 15:04:05")
					then, _ := time.Parse(format, fmtTime)

					date := time.Now()
					diff := date.Sub(then)
					days := int(diff.Hours() / 24)

					if days >= opts.General.RetroActionDays {
						if len(aTt.IDMembers) > 0 {
							for _, tu := range aTt.IDMembers {
								_, _, userName := GetMemberInfo(tu, tiktok)
								for _, u := range users {
									if userName == u.Trello {
										if tiktok.Config.LogToSlack {
											LogToSlack(tiktok.Config.BotName+" Retro Action Card: Sent warning to @"+u.Trello+" that this card `"+aTt.Name+"` is still not completed and has no activity within "+strconv.Itoa(opts.General.RetroActionDays)+" day warning period.", tiktok, attachments)
										}
										testPayload.Text = "*Warning!* You have a Retro Action Item that is still not complete and has no activity in the past " + strconv.Itoa(opts.General.RetroActionDays) + " days.\n<https://trello.com/c/" + aTt.ID + "|" + aTt.Name + ">"
										testPayload.Channel = u.SlackID

										err := WranglerDM(tiktok, testPayload)
										if err != nil {
											return err
										}
									}
								}
							}
						}
					}
				}
			}
		}

	}

	return nil
}

// CheckActionCards - Loop through retro boards and verify all retro cards are checked for in-action
func CheckActionCards(tiktok *TikTokConf, opts Config, teamID string) {

	var retroAll []RetroStruct

	// get sprint retros
	retroStruct, err := GetRetroID(tiktok, teamID)
	if err != nil {
		return
	}

	// get and append other retros in DB
	retroAdds, err := GetWBoards(tiktok)

	retroAll = append(retroStruct, retroAdds...)

	for _, r := range retroAll {
		err := RetroCheck(tiktok, opts, r.RetroID)
		if err != nil {
			return
		}
	}
}

//TemplateCard - Check for template cards and move them to top of backlog
func TemplateCard(tiktok *TikTokConf, opts Config) {

	var attachments Attachment

	LogToSlack("Scanning board "+opts.General.TeamName+" for template cards to ensure they are in the right spot.", tiktok, attachments)

	allTheThings, err := RetrieveAll(tiktok, opts.General.BoardID, "visible")
	if err != nil {
		errTrap(tiktok, "Trello error in TemplateCard `actions.go` for `"+opts.General.TeamName+"` board", err)
		return
	}

	for _, aTt := range allTheThings.Cards {
		if !aTt.Closed {

			for _, l := range aTt.Labels {
				if l.ID == opts.General.TemplateLabelID {

					if aTt.IDList != opts.General.BacklogID {
						err := MoveCardList(tiktok, aTt.ID, opts.General.BacklogID)
						if err != nil {
							return
						}
					}

					// change position
					err := CardPosition(tiktok, aTt.ID, "top")
					if err != nil {
						return
					}

				}
			}
		}
	}

	return
}
