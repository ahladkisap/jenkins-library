// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/validation"
	"github.com/spf13/cobra"
)

type integrationArtifactTriggerIntegrationTestOptions struct {
	IntegrationFlowServiceKey         string `json:"integrationFlowServiceKey,omitempty"`
	IntegrationFlowID                 string `json:"integrationFlowId,omitempty"`
	IntegrationFlowServiceEndpointURL string `json:"integrationFlowServiceEndpointUrl,omitempty"`
	ContentType                       string `json:"contentType,omitempty"`
	MessageBodyPath                   string `json:"messageBodyPath,omitempty"`
}

// IntegrationArtifactTriggerIntegrationTestCommand Test the service endpoint of your iFlow
func IntegrationArtifactTriggerIntegrationTestCommand() *cobra.Command {
	const STEP_NAME = "integrationArtifactTriggerIntegrationTest"

	metadata := integrationArtifactTriggerIntegrationTestMetadata()
	var stepConfig integrationArtifactTriggerIntegrationTestOptions
	var startTime time.Time
	var logCollector *log.CollectorHook
	var splunkClient *splunk.Splunk
	telemetryClient := &telemetry.Telemetry{}

	var createIntegrationArtifactTriggerIntegrationTestCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Test the service endpoint of your iFlow",
		Long:  `With this step you can test your intergration flow  exposed by SAP Cloud Platform Integration on a tenant using OData API.Learn more about the SAP Cloud Integration remote API for getting service endpoint of deployed integration artifact [here](https://help.sap.com/viewer/368c481cd6954bdfa5d0435479fd4eaf/Cloud/en-US/d1679a80543f46509a7329243b595bdb.html).`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			startTime = time.Now()
			log.SetStepName(STEP_NAME)
			log.SetVerbose(GeneralConfig.Verbose)

			GeneralConfig.GitHubAccessTokens = ResolveAccessTokens(GeneralConfig.GitHubTokens)

			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)

			err := PrepareConfig(cmd, &metadata, STEP_NAME, &stepConfig, config.OpenPiperFile)
			if err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}
			log.RegisterSecret(stepConfig.IntegrationFlowServiceKey)

			if len(GeneralConfig.HookConfig.SentryConfig.Dsn) > 0 {
				sentryHook := log.NewSentryHook(GeneralConfig.HookConfig.SentryConfig.Dsn, GeneralConfig.CorrelationID)
				log.RegisterHook(&sentryHook)
			}

			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
				splunkClient = &splunk.Splunk{}
				logCollector = &log.CollectorHook{CorrelationID: GeneralConfig.CorrelationID}
				log.RegisterHook(logCollector)
			}

			if len(GeneralConfig.HookConfig.ANSConfig.ServiceKey) == 0 {
				GeneralConfig.HookConfig.ANSConfig.ServiceKey = os.Getenv("PIPER_ansServiceKey")
			}
			if len(GeneralConfig.HookConfig.ANSConfig.ServiceKey) > 0 {
				log.RegisterSecret(GeneralConfig.HookConfig.ANSConfig.ServiceKey)
				ansHook, err := log.NewANSHook(GeneralConfig.HookConfig.ANSConfig, GeneralConfig.CorrelationID)
				if err != nil {
					log.Entry().WithError(err).Warn("failed to set up SAP Alert Notification Service log hook")
				} else {
					log.RegisterHook(&ansHook)
				}
			}

			validation, err := validation.New(validation.WithJSONNamesForStructFields(), validation.WithPredefinedErrorMessages())
			if err != nil {
				return err
			}
			if err = validation.ValidateStruct(stepConfig); err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}

			return nil
		},
		Run: func(_ *cobra.Command, _ []string) {
			stepTelemetryData := telemetry.CustomData{}
			stepTelemetryData.ErrorCode = "1"
			handler := func() {
				config.RemoveVaultSecretFiles()
				stepTelemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				stepTelemetryData.ErrorCategory = log.GetErrorCategory().String()
				stepTelemetryData.PiperCommitHash = GitCommit
				telemetryClient.SetData(&stepTelemetryData)
				telemetryClient.Send()
				if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
					splunkClient.Send(telemetryClient.GetData(), logCollector)
				}
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetryClient.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
				splunkClient.Initialize(GeneralConfig.CorrelationID,
					GeneralConfig.HookConfig.SplunkConfig.Dsn,
					GeneralConfig.HookConfig.SplunkConfig.Token,
					GeneralConfig.HookConfig.SplunkConfig.Index,
					GeneralConfig.HookConfig.SplunkConfig.SendLogs)
			}
			integrationArtifactTriggerIntegrationTest(stepConfig, &stepTelemetryData)
			stepTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addIntegrationArtifactTriggerIntegrationTestFlags(createIntegrationArtifactTriggerIntegrationTestCmd, &stepConfig)
	return createIntegrationArtifactTriggerIntegrationTestCmd
}

