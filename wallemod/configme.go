package wallemod

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/nlopes/slack"
)

/*
This module is called from Wall-E command Build Config
It builds a config module for a trello board for Wall-E to manage
SEE README.md for more Details about how to use this
*/

// BacklogName - Default name of Backlog List
var BacklogName = "Backlog"

// Upcoming - Default name of Upcoming/Un-Scoped List
var Upcoming = "Upcoming Un-Scoped"

// Scoped - Default name of Backlog List
var Scoped = "Ready for Points"

// NextSprintName - Default name of Next Sprint List
var NextSprintName = "Next Sprint"

// ReadyForWork - Default name of Ready for Work List
var ReadyForWork = "Ready for Work"

// WorkingName - Default name of Working List
var WorkingName = "Working"

// RFRName - Default name of PR List
var RFRName = "Ready for Review (PR)"

// DoneName - Default name of Done List
var DoneName = "Done"

// LabelRO - Default name of Roll-over Label
var LabelRO = "ROLL-OVER"

// SprintName - Default name of custom Field for sprint name
var SprintName = "Sprint"

// SprintPoints - Default name of custom field for story points
var SprintPoints = "Burndown"

// TemplateLabelID - Default label name of Template cards
var TemplateLabelID = "TEMPLATE CARD DO NOT MOVE"

// AllowMembersLabel - Default label name for cards to ignore "face"/"owner" alerts
var AllowMembersLabel = "DEMO"

// TrainingLabel - Default label name for training cards
var TrainingLabel = "Training"

// SilenceCardLabel - Default label name for label that will disable 98% of wall-e monitoring/alerting
var SilenceCardLabel = "Wall-E Hush"

// ConfigMe struct for passing config data around
type ConfigMe struct {
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
	CfsprintID        string
	CfpointsID        string
	TemplateLabelID   string
	AllowMembersLabel string
	TrainingLabel     string
	SilenceCardLabel  string
}

// LabelCollection - handles multiple plugin types per board
type LabelCollection []*Labels

// CustomCollection - handles mutiple custom fields
type CustomCollection []*Customs

// Customs struct
type Customs struct {
	ID         string `json:"id"`
	IDMOdel    string `json:"idModel"`
	ModelType  string `json:"modelType"`
	FieldGroup string `json:"fieldGroup"`
	Name       string `json:"name"`
	Pos        int    `json:"pos"`
	Type       string `json:"type"`
}

// Labels struct
type Labels struct {
	ID      string `json:"id"`
	IDBoard string `json:"idBoard"`
	Name    string `json:"name"`
	Color   string `json:"color"`
	Uses    int    `json:"uses"`
}

// GetBoardCustoms - grab all custom fields on a board
func GetBoardCustoms(boardID string, wOpts *WallConf) (customList CustomCollection, err error) {
	url := "https://api.trello.com/1/boards/" + boardID + "/customFields?key=" + wOpts.Walle.Tkey + "&token=" + wOpts.Walle.Ttoken

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		errTrap(wOpts, "Error http Request to trello in `configme.go` func `GetBoardCustoms`", err)
		return customList, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errTrap(wOpts, "Error http.Client request to trello in `configme.go` func `GetBoardCustoms`", err)
		return customList, err
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&customList)

	return customList, err
}

// GetBoardLabels - grab all labels on a board.  not part of Adilo/trello
func GetBoardLabels(boardID string, wOpts *WallConf) (labelList LabelCollection, err error) {
	url := "https://api.trello.com/1/boards/" + boardID + "/labels?key=" + wOpts.Walle.Tkey + "&token=" + wOpts.Walle.Ttoken

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		errTrap(wOpts, "Error http Request to trello in `configme.go` func `GetBoardLabels`", err)
		return labelList, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errTrap(wOpts, "Error http.Client request to trello in `configme.go` func `GetBoardLabels`", err)
		return labelList, err
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&labelList)

	return labelList, err
}

