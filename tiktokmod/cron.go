package tiktokmod

// Manages CRON job calls to functions

import (
	"fmt"
	"strings"
	"time"

	"github.com/robfig/cron"
)

// CronFunc - function for encompassing cron functions
type CronFunc func(baloo *BalooConf, config string)

func localLoad(baloo *BalooConf, teamID string) (opts Config, err error) {
	var attachments Attachment

	opts, err = LoadConf(baloo, teamID)

	if err != nil {
		LogToSlack("I couldn't find the team config file ("+teamID+".toml) specified in Cron job!.", baloo, attachments)
		return opts, err
	}

	return opts, err
}

// HolidayTroll - TikTok Holiday messaging Cron
func HolidayTroll(baloo *BalooConf, teamID string) {
	var attachments Attachment

	opts, err := localLoad(baloo, teamID)
	if err != nil {
		errTrap(baloo, "CRON ISSUE: Error Loading teamID `"+teamID+"` in HolidayTroll in `cron.go`", err)
		return
	}

	// Check for Holiday
	isHoliday, holiday := IsHoliday(baloo, time.Now())
	if isHoliday && opts.General.HolidaySupport {
		if strings.ToLower(holiday.Name) == "saas off-site" {
			Wrangler(baloo.Config.SlackHook, "I'm at the SaaS Off-Site today so I'm not doing my regular routine. "+holiday.Message, opts.General.ComplaintChannel, baloo.Config.SlackEmoji, attachments)
		} else {
			Wrangler(baloo.Config.SlackHook, "I'm not working today, it's a company Holiday! "+holiday.Message, opts.General.ComplaintChannel, baloo.Config.SlackEmoji, attachments)
		}
	}
}

// RecordThemeCount - Record theme card count for a specific board from a cronjob
func RecordThemeCount(baloo *BalooConf, teamID string) {
	var attachments Attachment

	if baloo.Config.LogToSlack {
		LogToSlack("Executing CRON `RecordThemeCount` on team *"+teamID+"*", baloo, attachments)
	}

	opts, err := LoadConf(baloo, teamID)
	if err != nil {
		errTrap(baloo, "CRON ISSUE: Error Loading teamID `"+teamID+"` in RecordThemeCount in `cron.go`", err)
		return
	}

	_, err = CountCards(opts, baloo, teamID)
	if err != nil {
		errTrap(baloo, "Error returned running Cron job `CountCards` function in cron.go for team "+teamID, err)
	}
}

// RecordPointCron - Record points for a specific board from a cronjob
func RecordPointCron(baloo *BalooConf, teamID string) {

	if baloo.Config.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `RecordPointCron` on team *"+teamID+"*", baloo, attachments)
	}

	opts, err := LoadConf(baloo, teamID)
	if err != nil {
		errTrap(baloo, "CRON ISSUE: Error Loading teamID `"+teamID+"` in RecordPointCron in `cron.go`", err)
		return
	}
	sOpts, err := GetDBSprint(baloo, teamID)
	if err != nil {
		errTrap(baloo, "CRON ISSUE: SQL error in `GetDBSprint` in `cron.go`", err)
		return
	}
	_, _ = GetAllPoints(baloo, opts, sOpts)

}

// SprintCron - Execute Sprint from a Cronjob
func SprintCron(baloo *BalooConf, teamID string) {

	if baloo.Config.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `SprintCron` on team *"+teamID+"*", baloo, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	opts, err := localLoad(baloo, teamID)
	if err != nil {
		errTrap(baloo, "CRON ISSUE: Error Loading teamID `"+teamID+"` in SprintCron in `cron.go`", err)
		return
	}

	returnMsg, err := Sprint(opts, baloo, false)
	if baloo.Config.DEBUG {
		fmt.Println(returnMsg)
	}

	return
}