func addIntegrationArtifactTriggerIntegrationTestFlags(cmd *cobra.Command, stepConfig *integrationArtifactTriggerIntegrationTestOptions) {
	cmd.Flags().StringVar(&stepConfig.IntegrationFlowServiceKey, "integrationFlowServiceKey", os.Getenv("PIPER_integrationFlowServiceKey"), "Service key JSON string to access the Process Integration Runtime service instance of plan 'integration-flow'")
	cmd.Flags().StringVar(&stepConfig.IntegrationFlowID, "integrationFlowId", os.Getenv("PIPER_integrationFlowId"), "Specifies the ID of the Integration Flow artifact")
	cmd.Flags().StringVar(&stepConfig.IntegrationFlowServiceEndpointURL, "integrationFlowServiceEndpointUrl", os.Getenv("PIPER_integrationFlowServiceEndpointUrl"), "Specifies the URL endpoint of the iFlow. Please provide in the format `<protocol>://<host>:<port>`. Supported protocols are `http` and `https`.")
	cmd.Flags().StringVar(&stepConfig.ContentType, "contentType", os.Getenv("PIPER_contentType"), "Specifies the content type of the file defined in messageBodyPath e.g. application/json")
	cmd.Flags().StringVar(&stepConfig.MessageBodyPath, "messageBodyPath", os.Getenv("PIPER_messageBodyPath"), "Speficfies the relative file path to the message body.")

	cmd.MarkFlagRequired("integrationFlowServiceKey")
	cmd.MarkFlagRequired("integrationFlowId")
	cmd.MarkFlagRequired("integrationFlowServiceEndpointUrl")
}

// retrieve step metadata
func integrationArtifactTriggerIntegrationTestMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "integrationArtifactTriggerIntegrationTest",
			Aliases:     []config.Alias{},
			Description: "Test the service endpoint of your iFlow",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "integrationFlowServiceKeyCredentialsId", Description: "Jenkins secret text credential ID containing the service key to the Process Integration Runtime service instance of plan 'integration-flow'", Type: "jenkins"},
				},
				Parameters: []config.StepParameters{
					{
						Name: "integrationFlowServiceKey",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "integrationFlowServiceKeyCredentialsId",
								Param: "integrationFlowServiceKey",
								Type:  "secret",
							},
						},
						Scope:     []string{"PARAMETERS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_integrationFlowServiceKey"),
					},
					{
						Name:        "integrationFlowId",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_integrationFlowId"),
					},
					{
						Name: "integrationFlowServiceEndpointUrl",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/integrationFlowServiceEndpoint",
							},
						},
						Scope:     []string{"PARAMETERS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_integrationFlowServiceEndpointUrl"),
					},
					{
						Name:        "contentType",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_contentType"),
					},
					{
						Name:        "messageBodyPath",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_messageBodyPath"),
					},
				},
			},
		},
	}
	return theMetaData
}
