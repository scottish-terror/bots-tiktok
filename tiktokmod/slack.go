package tiktokmod

// handles slack API interface for sending webhooks back with responses

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/parnurzeal/gorequest"
)

const (
	fileUploadURL       string = "https://slack.com/api/files.upload"
	channelCreateURL    string = "https://slack.com/api/channels.create"
	channelArchiveURL   string = "https://slack.com/api/channels.archive"
	channelListURL      string = "https://slack.com/api/conversations.list"
	channelInviteURL    string = "https://slack.com/api/channels.invite"
	channelTopicSetURL  string = "https://slack.com/api/channels.setTopic"
	channelUnArchiveURL string = "https://slack.com/api/channels.unarchive"
)

// Field - struct
type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// BasicSlackPayload - returns most basic slack response payload
type BasicSlackPayload struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

// ChannelListPayload - return payload containing list of slack channels
type ChannelListPayload struct {
	Ok       bool `json:"ok"`
	Channels []struct {
		ID                 string        `json:"id"`
		Name               string        `json:"name"`
		IsChannel          bool          `json:"is_channel"`
		IsGroup            bool          `json:"is_group"`
		IsIm               bool          `json:"is_im"`
		Created            int           `json:"created"`
		Creator            string        `json:"creator"`
		IsArchived         bool          `json:"is_archived"`
		IsGeneral          bool          `json:"is_general"`
		Unlinked           int           `json:"unlinked"`
		NameNormalized     string        `json:"name_normalized"`
		IsShared           bool          `json:"is_shared"`
		IsExtShared        bool          `json:"is_ext_shared"`
		IsOrgShared        bool          `json:"is_org_shared"`
		PendingShared      []interface{} `json:"pending_shared"`
		IsPendingExtShared bool          `json:"is_pending_ext_shared"`
		IsMember           bool          `json:"is_member"`
		IsPrivate          bool          `json:"is_private"`
		IsMpim             bool          `json:"is_mpim"`
		Topic              struct {
			Value   string `json:"value"`
			Creator string `json:"creator"`
			LastSet int    `json:"last_set"`
		} `json:"topic"`
		Purpose struct {
			Value   string `json:"value"`
			Creator string `json:"creator"`
			LastSet int    `json:"last_set"`
		} `json:"purpose"`
		PreviousNames []interface{} `json:"previous_names"`
		NumMembers    int           `json:"num_members"`
	} `json:"channels"`
	ResponseMetadata struct {
		NextCursor string `json:"next_cursor"`
	} `json:"response_metadata"`
	Error string `json:"error"`
}

// ChannelRespPayload - return payload from slack channel creation
type ChannelRespPayload struct {
	Ok      bool `json:"ok"`
	Channel struct {
		ID                 string      `json:"id"`
		Name               string      `json:"name"`
		IsChannel          bool        `json:"is_channel"`
		Created            int         `json:"created"`
		Creator            string      `json:"creator"`
		IsArchived         bool        `json:"is_archived"`
		IsGeneral          bool        `json:"is_general"`
		NameNormalized     string      `json:"name_normalized"`
		IsShared           bool        `json:"is_shared"`
		IsOrgShared        bool        `json:"is_org_shared"`
		IsMember           bool        `json:"is_member"`
		IsPrivate          bool        `json:"is_private"`
		IsMpim             bool        `json:"is_mpim"`
		LastRead           string      `json:"last_read"`
		Latest             interface{} `json:"latest"`
		UnreadCount        int         `json:"unread_count"`
		UnreadCountDisplay int         `json:"unread_count_display"`
		Members            []string    `json:"members"`
		Topic              struct {
			Value   string `json:"value"`
			Creator string `json:"creator"`
			LastSet int    `json:"last_set"`
		} `json:"topic"`
		Purpose struct {
			Value   string `json:"value"`
			Creator string `json:"creator"`
			LastSet int    `json:"last_set"`
		} `json:"purpose"`
		PreviousNames []interface{} `json:"previous_names"`
	} `json:"channel"`
	Error  string `json:"error"`
	Detail string `json:"detail"`
}

// BotDMPayload - struct for bot DMs
type BotDMPayload struct {
	Token          string       `json:"token,omitempty"`
	Channel        string       `json:"channel,omitempty"`
	Text           string       `json:"text,omitempty"`
	AsUser         bool         `json:"as_user,omitempty"`
	Attachments    []Attachment `json:"attachments,omitempty"`
	IconEmoji      string       `json:"icon_emoji,omitempty"`
	IconURL        string       `json:"icon_url,omitempty"`
	LinkNames      bool         `json:"link_names,omitempty"`
	Mkrdwn         bool         `json:"mrkdwn,omitempty"`
	Parse          string       `json:"parse,omitempty"`
	ReplyBroadcast bool         `json:"reply_broadcast,omitempty"`
	ThreadTS       string       `json:"thread_ts,omitempty"`
	UnfurlLinks    bool         `json:"unfurl_links,omitempty"`
	UnfurlMedia    bool         `json:"unfurl_media,omitempty"`
	Username       string       `json:"username,omitempty"`
}

