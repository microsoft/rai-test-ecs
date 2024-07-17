package ecsgoclient

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/raiecs/ecsclientgowrapper"
)

// EcsUpdateEventCallbackFunc is a callback func the user can register and will be called if a config update has been received
type EcsUpdateEventCallbackFunc func(optionsUpdateError error)

// ecsOptionsUpdateFunc is an internal type to auto update options that are registered on the client
type ecsOptionsUpdateFunc func(config string, logger ecsclientgowrapper.Logger) (error, bool)

// EcsOptionsMonitor contains the info for the options monitor (the func to do the update of TOptions, and the registered callbacks if the TOptions update was invoked)
type EcsOptionsMonitor struct {
	optionsUpdateFunc  ecsOptionsUpdateFunc
	configUpdateEvents []EcsUpdateEventCallbackFunc
}

// EcsConfigGetter is the interface that is internally used for fetching the config from ECS
type EcsConfigGetter interface {
	GetConfig(ecsRequestIdentifiers ecsclientgowrapper.EcsRequestIdentifiers) (string, error)
}

// EcsClientOptions are the ECS configuration options that the ECS-GO-Client exposes
type EcsClientOptions struct {
	// The ECS client name
	Client string

	// The ECS project team names
	ProjectTeams []string

	// The target filters, typically service level context (e.g. environment, region, etc.).
	TargetFilters map[string][]string

	// The Logger
	Logger ecsclientgowrapper.Logger

	// Path to default configurations.
	DefaultConfigPath string

	// Path to default groups.
	DefaultGroupsPath string

	// X509 certificate for authentication. Should be raw byte array of X.509 in PKCS #12 format (PFX) with private key.
	X509Cert []byte

	// TenantId if using Azure AD app authentication via SN/I.
	TenantId string

	// Client ID if using Azure AD app authentication via SN/I.
	ClientId string

	// Use legacy ECS Azure AD application if using Azure AD app authentication via SN/I. If true, scope https://ecs.skype.com/.default will be used.
	UseLegacyApp int

	// Authentication environment override. Needed if GCCMod and AAD app is in Azure Government. If NULL defaults to ECS client initialized environment.
	AuthenticationEnvironment *ecsclientgowrapper.ECS_ENVIRONMENT_TYPE

	// The method to be used as authentication for ECS Config Service requests.
	AuthenticationMethod ecsclientgowrapper.ECS_AUTHENTICATION_METHOD

	// The min loglevel
	LogLevel ecsclientgowrapper.ECS_LOG_LEVEL

	// Enable A&E ExP Control Tower based flighting for Cerberus.
	EnableExp int
}

type EcsClient struct {
	internalEcsClient  EcsConfigGetter
	logger             ecsclientgowrapper.Logger
	ecsOptionMonitors  map[any]*EcsOptionsMonitor
	callbackFuncsMutex sync.RWMutex
}

type OptionsUpdateReceiver interface {
	OnOptionsUpdateReceived([]byte) error
}

// Request identifier names:
const (
	EnvironmentRequestIdentifierName = "EnvironmentName"
	ServiceRequestIdentifierName     = "ServiceName"
)

