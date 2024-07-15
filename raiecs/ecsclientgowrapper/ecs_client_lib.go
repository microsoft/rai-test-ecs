package ecsclientgowrapper

import (
	"unsafe"
)

/*
#cgo CFLAGS: -I${SRCDIR}/libs/win-x64
#cgo LDFLAGS: -L${SRCDIR}/libs/win-x64 -Wl,-rpath=${SRCDIR}/libs/win-x64 -lecsclient
#cgo linux CFLAGS: -I${SRCDIR}/libs/linux-x64
#cgo linux LDFLAGS: -L${SRCDIR}/libs/linux-x64 -Wl,-rpath=${SRCDIR}/libs/linux-x64 -lecsclient
#include <stdio.h>
#include <stdlib.h>
#include <ecsclient.h>
*/
import "C"

// Represents an ECS client instance.
type EcsClient struct {
	ecsClientHandle *C.EcsClientHandle
}

// Represents the function signature of callback functions that can be registered.
type EcsConfigurationEventCallbackFunc func(ECS_EVENT_TYPE, string)

// Structure representing a key-value pair for request identifiers.
type EcsRequestIdentifier struct {
	// The key/name of the request identifier as UTF-8 string.
	Name string

	// The value associated with the request identifier as UTF-8 strings.
	Values []string
}

// Request identifiers, typically service level context (e.g. environment, region, etc.).
type EcsRequestIdentifiers []EcsRequestIdentifier

// Structure representing options for configuring ECS client.
type EcsClientOptions struct {
	// Path to default configurations.
	DefaultConfigPath string

	// Path to default groups.
	DefaultGroupsPath string

	// Default request identifiers, typically service level context (e.g. environment, region, etc.).
	DefaultRequestIdentifiers EcsRequestIdentifiers

	// X509 certificate for authentication. Should be raw byte array of X.509 in PKCS #12 format (PFX) with private key.
	X509Cert []byte

	// TenantId if using Azure AD app authentication via SN/I. If NULL defaults to Torus tenant specific to ECS client initialized environment.
	TenantId string

	// Client ID if using Azure AD app authentication via SN/I. If NULL but x509_cert is defined, plain MTLS will be used.
	ClientId string

	// Authentication environment override. Needed if GCCMod and AAD app is in Azure Government. If NULL defaults to ECS client initialized environment.
	AuthenticationEnvironment *ECS_ENVIRONMENT_TYPE

	// The method to be used as authentication for ECS Config Service requests.
	AuthenticationMethod ECS_AUTHENTICATION_METHOD

	// Callback function that should get called when the ecs configuration changes.
	EcsConfigurationEventCallbackFunc *EcsConfigurationEventCallbackFunc

	// The Logger
	Logger Logger

	// Log level for logging messages.
	LogLevel ECS_LOG_LEVEL

	// Enable A&E ExP Control Tower based flighting for Cerberus.
	EnableExp int
}

// Logger interface
type Logger interface {
	Log(logLevel ECS_LOG_LEVEL, msg string)
}

// CreateEcsClient instantiates a new EcsClient.
func CreateEcsClient(
	environment ECS_ENVIRONMENT_TYPE,
	client string,
	agents []string,
	clientOptions EcsClientOptions) (EcsClient, error) {
	cEnvironmentType := C.ECS_ENVIRONMENT_TYPE(C.int(int(environment)))

	cClient := C.CString(client)
	defer C.free(unsafe.Pointer(cClient))

	cAgents, cAgentsLen := cStringArray(agents...)
	defer freeCStringArray(cAgents, cAgentsLen)

	cClientOptions := cEcsClientOptions(clientOptions)
	defer freeCEcsClientOptions(cClientOptions)

	ecs_client_handle := (*C.EcsClientHandle)(C.malloc(C.sizeof_EcsClientHandle))

	status_code := C.ecs_create_client(cEnvironmentType, cClient, cAgents, cAgentsLen, cClientOptions, ecs_client_handle)

	return EcsClient{ecsClientHandle: ecs_client_handle}, statusCodeToError(status_code)
}

// GetConfig fetches the ECS config.
func (ecsClient EcsClient) GetConfig(ecsRequestIdentifiers EcsRequestIdentifiers) (string, error) {
	crequestIdentifiers, cRequestIdentifiersLen := cEcsRequestIdentifiers(ecsRequestIdentifiers)
	defer freeCEcsRequestIdentifiers(crequestIdentifiers, cRequestIdentifiersLen)

	var out_config *C.char
	defer func() {
		_ = C.ecs_free_str(out_config)
	}()

	status_code := C.ecs_client_get_config(*ecsClient.ecsClientHandle, crequestIdentifiers, cRequestIdentifiersLen, &out_config)
	go_out_config := C.GoString(out_config)

	return go_out_config, statusCodeToError(status_code)
}

// DestroyClient destroys the EcsClient instance.
func (ecsClient EcsClient) DestroyClient() error {
	return statusCodeToError(C.ecs_destroy_client(*ecsClient.ecsClientHandle))
}
