# zap-ecs

This package provides a zap logging layer that groups and serializes ECS compliant field keys into its appropriate JSON representations upon encoding. This is done via field aggregation and proxying on encoding time. Also, it providers helpers that create zap.Field instances of supported ECS fields

## ECS Compatibility status

Please note that this currently only encompasses a small subset of the complete ECS standard (https://www.elastic.co/guide/en/ecs/current), so you may consider it a WIP. You can check the currently supported keys/fields at ecs/ecs_keys.go

## Usage

### Creating a zap.Logger instance

The logger implements and exposes the following interface, which can be composed for decoration if needed by the consumer:

```go
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
```

The ecs logger requires the following parameter struct:

```go
type Options struct {
    BaseLoggerField zap.Field
    BaseTags        []string
    BaseLabels      []zap.Field
    Logger          *zap.Logger
}
```

In this context, BaseLoggerField is the zap.Field instance that identifies the logger and uses the "log.logger" field key, BaseTags is the zap.Field instance that identifies the log entry tags ("tags" field) and BaseLabels is an arbitrary set of fields that belong to the root level of any log entry and will be autoinjected on all messages

An example of a general purpose code based on environment variables could be the following

```go
// Definition of base tags and fields to be inyected by default on log entries
var (
    baseLoggerField = zap.String(zapEcsKeys.FieldLogger, "my-app_(uber-go/zap)")
        baseTags        = func() []string { return []string{os.Getenv("ENVIRONMENT")} }
            baseFields      = func() []zap.Field {
                return []zap.Field{
                    zap.String(ecs.FieldLabelApplication, os.Getenv("APPLICATION_NAME")),
                    zap.String(ecs.FieldLabelService, os.Getenv("LOGGING_SERVICE_NAME")),
                    zap.String(ecs.FieldLabelEnvironment, os.Getenv("ENVIRONMENT")),
                    zap.String(ecs.FieldLabelLibVersion, "v0.0.1"),
                    zap.String(ecs.FieldLabelLibLanguage, os.Getenv("GO_VERSION")),
                    zap.String(ecs.FieldLabelPodName, os.Getenv("MY_POD_NAME")),
                    zap.String(ecs.FieldLabelNodeName, os.Getenv("MY_NODE_NAME")),
        }
    }
)
```

Then, the logger instance can be built from the application:

```go
	cfg := zap.NewProductionConfig()

    // Adapt field names to ECS base:
    // https://www.elastic.co/guide/en/ecs/current/ecs-base.html
    cfg.EncoderConfig.MessageKey = zapEcsKeys.FieldMessage
    cfg.EncoderConfig.TimeKey = zapEcsKeys.FieldTimestamp
    cfg.EncoderConfig.LevelKey = "" // Omit it, we will generate it on our own (it conflicts with the ECS ObjectEncoder)
    cfg.DisableStacktrace = true    // Omit automatic stacktraces, these should be emitted by the recovery middleware


	l, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	ecsLogger := zapEcs.NewECSLogger(zapEcs.Options{
		BaseLoggerField: baseLoggerField,
		BaseTags:        baseTags,
		BaseLabels:      baseFields,
		Logger:          l,
		Level:           cfg.Level.Level(),
	})

	if ecsLogger == nil {
	    return nil, errors.New("log: failed to create ECS log")
	}
	
	// use ecsLogger as needed
```

### Helpers

For convenience, the encapsulated logger exposes the following methods from the native zap instance (use only if needed):

```go
AsCoreLogger() zapcore.Core

AsSugaredLogger() *zap.SugaredLogger
```

