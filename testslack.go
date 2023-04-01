package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/slack-go/slack"
	entitySlack "github.com/tokopedia/captainmarvel/cloud-platform-diary/internal/entity/slack"
	"github.com/tokopedia/tdk/go/log"
	"github.com/tokopedia/tdk/go/sql/sqldb"
)

type SlackRepo struct {
	slack       *slack.Client
	db          *sqldb.DB
	getQuery    []string
	insertQuery []string
}

const (
	getNewRelicIncident = iota
	getNewRelicIncidentByID
	getNewRelicIncidentByMsgTimestamp
	getMessageByTimestamp
	getMessageByIncidentID
)

const (
	insertMessage          = iota
	insertNewRelicIncident = iota
	updateNewRelicIncidentByID
	updateNewRelicIncidentByMsgTimestamp
	updateNewRelicIncidentStatusByID
	updateMessageByTimestamp
)

func New(db *sqldb.DB, slack *slack.Client) (*SlackRepo, error) {
	getQuery := []string{
		getNewRelicIncident:               "SELECT slack.message_timestamp, incident.incident_id, name, url, description, owner, generated_by, status, severity, condition_id, labels, root_cause, channel, start_time, recover_time FROM incident INNER JOIN slack ON slack.incident_id = incident.incident_id WHERE incident.incident_id = $1 AND incident.channel = $2",
		getNewRelicIncidentByID:           "SELECT incident_id, name, url, description, owner, generated_by, status, severity, condition_id, labels, root_cause, channel, start_time, recover_time FROM incident WHERE incident_id = $1 AND channel = $2",
		getNewRelicIncidentByMsgTimestamp: "SELECT slack.message_timestamp, incident.incident_id, name, url, description, owner, generated_by, status, severity, condition_id, labels, root_cause, channel, start_time, recover_time FROM incident INNER JOIN slack ON slack.incident_id = incident.incident_id WHERE incident.incident_id = $1 AND slack.message_timestamp = $2",
		getMessageByTimestamp:             "SELECT slack.id, trigger_id, workspace, user_ack, message_timestamp, incident.incident_id FROM slack INNER JOIN incident ON incident.incident_id = slack.incident_id WHERE slack.message_timestamp = $1 AND incident.channel = $2",
		getMessageByIncidentID:            "SELECT trigger_id, workspace, user_ack, message_timestamp, slack.incident_id FROM slack INNER JOIN incident ON incident.id = slack.incident_id WHERE incident.incident_id = $1 AND incident.channel = $2",
	}

	insertQuery := []string{
		insertMessage:                        "INSERT INTO slack (trigger_id, workspace, user_ack, message_timestamp, incident_id) VALUES ($1, $2, $3, $4, $5)",
		insertNewRelicIncident:               "INSERT INTO incident (incident_id, name, url, description, owner, generated_by, status, severity, condition_id, labels, root_cause, channel, start_time, recover_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)",
		updateNewRelicIncidentByID:           "UPDATE incident SET root_cause = $1 WHERE incident_id = $2",
		updateNewRelicIncidentByMsgTimestamp: "UPDATE incident SET root_cause = $1 WHERE incident_id = $2",
		updateNewRelicIncidentStatusByID:     "UPDATE incident SET status = $1, recover_time = $2 FROM slack WHERE incident.incident_id = $3 AND incident.channel = $4 AND slack.message_timestamp = $5",
		updateMessageByTimestamp:             "UPDATE slack SET trigger_id = $1, workspace = $2, user_ack = $3 FROM incident WHERE message_timestamp = $4 AND incident.channel = $5",
	}

	return &SlackRepo{
		slack:       slack,
		db:          db,
		getQuery:    getQuery,
		insertQuery: insertQuery,
	}, nil
}

func (sr *SlackRepo) GetNewRelicIncident(ctx context.Context, incidentID int, channelID string) (entitySlack.Incident, error) {
	incident := entitySlack.Incident{}
	err := sr.db.GetContext(ctx, &incident, sr.getQuery[getNewRelicIncident], incidentID, channelID)

	return incident, err
}

