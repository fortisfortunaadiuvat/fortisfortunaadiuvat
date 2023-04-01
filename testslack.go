package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"time"

	"github.com/slack-go/slack"
	entitySlack "github.com/tokopedia/captainmarvel/cloud-platform-diary/internal/entity/slack"
	"github.com/tokopedia/captainmarvel/cloud-platform-diary/internal/pkg/webhook"
	"github.com/tokopedia/tdk/go/log"
)

type UseCase struct {
	slackRepo slackRepository
}

func New(slack slackRepository) *UseCase {
	return &UseCase{
		slackRepo: slack,
	}
}

func (u *UseCase) ProcessIncident(ctx context.Context, data entitySlack.NewRelicReplyThread) (entitySlack.Incident, error) {
	// Get NewRelic Incident BY incident ID
	incident, _ := u.slackRepo.GetNewRelicIncident(ctx, data.GetIncidentID(), data.GetChannel())
	// if err != nil {
	// 	log.Info("Incident on database not found: %s", err)
	// }

	incidentTs := incident.MessageTimestamp
	if incidentTs == "" {
		// Store Incident
		if err := u.RegisterIncident(ctx, data); err != nil {
			log.Errorf("Failed send incident request: %v", err)
		}

		// Get NewRelic Incident BY incident ID
		i, err := u.slackRepo.GetNewRelicIncidentByID(ctx, data.GetIncidentID(), data.GetChannel())
		if err != nil {
			log.Errorf("Error GET incident on database: %s", err)
		}

		// Send Slack Message
		_, ts, err := u.slackRepo.SendMessage(ctx, data.GetChannel(), u.GetMessageSummary(data, i.StartTime, i.RecoverTime), u.GetColor(data), incidentTs, data.GetVendor(), data.GetURL())
		if err != nil {
			log.Errorf("Failed send slack message to channel %s because: %s", data.GetChannel(), err)
		}

		// Store Slack Message
		triggerID := ""
		workspace := ""
		userACK := ""
		if err := u.slackRepo.InsertMessage(ctx, triggerID, workspace, userACK, ts, data.GetIncidentID()); err != nil {
			log.Errorf("Error store message to database: %s", err)
		}
	} else {
		var recoverTime time.Time
		if data.GetState() == "closed" || data.GetState() == "acknowledged" {
			recoverTime = time.Now().Local()
		} else {
			recoverTime = time.Time{}
		}

		// Store Incident information
		if err := u.slackRepo.UpdateNewRelicIncidentStatusByID(ctx, data.GetState(), incidentTs, data.GetChannel(), data.GetIncidentID(), recoverTime); err != nil {
			log.Errorf("Error store message to database: %s", err)
		}

		// Get NewRelic Incident BY incident ID
		i, err := u.slackRepo.GetNewRelicIncidentByID(ctx, data.GetIncidentID(), data.GetChannel())
		if err != nil {
			log.Errorf("Error GET incident on database: %s", err)
		}

		// Update Slack Message
		_, _, err = u.slackRepo.UpdateMessage(ctx, data.GetChannel(), u.GetMessageSummary(data, i.StartTime, i.RecoverTime), u.GetColor(data), incidentTs, data.GetVendor(), data.GetURL())
		if err != nil {
			log.Errorf("Failed send slack message to channel %s because: %s", data.GetChannel(), err)
		}
	}

	// Get NewRelic Incident BY incident ID
	incidentMetadata, err := u.slackRepo.GetNewRelicIncident(ctx, data.GetIncidentID(), data.GetChannel())
	if err != nil {
		log.Errorf("Error GET incident on database: %s", err)
	}

	// Send Slack Message
	_, _, err = u.slackRepo.ReplyMessageInThread(ctx, data.GetChannel(), u.GetMessage(data), u.GetColor(data), incidentMetadata.MessageTimestamp, data.GetURL())
	if err != nil {
		log.Errorf("Failed send slack message to channel %s because: %s", data.GetChannel(), err)
	}

	return incidentMetadata, nil
}

func (u *UseCase) RegisterIncident(ctx context.Context, data entitySlack.NewRelicReplyThread) error {
	dt := time.Now()

	incidentID := data.GetIncidentID()
	incidentName := data.GetIncidentName()
	incidentURL := data.GetURL()
	incidentDescription := data.GetBody()
	incidentOwner := data.GetOwner()
	incidentGeneratedBy := data.GetVendor()
	incidentStatus := data.GetState()
	incidentSeverity := data.GetSeverity()
	incidentConditionID := data.GetConditionID()
	incidentLabels := data.GetLabels()
	incidentRootCause := "null"
	incidentChannel := data.GetChannel()
	incidentStartTime := dt.Local()
	incidentRecoverTime := time.Time{}

	if err := u.slackRepo.InsertNewRelicIncident(
		ctx,
		incidentID,
		incidentConditionID,
		incidentName,
		incidentURL,
		incidentDescription,
		incidentOwner,
		incidentGeneratedBy,
		incidentStatus,
		incidentSeverity,
		incidentRootCause,
		incidentChannel,
		incidentLabels,
		incidentStartTime,
		incidentRecoverTime,
	); err != nil {
		log.Errorf("Error store incident to database: %s", err)
		return err
	}

	return nil
}