// NewEcsClient creates a new ecs client which calls into the ecs C library to fetch the config
func NewEcsClient(ecsClientOptions EcsClientOptions) (*EcsClient, error) {
	var callbackFunction ecsclientgowrapper.EcsConfigurationEventCallbackFunc = func(event ecsclientgowrapper.ECS_EVENT_TYPE, message string) {}

	targetFilters := make([]ecsclientgowrapper.EcsRequestIdentifier, len(ecsClientOptions.TargetFilters))

	idx := 0
	for key, values := range ecsClientOptions.TargetFilters {
		targetFilters[idx] = ecsclientgowrapper.EcsRequestIdentifier{
			Name:   key,
			Values: values,
		}
		idx++
	}

	internalClientOptions := ecsclientgowrapper.EcsClientOptions{
		DefaultConfigPath:                 ecsClientOptions.DefaultConfigPath,
		DefaultGroupsPath:                 ecsClientOptions.DefaultGroupsPath,
		DefaultRequestIdentifiers:         targetFilters,
		X509Cert:                          ecsClientOptions.X509Cert,
		TenantId:                          ecsClientOptions.TenantId,
		ClientId:                          ecsClientOptions.ClientId,
		AuthenticationEnvironment:         ecsClientOptions.AuthenticationEnvironment,
		AuthenticationMethod:              ecsClientOptions.AuthenticationMethod,
		EcsConfigurationEventCallbackFunc: &callbackFunction,
		Logger:                            ecsClientOptions.Logger,
		LogLevel:                          ecsClientOptions.LogLevel,
		EnableExp:                         ecsClientOptions.EnableExp,
	}

	internalClient, err := ecsclientgowrapper.CreateEcsClient(ecsclientgowrapper.ECS_ENVIRONMENT_TYPE_PRODUCTION, ecsClientOptions.Client, ecsClientOptions.ProjectTeams, internalClientOptions)
	if err != nil {
		return nil, err
	}

	ecsClient := &EcsClient{
		internalEcsClient: internalClient,
		logger:            ecsClientOptions.Logger,
		ecsOptionMonitors: make(map[any]*EcsOptionsMonitor),
	}

	callbackFunction = func(event ecsclientgowrapper.ECS_EVENT_TYPE, message string) {
		ecsClient.invokeOptionsUpdate(false)
	}

	return ecsClient, nil
}

// NewEcsClient creates a new ecs client that fetches config from EcsConfigGetter - useful for mocking/testing
func NewEcsClientFromConfigGetter(ecsConfigGetter EcsConfigGetter, logger ecsclientgowrapper.Logger) *EcsClient {
	return &EcsClient{
		internalEcsClient: ecsConfigGetter,
		logger:            logger,
		ecsOptionMonitors: make(map[any]*EcsOptionsMonitor),
	}
}

func (ecsClient *EcsClient) TriggerAllUpdateEventCallbacks() {
	ecsClient.callbackFuncsMutex.Lock()
	defer ecsClient.callbackFuncsMutex.Unlock()

	for _, listener := range ecsClient.ecsOptionMonitors {
		for _, fn := range listener.configUpdateEvents {
			fn(nil)
		}
	}
}

func (ecsClient *EcsClient) invokeOptionsUpdate(isInitialUpdate bool) {
	config, err := ecsClient.internalEcsClient.GetConfig(ecsclientgowrapper.EcsRequestIdentifiers{})
	if err != nil {
		ecsClient.logger.Log(ecsclientgowrapper.ECS_LOG_LEVEL_ERROR, "updating config failed")

		ecsClient.callbackFuncsMutex.RLock()
		defer ecsClient.callbackFuncsMutex.RUnlock()
		for _, listener := range ecsClient.ecsOptionMonitors {
			for _, fn := range listener.configUpdateEvents {
				fn(err)
			}
		}

		return
	}

	ecsClient.callbackFuncsMutex.RLock()
	defer ecsClient.callbackFuncsMutex.RUnlock()
	for _, listener := range ecsClient.ecsOptionMonitors {
		err, updatedOptions := listener.optionsUpdateFunc(config, ecsClient.logger)
		if err != nil {
			ecsClient.logger.Log(ecsclientgowrapper.ECS_LOG_LEVEL_ERROR, fmt.Sprintf("on options update func failed with error: %v", err))
			for _, fn := range listener.configUpdateEvents {
				fn(err)
			}
			continue
		}

		if updatedOptions {
			if !isInitialUpdate {
				ecsClient.logger.Log(ecsclientgowrapper.ECS_LOG_LEVEL_INFORMATION, "Received ECS update - calling config update event")
			}

			for _, fn := range listener.configUpdateEvents {
				fn(err)
			}
		}
	}
}