func (sr *SlackRepo) GetNewRelicIncidentByID(ctx context.Context, incidentID int, channelID string) (entitySlack.Incident, error) {
	incident := entitySlack.Incident{}
	err := sr.db.GetContext(ctx, &incident, sr.getQuery[getNewRelicIncidentByID], incidentID, channelID)

	return incident, err
}

func (sr *SlackRepo) GetNewRelicIncidentByMsgTimestamp(ctx context.Context, incidentID int, messageTs string) (entitySlack.Incident, error) {
	incident := entitySlack.Incident{}
	err := sr.db.GetContext(ctx, &incident, sr.getQuery[getNewRelicIncidentByMsgTimestamp], incidentID, messageTs)

	return incident, err
}

func (sr *SlackRepo) GetMessageByTimestamp(ctx context.Context, messageTimestamp, channel string) (entitySlack.Slack, error) {
	slackMessage := entitySlack.Slack{}
	err := sr.db.GetContext(ctx, &slackMessage, sr.getQuery[getMessageByTimestamp], messageTimestamp, channel)

	return slackMessage, err
}

func (sr *SlackRepo) GetMessageByIncidentID(ctx context.Context, incidentID int, channel string) (entitySlack.Slack, error) {
	slackMessage := entitySlack.Slack{}
	err := sr.db.GetContext(ctx, &slackMessage, sr.getQuery[getMessageByIncidentID], incidentID, channel)

	return slackMessage, err
}

// InsertMessage stores Slack message/alert for future reference.
// We need to track alerts/messages that were sent, e.g. to reply updates in its thread or to update the message body.
func (sr *SlackRepo) InsertMessage(ctx context.Context, triggerID, workspace, userACK, messageTimestamp string, incidentID int) error {
	_, err := sr.db.ExecContext(ctx, sr.insertQuery[insertMessage], triggerID, workspace, userACK, messageTimestamp, incidentID)
	if err != nil {
		log.Errorf("Error insert slack message: %s", err)
		return err
	}
	return nil
}

func (sr *SlackRepo) InsertNewRelicIncident(ctx context.Context, ID, conditionID int, name, url, description, owner, generatedBy, status, severity, rootCause, channel, labels string, startTime, recoverTime time.Time) error {
	_, err := sr.db.ExecContext(ctx, sr.insertQuery[insertNewRelicIncident], ID, name, url, description, owner, generatedBy, status, severity, conditionID, labels, rootCause, channel, startTime, recoverTime)
	if err != nil {
		log.Errorf("Error insert incident root cause: %s", err)
		return err
	}

	return nil
}

func (sr *SlackRepo) UpdateMessageByTimestamp(ctx context.Context, triggerID, workspace, userACK, messageTimestamp, channel string) error {
	_, err := sr.db.ExecContext(ctx, sr.insertQuery[updateMessageByTimestamp], triggerID, workspace, userACK, messageTimestamp, channel)
	if err != nil {
		log.Errorf("Error update slack message by timestamp: %s", err)
		return err
	}
	return nil
}

func (sr *SlackRepo) UpdateNewRelicIncidentByID(ctx context.Context, rootCause string, incidentID int) error {
	_, err := sr.db.ExecContext(ctx, sr.insertQuery[updateNewRelicIncidentByID], rootCause, incidentID)
	if err != nil {
		log.Errorf("Error update incident status: %s", err)
		return err
	}
	return nil
}

func (sr *SlackRepo) UpdateNewRelicIncidentStatusByID(ctx context.Context, status, messageTimestamp, channelID string, incidentID int, recoverTime time.Time) error {
	_, err := sr.db.ExecContext(ctx, sr.insertQuery[updateNewRelicIncidentStatusByID], status, recoverTime, incidentID, channelID, messageTimestamp)
	if err != nil {
		log.Errorf("Error update incident: %s", err)
		return err
	}
	return nil
}

