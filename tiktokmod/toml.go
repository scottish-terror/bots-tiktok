package tiktokmod

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/jinzhu/copier"
)

// Cronjobs struct. For holding all cron job info from TOML
type Cronjobs struct {
	Cronjob []struct {
		Timing string
		Action string
		Config string
	}
}

// TikTokStruct primary configuration struct
type TikTokStruct struct {
	SlackHook               string
	SlackToken              string
	SlackOAuth              string
	Tkey                    string
	Ttoken                  string
	GitToken                string
	DBUser                  string
	DBPassword              string
	DEBUG                   bool
	LogChannel              string
	SlackEmoji              string
	LogToSlack              bool
	AdminSlackChannel       string
	ScrumControlChannel     string
	DupeCollectionID        string
	LoggingPrefix           string
	Version                 string
	BotID                   string
	BotName                 string
	TeamID                  string
	TeamName                string
	PointsPowerUpID         string
	BotTrelloID             string
	TrelloOrgID             string
	UseGCP                  bool
	SQLHost                 string
	SQLPort                 string
	SQLDBName               string
	AllowNativePasswords    bool
	AllowCleartextPasswords bool
	AllowAllFiles           bool
	ParseTime               bool
	GithubOrgName           string
}

//GeneralOptions struct for configs
type GeneralOptions struct {
	TeamName        string
	Sprintname      string
	TrelloOrg       string
	StaleTime       int
	MaxPoints       int
	ArchiveDoneDays int
	BackLogDays     int
	SprintDuration  int
	RetroActionDays int
	IgnoreWeekends  bool
	HolidaySupport  bool

	BacklogID         string
	Upcoming          string
	Scoped            string
	NextsprintID      string
	ReadyForWork      string
	Working           string
	ReadyForReview    string
	Done              string
	BoardID           string
	ROLabelID         string
	TemplateLabelID   string
	CfsprintID        string
	CfpointsID        string
	RetroCollectionID string
	AllowMembersLabel string
	TrainingLabel     string
	SilenceCardLabel  string
	DemoBoardID       string

	RetroChannel     string
	SprintChannel    string
	ComplaintChannel string

	StandupAlertChannel string
	StandupLink         string
	DemoAlertChannel    string
	DemoAlertLink       string
	RetroAlertChannel   string
	RetroAlertLink      string
	WDWAlertChannel     string
	WDWAlertLink        string
}

// Config - Struct of option file sections
type Config struct {
	General GeneralOptions
}

// TikTokConf - Struct of tiktok conf file section
type TikTokConf struct {
	Config TikTokStruct
}

var conf Config
var tiktok TikTokConf
var jobList Cronjobs

// LoadCronFile - CRON Tabs
func LoadCronFile() (*Cronjobs, error) {
	configFile := "cfg/crons.toml"
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, errors.New("cron file does not exist - crons.toml must exist in run directory")
	} else if err != nil {
		return nil, err
	}

	if _, err := toml.DecodeFile(configFile, &jobList); err != nil {
		return nil, err
	}

	return &jobList, nil
}

// LoadTikTokConf Main Config
func LoadTikTokConf() (*TikTokConf, error) {
	configFile := "cfg/tiktok.toml"
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, errors.New("config file does not exist - tiktok.toml must exist in run directory")
	} else if err != nil {
		return nil, err
	}

	if _, err := toml.DecodeFile(configFile, &tiktok); err != nil {
		return nil, err
	}

	return &tiktok, nil
}

// LoadConfig - load toml config file
func LoadConfig(configFile string) (*Config, error) {
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, errors.New("config file " + configFile + " does not exist")
	} else if err != nil {
		return nil, err
	}

	if _, err := toml.DecodeFile(configFile, &conf); err != nil {
		return nil, err
	}

	return &conf, nil
}

// SanityCheck - Check for Valid Config file.  Determines struct values exist and are not blank
func SanityCheck(ConfigLocation string, ms GeneralOptions) (sane bool, output string) {

	// struct field value
	msValuePtr := reflect.ValueOf(&ms)
	msValue := msValuePtr.Elem()

	// struct field name
	msTypePtr := reflect.TypeOf(&ms)
	msTvalue := msTypePtr.Elem()

	message := ""
	for i := 0; i < msValue.NumField(); i++ {
		field := msValue.Field(i)
		typed := msTvalue.Field(i)

		// Ignore fields that don't have the same type as a string
		if field.Type() != reflect.TypeOf("") {
			continue
		}

		str := field.Interface().(string)
		str = strings.TrimSpace(str)
		typ := typed.Name

		field.SetString(str)
		if str == "" {
			// ignore these fields which can be blank
			if typ == "RetroCollectionID" || typ == "DemoBoardID" || typ == "StandupAlertChannel" || typ == "StandupLink" || typ == "DemoAlertChannel" || typ == "DemoAlertLink" || typ == "RetroAlertChannel" || typ == "RetroAlertLink" || typ == "WDWAlertChannel" || typ == "WDWAlertLink" {
				str = ""
			} else {
				message = message + "Value " + typ + " can not be blank!\n"
			}
		}
	}
	if message == "" {
		return true, "\nValid Configuration File Found"
	}
	message = "\nConfiguration File " + ConfigLocation + " is invalid: \n\n" + message
	return false, message
}

// LoadConf - load a teams conf file to do something
func LoadConf(tiktok *TikTokConf, team string) (opts Config, err error) {

	var attachments Attachment

	configLocation := "cfg/" + team + ".toml"

	// Load the config file
	slopts, err := LoadConfig(configLocation)
	if err != nil {
		errTrap(tiktok, "Failure Loading requested team file `"+configLocation+"`.", err)
		return opts, err
	}

	copier.Copy(&opts, &slopts)

	// Run sanity check on the config file
	sane, output := SanityCheck(configLocation, opts.General)
	if !sane {
		if tiktok.Config.DEBUG {
			fmt.Println(output)
		}
		if tiktok.Config.LogToSlack {
			LogToSlack("Config file failed Sanity check for team file `"+configLocation+"`. ```"+output+"```", tiktok, attachments)
		}
		return opts, err
	}

	return opts, nil
}

// FindToml - Get list of TOML files
func FindToml(tiktok *TikTokConf) (tomls []os.FileInfo, err error) {

	tomls, err = ioutil.ReadDir("cfg/")

	if err != nil {
		errTrap(tiktok, "Error attempting to read directory listing for `./*.toml`", err)
		return nil, err
	}

	return tomls, nil
}

// ListAllTOML - list all the available TOML files in a string
func ListAllTOML(tiktok *TikTokConf) (message string) {

	tomls, _ := FindToml(tiktok)

	for _, f := range tomls {

		if f.Name() != "example.toml" && f.Name() != "crons.toml" && f.Name() != "tiktok.toml" {
			s := strings.Split(f.Name(), ".")

			if s[len(s)-1] == "toml" {
				opts, _ := LoadConf(tiktok, s[0])
				message = message + "<https://trello.com/b/" + opts.General.BoardID + "|" + opts.General.TeamName + " trello board>.  Refer to ID: [" + s[0] + "]\n"
			}

		}

	}

	return message

}
