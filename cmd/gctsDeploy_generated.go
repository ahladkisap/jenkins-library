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

type gctsDeployOptions struct {
	Username            string                 `json:"username,omitempty"`
	Password            string                 `json:"password,omitempty"`
	Repository          string                 `json:"repository,omitempty"`
	Host                string                 `json:"host,omitempty"`
	Client              string                 `json:"client,omitempty"`
	Commit              string                 `json:"commit,omitempty"`
	RemoteRepositoryURL string                 `json:"remoteRepositoryURL,omitempty"`
	Role                string                 `json:"role,omitempty" validate:"possible-values=SOURCE TARGET"`
	VSID                string                 `json:"vSID,omitempty"`
	Type                string                 `json:"type,omitempty" validate:"possible-values=GIT GITHUB GITLAB"`
	Branch              string                 `json:"branch,omitempty"`
	Scope               string                 `json:"scope,omitempty"`
	Rollback            bool                   `json:"rollback,omitempty"`
	Configuration       map[string]interface{} `json:"configuration,omitempty"`
}

// GctsDeployCommand Deploys a Git Repository to a local Repository and then to an ABAP System
func GctsDeployCommand() *cobra.Command {
	const STEP_NAME = "gctsDeploy"

	metadata := gctsDeployMetadata()
	var stepConfig gctsDeployOptions
	var startTime time.Time
	var logCollector *log.CollectorHook
	var splunkClient *splunk.Splunk
	telemetryClient := &telemetry.Telemetry{}

	var createGctsDeployCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Deploys a Git Repository to a local Repository and then to an ABAP System",
		Long: `This step deploys a remote Git repository to a local repository on an ABAP system and imports the content in the ABAP database.
If the repository does not yet exist in the system, this step also creates it.
If the repository already exists on the ABAP system, this step executes the remaining actions of the step, depending on the parameters provided for the step.
These actions include, for example, deploy a specific commit to the ABAP system, or deploy the current commit of a specific branch.
You can use this step for gCTS as of SAP S/4HANA 2020.`,
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
			log.RegisterSecret(stepConfig.Username)
			log.RegisterSecret(stepConfig.Password)

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
			gctsDeploy(stepConfig, &stepTelemetryData)
			stepTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addGctsDeployFlags(createGctsDeployCmd, &stepConfig)
	return createGctsDeployCmd
}

func addGctsDeployFlags(cmd *cobra.Command, stepConfig *gctsDeployOptions) {
	cmd.Flags().StringVar(&stepConfig.Username, "username", os.Getenv("PIPER_username"), "User that authenticates to the ABAP system. **Note** - Don't provide this parameter directly. Either set it in the environment, or in the Jenkins credentials store, and provide the ID as value of the `abapCredentialsId` parameter.")
	cmd.Flags().StringVar(&stepConfig.Password, "password", os.Getenv("PIPER_password"), "Password of the ABAP user that authenticates to the ABAP system. **Note** - Don´t provide this parameter directly. Either set it in the environment, or in the Jenkins credentials store, and provide the ID as value of the `abapCredentialsId` parameter.")
	cmd.Flags().StringVar(&stepConfig.Repository, "repository", os.Getenv("PIPER_repository"), "Specifies the name (ID) of the local repsitory on the ABAP system")
	cmd.Flags().StringVar(&stepConfig.Host, "host", os.Getenv("PIPER_host"), "Protocol and host of the ABAP system, including the port. Please provide in the format `<protocol>://<host>:<port>`. Supported protocols are `http` and `https`.")
	cmd.Flags().StringVar(&stepConfig.Client, "client", os.Getenv("PIPER_client"), "Client of the ABAP system to which you want to deploy the repository")
	cmd.Flags().StringVar(&stepConfig.Commit, "commit", os.Getenv("PIPER_commit"), "ID of a specific commit, if you want to deploy the content of the specified commit.")
	cmd.Flags().StringVar(&stepConfig.RemoteRepositoryURL, "remoteRepositoryURL", os.Getenv("PIPER_remoteRepositoryURL"), "URL of the remote repository")
	cmd.Flags().StringVar(&stepConfig.Role, "role", `SOURCE`, "Role of the local repository. Possible values are 'SOURCE' (for repositories on development systems - Default) and 'TARGET' (for repositories on target systems). Local repositories with a TARGET role cannot be the source of code changes.")
	cmd.Flags().StringVar(&stepConfig.VSID, "vSID", os.Getenv("PIPER_vSID"), "Virtual SID of the local repository. The vSID corresponds to the transport route that delivers content to the remote Git repository. For more information, see [Background Information - vSID](https://help.sap.com/viewer/4a368c163b08418890a406d413933ba7/latest/en-US/8edc17edfc374908bd8a1615ea5ab7b7.html) on SAP Help Portal.")
	cmd.Flags().StringVar(&stepConfig.Type, "type", `GIT`, "Type of the used source code management tool")
	cmd.Flags().StringVar(&stepConfig.Branch, "branch", os.Getenv("PIPER_branch"), "Name of a branch, if you want to deploy the content of a specific branch to the ABAP system.")
	cmd.Flags().StringVar(&stepConfig.Scope, "scope", os.Getenv("PIPER_scope"), "Scope of objects to be deployed. Possible values are CRNTCOMMIT (current commit - Default) and LASTACTION (last repository action). The default option deploys all objects that existed in the repository when the commit was created. LASTACTION only deploys the object difference of the last action in the repository.")
	cmd.Flags().BoolVar(&stepConfig.Rollback, "rollback", false, "Indication whether you want to roll back to the last working state of the repository, if any of the step actions *switch branch* or *pull commit* fail.")

	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("repository")
	cmd.MarkFlagRequired("host")
	cmd.MarkFlagRequired("client")
	cmd.MarkFlagRequired("remoteRepositoryURL")
}

// retrieve step metadata
func gctsDeployMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "gctsDeploy",
			Aliases:     []config.Alias{},
			Description: "Deploys a Git Repository to a local Repository and then to an ABAP System",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "abapCredentialsId", Description: "ID taken from the Jenkins credentials store containing the user name and password of the user that authenticates to the ABAP system.", Type: "jenkins"},
				},
				Parameters: []config.StepParameters{
					{
						Name: "username",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "abapCredentialsId",
								Param: "username",
								Type:  "secret",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_username"),
					},
					{
						Name: "password",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "abapCredentialsId",
								Param: "password",
								Type:  "secret",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_password"),
					},
					{
						Name:        "repository",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_repository"),
					},
					{
						Name:        "host",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_host"),
					},
					{
						Name:        "client",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_client"),
					},
					{
						Name:        "commit",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_commit"),
					},
					{
						Name:        "remoteRepositoryURL",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_remoteRepositoryURL"),
					},
					{
						Name:        "role",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `SOURCE`,
					},
					{
						Name:        "vSID",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_vSID"),
					},
					{
						Name:        "type",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `GIT`,
					},
					{
						Name:        "branch",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_branch"),
					},
					{
						Name:        "scope",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_scope"),
					},
					{
						Name:        "rollback",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "configuration",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "map[string]interface{}",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "gctsRepositoryConfigurations"}},
					},
				},
			},
		},
	}
	return theMetaData
}