func (sr *SlackRepo) ReplyMessageInThread(ctx context.Context, channelID, text, color, ts, url string) (string, string, error) {
	var msgBlocks []slack.Block
	title := ""
	body := text
	temp := strings.Split(text, "\n")

	// this will separate title with body message
	if len(temp) > 0 {
		title = temp[0]
		body = strings.Join(temp[1:], "\n")
	}

	// Append FieldBlock
	textBlockObj := slack.NewTextBlockObject("mrkdwn", body, false, false)
	msgBlocks = append(msgBlocks, slack.NewSectionBlock(textBlockObj, nil, nil))

	msg := slack.NewBlockMessage(msgBlocks...)
	b, err := json.Marshal(&msg)
	if err != nil {
		log.Errorf(err.Error())
		return "", "", err
	}

	m := slack.Message{}
	if err := json.Unmarshal(b, &m); err != nil {
		return "", "", err
	}

	attach := slack.Attachment{
		Color:     "FFFFFF",
		Title:     title,
		TitleLink: url,
	}

	block := slack.Attachment{
		Color: color,
		Blocks: slack.Blocks{
			BlockSet: m.Blocks.BlockSet,
		},
	}

	// Post Parameters
	param := slack.PostMessageParameters{
		ThreadTimestamp: ts,
	}

	attachment := slack.MsgOptionAttachments(attach, block)
	postParams := slack.MsgOptionPostMessageParameters(param)

	// channel, ts, _, err := sr.slack.SendMessage(channelID, attachment)
	channel, ts, err := sr.slack.PostMessage(channelID, attachment, postParams)
	if err != nil {
		return "", "", err
	}

	return channel, ts, err
}

func (sr *SlackRepo) SendMessage(ctx context.Context, channelID, text, color, ts, vendor, url string) (string, string, error) {
	var (
		msgBlocks []slack.Block
	)
	title := ""
	body := text
	temp := strings.Split(text, "\n")

	// this will separate title with body message
	if len(temp) > 0 {
		title = temp[0]
		body = strings.Join(temp[1:], "\n")
	}

	// TitleBlock
	// titleBlockObj := slack.NewTextBlockObject("mrkdwn", title, false, false)
	// titleBlocks = append(titleBlocks, slack.NewHeaderBlock(titleBlockObj))

	// Append FieldBlock
	textBlockObj := slack.NewTextBlockObject("mrkdwn", body, false, false)
	// msgBlocks = append(msgBlocks, slack.NewSectionBlock(textBlockObj, nil, nil))

	// Append ActionBlock
	ackBtnTxt := slack.NewTextBlockObject("plain_text", "Acknowledge", false, false)
	ackBtn := slack.NewButtonBlockElement("reason_btn", "", ackBtnTxt)
	ignoreBtnTxt := slack.NewTextBlockObject("plain_text", "Ignore", false, false)
	ignoreBtn := slack.NewButtonBlockElement("ignore_btn", "ignored", ignoreBtnTxt)
	// msgBlocks = append(msgBlocks, slack.NewActionBlock("", ackBtn, ignoreBtn))

	switch vendor {
	case "New Relic":
		msgBlocks = append(msgBlocks, slack.NewSectionBlock(textBlockObj, nil, nil))
		msgBlocks = append(msgBlocks, slack.NewActionBlock("", ackBtn, ignoreBtn))
	case "Google Cloud Platform":
		msgBlocks = append(msgBlocks, slack.NewSectionBlock(textBlockObj, nil, nil))
	default:
		msgBlocks = append(msgBlocks, slack.NewSectionBlock(textBlockObj, nil, nil))
	}

	msg := slack.NewBlockMessage(msgBlocks...)

	// Message Block
	b, err := json.Marshal(&msg)
	if err != nil {
		log.Errorf(err.Error())
		return "", "", err
	}

	m := slack.Message{}
	if err := json.Unmarshal(b, &m); err != nil {
		return "", "", err
	}

	// Attachment Parameters
	attach := slack.Attachment{
		Color:     "FFFFFF",
		Title:     title,
		TitleLink: url,
	}

	block := slack.Attachment{
		Color: color,
		Blocks: slack.Blocks{
			BlockSet: m.Blocks.BlockSet,
		},
	}

	// Post Parameters
	param := slack.PostMessageParameters{
		ThreadTimestamp: ts,
	}

	attachment := slack.MsgOptionAttachments(attach, block)
	postParams := slack.MsgOptionPostMessageParameters(param)

	channel, ts, err := sr.slack.PostMessage(channelID, attachment, postParams)
	if err != nil {
		return "", "", err
	}

	return channel, ts, err
}

