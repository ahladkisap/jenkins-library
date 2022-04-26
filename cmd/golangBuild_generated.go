// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/gcs"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/validation"
	"github.com/bmatcuk/doublestar"
	"github.com/spf13/cobra"
)

type golangBuildOptions struct {
	BuildFlags                   []string `json:"buildFlags,omitempty"`
	BuildSettingsInfo            string   `json:"buildSettingsInfo,omitempty"`
	CgoEnabled                   bool     `json:"cgoEnabled,omitempty"`
	CoverageFormat               string   `json:"coverageFormat,omitempty" validate:"possible-values=cobertura html"`
	CreateBOM                    bool     `json:"createBOM,omitempty"`
	CustomTLSCertificateLinks    []string `json:"customTlsCertificateLinks,omitempty"`
	ExcludeGeneratedFromCoverage bool     `json:"excludeGeneratedFromCoverage,omitempty"`
	LdflagsTemplate              string   `json:"ldflagsTemplate,omitempty"`
	Output                       string   `json:"output,omitempty"`
	Packages                     []string `json:"packages,omitempty"`
	Publish                      bool     `json:"publish,omitempty"`
	TargetRepositoryPassword     string   `json:"targetRepositoryPassword,omitempty"`
	TargetRepositoryUser         string   `json:"targetRepositoryUser,omitempty"`
	TargetRepositoryURL          string   `json:"targetRepositoryURL,omitempty"`
	ReportCoverage               bool     `json:"reportCoverage,omitempty"`
	RunTests                     bool     `json:"runTests,omitempty"`
	RunIntegrationTests          bool     `json:"runIntegrationTests,omitempty"`
	TargetArchitectures          []string `json:"targetArchitectures,omitempty"`
	TestOptions                  []string `json:"testOptions,omitempty"`
	TestResultFormat             string   `json:"testResultFormat,omitempty" validate:"possible-values=junit standard"`
	PrivateModules               string   `json:"privateModules,omitempty"`
	PrivateModulesGitToken       string   `json:"privateModulesGitToken,omitempty"`
	ArtifactVersion              string   `json:"artifactVersion,omitempty"`
}

type golangBuildCommonPipelineEnvironment struct {
	custom struct {
		buildSettingsInfo string
	}
}

func (p *golangBuildCommonPipelineEnvironment) persist(path, resourceName string) {
	content := []struct {
		category string
		name     string
		value    interface{}
	}{
		{category: "custom", name: "buildSettingsInfo", value: p.custom.buildSettingsInfo},
	}

	errCount := 0
	for _, param := range content {
		err := piperenv.SetResourceParameter(path, resourceName, filepath.Join(param.category, param.name), param.value)
		if err != nil {
			log.Entry().WithError(err).Error("Error persisting piper environment.")
			errCount++
		}
	}
	if errCount > 0 {
		log.Entry().Error("failed to persist Piper environment")
	}
}

type golangBuildReports struct {
}

func (p *golangBuildReports) persist(stepConfig golangBuildOptions, gcpJsonKeyFilePath string, gcsBucketId string, gcsFolderPath string, gcsSubFolder string) {
	if gcsBucketId == "" {
		log.Entry().Info("persisting reports to GCS is disabled, because gcsBucketId is empty")
		return
	}
	log.Entry().Info("Uploading reports to Google Cloud Storage...")
	content := []gcs.ReportOutputParam{
		{FilePattern: "**/bom.xml", ParamRef: "", StepResultType: "sbom"},
		{FilePattern: "**/TEST-*.xml", ParamRef: "", StepResultType: "junit"},
		{FilePattern: "**/cobertura-coverage.xml", ParamRef: "", StepResultType: "cobertura-coverage"},
	}
	envVars := []gcs.EnvVar{
		{Name: "GOOGLE_APPLICATION_CREDENTIALS", Value: gcpJsonKeyFilePath, Modified: false},
	}
	gcsClient, err := gcs.NewClient(gcs.WithEnvVars(envVars))
	if err != nil {
		log.Entry().Errorf("creation of GCS client failed: %v", err)
		return
	}
	defer gcsClient.Close()
	structVal := reflect.ValueOf(&stepConfig).Elem()
	inputParameters := map[string]string{}
	for i := 0; i < structVal.NumField(); i++ {
		field := structVal.Type().Field(i)
		if field.Type.String() == "string" {
			paramName := strings.Split(field.Tag.Get("json"), ",")
			paramValue, _ := structVal.Field(i).Interface().(string)
			inputParameters[paramName[0]] = paramValue
		}
	}
	if err := gcs.PersistReportsToGCS(gcsClient, content, inputParameters, gcsFolderPath, gcsBucketId, gcsSubFolder, doublestar.Glob, os.Stat); err != nil {
		log.Entry().Errorf("failed to persist reports: %v", err)
	}
}

