package objects

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Object is a field wrapper which encodes them into a plain, key/value struct upon marshal
type Object struct {
	fields []zap.Field
}

func AsObject(fields ...zap.Field) *Object {
	return &Object{fields}
}

// MarshalLogObject marshals the object as required by the zap serializer
func (f *Object) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	for _, field := range f.fields {
		// Encode either field's interface or string, in that order of availability
		if IsNilValue(field.Interface) {
			enc.AddString(field.Key, field.String)
		} else if err := EncodeFieldInterface(enc, field); err != nil {
			return err
		}
	}
	return nil
}
