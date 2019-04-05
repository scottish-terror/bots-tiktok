package tiktokmod

// Manages CRON job calls to functions

import (
	"fmt"
	"strings"
	"time"

	"github.com/robfig/cron"
)

// CronFunc - function for encompassing cron functions
type CronFunc func(tiktok *TikTokConf, config string)

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
func HolidayTroll(tiktok *TikTokConf, teamID string) {
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

// RecordThemeCount - Record theme card count for a specific board from a cronjob
func RecordThemeCount(tiktok *TikTokConf, teamID string) {
	var attachments Attachment

	if tiktok.Config.LogToSlack {
		LogToSlack("Executing CRON `RecordThemeCount` on team *"+teamID+"*", tiktok, attachments)
	}

	opts, err := LoadConf(tiktok, teamID)
	if err != nil {
		errTrap(tiktok, "CRON ISSUE: Error Loading teamID `"+teamID+"` in RecordThemeCount in `cron.go`", err)
		return
	}

	_, err = CountCards(opts, tiktok, teamID)
	if err != nil {
		errTrap(tiktok, "Error returned running Cron job `CountCards` function in cron.go for team "+teamID, err)
	}
}

// RecordPointCron - Record points for a specific board from a cronjob
func RecordPointCron(tiktok *TikTokConf, teamID string) {

	if tiktok.Config.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `RecordPointCron` on team *"+teamID+"*", tiktok, attachments)
	}

	opts, err := LoadConf(tiktok, teamID)
	if err != nil {
		errTrap(tiktok, "CRON ISSUE: Error Loading teamID `"+teamID+"` in RecordPointCron in `cron.go`", err)
		return
	}
	sOpts, err := GetDBSprint(tiktok, teamID)
	if err != nil {
		errTrap(tiktok, "CRON ISSUE: SQL error in `GetDBSprint` in `cron.go`", err)
		return
	}
	_, _ = GetAllPoints(tiktok, opts, sOpts)

}

// SprintCron - Execute Sprint from a Cronjob
func SprintCron(tiktok *TikTokConf, teamID string) {

	if tiktok.Config.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `SprintCron` on team *"+teamID+"*", tiktok, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	opts, err := localLoad(tiktok, teamID)
	if err != nil {
		errTrap(tiktok, "CRON ISSUE: Error Loading teamID `"+teamID+"` in SprintCron in `cron.go`", err)
		return
	}

	returnMsg, err := Sprint(opts, tiktok, false)
	if tiktok.Config.DEBUG {
		fmt.Println(returnMsg)
	}

	return
}

