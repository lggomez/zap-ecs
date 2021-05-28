package zapecs

import (
	"strings"

	"github.com/lggomez/zap-ecs/ecs"
	"github.com/lggomez/zap-ecs/internal/objects"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Derived interface from zap.Logger
type Logger interface {
	SetLevel(level Level)

	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Panic(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)

	Flush() error
}

type zapECSLogger struct {
	baseLoggerField zap.Field
	baseTags        []string
	baseLabels      []zap.Field
	logger          *zap.Logger
}

type Options struct {
	BaseLoggerField zap.Field
	BaseTags        []string
	BaseLabels      []zap.Field
	Logger          *zap.Logger
}

func NewECSLogger(o Options) Logger {
	return &zapECSLogger{
		baseLoggerField: o.BaseLoggerField,
		baseTags:        o.BaseTags,
		baseLabels:      o.BaseLabels,
		logger:          o.Logger,
	}
}

type fieldAccumulators struct {
	l Level

	labelsFieldsAccum []zap.Field
	logFieldsAccum    []zap.Field
	httpFieldsAccum   []zap.Field
	eventFieldsAccum  []zap.Field
	errorFieldsAccum  []zap.Field
	traceFieldsAccum  []zap.Field
}

func newFieldAccumulators(labelsSize int, l Level) *fieldAccumulators {
	a := &fieldAccumulators{l: l}
	// Labels final size is non-deterministic, so we allocate it with an extra threshold
	a.labelsFieldsAccum = make([]zap.Field, 0, labelsSize+labelsSize/2)
	a.logFieldsAccum = make([]zap.Field, 0, 2)
	a.httpFieldsAccum = make([]zap.Field, 0, 7)
	a.eventFieldsAccum = make([]zap.Field, 0, 7)
	a.errorFieldsAccum = make([]zap.Field, 0, 7)
	a.traceFieldsAccum = make([]zap.Field, 0, 1)
	return a
}

// reduceKey gets the last element from a dotted path object key
func reduceKey(field zap.Field) zap.Field {
	keyTokens := strings.Split(field.Key, ".")
	field.Key = keyTokens[len(keyTokens)-1]
	return field
}

func (a *fieldAccumulators) appendField(f zap.Field) {
	if strings.HasPrefix(f.Key, ecs.LogPrefix) {
		// Field is part of the log object
		a.logFieldsAccum = append(a.logFieldsAccum, reduceKey(f))
	} else if strings.HasPrefix(f.Key, ecs.HTTPPrefix) {
		// Field is part of the http object
		// Don't sanitize key
		a.httpFieldsAccum = append(a.httpFieldsAccum, f)
	} else if strings.HasPrefix(f.Key, ecs.EventPrefix) {
		// Field is part of the event object
		a.eventFieldsAccum = append(a.eventFieldsAccum, reduceKey(f))
	} else if strings.HasPrefix(f.Key, ecs.ErrorPrefix) {
		// Field is part of the error object
		a.errorFieldsAccum = append(a.errorFieldsAccum, reduceKey(f))
	} else if strings.HasPrefix(f.Key, ecs.TracePrefix) {
		// Field is part of the error object
		a.traceFieldsAccum = append(a.traceFieldsAccum, reduceKey(f))
	} else {
		// No match: field is part of the labels object
		a.labelsFieldsAccum = append(a.labelsFieldsAccum, reduceKey(f))
	}
}

func (a *fieldAccumulators) emitLogFields(baseLoggerField zap.Field) []zap.Field {
	ret := make([]zap.Field, 0, 8)

	// Add logger fields
	a.logFieldsAccum = append(a.logFieldsAccum,
		reduceKey(baseLoggerField),
		reduceKey(zap.String(ecs.FieldLogLevel, a.l.String())))

	// Encode labels log object and add field
	ret = append(ret,
		zap.Object(ecs.LogBaseLevelKey, objects.AsObject(a.logFieldsAccum...)),
		objects.NestedObject(ecs.HTTPBaseLevelKey, objects.HTTPECSMapper, a.httpFieldsAccum...).AsField(),
		zap.Object(ecs.EventBaseLevelKey, objects.AsObject(a.eventFieldsAccum...)),
		zap.Object(ecs.ErrorBaseLevelKey, objects.AsObject(a.errorFieldsAccum...)),
		zap.Object(ecs.TraceBaseLevelKey, objects.AsObject(a.traceFieldsAccum...)),
		zap.Object(ecs.FieldLabels, objects.AsObject(a.labelsFieldsAccum...)))

	return ret
}

func (l zapECSLogger) encodeFields(fields []zap.Field, lvl Level) []zap.Field {
	// Do no process entries more than once per key (ignore duplicates)
	processedFields := make(map[string]struct{}, len(fields))

	// Prepare field slices
	logFields := make([]zap.Field, 0, len(fields)+len(l.baseLabels))
	accums := newFieldAccumulators(len(l.baseLabels), lvl)
	entryTags := make([]string, 0, len(l.baseTags))
	entryTags = append(entryTags, l.baseTags...)

	// Filter fields into ECS and label fields, merging tags in the process
	for _, field := range fields {
		if _, found := processedFields[field.Key]; found {
			// discard duplicate fields by key
			continue
		}

		processedFields[field.Key] = struct{}{}
		// Check if the field is an ECS standard field or not, and process accordingly
		if ecs.IsECSFieldName(field.Key) {
			switch field.Key {
			case ecs.FieldTags:
				// In the case of tags, we'll be merging the field tags add create the field later
				if field.Interface != nil {
					tags := field.Interface.([]string)
					entryTags = append(entryTags, tags...)
				}
			default:
				accums.appendField(field)
			}
		} else {
			accums.appendField(field)
		}
	}

	// Add tags field
	if len(entryTags) > 0 {
		logFields = append(logFields, zap.Strings(ecs.FieldTags, entryTags))
	}

	// Add the rest of the fields
	logFields = append(logFields, l.baseLabels...)
	logFields = append(logFields, accums.emitLogFields(l.baseLoggerField)...)

	return logFields
}

func (l zapECSLogger) Debug(msg string, fields ...zap.Field) {
	filteredFields := l.encodeFields(fields, DebugLevel)
	l.logger.Debug(msg, filteredFields...)
}

func (l zapECSLogger) Info(msg string, fields ...zap.Field) {
	filteredFields := l.encodeFields(fields, InfoLevel)
	l.logger.Info(msg, filteredFields...)
}

func (l zapECSLogger) Warn(msg string, fields ...zap.Field) {
	filteredFields := l.encodeFields(fields, WarnLevel)
	l.logger.Warn(msg, filteredFields...)
}

func (l zapECSLogger) Error(msg string, fields ...zap.Field) {
	filteredFields := l.encodeFields(fields, ErrorLevel)
	l.logger.Error(msg, filteredFields...)
}

func (l zapECSLogger) Panic(msg string, fields ...zap.Field) {
	filteredFields := l.encodeFields(fields, PanicLevel)
	l.logger.Panic(msg, filteredFields...)
}

func (l zapECSLogger) Fatal(msg string, fields ...zap.Field) {
	filteredFields := l.encodeFields(fields, FatalLevel)
	l.logger.Fatal(msg, filteredFields...)
}

func (l zapECSLogger) Flush() error {
	return l.logger.Sync()
}

func (l zapECSLogger) SetLevel(level Level) {
	l.logger.Core().Enabled(level)
}

func (l zapECSLogger) AsLoggerCore() zapcore.Core {
	return l.logger.Core()
}

func (l zapECSLogger) AsSugaredLogger() *zap.SugaredLogger {
	return l.logger.Sugar()
}