// AckMessage provides an ack form.
// Ack form comes up whenever user does an ack on a Slack Message (e.g. clicked Ack button to open an Ack form)
func (u *UseCase) AckMessage(ctx context.Context, message slack.InteractionCallback) (entitySlack.Incident, string, string, error) {
	replaceOriginalMessage := true
	tsMessage := message.Message.Timestamp
	channelID := message.Container.ChannelID
	triggerID := message.TriggerID
	username := message.User.Name
	workspace := message.Team.Domain

	// Record Slack Message related metadata.
	if err := u.slackRepo.UpdateMessageByTimestamp(ctx, triggerID, workspace, username, tsMessage, channelID); err != nil {
		log.Errorf("Error store message to database: %s", err)
	}

	// Retrieve the original Slack Message that's currently being Ack'ed
	slackMessage, err := u.slackRepo.GetMessageByTimestamp(ctx, tsMessage, channelID)
	if err != nil {
		log.Errorf("Error GET message on database: %s", err)
	}

	// Get incident ID and condition alert ID
	incident, err := u.slackRepo.GetNewRelicIncidentByMsgTimestamp(ctx, slackMessage.IncidentID, slackMessage.MessageTimestamp)
	if err != nil {
		log.Errorf("Error GET incident on database: %s", err)
	}

	// Construct Ack form
	blockActions := message.ActionCallback.BlockActions
	optionsData := u.GetOptionStr(incident.ConditionID)
	incidentTitle := u.GetTitle(incident.GeneratedBy, incident.Status, incident.Name, incident.URL)
	incidentColor := u.GetColorStr(incident.Status)
	incidentMessage := u.GetMessageString(incident.Description, incident.Status, incident.StartTime, incident.RecoverTime)

	// Respond with Ack form
	result, err := u.slackRepo.SubmitButtonAction(blockActions, optionsData, channelID, tsMessage, triggerID, incidentTitle, incidentMessage, incidentColor, username, incident.URL, replaceOriginalMessage)
	if err != nil {
		log.Errorf("Failed update slack block message because: %s", err)
	}

	return incident, result, slackMessage.MessageTimestamp, nil
}

// SubmitAckForm accepts Ack form submission and processes it (e.g. updates the Slack Message with new information).
func (u *UseCase) SubmitAckForm(ctx context.Context, message slack.InteractionCallback, messageTimestamp, channel string) (entitySlack.Incident, string, error) {
	var actionValue string
	replaceOriginalMessage := true
	username := message.User.Name
	viewState := message.View.State.Values

	for _, state := range viewState {
		for _, value := range state {
			selectedOptions := value.SelectedOption.Value
			textValue := value.Value
			if len(textValue) > 0 {
				actionValue = textValue
			}

			if len(selectedOptions) > 0 {
				actionValue = selectedOptions
			}
		}
	}

	// Retrieve Slack Message
	slackMessage, err := u.slackRepo.GetMessageByTimestamp(ctx, messageTimestamp, channel)
	if err != nil {
		log.Errorf("Error GET message on database: %s", err)
	}

	// Store Incident information from Ack form
	if err := u.slackRepo.UpdateNewRelicIncidentByID(ctx, actionValue, slackMessage.IncidentID); err != nil {
		log.Errorf("Error store message to database: %s", err)
	}

	// Retrieve Incident information.
	incident, err := u.slackRepo.GetNewRelicIncidentByMsgTimestamp(ctx, slackMessage.IncidentID, slackMessage.MessageTimestamp)
	if err != nil {
		log.Errorf("Error GET incident on database: %s", err)
	}

	// Update Slack Message to reflect new information from Ack form.
	incidentTitle := u.GetTitle(incident.GeneratedBy, incident.Status, incident.Name, incident.URL)
	incidentColor := u.GetColorStr(incident.Status)
	incidentMessage := u.GetMessageString(incident.Description, incident.Status, incident.StartTime, incident.RecoverTime)
	_, err = u.slackRepo.ReplaceMessage(incident.Channel, slackMessage.MessageTimestamp, actionValue, incidentTitle, incidentMessage, incidentColor, username, incident.URL, replaceOriginalMessage)
	if err != nil {
		log.Errorf("Failed update slack block message because: %s", err)
	}

	return incident, actionValue, nil
}