// PrCron - Execute PR Scan from a Cronjob
func PrCron(tiktok *TikTokConf, teamID string) {
	var attachments Attachment

	opts, err := localLoad(tiktok, teamID)
	if err != nil {
		errTrap(tiktok, "CRON ISSUE: Error Loading teamID `"+teamID+"` in PrCron in `cron.go`", err)
		return
	}

	// Check for Holiday
	isHoliday, holiday := IsHoliday(tiktok, time.Now())
	if isHoliday && opts.General.HolidaySupport {
		if tiktok.Config.LogToSlack {
			LogToSlack("Today is Holiday, skipping PR Scan. ("+holiday.Name+")", tiktok, attachments)
		}

		return
	}

	if tiktok.Config.LogToSlack {
		LogToSlack("Executing CRON `PrCron` on team *"+teamID+"*", tiktok, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	returnMsg, err := StalePRcards(opts, tiktok)
	if tiktok.Config.DEBUG {
		fmt.Println(returnMsg)
	}

	return
}

// PointsCron - Execute Points Sync from a Cronjob
func PointsCron(tiktok *TikTokConf, teamID string) {
	var attachments Attachment

	opts, err := localLoad(tiktok, teamID)
	if err != nil {
		errTrap(tiktok, "CRON ISSUE: Error Loading teamID `"+teamID+"` in PointsCron in `cron.go`", err)
		return
	}

	// Check for Holiday
	isHoliday, holiday := IsHoliday(tiktok, time.Now())
	if isHoliday && opts.General.HolidaySupport {
		if tiktok.Config.LogToSlack {
			LogToSlack("Today is Holiday, skipping Points Sync/Alert Scan. ("+holiday.Name+")", tiktok, attachments)
		}

		return
	}

	if tiktok.Config.LogToSlack {
		LogToSlack("Executing CRON `PointsCron` on team *"+teamID+"*", tiktok, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	returnMsg := PointCleanup(opts, tiktok, teamID)
	if tiktok.Config.DEBUG {
		fmt.Println(returnMsg)
	}
	return
}

// ArchiveCron - Execute Board Card Archiver from a Cronjob
func ArchiveCron(tiktok *TikTokConf, teamID string) {
	var attachments Attachment

	opts, err := localLoad(tiktok, teamID)
	if err != nil {
		errTrap(tiktok, "CRON ISSUE: Error Loading teamID `"+teamID+"` in ArchiveCron in `cron.go`", err)
		return
	}

	// Check for Holiday
	isHoliday, holiday := IsHoliday(tiktok, time.Now())
	if isHoliday && opts.General.HolidaySupport {
		if tiktok.Config.LogToSlack {
			LogToSlack("Today is Holiday, skipping regulary Archiving Process. ("+holiday.Name+")", tiktok, attachments)
		}
		return
	}

	if tiktok.Config.LogToSlack {
		LogToSlack("Executing CRON `ArchiveCron` on team *"+teamID+"*", tiktok, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	returnMsg, _ := CleanDone(opts, tiktok)
	if returnMsg != "" {
		if tiktok.Config.DEBUG {
			fmt.Println(returnMsg)
		}
		if tiktok.Config.LogToSlack {
			LogToSlack(returnMsg, tiktok, attachments)
		}
	}
}

// BackLogArchiveCron - Archive old cards in the backlog
func BackLogArchiveCron(tiktok *TikTokConf, teamID string) {
	var attachments Attachment

	opts, err := localLoad(tiktok, teamID)
	if err != nil {
		errTrap(tiktok, "CRON ISSUE: Error Loading teamID `"+teamID+"` in BackLogArchiveCron in `cron.go`", err)
		return
	}

	// Check for Holiday
	isHoliday, holiday := IsHoliday(tiktok, time.Now())
	if isHoliday && opts.General.HolidaySupport {
		if tiktok.Config.LogToSlack {
			LogToSlack("Today is Holiday, skipping regulary Archiving Process. ("+holiday.Name+")", tiktok, attachments)
		}
		return
	}

	if tiktok.Config.LogToSlack {
		LogToSlack("Executing CRON `BackLogArchiveCron` on team *"+teamID+"*", tiktok, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)
	err = ArchiveBacklog(tiktok, opts)
	if err != nil {
		errTrap(tiktok, "Error while attempting to archive backlog cards during CRON job on `"+teamID+"`", err)
	}
}

// CleanBacklog - Clean up the backlog
func CleanBacklog(tiktok *TikTokConf, teamID string) {
	var attachments Attachment

	opts, err := localLoad(tiktok, teamID)
	if err != nil {
		errTrap(tiktok, "CRON ISSUE: Error Loading teamID `"+teamID+"` in ArchiveCron in `cron.go`", err)
		return
	}

	// Check for Holiday
	isHoliday, holiday := IsHoliday(tiktok, time.Now())
	if isHoliday && opts.General.HolidaySupport {
		if tiktok.Config.LogToSlack {
			LogToSlack("Today is Holiday, skipping regulary Archiving Process. ("+holiday.Name+")", tiktok, attachments)
		}
		return
	}

	if tiktok.Config.LogToSlack {
		LogToSlack("Executing CRON `CleanBacklog` on team *"+teamID+"*", tiktok, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	err = CleanBackLog(opts, tiktok)
	if err != nil {
		errTrap(tiktok, "Error while attempting to cleanup the backlog on `"+teamID+"`", err)
	}
}

// CriticalBugCron - Check for cards with Critical Bug Labels
func CriticalBugCron(tiktok *TikTokConf, teamID string) {
	var attachments Attachment

	opts, err := localLoad(tiktok, teamID)
	if err != nil {
		errTrap(tiktok, "CRON ISSUE: Error Loading teamID `"+teamID+"` in CriticalBugCron in `cron.go`", err)
		return
	}

	// Check for Holiday
	isHoliday, holiday := IsHoliday(tiktok, time.Now())
	if isHoliday && opts.General.HolidaySupport {
		if tiktok.Config.LogToSlack {
			LogToSlack("Today is Holiday, skipping regulary Critical Bug Check. ("+holiday.Name+")", tiktok, attachments)
		}
		return
	}

	if tiktok.Config.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `critical-bug` on team *"+teamID+"*", tiktok, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	_ = CheckBugs(opts, tiktok)
}

// EpicLinks - Check that feature cards have Epic Links
func EpicLinks(tiktok *TikTokConf, teamID string) {
	var attachments Attachment

	opts, err := localLoad(tiktok, teamID)
	if err != nil {
		errTrap(tiktok, "CRON ISSUE: Error Loading teamID `"+teamID+"` in EpicLinks in `cron.go`", err)
		return
	}

	// Check for Holiday
	isHoliday, holiday := IsHoliday(tiktok, time.Now())
	if isHoliday && opts.General.HolidaySupport {
		if tiktok.Config.LogToSlack {
			LogToSlack("Today is Holiday, skipping daily Epic Link Check. ("+holiday.Name+")", tiktok, attachments)
		}

		return
	}

	if tiktok.Config.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `Epic-Links` on team *"+teamID+"*", tiktok, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	EpicLink(tiktok, opts)
}

// CardDataLoad - Grab card data and load into DB
func CardDataLoad(tiktok *TikTokConf, teamID string) {
	var attachments Attachment

	opts, err := localLoad(tiktok, teamID)
	if err != nil {
		errTrap(tiktok, "CRON ISSUE: Error Loading teamID `"+teamID+"` in CardDataLoad in `cron.go`", err)
		return
	}

	if tiktok.Config.LogToSlack {
		LogToSlack("Executing CRON `CardDataLoad` on team *"+teamID+"*", tiktok, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	CardPlay(tiktok, opts, "", teamID, false)
}

// ChapCount - Record Chapter Card Count for backlog
func ChapCount(tiktok *TikTokConf, teamID string) {

	if tiktok.Config.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `ChapCount` to record chapter counts in the `backlog` for team *"+teamID+"*", tiktok, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	err := RecordChapters(tiktok, teamID, "backlog")
	if tiktok.Config.DEBUG {
		fmt.Println(err.Error())
	}
}

// RetroActionCron - Check Retro boards for open action items
func RetroActionCron(tiktok *TikTokConf, teamID string) {

	opts, err := localLoad(tiktok, teamID)
	if err != nil {
		errTrap(tiktok, "CRON ISSUE: Error Loading teamID `"+teamID+"` in PRSummaryCron in `cron.go`", err)
		return
	}

	if tiktok.Config.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `retroaction` to check retro boards for action items still pending on *"+teamID+"*", tiktok, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	CheckActionCards(tiktok, opts, teamID)
}

// TemplateCheck - Check to ensure template cards are where they should be
func TemplateCheck(tiktok *TikTokConf, teamID string) {

	opts, err := localLoad(tiktok, teamID)
	if err != nil {
		errTrap(tiktok, "CRON ISSUE: Error Loading teamID `"+teamID+"` in PRSummaryCron in `cron.go`", err)
		return
	}

	if tiktok.Config.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `TemplateCheck` to check template cards on *"+teamID+"*", tiktok, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	TemplateCard(tiktok, opts)
}

// PRSummaryCron - Summarize PR's before standup
func PRSummaryCron(tiktok *TikTokConf, teamID string) {
	var attachments Attachment

	opts, err := localLoad(tiktok, teamID)
	if err != nil {
		errTrap(tiktok, "CRON ISSUE: Error Loading teamID `"+teamID+"` in PRSummaryCron in `cron.go`", err)
		return
	}

	// Check for Holiday
	isHoliday, holiday := IsHoliday(tiktok, time.Now())
	if isHoliday && opts.General.HolidaySupport {
		if tiktok.Config.LogToSlack {
			LogToSlack("Today is Holiday, skipping daily PR Summaries. ("+holiday.Name+")", tiktok, attachments)
		}

		return
	}

	if tiktok.Config.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `PRSummaryCron` on team *"+teamID+"*", tiktok, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	returnMsg, err := PRSummary(opts, tiktok)
	if tiktok.Config.DEBUG {
		fmt.Println(returnMsg + " - " + err.Error())
	}
}

// TrollCron - Execute Board Trolling from a Cronjob
func TrollCron(tiktok *TikTokConf, teamID string) {
	var attachments Attachment

	opts, err := localLoad(tiktok, teamID)
	if err != nil {
		errTrap(tiktok, "CRON ISSUE: Error Loading teamID `"+teamID+"` in TrollCron in `cron.go`", err)
		return
	}

	// Check for Holiday
	isHoliday, holiday := IsHoliday(tiktok, time.Now())
	if isHoliday && opts.General.HolidaySupport {
		if tiktok.Config.LogToSlack {
			LogToSlack("Today is Holiday, skipping Board Trolling/Alerting. ("+holiday.Name+")", tiktok, attachments)
		}

		return
	}

	if tiktok.Config.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `TrollCron` on team *"+teamID+"*", tiktok, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	returnMsg, err := AlertRunner(opts, tiktok)
	if tiktok.Config.DEBUG {
		fmt.Println(returnMsg)
	}

	// Run PR Column Skipped Check
	SkippedPR(tiktok, opts)

	return
}

// StandupAlert - Send alert message about Standup
func StandupAlert(tiktok *TikTokConf, teamID string) {

	opts, err := localLoad(tiktok, teamID)
	if err != nil {
		errTrap(tiktok, "CRON ISSUE: Error Loading teamID `"+teamID+"` in StandupAlert in `cron.go`", err)
		return
	}

	if tiktok.Config.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `standupalert` for *"+teamID+"*", tiktok, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	SendAlert(tiktok, opts, "standup")
}

// DemoAlert - Send alert message about Demos
func DemoAlert(tiktok *TikTokConf, teamID string) {

	opts, err := localLoad(tiktok, teamID)
	if err != nil {
		errTrap(tiktok, "CRON ISSUE: Error Loading teamID `"+teamID+"` in DemoAlert in `cron.go`", err)
		return
	}

	if tiktok.Config.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `demoalert` for *"+teamID+"*", tiktok, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	SendAlert(tiktok, opts, "demo")
}

// RetroAlert - Send alert message about Retro
func RetroAlert(tiktok *TikTokConf, teamID string) {

	opts, err := localLoad(tiktok, teamID)
	if err != nil {
		errTrap(tiktok, "CRON ISSUE: Error Loading teamID `"+teamID+"` in RetroAlert in `cron.go`", err)
		return
	}

	if tiktok.Config.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `retroalert` for *"+teamID+"*", tiktok, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	SendAlert(tiktok, opts, "retro")
}

// WDWAlert - Send alert message about WDW
func WDWAlert(tiktok *TikTokConf, teamID string) {

	opts, err := localLoad(tiktok, teamID)
	if err != nil {
		errTrap(tiktok, "CRON ISSUE: Error Loading teamID `"+teamID+"` in WDWAlert in `cron.go`", err)
		return
	}

	if tiktok.Config.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `wdwalert` for *"+teamID+"*", tiktok, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	SendAlert(tiktok, opts, "wdw")
}

// SDLCAlert - Send alert message about SDLC Monthly review meeting
func SDLCAlert(tiktok *TikTokConf, teamID string) {

	opts, err := localLoad(tiktok, teamID)
	if err != nil {
		errTrap(tiktok, "CRON ISSUE: Error Loading teamID `"+teamID+"` in WDWAlert in `cron.go`", err)
		return
	}

	if tiktok.Config.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `wdwalert` for *"+teamID+"*", tiktok, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	SendAlert(tiktok, opts, "sdlc")
}

func newCron(handler CronFunc, tiktok *TikTokConf, config string) func() {
	return func() { handler(tiktok, config) }
}

// CronLoad - Load or re-load all cron jobs.  Re-read the toml file
func CronLoad(tiktok *TikTokConf) (cronjobs *Cronjobs, c *cron.Cron, err error) {
	// initiate CRON system and call load routine
	c = cron.New()

	// Load Crons
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

	for _, j := range cronjobs.Cronjob {

		var attachments Attachment

		LogToSlack("Loading Cron: `"+j.Action+"` @ `"+j.Timing+"` on `"+j.Config+"`", tiktok, attachments)

		switch j.Action {
		case "standupalert":
			c.AddFunc(j.Timing, newCron(StandupAlert, tiktok, j.Config))
		case "demoalert":
			c.AddFunc(j.Timing, newCron(DemoAlert, tiktok, j.Config))
		case "retroalert":
			c.AddFunc(j.Timing, newCron(RetroAlert, tiktok, j.Config))
		case "wdwalert":
			c.AddFunc(j.Timing, newCron(WDWAlert, tiktok, j.Config))
		case "sdlcalert":
			c.AddFunc(j.Timing, newCron(SDLCAlert, tiktok, j.Config))
		case "pr":
			c.AddFunc(j.Timing, newCron(PrCron, tiktok, j.Config))
		case "troll":
			c.AddFunc(j.Timing, newCron(TrollCron, tiktok, j.Config))
		case "sprint":
			c.AddFunc(j.Timing, newCron(SprintCron, tiktok, j.Config))
		case "pts":
			c.AddFunc(j.Timing, newCron(PointsCron, tiktok, j.Config))
		case "archive":
			c.AddFunc(j.Timing, newCron(ArchiveCron, tiktok, j.Config))
		case "backlogarchive":
			c.AddFunc(j.Timing, newCron(BackLogArchiveCron, tiktok, j.Config))
		case "clean-backlog":
			c.AddFunc(j.Timing, newCron(CleanBacklog, tiktok, j.Config))
		case "pr-summary":
			c.AddFunc(j.Timing, newCron(PRSummaryCron, tiktok, j.Config))
		case "record-pts":
			c.AddFunc(j.Timing, newCron(RecordPointCron, tiktok, j.Config))
		case "count-cards":
			c.AddFunc(j.Timing, newCron(RecordThemeCount, tiktok, j.Config))
		case "holidays":
			c.AddFunc(j.Timing, newCron(HolidayTroll, tiktok, j.Config))
		case "epic-links":
			c.AddFunc(j.Timing, newCron(EpicLinks, tiktok, j.Config))
		case "chapter-count":
			c.AddFunc(j.Timing, newCron(ChapCount, tiktok, j.Config))
		case "critical-bug":
			c.AddFunc(j.Timing, newCron(CriticalBugCron, tiktok, j.Config))
		case "cardloader":
			c.AddFunc(j.Timing, newCron(CardDataLoad, tiktok, j.Config))
		case "retroaction":
			c.AddFunc(j.Timing, newCron(RetroActionCron, tiktok, j.Config))
		case "templatecheck":
			c.AddFunc(j.Timing, newCron(TemplateCheck, tiktok, j.Config))
		default:
			msg := "Warning INVALID Cron Load action called `" + j.Action + "` for Cron entry:  ```" + j.Timing + "  " + j.Config + "```"
			LogToSlack(msg, tiktok, attachments)
		}
	}

	c.Start()

	return cronjobs, c, nil
}
