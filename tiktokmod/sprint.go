package tiktokmod

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Sprint - Verify sprint is acceptable to execute then do or do not
func Sprint(opts Config, tiktok *TikTokConf, retroNo bool) (message string, err error) {
	var countcards int
	var countcardsbl int
	var newsprintcount int
	var attachments Attachment
	var weHaveSpike bool
	var sOpts SprintData
	var rboardID string
	var workingDays int
	var hush bool
	var retroMessage string
	var commentUpdate string

	// Trello args maps for custom fields
	var m map[string]string
	m = make(map[string]string)
	m["fields"] = "name"
	m["customFieldItems"] = "true"

	if tiktok.Config.DEBUG {
		fmt.Println("Executing Sprint Setup for `" + opts.General.TeamName + "` board!")
	}
	if tiktok.Config.LogToSlack {
		LogToSlack("Executing Sprint Setup for `"+opts.General.TeamName+"` board!", tiktok, attachments)
	}

	// Grab current sprint info
	spOpts, err := GetDBSprint(tiktok, strings.ToLower(opts.General.Sprintname))
	if err != nil {
		errTrap(tiktok, "GetDBSprint Error: SQL error in function `sprintgo` in `sprint.go`", err)
		return
	}
	_, _ = GetAllPoints(tiktok, opts, spOpts)

	// Record current Sprint squad point data to SQLDB
	squadTotals, nonPoints, err := SprintSquadPoints(tiktok, opts, spOpts.SprintName)
	if err != nil {
		errTrap(tiktok, "Failed to retrieve current sprint squad points for recording, check the logs. Continuing on...", err)
	}
	_ = RecordSquadSprintData(tiktok, squadTotals, spOpts.SprintName, nonPoints)

	// Dupe old cardtracker table to new table name for historical data
	tN := strings.Replace(spOpts.SprintName, "-", "_", -1)
	tableName := "tiktok_" + tN
	if tiktok.Config.LogToSlack {
		LogToSlack("Duplicating tiktok_cardtracker to new table `"+tableName+"` for historical records...this may take a few...", tiktok, attachments)
	}
	err = DupeTable(tiktok, tableName, "tiktok_cardtracker")
	if err != nil {
		errTrap(tiktok, "Error attempting to dupe table tiktok_cardtracker to "+tableName, err)
	}

	// create new sprint name
	rightnow := time.Now().Local()
	today := rightnow.Format("01-02-2006")
	newSprintName := opts.General.Sprintname + "-" + today

	// Load Squad Information
	allSquads, err := GetDBSquads(tiktok, opts.General.BoardID)
	if err != nil {
		errTrap(tiktok, "Failed DB Call to get squad information in sprint.go func `sprintgo`", err)
		return "Failed DB Call to get squad information", err
	}

	if tiktok.Config.DEBUG {
		fmt.Println("Created a new Sprint Name for `" + opts.General.TeamName + "` board - " + newSprintName)
	}
	if tiktok.Config.LogToSlack {
		LogToSlack("Created a new Sprint Name for `"+opts.General.TeamName+"` board - "+newSprintName, tiktok, attachments)
	}

	// Complain if cards don't have Theme Labels
	if tiktok.Config.LogToSlack {
		LogToSlack("Checking Next Sprint list for Card Themes on `"+opts.General.TeamName+"` board", tiktok, attachments)
	}
	jmessage, _ := CheckThemes(tiktok, opts, opts.General.NextsprintID)
	if jmessage != "" {
		attachments.Color = "#ff0000"
		attachments.Text = jmessage
		Wrangler(tiktok.Config.SlackHook, "*WARNING*! The following cards do *not* have appropriate Theme Labels on them: ", opts.General.ComplaintChannel, tiktok.Config.SlackEmoji, attachments)
	}
	attachments.Text = ""
	attachments.Color = ""

	allTheThings, err := RetrieveAll(tiktok, opts.General.BoardID, "visible")
	if err != nil {
		errTrap(tiktok, "Trello error in RetrieveAll function `sprintgo` in `sprint.go` for `"+opts.General.TeamName+"` board", err)
		return "Error in RetrieveAll cards API query, see logs.", err
	}

	for _, aTt := range allTheThings.Cards {
		if !aTt.Closed {

			if aTt.IDList == opts.General.ReadyForWork || aTt.IDList == opts.General.Working || aTt.IDList == opts.General.ReadyForReview {

				moveitmoveit := false
				for _, l := range aTt.Labels {

					if l.ID == opts.General.ROLabelID {
						moveitmoveit = true
					}

				}

				if moveitmoveit {
					commentUpdate = ""

					// move card to next sprint
					err := MoveCardList(tiktok, aTt.ID, opts.General.NextsprintID)

					if err != nil {
						errTrap(tiktok, "Error moving card `"+aTt.ID+"` to *Next Sprint* ... skipping", err)
					} else {
						if tiktok.Config.LogToSlack {
							LogToSlack("Moving card _"+aTt.Name+"_ ("+aTt.ID+") to *Next Sprint* column on `"+opts.General.TeamName+"` board.", tiktok, attachments)
						}
						countcards++
						commentUpdate = commentUpdate + "Moving incomplete card from current sprint, per WDW/planning discussions.\n"

						// sort card to top of sprint
						err := ReOrderCardInList(tiktok, aTt.ID, "top")
						if err != nil {
							errTrap(tiktok, "Couldn't not move card to top of list on card `"+aTt.Name+"` in `ReOrderCardInList` in `sprint.go`", err)
						} else {
							commentUpdate = commentUpdate + "Moving to top of list in priority per SDLC\n"
						}

						// remove ROLL-OVER Label from card
						err = removeLabel(aTt.ID, opts.General.ROLabelID, tiktok)
						if err != nil {
							errTrap(tiktok, "Couldn't remove Roll Over label on card `"+aTt.Name+"` in `ReOrderCardInList` in `sprint.go`", err)
						} else {
							commentUpdate = commentUpdate + "Removed ROLL-OVER label\n"
						}
						// add card comment
						err = CommentCard(aTt.ID, commentUpdate, tiktok)
						if err != nil {
							errTrap(tiktok, "Couldn't put change comments on card `"+aTt.Name+"` in `ReOrderCardInList` in `sprint.go`", err)
						}
					}

				} else {
					// move card to backlog
					err := MoveCardList(tiktok, aTt.ID, opts.General.BacklogID)
					if err != nil {
						errTrap(tiktok, "Error moving card `"+aTt.ID+"` to *Backlog* ... skipping", err)
					} else {
						if tiktok.Config.LogToSlack {
							LogToSlack("Moving card _"+aTt.Name+"_ ("+aTt.ID+") to *Backlog* column on `"+opts.General.TeamName+"` board.", tiktok, attachments)
						}
						countcardsbl++

						err = PutCustomField(aTt.ID, opts.General.CfsprintID, tiktok, "text", " ")
						if err != nil {
							errTrap(tiktok, "Trello error in PutCustomField `sprint.go` while moving card to backlog for `"+opts.General.TeamName+"` board", err)
						}
						err = CommentCard(aTt.ID, "Moving card to backlog from current sprint per WDW planning discussion.", tiktok)
						if err != nil {
							errTrap(tiktok, "Couldn't put change comments on card `"+aTt.Name+"` in `ReOrderCardInList` in `sprint.go`", err)
						}
					}

				}

			}
		}

	}

	// Move all cards in "Next Sprint" column to "Ready for Work"
	var totalPoints int
	var points int

	// re-read the board because we may have moved cards in the above function
	allTheThings, err = RetrieveAll(tiktok, opts.General.BoardID, "visible")
	if err != nil {
		errTrap(tiktok, "Trello error in RetrieveAll function `sprintgo` in `sprint.go` for `"+opts.General.TeamName+"` board", err)
		return "Error in RetrieveAll cards API query, see logs.", err
	}

	for _, aTt := range allTheThings.Cards {
		if !aTt.Closed {
			if aTt.IDList == opts.General.NextsprintID {

				// update custom field for sprint name
				for _, cusval := range aTt.CustomFieldItems {
					if cusval.IDCustomField == opts.General.CfsprintID {
						oldSprintname := string(cusval.Value.Text)
						commentUpdate := "Renaming sprint field from (" + oldSprintname + ") to " + newSprintName + "\n"
						_ = CommentCard(aTt.ID, commentUpdate, tiktok)
					}
				}
				err = PutCustomField(aTt.ID, opts.General.CfsprintID, tiktok, "text", newSprintName)
				if err != nil {
					errTrap(tiktok, "Trello error in PutCustomField `sprint.go` for `"+opts.General.TeamName+"` board", err)
				}

				// update custom field burndown story points
				pluginCard, _ := GetPowerUpField(aTt.ID, tiktok)

				for _, p := range pluginCard {

					if p.IDPlugin == tiktok.Config.PointsPowerUpID {

						var plugins PointsHistory

						pluginJSON := []byte(p.Value)
						json.Unmarshal(pluginJSON, &plugins)
						points = plugins.Points
						totalPoints = totalPoints + points
					}
				}
				spoints := strconv.Itoa(points)
				err = PutCustomField(aTt.ID, opts.General.CfpointsID, tiktok, "number", spoints)
				if err != nil {
					errTrap(tiktok, "Trello error in PutCustomField `sprint.go` trying to update burndown custom point field for `"+opts.General.TeamName+"` board", err)
				}

				// update squad points
				for _, labels := range aTt.Labels {

					for s, squad := range allSquads {
						if opts.General.BoardID == squad.BoardID && squad.LabelID == labels.ID {
							tPts := squad.SquadPts
							allSquads[s].SquadPts = tPts + points
							if tiktok.Config.DEBUG {
								fmt.Println(squad.Squadname + " found so adding " + strconv.Itoa(points) + " to the existing " + strconv.Itoa(tPts) + " for total of " + strconv.Itoa(allSquads[s].SquadPts))
							}
							if tiktok.Config.LogToSlack {
								attachments.Color = ""
								attachments.Text = ""
								LogToSlack(squad.Squadname+" found so adding "+strconv.Itoa(points)+" to the existing "+strconv.Itoa(tPts)+" for total of "+strconv.Itoa(allSquads[s].SquadPts), tiktok, attachments)
							}
						}
					}

				}

				hush = false

				for _, labels := range aTt.Labels {
					if labels.ID == opts.General.SilenceCardLabel {
						hush = true
					}
				}

				if !hush {

					attachments.Color = ""
					attachments.Text = ""

					// Remove any members from the card
					for _, m := range aTt.IDMembers {
						err := RemoveHead(tiktok, aTt.ID, m)
						if err != nil {
							errTrap(tiktok, "Trello RemoveMember function error in SprintGo in `sprint.go`", err)
						} else {
							if tiktok.Config.LogToSlack {
								LogToSlack("Removing "+m+" from card `"+aTt.Name+"`.", tiktok, attachments)
							}
						}
					}

					// verify if we have a {SPIKE} card or not
					spikeText := Between(aTt.ID, "{", "}")
					if strings.ToLower(spikeText) == "spike" {
						weHaveSpike = true
					} else {
						weHaveSpike = false
					}

					if points > opts.General.MaxPoints {
						// send an alert and don't move the card
						if tiktok.Config.LogToSlack {
							LogToSlack("Found card greater than "+strconv.Itoa(opts.General.MaxPoints)+" points in `Next Sprint` column. Card will *not* be moved.  Sending an alert to "+opts.General.ComplaintChannel, tiktok, attachments)
						}

						amessage := "Card #" + strconv.Itoa(aTt.IDShort) + " contains _*" + spoints + "*_ points!\n"
						amessage = amessage + "Please address it. - <" + aTt.ShortURL + "|" + aTt.Name + ">"
						attachments.Color = "#ff0000"
						attachments.Text = amessage

						Wrangler(tiktok.Config.SlackHook, "<!here> *WARNING!* High Point Card Found!", opts.General.SprintChannel, tiktok.Config.SlackEmoji, attachments)

					} else if spoints == "0" && !weHaveSpike {
						// send an alert and don't move the card if points is 0 AND its not a {SPIKE}
						if tiktok.Config.LogToSlack {
							LogToSlack("Found card with *zero* points in `Next Sprint` column. Card will *not* be moved.  Sending an alert to "+opts.General.ComplaintChannel, tiktok, attachments)
						}

						amessage := "Card <" + aTt.ShortURL + "|" + aTt.Name + "> contains _*NO*_ points!\n"
						attachments.Color = "#ff0000"
						attachments.Text = amessage

						Wrangler(tiktok.Config.SlackHook, "<!here> *WARNING!* Card with No Points!", opts.General.SprintChannel, tiktok.Config.SlackEmoji, attachments)

					} else {
						// otherwise move card
						if tiktok.Config.LogToSlack {
							attachments.Color = ""
							attachments.Text = ""
							LogToSlack("Moving card _"+aTt.Name+"_ ("+aTt.ID+") to *Ready for Work* column on `"+opts.General.TeamName+"` board, for the next sprint.", tiktok, attachments)
						}
						_ = MoveCardList(tiktok, aTt.ID, opts.General.ReadyForWork)
						newsprintcount++
					}
				} else { // card is silenced so still need to move it
					if tiktok.Config.LogToSlack {
						attachments.Color = ""
						attachments.Text = ""
						LogToSlack("Moving card _"+aTt.Name+"_ ("+aTt.ID+") to *Ready for Work* column on `"+opts.General.TeamName+"` board, for the next sprint.", tiktok, attachments)
					}
					_ = MoveCardList(tiktok, aTt.ID, opts.General.ReadyForWork)
					newsprintcount++
				}
			}
		}
	}

	// Create Retro Board for next sprint
	if retroNo {
		attachments.Color = ""
		attachments.Text = ""

		if tiktok.Config.DEBUG {
			fmt.Println("Supressing creation of Retroboard due to suppress command being given.")
		}
		if tiktok.Config.LogToSlack {
			LogToSlack("Suppressing creation of Retro Board for `"+opts.General.TeamName+" on the next sprint retro due to over-ride! command being given.", tiktok, attachments)
		}
	} else {
		boardName := "Retro: " + newSprintName
		trellout, err := CreateBoard(boardName, opts.General.TrelloOrg, tiktok)
		if err != nil {
			errTrap(tiktok, "Trello error in CreateBoard `sprint.go` for `"+opts.General.TeamName+"` board", err)
			return "Trello error in CreateBoard `sprint.go` for `" + opts.General.TeamName + "` board", err
		}
		rboardID = trellout.ID

		// Create lists on new board.  Create in reverse order you want them to display in
		err = CreateList(rboardID, "Completed", tiktok)
		err = CreateList(rboardID, "Action Items", tiktok)
		err = CreateList(rboardID, "Vent", tiktok)
		err = CreateList(rboardID, "Stop Doing", tiktok)
		err = CreateList(rboardID, "Start Doing", tiktok)
		err = CreateList(rboardID, "What Needs Improvement", tiktok)
		err = CreateList(rboardID, "What Went Well", tiktok)

		if tiktok.Config.DEBUG {
			fmt.Println("Creating Sprint Retro Board: " + boardName)
		}
		if tiktok.Config.LogToSlack {
			LogToSlack("Created next sprint Retro Board _"+boardName+"_ for `"+opts.General.TeamName+"`, for the next sprint retro.", tiktok, attachments)
		}

		// Assign new board to RETRO collections
		out := AssignCollection(rboardID, opts.General.RetroCollectionID, tiktok)

		if tiktok.Config.DEBUG {
			fmt.Println(out)
		}
		if tiktok.Config.LogToSlack {
			LogToSlack(out+" for Retro board _"+boardName+"_ for `"+opts.General.TeamName+"`", tiktok, attachments)
		}

		// Add team members to the board
		attachments.Color = ""
		attachments.Text = ""
		retroMessage = ""

		retroUsers, err := GetDBUsers(tiktok)

		for _, u := range retroUsers {
			err = AddBoardMember(tiktok, rboardID, u.Trello)
			if err != nil {
				errTrap(tiktok, "Error adding member "+u.Name+" to new Retro Board.  Trello error in AddBoardMember `sprint.go`", err)
			}

			retroMessage = retroMessage + "Member " + u.Name + " (" + u.Trello + ") \n"
		}

		if tiktok.Config.LogToSlack {
			attachments.Color = "#0000ff"
			attachments.Text = retroMessage
			LogToSlack("Following users added to new Retro board "+boardName+" ("+rboardID+")", tiktok, attachments)
		}

		// Output
		attachments.Color = "#00aaff"
		attachments.Text = "I created this sprints Retro board and its called " + boardName + "!\n https://trello.com/b/" + rboardID + "/"
		Wrangler(tiktok.Config.SlackHook, "*Notice!*", opts.General.RetroChannel, tiktok.Config.SlackEmoji, attachments)

	}

	attachments.Color = ""
	attachments.Text = ""

	// Add Demo card list to demo board if it exists
	if opts.General.DemoBoardID != "" {
		listName := "DEMO: Sprint " + newSprintName
		aTt, _ := RetrieveAll(tiktok, opts.General.DemoBoardID, "visible")
		demoBoardID := aTt.ID
		err = CreateList(demoBoardID, listName, tiktok)
		if err != nil {
			errTrap(tiktok, "Error attempting to add list called `"+listName+"` to Demo board `"+opts.General.DemoBoardID+"` in `sprint.go`", err)
		} else {
			if tiktok.Config.LogToSlack {
				LogToSlack("Adding list named `"+listName+"` to the DEMO Board "+opts.General.DemoBoardID+" ("+demoBoardID+") for cards this sprint", tiktok, attachments)
			}
		}
	} else {
		if tiktok.Config.LogToSlack {
			LogToSlack("Skipping creation of new LIST on Demo board as none is specified in the TOML file.", tiktok, attachments)
		}
	}

	sprintStartTime := time.Now().Local()
	sprintStartTime.Format("2006-01-02 15:04:05")

	// Figure out working days in sprint accounting for holidays
	if tiktok.Config.LogToSlack {
		LogToSlack("Calculating working days next sprint based on known Holidays", tiktok, attachments)
	}
	oneDay := int64(86400)
	startDate := int64(sprintStartTime.Unix())
	eD := startDate + (oneDay * int64(opts.General.SprintDuration))
	endDate := int64(eD)

	workingDays = 0
	for timestamp := startDate; timestamp < endDate; timestamp += oneDay {
		valid, holiday := IsHoliday(tiktok, time.Unix(timestamp, 0))
		if !valid {
			workingDays = workingDays + 1
		} else {
			if tiktok.Config.LogToSlack {
				LogToSlack("Holiday found `"+holiday.Name+"` skipping as a work day.", tiktok, attachments)
			}
		}
	}

	// estimate # of weekends based on sprint length (divide by 7 basically)
	totalWeekendDays := (float64(opts.General.SprintDuration) / float64(7)) * 2

	//subtract 4 days for weekends
	wDays := workingDays - int(totalWeekendDays)

	if tiktok.Config.LogToSlack {
		LogToSlack("Based on upcoming Holidays and "+strconv.Itoa(int(totalWeekendDays))+" weekend days this will make "+strconv.Itoa(wDays)+" working days this next sprint", tiktok, attachments)
	}

	// Update SQL DB with Sprint Data
	sOpts.SprintStart = sprintStartTime
	sOpts.Duration = opts.General.SprintDuration
	sOpts.RetroID = rboardID
	sOpts.SprintName = newSprintName
	sOpts.TeamID = strings.ToLower(opts.General.Sprintname)
	sOpts.WorkingDays = wDays

	err = PutDBSprint(tiktok, sOpts)
	if err != nil {
		errTrap(tiktok, "Error writing sprint data to SQL DB via func `PutDBSprint` in `sprint.go`", err)
	}

	// Re-record points for new sprint
	_, _ = GetAllPoints(tiktok, opts, sOpts)

	// Update slack with goodness
	hmessage := "*New Sprint Active* - (<https://trello.com/b/" + opts.General.BoardID + "|" + newSprintName + ">)"
	amessage := "Total cards moved from current sprint to next sprint: " + strconv.Itoa(countcards) + "\n"
	amessage = amessage + "Total cards moved to Backlog: " + strconv.Itoa(countcardsbl) + "\n"
	amessage = amessage + "Total cards in Next Sprint: " + strconv.Itoa(newsprintcount) + "\n\n"
	for _, s := range allSquads {
		if opts.General.BoardID == s.BoardID {
			amessage = amessage + "Total `" + s.Squadname + "` Points: " + strconv.Itoa(s.SquadPts) + "\n"
		}
	}
	amessage = amessage + "Total points added for this Sprint: " + strconv.Itoa(totalPoints) + "\n"

	attachments.Color = "#00ba2b"
	attachments.Text = amessage

	Wrangler(tiktok.Config.SlackHook, hmessage, opts.General.SprintChannel, tiktok.Config.SlackEmoji, attachments)

	if tiktok.Config.DEBUG {
		fmt.Println("Total Cards moved from Sprint to Sprint: " + strconv.Itoa(countcards))
		fmt.Println("Total Cards moved to Backlog: " + strconv.Itoa(countcardsbl))
		fmt.Println("Total Cards moved into new Sprint: " + strconv.Itoa(newsprintcount))
		fmt.Println("Total Points aded for this Sprint: " + strconv.Itoa(totalPoints))
	}

	return "Done Executing Sprint Setup for `" + opts.General.TeamName + "` board\n", nil
}