// BuildConfig - grab UID's out of trello to help setup config files
func BuildConfig(boardID string, wOpts *WallConf, user string, api *slack.Client) {

	var config ConfigMe
	var attachments Attachment
	var myPayload BotDMPayload
	var message string

	attachments.Text = ""
	attachments.Color = ""

	config.BoardID = boardID
	userInfo, _ := api.GetUserInfo(user)

	// handle columns
	listData, err := GetLists(wOpts, boardID)
	if err != nil {
		errTrap(wOpts, "Trying to run Config Builder for "+userInfo.Name+" but had request for all lists on board `"+boardID+"` returned error.", err)
		return
	}

	for _, l := range listData {
		listName := strings.ToLower(l.Name)

		if listName == strings.ToLower(BacklogName) {
			config.BacklogID = l.ID
		}
		if listName == strings.ToLower(Upcoming) {
			config.Upcoming = l.ID
		}
		if listName == strings.ToLower(Scoped) {
			config.Scoped = l.ID
		}
		if listName == strings.ToLower(NextSprintName) {
			config.NextsprintID = l.ID
		}
		if listName == strings.ToLower(ReadyForWork) {
			config.ReadyForWork = l.ID
		}
		if listName == strings.ToLower(WorkingName) {
			config.Working = l.ID
		}
		if listName == strings.ToLower(RFRName) {
			config.ReadyForReview = l.ID
		}
		if listName == strings.ToLower(DoneName) {
			config.Done = l.ID
		}
	}

	if config.BacklogID == "" {
		message = message + "Failed to find ID for list " + BacklogName + "\n"
	}
	if config.Upcoming == "" {
		message = message + "Failed to find ID for list " + Upcoming + "\n"
	}
	if config.Scoped == "" {
		message = message + "Failed to find ID for list " + Scoped + "\n"
	}
	if config.NextsprintID == "" {
		message = message + "Failed to find ID for list " + NextSprintName + "\n"
	}
	if config.ReadyForWork == "" {
		message = message + "Failed to find ID for list " + ReadyForWork + "\n"
	}
	if config.Working == "" {
		message = message + "Failed to find ID for list " + WorkingName + "\n"
	}
	if config.ReadyForReview == "" {
		message = message + "Failed to find ID for list " + RFRName + "\n"
	}
	if config.Done == "" {
		message = message + "Failed to find ID for list " + DoneName + "\n"
	}

	myPayload.Text = message
	myPayload.Channel = userInfo.ID
	_ = WranglerDM(wOpts, myPayload)

	// handle labels
	labelSet, _ := GetBoardLabels(boardID, wOpts)

	for _, l := range labelSet {
		if l.Name == LabelRO {
			config.ROLabelID = l.ID
		}
		if l.Name == TemplateLabelID {
			config.TemplateLabelID = l.ID
		}
		if l.Name == AllowMembersLabel {
			config.AllowMembersLabel = l.ID
		}
		if l.Name == TrainingLabel {
			config.TrainingLabel = l.ID
		}
		if l.Name == SilenceCardLabel {
			config.SilenceCardLabel = l.ID
		}
	}

	message = ""
	if config.ROLabelID == "" {
		message = message + "Failed to find ID for Label " + LabelRO + "\n"
	}
	if config.TemplateLabelID == "" {
		message = message + "Failed to find id for Label " + TemplateLabelID + "\n"
	}
	if config.AllowMembersLabel == "" {
		message = message + "Failed to find id for label " + AllowMembersLabel + "\n"
	}
	if config.TrainingLabel == "" {
		message = message + "Failed to find id for label " + TrainingLabel + "\n"
	}
	if config.SilenceCardLabel == "" {
		message = message + "Failed to find id for label " + SilenceCardLabel + "\n"
	}

	myPayload.Text = message
	_ = WranglerDM(wOpts, myPayload)

	// handle custom fields
	customSet, _ := GetBoardCustoms(boardID, wOpts)

	for _, c := range customSet {
		if c.Name == SprintName {
			config.CfsprintID = c.ID
		}
		if c.Name == SprintPoints {
			config.CfpointsID = c.ID
		}
	}

	message = ""
	if config.CfsprintID == "" {
		message = message + "Failed to find ID for Custom Field " + SprintName + "\n"
	}
	if config.CfpointsID == "" {
		message = message + "Failed to find ID for Custom Field " + SprintPoints + "\n"
	}

	myPayload.Text = message
	_ = WranglerDM(wOpts, myPayload)

	message = "*Hey, here's the config data you need to make your TOML file!*\n"

	message = message + "```"
	message = message + "BacklogID          = " + config.BacklogID + "\n"
	message = message + "Upcoming           = " + config.Upcoming + "\n"
	message = message + "Scoped             = " + config.Scoped + "\n"
	message = message + "NextsprintID       = " + config.NextsprintID + "\n"
	message = message + "ReadyForWork       = " + config.ReadyForWork + "\n"
	message = message + "Working            = " + config.Working + "\n"
	message = message + "ReadyForReview     = " + config.ReadyForReview + "\n"
	message = message + "Done               = " + config.Done + "\n"
	message = message + "BoardID            = " + boardID + "\n"
	message = message + "ROLabelID          = " + config.ROLabelID + "\n"
	message = message + "CfsprintID         = " + config.CfsprintID + "\n"
	message = message + "CfpointsID         = " + config.CfpointsID + "\n"
	message = message + "TemplateLabelID    = " + config.TemplateLabelID + "\n"
	message = message + "AllowMembersLabel  = " + config.AllowMembersLabel + "\n"
	message = message + "TrainingLabel      = " + config.TrainingLabel + "\n"
	message = message + "SilenceCardLabel   = " + config.SilenceCardLabel + "\n"
	message = message + "```"

	myPayload.Text = message
	_ = WranglerDM(wOpts, myPayload)
}
