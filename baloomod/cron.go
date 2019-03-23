package baloomod

// Manages CRON job calls to functions

import (
	"fmt"
	"strings"
	"time"

	"github.com/robfig/cron"
)

// CronFunc - function for encompassing cron functions
type CronFunc func(wOpts *WallConf, config string)

func localLoad(wOpts *WallConf, teamID string) (opts Config, err error) {
	var attachments Attachment

	opts, err = LoadConf(wOpts, teamID)

	if err != nil {
		LogToSlack("I couldn't find the team config file ("+teamID+".toml) specified in Cron job!.", wOpts, attachments)
		return opts, err
	}

	return opts, err
}

// HolidayTroll - WallE Holiday messaging Cron
func HolidayTroll(wOpts *WallConf, teamID string) {
	var attachments Attachment

	opts, err := localLoad(wOpts, teamID)
	if err != nil {
		errTrap(wOpts, "CRON ISSUE: Error Loading teamID `"+teamID+"` in HolidayTroll in `cron.go`", err)
		return
	}

	// Check for Holiday
	isHoliday, holiday := IsHoliday(wOpts, time.Now())
	if isHoliday && opts.General.HolidaySupport {
		if strings.ToLower(holiday.Name) == "saas off-site" {
			Wrangler(wOpts.Walle.SlackHook, "I'm at the SaaS Off-Site today so I'm not doing my regular routine. "+holiday.Message, opts.General.ComplaintChannel, wOpts.Walle.SlackEmoji, attachments)
		} else {
			Wrangler(wOpts.Walle.SlackHook, "I'm not working today, it's a company Holiday! "+holiday.Message, opts.General.ComplaintChannel, wOpts.Walle.SlackEmoji, attachments)
		}
	}
}

// RecordThemeCount - Record theme card count for a specific board from a cronjob
func RecordThemeCount(wOpts *WallConf, teamID string) {
	var attachments Attachment

	if wOpts.Walle.LogToSlack {
		LogToSlack("Executing CRON `RecordThemeCount` on team *"+teamID+"*", wOpts, attachments)
	}

	opts, err := LoadConf(wOpts, teamID)
	if err != nil {
		errTrap(wOpts, "CRON ISSUE: Error Loading teamID `"+teamID+"` in RecordThemeCount in `cron.go`", err)
		return
	}

	_, err = CountCards(opts, wOpts, teamID)
	if err != nil {
		errTrap(wOpts, "Error returned running Cron job `CountCards` function in cron.go for team "+teamID, err)
	}
}

// RecordPointCron - Record points for a specific board from a cronjob
func RecordPointCron(wOpts *WallConf, teamID string) {

	if wOpts.Walle.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `RecordPointCron` on team *"+teamID+"*", wOpts, attachments)
	}

	opts, err := LoadConf(wOpts, teamID)
	if err != nil {
		errTrap(wOpts, "CRON ISSUE: Error Loading teamID `"+teamID+"` in RecordPointCron in `cron.go`", err)
		return
	}
	sOpts, err := GetDBSprint(wOpts, teamID)
	if err != nil {
		errTrap(wOpts, "CRON ISSUE: SQL error in `GetDBSprint` in `cron.go`", err)
		return
	}
	_, _ = GetAllPoints(wOpts, opts, sOpts)

}

// SprintCron - Execute Sprint from a Cronjob
func SprintCron(wOpts *WallConf, teamID string) {

	if wOpts.Walle.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `SprintCron` on team *"+teamID+"*", wOpts, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	opts, err := localLoad(wOpts, teamID)
	if err != nil {
		errTrap(wOpts, "CRON ISSUE: Error Loading teamID `"+teamID+"` in SprintCron in `cron.go`", err)
		return
	}

	returnMsg, err := Sprint(opts, wOpts, false)
	if wOpts.Walle.DEBUG {
		fmt.Println(returnMsg)
	}

	return
}

