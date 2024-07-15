package ecsclientgowrapper

import (
	"fmt"
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
void eventCallback(EcsClientHandle, ECS_EVENT_CODE, char*);
void logCallback(ECS_LOG_LEVEL, char*);
*/
import "C"

// eventCallback is a GO function that is called by the ecsclientlib on config changes.
//
//export eventCallback
func eventCallback(a C.EcsClientHandle, b C.ECS_EVENT_CODE, c *C.char) {
	if eventCallbackFunc != nil {
		eventType := ECS_EVENT_TYPE(int(b))
		message := C.GoString(c)

		(*eventCallbackFunc)(eventType, message)
	}
}

// logCallback is a GO function that is called by the ecsclientlib for each log message.
//
//export logCallback
func logCallback(a C.ECS_LOG_LEVEL, b *C.char) {
	if logger != nil {
		logLevel := ECS_LOG_LEVEL(int(a))
		logMessage := C.GoString(b)

		logger.Log(logLevel, logMessage)
	}
}

var eventCallbackFunc *EcsConfigurationEventCallbackFunc
var logger Logger

// cEcsClientOptions converts the Go EcsClientOptions to C.EcsClientOptions.
//
// Note that the C.EcsClientOptions are allocated in the C heap and therefore
// must be freed by calling freeCEcsClientOptions().
func cEcsClientOptions(clientOptions EcsClientOptions) *C.EcsClientOptions {
	cClientOptions := C.EcsClientOptions{}

	cClientOptions.default_config_path = C.CString(clientOptions.DefaultConfigPath)

	cClientOptions.default_groups_path = C.CString(clientOptions.DefaultGroupsPath)

	defaultRequestIdentifiers, defaultRequestIdentifiersLen := cEcsRequestIdentifiers(clientOptions.DefaultRequestIdentifiers)
	cClientOptions.default_request_identifiers = defaultRequestIdentifiers
	cClientOptions.default_request_identifiers_length = defaultRequestIdentifiersLen

	if len(clientOptions.X509Cert) > 0 {
		cCertBytes := C.CBytes(clientOptions.X509Cert)
		cClientOptions.x509_cert = (*C.uchar)(cCertBytes)
		cClientOptions.x509_cert_length = C.int(len(clientOptions.X509Cert))
	} else {
		cClientOptions.x509_cert = nil
		cClientOptions.x509_cert_length = C.int(0)
	}

	if clientOptions.EcsConfigurationEventCallbackFunc != nil {
		eventCallbackFunc = clientOptions.EcsConfigurationEventCallbackFunc
	}

	cClientOptions.event_callback = (*[0]byte)(C.eventCallback)

	if clientOptions.Logger != nil {
		logger = clientOptions.Logger
	}

	cClientOptions.log_callback = (*[0]byte)(C.logCallback)
	cClientOptions.log_level = C.ECS_LOG_LEVEL(clientOptions.LogLevel)

	cClientOptions.tenant_id = C.CString(clientOptions.TenantId)

	cClientOptions.client_id = C.CString(clientOptions.ClientId)

	if clientOptions.AuthenticationEnvironment != nil {
		auth_env := (*C.ECS_ENVIRONMENT_TYPE)(C.malloc(C.size_t(unsafe.Sizeof(uintptr(0)))))
		*auth_env = (C.ECS_ENVIRONMENT_TYPE)(C.int(*clientOptions.AuthenticationEnvironment))
		cClientOptions.auth_env = auth_env
	}

	cClientOptions.authentication_method = C.ECS_AUTHENTICATION_METHOD(clientOptions.AuthenticationMethod)

	cClientOptions.enable_exp = C.int(clientOptions.EnableExp)

	return &cClientOptions
}

// freeCEcsClientOptions frees the C.EcsClientOptions in the C heap.
func freeCEcsClientOptions(cEcsClientOptions *C.EcsClientOptions) {
	C.free(unsafe.Pointer(cEcsClientOptions.default_config_path))
	C.free(unsafe.Pointer(cEcsClientOptions.default_groups_path))

	freeCEcsRequestIdentifiers(cEcsClientOptions.default_request_identifiers, cEcsClientOptions.default_request_identifiers_length)

	C.free(unsafe.Pointer(cEcsClientOptions.x509_cert))
	C.free(unsafe.Pointer(cEcsClientOptions.tenant_id))
	C.free(unsafe.Pointer(cEcsClientOptions.client_id))

	if cEcsClientOptions.auth_env != nil {
		C.free(unsafe.Pointer(cEcsClientOptions.auth_env))
	}
}

// cEcsRequestIdentifiers converts the Go EcsRequestIdentifiers to a C.EcsRequestIdentifier array.
//
// Note that the C.EcsRequestIdentifier array is allocated in the C heap and therefore
// must be freed by calling freeCEcsRequestIdentifiers().
func cEcsRequestIdentifiers(ecsRequestIdentifiers EcsRequestIdentifiers) (*C.EcsRequestIdentifier, C.int) {
	requestIdentifiers := C.malloc(C.size_t(len(ecsRequestIdentifiers)) * C.sizeof_EcsRequestIdentifier)

	// we pick arbitrary size [1<<30 - 1] here just to make sure that our array is big enough
	pRequestIdentifiers := (*[1<<30 - 1]C.EcsRequestIdentifier)(requestIdentifiers)

	for i, defaultRequestIdentifier := range ecsRequestIdentifiers {
		requestIdentifier := (*C.EcsRequestIdentifier)(C.malloc(C.sizeof_EcsRequestIdentifier))
		requestIdentifier.name = C.CString(defaultRequestIdentifier.Name)

		values, valuesLen := cStringArray(defaultRequestIdentifier.Values...)
		requestIdentifier.values = values
		requestIdentifier.values_length = valuesLen
		pRequestIdentifiers[i] = *requestIdentifier
	}

	return (*C.EcsRequestIdentifier)(requestIdentifiers), C.int(len(ecsRequestIdentifiers))
}

// freeCEcsClientOptions frees the C.EcsRequestIdentifier array in the C heap.
func freeCEcsRequestIdentifiers(cEcsRequestIdentifiers *C.EcsRequestIdentifier, cEcsRequestIdentifiersLen C.int) {
	defer C.free(unsafe.Pointer(cEcsRequestIdentifiers))
	requestIdentifiers := unsafe.Slice(cEcsRequestIdentifiers, int(cEcsRequestIdentifiersLen))
	for _, requestIdentifier := range requestIdentifiers {
		C.free(unsafe.Pointer(requestIdentifier.name))
		freeCStringArray(requestIdentifier.values, requestIdentifier.values_length)
	}
}

// cStringArray converts the Go string slice to a *C.char array.
//
// Note that the *C.char array is allocated in the C heap and therefore
// must be freed by calling freeCStringArray().
func cStringArray(params ...string) (**C.char, C.int) {
	ret := C.malloc(C.size_t(len(params)) * C.size_t(unsafe.Sizeof(uintptr(0))))

	// we pick arbitrary size [1<<30 - 1] here just to make sure that our array is big enough
	pRet := (*[1<<30 - 1]*C.char)(ret)

	for i, item := range params {
		pRet[i] = C.CString(item)
	}

	return (**C.char)(ret), C.int(len(params))
}

// freeCStringArray frees the *C.char array in the C heap.
func freeCStringArray(cArray **C.char, cArrayLen C.int) {
	defer C.free(unsafe.Pointer(cArray))
	pRet := unsafe.Slice(cArray, int(cArrayLen))
	for _, item := range pRet {
		C.free(unsafe.Pointer(item))
	}
}

// statusCodeToError maps the C.ECS_STATUS_CODE to a Go error or nil.
func statusCodeToError(statusCode C.ECS_STATUS_CODE) error {
	if statusCode == C.ECS_STATUS_SUCCESS {
		return nil
	}

	return fmt.Errorf("ECS operation failed with status %v", statusCode)
}
