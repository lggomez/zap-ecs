package ecs

// Encoding keys used for ECS compliance:
// https://www.elastic.co/guide/en/ecs/current/ecs-base.html
const (
	// Internal base fields to be used by the logger
	FieldTimestamp = "@timestamp"
	FieldMessage   = "message"
	FieldLabels    = "labels"
	FieldTags      = "tags"

	// Internal label fields to be used by the logger
	FieldLabelApplication = "application"
	FieldLabelService     = "service"
	FieldLabelEnvironment = "environment"
	FieldLabelLibVersion  = "lib_version"
	FieldLabelLibLanguage = "lib_language"
	FieldLabelPodName     = "pod_name"
	FieldLabelNodeName    = "node_name"

	FieldLogger   = "log.logger"
	FieldLogLevel = "log.level"

	// Public fields to be available for consumers
	FieldServiceName = "service.name"

	FieldErrorMessage = "error.message"
	FieldStackTrace   = "error.stack_trace"
	FieldErrorType    = "error.type"

	FieldEventAction   = "event.action"
	FieldEventKind     = "event.kind"
	FieldEventCategory = "event.category"
	FieldEventModule   = "event.module"
	FieldEventType     = "event.type"
	FieldEventOriginal = "event.original"
	FieldEventOutcome  = "event.outcome"

	FieldTraceID = "trace.id"

	FieldHTTPRequestBodyContent   = "http.request.body.content"
	FieldHTTPRequestMethod        = "http.request.method"
	FieldHTTPRequestBodyHeaders   = "http.request.body.headers"
	FieldHTTPRequestReferrer      = "http.request.referrer"
	FieldHTTPResponseBodyContent  = "http.response.body.content"
	FieldHTTPResponseStatusCode   = "http.response.status_code"
	FieldHTTPResponseBodyReferrer = "http.response.body.referrer"
)

func IsECSFieldName(fieldName string) bool {
	_, found := ecsKeysMap[fieldName]
	return found
}

// Internal lookup map for ECS keys on log fields
var ecsKeysMap = map[string]struct{}{
	FieldTags: {},

	FieldLogger:           {},
	FieldLogLevel:         {},
	FieldLabelApplication: {},
	FieldLabelService:     {},
	FieldLabelEnvironment: {},
	FieldLabelLibVersion:  {},
	FieldLabelLibLanguage: {},
	FieldLabelPodName:     {},
	FieldLabelNodeName:    {},

	FieldServiceName: {},

	FieldErrorMessage: {},
	FieldStackTrace:   {},
	FieldErrorType:    {},

	FieldEventAction:   {},
	FieldEventCategory: {},
	FieldEventModule:   {},
	FieldEventType:     {},
	FieldEventOriginal: {},
	FieldEventOutcome:  {},

	FieldTraceID: {},

	FieldHTTPRequestBodyContent:   {},
	FieldHTTPRequestMethod:        {},
	FieldHTTPRequestBodyHeaders:   {},
	FieldHTTPRequestReferrer:      {},
	FieldHTTPResponseBodyContent:  {},
	FieldHTTPResponseStatusCode:   {},
	FieldHTTPResponseBodyReferrer: {},
}

const (
	LogPrefix       = "log."
	LogBaseLevelKey = "log"

	HTTPPrefix       = "http."
	HTTPBaseLevelKey = "http"

	EventPrefix       = "event."
	EventBaseLevelKey = "event"

	ErrorPrefix       = "error."
	ErrorBaseLevelKey = "error"

	TracePrefix       = "trace."
	TraceBaseLevelKey = "trace"
)
