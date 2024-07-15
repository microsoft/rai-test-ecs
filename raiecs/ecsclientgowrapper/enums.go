package ecsclientgowrapper

// Enumeration representing the different types of ECS environments.
type ECS_ENVIRONMENT_TYPE int

const (
	// Integration environment.
	ECS_ENVIRONMENT_TYPE_INTEGRATION ECS_ENVIRONMENT_TYPE = iota

	// Production environment.
	ECS_ENVIRONMENT_TYPE_PRODUCTION

	// Department of Defense environment.
	ECS_ENVIRONMENT_TYPE_DOD

	// Government Cloud Computing High environment.
	ECS_ENVIRONMENT_TYPE_GCCH

	// AG08 environment.
	ECS_ENVIRONMENT_TYPE_AG08

	// AG09 environment.
	ECS_ENVIRONMENT_TYPE_AG09

	// Mooncake environment.
	ECS_ENVIRONMENT_TYPE_MOONCAKE

	// Government Cloud Computing Moderate/Low environment.
	ECS_ENVIRONMENT_TYPE_GCCMOD ECS_ENVIRONMENT_TYPE = 8
)

// Enumeration representing the event codes returned by the ECS API functions.
type ECS_EVENT_TYPE int

const (
	// Successfully loaded configuration data.
	ECS_EVENT_CONFIGURATION_CHANGED ECS_EVENT_TYPE = iota

	// Successfully loaded configuration data from cache.
	ECS_EVENT_CONFIGURATION_CHANGED_FROM_CACHE

	// Configuration error.
	ECS_EVENT_CONFIGURATION_ERROR
)

// Enumeration representing the log levels used by the ECS API functions.
type ECS_LOG_LEVEL int

const (
	// Specifies that a logging category should not write any messages.
	ECS_LOG_LEVEL_NONE ECS_LOG_LEVEL = 0

	// Logs that track the general flow of the application. These logs should have long-term value.
	ECS_LOG_LEVEL_INFORMATION ECS_LOG_LEVEL = 2

	// Logs that highlight an abnormal or unexpected event in the application flow, but do not otherwise cause the application execution to stop.
	ECS_LOG_LEVEL_WARNING ECS_LOG_LEVEL = 3

	// Logs that highlight when the current flow of execution is stopped due to a failure.
	ECS_LOG_LEVEL_ERROR ECS_LOG_LEVEL = 4

	// Logs that describe an unrecoverable application or system crash, or a catastrophic failure that requires immediate attention.
	ECS_LOG_LEVEL_CRITICAL ECS_LOG_LEVEL = 5
)

// Enumeration representing the method to be used as authentication for ECS Config Service requests.
type ECS_AUTHENTICATION_METHOD int

const (
	// Specifies that no authentication should be used.
	ECS_AUTHENTICATION_METHOD_NONE ECS_AUTHENTICATION_METHOD = 0

	// Azure Application that uses as credential a X509Certificate2 registered to validate Subject Name and Issuer.
	ECS_AUTHENTICATION_METHOD_AZUREADCLIENTCERTIFICATEWITHSNI ECS_AUTHENTICATION_METHOD = 2

	// System Assigned Managed Identity. See also https://learn.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/overview#managed-identity-types .
	ECS_AUTHENTICATION_METHOD_SYSTEMASSIGNEDMANAGEDIDENTITY ECS_AUTHENTICATION_METHOD = 3

	// User Assigned Managed Identity. See also https://learn.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/overview#managed-identity-types .
	ECS_AUTHENTICATION_METHOD_USERASSIGNEDMANAGEDIDENTITY ECS_AUTHENTICATION_METHOD = 4
)
