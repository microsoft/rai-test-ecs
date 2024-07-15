package ecsgoclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/raiecs/ecsclientgowrapper"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	validConfigUpdate1 = `
	{
		"TestProjectTeam":
		{
			"ConfigName":
			{
				"TestProperty": "TestValue1",
				"TestIntegerWithMaxValue100": 1
			}
		},
		"Headers":
		{
			"ETag": "someEtag",
			"Expires": "Tue, 19 Dec 2023 13:01:09 GMT",
			"CountryCode": null,
			"StatusCode": "200"
		},
		"ConfigIDs":
		{
			"TestProjectTeam": "P-D-1129197-1-172"
		}
	}
	`
	validConfigUpdate2 = `
	{
		"TestProjectTeam":
		{
			"ConfigName":
			{
				"TestProperty": "TestValue2",
				"TestIntegerWithMaxValue100": 2
			}
		},
		"Headers":
		{
			"ETag": "someEtag",
			"Expires": "Tue, 19 Dec 2023 13:01:09 GMT",
			"CountryCode": null,
			"StatusCode": "200"
		},
		"ConfigIDs":
		{
			"TestProjectTeam": "P-D-1129197-1-172"
		}
	}
	`
	invalidConfigUpdate = `
	{
		"TestProjectTeam":
		{
			"ConfigName":
			{
				"TestProperty": "TestValue",
				"TestIntegerWithMaxValue100": 200
			}
		},
		"Headers":
		{
			"ETag": "someEtag",
			"Expires": "Tue, 19 Dec 2023 13:01:09 GMT",
			"CountryCode": null,
			"StatusCode": "200"
		},
		"ConfigIDs":
		{
			"TestProjectTeam": "P-D-1129197-1-172"
		}
	}
	`
)

type mockConfigGetter struct {
	mock.Mock
}

func (mockConfigGetter *mockConfigGetter) GetConfig(ecsRequestIdentifiers ecsclientgowrapper.EcsRequestIdentifiers) (string, error) {
	args := mockConfigGetter.Called(ecsRequestIdentifiers)
	return args.String(0), args.Error(1)
}

type TestConfig struct {
	TestProperty string `json:"TestProperty"`

	TestIntegerWithMaxValue100 int `json:"TestIntegerWithMaxValue100"`
}

func (testConfig *TestConfig) OnOptionsUpdateReceived(bytes []byte) error {
	var parsedConfig TestConfig
	if err := json.Unmarshal(bytes, &parsedConfig); err != nil {
		return fmt.Errorf("failed to unmarshal options, err: %v", err)
	}

	if err := parsedConfig.Validate(); err != nil {
		return err
	}

	*testConfig = parsedConfig
	return nil
}

func (testConfig TestConfig) Validate() error {
	if testConfig.TestIntegerWithMaxValue100 > 100 {
		return errors.New("test config validation failed")
	}

	return nil
}

type NoopLogger struct {
}

func (ecsLogger *NoopLogger) Log(logLevel ecsclientgowrapper.ECS_LOG_LEVEL, msg string) {
}

// Test receiving the initial config
func TestEcsGoClientInitialConfigReceived(t *testing.T) {
	ecsConfigGetter := mockConfigGetter{}
	ecsConfigGetter.On("GetConfig", mock.Anything).Return(validConfigUpdate1, nil)

	ecsClientInstance := NewEcsClientFromConfigGetter(&ecsConfigGetter, &NoopLogger{})

	testConfig := &TestConfig{}

	err := ecsClientInstance.AddOptionsMonitorToEcsClient(testConfig, "TestProjectTeam", "ConfigName")
	require.NoError(t, err)
	require.Equal(t, "TestValue1", testConfig.TestProperty)
	require.Equal(t, 1, testConfig.TestIntegerWithMaxValue100)
}

// Tests that an error is returned by the validation if the received config does not pass the validation
func TestEcsGoClientInvalidInitialConfigReceived(t *testing.T) {
	ecsConfigGetter := mockConfigGetter{}
	ecsConfigGetter.On("GetConfig", mock.Anything).Return(invalidConfigUpdate, nil)

	ecsClientInstance := NewEcsClientFromConfigGetter(&ecsConfigGetter, &NoopLogger{})

	testConfig := &TestConfig{}
	err := ecsClientInstance.AddOptionsMonitorToEcsClient(testConfig, "TestProjectTeam", "ConfigName")
	require.Error(t, err)
	require.Equal(t, "", testConfig.TestProperty)
}

// Tests that an error in getting the config is forwarded
func TestEcsGoClientErrorReceived(t *testing.T) {
	ecsConfigGetter := mockConfigGetter{}
	ecsConfigGetter.On("GetConfig", mock.Anything).Return("", errors.New("some error"))

	ecsClientInstance := NewEcsClientFromConfigGetter(&ecsConfigGetter, &NoopLogger{})

	testConfig := &TestConfig{}

	err := ecsClientInstance.AddOptionsMonitorToEcsClient(testConfig, "TestProjectTeam", "ConfigName")
	require.Error(t, err)
	require.Equal(t, "", testConfig.TestProperty)
}

// Tests that the initial config and a config update event are received
func TestEcsGoClientConfigUpdateReceived(t *testing.T) {
	ecsConfigGetter := mockConfigGetter{}
	configUpdateEvent1 := ecsConfigGetter.On("GetConfig", mock.Anything).Return(validConfigUpdate1, nil)

	ecsClientInstance := NewEcsClientFromConfigGetter(&ecsConfigGetter, &NoopLogger{})

	testConfig := &TestConfig{}

	err := ecsClientInstance.AddOptionsMonitorToEcsClient(testConfig, "TestProjectTeam", "ConfigName")

	require.NoError(t, err)
	require.Equal(t, "TestValue1", testConfig.TestProperty)
	require.Equal(t, 1, testConfig.TestIntegerWithMaxValue100)

	var optionsUpdateError error
	err = ecsClientInstance.RegisterUpdateEventCallbackFunc(testConfig, func(innerOptionsUpdateError error) {
		optionsUpdateError = innerOptionsUpdateError
	})
	require.NoError(t, err)

	configUpdateEvent1.Unset()
	ecsConfigGetter.On("GetConfig", mock.Anything).Return(validConfigUpdate2, nil)

	ecsClientInstance.invokeOptionsUpdate(false)
	require.NoError(t, optionsUpdateError)
	require.Equal(t, "TestValue2", testConfig.TestProperty)
	require.Equal(t, 2, testConfig.TestIntegerWithMaxValue100)
}

// Tests that a valid initial config and an invalid config update event are received
// The invalid config should be rejected and initial config kept
func TestEcsGoClientInvalidConfigUpdateReceived(t *testing.T) {
	ecsConfigGetter := mockConfigGetter{}
	configUpdateEvent1 := ecsConfigGetter.On("GetConfig", mock.Anything).Return(validConfigUpdate1, nil)

	ecsClientInstance := NewEcsClientFromConfigGetter(&ecsConfigGetter, &NoopLogger{})

	testConfig := &TestConfig{}

	err := ecsClientInstance.AddOptionsMonitorToEcsClient(testConfig, "TestProjectTeam", "ConfigName")
	require.NoError(t, err)
	require.Equal(t, "TestValue1", testConfig.TestProperty)
	require.Equal(t, 1, testConfig.TestIntegerWithMaxValue100)

	var optionsUpdateError error
	err = ecsClientInstance.RegisterUpdateEventCallbackFunc(testConfig, func(innerOptionsUpdateError error) {
		optionsUpdateError = innerOptionsUpdateError
	})
	require.NoError(t, err)

	configUpdateEvent1.Unset()
	ecsConfigGetter.On("GetConfig", mock.Anything).Return(invalidConfigUpdate, nil)

	ecsClientInstance.invokeOptionsUpdate(false)
	require.Error(t, optionsUpdateError)
	require.Equal(t, "TestValue1", testConfig.TestProperty)
	require.Equal(t, 1, testConfig.TestIntegerWithMaxValue100)
}

// Tests that a valid initial config and an error on config update event are received
// The error should be propagated and initial config should be kept
func TestEcsGoClientConfigUpdateErrorReceived(t *testing.T) {
	ecsConfigGetter := mockConfigGetter{}
	configUpdateEvent1 := ecsConfigGetter.On("GetConfig", mock.Anything).Return(validConfigUpdate1, nil)

	ecsClientInstance := NewEcsClientFromConfigGetter(&ecsConfigGetter, &NoopLogger{})

	testConfig := &TestConfig{}
	err := ecsClientInstance.AddOptionsMonitorToEcsClient(testConfig, "TestProjectTeam", "ConfigName")
	require.NoError(t, err)
	require.Equal(t, "TestValue1", testConfig.TestProperty)
	require.Equal(t, 1, testConfig.TestIntegerWithMaxValue100)

	var optionsUpdateError error
	err = ecsClientInstance.RegisterUpdateEventCallbackFunc(testConfig, func(innerOptionsUpdateError error) {
		optionsUpdateError = innerOptionsUpdateError
	})
	require.NoError(t, err)

	configUpdateEvent1.Unset()
	ecsConfigGetter.On("GetConfig", mock.Anything).Return("", errors.New("some error"))

	ecsClientInstance.invokeOptionsUpdate(false)
	require.Error(t, optionsUpdateError)
	require.Equal(t, "TestValue1", testConfig.TestProperty)
	require.Equal(t, 1, testConfig.TestIntegerWithMaxValue100)
}

// Tests that the all callback events are triggered as expected
func TestEcsGoClientConfigUpdateEventsCallback(t *testing.T) {
	ecsConfigGetter := mockConfigGetter{}
	configUpdateEvent1 := ecsConfigGetter.On("GetConfig", mock.Anything).Return(validConfigUpdate1, nil)

	ecsClientInstance := NewEcsClientFromConfigGetter(&ecsConfigGetter, &NoopLogger{})

	testConfig := &TestConfig{}
	err := ecsClientInstance.AddOptionsMonitorToEcsClient(testConfig, "TestProjectTeam", "ConfigName")
	require.NoError(t, err)

	configUpdateCounter := 0
	err = ecsClientInstance.RegisterUpdateEventCallbackFunc(testConfig, func(innerOptionsUpdateError error) {
		if innerOptionsUpdateError == nil {
			configUpdateCounter++
		}
	})
	require.NoError(t, err)

	configUpdateEvent1.Unset()
	configUpdateEvent2 := ecsConfigGetter.On("GetConfig", mock.Anything).Return(validConfigUpdate2, nil)

	ecsClientInstance.invokeOptionsUpdate(false)
	require.Equal(t, 1, configUpdateCounter)

	configUpdateEvent2.Unset()
	ecsConfigGetter.On("GetConfig", mock.Anything).Return(validConfigUpdate1, nil)

	ecsClientInstance.invokeOptionsUpdate(false)
	require.Equal(t, 2, configUpdateCounter)
}

// Tests that the all callback events are triggered as expected, and errors/invalid configs are triggering an error
func TestEcsGoClientConfigUpdateEventsCallbackWithErrors(t *testing.T) {
	ecsConfigGetter := mockConfigGetter{}
	configUpdateEvent1 := ecsConfigGetter.On("GetConfig", mock.Anything).Return(validConfigUpdate1, nil)

	ecsClientInstance := NewEcsClientFromConfigGetter(&ecsConfigGetter, &NoopLogger{})

	testConfig := &TestConfig{}
	err := ecsClientInstance.AddOptionsMonitorToEcsClient(testConfig, "TestProjectTeam", "ConfigName")
	require.NoError(t, err)

	configUpdateCounter := 0
	err = ecsClientInstance.RegisterUpdateEventCallbackFunc(testConfig, func(innerOptionsUpdateError error) {
		if innerOptionsUpdateError == nil {
			configUpdateCounter++
		}
	})
	require.NoError(t, err)

	configUpdateEvent1.Unset()
	configUpdateEvent2 := ecsConfigGetter.On("GetConfig", mock.Anything).Return(invalidConfigUpdate, nil)

	ecsClientInstance.invokeOptionsUpdate(false)
	require.Equal(t, 0, configUpdateCounter)

	configUpdateEvent2.Unset()
	configUpdateEvent3 := ecsConfigGetter.On("GetConfig", mock.Anything).Return("", errors.New("some error"))

	ecsClientInstance.invokeOptionsUpdate(false)
	require.Equal(t, 0, configUpdateCounter)

	configUpdateEvent3.Unset()
	ecsConfigGetter.On("GetConfig", mock.Anything).Return(validConfigUpdate2, nil)

	ecsClientInstance.invokeOptionsUpdate(false)
	require.Equal(t, 1, configUpdateCounter)
}

// Checks that a config update with the same content does not trigger the callback function
func TestEcsGoClientConfigUpdateEventsCallbackSameConfig(t *testing.T) {
	ecsConfigGetter := mockConfigGetter{}
	ecsConfigGetter.On("GetConfig", mock.Anything).Return(validConfigUpdate1, nil)

	ecsClientInstance := NewEcsClientFromConfigGetter(&ecsConfigGetter, &NoopLogger{})

	testConfig := &TestConfig{}

	err := ecsClientInstance.AddOptionsMonitorToEcsClient(testConfig, "TestProjectTeam", "ConfigName")
	require.NoError(t, err)

	configUpdateCounter := 0
	err = ecsClientInstance.RegisterUpdateEventCallbackFunc(testConfig, func(innerOptionsUpdateError error) {
		if innerOptionsUpdateError == nil {
			configUpdateCounter++
		}
	})
	require.NoError(t, err)

	ecsClientInstance.invokeOptionsUpdate(false)
	require.Equal(t, 0, configUpdateCounter)

	ecsClientInstance.invokeOptionsUpdate(false)
	require.Equal(t, 0, configUpdateCounter)
}
