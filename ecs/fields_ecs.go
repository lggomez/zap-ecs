package ecs

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Internal lookup map for ECS key headers to be obfuscated/truncated
var secretEcsKeysMap = map[string]struct{}{
	"x-authorization":      {},
	"authorization":        {},
	"cookie":               {},
	"x-san-iatx-user-pass": {},
}

const (
	secretPlaceholderValue = "SECRET"
)

func SanitizeHeaders(val []http.Header) []string {
	plainHeaders := make([]string, 0, len(val))
	for _, header := range val {
		for key := range header {
			if _, found := secretEcsKeysMap[strings.ToLower(key)]; found {
				header[key] = []string{secretPlaceholderValue}
			}
			plainHeaders = append(plainHeaders, fmt.Sprintf("%s=%s", key, strings.Join(header[key], ",")))
		}
	}
	return plainHeaders
}

/*
	BASE FIELDS
*/

// Tags constructs a field that carries a string list, with no marshaling as it will be
// handled later on by the logger
func Tags(val []string) zap.Field {
	return zap.Field{Key: FieldTags, Type: zapcore.SkipType, Interface: val}
}

// Duration constructs a field with the given key and value.
func Duration(key string, val time.Duration) zap.Field {
	// Don't use the duration field as its encoder translates to seconds only: https://github.com/uber-go/zap/issues/649
	return zap.String(key, val.String())
}

// ServiceName constructs a String field with the FieldServiceName ECS standard key
func ServiceName(val string) zap.Field {
	return zap.String(FieldServiceName, val)
}

// EventAction constructs a String field with the FieldEventAction ECS standard key
func EventAction(val string) zap.Field {
	return zap.String(FieldEventAction, val)
}

// EventKind constructs a String field with the FieldEventKind ECS standard key
func EventKind(val string) zap.Field {
	return zap.String(FieldEventKind, val)
}

// EventCategory constructs a String field with the FieldEventCategory ECS standard key
func EventCategory(val string) zap.Field {
	return zap.String(FieldEventCategory, val)
}

// EventModule constructs a String field with the FieldEventModule ECS standard key
func EventModule(val string) zap.Field {
	return zap.String(FieldEventModule, val)
}

// EventType constructs a String field with the FieldEventType ECS standard key
func EventType(val string) zap.Field {
	return zap.String(FieldEventType, val)
}

// EventOriginal constructs a String field with the FieldEventOriginal ECS standard key
func EventOriginal(val string) zap.Field {
	return zap.String(FieldEventOriginal, val)
}

// EventOutcome constructs a String field with the FieldEventOutcome ECS standard key
func EventOutcome(val string) zap.Field {
	return zap.String(FieldEventOutcome, val)
}

// TraceID constructs a String field with the FieldTraceID ECS standard key
func TraceID(val string) zap.Field {
	return zap.String(FieldTraceID, val)
}

/*
	ERROR FIELDS
*/

// Err constructs an "error" field that carries an error. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Err(err error) zap.Field {
	if err != nil {
		return zap.String(FieldErrorMessage, err.Error())
	}

	return zap.Stringp(FieldErrorMessage, nil)
}

/*
	HTTP FIELDS
*/

// HTTPRequestBodyContent constructs a String field with the FieldHTTPRequestBodyContent ECS standard key
func HTTPRequestBodyContent(val string) zap.Field {
	return zap.String(FieldHTTPRequestBodyContent, val)
}

// HTTPRequestMethod constructs a String field with the FieldHTTPRequestMethod ECS standard key
func HTTPRequestMethod(val string) zap.Field {
	return zap.String(FieldHTTPRequestMethod, val)
}

// HTTPRequestBodyHeaders constructs a String field with the FieldHTTPRequestBodyHeaders ECS standard key
func HTTPRequestBodyHeaders(val []http.Header) zap.Field {
	plainHeaders := SanitizeHeaders(val)
	return zap.Strings(FieldHTTPRequestBodyHeaders, plainHeaders)
}

// HTTPRequestReferrer constructs a String field with the FieldHTTPRequestReferrer ECS standard key
func HTTPRequestReferrer(val string) zap.Field {
	return zap.String(FieldHTTPRequestReferrer, val)
}

// HTTPResponseBodyContent constructs a String field with the FieldHTTPResponseBodyContent ECS standard key
func HTTPResponseBodyContent(val string) zap.Field {
	return zap.String(FieldHTTPResponseBodyContent, val)
}

// HTTPResponseStatusCode constructs a String field with the FieldHTTPResponseStatusCode ECS standard key
func HTTPResponseStatusCode(val string) zap.Field {
	return zap.String(FieldHTTPResponseStatusCode, val)
}

// HTTPResponseBodyReferrer constructs a String field with the FieldHTTPResponseBodyReferrer ECS standard key
func HTTPResponseBodyReferrer(val string) zap.Field {
	return zap.String(FieldHTTPResponseBodyReferrer, val)
}
