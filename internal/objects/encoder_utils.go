package objects

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// EncodeFieldInterface is a proxy between the zap field and the encoder, determining the appropriate
// representation of each field value
func EncodeFieldInterface(enc zapcore.ObjectEncoder, field zap.Field) error {
	key := field.Key
	switch field.Type {
	// ArrayMarshalerType indicates that the field carries an ArrayMarshaler.
	case zapcore.ArrayMarshalerType:
		if err := enc.AddArray(key, field.Interface.(zapcore.ArrayMarshaler)); err != nil {
			return err
		}
		// ObjectMarshalerType indicates that the field carries an ObjectMarshaler.
	case zapcore.ObjectMarshalerType:
		if err := enc.AddObject(key, field.Interface.(zapcore.ObjectMarshaler)); err != nil {
			return err
		}
		// BinaryType indicates that the field carries an opaque binary blob.
	case zapcore.BinaryType:
		enc.AddBinary(key, field.Interface.([]byte))
		// BoolType indicates that the field carries a bool.
	case zapcore.BoolType:
		enc.AddBool(key, field.Interface.(bool))
		// ByteStringType indicates that the field carries UTF-8 encoded bytes.
	case zapcore.ByteStringType:
		enc.AddByteString(key, field.Interface.([]byte))
		// DurationType indicates that the field carries a time.Duration.
	case zapcore.DurationType:
		enc.AddDuration(key, field.Interface.(time.Duration))
		// Float64Type indicates that the field carries a float64.
	case zapcore.Float64Type:
		enc.AddFloat64(key, field.Interface.(float64))
		// Float32Type indicates that the field carries a float32.
	case zapcore.Float32Type:
		enc.AddFloat32(key, field.Interface.(float32))
		// Int64Type indicates that the field carries an int64.
	case zapcore.Int64Type:
		enc.AddInt64(key, field.Interface.(int64))
		// Int32Type indicates that the field carries an int32.
	case zapcore.Int32Type:
		enc.AddInt32(key, field.Interface.(int32))
		// Int16Type indicates that the field carries an int16.
	case zapcore.Int16Type:
		enc.AddInt16(key, field.Interface.(int16))
		// Int8Type indicates that the field carries an int8.
	case zapcore.Int8Type:
		enc.AddInt8(key, field.Interface.(int8))
		// StringType indicates that the field carries a string.
	case zapcore.StringType:
		enc.AddString(key, field.Interface.(string))
		// TimeType indicates that the field carries a time.Time that is
		// representable by a UnixNano() stored as an int64.
	case zapcore.TimeType:
		enc.AddInt64(key, field.Integer)
		// TimeFullType indicates that the field carries a time.Time stored as-is.
	case zapcore.TimeFullType:
		enc.AddTime(key, field.Interface.(time.Time))
		// Uint64Type indicates that the field carries a uint64.
	case zapcore.Uint64Type:
		enc.AddUint64(key, field.Interface.(uint64))
		// Uint32Type indicates that the field carries a uint32.
	case zapcore.Uint32Type:
		enc.AddUint32(key, field.Interface.(uint32))
		// Uint16Type indicates that the field carries a uint16.
	case zapcore.Uint16Type:
		enc.AddUint16(key, field.Interface.(uint16))
		// Uint8Type indicates that the field carries a uint8.
	case zapcore.Uint8Type:
		enc.AddUint8(key, field.Interface.(uint8))
		// UintptrType indicates that the field carries a uintptr.
	case zapcore.UintptrType:
		enc.AddUintptr(key, field.Interface.(uintptr))
		// ReflectType indicates that the field carries an interface{}, which should
		// be serialized using reflection.
	case zapcore.ReflectType:
		if err := enc.AddReflected(key, field.Interface); err != nil {
			return err
		}
		// StringerType indicates that the field carries a fmt.Stringer.
	case zapcore.StringerType:
		enc.AddString(key, field.Interface.(fmt.Stringer).String())
		// ErrorType indicates that the field carries an error.
	case zapcore.ErrorType:
		enc.AddString(key, field.Interface.(error).Error())
		// SkipType indicates that the field is a no-op.
	case zapcore.SkipType:
		fallthrough //nolint:gocritic // we want to list this case
	default:
		break
	}
	return nil
}
