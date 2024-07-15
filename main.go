package main

import (
	"fmt"
	"time"

	ecsgoclient "github.com/raiecs"
	"github.com/raiecs/ecsclientgowrapper"
)

func main() {
	fmt.Println("Hello, World!")
	env := ecsclientgowrapper.ECS_ENVIRONMENT_TYPE_INTEGRATION
	options := ecsgoclient.EcsClientOptions{
		Client:       "ECS_Test_Agrawalsh",
		ProjectTeams: []string{"ResponsibleAI"},
		TargetFilters: map[string][]string{
			"EnvironmentName": {"YourEnvironment"},
			"ServiceName":     {"YourService"},
		},
		Logger:                    ECSLogger{},
		AuthenticationEnvironment: &env,
		AuthenticationMethod:      ecsclientgowrapper.ECS_AUTHENTICATION_METHOD_USERASSIGNEDMANAGEDIDENTITY,
		TenantId:                  "72f988bf-86f1-41af-91ab-2d7cd011db47",
		ClientId:                  "add30c60-c1ac-43e6-992a-52c4da308a92",
	}

	var err error
	var ecsClient *ecsgoclient.EcsClient
	ecsClient, err = ecsgoclient.NewEcsClient(options)

	if err != nil {
		// throw error + log
	}

	// Step 2: Get new options registered with ECS
	config_obj := registerOptionsMonitor(ecsClient)
	old_config := Config{config: config_obj.config}
	for {
		if old_config.config != config_obj.config {
			fmt.Printf("New Config Recieved: %s", config_obj.config)
			old_config.config = config_obj.config
		}
		time.Sleep(1 * time.Second)
	}

}

type ECSLogger struct{}

func (ecslogger ECSLogger) Log(logLevel ecsclientgowrapper.ECS_LOG_LEVEL, msg string) {
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