// PrCron - Execute PR Scan from a Cronjob
func PrCron(baloo *BalooConf, teamID string) {
	var attachments Attachment

	opts, err := localLoad(baloo, teamID)
	if err != nil {
		errTrap(baloo, "CRON ISSUE: Error Loading teamID `"+teamID+"` in PrCron in `cron.go`", err)
		return
	}

	// Check for Holiday
	isHoliday, holiday := IsHoliday(baloo, time.Now())
	if isHoliday && opts.General.HolidaySupport {
		if baloo.Config.LogToSlack {
			LogToSlack("Today is Holiday, skipping PR Scan. ("+holiday.Name+")", baloo, attachments)
		}

		return
	}

	if baloo.Config.LogToSlack {
		LogToSlack("Executing CRON `PrCron` on team *"+teamID+"*", baloo, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	returnMsg, err := StalePRcards(opts, baloo)
	if baloo.Config.DEBUG {
		fmt.Println(returnMsg)
	}

	return
}

// PointsCron - Execute Points Sync from a Cronjob
func PointsCron(baloo *BalooConf, teamID string) {
	var attachments Attachment

	opts, err := localLoad(baloo, teamID)
	if err != nil {
		errTrap(baloo, "CRON ISSUE: Error Loading teamID `"+teamID+"` in PointsCron in `cron.go`", err)
		return
	}

	// Check for Holiday
	isHoliday, holiday := IsHoliday(baloo, time.Now())
	if isHoliday && opts.General.HolidaySupport {
		if baloo.Config.LogToSlack {
			LogToSlack("Today is Holiday, skipping Points Sync/Alert Scan. ("+holiday.Name+")", baloo, attachments)
		}

		return
	}

	if baloo.Config.LogToSlack {
		LogToSlack("Executing CRON `PointsCron` on team *"+teamID+"*", baloo, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	returnMsg := PointCleanup(opts, baloo, teamID)
	if baloo.Config.DEBUG {
		fmt.Println(returnMsg)
	}
	return
}

// ArchiveCron - Execute Board Card Archiver from a Cronjob
func ArchiveCron(baloo *BalooConf, teamID string) {
	var attachments Attachment

	opts, err := localLoad(baloo, teamID)
	if err != nil {
		errTrap(baloo, "CRON ISSUE: Error Loading teamID `"+teamID+"` in ArchiveCron in `cron.go`", err)
		return
	}

	// Check for Holiday
	isHoliday, holiday := IsHoliday(baloo, time.Now())
	if isHoliday && opts.General.HolidaySupport {
		if baloo.Config.LogToSlack {
			LogToSlack("Today is Holiday, skipping regulary Archiving Process. ("+holiday.Name+")", baloo, attachments)
		}
		return
	}

	if baloo.Config.LogToSlack {
		LogToSlack("Executing CRON `ArchiveCron` on team *"+teamID+"*", baloo, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	returnMsg, _ := CleanDone(opts, baloo)
	if returnMsg != "" {
		if baloo.Config.DEBUG {
			fmt.Println(returnMsg)
		}
		if baloo.Config.LogToSlack {
			LogToSlack(returnMsg, baloo, attachments)
		}
	}
}

// BackLogArchiveCron - Archive old cards in the backlog
func BackLogArchiveCron(baloo *BalooConf, teamID string) {
	var attachments Attachment

	opts, err := localLoad(baloo, teamID)
	if err != nil {
		errTrap(baloo, "CRON ISSUE: Error Loading teamID `"+teamID+"` in BackLogArchiveCron in `cron.go`", err)
		return
	}

	// Check for Holiday
	isHoliday, holiday := IsHoliday(baloo, time.Now())
	if isHoliday && opts.General.HolidaySupport {
		if baloo.Config.LogToSlack {
			LogToSlack("Today is Holiday, skipping regulary Archiving Process. ("+holiday.Name+")", baloo, attachments)
		}
		return
	}

	if baloo.Config.LogToSlack {
		LogToSlack("Executing CRON `BackLogArchiveCron` on team *"+teamID+"*", baloo, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)
	err = ArchiveBacklog(baloo, opts)
	if err != nil {
		errTrap(baloo, "Error while attempting to archive backlog cards during CRON job on `"+teamID+"`", err)
	}
}

// CleanBacklog - Clean up the backlog
func CleanBacklog(baloo *BalooConf, teamID string) {
	var attachments Attachment

	opts, err := localLoad(baloo, teamID)
	if err != nil {
		errTrap(baloo, "CRON ISSUE: Error Loading teamID `"+teamID+"` in ArchiveCron in `cron.go`", err)
		return
	}

	// Check for Holiday
	isHoliday, holiday := IsHoliday(baloo, time.Now())
	if isHoliday && opts.General.HolidaySupport {
		if baloo.Config.LogToSlack {
			LogToSlack("Today is Holiday, skipping regulary Archiving Process. ("+holiday.Name+")", baloo, attachments)
		}
		return
	}

	if baloo.Config.LogToSlack {
		LogToSlack("Executing CRON `CleanBacklog` on team *"+teamID+"*", baloo, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	err = CleanBackLog(opts, baloo)
	if err != nil {
		errTrap(baloo, "Error while attempting to cleanup the backlog on `"+teamID+"`", err)
	}
}

// CriticalBugCron - Check for cards with Critical Bug Labels
func CriticalBugCron(baloo *BalooConf, teamID string) {
	var attachments Attachment

	opts, err := localLoad(baloo, teamID)
	if err != nil {
		errTrap(baloo, "CRON ISSUE: Error Loading teamID `"+teamID+"` in CriticalBugCron in `cron.go`", err)
		return
	}

	// Check for Holiday
	isHoliday, holiday := IsHoliday(baloo, time.Now())
	if isHoliday && opts.General.HolidaySupport {
		if baloo.Config.LogToSlack {
			LogToSlack("Today is Holiday, skipping regulary Critical Bug Check. ("+holiday.Name+")", baloo, attachments)
		}
		return
	}

	if baloo.Config.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `critical-bug` on team *"+teamID+"*", baloo, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	_ = CheckBugs(opts, baloo)
}

// EpicLinks - Check that feature cards have Epic Links
func EpicLinks(baloo *BalooConf, teamID string) {
	var attachments Attachment

	opts, err := localLoad(baloo, teamID)
	if err != nil {
		errTrap(baloo, "CRON ISSUE: Error Loading teamID `"+teamID+"` in EpicLinks in `cron.go`", err)
		return
	}

	// Check for Holiday
	isHoliday, holiday := IsHoliday(baloo, time.Now())
	if isHoliday && opts.General.HolidaySupport {
		if baloo.Config.LogToSlack {
			LogToSlack("Today is Holiday, skipping daily Epic Link Check. ("+holiday.Name+")", baloo, attachments)
		}

		return
	}

	if baloo.Config.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `Epic-Links` on team *"+teamID+"*", baloo, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	EpicLink(baloo, opts)
}

// CardDataLoad - Grab card data and load into DB
func CardDataLoad(baloo *BalooConf, teamID string) {
	var attachments Attachment

	opts, err := localLoad(baloo, teamID)
	if err != nil {
		errTrap(baloo, "CRON ISSUE: Error Loading teamID `"+teamID+"` in CardDataLoad in `cron.go`", err)
		return
	}

	if baloo.Config.LogToSlack {
		LogToSlack("Executing CRON `CardDataLoad` on team *"+teamID+"*", baloo, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	CardPlay(baloo, opts, "", teamID, false)
}

// ChapCount - Record Chapter Card Count for backlog
func ChapCount(baloo *BalooConf, teamID string) {

	if baloo.Config.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `ChapCount` to record chapter counts in the `backlog` for team *"+teamID+"*", baloo, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	err := RecordChapters(baloo, teamID, "backlog")
	if baloo.Config.DEBUG {
		fmt.Println(err.Error())
	}
}

// RetroActionCron - Check Retro boards for open action items
func RetroActionCron(baloo *BalooConf, teamID string) {

	opts, err := localLoad(baloo, teamID)
	if err != nil {
		errTrap(baloo, "CRON ISSUE: Error Loading teamID `"+teamID+"` in PRSummaryCron in `cron.go`", err)
		return
	}

	if baloo.Config.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `retroaction` to check retro boards for action items still pending on *"+teamID+"*", baloo, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	CheckActionCards(baloo, opts, teamID)
}

// TemplateCheck - Check to ensure template cards are where they should be
func TemplateCheck(baloo *BalooConf, teamID string) {

	opts, err := localLoad(baloo, teamID)
	if err != nil {
		errTrap(baloo, "CRON ISSUE: Error Loading teamID `"+teamID+"` in PRSummaryCron in `cron.go`", err)
		return
	}

	if baloo.Config.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `TemplateCheck` to check template cards on *"+teamID+"*", baloo, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	TemplateCard(baloo, opts)
}

// PRSummaryCron - Summarize PR's before standup
func PRSummaryCron(baloo *BalooConf, teamID string) {
	var attachments Attachment

	opts, err := localLoad(baloo, teamID)
	if err != nil {
		errTrap(baloo, "CRON ISSUE: Error Loading teamID `"+teamID+"` in PRSummaryCron in `cron.go`", err)
		return
	}

	// Check for Holiday
	isHoliday, holiday := IsHoliday(baloo, time.Now())
	if isHoliday && opts.General.HolidaySupport {
		if baloo.Config.LogToSlack {
			LogToSlack("Today is Holiday, skipping daily PR Summaries. ("+holiday.Name+")", baloo, attachments)
		}

		return
	}

	if baloo.Config.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `PRSummaryCron` on team *"+teamID+"*", baloo, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	returnMsg, err := PRSummary(opts, baloo)
	if baloo.Config.DEBUG {
		fmt.Println(returnMsg + " - " + err.Error())
	}
}

// TrollCron - Execute Board Trolling from a Cronjob
func TrollCron(baloo *BalooConf, teamID string) {
	var attachments Attachment

	opts, err := localLoad(baloo, teamID)
	if err != nil {
		errTrap(baloo, "CRON ISSUE: Error Loading teamID `"+teamID+"` in TrollCron in `cron.go`", err)
		return
	}

	// Check for Holiday
	isHoliday, holiday := IsHoliday(baloo, time.Now())
	if isHoliday && opts.General.HolidaySupport {
		if baloo.Config.LogToSlack {
			LogToSlack("Today is Holiday, skipping Board Trolling/Alerting. ("+holiday.Name+")", baloo, attachments)
		}

		return
	}

	if baloo.Config.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `TrollCron` on team *"+teamID+"*", baloo, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	returnMsg, err := AlertRunner(opts, baloo)
	if baloo.Config.DEBUG {
		fmt.Println(returnMsg)
	}

	// Run PR Column Skipped Check
	SkippedPR(baloo, opts)

	return
}

// StandupAlert - Send alert message about Standup
func StandupAlert(baloo *BalooConf, teamID string) {

	opts, err := localLoad(baloo, teamID)
	if err != nil {
		errTrap(baloo, "CRON ISSUE: Error Loading teamID `"+teamID+"` in StandupAlert in `cron.go`", err)
		return
	}

	if baloo.Config.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `standupalert` for *"+teamID+"*", baloo, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	SendAlert(baloo, opts, "standup")
}

// DemoAlert - Send alert message about Demos
func DemoAlert(baloo *BalooConf, teamID string) {

	opts, err := localLoad(baloo, teamID)
	if err != nil {
		errTrap(baloo, "CRON ISSUE: Error Loading teamID `"+teamID+"` in DemoAlert in `cron.go`", err)
		return
	}

	if baloo.Config.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `demoalert` for *"+teamID+"*", baloo, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	SendAlert(baloo, opts, "demo")
}

// RetroAlert - Send alert message about Retro
func RetroAlert(baloo *BalooConf, teamID string) {

	opts, err := localLoad(baloo, teamID)
	if err != nil {
		errTrap(baloo, "CRON ISSUE: Error Loading teamID `"+teamID+"` in RetroAlert in `cron.go`", err)
		return
	}

	if baloo.Config.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `retroalert` for *"+teamID+"*", baloo, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	SendAlert(baloo, opts, "retro")
}

// WDWAlert - Send alert message about WDW
func WDWAlert(baloo *BalooConf, teamID string) {

	opts, err := localLoad(baloo, teamID)
	if err != nil {
		errTrap(baloo, "CRON ISSUE: Error Loading teamID `"+teamID+"` in WDWAlert in `cron.go`", err)
		return
	}

	if baloo.Config.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `wdwalert` for *"+teamID+"*", baloo, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	SendAlert(baloo, opts, "wdw")
}

// SDLCAlert - Send alert message about SDLC Monthly review meeting
func SDLCAlert(baloo *BalooConf, teamID string) {

	opts, err := localLoad(baloo, teamID)
	if err != nil {
		errTrap(baloo, "CRON ISSUE: Error Loading teamID `"+teamID+"` in WDWAlert in `cron.go`", err)
		return
	}

	if baloo.Config.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `wdwalert` for *"+teamID+"*", baloo, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	SendAlert(baloo, opts, "sdlc")
}

func newCron(handler CronFunc, baloo *BalooConf, config string) func() {
	return func() { handler(baloo, config) }
}

// CronLoad - Load or re-load all cron jobs.  Re-read the toml file
func CronLoad(baloo *BalooConf) (cronjobs *Cronjobs, c *cron.Cron, err error) {
	// initiate CRON system and call load routine
	c = cron.New()

	// Load Crons
	cronjobs, err = LoadCronFile()
	if err != nil {
		if baloo.Config.LogToSlack {
			var attachments Attachment
			LogToSlack("*WARNING!* Can not find a valid `cron.toml` file to load!! Cron's are not running!", baloo, attachments)
		}
		fmt.Println(err)
		return cronjobs, c, err
	}

	c.Stop()

	for _, j := range cronjobs.Cronjob {

		var attachments Attachment

		LogToSlack("Loading Cron: `"+j.Action+"` @ `"+j.Timing+"` on `"+j.Config+"`", baloo, attachments)

		switch j.Action {
		case "standupalert":
			c.AddFunc(j.Timing, newCron(StandupAlert, baloo, j.Config))
		case "demoalert":
			c.AddFunc(j.Timing, newCron(DemoAlert, baloo, j.Config))
		case "retroalert":
			c.AddFunc(j.Timing, newCron(RetroAlert, baloo, j.Config))
		case "wdwalert":
			c.AddFunc(j.Timing, newCron(WDWAlert, baloo, j.Config))
		case "sdlcalert":
			c.AddFunc(j.Timing, newCron(SDLCAlert, baloo, j.Config))
		case "pr":
			c.AddFunc(j.Timing, newCron(PrCron, baloo, j.Config))
		case "troll":
			c.AddFunc(j.Timing, newCron(TrollCron, baloo, j.Config))
		case "sprint":
			c.AddFunc(j.Timing, newCron(SprintCron, baloo, j.Config))
		case "pts":
			c.AddFunc(j.Timing, newCron(PointsCron, baloo, j.Config))
		case "archive":
			c.AddFunc(j.Timing, newCron(ArchiveCron, baloo, j.Config))
		case "backlogarchive":
			c.AddFunc(j.Timing, newCron(BackLogArchiveCron, baloo, j.Config))
		case "clean-backlog":
			c.AddFunc(j.Timing, newCron(CleanBacklog, baloo, j.Config))
		case "pr-summary":
			c.AddFunc(j.Timing, newCron(PRSummaryCron, baloo, j.Config))
		case "record-pts":
			c.AddFunc(j.Timing, newCron(RecordPointCron, baloo, j.Config))
		case "count-cards":
			c.AddFunc(j.Timing, newCron(RecordThemeCount, baloo, j.Config))
		case "holidays":
			c.AddFunc(j.Timing, newCron(HolidayTroll, baloo, j.Config))
		case "epic-links":
			c.AddFunc(j.Timing, newCron(EpicLinks, baloo, j.Config))
		case "chapter-count":
			c.AddFunc(j.Timing, newCron(ChapCount, baloo, j.Config))
		case "critical-bug":
			c.AddFunc(j.Timing, newCron(CriticalBugCron, baloo, j.Config))
		case "cardloader":
			c.AddFunc(j.Timing, newCron(CardDataLoad, baloo, j.Config))
		case "retroaction":
			c.AddFunc(j.Timing, newCron(RetroActionCron, baloo, j.Config))
		case "templatecheck":
			c.AddFunc(j.Timing, newCron(TemplateCheck, baloo, j.Config))
		default:
			msg := "Warning INVALID Cron Load action called `" + j.Action + "` for Cron entry:  ```" + j.Timing + "  " + j.Config + "```"
			LogToSlack(msg, baloo, attachments)
		}
	}

	c.Start()

	return cronjobs, c, nil
}
