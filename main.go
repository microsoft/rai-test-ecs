package main

import (
	"fmt"
	"time"

	ecsgoclient "github.com/raiecs"
	"github.com/raiecs/ecsclientgowrapper"
)

func main() {
	fmt.Println("Hello!")
	fmt.Println("Creating ECS Client Options")
	env := ecsclientgowrapper.ECS_ENVIRONMENT_TYPE_INTEGRATION
	options := ecsgoclient.EcsClientOptions{
		Client:       "AzureService",
		ProjectTeams: []string{"ResponsibleAI"},
		TargetFilters: map[string][]string{
			"EnvironmentName": {"Pre-Test"},
			"ServiceName":     {"RAI-ECS-Test"},
		},
		Logger:                    &ECSLogger{},
		LogLevel:                  ecsclientgowrapper.ECS_LOG_LEVEL_INFORMATION,
		AuthenticationEnvironment: &env,
		AuthenticationMethod:      ecsclientgowrapper.ECS_AUTHENTICATION_METHOD_USERASSIGNEDMANAGEDIDENTITY,
		TenantId:                  "72f988bf-86f1-41af-91ab-2d7cd011db47",
		ClientId:                  "5aebd2f3-8d5c-45aa-ab66-087ca711c3c4",
	}

	var err error
	var ecsClient *ecsgoclient.EcsClient
	fmt.Println("Creating ECS Client")
	ecsClient, err = ecsgoclient.NewEcsClient(options)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Step 2: Get new options registered with ECS
	fmt.Println("Registering Monitor")
	config_obj := registerOptionsMonitor(ecsClient)
	old_config := Config{config: config_obj.config}
	for {
		fmt.Println("Waiting for config...")
		if old_config.config != config_obj.config {
			fmt.Printf("New Config Recieved: %s", config_obj.config)
			old_config.config = config_obj.config
		}
		time.Sleep(15 * time.Second)
	}

}

type ECSLogger struct{}

func (ecslogger *ECSLogger) Log(logLevel ecsclientgowrapper.ECS_LOG_LEVEL, msg string) {
	fmt.Printf("Received log. %v: %v \n", logLevel, msg)
}

type Config struct {
	config string
}

func (o *Config) OnOptionsUpdateReceived(bytes []byte) error {
	parsedOptions := string(bytes)
	o.config = parsedOptions
	return nil
}

func registerOptionsMonitor(client *ecsgoclient.EcsClient) *Config {
	options := &Config{config: ""}
	client.AddOptionsMonitorToEcsClient(options, "ResponsibleAI", "SampleOptions")
	return options
}