func (u *UseCase) GetOptionStr(data int) []string {
	optionData := []string{}

	for _, v := range webhook.DiaryWebhookConfig.Slack.NewRelic {
		if data == v.AlertConditionID {
			for _, cause := range v.AlertConditionCause {
				entry := cause
				optionData = append(optionData, u.GetDataOptions(entry))
			}
		}
	}

	return optionData
}

func (u *UseCase) GetLabels(data, key string) string {
	labels := map[string]string{}
	json.Unmarshal([]byte(data), &labels)

	value := labels[key]
	return value
}

func (u *UseCase) GetTitle(generatedBy, status, name, url string) string {
	title := fmt.Sprintf("[%s:newrelic:] %s", generatedBy, name)

	messageTitle := fmt.Sprintf("%s\n", title)
	return messageTitle
}

func (u *UseCase) GetMessageString(body, status string, startTime, recoverTime time.Time) string {
	start := startTime.Local().Unix() - 7*3600
	recover := recoverTime.Local().Unix() - 7*3600

	ttr := ""
	if status == "closed" && !startTime.IsZero() {
		duration := time.Now().Local().Sub(time.Unix(start, 0)).Seconds()
		hour := int(duration / 3600)
		left := int(duration) % 3600
		minute := int(left / 60)
		second := int(left % 60)

		str := ""
		if hour > 0 {
			str += fmt.Sprintf("%dh ", hour)
		}
		if minute > 0 {
			str += fmt.Sprintf("%dm ", minute)
		}
		if second > 0 {
			str += fmt.Sprintf("%ds ", second)
		}

		ttr = fmt.Sprintf("\n*Time to Resolve* : *%s* (*%s*)", str, time.Unix(recover, 0).Format(time.RFC1123))
	}

	message := fmt.Sprintf("*Current Status* : *`%s`*\n*Incident Time* : %s\n%s\n\n*Incident* : \n%s\n\n ", status, time.Unix(start, 0).Format(time.RFC1123), ttr, body)

	return message
}

func (u *UseCase) GetMessage(data entitySlack.NewRelicReplyThread) string {
	body := data.GetBody()
	message := fmt.Sprintf("%s\n*Status : *`%s`\n*Incident : *\n%s", data.GetTitle(), data.GetState(), body)

	return message
}

func (u *UseCase) GetMessageSummary(data entitySlack.NewRelicReplyThread, start, recover time.Time) string {
	// url := data.GetURL()
	title := data.GetTitle()
	body := data.GetBody()

	status := data.GetState()

	startTime := start.Local().Unix() - 7*3600
	recoverTime := recover.Local().Unix() - 7*3600

	ttr := ""
	if status == "closed" || status == "acknowledged" && !time.Unix(startTime, 0).IsZero() {
		duration := time.Now().Local().Sub(time.Unix(startTime, 0)).Seconds()
		hour := int(duration / 3600)
		left := int(duration) % 3600
		minute := int(left / 60)
		second := int(left % 60)

		str := ""
		if hour > 0 {
			str += fmt.Sprintf("%dh ", hour)
		}
		if minute > 0 {
			str += fmt.Sprintf("%dm ", minute)
		}
		if second > 0 {
			str += fmt.Sprintf("%ds ", second)
		}

		ttr = fmt.Sprintf("\n*Time to Resolve* : *%s* (*%s*)", str, time.Unix(recoverTime, 0).Format(time.RFC1123))
	}

	message := fmt.Sprintf("%s\n*Current Status* : *`%s`*\n*Incident Time* : %s\n%s\n\n*Incident* : \n%s\n\n ", title, status, time.Unix(startTime, 0).Format(time.RFC1123), ttr, body)

	return message
}

func (u *UseCase) GetColor(data entitySlack.NewRelicReplyThread) string {
	color := "FFFFFF"
	if data.GetState() == "open" {
		color = "FF0000"
	} else {
		color = "00BF85"
	}

	return color
}

func (u *UseCase) GetColorStr(status string) string {
	color := "FFFFFF"
	if status == "open" {
		color = "FF0000"
	} else {
		color = "00BF85"
	}

	return color
}

func (u *UseCase) GetIncidentName(data entitySlack.NewRelicReplyThread) string {
	var incidentName string
	for _, v := range webhook.DiaryWebhookConfig.Slack.NewRelic {
		if data.GetConditionID() == v.AlertConditionID {
			incidentName = v.AlertConditionName
		}
	}
	return incidentName
}

func (u *UseCase) GetDataOptions(data string) string {
	s := data
	text := strings.Replace(s, "-", " ", -1)

	return text
}

func (u *UseCase) ConvertValues(data string) string {
	s := data
	lower := strings.ToLower(s)
	text := strings.Replace(lower, " ", "-", -1)

	return text
}
