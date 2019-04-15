package tiktokmod

// Manages CRON job calls to functions

import (
	"fmt"
	"strings"
	"time"

	"github.com/robfig/cron"
)

// cronFunc - function for encompassing cron functions
type cronFunc func(tiktok *TikTokConf, config string, job string, holiday bool)

func newCron(handler cronFunc, tiktok *TikTokConf, config string, job string, holiday bool) func() {
	return func() { handler(tiktok, config, job, holiday) }
}

func localLoad(tiktok *TikTokConf, teamID string) (opts Config, err error) {
	var attachments Attachment

	opts, err = LoadConf(tiktok, teamID)

	if err != nil {
		LogToSlack("I couldn't find the team config file ("+teamID+".toml) specified in Cron job!.", tiktok, attachments)
		return opts, err
	}

	return opts, err
}

// HolidayTroll - TikTok Holiday messaging Cron
func HolidayTroll(tiktok *TikTokConf, teamID string, job string, dummy bool) {
	var attachments Attachment

	opts, err := localLoad(tiktok, teamID)
	if err != nil {
		errTrap(tiktok, "CRON ISSUE: Error Loading teamID `"+teamID+"` in HolidayTroll in `cron.go`", err)
		return
	}

	// Check for Holiday
	isHoliday, holiday := IsHoliday(tiktok, time.Now())
	if isHoliday && opts.General.HolidaySupport {
		if strings.ToLower(holiday.Name) == "saas off-site" {
			Wrangler(tiktok.Config.SlackHook, "I'm at the SaaS Off-Site today so I'm not doing my regular routine. "+holiday.Message, opts.General.ComplaintChannel, tiktok.Config.SlackEmoji, attachments)
		} else {
			Wrangler(tiktok.Config.SlackHook, "I'm not working today, it's a company Holiday! "+holiday.Message, opts.General.ComplaintChannel, tiktok.Config.SlackEmoji, attachments)
		}
	}
}

