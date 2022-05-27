package cmd

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/ans"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/sirupsen/logrus"
	"time"
)

func ansSendEvent(config ansSendEventOptions, telemetryData *telemetry.CustomData) {
	err := runAnsSendEvent(&config, &ans.ANS{})
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAnsSendEvent(config *ansSendEventOptions, c ans.Client) error {
	ansServiceKey, err := ans.UnmarshallServiceKeyJSON(config.AnsServiceKey)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return err
	}
	c.SetServiceKey(ansServiceKey)

	event := ans.Event{
		EventType: "Piper",
		Resource: &ans.Resource{
			ResourceType: "Pipeline",
			ResourceName: "Pipeline",
		},
		Subject:        fmt.Sprint(log.Entry().Data["stepName"]),
		Body:           fmt.Sprintf("Call from Piper step: %s", log.Entry().Data["stepName"]),
	}
	event.SetSeverityAndCategory(logrus.InfoLevel)

	if err = event.MergeWithJSON([]byte(config.EventJSON)); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return err
	}
	// We set the time
	event.EventTimestamp = time.Now().Unix()
	if err = c.Send(event); err != nil {
		log.SetErrorCategory(log.ErrorService)
	}
	return err
}
