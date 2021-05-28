package objects

import (
	stdJson "encoding/json"

	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var json = jsoniter.ConfigFastest

// HTTPObject is a specialized nested struct which transforms into a serialized field
type HTTPObject struct {
	fields  []zap.Field
	target  HTTPMarshalObject
	mapper  func(zap.Field, *HTTPMarshalObject)
	baseKey string
}

type HTTPMarshalObject struct {
	Request  *HTTPRequestMarshalObject  `json:"request,omitempty"`
	Response *HTTPResponseMarshalObject `json:"response,omitempty"`
}

type HTTPRequestMarshalObject struct {
	Body            *HTTPBodyMarshalObject `json:"body,omitempty"`
	RequestMethod   string                 `json:"method,omitempty"`
	RequestReferrer string                 `json:"referrer,omitempty"`
}

type HTTPResponseMarshalObject struct {
	Body             *HTTPBodyMarshalObject `json:"body,omitempty"`
	ResponseReferrer string                 `json:"referrer,omitempty"`
}

type HTTPBodyMarshalObject struct {
	BodyContent string `json:"content,omitempty"`
	Headers     string `json:"headers,omitempty"`
	StatusCode  string `json:"status_code,omitempty"`
}

func NestedObject(baseKey string, mapper func(zap.Field, *HTTPMarshalObject), fields ...zap.Field) *HTTPObject {
	return &HTTPObject{baseKey: baseKey, fields: fields, mapper: mapper}
}

func (f *HTTPObject) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	data, err := json.Marshal(&f.target)
	enc.AddByteString(f.baseKey, data)
	return err
}

func (f *HTTPObject) AsField() zap.Field {
	data, err := f.MarshalJSON()
	if err != nil {
		data = []byte(err.Error()) // TODO: should we propagate or log separately?
	}
	return zap.Any(f.baseKey, stdJson.RawMessage(data))
}

func (f *HTTPObject) MarshalJSON() ([]byte, error) {
	for _, field := range f.fields {
		f.mapper(field, &f.target)
	}
	return json.Marshal(&f.target)
}
