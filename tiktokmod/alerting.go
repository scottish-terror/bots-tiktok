package tiktokmod

import (
	"encoding/json"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type lists struct {
	channelID   string
	channelName string
}

// AlertRunner - Run the alerts in Planning / Next Sprint / Ready for Work / Working
func AlertRunner(opts Config, tiktok *TikTokConf) (string, error) {

	var attachments Attachment
	var messageAlertOut string
	var rHmessage string
	var mHmessage string
	var tMessage string
	var temp string
	var hush bool
	var weHaveSpike bool

	allTheThings, err := RetrieveAll(tiktok, opts.General.BoardID, "visible")
	if err != nil {
		errTrap(tiktok, "Error retrieving all cards on board "+opts.General.BoardID+" in `alerting.go` func `AlertRunner`", err)
		return "", err
	}

	if tiktok.Config.LogToSlack {
		attachments.Color = ""
		attachments.Text = ""
		LogToSlack("I'm trolling cards in the `"+opts.General.TeamName+"` board for zero points or points greater than "+strconv.Itoa(opts.General.MaxPoints)+" as well as member checking and spike checking.", tiktok, attachments)
	}

	for _, aTt := range allTheThings.Cards {
		if aTt.IDList == opts.General.NextsprintID || aTt.IDList == opts.General.ReadyForWork || aTt.IDList == opts.General.Working {

			hush = false
			for _, l := range aTt.Labels {
				if l.ID == opts.General.SilenceCardLabel {
					hush = true
				}
			}

			if !hush {
				// verify if we have a {SPIKE} card or not
				spikeText := Between(aTt.Name, "{", "}")
				if strings.ToLower(spikeText) == "spike" {
					weHaveSpike = true
				} else {
					weHaveSpike = false
				}

				points := GetCardPoints(opts, tiktok, aTt.ID)
				spoints := strconv.Itoa(points)

				if points > opts.General.MaxPoints {
					messageAlertOut = messageAlertOut + "<" + aTt.ShortURL + "|" + aTt.Name + "> contains *" + spoints + "* points!\n"
				}

				if spoints == "0" && !weHaveSpike {
					messageAlertOut = messageAlertOut + "<" + aTt.ShortURL + "|" + aTt.Name + "> contains *ZERO* points!\n"
				}

				// Check if card should NOT have an owner head on it and remove
				if aTt.IDList == opts.General.ReadyForWork || aTt.IDList == opts.General.NextsprintID {

					if len(aTt.IDMembers) > 0 {

						for _, head := range aTt.IDMembers {
							err := RemoveHead(tiktok, aTt.ID, head)
							if err != nil {
								errTrap(tiktok, "Error attempting to remove head from card <"+aTt.ShortURL+"|"+aTt.Name+"> in `AlertRunner` in `alerting.go`", err)
							}
						}

						rHmessage = rHmessage + "<" + aTt.ShortURL + "|" + aTt.Name + ">\n"
					}
				}

				// Check card should have an owner head on it and alert if not
				if aTt.IDList == opts.General.Working {
					if len(aTt.IDMembers) == 0 {
						mHmessage = mHmessage + "<" + aTt.ShortURL + "|" + aTt.Name + ">\n"
					}
				}
			}
		}
	}

	if rHmessage != "" {
		attachments.Color = "#ff0000"
		attachments.Text = "These cards should not be assigned yet!\n" + rHmessage
		Wrangler(tiktok.Config.SlackHook, "<!here> NOTICE!  I have *removed* people from these cards", opts.General.ComplaintChannel, tiktok.Config.SlackEmoji, attachments)
	}

	if mHmessage != "" {
		attachments.Color = "#ff0000"
		attachments.Text = "I'm sad! These cards are in the working column but have nobody assigned to them!\n" + mHmessage
		Wrangler(tiktok.Config.SlackHook, "<!here> Warning Un-Assigned Work!!", opts.General.ComplaintChannel, tiktok.Config.SlackEmoji, attachments)
	}

	if messageAlertOut != "" {
		attachments.Color = "#ff0000"
		attachments.Text = "These cards have too many or not enough points!\n" + messageAlertOut
		Wrangler(tiktok.Config.SlackHook, "<!here> Warning cards with Point issues!!", opts.General.ComplaintChannel, tiktok.Config.SlackEmoji, attachments)
	}

	temp, _ = CheckThemes(tiktok, opts, opts.General.Upcoming)
	tMessage = tMessage + temp
	temp, _ = CheckThemes(tiktok, opts, opts.General.Scoped)
	tMessage = tMessage + temp
	temp, _ = CheckThemes(tiktok, opts, opts.General.ReadyForWork)
	tMessage = tMessage + temp

	if tMessage != "" {
		attachments.Color = "#ff0000"
		attachments.Text = tMessage
		Wrangler(tiktok.Config.SlackHook, "*WARNING*! The following cards do *not* have appropriate Theme Labels on them: ", opts.General.ComplaintChannel, tiktok.Config.SlackEmoji, attachments)
	}

	return "", nil
}

// GetCardPoints - Get the points on a card from the power-up
func GetCardPoints(opts Config, tiktok *TikTokConf, cardID string) (points int) {
	pluginCard, _ := GetPowerUpField(cardID, tiktok)

	for _, p := range pluginCard {

		if p.IDPlugin == tiktok.Config.PointsPowerUpID {

			var plugins PointsHistory

			pluginJSON := []byte(p.Value)
			json.Unmarshal(pluginJSON, &plugins)
			points = plugins.Points
		}
	}

	return points
}

// StalePRcards - Check for cards that are aged out in the PR column
func StalePRcards(opts Config, tiktok *TikTokConf) (message string, err error) {

	var attachments Attachment
	var smessage string
	var uMessage string
	var tMessage string
	var prFound bool

	LogToSlack("I'm trolling the PR Column cards in the `"+opts.General.TeamName+"` board.", tiktok, attachments)

	allTheThings, err := RetrieveAll(tiktok, opts.General.BoardID, "visible")
	if err != nil {
		errTrap(tiktok, "Error retrieving all cards from func `RetrieveAll` in `StalePRCards` in `alerting.go` with board "+opts.General.TeamName, err)
	}

	for _, aTt := range allTheThings.Cards {

		if aTt.IDList == opts.General.ReadyForReview {
			cardAction, err := GetCardAction(tiktok, aTt.ID, 1)
			if err != nil {
				errTrap(tiktok, "Error from `GetCardAction` in `StalePRCards` in `alerting.go`", err)
			}

			for _, actions := range cardAction {
				createdAt := actions.Date
				expiresAt := time.Now()
				diff := expiresAt.Sub(createdAt)
				staleTimer := time.Duration(opts.General.StaleTime) * time.Hour

				// compenstate for weekends
				if opts.General.IgnoreWeekends {
					if int(time.Now().Weekday()) == 1 {
						diff = diff - time.Duration(48)*time.Hour
					}
				}

				// compenstate if yesterday was a holiday
				isHoliday, _ := IsHoliday(tiktok, time.Now().AddDate(0, 0, -1))
				if isHoliday {
					diff = diff - time.Duration(24)*time.Hour
				}

				if tiktok.Config.LogToSlack {
					LogToSlack("Time in list for card <"+aTt.ShortURL+"|"+aTt.Name+"> is "+diff.String(), tiktok, attachments)
				}

				if diff > staleTimer {
					// retrieve github PR from trello attachments if it exists
					if aTt.Badges.Attachments > 0 {
						attached, _ := GetAttachments(tiktok, aTt.ID)
						prFound = false
						for _, a := range attached {
							if !a.IsUpload && strings.Contains(a.URL, "github.com") && strings.Contains(a.URL, "/pull/") {
								prFound = true
								if tiktok.Config.LogToSlack {
									LogToSlack("<"+aTt.ShortURL+"|"+aTt.Name+"> has PR attached: <"+a.URL+"|"+a.Name+">", tiktok, attachments)
								}
								// get PR out of URL
								u, err := url.Parse(a.URL)
								if err != nil {
									errTrap(tiktok, "Error parsing Github URL in Trello card attachment in `StalePRCards` in `alerting.go` - URL: "+a.URL, err)
								}
								splitPath := strings.Split(u.Path, "/")
								if len(splitPath) > 0 {
									locale := len(splitPath) - 1
									prNum, _ := strconv.Atoi(splitPath[locale])
									locale = len(splitPath) - 3
									repoName := splitPath[locale]

									prDetail, err := GitPR(tiktok, repoName, prNum, tiktok.Config.GithubOrgName)
									if err == nil {
										// look in github to see if PR is closed/merged
										if *prDetail.Merged {
											loc, _ := time.LoadLocation("America/Los_Angeles")
											prMergeTime := *prDetail.MergedAt
											lastUpdate := prMergeTime.In(loc).Format("2006-01-02 15:04:05")
											uMessage = "*PLEASE NOTE* : The Github Pull Request for this card was merged on `" + lastUpdate + "`, does this card need to be closed in Trello? <" + aTt.URL + "|" + aTt.Name + ">\n"
											tMessage = ""
											if len(aTt.IDMembers) > 0 {
												for _, u := range aTt.IDMembers {
													user, err := GetUser(tiktok, "trello", u)
													if err == nil {
														if user.SlackID != "" {
															tMessage = tMessage + "@" + user.SlackID + " "
														}
													}
												}
												uMessage = tMessage + " The Github Pull Request for this card was merged on `" + lastUpdate + "`, does this card need to be closed in Trello? <" + aTt.URL + "|" + aTt.Name + ">\n"

												if tiktok.Config.LogToSlack {
													LogToSlack("PR <"+*prDetail.HTMLURL+"|"+*prDetail.Title+"> is merged but card still open, alerting owners ("+tMessage+") and channel", tiktok, attachments)
												}
											}
											Wrangler(tiktok.Config.SlackHook, uMessage, opts.General.ComplaintChannel, tiktok.Config.SlackEmoji, attachments)
										} else {
											// look in github to see if PR has been commented on in past 24 hours
											upAt := *prDetail.UpdatedAt
											expiresAt := time.Now()
											diff := expiresAt.Sub(upAt)

											// compenstate for weekends
											if opts.General.IgnoreWeekends {
												if int(time.Now().Weekday()) == 1 {
													diff = diff - time.Duration(48)*time.Hour
												}
											}

											// compenstate if yesterday was a holiday
											isHoliday, _ := IsHoliday(tiktok, time.Now().AddDate(0, 0, -1))
											if isHoliday {
												diff = diff - time.Duration(24)*time.Hour
											}

											if tiktok.Config.LogToSlack {
												LogToSlack("PR <"+*prDetail.HTMLURL+"|"+*prDetail.Title+"> was last modifed/updated "+diff.String()+" ago", tiktok, attachments)
											}

											if diff > staleTimer {
												smessage = smessage + "<" + aTt.ShortURL + "|" + aTt.Name + ">\n"
											} else {
												if tiktok.Config.LogToSlack {
													LogToSlack("<"+aTt.URL+"|"+aTt.Name+"> is lagging in trello but has current updates in Github, no alerting.  PR Is here <"+*prDetail.HTMLURL+"|"+*prDetail.Title+">", tiktok, attachments)
												}
											}
										}
									}
								}
							}
						}
						if !prFound {
							if tiktok.Config.LogToSlack {
								LogToSlack("No github PR's found attached to <"+aTt.ShortURL+"|"+aTt.Name+">", tiktok, attachments)
							}
							smessage = smessage + "<" + aTt.ShortURL + "|" + aTt.Name + ">\n"
						}
					} else {
						// no PR attached so assuming the worst
						if tiktok.Config.LogToSlack {
							LogToSlack("No PR attached to <"+aTt.ShortURL+"|"+aTt.Name+"> and its over time so sending warning message.", tiktok, attachments)
						}
						smessage = smessage + "<" + aTt.ShortURL + "|" + aTt.Name + ">\n"
					}
				} else {
					if tiktok.Config.LogToSlack {
						LogToSlack("Ignoring <"+aTt.ShortURL+"|"+aTt.Name+">.", tiktok, attachments)
					}
				}

			}

		}
	}

	if smessage != "" {
		attachments.Color = "#ff0000"
		attachments.Text = "These are " + strconv.Itoa(opts.General.StaleTime) + " hours or older\n" + smessage
		Wrangler(tiktok.Config.SlackHook, "<!here> WARNING!! Lagging PR Card(s)!!", opts.General.ComplaintChannel, tiktok.Config.SlackEmoji, attachments)
	}

	return "", nil
}

// SkippedPR - Alert if cards have skipped PR column
func SkippedPR(tiktok *TikTokConf, opts Config) {
	var message string
	var attachments Attachment
	var commentMsg string
	var checkThisCard bool

	users, err := GetDBUsers(tiktok)
	if err != nil {
		errTrap(tiktok, "Error getting user data from `GetDBUsers` in `SkippedPR` in `alerting.go`", err)
		return
	}

	allTheThings, err := RetrieveAll(tiktok, opts.General.BoardID, "visible")
	if err != nil {
		errTrap(tiktok, "Trello error in RetrieveAll in `SkippedPR` in `alerting.go` for `"+opts.General.TeamName+"` board", err)
		return
	}

	message = ""
	for _, aTt := range allTheThings.Cards {
		if !aTt.Closed {
			if aTt.IDList == opts.General.Done {

				checkThisCard = true

				// check if card has already been commented on by TikTok and skip it if so
				cardComments, err := GetCardComments(aTt.ID, tiktok)
				if err != nil {
					errTrap(tiktok, "Error on return from `GetCardComments` in `SkippedPR` in `Trello.go`", err)
					return
				}
				for _, c := range cardComments {
					if c.MemberCreator.Username == tiktok.Config.BotTrelloID {
						if strings.Contains(c.Data.Text, tiktok.Config.BotName+" PR Message:") {
							checkThisCard = false
						}
					}
				}

				if checkThisCard {
					cardListHistory := GetCardListHistory(aTt.ID, tiktok)

					for _, h := range cardListHistory {
						if h.Data.ListAfter.ID == opts.General.Done {
							if h.Data.ListBefore.ID != opts.General.ReadyForReview {
								message = message + "<https://trello.com/c/" + aTt.ID + "|" + aTt.Name + ">\n"
								commentMsg = tiktok.Config.BotName + " PR Message: Couldn't find a card owner on this card that has skipped the Review process so I sent a general alert to the " + opts.General.ComplaintChannel + " slack channel about it."

								for _, u := range users {
									if len(aTt.IDMembers) > 0 {
										_, _, userName := GetMemberInfo(aTt.IDMembers[0], tiktok)
										if userName == u.Trello {
											commentMsg = tiktok.Config.BotName + " PR Message: Sent warning to @" + u.Trello + " that this card skipped the Review process and they should put an update in it with an explanation."
											Wrangler(tiktok.Config.SlackHook, "*Warning!* This card with your face on it, appears to have skipped the `Review` column, please resolve this by adding notes as to why this happened. Even spikes should be reviewed! Thank you!\n<https://trello.com/c/"+aTt.ID+"|"+aTt.Name+">", "@"+u.SlackID, tiktok.Config.SlackEmoji, attachments)
										}
									}
								}

								// put comment on the card so it gets ignored next round
								err = CommentCard(aTt.ID, commentMsg, tiktok)
								if err != nil {
									errTrap(tiktok, "Error attempting to comment on card "+aTt.ID+" bailing out of SkippedPR routine", err)
									return
								}

								break
							}
						}
					}
				}
			}
		}
	}

	if message != "" {
		attachments.Color = "#ff0000"
		attachments.Text = message
		headerMsg := "*Warning* The following cards appear to have skipped the review column in trello.  If you are an owner of one of these cards I will slack you directly about putting a note in it regarding why it skipped `Ready for Review`!\nPlease review these!"
		Wrangler(tiktok.Config.SlackHook, headerMsg, opts.General.ComplaintChannel, tiktok.Config.SlackEmoji, attachments)
	}
}

// CheckBugs - Check for bugs and alert on them
func CheckBugs(opts Config, tiktok *TikTokConf) (critBugNum int) {
	var message string
	var amessage string
	var criticalID string
	var attachments Attachment

	bugLabels, err := GetBugID(tiktok, opts.General.BoardID)
	if err != nil {
		errTrap(tiktok, "Error getting user data from `GetDBUsers` in `SkippedPR` in `alerting.go`", err)
		return 0
	}
	for _, bugs := range bugLabels {
		if strings.ToLower(bugs.BugLevel) == "critical" {
			criticalID = bugs.LabelID
		}
	}

	allTheThings, err := RetrieveAll(tiktok, opts.General.BoardID, "visible")
	if err != nil {
		errTrap(tiktok, "Trello error in RetrieveAll in `CheckBugs` in `alerting.go` for `"+opts.General.TeamName+"` board", err)
		return 0
	}

	critBugNum = 0
	for _, aTt := range allTheThings.Cards {
		if !aTt.Closed {
			// Alert on Critical Bugs
			if aTt.IDList == opts.General.BacklogID || aTt.IDList == opts.General.Upcoming || aTt.IDList == opts.General.Scoped {
				for _, labels := range aTt.Labels {
					if labels.ID == criticalID {
						message = message + "<https://trello.com/c/" + aTt.ID + "|" + aTt.Name + ">\n"
						critBugNum = critBugNum + 1
					}
				}
			}
		}
	}

	if critBugNum > 0 {
		if critBugNum == 1 {
			amessage = "<!here> *CRITICAL* Bug Opened!"
		}
		if critBugNum > 1 {
			amessage = "<!here> *CRITICAL* Bugs Opened!\n" + strconv.Itoa(critBugNum) + " new critical bugs.\n"
		}

		attachments.Color = "#FF0000"
		attachments.Text = message
		Wrangler(tiktok.Config.SlackHook, amessage, opts.General.ComplaintChannel, tiktok.Config.SlackEmoji, attachments)
	} else {
		if tiktok.Config.LogToSlack {
			LogToSlack("No Critical Bugs Found", tiktok, attachments)
		}
	}

	return critBugNum

}

// SendAlert - send a slack alert message about a sprint meeting reminder (standup/demo/retro/wdw)
func SendAlert(tiktok *TikTokConf, opts Config, alertType string) {
	var attachments Attachment
	var meetType string
	var location string
	var channel string
	var message string

	// Check for Holiday
	isHoliday, holiday := IsHoliday(tiktok, time.Now())
	if isHoliday && opts.General.HolidaySupport {
		if tiktok.Config.LogToSlack {
			LogToSlack("Today is Holiday, skipping "+alertType+" slack alert. ("+holiday.Name+")", tiktok, attachments)
		}

		return
	}

	switch alertType {
	case "standup":
		meetType = "Stand-Up"
		channel = opts.General.StandupAlertChannel
		location = opts.General.StandupLink
	case "retro":
		meetType = "Retro"
		channel = opts.General.RetroAlertChannel
		location = opts.General.RetroAlertLink
	case "demo":
		meetType = "Demos"
		channel = opts.General.DemoAlertChannel
		location = opts.General.DemoAlertLink
	case "wdw":
		meetType = "WDW"
		channel = opts.General.WDWAlertChannel
		location = opts.General.WDWAlertLink
	case "sdlc":
		meetType = "White Fences & SDLC review"
		channel = opts.General.WDWAlertChannel
		location = ""
	}

	preMsg := []string{
		"Hey everybody its time for " + meetType,
		"Oh ya its " + meetType + " time!",
		"Tick tock, join us for " + meetType + "!",
		"Hey there, its that time again! Let's do " + meetType,
	}

	rand.Seed(time.Now().Unix())
	message = "<!here> " + preMsg[rand.Intn(len(preMsg))] + " - " + location

	Wrangler(tiktok.Config.SlackHook, message, channel, tiktok.Config.SlackEmoji, attachments)

	return
}