func (sr *SlackRepo) UpdateMessage(ctx context.Context, channelID, text, color, ts, vendor, url string) (string, string, error) {
	var msgBlocks []slack.Block
	title := ""
	body := text
	temp := strings.Split(text, "\n")

	// this will separate title with body message
	if len(temp) > 0 {
		title = temp[0]
		body = strings.Join(temp[1:], "\n")
	}

	// Append FieldBlock
	textBlockObj := slack.NewTextBlockObject("mrkdwn", body, false, false)
	// msgBlocks = append(msgBlocks, slack.NewSectionBlock(textBlockObj, nil, nil))

	// Append ActionBlock
	ackBtnTxt := slack.NewTextBlockObject("plain_text", "Acknowledge", false, false)
	ackBtn := slack.NewButtonBlockElement("reason_btn", "", ackBtnTxt)
	ignoreBtnTxt := slack.NewTextBlockObject("plain_text", "Ignore", false, false)
	ignoreBtn := slack.NewButtonBlockElement("ignore_btn", "ignored", ignoreBtnTxt)
	// msgBlocks = append(msgBlocks, slack.NewActionBlock("", ackBtn, ignoreBtn))

	switch vendor {
	case "New Relic":
		msgBlocks = append(msgBlocks, slack.NewSectionBlock(textBlockObj, nil, nil))
		msgBlocks = append(msgBlocks, slack.NewActionBlock("", ackBtn, ignoreBtn))
	case "Google Cloud Platform":
		msgBlocks = append(msgBlocks, slack.NewSectionBlock(textBlockObj, nil, nil))
	default:
		msgBlocks = append(msgBlocks, slack.NewSectionBlock(textBlockObj, nil, nil))
	}

	msg := slack.NewBlockMessage(msgBlocks...)
	b, err := json.Marshal(&msg)
	if err != nil {
		log.Errorf(err.Error())
		return "", "", err
	}

	m := slack.Message{}
	if err := json.Unmarshal(b, &m); err != nil {
		return "", "", err
	}

	// Attachment Parameters
	attach := slack.Attachment{
		Color:     "FFFFFF",
		Title:     title,
		TitleLink: url,
	}

	block := slack.Attachment{
		Color: color,
		Blocks: slack.Blocks{
			BlockSet: m.Blocks.BlockSet,
		},
	}

	// block := slack.MsgOptionBlocks(m.Blocks.BlockSet...)
	attachment := slack.MsgOptionAttachments(attach, block)

	channel, ts, _, err := sr.slack.UpdateMessage(channelID, ts, attachment)
	if err != nil {
		return "", "", err
	}

	return channel, ts, err
}

func (sr *SlackRepo) SubmitButtonAction(blockActions []*slack.BlockAction, options []string, channelID, tsMessage, triggerID, incidentTitle, incidentMessage, incidentColor, username, url string, replaceMessage bool) (string, error) {
	var actionValue string
	for _, actions := range blockActions {
		switch actions.ActionID {
		case "reason_btn":
			modalRequest := sr.GenerateModalRequest(options)
			modalRequest.CallbackID = "orderModalSubmission"
			_, err := sr.slack.OpenView(triggerID, modalRequest)
			if err != nil {
				log.Errorf("Error opening view : %s", err)
			}
		case "ignore_btn":
			actionValue = actions.Value
			_, err := sr.ReplaceMessage(channelID, tsMessage, actionValue, incidentTitle, incidentMessage, incidentColor, username, url, replaceMessage)
			if err != nil {
				log.Errorf("Error replaced message : %s", err)
				return "", err
			}
		default:
			log.Info("Invalid action : %s", actions.ActionID)
			return "", nil
		}
	}

	return actionValue, nil
}

