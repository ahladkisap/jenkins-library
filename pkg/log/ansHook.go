package log

import (
	"encoding/json"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/ans"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strings"
)

// ANSHook is used to set the hook features for the logrus hook
type ANSHook struct {
	client        ans.Client
	eventTemplate ans.Event
	firing        bool
}

// NewANSHook creates a new ANS hook for logrus
func NewANSHook(config ans.Configuration, correlationID string) (hook ANSHook, err error) {
	return newANSHook(config, correlationID, &ans.ANS{})
}

func newANSHook(config ans.Configuration, correlationID string, client ans.Client) (hook ANSHook, err error) {
	var ansServiceKey ans.ServiceKey
	ansServiceKey, err = ans.UnmarshallServiceKeyJSON(config.ServiceKey); if err != nil {
		err = errors.Wrap(err, "cannot initialize SAP Alert Notification Service due to faulty serviceKey json")
		return
	}
	client.SetServiceKey(ansServiceKey)

	err = client.CheckCorrectSetup(); if err != nil {
		err = errors.Wrap(err, "check http request to SAP Alert Notification Service failed; not setting up the ANS hook")
		return
	}

	event := ans.Event{
		EventType: "Piper",
		Tags:      map[string]interface{}{"ans:correlationId": correlationID, "ans:sourceEventId": correlationID},
		Resource: &ans.Resource{
			ResourceType: "Pipeline",
			ResourceName: "Pipeline",
		},
	}
	if len(config.EventTemplateFilePath) > 0 {
		eventTemplateString, err := ioutil.ReadFile(config.EventTemplateFilePath); if err != nil {
			Entry().WithField("stepName", "ANS").Warnf("provided SAP Alert Notification Service event template file with path '%s' could not be read: %v", config.EventTemplateFilePath, err)
		} else {
			err = event.MergeWithJSON(eventTemplateString); if err != nil {
				Entry().WithField("stepName", "ANS").Warnf("provided SAP Alert Notification Service event template '%s' could not be unmarshalled: %v", eventTemplateString, err)
			}
		}
	}
	if len(config.EventTemplate) == 0 {
		config.EventTemplate = os.Getenv("PIPER_ansEventTemplate")
	}
	if len(config.EventTemplate) > 0 {
		if err = event.MergeWithJSON([]byte(config.EventTemplate)); err != nil {
			Entry().WithField("stepName", "ANS").Warnf("provided SAP Alert Notification Service event template '%s' could not be unmarshalled: %v", config.EventTemplate, err)
		}
	}
	if len(event.Severity) > 0 {
		Entry().WithField("stepName", "ANS").Warnf("event severity set to '%s' will be overwritten according to the log level", event.Severity)
		event.Severity = ""
	}
	if len(event.Category) > 0 {
		Entry().WithField("stepName", "ANS").Warnf("event category set to '%s' will be overwritten according to the log level", event.Category)
		event.Category = ""
	}
	if err = event.Validate(); err != nil {
		err = errors.Wrap(err, "did not initialize SAP Alert Notification Service due to faulty event template json")
		return
	}
	hook = ANSHook{
		client:        client,
		eventTemplate: event,
	}
	return
}

// Levels returns the supported log level of the hook.
func (ansHook *ANSHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.WarnLevel, logrus.ErrorLevel, logrus.PanicLevel, logrus.FatalLevel}
}

// Fire creates a new event from the logrus and sends an event to the ANS backend
func (ansHook *ANSHook) Fire(entry *logrus.Entry) (err error) {
	if ansHook.firing {
		return fmt.Errorf("ANS hook has already been fired")
	}
	ansHook.firing = true
	defer func() { ansHook.firing = false }()

	if len(strings.TrimSpace(entry.Message)) == 0 {
		return
	}
	var event ans.Event
	event, err = ansHook.eventTemplate.Copy(); if err != nil {
		return
	}

	logLevel := entry.Level
	for k, v := range entry.Data {
		event.Tags[k] = v
	}
	if errorCategory := GetErrorCategory().String(); errorCategory != "undefined" {
		event.Tags["errorCategory"] = errorCategory
	}

	event.EventTimestamp = entry.Time.Unix()
	if event.Subject == "" {
		event.Subject = fmt.Sprint(entry.Data["stepName"])
	}
	event.Body = entry.Message
	event.SetSeverityAndCategory(logLevel)
	event.Tags["logLevel"] = logLevel.String()

	return ansHook.client.Send(event)
}

func copyEvent(source ans.Event) (destination ans.Event, err error) {
	var sourceJSON []byte
	sourceJSON, err = json.Marshal(source); if err != nil {
		return
	}
	err = destination.MergeWithJSON(sourceJSON)
	return
}
