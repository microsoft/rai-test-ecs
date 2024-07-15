/**
 * @file ecsclient.h
 * @brief This is a C API exposing a subset of the Microsoft.Skype.ECS.Client .NET library compiled with NativeAOT.
 */

#ifndef ECSCLIENT_H
#define ECSCLIENT_H

/* Define DYNAMIC_LOAD to enable dynamic loading of the library.
 * Call ecs_load_library(path) before using any of the functions.
 */
#ifdef ECSCLIENT_DYNAMIC_LOAD

#ifdef _WIN32
#include "windows.h"
#define symLoad GetProcAddress
#else
#include "dlfcn.h"
#include <unistd.h>
#define symLoad dlsym
#endif

#ifdef _WIN32
HINSTANCE ecsclient_handle;
#else
void* ecsclient_handle;
#endif

void ecs_load_library(const char* path) {
#ifdef _WIN32
	ecsclient_handle = LoadLibraryA(path);
#else
	ecsclient_handle = dlopen(path, RTLD_LAZY);
#endif
}

#endif /* ECSCLIENT_DYNAMIC_LOAD */

 /**
  * @enum ECS_ENVIRONMENT_TYPE
  * @brief Enumeration representing the different types of ECS environments.
  */
typedef enum
{
	ECS_ENVIRONMENT_TYPE_INTEGRATION = 0, /**< Integration environment. */
	ECS_ENVIRONMENT_TYPE_PRODUCTION = 1,  /**< Production environment. */
	ECS_ENVIRONMENT_TYPE_DOD = 2,         /**< Department of Defense environment. */
	ECS_ENVIRONMENT_TYPE_GCCH = 3,        /**< Government Cloud Computing High environment. */
	ECS_ENVIRONMENT_TYPE_AG08 = 4,        /**< AG08 environment. */
	ECS_ENVIRONMENT_TYPE_AG09 = 5,        /**< AG09 environment. */
	ECS_ENVIRONMENT_TYPE_MOONCAKE = 6,    /**< Mooncake environment. */
	ECS_ENVIRONMENT_TYPE_GCCMOD = 8,      /**< Government Cloud Computing Moderate/Low environment. */
	ECS_ENVIRONMENT_TYPE_CANARY = 9       /**< Canary environment. */
} ECS_ENVIRONMENT_TYPE;

/**
 * @enum ECS_STATUS_CODE
 * @brief Enumeration representing the status codes returned by the ECS API functions.
 */
typedef enum {
	ECS_STATUS_SUCCESS = 0,          /**< Success status code. */
	ECS_STATUS_ERROR_UNDEFINED = -1, /**< Undefined error status code. */
} ECS_STATUS_CODE;

/**
 * @enum ECS_EVENT_CODE
 * @brief Enumeration representing the event codes returned by the ECS API functions.
 */
typedef enum {
	ECS_EVENT_CONFIGURATION_CHANGED = 0,            /**< Successfully loaded configuration data. */
	ECS_EVENT_CONFIGURATION_CHANGED_FROM_CACHE = 1, /**< Successfully loaded configuration data from cache. */
	ECS_EVENT_CONFIGURATION_ERROR = 2,              /**< Error loading configuration data. */
} ECS_EVENT_CODE;

/**
 * @enum ECS_LOG_LEVEL
 * @brief Enumeration representing the log levels returned by the ECS API functions.
 */
typedef enum {
	ECS_LOG_LEVEL_NONE = 0,        /**< Specifies that a logging category should not write any messages. */
	ECS_LOG_LEVEL_INFORMATION = 2, /**< Logs that track the general flow of the application. These logs should have long-term value. */
	ECS_LOG_LEVEL_WARNING = 3,     /**< Logs that highlight an abnormal or unexpected event in the application flow, but do not otherwise cause the application execution to stop. */
	ECS_LOG_LEVEL_ERROR = 4,       /**< Logs that highlight when the current flow of execution is stopped due to a failure.  */
	ECS_LOG_LEVEL_CRITICAL = 5,    /**< Logs that describe an unrecoverable application or system crash, or a catastrophic failure that requires immediate attention. */
} ECS_LOG_LEVEL;

/**
 * @enum ECS_AUTHENTICATION_METHOD
 * @brief Enumeration representing the method to be used as authentication for ECS Config Service requests.
 */