// PrCron - Execute PR Scan from a Cronjob
func PrCron(wOpts *WallConf, teamID string) {
	var attachments Attachment

	opts, err := localLoad(wOpts, teamID)
	if err != nil {
		errTrap(wOpts, "CRON ISSUE: Error Loading teamID `"+teamID+"` in PrCron in `cron.go`", err)
		return
	}

	// Check for Holiday
	isHoliday, holiday := IsHoliday(wOpts, time.Now())
	if isHoliday && opts.General.HolidaySupport {
		if wOpts.Walle.LogToSlack {
			LogToSlack("Today is Holiday, skipping PR Scan. ("+holiday.Name+")", wOpts, attachments)
		}

		return
	}

	if wOpts.Walle.LogToSlack {
		LogToSlack("Executing CRON `PrCron` on team *"+teamID+"*", wOpts, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	returnMsg, err := StalePRcards(opts, wOpts)
	if wOpts.Walle.DEBUG {
		fmt.Println(returnMsg)
	}

	return
}

// PointsCron - Execute Points Sync from a Cronjob
func PointsCron(wOpts *WallConf, teamID string) {
	var attachments Attachment

	opts, err := localLoad(wOpts, teamID)
	if err != nil {
		errTrap(wOpts, "CRON ISSUE: Error Loading teamID `"+teamID+"` in PointsCron in `cron.go`", err)
		return
	}

	// Check for Holiday
	isHoliday, holiday := IsHoliday(wOpts, time.Now())
	if isHoliday && opts.General.HolidaySupport {
		if wOpts.Walle.LogToSlack {
			LogToSlack("Today is Holiday, skipping Points Sync/Alert Scan. ("+holiday.Name+")", wOpts, attachments)
		}

		return
	}

	if wOpts.Walle.LogToSlack {
		LogToSlack("Executing CRON `PointsCron` on team *"+teamID+"*", wOpts, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	returnMsg := PointCleanup(opts, wOpts, teamID)
	if wOpts.Walle.DEBUG {
		fmt.Println(returnMsg)
	}
	return
}

// ArchiveCron - Execute Board Card Archiver from a Cronjob
func ArchiveCron(wOpts *WallConf, teamID string) {
	var attachments Attachment

	opts, err := localLoad(wOpts, teamID)
	if err != nil {
		errTrap(wOpts, "CRON ISSUE: Error Loading teamID `"+teamID+"` in ArchiveCron in `cron.go`", err)
		return
	}

	// Check for Holiday
	isHoliday, holiday := IsHoliday(wOpts, time.Now())
	if isHoliday && opts.General.HolidaySupport {
		if wOpts.Walle.LogToSlack {
			LogToSlack("Today is Holiday, skipping regulary Archiving Process. ("+holiday.Name+")", wOpts, attachments)
		}
		return
	}

	if wOpts.Walle.LogToSlack {
		LogToSlack("Executing CRON `ArchiveCron` on team *"+teamID+"*", wOpts, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	returnMsg, _ := CleanDone(opts, wOpts)
	if returnMsg != "" {
		if wOpts.Walle.DEBUG {
			fmt.Println(returnMsg)
		}
		if wOpts.Walle.LogToSlack {
			LogToSlack(returnMsg, wOpts, attachments)
		}
	}
}

// BackLogArchiveCron - Archive old cards in the backlog
func BackLogArchiveCron(wOpts *WallConf, teamID string) {
	var attachments Attachment

	opts, err := localLoad(wOpts, teamID)
	if err != nil {
		errTrap(wOpts, "CRON ISSUE: Error Loading teamID `"+teamID+"` in BackLogArchiveCron in `cron.go`", err)
		return
	}

	// Check for Holiday
	isHoliday, holiday := IsHoliday(wOpts, time.Now())
	if isHoliday && opts.General.HolidaySupport {
		if wOpts.Walle.LogToSlack {
			LogToSlack("Today is Holiday, skipping regulary Archiving Process. ("+holiday.Name+")", wOpts, attachments)
		}
		return
	}

	if wOpts.Walle.LogToSlack {
		LogToSlack("Executing CRON `BackLogArchiveCron` on team *"+teamID+"*", wOpts, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)
	err = ArchiveBacklog(wOpts, opts)
	if err != nil {
		errTrap(wOpts, "Error while attempting to archive backlog cards during CRON job on `"+teamID+"`", err)
	}
}

// CleanBacklog - Clean up the backlog
func CleanBacklog(wOpts *WallConf, teamID string) {
	var attachments Attachment

	opts, err := localLoad(wOpts, teamID)
	if err != nil {
		errTrap(wOpts, "CRON ISSUE: Error Loading teamID `"+teamID+"` in ArchiveCron in `cron.go`", err)
		return
	}

	// Check for Holiday
	isHoliday, holiday := IsHoliday(wOpts, time.Now())
	if isHoliday && opts.General.HolidaySupport {
		if wOpts.Walle.LogToSlack {
			LogToSlack("Today is Holiday, skipping regulary Archiving Process. ("+holiday.Name+")", wOpts, attachments)
		}
		return
	}

	if wOpts.Walle.LogToSlack {
		LogToSlack("Executing CRON `CleanBacklog` on team *"+teamID+"*", wOpts, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	err = CleanBackLog(opts, wOpts)
	if err != nil {
		errTrap(wOpts, "Error while attempting to cleanup the backlog on `"+teamID+"`", err)
	}
}

// CriticalBugCron - Check for cards with Critical Bug Labels
func CriticalBugCron(wOpts *WallConf, teamID string) {
	var attachments Attachment

	opts, err := localLoad(wOpts, teamID)
	if err != nil {
		errTrap(wOpts, "CRON ISSUE: Error Loading teamID `"+teamID+"` in CriticalBugCron in `cron.go`", err)
		return
	}

	// Check for Holiday
	isHoliday, holiday := IsHoliday(wOpts, time.Now())
	if isHoliday && opts.General.HolidaySupport {
		if wOpts.Walle.LogToSlack {
			LogToSlack("Today is Holiday, skipping regulary Critical Bug Check. ("+holiday.Name+")", wOpts, attachments)
		}
		return
	}

	if wOpts.Walle.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `critical-bug` on team *"+teamID+"*", wOpts, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	_ = CheckBugs(opts, wOpts)
}

// EpicLinks - Check that feature cards have Epic Links
func EpicLinks(wOpts *WallConf, teamID string) {
	var attachments Attachment

	opts, err := localLoad(wOpts, teamID)
	if err != nil {
		errTrap(wOpts, "CRON ISSUE: Error Loading teamID `"+teamID+"` in EpicLinks in `cron.go`", err)
		return
	}

	// Check for Holiday
	isHoliday, holiday := IsHoliday(wOpts, time.Now())
	if isHoliday && opts.General.HolidaySupport {
		if wOpts.Walle.LogToSlack {
			LogToSlack("Today is Holiday, skipping daily Epic Link Check. ("+holiday.Name+")", wOpts, attachments)
		}

		return
	}

	if wOpts.Walle.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `Epic-Links` on team *"+teamID+"*", wOpts, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	EpicLink(wOpts, opts)
}

// CardDataLoad - Grab card data and load into DB
func CardDataLoad(wOpts *WallConf, teamID string) {
	var attachments Attachment

	opts, err := localLoad(wOpts, teamID)
	if err != nil {
		errTrap(wOpts, "CRON ISSUE: Error Loading teamID `"+teamID+"` in CardDataLoad in `cron.go`", err)
		return
	}

	if wOpts.Walle.LogToSlack {
		LogToSlack("Executing CRON `CardDataLoad` on team *"+teamID+"*", wOpts, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	CardPlay(wOpts, opts, "", teamID, false)
}

// ChapCount - Record Chapter Card Count for backlog
func ChapCount(wOpts *WallConf, teamID string) {

	if wOpts.Walle.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `ChapCount` to record chapter counts in the `backlog` for team *"+teamID+"*", wOpts, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	err := RecordChapters(wOpts, teamID, "backlog")
	if wOpts.Walle.DEBUG {
		fmt.Println(err.Error())
	}
}

// RetroActionCron - Check Retro boards for open action items
func RetroActionCron(wOpts *WallConf, teamID string) {

	opts, err := localLoad(wOpts, teamID)
	if err != nil {
		errTrap(wOpts, "CRON ISSUE: Error Loading teamID `"+teamID+"` in PRSummaryCron in `cron.go`", err)
		return
	}

	if wOpts.Walle.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `retroaction` to check retro boards for action items still pending on *"+teamID+"*", wOpts, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	CheckActionCards(wOpts, opts, teamID)
}

// TemplateCheck - Check to ensure template cards are where they should be
func TemplateCheck(wOpts *WallConf, teamID string) {

	opts, err := localLoad(wOpts, teamID)
	if err != nil {
		errTrap(wOpts, "CRON ISSUE: Error Loading teamID `"+teamID+"` in PRSummaryCron in `cron.go`", err)
		return
	}

	if wOpts.Walle.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `TemplateCheck` to check template cards on *"+teamID+"*", wOpts, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	TemplateCard(wOpts, opts)
}

// PRSummaryCron - Summarize PR's before standup
func PRSummaryCron(wOpts *WallConf, teamID string) {
	var attachments Attachment

	opts, err := localLoad(wOpts, teamID)
	if err != nil {
		errTrap(wOpts, "CRON ISSUE: Error Loading teamID `"+teamID+"` in PRSummaryCron in `cron.go`", err)
		return
	}

	// Check for Holiday
	isHoliday, holiday := IsHoliday(wOpts, time.Now())
	if isHoliday && opts.General.HolidaySupport {
		if wOpts.Walle.LogToSlack {
			LogToSlack("Today is Holiday, skipping daily PR Summaries. ("+holiday.Name+")", wOpts, attachments)
		}

		return
	}

	if wOpts.Walle.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `PRSummaryCron` on team *"+teamID+"*", wOpts, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	returnMsg, err := PRSummary(opts, wOpts)
	if wOpts.Walle.DEBUG {
		fmt.Println(returnMsg + " - " + err.Error())
	}
}

// TrollCron - Execute Board Trolling from a Cronjob
func TrollCron(wOpts *WallConf, teamID string) {
	var attachments Attachment

	opts, err := localLoad(wOpts, teamID)
	if err != nil {
		errTrap(wOpts, "CRON ISSUE: Error Loading teamID `"+teamID+"` in TrollCron in `cron.go`", err)
		return
	}

	// Check for Holiday
	isHoliday, holiday := IsHoliday(wOpts, time.Now())
	if isHoliday && opts.General.HolidaySupport {
		if wOpts.Walle.LogToSlack {
			LogToSlack("Today is Holiday, skipping Board Trolling/Alerting. ("+holiday.Name+")", wOpts, attachments)
		}

		return
	}

	if wOpts.Walle.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `TrollCron` on team *"+teamID+"*", wOpts, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	returnMsg, err := AlertRunner(opts, wOpts)
	if wOpts.Walle.DEBUG {
		fmt.Println(returnMsg)
	}

	// Run PR Column Skipped Check
	SkippedPR(wOpts, opts)

	return
}

// StandupAlert - Send alert message about Standup
func StandupAlert(wOpts *WallConf, teamID string) {

	opts, err := localLoad(wOpts, teamID)
	if err != nil {
		errTrap(wOpts, "CRON ISSUE: Error Loading teamID `"+teamID+"` in StandupAlert in `cron.go`", err)
		return
	}

	if wOpts.Walle.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `standupalert` for *"+teamID+"*", wOpts, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	SendAlert(wOpts, opts, "standup")
}

// DemoAlert - Send alert message about Demos
func DemoAlert(wOpts *WallConf, teamID string) {

	opts, err := localLoad(wOpts, teamID)
	if err != nil {
		errTrap(wOpts, "CRON ISSUE: Error Loading teamID `"+teamID+"` in DemoAlert in `cron.go`", err)
		return
	}

	if wOpts.Walle.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `demoalert` for *"+teamID+"*", wOpts, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	SendAlert(wOpts, opts, "demo")
}

// RetroAlert - Send alert message about Retro
func RetroAlert(wOpts *WallConf, teamID string) {

	opts, err := localLoad(wOpts, teamID)
	if err != nil {
		errTrap(wOpts, "CRON ISSUE: Error Loading teamID `"+teamID+"` in RetroAlert in `cron.go`", err)
		return
	}

	if wOpts.Walle.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `retroalert` for *"+teamID+"*", wOpts, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	SendAlert(wOpts, opts, "retro")
}

// WDWAlert - Send alert message about WDW
func WDWAlert(wOpts *WallConf, teamID string) {

	opts, err := localLoad(wOpts, teamID)
	if err != nil {
		errTrap(wOpts, "CRON ISSUE: Error Loading teamID `"+teamID+"` in WDWAlert in `cron.go`", err)
		return
	}

	if wOpts.Walle.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `wdwalert` for *"+teamID+"*", wOpts, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	SendAlert(wOpts, opts, "wdw")
}

// SDLCAlert - Send alert message about SDLC Monthly review meeting
func SDLCAlert(wOpts *WallConf, teamID string) {

	opts, err := localLoad(wOpts, teamID)
	if err != nil {
		errTrap(wOpts, "CRON ISSUE: Error Loading teamID `"+teamID+"` in WDWAlert in `cron.go`", err)
		return
	}

	if wOpts.Walle.LogToSlack {
		var attachments Attachment
		LogToSlack("Executing CRON `wdwalert` for *"+teamID+"*", wOpts, attachments)
	}

	// Go runs faster then 1 second on cron jobs, so multiple runs happen per second.  this sleeps 1 sec
	time.Sleep(1000 * time.Millisecond)

	SendAlert(wOpts, opts, "sdlc")
}

func newCron(handler CronFunc, wOpts *WallConf, config string) func() {
	return func() { handler(wOpts, config) }
}

// CronLoad - Load or re-load all cron jobs.  Re-read the toml file
func CronLoad(wOpts *WallConf) (cronjobs *Cronjobs, c *cron.Cron, err error) {
	// initiate CRON system and call load routine
	c = cron.New()

	// Load Wall*E Crons
	cronjobs, err = LoadCronFile()
	if err != nil {
		if wOpts.Walle.LogToSlack {
			var attachments Attachment
			LogToSlack("*WARNING!* Can not find a valid `cron.toml` file to load!! Cron's are not running!", wOpts, attachments)
		}
		fmt.Println(err)
		return cronjobs, c, err
	}

	c.Stop()

	for _, j := range cronjobs.Cronjob {

		var attachments Attachment

		LogToSlack("Loading Cron: `"+j.Action+"` @ `"+j.Timing+"` on `"+j.Config+"`", wOpts, attachments)

		switch j.Action {
		case "standupalert":
			c.AddFunc(j.Timing, newCron(StandupAlert, wOpts, j.Config))
		case "demoalert":
			c.AddFunc(j.Timing, newCron(DemoAlert, wOpts, j.Config))
		case "retroalert":
			c.AddFunc(j.Timing, newCron(RetroAlert, wOpts, j.Config))
		case "wdwalert":
			c.AddFunc(j.Timing, newCron(WDWAlert, wOpts, j.Config))
		case "sdlcalert":
			c.AddFunc(j.Timing, newCron(SDLCAlert, wOpts, j.Config))
		case "pr":
			c.AddFunc(j.Timing, newCron(PrCron, wOpts, j.Config))
		case "troll":
			c.AddFunc(j.Timing, newCron(TrollCron, wOpts, j.Config))
		case "sprint":
			c.AddFunc(j.Timing, newCron(SprintCron, wOpts, j.Config))
		case "pts":
			c.AddFunc(j.Timing, newCron(PointsCron, wOpts, j.Config))
		case "archive":
			c.AddFunc(j.Timing, newCron(ArchiveCron, wOpts, j.Config))
		case "backlogarchive":
			c.AddFunc(j.Timing, newCron(BackLogArchiveCron, wOpts, j.Config))
		case "clean-backlog":
			c.AddFunc(j.Timing, newCron(CleanBacklog, wOpts, j.Config))
		case "pr-summary":
			c.AddFunc(j.Timing, newCron(PRSummaryCron, wOpts, j.Config))
		case "record-pts":
			c.AddFunc(j.Timing, newCron(RecordPointCron, wOpts, j.Config))
		case "count-cards":
			c.AddFunc(j.Timing, newCron(RecordThemeCount, wOpts, j.Config))
		case "holidays":
			c.AddFunc(j.Timing, newCron(HolidayTroll, wOpts, j.Config))
		case "epic-links":
			c.AddFunc(j.Timing, newCron(EpicLinks, wOpts, j.Config))
		case "chapter-count":
			c.AddFunc(j.Timing, newCron(ChapCount, wOpts, j.Config))
		case "critical-bug":
			c.AddFunc(j.Timing, newCron(CriticalBugCron, wOpts, j.Config))
		case "cardloader":
			c.AddFunc(j.Timing, newCron(CardDataLoad, wOpts, j.Config))
		case "retroaction":
			c.AddFunc(j.Timing, newCron(RetroActionCron, wOpts, j.Config))
		case "templatecheck":
			c.AddFunc(j.Timing, newCron(TemplateCheck, wOpts, j.Config))
		default:
			msg := "Warning INVALID Cron Load action called `" + j.Action + "` for Cron entry:  ```" + j.Timing + "  " + j.Config + "```"
			LogToSlack(msg, wOpts, attachments)
		}
	}

	c.Start()

	return cronjobs, c, nil
}
