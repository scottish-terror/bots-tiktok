package baloomod

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// CardAttachment - struct for card attachment data
type CardAttachment struct {
	ID        string              `json:"id"`
	Name      string              `json:"name"`
	Pos       float32             `json:"pos"`
	Bytes     int                 `json:"int"`
	Date      string              `json:"date"`
	EdgeColor string              `json:"edgeColor"`
	IDMember  string              `json:"idMember"`
	IsUpload  bool                `json:"isUpload"`
	MimeType  string              `json:"mimeType"`
	Previews  []AttachmentPreview `json:"previews"`
	URL       string              `json:"url"`
}

// AttachmentPreview - struct for attachment preview data
type AttachmentPreview struct {
	ID     string `json:"_id"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Bytes  int    `json:"bytes"`
	Scaled bool   `json:"scaled"`
}

// PluginCollection handles multiple plugin types per board
type PluginCollection []*PluginCard

// PValue custom field value struct
type PValue struct {
	Points        string   `json:"points"`
	PointsHistory []string `json:"pointsHistory,omitempty"`
}

//PluginCard - PowerUp Plugin Struct
type PluginCard struct {
	ID       string `json:"id"`
	IDPlugin string `json:"idPlugin"`
	Scope    string `json:"scope"`
	IDModel  string `json:"idModel"`
	Value    string `json:"value"`
	Access   string `json:"access"`
}

// PointsHistory struct for managing points and points history for plugin
type PointsHistory struct {
	PointsHistory []string `json:"pointsHistory"`
	Points        int      `json:"points"`
}

//Boards - struct to create trello boards
type Boards struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	Desc           string      `json:"desc"`
	DescData       interface{} `json:"descData"`
	Closed         bool        `json:"closed"`
	IDOrganization string      `json:"idOrganization"`
	Pinned         bool        `json:"pinned"`
	URL            string      `json:"url"`
	ShortURL       string      `json:"shortUrl"`
	Prefs          struct {
		PermissionLevel       string      `json:"permissionLevel"`
		Voting                string      `json:"voting"`
		Comments              string      `json:"comments"`
		Invitations           string      `json:"invitations"`
		SelfJoin              bool        `json:"selfJoin"`
		CardCovers            bool        `json:"cardCovers"`
		CardAging             string      `json:"cardAging"`
		CalendarFeedEnabled   bool        `json:"calendarFeedEnabled"`
		Background            string      `json:"background"`
		BackgroundImage       interface{} `json:"backgroundImage"`
		BackgroundImageScaled interface{} `json:"backgroundImageScaled"`
		BackgroundTile        bool        `json:"backgroundTile"`
		BackgroundBrightness  string      `json:"backgroundBrightness"`
		BackgroundColor       string      `json:"backgroundColor"`
		BackgroundBottomColor string      `json:"backgroundBottomColor"`
		BackgroundTopColor    string      `json:"backgroundTopColor"`
		CanBePublic           bool        `json:"canBePublic"`
		CanBeOrg              bool        `json:"canBeOrg"`
		CanBePrivate          bool        `json:"canBePrivate"`
		CanInvite             bool        `json:"canInvite"`
	} `json:"prefs"`
	LabelNames struct {
		Green  string `json:"green"`
		Yellow string `json:"yellow"`
		Orange string `json:"orange"`
		Red    string `json:"red"`
		Purple string `json:"purple"`
		Blue   string `json:"blue"`
		Sky    string `json:"sky"`
		Lime   string `json:"lime"`
		Pink   string `json:"pink"`
		Black  string `json:"black"`
	} `json:"labelNames"`
	Limits struct {
	} `json:"limits"`
}

// BoardData - struct for nested Board/List/Card API call
type BoardData struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	Desc           string      `json:"desc"`
	DescData       interface{} `json:"descData"`
	Closed         bool        `json:"closed"`
	IDOrganization string      `json:"idOrganization"`
	Pinned         bool        `json:"pinned"`
	URL            string      `json:"url"`
	ShortURL       string      `json:"shortUrl"`
	Prefs          struct {
		PermissionLevel       string `json:"permissionLevel"`
		Voting                string `json:"voting"`
		Comments              string `json:"comments"`
		Invitations           string `json:"invitations"`
		SelfJoin              bool   `json:"selfJoin"`
		CardCovers            bool   `json:"cardCovers"`
		CardAging             string `json:"cardAging"`
		CalendarFeedEnabled   bool   `json:"calendarFeedEnabled"`
		Background            string `json:"background"`
		BackgroundImage       string `json:"backgroundImage"`
		BackgroundImageScaled []struct {
			Width  int    `json:"width"`
			Height int    `json:"height"`
			URL    string `json:"url"`
		} `json:"backgroundImageScaled"`
		BackgroundTile        bool   `json:"backgroundTile"`
		BackgroundBrightness  string `json:"backgroundBrightness"`
		BackgroundBottomColor string `json:"backgroundBottomColor"`
		BackgroundTopColor    string `json:"backgroundTopColor"`
		CanBePublic           bool   `json:"canBePublic"`
		CanBeOrg              bool   `json:"canBeOrg"`
		CanBePrivate          bool   `json:"canBePrivate"`
		CanInvite             bool   `json:"canInvite"`
	} `json:"prefs"`
	LabelNames struct {
		Green  string `json:"green"`
		Yellow string `json:"yellow"`
		Orange string `json:"orange"`
		Red    string `json:"red"`
		Purple string `json:"purple"`
		Blue   string `json:"blue"`
		Sky    string `json:"sky"`
		Lime   string `json:"lime"`
		Pink   string `json:"pink"`
		Black  string `json:"black"`
	} `json:"labelNames"`
	Cards []struct {
		ID                    string        `json:"id"`
		CheckItemStates       interface{}   `json:"checkItemStates"`
		Closed                bool          `json:"closed"`
		DateLastActivity      time.Time     `json:"dateLastActivity"`
		Desc                  string        `json:"desc"`
		DescData              interface{}   `json:"descData"`
		IDBoard               string        `json:"idBoard"`
		IDList                string        `json:"idList"`
		IDMembersVoted        []interface{} `json:"idMembersVoted"`
		IDShort               int           `json:"idShort"`
		IDAttachmentCover     interface{}   `json:"idAttachmentCover"`
		IDLabels              []interface{} `json:"idLabels"`
		ManualCoverAttachment bool          `json:"manualCoverAttachment"`
		Name                  string        `json:"name"`
		Pos                   int           `json:"pos"`
		ShortLink             string        `json:"shortLink"`
		Badges                struct {
			Votes             int `json:"votes"`
			AttachmentsByType struct {
				Trello struct {
					Board int `json:"board"`
					Card  int `json:"card"`
				} `json:"trello"`
			} `json:"attachmentsByType"`
			ViewingMemberVoted bool        `json:"viewingMemberVoted"`
			Subscribed         bool        `json:"subscribed"`
			Fogbugz            string      `json:"fogbugz"`
			CheckItems         int         `json:"checkItems"`
			CheckItemsChecked  int         `json:"checkItemsChecked"`
			Comments           int         `json:"comments"`
			Attachments        int         `json:"attachments"`
			Description        bool        `json:"description"`
			Due                interface{} `json:"due"`
			DueComplete        bool        `json:"dueComplete"`
		} `json:"badges"`
		DueComplete  bool          `json:"dueComplete"`
		Due          interface{}   `json:"due"`
		IDChecklists []interface{} `json:"idChecklists"`
		IDMembers    []string      `json:"idMembers"`
		Labels       []struct {
			ID      string `json:"id"`
			IDBoard string `json:"idBoard"`
			Name    string `json:"name"`
			Color   string `json:"color"`
		} `json:"labels"`
		ShortURL         string `json:"shortUrl"`
		Subscribed       bool   `json:"subscribed"`
		URL              string `json:"url"`
		CustomFieldItems []struct {
			ID    string `json:"id"`
			Value struct {
				Text   string `json:"text"`
				Number string `json:"number"`
			} `json:"value"`
			IDCustomField string `json:"idCustomField"`
			IDModel       string `json:"idModel"`
			ModelType     string `json:"modelType"`
		} `json:"customFieldItems"`
	} `json:"cards"`
}

// CardDescHistory - struct to contain description history of any given trello card
type CardDescHistory []struct {
	ID              string `json:"id"`
	IDMemberCreator string `json:"idMemberCreator"`
	Data            struct {
		List struct {
			Name string `json:"name"`
			ID   string `json:"id"`
		} `json:"list"`
		Board struct {
			ShortLink string `json:"shortLink"`
			Name      string `json:"name"`
			ID        string `json:"id"`
		} `json:"board"`
		Card struct {
			ShortLink string `json:"shortLink"`
			IDShort   int    `json:"idShort"`
			Name      string `json:"name"`
			ID        string `json:"id"`
			Desc      string `json:"desc"`
		} `json:"card"`
		Old struct {
			Desc string `json:"desc"`
		} `json:"old"`
	} `json:"data"`
	Type   string    `json:"type"`
	Date   time.Time `json:"date"`
	Limits struct {
	} `json:"limits"`
	MemberCreator struct {
		ID         string `json:"id"`
		AvatarHash string `json:"avatarHash"`
		AvatarURL  string `json:"avatarUrl"`
		FullName   string `json:"fullName"`
		Initials   string `json:"initials"`
		Username   string `json:"username"`
	} `json:"memberCreator"`
}

// CardListHistory - struct to contain list history of any given trello card
type CardListHistory []struct {
	ID              string `json:"id"`
	IDMemberCreator string `json:"idMemberCreator"`
	Data            struct {
		ListAfter struct {
			Name string `json:"name"`
			ID   string `json:"id"`
		} `json:"listAfter"`
		ListBefore struct {
			Name string `json:"name"`
			ID   string `json:"id"`
		} `json:"listBefore"`
		Board struct {
			ShortLink string `json:"shortLink"`
			Name      string `json:"name"`
			ID        string `json:"id"`
		} `json:"board"`
		Card struct {
			ShortLink string `json:"shortLink"`
			IDShort   int    `json:"idShort"`
			Name      string `json:"name"`
			ID        string `json:"id"`
			IDList    string `json:"idList"`
		} `json:"card"`
		Old struct {
			IDList string `json:"idList"`
		} `json:"old"`
	} `json:"data"`
	Type   string    `json:"type"`
	Date   time.Time `json:"date"`
	Limits struct {
	} `json:"limits"`
	MemberCreator struct {
		ID         string `json:"id"`
		AvatarHash string `json:"avatarHash"`
		AvatarURL  string `json:"avatarUrl"`
		FullName   string `json:"fullName"`
		Initials   string `json:"initials"`
		Username   string `json:"username"`
	} `json:"memberCreator"`
}

// CardAction - struct for holding card actions
type CardAction []struct {
	ID              string `json:"id"`
	IDMemberCreator string `json:"idMemberCreator"`
	Data            struct {
		List struct {
			Name string `json:"name"`
			ID   string `json:"id"`
		} `json:"list"`
		Board struct {
			ShortLink string `json:"shortLink"`
			Name      string `json:"name"`
			ID        string `json:"id"`
		} `json:"board"`
		Card struct {
			ShortLink string `json:"shortLink"`
			IDShort   int    `json:"idShort"`
			Name      string `json:"name"`
			ID        string `json:"id"`
		} `json:"card"`
		Text string `json:"text"`
	} `json:"data"`
	Type   string    `json:"type"`
	Date   time.Time `json:"date"`
	Limits struct {
		Reactions struct {
			PerAction struct {
				Status    string `json:"status"`
				DisableAt int    `json:"disableAt"`
				WarnAt    int    `json:"warnAt"`
			} `json:"perAction"`
			UniquePerAction struct {
				Status    string `json:"status"`
				DisableAt int    `json:"disableAt"`
				WarnAt    int    `json:"warnAt"`
			} `json:"uniquePerAction"`
		} `json:"reactions"`
	} `json:"limits"`
	MemberCreator struct {
		ID               string      `json:"id"`
		AvatarHash       string      `json:"avatarHash"`
		AvatarURL        string      `json:"avatarUrl"`
		FullName         string      `json:"fullName"`
		IDMemberReferrer interface{} `json:"idMemberReferrer"`
		Initials         string      `json:"initials"`
		Username         string      `json:"username"`
	} `json:"memberCreator"`
}

// Theme - struct of theme points
type Theme struct {
	ID      string `json:"id"`
	IDBoard string `json:"idBoard"`
	Name    string `json:"name"`
	Color   string `json:"color"`
	Pts     int
}

// CardData struct
type CardData struct {
	CardID    string
	CardName  string
	CardURL   string
	WorkStart time.Time
	PRStart   time.Time
	Heads     []string
}

// Member - data struct
type Member struct {
	ID                       string        `json:"id"`
	AvatarHash               string        `json:"avatarHash"`
	AvatarURL                string        `json:"avatarUrl"`
	Bio                      string        `json:"bio"`
	BioData                  interface{}   `json:"bioData"`
	Confirmed                bool          `json:"confirmed"`
	FullName                 string        `json:"fullName"`
	IDEnterprisesDeactivated []interface{} `json:"idEnterprisesDeactivated"`
	IDPremOrgsAdmin          []interface{} `json:"idPremOrgsAdmin"`
	Initials                 string        `json:"initials"`
	MemberType               string        `json:"memberType"`
	Products                 []interface{} `json:"products"`
	Status                   string        `json:"status"`
	URL                      string        `json:"url"`
	Username                 string        `json:"username"`
	AvatarSource             interface{}   `json:"avatarSource"`
	Email                    interface{}   `json:"email"`
	GravatarHash             interface{}   `json:"gravatarHash"`
	IDBoards                 []string      `json:"idBoards"`
	IDEnterprise             interface{}   `json:"idEnterprise"`
	IDOrganizations          []string      `json:"idOrganizations"`
	IDEnterprisesAdmin       []interface{} `json:"idEnterprisesAdmin"`
	Limits                   struct {
		Boards struct {
			TotalPerMember struct {
				Status    string `json:"status"`
				DisableAt int    `json:"disableAt"`
				WarnAt    int    `json:"warnAt"`
			} `json:"totalPerMember"`
		} `json:"boards"`
		Orgs struct {
			TotalPerMember struct {
				Status    string `json:"status"`
				DisableAt int    `json:"disableAt"`
				WarnAt    int    `json:"warnAt"`
			} `json:"totalPerMember"`
		} `json:"orgs"`
	} `json:"limits"`
	LoginTypes               interface{}   `json:"loginTypes"`
	MarketingOptIn           interface{}   `json:"marketingOptIn"`
	MessagesDismissed        interface{}   `json:"messagesDismissed"`
	OneTimeMessagesDismissed interface{}   `json:"oneTimeMessagesDismissed"`
	Prefs                    interface{}   `json:"prefs"`
	Trophies                 []interface{} `json:"trophies"`
	UploadedAvatarHash       interface{}   `json:"uploadedAvatarHash"`
	UploadedAvatarURL        interface{}   `json:"uploadedAvatarUrl"`
	PremiumFeatures          []interface{} `json:"premiumFeatures"`
	IDBoardsPinned           interface{}   `json:"idBoardsPinned"`
}

// CardComment - struct for managing trello card comments
type CardComment []struct {
	ID              string `json:"id"`
	IDMemberCreator string `json:"idMemberCreator"`
	Data            struct {
		List struct {
			Name string `json:"name"`
			ID   string `json:"id"`
		} `json:"list"`
		Board struct {
			ShortLink string `json:"shortLink"`
			Name      string `json:"name"`
			ID        string `json:"id"`
		} `json:"board"`
		Card struct {
			ShortLink string `json:"shortLink"`
			IDShort   int    `json:"idShort"`
			Name      string `json:"name"`
			ID        string `json:"id"`
		} `json:"card"`
		Text string `json:"text"`
	} `json:"data"`
	Type   string    `json:"type"`
	Date   time.Time `json:"date"`
	Limits struct {
		Reactions struct {
			PerAction struct {
				Status    string `json:"status"`
				DisableAt int    `json:"disableAt"`
				WarnAt    int    `json:"warnAt"`
			} `json:"perAction"`
			UniquePerAction struct {
				Status    string `json:"status"`
				DisableAt int    `json:"disableAt"`
				WarnAt    int    `json:"warnAt"`
			} `json:"uniquePerAction"`
		} `json:"reactions"`
	} `json:"limits"`
	MemberCreator struct {
		ID         string `json:"id"`
		AvatarHash string `json:"avatarHash"`
		AvatarURL  string `json:"avatarUrl"`
		FullName   string `json:"fullName"`
		Initials   string `json:"initials"`
		Username   string `json:"username"`
	} `json:"memberCreator"`
}

// ListData - struct of board lists data
type ListData []struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Closed     bool   `json:"closed"`
	IDBoard    string `json:"idBoard"`
	Pos        int    `json:"pos"`
	Subscribed bool   `json:"subscribed"`
}

// Themes - array of theme structs
type Themes []Theme

// AddBoardMember - Add a trello member to a board
func AddBoardMember(wOpts *WallConf, boardID string, memberID string) error {
	url := "https://api.trello.com/1/boards/" + boardID + "/members/" + memberID + "?type=normal&key=" + wOpts.Walle.Tkey + "&token=" + wOpts.Walle.Ttoken

	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		errTrap(wOpts, "Error in http.NewRequest in `AddBoardMember` in `trello.go`", err)
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errTrap(wOpts, "Error in client.Do in `AddBoardMember` in `trello.go`", err)
		return err
	}
	defer resp.Body.Close()

	return err
}

// CreateList - adlio doesn't have this function so here it is
func CreateList(boardID string, listName string, wOpts *WallConf) error {
	url := "https://api.trello.com/1/lists/"
	var jsonStr = []byte(`{
		"name":"` + listName + `",
		"idBoard":"` + boardID + `",
		"key":"` + wOpts.Walle.Tkey + `",
		"token":"` + wOpts.Walle.Ttoken + `"
		}`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		errTrap(wOpts, "Error in http.NewRequest in `CreateList` in `trello.go`", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errTrap(wOpts, "Error in client.Do in `CreateList` in `trello.go`", err)
		return err
	}
	defer resp.Body.Close()
	return err
}

// CreateCard - custom card creation
func CreateCard(cardName string, listID string, wOpts *WallConf) error {
	url := "https://api.trello.com/1/cards"
	var jsonStr = []byte(`{
		"name":"` + cardName + `",
		"idList":"` + listID + `",
		"key":"` + wOpts.Walle.Tkey + `",
		"token":"` + wOpts.Walle.Ttoken + `"
		}`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		errTrap(wOpts, "Error in http.NewRequest in `CreateCard` in `trello.go`", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errTrap(wOpts, "Error in client.Do in `CreateList` in `trello.go`", err)
		return err
	}
	defer resp.Body.Close()
	return err
}

// CreateBoard - adlio doesn't have this function so here it is
func CreateBoard(boardName string, orgName string, wOpts *WallConf) (trellrep Boards, err error) {

	url := "https://api.trello.com/1/boards/"

	var jsonStr = []byte(`{
		"name":"` + boardName + `",
		"defaultLabels":"true",
		"defaultLists":"false",
		"idOrganization":"` + orgName + `",
		"keepFromSource":"none",
		"powerUps":"voting",
		"prefs_permissionLevel":"org",
		"prefs_voting":"members",
		"prefs_comments":"members",
		"prefs_invitations":"members",
		"prefs_selfJoin":"true",
		"prefs_cardCovers":"true",
		"prefs_background":"green",
		"prefs_cardAging":"regular",
		"key":"` + wOpts.Walle.Tkey + `",
		"token":"` + wOpts.Walle.Ttoken + `"
		}`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		errTrap(wOpts, "Error in http.NewRequest in `CreateBoard` in `trello.go`", err)
		return trellrep, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errTrap(wOpts, "Error in client.Do in `CreateBoard` in `trello.go`", err)
		return trellrep, err
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&trellrep)

	return trellrep, err

}

//AssignCollection - Assign a board to a speific collection (via its ID)
func AssignCollection(boardID string, collectionID string, wOpts *WallConf) string {
	url := "https://api.trello.com/1/boards/" + boardID + "/idTags?value=" + collectionID + "&key=" + wOpts.Walle.Tkey + "&token=" + wOpts.Walle.Ttoken

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		errTrap(wOpts, "Error in http.NewRequest in `AssignCollection` in `trello.go`", err)
		return "Error see logs"
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errTrap(wOpts, "Error in client.Do in `AssignCollection` in `trello.go`", err)
		return "Error see logs"
	}
	defer resp.Body.Close()

	return "Board was assigned to Collection " + collectionID
}

// GetPowerUpField - adlio/trello doesn't support Powerup/Plugin card fields, so buildling this function to add that
func GetPowerUpField(cardID string, wOpts *WallConf) (pluginCard PluginCollection, err error) {
	url := "https://api.trello.com/1/cards/" + cardID + "/plugindata?key=" + wOpts.Walle.Tkey + "&token=" + wOpts.Walle.Ttoken

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		errTrap(wOpts, "Error in http.NewRequest in `GetPowerUpField` in `trello.go`", err)
		return pluginCard, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errTrap(wOpts, "Error in client.Do in `GetPowerUpField` in `trello.go`", err)
		return pluginCard, err
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&pluginCard)

	return pluginCard, err
}

// adlia/trello doesn't support label removal, so this function does that thing
func removeLabel(cardID string, labelID string, wOpts *WallConf) (err error) {
	url := "https://api.trello.com/1/cards/" + cardID + "/idLabels/" + labelID

	var jsonStr = []byte(`{"key": "` + wOpts.Walle.Tkey + `", "token": "` + wOpts.Walle.Ttoken + `"}`)
	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		errTrap(wOpts, "Error in http.NewRequest in `removeLabel` in `trello.go`", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errTrap(wOpts, "Error in client.Do in `removeLabel` in `trello.go`", err)
		return err
	}
	defer resp.Body.Close()

	return err
}

// PutCustomField - used in multiple places
// adlio/trello doesn't support custom card fields, so this function posts to changes in custom fields
func PutCustomField(cardID string, customID string, wOpts *WallConf, someValueType string, somevalue string) (err error) {
	url := "https://api.trello.com/1/card/" + cardID + "/customField/" + customID + "/item"

	var jsonStr = []byte(`{"key": "` + wOpts.Walle.Tkey + `", "token": "` + wOpts.Walle.Ttoken + `", "value": { "` + someValueType + `": "` + somevalue + `" }}`)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		errTrap(wOpts, "Error in http.NewRequest in `PutCustomField` in `trello.go`", err)
		return err
	}
	req.Header.Set("X-Custom-Header", "")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errTrap(wOpts, "Error in client.Do in `PutCustomField` in `trello.go`", err)
		return err
	}
	defer resp.Body.Close()

	return err
}

// RemoveHead - Remove member from trello card
func RemoveHead(wOpts *WallConf, cardID string, memberID string) error {

	url := "https://api.trello.com/1/cards/" + cardID + "/idMembers/" + memberID

	var jsonStr = []byte(`{"key": "` + wOpts.Walle.Tkey + `", "token": "` + wOpts.Walle.Ttoken + `"}`)
	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		errTrap(wOpts, "Error in http.NewRequest in `RemoveHead` in `trello.go`", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errTrap(wOpts, "Error in client.Do in `RemoveHead` in `trello.go`", err)
		return err
	}
	defer resp.Body.Close()

	return err

}

// GetCardListHistory - return list history of a card
func GetCardListHistory(cardID string, wOpts *WallConf) (cardListHistory CardListHistory) {

	url := "https://api.trello.com/1/cards/" + cardID + "/actions?filter=updateCard:idList&key=" + wOpts.Walle.Tkey + "&token=" + wOpts.Walle.Ttoken

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		errTrap(wOpts, "Error in http.NewRequest in `GetTimePutList` in `trello.go`", err)
		return cardListHistory
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errTrap(wOpts, "Error in client.Do in `GetTimePutList` in `trello.go`", err)
		return cardListHistory
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&cardListHistory)

	return cardListHistory
}

// SkipColumn - check if a card was ever in a specific column
func SkipColumn(wOpts *WallConf, skippedColumn string, cardID string) (skipped bool) {

	cardListHistory := GetCardListHistory(cardID, wOpts)

	for _, h := range cardListHistory {
		if h.Data.ListBefore.ID == skippedColumn || h.Data.ListAfter.ID == skippedColumn {
			return false
		}
	}

	return false
}

// GetTimePutList - Get datetime card was last put in a given list
func GetTimePutList(listID string, cardID string, opts Config, wOpts *WallConf) (found bool, cardListTime time.Time) {

	var cardListHistory CardListHistory

	url := "https://api.trello.com/1/cards/" + cardID + "/actions?filter=updateCard:idList&key=" + wOpts.Walle.Tkey + "&token=" + wOpts.Walle.Ttoken

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		errTrap(wOpts, "Error in http.NewRequest in `GetTimePutList` in `trello.go`", err)
		return false, cardListTime
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errTrap(wOpts, "Error in client.Do in `GetTimePutList` in `trello.go`", err)
		return false, cardListTime
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&cardListHistory)

	for h := range cardListHistory {
		if cardListHistory[h].Data.ListAfter.ID == listID {
			cardListTime = cardListHistory[h].Date

			return true, cardListTime
		}
	}

	return false, cardListTime
}

// DupeTrelloBoard - Duplicate an entire trello board and assign it to the Dupe Collection
func DupeTrelloBoard(boardID string, newName string, trelloOrg string, wOpts *WallConf) (output string, err error) {

	var newBoard Boards

	url := "https://api.trello.com/1/boards"
	var jsonStr = []byte(`{
		"name":"` + newName + `",
		"idBoardSource":"` + boardID + `",
		"idOrganization":"` + trelloOrg + `",
		"prefs_permissionLevel":"org",
		"keepFromSource":"cards",
		"prefs_background":"red",
		"key":"` + wOpts.Walle.Tkey + `",
		"token":"` + wOpts.Walle.Ttoken + `"
		}`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		errTrap(wOpts, "Error in http.NewRequest in `DupeTrelloBoard` in `trello.go`", err)
		return "Error in http.NewRequest in `DupeTrelloBoard` in `trello.go`", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errTrap(wOpts, "Error in client.Do in `DupeTrelloBoard` in `trello.go`", err)
		return "Error in client.Do in `DupeTrelloBoard` in `trello.go`", err
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&newBoard)
	if err != nil {
		errTrap(wOpts, "Error in json.NewDecoder in `DupeTrelloBoard` in `trello.go`", err)
		return "Error in json.NewDecoder in `DupeTrelloBoard` in `trello.go`", err
	}

	message := "Board duplicated to <" + newBoard.ShortURL + "|" + newName + "> ID#: `" + newBoard.ID + "`\n"

	if wOpts.Walle.DupeCollectionID != "" {
		output := AssignCollection(newBoard.ID, wOpts.Walle.DupeCollectionID, wOpts)
		if wOpts.Walle.LogToSlack {
			var attachments Attachment
			attachments.Text = ""
			attachments.Color = ""
			LogToSlack(output, wOpts, attachments)
		}
		message = message + output
	}

	return message, nil
}

// RetrieveAll - Get all board data (cards/lists/etc)
// whichCards ==  none / all / closed / open / visible per https://developers.trello.com/reference#cards-nested-resource
func RetrieveAll(wOpts *WallConf, boardID string, whichCards string) (allTheThings BoardData, err error) {

	// handle invalid card types
	switch whichCards {
	case "none":
		whichCards = "none"
	case "closed":
		whichCards = "closed"
	case "open":
		whichCards = "open"
	case "visible":
		whichCards = "visible"
	default:
		whichCards = "all"
	}

	url := "https://api.trello.com/1/boards/" + boardID + "/?card_customFieldItems=true&cards=" + whichCards + "&key=" + wOpts.Walle.Tkey + "&token=" + wOpts.Walle.Ttoken

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		errTrap(wOpts, "ERROR in RetrieveAll in `nested.go` *GET*", err)
		return allTheThings, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errTrap(wOpts, "ERROR in RetrieveAll in `nested.go` client.Do(req)", err)
		return allTheThings, err
	}
	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(&allTheThings)

	return allTheThings, nil
}

// GetLabel - Get labels on a board
func GetLabel(wOpts *WallConf, boardID string) (allThemes Themes, err error) {
	url := "https://api.trello.com/1/board/" + boardID + "/labels?key=" + wOpts.Walle.Tkey + "&token=" + wOpts.Walle.Ttoken

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		errTrap(wOpts, "Error in http.NewRequest in `GetLabel` in `trello.go`", err)
		return allThemes, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errTrap(wOpts, "Error in client.Do in `GetLabel` in `trello.go`", err)
		return allThemes, err
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&allThemes)

	return allThemes, nil

}

// GetDescHistory - retrieve card description history
func GetDescHistory(wOpts *WallConf, cardID string) (descHistory CardDescHistory, err error) {
	url := "https://trello.com/1/cards/" + cardID + "/actions?filter=updateCard:desc&key=" + wOpts.Walle.Tkey + "&token=" + wOpts.Walle.Ttoken

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		errTrap(wOpts, "Error in http.NewRequest in `GetDescHistory` in `trello.go`", err)
		return descHistory, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errTrap(wOpts, "Error in client.Do in `GetDescHistory` in `trello.go`", err)
		return descHistory, err
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&descHistory)

	return descHistory, err
}

// GetAttachments - get trello card attachment data
func GetAttachments(wOpts *WallConf, cardID string) (cAttach []CardAttachment, err error) {

	url := "https://trello.com/1/cards/" + cardID + "/attachments?key=" + wOpts.Walle.Tkey + "&token=" + wOpts.Walle.Ttoken

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		errTrap(wOpts, "Error in http.NewRequest in `GetAttachments` in `trello.go`", err)
		return cAttach, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errTrap(wOpts, "Error in client.Do in `GetAttachments` in `trello.go`", err)
		return cAttach, err
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&cAttach)

	return cAttach, nil
}

// GetWBoards - search and retrieve all Trello boards in organization that have {W} in the title.  used for alerting on retro items
func GetWBoards(wOpts *WallConf) (retroArray []RetroStruct, err error) {

	type tempID []struct {
		Name string `json:"name"`
		ID   string `json:"id"`
	}

	var holdID tempID
	var tempArray RetroStruct

	url := "https://api.trello.com/1/organizations/" + wOpts.Walle.TrelloOrgID + "/boards?filter=open&fields=id%2Cname&key=" + wOpts.Walle.Tkey + "&token=" + wOpts.Walle.Ttoken

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		errTrap(wOpts, "Error in http.NewRequest in `GetWBoards` in `trello.go`", err)
		return retroArray, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errTrap(wOpts, "Error in client.Do in `GetWBoards` in `trello.go`", err)
		return retroArray, err
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&holdID)

	for _, h := range holdID {
		if strings.Contains(h.Name, "{W}") {
			tempArray.RetroID = h.ID
			tempArray.TeamID = "{W}"
			retroArray = append(retroArray, tempArray)
		}
	}

	return retroArray, nil

}

// GetMemberInfo - get members data
func GetMemberInfo(headID string, wOpts *WallConf) (fullname string, avatarhash string, userName string) {
	var memberData Member

	url := "https://api.trello.com/1/member/" + headID + "/?key=" + wOpts.Walle.Tkey + "&token=" + wOpts.Walle.Ttoken

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		errTrap(wOpts, "Error in http.NewRequest in `GetMemberInfo` in `trello.go`", err)
		fmt.Println(err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errTrap(wOpts, "Error in client.Do in `GetMemberInfo` in `trello.go`", err)
		fmt.Println(err)
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&memberData)

	return memberData.FullName, memberData.AvatarHash, memberData.Username
}

// CommentCard - add a comment to a card
func CommentCard(cardID string, comment string, wOpts *WallConf) error {
	url := "https://api.trello.com/1/cards/" + cardID + "/actions/comments"
	var jsonStr = []byte(`{
		"text":"` + comment + `",
		"key":"` + wOpts.Walle.Tkey + `",
		"token":"` + wOpts.Walle.Ttoken + `"
		}`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		errTrap(wOpts, "Error in http.NewRequest in `CommentCard` in `trello.go`", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errTrap(wOpts, "Error in client.Do in `CommentCard` in `trello.go`", err)
		return err
	}
	defer resp.Body.Close()

	return nil
}

// GetCardComments - retrieve all comments on a given trello card
func GetCardComments(cardID string, wOpts *WallConf) (cardComments CardComment, err error) {

	url := "https://api.trello.com/1/cards/" + cardID + "/actions?filter=commentCard&key=" + wOpts.Walle.Tkey + "&token=" + wOpts.Walle.Ttoken

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		errTrap(wOpts, "Error in http.NewRequest in `GetCardComments` in `trello.go`", err)
		return cardComments, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errTrap(wOpts, "Error in client.Do in `GetCardComments` in `trello.go`", err)
		return cardComments, err
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&cardComments)

	return cardComments, nil

}

// GetLists - Get all lists data in a board
func GetLists(wOpts *WallConf, boardID string) (listData ListData, err error) {

	url := "https://api.trello.com/1/boards/" + boardID + "/lists?key=" + wOpts.Walle.Tkey + "&token=" + wOpts.Walle.Ttoken

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		errTrap(wOpts, "ERROR in GetListID in `trello.go` *GET*", err)
		return listData, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errTrap(wOpts, "ERROR in GetListID in `trello.go` client.Do(req)", err)
		return listData, err
	}
	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(&listData)

	return listData, nil
}

// CardPosition - change card position
func CardPosition(wOpts *WallConf, cardID string, position string) error {
	url := "https://api.trello.com/1/cards/" + cardID
	var jsonStr = []byte(`{
		"pos":"` + position + `",
		"key":"` + wOpts.Walle.Tkey + `",
		"token":"` + wOpts.Walle.Ttoken + `"
		}`)

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		errTrap(wOpts, "Error in http.NewRequest in `CardPosition` in `trello.go`", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errTrap(wOpts, "Error in client.Do in `CardPosition` in `trello.go`", err)
		return err
	}
	defer resp.Body.Close()

	return nil
}

// MoveCardList - Move a card to a different list
func MoveCardList(wOpts *WallConf, cardID string, newList string) error {
	url := "https://api.trello.com/1/cards/" + cardID
	var jsonStr = []byte(`{
		"idList":"` + newList + `",
		"key":"` + wOpts.Walle.Tkey + `",
		"token":"` + wOpts.Walle.Ttoken + `"
		}`)

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		errTrap(wOpts, "Error in http.NewRequest in `MoveCardList` in `trello.go`", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errTrap(wOpts, "Error in client.Do in `MoveCardList` in `trello.go`", err)
		return err
	}
	defer resp.Body.Close()

	return nil
}

// ReOrderCardInList - Change placement of a card in a list
// newPos == "top", "bottom" or positive float
func ReOrderCardInList(wOpts *WallConf, cardID string, newPos string) error {
	url := "https://api.trello.com/1/cards/" + cardID
	var jsonStr = []byte(`{
		"pos":"` + newPos + `", 
		"key":"` + wOpts.Walle.Tkey + `",
		"token":"` + wOpts.Walle.Ttoken + `"
		}`)

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		errTrap(wOpts, "Error in http.NewRequest in `ReOrderCardInList` in `trello.go`", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errTrap(wOpts, "Error in client.Do in `ReOrderCardInList` in `trello.go`", err)
		return err
	}
	defer resp.Body.Close()

	return nil
}

// GetCardAction - retrieve card actions
func GetCardAction(wOpts *WallConf, cardID string, limit int) (actions CardAction, err error) {
	url := "https://trello.com/1/cards/" + cardID + "/actions?filter=all&limit=" + strconv.Itoa(limit) + "&key=" + wOpts.Walle.Tkey + "&token=" + wOpts.Walle.Ttoken

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		errTrap(wOpts, "Error in http.NewRequest in `GetDescHistory` in `trello.go`", err)
		return actions, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errTrap(wOpts, "Error in client.Do in `GetDescHistory` in `trello.go`", err)
		return actions, err
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&actions)

	return actions, err
}

// GetCreateDate - Retrieve creation date of a trello card
func GetCreateDate(wOpts *WallConf, cardID string) (createDate time.Time, err error) {
	if cardID == "" {
		return time.Time{}, nil
	}

	ts, err := strconv.ParseUint(cardID[:8], 16, 64)
	if err != nil {
		errTrap(wOpts, "ID `"+cardID+"` failed to convert to timestamp in `GetCreateDate`.", err)
	} else {
		createDate = time.Unix(int64(ts), 0)
	}

	return createDate, nil
}
