package objects

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/lggomez/zap-ecs/ecs"
)

func HTTPECSMapper(field zap.Field, object *HTTPMarshalObject) {
	switch field.Key {
	case ecs.FieldHTTPRequestBodyContent:
		requestNilGuard(object)
		object.Request.Body.BodyContent = getValueAsString(field)
	case ecs.FieldHTTPRequestBodyHeaders:
		requestNilGuard(object)
		object.Request.Body.Headers = getValueAsString(field)
	case ecs.FieldHTTPRequestMethod:
		requestNilGuard(object)
		object.Request.RequestMethod = getValueAsString(field)
	case ecs.FieldHTTPRequestReferrer:
		requestNilGuard(object)
		object.Request.RequestReferrer = getValueAsString(field)
	case ecs.FieldHTTPResponseBodyContent:
		responseNilGuard(object)
		object.Response.Body.BodyContent = getValueAsString(field)
	case ecs.FieldHTTPResponseStatusCode:
		responseNilGuard(object)
		object.Response.Body.StatusCode = getValueAsString(field)
	case ecs.FieldHTTPResponseBodyReferrer:
		responseNilGuard(object)
		object.Response.ResponseReferrer = getValueAsString(field)
	}
}

func requestNilGuard(object *HTTPMarshalObject) {
	if object.Request == nil {
		object.Request = &HTTPRequestMarshalObject{}
	}
	if object.Request.Body == nil {
		object.Request.Body = &HTTPBodyMarshalObject{}
	}
}

func responseNilGuard(object *HTTPMarshalObject) {
	if object.Response == nil {
		object.Response = &HTTPResponseMarshalObject{}
	}
	if object.Response.Body == nil {
		object.Response.Body = &HTTPBodyMarshalObject{}
	}
}

func getValueAsString(field zap.Field) string {
	if field.String != "" {
		return field.String
	}
	if !IsNilValue(field.Interface) {
		if headers, ok := field.Interface.([]http.Header); ok {
			return fmt.Sprintf("%v", ecs.SanitizeHeaders(headers))
		}
	}
	return fmt.Sprintf("%v", field.Interface)
}