// RegisterUpdateEventCallbackFunc registers another callback function to an options monitor for a certain option TOptions
func (ecsClient *EcsClient) RegisterUpdateEventCallbackFunc(options OptionsUpdateReceiver, configUpdateEvent EcsUpdateEventCallbackFunc) error {
	ecsClient.callbackFuncsMutex.Lock()
	defer ecsClient.callbackFuncsMutex.Unlock()

	if ecsUpdateListener, ok := ecsClient.ecsOptionMonitors[options]; ok {
		ecsUpdateListener.configUpdateEvents = append(ecsUpdateListener.configUpdateEvents, configUpdateEvent)
	} else {
		return fmt.Errorf("no OptionsMonitor for provided options are registered - configUpdateEvent would never get called")
	}
	return nil
}

// AddOptionsMonitorToEcsClient adds a TOptions struct to the ecsClient for monitoring. The ecsClient will update the values of the options and
// call the callback function whenever a config update has been registered.
func (ecsClient *EcsClient) AddOptionsMonitorToEcsClient(options OptionsUpdateReceiver, projectTeam string, optionName string) error {
	var checkSum *string
	var updateFunc ecsOptionsUpdateFunc = func(config string, logger ecsclientgowrapper.Logger) (error, bool) {
		var fullConfig map[string]interface{}
		if err := json.Unmarshal([]byte(config), &fullConfig); err != nil {
			return fmt.Errorf("failed to unmarshal ecs config"), false
		}
		logger.Log(ecsclientgowrapper.ECS_LOG_LEVEL_INFORMATION, "Full Config Keys: ")
		for key := range fullConfig {
			logger.Log(ecsclientgowrapper.ECS_LOG_LEVEL_INFORMATION, fmt.Sprint("Key: %s", key))
		}

		clientConfig, ok := fullConfig[projectTeam]
		if !ok {
			return fmt.Errorf("failed to find projectTeam property '%v'", projectTeam), false
		}

		typedClientConfig, ok := clientConfig.(map[string]interface{})
		if !ok {
			return fmt.Errorf("failed to parse property '%v'", projectTeam), false
		}

		optionConfig, ok := typedClientConfig[optionName]
		if !ok {
			return fmt.Errorf("failed to parse property '%v' in '%v'", optionName, projectTeam), false
		}

		jsonOpts, err := json.Marshal(optionConfig)
		if err != nil {
			return fmt.Errorf("failed to get bytes for options '%v', err: %v", optionName, err), false
		}

		newCheckSum, err := getCheckSum(jsonOpts)
		if err != nil {
			return err, false
		}

		// only do the update if we see that the checkSum has changed:
		if checkSum == nil || *checkSum != newCheckSum {
			err = options.OnOptionsUpdateReceived(jsonOpts)
			if err != nil {
				return fmt.Errorf("failed options update: '%v'", err), false
			}

			checkSum = &newCheckSum
			ecsClient.logger.Log(ecsclientgowrapper.ECS_LOG_LEVEL_INFORMATION, fmt.Sprintf("Received ECS config: %v", string(jsonOpts)))
			return nil, true
		}

		return nil, false
	}

	ecsClient.callbackFuncsMutex.Lock()
	if _, ok := ecsClient.ecsOptionMonitors[options]; ok {
		ecsClient.callbackFuncsMutex.Unlock()
		return fmt.Errorf("there is already an options monitor registered for the same options")
	}

	var initialCallbackError error
	initCallbackFunc := func(optionsUpdateError error) {
		if optionsUpdateError != nil {
			initialCallbackError = optionsUpdateError
			return
		}
	}

	ecsClient.ecsOptionMonitors[options] = &EcsOptionsMonitor{
		optionsUpdateFunc:  updateFunc,
		configUpdateEvents: []EcsUpdateEventCallbackFunc{initCallbackFunc},
	}

	ecsClient.callbackFuncsMutex.Unlock()
	ecsClient.invokeOptionsUpdate(true)

	return initialCallbackError
}

func getCheckSum(byteArr []byte) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, bytes.NewReader(byteArr)); err != nil {
		return "", err
	}

	sha256Sum := h.Sum(nil)
	return hex.EncodeToString(sha256Sum[:]), nil
}