// GolangBuildCommand This step will execute a golang build.
func GolangBuildCommand() *cobra.Command {
	const STEP_NAME = "golangBuild"

	metadata := golangBuildMetadata()
	var stepConfig golangBuildOptions
	var startTime time.Time
	var commonPipelineEnvironment golangBuildCommonPipelineEnvironment
	var reports golangBuildReports
	var logCollector *log.CollectorHook
	var splunkClient *splunk.Splunk
	telemetryClient := &telemetry.Telemetry{}

	var createGolangBuildCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "This step will execute a golang build.",
		Long: `This step will build a golang project.
It will also execute golang-based tests using [gotestsum](https://github.com/gotestyourself/gotestsum) and with that allows for reporting test results and test coverage.

Besides execution of the default tests the step allows for running an additional integration test run using ` + "`" + `-tags=integration` + "`" + ` using pattern ` + "`" + `./...` + "`" + `

If the build is successful the resulting artifact can be uploaded to e.g. a binary repository automatically.`,
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
			log.RegisterSecret(stepConfig.TargetRepositoryPassword)
			log.RegisterSecret(stepConfig.TargetRepositoryUser)
			log.RegisterSecret(stepConfig.PrivateModulesGitToken)

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
				commonPipelineEnvironment.persist(GeneralConfig.EnvRootPath, "commonPipelineEnvironment")
				reports.persist(stepConfig, GeneralConfig.GCPJsonKeyFilePath, GeneralConfig.GCSBucketId, GeneralConfig.GCSFolderPath, GeneralConfig.GCSSubFolder)
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
			golangBuild(stepConfig, &stepTelemetryData, &commonPipelineEnvironment)
			stepTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addGolangBuildFlags(createGolangBuildCmd, &stepConfig)
	return createGolangBuildCmd
}

func addGolangBuildFlags(cmd *cobra.Command, stepConfig *golangBuildOptions) {
	cmd.Flags().StringSliceVar(&stepConfig.BuildFlags, "buildFlags", []string{}, "Defines list of build flags to be used.")
	cmd.Flags().StringVar(&stepConfig.BuildSettingsInfo, "buildSettingsInfo", os.Getenv("PIPER_buildSettingsInfo"), "build settings info is typically filled by the step automatically to create information about the build settings that were used during the maven build . This information is typically used for compliance related processes.")
	cmd.Flags().BoolVar(&stepConfig.CgoEnabled, "cgoEnabled", false, "If active: enables the creation of Go packages that call C code.")
	cmd.Flags().StringVar(&stepConfig.CoverageFormat, "coverageFormat", `html`, "Defines the format of the coverage repository.")
	cmd.Flags().BoolVar(&stepConfig.CreateBOM, "createBOM", false, "Creates the bill of materials (BOM) using CycloneDX plugin. It requires Go 1.17 or newer.")
	cmd.Flags().StringSliceVar(&stepConfig.CustomTLSCertificateLinks, "customTlsCertificateLinks", []string{}, "List of download links to custom TLS certificates. This is required to ensure trusted connections to instances with repositories (like nexus) when publish flag is set to true.")
	cmd.Flags().BoolVar(&stepConfig.ExcludeGeneratedFromCoverage, "excludeGeneratedFromCoverage", true, "Defines if generated files should be excluded, according to [https://golang.org/s/generatedcode](https://golang.org/s/generatedcode).")
	cmd.Flags().StringVar(&stepConfig.LdflagsTemplate, "ldflagsTemplate", os.Getenv("PIPER_ldflagsTemplate"), "Defines the content of -ldflags option in a golang template format.")
	cmd.Flags().StringVar(&stepConfig.Output, "output", os.Getenv("PIPER_output"), "Defines the build result or output directory as per `go build` documentation.")
	cmd.Flags().StringSliceVar(&stepConfig.Packages, "packages", []string{}, "List of packages to be build as per `go build` documentation.")
	cmd.Flags().BoolVar(&stepConfig.Publish, "publish", false, "Configures the build to publish artifacts to a repository.")
	cmd.Flags().StringVar(&stepConfig.TargetRepositoryPassword, "targetRepositoryPassword", os.Getenv("PIPER_targetRepositoryPassword"), "Password for the target repository where the compiled binaries shall be uploaded - typically provided by the CI/CD environment.")
	cmd.Flags().StringVar(&stepConfig.TargetRepositoryUser, "targetRepositoryUser", os.Getenv("PIPER_targetRepositoryUser"), "Username for the target repository where the compiled binaries shall be uploaded - typically provided by the CI/CD environment.")
	cmd.Flags().StringVar(&stepConfig.TargetRepositoryURL, "targetRepositoryURL", os.Getenv("PIPER_targetRepositoryURL"), "URL of the target repository where the compiled binaries shall be uploaded - typically provided by the CI/CD environment.")
	cmd.Flags().BoolVar(&stepConfig.ReportCoverage, "reportCoverage", true, "Defines if a coverage report should be created.")
	cmd.Flags().BoolVar(&stepConfig.RunTests, "runTests", true, "Activates execution of tests using [gotestsum](https://github.com/gotestyourself/gotestsum).")
	cmd.Flags().BoolVar(&stepConfig.RunIntegrationTests, "runIntegrationTests", false, "Activates execution of a second test run using tag `integration`.")
	cmd.Flags().StringSliceVar(&stepConfig.TargetArchitectures, "targetArchitectures", []string{`linux,amd64`}, "Defines the target architectures for which the build should run using OS and architecture separated by a comma.")
	cmd.Flags().StringSliceVar(&stepConfig.TestOptions, "testOptions", []string{}, "Options to pass to test as per `go test` documentation (comprises e.g. flags, packages).")
	cmd.Flags().StringVar(&stepConfig.TestResultFormat, "testResultFormat", `junit`, "Defines the output format of the test results.")
	cmd.Flags().StringVar(&stepConfig.PrivateModules, "privateModules", os.Getenv("PIPER_privateModules"), "Tells go which modules shall be considered to be private (by setting [GOPRIVATE](https://pkg.go.dev/cmd/go#hdr-Configuration_for_downloading_non_public_code)).")
	cmd.Flags().StringVar(&stepConfig.PrivateModulesGitToken, "privateModulesGitToken", os.Getenv("PIPER_privateModulesGitToken"), "GitHub personal access token as per https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line.")
	cmd.Flags().StringVar(&stepConfig.ArtifactVersion, "artifactVersion", os.Getenv("PIPER_artifactVersion"), "Version of the artifact to be built.")

	cmd.MarkFlagRequired("targetArchitectures")
}

