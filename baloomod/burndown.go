package baloomod

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

//GetAllPoints - GetAll Points in a sprint
func GetAllPoints(wOpts *WallConf, opts Config, sOpts SprintData) (message string, valid bool) {

	var attachments Attachment
	var plugins PointsHistory
	var sprintName string
	var numCards int

	// Trello args maps for custom fields
	var m map[string]string
	m = make(map[string]string)
	m["fields"] = "name"
	m["customFieldItems"] = "true"

	rfwpts := 0
	wkgpts := 0
	rfrpts := 0
	dnepts := 0

	allTheThings, err := RetrieveAll(wOpts, opts.General.BoardID, "visible")
	if err == nil {
		for _, aTt := range allTheThings.Cards {
			if !aTt.Closed {
				// wip sprint cards
				if aTt.IDList == opts.General.ReadyForWork || aTt.IDList == opts.General.Working || aTt.IDList == opts.General.ReadyForReview || aTt.IDList == opts.General.Done {

					sprintName = ""

					for _, cc := range aTt.CustomFieldItems {
						if cc.IDCustomField == opts.General.CfsprintID {
							sprintName = cc.Value.Text
						}
					}

					pluginCard, _ := GetPowerUpField(aTt.ID, wOpts)

					for _, pl := range pluginCard {
						// zero out this struct field, as sometimes its non-existent in the json payload
						plugins.Points = 0

						if pl.IDPlugin == wOpts.Walle.PointsPowerUpID {
							pluginJSON := []byte(pl.Value)
							json.Unmarshal(pluginJSON, &plugins)

							switch {

							case aTt.IDList == opts.General.ReadyForWork:
								rfwpts = rfwpts + plugins.Points
								numCards = numCards + 1
							case aTt.IDList == opts.General.Working:
								wkgpts = wkgpts + plugins.Points
								numCards = numCards + 1
							case aTt.IDList == opts.General.ReadyForReview:
								rfrpts = rfrpts + plugins.Points
								numCards = numCards + 1
							case aTt.IDList == opts.General.Done:
								if sprintName != "" {
									if sOpts.SprintName == sprintName {
										if wOpts.Walle.LogToSlack && wOpts.Walle.DEBUG {
											LogToSlack("Done Card w/ SprintName `"+sprintName+"` found, adding "+strconv.Itoa(plugins.Points)+" points. Card: "+aTt.ShortURL, wOpts, attachments)
										}
										dnepts = dnepts + plugins.Points
										numCards = numCards + 1
									}
								} else {
									if wOpts.Walle.LogToSlack {
										LogToSlack("Done Card w/ missing Sprint Name (`"+aTt.Name+"`) found. Card: "+aTt.ShortURL, wOpts, attachments)
									}
									value, cardListTime := GetTimePutList(opts.General.Done, aTt.ID, opts, wOpts)
									if value {
										format := "2006-01-02 15:04:05"
										fmtTime := cardListTime.Format("2006-01-02 15:04:05")
										cardTime, _ := time.Parse(format, fmtTime)
										if cardTime.After(sOpts.SprintStart) {
											dnepts = dnepts + plugins.Points
											if wOpts.Walle.LogToSlack && wOpts.Walle.DEBUG {
												LogToSlack("Card (`"+aTt.Name+"`) also in current sprint time frame so adding "+strconv.Itoa(plugins.Points)+" points", wOpts, attachments)
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

		totalPoints := rfwpts + wkgpts + rfrpts + dnepts

		if totalPoints > 0 {
			db, status, _ := ConnectDB(wOpts, "walle")
			if status {

				today := time.Now().Local()
				today.Format("2006-01-02 15:04:05")

				stmt, _ := db.Prepare("INSERT walle_burndown SET pointdate=?,team=?,totalpoints=?,rfwpts=?,wkgpts=?,uatpts=?,dnepts=?,numcards=?")

				_, err := stmt.Exec(today, sOpts.TeamID, totalPoints, rfwpts, wkgpts, rfrpts, dnepts, numCards)

				if err != nil {
					errTrap(wOpts, "SQL Error in walle_burndown table insert:", err)
				}

			}
			if wOpts.Walle.DEBUG {
				fmt.Println("Failed connection, bailing out...")
			}
		} else {
			if wOpts.Walle.DEBUG {
				fmt.Print("Trying to add points for " + opts.General.TeamName + " sprint and Zero Points were found, somethings awry!")
			}
			if wOpts.Walle.LogToSlack {
				LogToSlack("Trying to add points for "+opts.General.TeamName+" sprint and Zero Points were found, somethings awry!", wOpts, attachments)
			}

			return "Invalid points.", false
		}

		avgPtsCard := float64(totalPoints) / float64(numCards)
		hmessage := "Recording today's sprint points for *" + opts.General.TeamName + "* off this <https://trello.com/b/" + opts.General.BoardID + "|Trello Board>"
		message = "Points in Ready For Work: " + strconv.Itoa(rfwpts) + "\n"
		message = message + "Points in Working: " + strconv.Itoa(wkgpts) + "\n"
		message = message + "Points in PR: " + strconv.Itoa(rfrpts) + "\n"
		message = message + "Points in Done: " + strconv.Itoa(dnepts) + "\n"
		message = message + "Total Points in Sprint: " + strconv.Itoa(totalPoints) + "\n"
		message = message + "Total Cards in Sprint: " + strconv.Itoa(numCards) + "\n"
		message = message + "Avg Points Per Card: " + strconv.FormatFloat(avgPtsCard, 'f', 2, 64)

		if wOpts.Walle.LogToSlack {
			attachments.Color = "#0000ff"
			attachments.Text = message
			LogToSlack(hmessage, wOpts, attachments)
		}

	} else {
		errTrap(wOpts, "Error attempting to get all trello cards (nested call) in burndown.go for board "+sOpts.TeamID, err)
		return "Failed get all cards", false
	}

	return message, true
}

// SprintSquadPoints - Determine squad points used on a specific sprint by sprint name
func SprintSquadPoints(wOpts *WallConf, opts Config, sprintName string) (totalpoints Squads, nonPoints int, err error) {
	var checker bool
	var points int

	// Initialize non-points so its not nil
	nonPoints = 0

	// Trello args maps for custom fields
	var m map[string]string
	m = make(map[string]string)
	m["fields"] = "name"
	m["customFieldItems"] = "true"

	// Load Squad Information
	totalpoints, err = GetDBSquads(wOpts, opts.General.BoardID)
	if err != nil {
		errTrap(wOpts, "Failed DB Call to get squad information in `burndown.go` func `SprintSquadPoints`", err)
		return totalpoints, nonPoints, err
	}

	allTheThings, err := RetrieveAll(wOpts, opts.General.BoardID, "visible")
	if err != nil {
		errTrap(wOpts, "Trello error in SprintSquadPoints `burndown.go` for `"+opts.General.TeamName+"` board", err)
		return
	}

	for _, aTt := range allTheThings.Cards {
		if aTt.IDList == opts.General.ReadyForWork || aTt.IDList == opts.General.Working || aTt.IDList == opts.General.ReadyForReview || aTt.IDList == opts.General.Done {
			for _, cusval := range aTt.CustomFieldItems {
				if cusval.IDCustomField == opts.General.CfsprintID {
					if cusval.Value.Text == sprintName {
						points = 0

						pluginCard, _ := GetPowerUpField(aTt.ID, wOpts)
						for _, p := range pluginCard {

							if p.IDPlugin == wOpts.Walle.PointsPowerUpID {

								var plugins PointsHistory

								pluginJSON := []byte(p.Value)
								json.Unmarshal(pluginJSON, &plugins)
								points = plugins.Points
							}
						}

						checker = false
						for _, labels := range aTt.Labels {

							for s, squad := range totalpoints {
								if opts.General.BoardID == squad.BoardID && squad.LabelID == labels.ID {
									tPts := squad.SquadPts
									totalpoints[s].SquadPts = tPts + points
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
		}
	}

	return totalpoints, nonPoints, nil
}

//ChapterCount - Card count by chapter on given list
func ChapterCount(wOpts *WallConf, opts Config, listID string) (allChapter Chapters, totalCards int, err error) {

	allChapter, err = GetDBChapters(wOpts, opts.General.BoardID)
	if err != nil {
		errTrap(wOpts, "Failed DB Call to get chapter information in `burndown.go` func `ChapterCount`", err)
		return allChapter, 0, err
	}

	allTheThings, err := RetrieveAll(wOpts, opts.General.BoardID, "visible")
	if err != nil {
		errTrap(wOpts, "Trello error in `ChapterCount` in `burndown.go` for `"+opts.General.TeamName+"` board", err)
		return allChapter, 0, err
	}

	for _, aTt := range allTheThings.Cards {
		if !aTt.Closed {
			if aTt.IDList == listID {
				totalCards = totalCards + 1

				for _, labels := range aTt.Labels {

					for s, chapter := range allChapter {
						if opts.General.BoardID == chapter.BoardID && chapter.LabelID == labels.ID {
							allChapter[s].ChapterCount = allChapter[s].ChapterCount + 1
						}
					}
				}
			}
		}
	}

	return allChapter, totalCards, nil
}

//ChapterPoint - Point count by chapter on given list
func ChapterPoint(wOpts *WallConf, opts Config, listID string) (allChapter Chapters, noChapter int, err error) {

	var points int
	var checker bool

	allChapter, err = GetDBChapters(wOpts, opts.General.BoardID)
	if err != nil {
		errTrap(wOpts, "Failed DB Call to get chapter information in `burndown.go` func `ChapterCount`", err)
		return allChapter, 0, err
	}

	allTheThings, err := RetrieveAll(wOpts, opts.General.BoardID, "visible")
	if err != nil {
		errTrap(wOpts, "Trello error in `ChapterCount` in `burndown.go` for `"+opts.General.TeamName+"` board", err)
		return allChapter, 0, err
	}

	for _, aTt := range allTheThings.Cards {
		if !aTt.Closed {
			if aTt.IDList == listID {
				points = 0

				pluginCard, _ := GetPowerUpField(aTt.ID, wOpts)
				for _, p := range pluginCard {

					if p.IDPlugin == wOpts.Walle.PointsPowerUpID {

						var plugins PointsHistory

						pluginJSON := []byte(p.Value)
						json.Unmarshal(pluginJSON, &plugins)
						points = plugins.Points
					}
				}

				checker = false
				for _, labels := range aTt.Labels {

					for s, chapter := range allChapter {
						if opts.General.BoardID == chapter.BoardID && chapter.LabelID == labels.ID {
							tPts := chapter.ChapterPoints
							allChapter[s].ChapterPoints = tPts + points
							checker = true
						}
					}
				}
				if !checker {
					noChapter = noChapter + points
				}

			}
		}
	}

	return allChapter, noChapter, nil
}