func (sr *SlackRepo) ReplaceMessage(channelID, tsMessage, actionValue, incidentTitle, incidentMessage, incidentColor, username, url string, replaceMessage bool) (string, error) {
	var msgBlocks []slack.Block
	newMessage := fmt.Sprintf("\n*Action :* \nThe user `%s` has *acknowledged* with `%s`", username, actionValue)

	// Append FieldBlock
	textOriginalBlockObj := slack.NewTextBlockObject("mrkdwn", incidentMessage, false, false)
	textNewBlockObj := slack.NewTextBlockObject("mrkdwn", newMessage, false, false)
	msgBlocks = append(msgBlocks, slack.NewSectionBlock(textOriginalBlockObj, nil, nil))
	msgBlocks = append(msgBlocks, slack.NewSectionBlock(textNewBlockObj, nil, nil))

	respMessage := slack.NewBlockMessage(msgBlocks...)
	respMessage.ReplaceOriginal = replaceMessage

	b, err := json.Marshal(&respMessage)
	if err != nil {
		log.Errorf(err.Error())
		return "", err
	}

	m := slack.Message{}
	if err := json.Unmarshal(b, &m); err != nil {
		return "", err
	}

	// Attachment Parameters
	attach := slack.Attachment{
		Color:     "FFFFFF",
		Title:     incidentTitle,
		TitleLink: url,
	}

	block := slack.Attachment{
		Color: incidentColor,
		Blocks: slack.Blocks{
			BlockSet: m.Blocks.BlockSet,
		},
	}

	attachment := slack.MsgOptionAttachments(attach, block)

	message, _, _, err := sr.slack.UpdateMessage(channelID, tsMessage, attachment)
	if err != nil {
		log.Errorf("Error update message: %s", err)
		return "", err
	}

	return message, err
}

func (sr *SlackRepo) GenerateModalRequest(options []string) slack.ModalViewRequest {
	// Create ModalViewRequest with a header two input
	titleText := slack.NewTextBlockObject("plain_text", "Cloud Platform Diary", false, false)
	closeText := slack.NewTextBlockObject("plain_text", "Close", false, false)
	submitText := slack.NewTextBlockObject("plain_text", "Submit", false, false)

	headerText := slack.NewTextBlockObject("mrkdwn", "Please select the cause of the warning", false, false)
	headerSectionBlock := slack.NewSectionBlock(headerText, nil, nil)

	causeOptions := sr.CreateOptionBlockObject(options)
	causeOptionText := slack.NewTextBlockObject(slack.PlainTextType, "Reason", false, false)
	causeOptionElement := slack.NewOptionsSelectBlockElement(slack.OptTypeStatic, nil, "reason_option", causeOptions...)
	causeOptionBlock := slack.NewInputBlock("reason_option", causeOptionText, causeOptionElement)

	causeText := slack.NewTextBlockObject("plain_text", "Other Reason", false, false)
	causePlaceholder := slack.NewTextBlockObject("plain_text", "Only fill this if you select others in Reason", false, false)
	causeElement := slack.NewPlainTextInputBlockElement(causePlaceholder, "reason_other")
	causeBlock := slack.NewInputBlock("reason_other", causeText, causeElement)
	causeBlock.Optional = true

	blocks := slack.Blocks{
		BlockSet: []slack.Block{
			headerSectionBlock,
			causeOptionBlock,
			causeBlock,
		},
	}

	var modalRequest slack.ModalViewRequest
	modalRequest.Type = slack.ViewType("modal")
	modalRequest.Title = titleText
	modalRequest.Close = closeText
	modalRequest.Submit = submitText
	modalRequest.Blocks = blocks

	return modalRequest
}

func (sr *SlackRepo) CreateOptionBlockObject(options []string) []*slack.OptionBlockObject {
	optionBlockObjects := make([]*slack.OptionBlockObject, 0, len(options))

	for _, opt := range options {
		optionText := slack.NewTextBlockObject(slack.PlainTextType, opt, false, false)
		optionBlockObjects = append(optionBlockObjects, slack.NewOptionBlockObject(opt, optionText, nil))
	}

	otherOptionText := slack.NewTextBlockObject(slack.PlainTextType, "others (provide the reason in the next field)", false, false)
	optionBlockObjects = append(optionBlockObjects, slack.NewOptionBlockObject("others", otherOptionText, nil))

	return optionBlockObjects
}

this is code for create a notfication from newrelic to a slack channel i want to change this code to create notifications to spesific email