// retrieve step metadata
func golangBuildMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "golangBuild",
			Aliases:     []config.Alias{},
			Description: "This step will execute a golang build.",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "golangPrivateModulesGitTokenCredentialsId", Description: "Jenkins 'Username with password' credentials ID containing username/password for http access to your git repos where your go private modules are stored.", Type: "jenkins"},
				},
				Parameters: []config.StepParameters{
					{
						Name:        "buildFlags",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name: "buildSettingsInfo",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/buildSettingsInfo",
							},
						},
						Scope:     []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_buildSettingsInfo"),
					},
					{
						Name:        "cgoEnabled",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "coverageFormat",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `html`,
					},
					{
						Name:        "createBOM",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "STEPS", "STAGES", "PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "customTlsCertificateLinks",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name:        "excludeGeneratedFromCoverage",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     true,
					},
					{
						Name:        "ldflagsTemplate",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_ldflagsTemplate"),
					},
					{
						Name:        "output",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_output"),
					},
					{
						Name:        "packages",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name:        "publish",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name: "targetRepositoryPassword",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/rawRepositoryPassword",
							},

							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/repositoryPassword",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_targetRepositoryPassword"),
					},
					{
						Name: "targetRepositoryUser",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/rawRepositoryUsername",
							},

							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/repositoryUsername",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_targetRepositoryUser"),
					},
					{
						Name: "targetRepositoryURL",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/rawRepositoryURL",
							},

							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/repositoryUrl",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_targetRepositoryURL"),
					},
					{
						Name:        "reportCoverage",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     true,
					},
					{
						Name:        "runTests",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     true,
					},
					{
						Name:        "runIntegrationTests",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "targetArchitectures",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "STEPS", "STAGES", "PARAMETERS"},
						Type:        "[]string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     []string{`linux,amd64`},
					},
					{
						Name:        "testOptions",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name:        "testResultFormat",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `junit`,
					},
					{
						Name:        "privateModules",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_privateModules"),
					},
					{
						Name: "privateModulesGitToken",
						ResourceRef: []config.ResourceReference{
							{
								Name: "golangPrivateModulesGitTokenCredentialsId",
								Type: "secret",
							},

							{
								Name:    "golangPrivateModulesGitTokenVaultSecret",
								Type:    "vaultSecret",
								Default: "golang",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_privateModulesGitToken"),
					},
					{
						Name: "artifactVersion",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "artifactVersion",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_artifactVersion"),
					},
				},
			},
			Containers: []config.Container{
				{Name: "golang", Image: "golang:1", Options: []config.Option{{Name: "-u", Value: "0"}}},
			},
			Outputs: config.StepOutputs{
				Resources: []config.StepResources{
					{
						Name: "commonPipelineEnvironment",
						Type: "piperEnvironment",
						Parameters: []map[string]interface{}{
							{"name": "custom/buildSettingsInfo"},
						},
					},
					{
						Name: "reports",
						Type: "reports",
						Parameters: []map[string]interface{}{
							{"filePattern": "**/bom.xml", "type": "sbom"},
							{"filePattern": "**/TEST-*.xml", "type": "junit"},
							{"filePattern": "**/cobertura-coverage.xml", "type": "cobertura-coverage"},
						},
					},
				},
			},
		},
	}
	return theMetaData
}
