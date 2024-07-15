package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/raiecs/ecsclientgowrapper"

	ecsgoclient "github.com/raiecs"
	"github.com/spf13/cobra"
)

var CMD = &cobra.Command{
	Use: "ecs",
	Run: runExample,
}

var (
	ClientName      string
	ProjectTeam     string
	EnvironmentName string
	ServiceName     string
)

func init() {
	CMD.PersistentFlags().StringVar(&ClientName, "client", "", "client")
	CMD.PersistentFlags().StringVar(&ProjectTeam, "projectTeam", "", "project team")
	CMD.PersistentFlags().StringVar(&EnvironmentName, "environment", "", "environment")
	CMD.PersistentFlags().StringVar(&ServiceName, "service", "", "service")
}

func main() {
	if err := CMD.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
}

type EcsConfig struct {
	Moderation EcsModerationConfig `json:"Moderation"`
}

type EcsModerationConfig struct {
	PolicyId string `json:"PolicyId"`
}

func (ecsConfig *EcsConfig) OnOptionsUpdateReceived(bytes []byte) error {
	var parsedConfig EcsConfig
	if err := json.Unmarshal(bytes, &parsedConfig); err != nil {
		return fmt.Errorf("failed to unmarshal options, err: %v", err)
	}

	if err := parsedConfig.Validate(); err != nil {
		return err
	}

	*ecsConfig = parsedConfig
	return nil
}

func (testConfig EcsConfig) Validate() error {
	return nil
}

type ConsoleLogger struct {
}

func (consoleLogger ConsoleLogger) Log(logLevel ecsclientgowrapper.ECS_LOG_LEVEL, msg string) {
	fmt.Printf("Received log. %v: %v \n", logLevel, msg)
}

func runExample(command *cobra.Command, args []string) {
	targetFilters := map[string][]string{
		"EnvironmentName": {EnvironmentName},
		"ServiceName":     {ServiceName},
	}

	consoleLogger := ConsoleLogger{}

	options := ecsgoclient.EcsClientOptions{
		Client:        ClientName,
		ProjectTeams:  []string{ProjectTeam},
		TargetFilters: targetFilters,
		LogLevel:      ecsclientgowrapper.ECS_LOG_LEVEL_INFORMATION,
		Logger:        consoleLogger,
	}

	ecsClientInstance, err := ecsgoclient.NewEcsClient(options)
	if err != nil {
		fmt.Printf("Creating ECS client failed with error: %v", err.Error())
		os.Exit(1)
	}

	ecsConfig := &EcsConfig{}
	err = ecsClientInstance.AddOptionsMonitorToEcsClient(ecsConfig, ProjectTeam, "EcsConfig")
	if err != nil {
		fmt.Printf("Creating ECS client failed with error: %v", err.Error())
		os.Exit(1)
	}

	err = ecsClientInstance.RegisterUpdateEventCallbackFunc(ecsConfig, func(innerOptionsUpdateError error) {
		if innerOptionsUpdateError != nil {
			fmt.Printf("Received error. %v", innerOptionsUpdateError.Error())
			return
		}

		fmt.Printf("Received config update event")
	})
	if err != nil {
		fmt.Printf("Registering event callback failed with error: %v", err.Error())
		os.Exit(1)
	}

	ecsConfigString, err := json.Marshal(*ecsConfig)
	if err != nil {
		fmt.Printf("%v", err.Error())
		os.Exit(1)
	}

	fmt.Printf("Received configuration: \"%v\"\n", string(ecsConfigString))
}