typedef enum {
	ECS_AUTHENTICATION_METHOD_NONE = 0,        /**< Specifies that no authentication should be used. */
	ECS_AUTHENTICATION_METHOD_AZUREADCLIENTCERTIFICATEWITHSNI = 2, /**< Azure Application that uses as credential a X509Certificate2 registered to validate Subject Name and Issuer. */
	ECS_AUTHENTICATION_METHOD_SYSTEMASSIGNEDMANAGEDIDENTITY = 3,     /**< System Assigned Managed Identity. See also https://learn.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/overview#managed-identity-types . */
	ECS_AUTHENTICATION_METHOD_USERASSIGNEDMANAGEDIDENTITY = 4,       /**< User Assigned Managed Identity. See also https://learn.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/overview#managed-identity-types . */
} ECS_AUTHENTICATION_METHOD;

/**
 * @struct EcsRequestIdentifier
 * @brief Structure representing a key-value pair for request identifiers.
 */
typedef struct {
	const char* name;  /**< The key/name of the request identifier as UTF-8 string. */
	const char** values; /**< The value associated with the request identifier as UTF-8 string. */
	int values_length;  /**< The length of the values array. */
} EcsRequestIdentifier;

/**
 * @typedef EcsClientHandle
 * @brief Opaque handle representing an ECS client instance.
 */
typedef void* EcsClientHandle;

/**
 * @typedef EcsConfigurationEventCallbackFunc
 * @brief ECS client event callback function signature.
 *
 * @param ecs_client_handle The ECS client handle.
 * @param status_code The ECS_EVENT_CODE indicating event type.
 * @param event_message A UTF-8 string containing more detailed status. This may be NULL. The callback does not own the string and should not free it.
 */
typedef void (*EcsConfigurationEventCallbackFunc)(EcsClientHandle, ECS_EVENT_CODE, const char*);

/**
 * @typedef EcsClientLogCallbackFunc
 * @brief ECS client event callback function signature.
 *
 * @param ecs_client_handle The ECS client handle.
 * @param status_code The ECS_LOG_LEVEL indicating log level.
 * @param event_message A UTF-8 string containing log message. The callback does not own the string and should not free it.
 */
typedef void (*EcsClientLogCallbackFunc)(ECS_LOG_LEVEL, const char*);

/**
 * @struct EcsClientOptions
 * @brief Structure representing options for configuring ECS client.
 */
typedef struct {
	const char* default_config_path; /**< Path to default configurations. */
	const char* default_groups_path; /**< Path to default groups. */
	const EcsRequestIdentifier* default_request_identifiers; /**< Default request identifiers, typically service level context (e.g. environment, region, etc.). */
	int default_request_identifiers_length; /**< Length of default_request_identifiers. */
	const unsigned char* x509_cert; /**< X509 certificate for authentication. Should be raw byte array of X.509 in PKCS #12 format (PFX) with private key. If using Windows this can also be a PCCERT_CONTEXT. */
	int x509_cert_length; /**< Length of x509_cert. If x509_cert is PCCERT_CONTEXT, then this must be zero. */
	const char* tenant_id; /**< Tenant ID if using Azure AD app authentication via SN/I. If NULL defaults to Torus tenant specific to ECS client initialized environment. */
	const char* client_id; /**< Client ID if using Azure AD app authentication via SN/I. If NULL but x509_cert is defined, plain MTLS will be used. */
	ECS_AUTHENTICATION_METHOD authentication_method; /**< The method to be used as authentication for ECS Config Service requests. */
	ECS_ENVIRONMENT_TYPE* auth_env; /**< Authentication environment override. Needed if GCCMod and AAD app is in Azure Government. If NULL defaults to ECS client initialized environment. */
	EcsConfigurationEventCallbackFunc event_callback; /**< ECS client event callback function. */
	EcsClientLogCallbackFunc log_callback; /**< ECS client log callback function. */
	ECS_LOG_LEVEL log_level; /**< ECS client log level. It's none by default. */
	int enable_exp; /**< Enable A&E ExP Control Tower based flighting for Cerberus . */
} EcsClientOptions;

/**
 * @brief Creates an ECS client with the given environment type, client identifier, and agents.
 *
 * @param env The environment type for the ECS client.
 * @param client A UTF-8 string representing the client identifier.
 * @param agents An array of UTF-8 strings representing the agents for the ECS client.
 * @param agents_length The length of the agents array.
 * @param options The ECS client options.
 * @param[out] out_ecs_client_handle A pointer to receive the created ECS client handle.
 * @return ECS_STATUS_CODE indicating success or failure.
 */