// Attachment - struct
type Attachment struct {
	Fallback   string   `json:"fallback,omitempty"`
	Color      string   `json:"color,omitempty"`
	PreText    string   `json:"pretext,omitempty"`
	AuthorName string   `json:"author_name,omitempty"`
	AuthorLink string   `json:"author_link,omitempty"`
	AuthorIcon string   `json:"author_icon,omitempty"`
	Title      string   `json:"title,omitempty"`
	TitleLink  string   `json:"title_link,omitempty"`
	Text       string   `json:"text,omitempty"`
	ImageURL   string   `json:"image_url,omitempty"`
	Fields     []*Field `json:"fields,omitempty"`
	Footer     string   `json:"footer,omitempty"`
	FooterIcon string   `json:"footer_icon,omitempty"`
	Timestamp  int64    `json:"ts,omitempty"`
	MarkdownIn []string `json:"mrkdwn_in,omitempty"`
}

// Payload - struct
type Payload struct {
	Parse       string       `json:"parse,omitempty"`
	Username    string       `json:"username,omitempty"`
	IconURL     string       `json:"icon_url,omitempty"`
	IconEmoji   string       `json:"icon_emoji,omitempty"`
	Channel     string       `json:"channel,omitempty"`
	Text        string       `json:"text,omitempty"`
	LinkNames   string       `json:"link_names,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
	UnfurlLinks bool         `json:"unfurl_links,omitempty"`
	UnfurlMedia bool         `json:"unfurl_media,omitempty"`
}

// AddField - add fields
func (attachment *Attachment) AddField(field Field) *Attachment {
	attachment.Fields = append(attachment.Fields, &field)
	return attachment
}

func redirectPolicyFunc(req gorequest.Request, via []gorequest.Request) error {
	return fmt.Errorf("Incorrect token (redirection)")
}

// PostSnippet - Post a snippet of any type to slack channel
func PostSnippet(baloo *BalooConf, fileType string, fileContent string, channel string, title string) error {

	form := url.Values{}

	form.Set("token", baloo.Config.SlackToken)
	form.Set("channels", channel)
	form.Set("content", fileContent)
	form.Set("filetype", fileType)
	form.Set("title", title)

	s := form.Encode()

	req, err := http.NewRequest("POST", fileUploadURL, strings.NewReader(s))

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Bearer "+baloo.Config.SlackToken)

	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		errTrap(baloo, "Slack PostSnippet - http.Do() error: ", err)
		return err
	}
	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		errTrap(baloo, "Slack PostSnippet - ioutil.ReadAll() error: ", err)
		return err
	}

	return nil
}

// Send - send message
func Send(webhookURL string, proxy string, payload Payload) []error {
	request := gorequest.New().Proxy(proxy)
	resp, _, err := request.
		Post(webhookURL).
		RedirectPolicy(redirectPolicyFunc).
		Send(payload).
		End()

	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return []error{fmt.Errorf("Error sending msg. Status: %v", resp.Status)}
	}

	return nil
}

// WranglerDM - Send chat.Post API DM messages "as the bot"
func WranglerDM(baloo *BalooConf, payload BotDMPayload) error {
	url := "https://slack.com/api/chat.postMessage"

	payload.Token = baloo.Config.SlackToken
	payload.AsUser = true

	jsonStr, err := json.Marshal(&payload)
	if err != nil {
		errTrap(baloo, "Error attempting to marshal struct to json for slack BotDMPayload", err)
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		errTrap(baloo, "Error in http.NewRequest in `CreateList` in `trello.go`", err)
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+baloo.Config.SlackToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errTrap(baloo, "Error in client.Do in `CreateList` in `trello.go`", err)
		return err
	}
	defer resp.Body.Close()
	return err

}

// Wrangler - wrangle slack calls
func Wrangler(webhookURL string, message string, myChannel string, emojiName string, attachments Attachment) {

	payload := Payload{
		Text:        message,
		Username:    "BalooConf",
		Channel:     myChannel,
		IconEmoji:   emojiName,
		Attachments: []Attachment{attachments},
	}
	err := Send(webhookURL, "", payload)
	if len(err) > 0 {
		fmt.Printf("Slack Messaging Error in Wrangler function in slack.go: %s\n", err)
	}
}

//LogToSlack - Dump Logs to a Slack Channel
func LogToSlack(message string, baloo *BalooConf, attachments Attachment) {
	now := time.Now().Local()
	if baloo.Config.LoggingPrefix != "" {
		message = "`" + baloo.Config.LoggingPrefix + "` - *" + now.Format("01/02/2006 15:04:05") + " :* " + message
	} else {
		message = "*" + now.Format("01/02/2006 15:04:05") + " :* " + message
	}
	Wrangler(baloo.Config.SlackHook, message, baloo.Config.LogChannel, baloo.Config.SlackEmoji, attachments)
}

// CreateChannel - create a slack channel and return the payload
func CreateChannel(baloo *BalooConf, channelName string, errValidate bool) (slackPayload ChannelRespPayload, success bool, err error) {
	var validate string

	form := url.Values{}

	if errValidate {
		validate = "true"
	} else {
		validate = "false"
	}
	form.Set("token", baloo.Config.SlackOAuth)
	form.Set("name", channelName)
	form.Set("validate", validate)

	s := form.Encode()

	req, err := http.NewRequest("POST", channelCreateURL, strings.NewReader(s))

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Bearer "+baloo.Config.SlackToken)

	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		errTrap(baloo, "Slack CreateChannel - http.Do() error: ", err)
		return slackPayload, false, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&slackPayload)

	if slackPayload.Ok {
		return slackPayload, true, nil
	}

	return slackPayload, false, nil
}

// ArchiveChannel - archive a slack channel created by bot and return payload. must send channel slackID, not channel name
func ArchiveChannel(baloo *BalooConf, channelID string) (slackPayload ChannelRespPayload, success bool, err error) {

	form := url.Values{}

	form.Set("token", baloo.Config.SlackOAuth)
	form.Set("channel", channelID)

	s := form.Encode()

	req, err := http.NewRequest("POST", channelArchiveURL, strings.NewReader(s))

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Bearer "+baloo.Config.SlackToken)

	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		errTrap(baloo, "Slack ArchiveChannel - http.Do() error: ", err)
		return slackPayload, false, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&slackPayload)

	if slackPayload.Ok {
		return slackPayload, true, nil
	}

	return slackPayload, false, nil
}

// UnArchiveChannel - un-archive a slack channel and return payload. must send channel slackID, not channel name
func UnArchiveChannel(baloo *BalooConf, channelID string) (slackPayload BasicSlackPayload, err error) {

	form := url.Values{}

	form.Set("token", baloo.Config.SlackOAuth)
	form.Set("channel", channelID)

	s := form.Encode()

	req, err := http.NewRequest("POST", channelUnArchiveURL, strings.NewReader(s))

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Bearer "+baloo.Config.SlackToken)

	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		errTrap(baloo, "Slack UnArchiveChannel - http.Do() error: ", err)
		return slackPayload, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&slackPayload)

	if slackPayload.Ok {
		return slackPayload, nil
	}

	return slackPayload, nil
}

// ChannelList - Return slice of slack channels
func ChannelList(baloo *BalooConf, noArchived bool) (slackPayload ChannelListPayload, success bool, err error) {
	var ignoreArchive string

	form := url.Values{}

	if noArchived {
		ignoreArchive = "true"
	} else {
		ignoreArchive = "false"
	}

	form.Set("token", baloo.Config.SlackToken)
	form.Set("exclude_archived", ignoreArchive)
	form.Set("limit", "1000")
	form.Set("types", "public_channel")

	s := form.Encode()

	req, err := http.NewRequest("POST", channelListURL, strings.NewReader(s))

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Bearer "+baloo.Config.SlackToken)

	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		errTrap(baloo, "Slack ChannelList - http.Do() error: ", err)
		return slackPayload, false, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&slackPayload)

	if slackPayload.Ok {
		return slackPayload, true, nil
	}

	return slackPayload, false, nil
}

// ChannelInvite - Invite a user to a specific channel. Expects slack channel ID and user ID not slack names
func ChannelInvite(baloo *BalooConf, channelID string, userID string) (slackPayload BasicSlackPayload, err error) {

	form := url.Values{}

	form.Set("token", baloo.Config.SlackOAuth)
	form.Set("channel", channelID)
	form.Set("user", userID)

	s := form.Encode()

	req, err := http.NewRequest("POST", channelInviteURL, strings.NewReader(s))

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Bearer "+baloo.Config.SlackToken)

	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		errTrap(baloo, "Slack ChannelInvite - http.Do() error: ", err)
		return slackPayload, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&slackPayload)

	if slackPayload.Ok {
		return slackPayload, nil
	}

	return slackPayload, nil
}

// ChannelTopicSet - Set topic in a channel. Expects slack channel ID not name. Bot MUST be in channel to work
func ChannelTopicSet(baloo *BalooConf, channelID string, topic string) (slackPayload BasicSlackPayload, err error) {

	form := url.Values{}

	form.Set("token", baloo.Config.SlackToken)
	form.Set("channel", channelID)
	form.Set("topic", topic)

	s := form.Encode()

	req, err := http.NewRequest("POST", channelTopicSetURL, strings.NewReader(s))

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Bearer "+baloo.Config.SlackToken)

	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		errTrap(baloo, "Slack ChannelTopicSet - http.Do() error: ", err)
		return slackPayload, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&slackPayload)

	if slackPayload.Ok {
		return slackPayload, nil
	}

	return slackPayload, nil
}