// StandardCron - Execute requested cron job
func StandardCron(tiktok *TikTokConf, teamID string, job string, holiday bool) {
	var attachments Attachment
	var err error
	var returnMsg string
	var opts Config

	opts, err = localLoad(tiktok, teamID)
	if err != nil {
		errTrap(tiktok, "CRON ISSUE: Error Loading teamID `"+teamID+"` in `"+job+"` in `cron.go`", err)
		return
	}

	if holiday {
		isHoliday, holiday := IsHoliday(tiktok, time.Now())
		if isHoliday && opts.General.HolidaySupport {
			if tiktok.Config.LogToSlack {
				LogToSlack("Today is Holiday, skipping cron job `"+job+"`. ("+holiday.Name+")", tiktok, attachments)
			}

			return
		}
	}

	if tiktok.Config.LogToSlack {
		LogToSlack("Executing CRON `"+job+"` on team *"+teamID+"*", tiktok, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	switch job {
	case "troll":
		returnMsg, err = AlertRunner(opts, tiktok)
		SkippedPR(tiktok, opts)
	case "pr-summary":
		returnMsg, err = PRSummary(opts, tiktok)
	case "templatecheck":
		TemplateCard(tiktok, opts)
	case "retroaction":
		CheckActionCards(tiktok, opts, teamID)
	case "chapter-count":
		err = RecordChapters(tiktok, teamID, "backlog")
	case "count-cards":
		_, err = CountCards(opts, tiktok, teamID)
	case "record-pts":
		sOpts, err := GetDBSprint(tiktok, teamID)
		if err != nil {
			errTrap(tiktok, "CRON ISSUE: SQL error in `GetDBSprint` in `cron.go`", err)
			return
		}
		_, _ = GetAllPoints(tiktok, opts, sOpts)
	case "sprint":
		returnMsg, err = Sprint(opts, tiktok, false)
	case "pr-alert":
		returnMsg, err = StalePRcards(opts, tiktok)
	case "points":
		returnMsg = PointCleanup(opts, tiktok, teamID)
	case "archive":
		returnMsg, err = CleanDone(opts, tiktok)
	case "":
		err = CleanBackLog(opts, tiktok)
	case "backlogarchive":
		err = ArchiveBacklog(tiktok, opts)
	case "critical-bug":
		_ = CheckBugs(opts, tiktok)
	case "epic-links":
		EpicLink(tiktok, opts)
	case "cardloader":
		CardPlay(tiktok, opts, "", teamID, false)
	case "standupalert":
		SendAlert(tiktok, opts, "standup")
	case "demoalert":
		SendAlert(tiktok, opts, "demo")
	case "retroalert":
		SendAlert(tiktok, opts, "retro")
	case "wdwalert":
		SendAlert(tiktok, opts, "wdw")
	}

	if tiktok.Config.LogToSlack {
		LogToSlack("Cron job "+job+" returned message "+returnMsg, tiktok, attachments)
	}
	if err != nil {
		errTrap(tiktok, "Error returned running Cron job `"+job+"` function in cron.go for team "+teamID, err)
	}

	return
}

// CronLoad - Load or re-load all cron jobs.  Re-read the toml file
func CronLoad(tiktok *TikTokConf) (cronjobs *Cronjobs, c *cron.Cron, err error) {
	var attachments Attachment

	c = cron.New()

	cronjobs, err = LoadCronFile()
	if err != nil {
		if tiktok.Config.LogToSlack {
			var attachments Attachment
			LogToSlack("*WARNING!* Can not find a valid `cron.toml` file to load!! Cron's are not running!", tiktok, attachments)
		}
		fmt.Println(err)
		return cronjobs, c, err
	}

	c.Stop()

	cMessage := "```"
	for _, j := range cronjobs.Cronjob {

		cMessage = cMessage + j.Action + " @ " + j.Timing + " for board " + j.Config + "\n"

		switch j.Action {
		case "holidays":
			c.AddFunc(j.Timing, newCron(HolidayTroll, tiktok, j.Config, "", true))
		case "standupalert":
			c.AddFunc(j.Timing, newCron(StandardCron, tiktok, j.Config, "standupalert", true))
		case "demoalert":
			c.AddFunc(j.Timing, newCron(StandardCron, tiktok, j.Config, "demoalert", true))
		case "retroalert":
			c.AddFunc(j.Timing, newCron(StandardCron, tiktok, j.Config, "retroalert", true))
		case "wdwalert":
			c.AddFunc(j.Timing, newCron(StandardCron, tiktok, j.Config, "wdwalert", true))
		case "pr-alert":
			c.AddFunc(j.Timing, newCron(StandardCron, tiktok, j.Config, "pr-alert", true))
		case "troll":
			c.AddFunc(j.Timing, newCron(StandardCron, tiktok, j.Config, "troll", true))
		case "sprint":
			c.AddFunc(j.Timing, newCron(StandardCron, tiktok, j.Config, "sprint", false))
		case "points":
			c.AddFunc(j.Timing, newCron(StandardCron, tiktok, j.Config, "points", true))
		case "archive":
			c.AddFunc(j.Timing, newCron(StandardCron, tiktok, j.Config, "archive", false))
		case "backlogarchive":
			c.AddFunc(j.Timing, newCron(StandardCron, tiktok, j.Config, "backlogarchive", false))
		case "clean-backlog":
			c.AddFunc(j.Timing, newCron(StandardCron, tiktok, j.Config, "backlogarchive", false))
		case "pr-summary":
			c.AddFunc(j.Timing, newCron(StandardCron, tiktok, j.Config, "pr-summary", true))
		case "record-pts":
			c.AddFunc(j.Timing, newCron(StandardCron, tiktok, j.Config, "record-pts", false))
		case "count-cards":
			c.AddFunc(j.Timing, newCron(StandardCron, tiktok, j.Config, "count-cards", false))
		case "epic-links":
			c.AddFunc(j.Timing, newCron(StandardCron, tiktok, j.Config, "epic-links", true))
		case "chapter-count":
			c.AddFunc(j.Timing, newCron(StandardCron, tiktok, j.Config, "chapter-count", false))
		case "critical-bug":
			c.AddFunc(j.Timing, newCron(StandardCron, tiktok, j.Config, "critical-bug", true))
		case "cardloader":
			c.AddFunc(j.Timing, newCron(StandardCron, tiktok, j.Config, "cardloader", false))
		case "retroaction":
			c.AddFunc(j.Timing, newCron(StandardCron, tiktok, j.Config, "retroaction", true))
		case "templatecheck":
			c.AddFunc(j.Timing, newCron(StandardCron, tiktok, j.Config, "templatecheck", false))
		default:
			if tiktok.Config.LogToSlack {
				LogToSlack("Warning INVALID Cron Load action called `"+j.Action+"` for Cron entry:  ```"+j.Timing+"  "+j.Config+"```", tiktok, attachments)
			}
			if tiktok.Config.DEBUG {
				fmt.Println("Warning INVALID Cron Load action called `" + j.Action + "` for Cron entry:  ```" + j.Timing + "  " + j.Config + "```")
			}
		}
	}
	cMessage = cMessage + "```"

	attachments.Text = cMessage
	attachments.Color = "#0000FF"
	if tiktok.Config.LogToSlack {
		LogToSlack("Loading Cron Jobs:", tiktok, attachments)
	}

	c.Start()

	return cronjobs, c, nil
}