ECS_STATUS_CODE ecs_create_client(
	ECS_ENVIRONMENT_TYPE env,
	const char* client,
	const char** agents,
	int agents_length,
	const EcsClientOptions* options,
	EcsClientHandle* out_ecs_client_handle);

#ifdef ECSCLIENT_DYNAMIC_LOAD
typedef ECS_STATUS_CODE(*EcsCreateClientFunc)(
	ECS_ENVIRONMENT_TYPE env,
	const char* client,
	const char** agents,
	int agents_length,
	const EcsClientOptions* options,
	EcsClientHandle* out_ecs_client_handle);

ECS_STATUS_CODE ecs_create_client(ECS_ENVIRONMENT_TYPE env, const char* client, const char** agents, int agents_length, const EcsClientOptions* options, EcsClientHandle* ecs_client_handle)
{
	EcsCreateClientFunc func = (EcsCreateClientFunc)symLoad(ecsclient_handle, "ecs_create_client");
	return func(env, client, agents, agents_length, options, ecs_client_handle);
}
#endif

/**
 * @brief Destroys the given ECS client instance.
 *
 * @param ecs_client_handle The ECS client handle to be destroyed.
 * @return ECS_STATUS_CODE indicating success or failure.
 */
ECS_STATUS_CODE ecs_destroy_client(
	EcsClientHandle ecs_client_handle);

#ifdef ECSCLIENT_DYNAMIC_LOAD
typedef ECS_STATUS_CODE(*EcsDestroyClientFunc)(
	EcsClientHandle ecs_client_handle);

ECS_STATUS_CODE ecs_destroy_client(EcsClientHandle client)
{
	EcsDestroyClientFunc func = (EcsDestroyClientFunc)symLoad(ecsclient_handle, "ecs_destroy_client");
	return func(client);
}
#endif

/**
 * @brief Retrieves the configuration for the given ECS client and request identifiers.
 *
 * @param ecs_client_handle The ECS client handle.
 * @param request_identifiers An array of EcsRequestIdentifier structures.
 * @param request_identifiers_length The length of the request_identifiers array.
 * @param[out] out_config A pointer to receive the UTF-8 string containing the configuration. It's the callers responsibility to free the string.
 * @return ECS_STATUS_CODE indicating success or failure.
 */
ECS_STATUS_CODE ecs_client_get_config(
	EcsClientHandle ecs_client_handle,
	const EcsRequestIdentifier* request_identifiers,
	int request_identifiers_length,
	char** out_config);

#ifdef ECSCLIENT_DYNAMIC_LOAD
typedef ECS_STATUS_CODE(*EcsClientGetConfigFunc)(
	EcsClientHandle ecs_client_handle,
	const EcsRequestIdentifier* request_identifiers,
	int request_identifiers_length,
	char** out_config);

ECS_STATUS_CODE ecs_client_get_config(EcsClientHandle client, const EcsRequestIdentifier* request_identifiers, int request_identifiers_length, char** out_config)
{
	EcsClientGetConfigFunc func = (EcsClientGetConfigFunc)symLoad(ecsclient_handle, "ecs_client_get_config");
	return func(client, request_identifiers, request_identifiers_length, out_config);
}
#endif

/**
 * @brief Frees the memory allocated for a UTF-8 string returned by the ECS API functions.
 *
 * @param str The UTF-8 string to be freed.
 * @return ECS_STATUS_CODE indicating success or failure.
 */
ECS_STATUS_CODE ecs_free_str(char* str);

#ifdef ECSCLIENT_DYNAMIC_LOAD
typedef ECS_STATUS_CODE(*EcsFreeStrFunc)(char* str);

ECS_STATUS_CODE ecs_free_str(char* str)
{
	EcsFreeStrFunc func = (EcsFreeStrFunc)symLoad(ecsclient_handle, "ecs_free_str");
	return func(str);
}
#endif

/**
 * @brief Get last error message for current thread as a UTF-8 string.
 * This will be populated anytime functions return ECS_STATUS_CODE that's not ECS_STATUS_SUCCESS.
 *
 * @return The UTF-8 string of last error message. This may be NULL. It's the callers responsibility to free the string.
 */
char* ecs_get_last_error();

#ifdef ECSCLIENT_DYNAMIC_LOAD
typedef char* (*EcsGetLastErrorFunc)();

char* ecs_get_last_error()
{
	EcsGetLastErrorFunc func = (EcsGetLastErrorFunc)symLoad(ecsclient_handle, "ecs_get_last_error");
	return func();
}
#endif

#endif /* ECSCLIENT_H */