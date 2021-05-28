package zapecs

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/lggomez/zap-ecs/ecs"
	"github.com/lggomez/zap-ecs/internal/test"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// applicationVersionGetter returns the toolkit version constant from the manifest file
var applicationVersionGetter = func() string {
	return "v0.0.1-test"
}

var (
	baseLoggerField = zap.String(ecs.FieldLogger, "ecs_(uber-go/zap)")
	baseTags        = func() []string { return []string{os.Getenv("ENVIRONMENT")} }
	baseFields      = func() []zap.Field {
		return []zap.Field{
			zap.String(ecs.FieldLabelApplication, os.Getenv("APPLICATION_NAME")),
			zap.String(ecs.FieldLabelService, os.Getenv("LOGGING_SERVICE_NAME")),
			zap.String(ecs.FieldLabelEnvironment, os.Getenv("ENVIRONMENT")),
			zap.String(ecs.FieldLabelLibVersion, applicationVersionGetter()),
			zap.String(ecs.FieldLabelLibLanguage, os.Getenv("GO_VERSION")),
			zap.String(ecs.FieldLabelPodName, os.Getenv("MY_POD_NAME")),
			zap.String(ecs.FieldLabelNodeName, os.Getenv("MY_NODE_NAME")),
		}
	}
)

func buildLoggerConfig() zap.Config {
	cfg := zap.NewProductionConfig()

	// Adapt field names to ECS base:
	// https://www.elastic.co/guide/en/ecs/current/ecs-base.html
	cfg.EncoderConfig.MessageKey = ecs.FieldMessage
	cfg.EncoderConfig.TimeKey = ecs.FieldTimestamp
	cfg.EncoderConfig.LevelKey = "" // Omit it, we will generate it on our own (it conflicts with the ECS ObjectEncoder)
	cfg.DisableStacktrace = true

	return cfg
}

//nolint:golint //we want to return the encapsulated type
func NewBufferedLogger(baseTags []string, baseFields []zap.Field) (*bytes.Buffer, *zapECSLogger) {
	buf := &bytes.Buffer{}
	ws := zapcore.AddSync(buf)

	jsonEncoder := zapcore.NewJSONEncoder(buildLoggerConfig().EncoderConfig)
	core := zapcore.NewCore(jsonEncoder, ws, zap.DebugLevel)

	return buf, &zapECSLogger{
		baseLoggerField: baseLoggerField,
		baseTags:        baseTags,
		baseLabels:      baseFields,
		logger:          zap.New(core),
	}
}

// SanitizeTestTimestamp replaces the current time.Now() generated timestamp
// with a fixed one to allow string assertions
func SanitizeTestTimestamp(data []byte) []byte {
	s := string(data)
	tsToken := fmt.Sprintf("\"%v\":", ecs.FieldTimestamp)
	start := strings.Index(s, tsToken)
	end := strings.Index(s, ",\"message\"")
	ret := []byte(s[0:start+len(tsToken)] + "1600000000.0000000" + s[end:])
	return ret
}

func Test_LoggerNoTags(t *testing.T) {
	// Set up environment
	prev := os.Getenv("ENVIRONMENT")
	os.Setenv("ENVIRONMENT", "test-environment")
	defer func() {
		os.Setenv("ENVIRONMENT", prev)
	}()

	// Set up log
	buf, l := NewBufferedLogger(nil, nil)

	testName := "TEST/no_tags_simple"
	t.Run(testName, func(t *testing.T) {
		buf.Truncate(0)
		l.Info("this is a test message", zap.String("foo", "a"), ecs.Duration("elapsed_example", 1532*time.Millisecond))
		test.AssertBytesAsJSON(t, testName, SanitizeTestTimestamp(buf.Bytes()))
	})
}

type stringerMock struct{}

func (s *stringerMock) String() string {
	return "foo"
}

var mockStackTrace = "goroutine 22 [running]:\\nruntime/debug.Stack(0x1483cf0, 0xd, 0xc0001587e8)\\n\\t/usr/local/go/src/runtime/debug/stack.go:24 +0x9d\\nmy-library/platform/log.Test_LoggerFormat.func5(0xc00024e5a0)\\n\\t/Users/home/my-library/log/logger_test.go:193 +0x260e\\ntesting.tRunner(0xc00024e5a0, 0xc000250360)\\n\\t/usr/local/go/src/testing/testing.go:1050 +0xdc\\ncreated by testing.(*T).Run\\n\\t/usr/local/go/src/testing/testing.go:1095 +0x28b\\n\": \"my-library/platform/log.Stack\\n\\t/Users/home/my-library/log/fields.go:237\\nmy-library/platform/log.Test_LoggerFormat.func5\\n\\t/Users/home/my-library/log/logger_test.go:193\\ntesting.tRunner\\n\\t/usr/local/go/src/testing/testing.go:1050"

func Test_LoggerFormat(t *testing.T) {
	// Set up environment
	prevGetter := applicationVersionGetter
	applicationVersionGetter = func() string {
		return "test-local-kit"
	}
	prev := os.Getenv("ENVIRONMENT")
	prev2 := os.Getenv("APPLICATION_NAME")
	prev3 := os.Getenv("LOGGING_SERVICE_NAME")
	prev4 := os.Getenv("MY_POD_NAME")
	prev5 := os.Getenv("MY_NODE_NAME")
	prev6 := os.Getenv("GO_VERSION")
	os.Setenv("ENVIRONMENT", "test-environment")
	os.Setenv("APPLICATION_NAME", "test-application")
	os.Setenv("LOGGING_SERVICE_NAME", "test-logging-service")
	os.Setenv("MY_POD_NAME", "local-pod")
	os.Setenv("MY_NODE_NAME", "local-node")
	os.Setenv("GO_VERSION", "go version go1.14.12 darwin/amd64")
	defer func() {
		applicationVersionGetter = prevGetter
		os.Setenv("ENVIRONMENT", prev)
		os.Setenv("APPLICATION_NAME", prev2)
		os.Setenv("LOGGING_SERVICE_NAME", prev3)
		os.Setenv("MY_POD_NAME", prev4)
		os.Setenv("MY_NODE_NAME", prev5)
		os.Setenv("GO_VERSION", prev6)
	}()

	// Set up log
	buf, l := NewBufferedLogger(baseTags(), baseFields())

	suts := map[string]func(msg string, fields ...zap.Field){
		"Info":  l.Info,
		"Warn":  l.Warn,
		"Error": l.Error,
		"Debug": l.Debug,
		//"Fatal": l.Fatal, fatal exits the application, so we cannot do assertions on this suite
		//"Panic": l.Panic, panic throws a panic, assert separately
		"2nd_pass_Info":  l.Info,
		"2nd_pass_Warn":  l.Warn,
		"2nd_pass_Error": l.Error,
		"2nd_pass_Debug": l.Debug,
	}

	for loggerName, sut := range suts {
		testName := fmt.Sprintf("%s_simple", loggerName)
		t.Run(testName, func(t *testing.T) {
			buf.Truncate(0)
			sut("this is a test message", zap.String("foo", "a"), ecs.Duration("elapsed_example", 1532*time.Millisecond))
			test.AssertBytesAsJSON(t, testName, SanitizeTestTimestamp(buf.Bytes()))
		})

		testName = fmt.Sprintf("%s_simple_extra_tags", loggerName)
		t.Run(testName, func(t *testing.T) {
			buf.Truncate(0)
			sut("this is a test message", zap.String("foo", "a"), ecs.Duration("elapsed_example", 1532*time.Millisecond), ecs.Tags([]string{"tag1", "tag2", "tag3"}))
			test.AssertBytesAsJSON(t, testName, SanitizeTestTimestamp(buf.Bytes()))
		})

		testName = fmt.Sprintf("%s_all_ecs_fields", loggerName)
		t.Run(testName, func(t *testing.T) {
			buf.Truncate(0)
			sut("this is a test message",
				zap.String("foo", "bar"),
				ecs.Duration("elapsed_example", 1532*time.Millisecond), ecs.Tags([]string{"tag1", "tag2", "tag3"}),
				zap.String(ecs.FieldServiceName, "logger.test"),

				zap.String(ecs.FieldErrorMessage, "fail"),
				zap.String(ecs.FieldStackTrace, mockStackTrace),
				zap.String(ecs.FieldErrorType, "panic"),

				zap.String(ecs.FieldEventAction, "test-started"),
				zap.String(ecs.FieldEventKind, "test"),
				zap.String(ecs.FieldEventCategory, "test_case"),
				zap.String(ecs.FieldEventModule, "log_test"),
				zap.String(ecs.FieldEventType, "test"),
				zap.String(ecs.FieldEventOriginal, "test"),
				zap.String(ecs.FieldEventOutcome, "test-outcome"),

				zap.String(ecs.FieldTraceID, " 1c6ee3fc-c19e-4a63-bcb4-dd1f1862bad0"),

				zap.String(ecs.FieldHTTPRequestBodyContent, "{\"foo\": 42}"),
				zap.String(ecs.FieldHTTPRequestMethod, "POST"),
				zap.Any(ecs.FieldHTTPRequestBodyHeaders, []http.Header{{"foo": []string{"bar"}}}),
				zap.String(ecs.FieldHTTPRequestReferrer, "https://www2.luisgg.com.ar/"),
				zap.String(ecs.FieldHTTPResponseBodyContent, "{\"result\": \"OK\"}"),
				zap.String(ecs.FieldHTTPResponseStatusCode, fmt.Sprintf("%v", http.StatusCreated)),
				zap.String(ecs.FieldHTTPResponseBodyReferrer, "https://www3.luisgg.com.ar/"),
			)
			test.AssertBytesAsJSON(t, testName, SanitizeTestTimestamp(buf.Bytes()))
		})

		testName = fmt.Sprintf("%s_ecs_field_idempotence", loggerName)
		t.Run(testName, func(t *testing.T) {
			buf.Truncate(0)
			sut("this is a test message",
				zap.String("foo", "bar"),
				ecs.Duration("elapsed_example", 1532*time.Millisecond), ecs.Tags([]string{"tag1", "tag2", "tag3"}),

				zap.String(ecs.FieldServiceName, "logger.test"),

				zap.String(ecs.FieldErrorMessage, "fail"),
				zap.String(ecs.FieldStackTrace, mockStackTrace),
				zap.String(ecs.FieldErrorType, "panic"),

				zap.String(ecs.FieldEventAction, "test-started"),
				zap.String(ecs.FieldEventKind, "test"),
				zap.String(ecs.FieldEventCategory, "test_case"),
				zap.String(ecs.FieldEventModule, "log_test"),
				zap.String(ecs.FieldEventType, "test"),
				zap.String(ecs.FieldEventOriginal, "test"),
				zap.String(ecs.FieldEventOutcome, "test-outcome"),

				zap.String(ecs.FieldTraceID, " 1c6ee3fc-c19e-4a63-bcb4-dd1f1862bad0"),

				zap.String(ecs.FieldHTTPRequestBodyContent, "{\"foo\": 42}"),
				zap.String(ecs.FieldHTTPRequestMethod, "POST"),
				zap.Any(ecs.FieldHTTPRequestBodyHeaders, []http.Header{{"foo": []string{"bar"}}}),
				zap.String(ecs.FieldHTTPRequestReferrer, "https://www2.luisgg.com.ar/"),
				zap.String(ecs.FieldHTTPResponseBodyContent, "{\"result\": \"OK\"}"),
				zap.String(ecs.FieldHTTPResponseStatusCode, fmt.Sprintf("%v", http.StatusCreated)),
				zap.String(ecs.FieldHTTPResponseBodyReferrer, "https://www3.luisgg.com.ar/"),

				ecs.ServiceName("logger.test"),

				ecs.EventAction("test-started"),
				ecs.EventKind("test"),
				ecs.EventCategory("test_case"),
				ecs.EventModule("log_test"),
				ecs.EventType("test"),
				ecs.EventOriginal("test"),
				ecs.EventOutcome("test-outcome"),

				ecs.TraceID(" 1c6ee3fc-c19e-4a63-bcb4-dd1f1862bad0"),

				ecs.HTTPRequestBodyContent("{\"foo\": 42}"),
				ecs.HTTPRequestMethod("POST"),
				ecs.HTTPRequestBodyHeaders([]http.Header{{"foo": []string{"bar"}}}),
				ecs.HTTPRequestReferrer("https://www2.luisgg.com.ar/"),
				ecs.HTTPResponseBodyContent("{\"result\": \"OK\"}"),
				ecs.HTTPResponseStatusCode(fmt.Sprintf("%v", http.StatusCreated)),
				ecs.HTTPResponseBodyReferrer("https://www3.luisgg.com.ar/"),
			)
			test.AssertBytesAsJSON(t, testName, SanitizeTestTimestamp(buf.Bytes()))
		})

		testName = fmt.Sprintf("%s_all_field_types", loggerName)
		t.Run(testName, func(t *testing.T) {
			buf.Truncate(0)
			sut("this is a test message",
				ecs.Err(errors.New("fail")),
				ecs.Tags([]string{"tag1", "tag2", "tag3"}),
				zap.Binary("binary_example", []byte{0, 1, 0, 1, 0, 1}),
				zap.Bool("bool_example", false),
				zap.Boolp("boolp_example", func(val bool) *bool { return &val }(false)), // return pointer via annonymous function
				zap.ByteString("bytestring_example", []byte("getIntPointer(val int) *int {")),
				zap.Float64("float64_example", math.Pow(math.Pi, 2)),
				zap.Float64p("float64p_example", func(val float64) *float64 { return &val }(math.Pow(math.Pi, 2))), // return pointer via annonymous function
				zap.Float32("float32_example", math.Pi),
				zap.Float32p("float32p_example", func(val float32) *float32 { return &val }(math.Pi)),
				zap.Int("int_example", 42),
				zap.Intp("intp_example", func(val int) *int { return &val }(42)),
				zap.Int8("int8_example", 42),
				zap.Int8p("int8p_example", func(val int8) *int8 { return &val }(42)),
				zap.Int16("int16_example", 42),
				zap.Int16p("int16p_example", func(val int16) *int16 { return &val }(42)),
				zap.Int32("int32_example", 42),
				zap.Int32p("int32p_example", func(val int32) *int32 { return &val }(42)),
				zap.Int64("int64_example", 42),
				zap.Int64p("int64p_example", func(val int64) *int64 { return &val }(42)),
				zap.String("string_example", "foo"),
				zap.Strings("strings_example", []string{"foo, bar, baz, qux"}),
				zap.Stringp("stringp_example", func(val string) *string { return &val }("foo")),
				zap.Uint("uint_example", 42),
				zap.Uintp("uintp_example", func(val uint) *uint { return &val }(42)),
				zap.Uint8("uint8_example", 42),
				zap.Uint8p("uint8p_example", func(val uint8) *uint8 { return &val }(42)),
				zap.Uint16("uint16_example", 42),
				zap.Uint16p("uint16p_example", func(val uint16) *uint16 { return &val }(42)),
				zap.Uint32("uint32_example", 42),
				zap.Uint32p("uint32p_example", func(val uint32) *uint32 { return &val }(42)),
				zap.Uint64("uint64_example", 42),
				zap.Uint64p("uint64p_example", func(val uint64) *uint64 { return &val }(42)),
				zap.Uintptr("uintptr_example", func(val int64) uintptr { return uintptr(unsafe.Pointer(&val)) }(42)),
				zap.Reflect("reflect_example", "foo"),
				zap.Stringer("stringer_example", &stringerMock{}),
				zap.Time("time_fullexample", time.Date(1990, time.November, 26, 17, 56, 11, 31, time.UTC).UTC()),
				zap.Timep("timep_fullexample", func(val time.Time) *time.Time { return &val }(time.Date(1990, time.November, 26, 17, 56, 11, 31, time.UTC).UTC())),
				zap.Time("time_example", time.Unix(0, math.MinInt64).Add(-1*time.Hour)),
				zap.Timep("timep_example", func(val time.Time) *time.Time { return &val }(time.Unix(0, math.MinInt64).Add(-1*time.Hour))),
				ecs.Duration("elapsed_example", 1532*time.Millisecond),
				zap.Any("any_example", map[string]interface{}{"foo": math.SqrtPhi}),
			)
			test.AssertBytesAsJSON(t, testName, SanitizeTestTimestamp(buf.Bytes()))
		})

		testName = fmt.Sprintf("%s_all_fields", loggerName)
		t.Run(testName, func(t *testing.T) {
			buf.Truncate(0)
			sut("this is a test message",
				zap.String(ecs.FieldServiceName, "logger.test"),

				zap.String(ecs.FieldErrorMessage, "fail"),
				zap.String(ecs.FieldStackTrace, mockStackTrace),
				zap.String(ecs.FieldErrorType, "panic"),

				zap.String(ecs.FieldEventAction, "test-started"),
				zap.String(ecs.FieldEventKind, "test"),
				zap.String(ecs.FieldEventCategory, "test_case"),
				zap.String(ecs.FieldEventModule, "log_test"),
				zap.String(ecs.FieldEventType, "test"),
				zap.String(ecs.FieldEventOriginal, "test"),
				ecs.EventOutcome("test-outcome"),

				zap.String(ecs.FieldTraceID, " 1c6ee3fc-c19e-4a63-bcb4-dd1f1862bad0"),

				zap.String(ecs.FieldHTTPRequestBodyContent, "{\"foo\": 42}"),
				zap.String(ecs.FieldHTTPRequestMethod, "POST"),
				zap.Any(ecs.FieldHTTPRequestBodyHeaders, []http.Header{{"header1": []string{"foo1"}}}),
				zap.String(ecs.FieldHTTPRequestReferrer, "https://www2.luisgg.com.ar/"),
				zap.String(ecs.FieldHTTPResponseBodyContent, "{\"result\": \"OK\"}"),
				zap.String(ecs.FieldHTTPResponseStatusCode, fmt.Sprintf("%v", http.StatusCreated)),
				zap.String(ecs.FieldHTTPResponseBodyReferrer, "https://www3.luisgg.com.ar/"),

				ecs.ServiceName("logger.test"),

				ecs.EventAction("test-started"),
				ecs.EventKind("test"),
				ecs.EventCategory("test_case"),
				ecs.EventModule("log_test"),
				ecs.EventType("test"),
				ecs.EventOriginal("test"),
				ecs.EventOutcome("test-outcome"),

				ecs.TraceID(" 1c6ee3fc-c19e-4a63-bcb4-dd1f1862bad0"),

				ecs.HTTPRequestBodyContent("{\"foo\": 42}"),
				ecs.HTTPRequestMethod("POST"),
				ecs.HTTPRequestBodyHeaders([]http.Header{{"foo": []string{"bar"}}}),
				ecs.HTTPRequestReferrer("https://www2.luisgg.com.ar/"),
				ecs.HTTPResponseBodyContent("{\"result\": \"OK\"}"),
				ecs.HTTPResponseStatusCode(fmt.Sprintf("%v", http.StatusCreated)),
				ecs.HTTPResponseBodyReferrer("https://www3.luisgg.com.ar/"),

				ecs.Err(errors.New("fail")),
				ecs.Tags([]string{"tag1", "tag2", "tag3"}),
				zap.Binary("binary_example", []byte{0, 1, 0, 1, 0, 1}),
				zap.Bool("bool_example", false),
				zap.Boolp("boolp_example", func(val bool) *bool { return &val }(false)), // return pointer via annonymous function
				zap.ByteString("bytestring_example", []byte("getIntPointer(val int) *int {")),
				zap.Float64("float64_example", math.Pow(math.Pi, 2)),
				zap.Float64p("float64p_example", func(val float64) *float64 { return &val }(math.Pow(math.Pi, 2))), // return pointer via annonymous function
				zap.Float32("float32_example", math.Pi),
				zap.Float32p("float32p_example", func(val float32) *float32 { return &val }(math.Pi)),
				zap.Int("int_example", 42),
				zap.Intp("intp_example", func(val int) *int { return &val }(42)),
				zap.Int8("int8_example", 42),
				zap.Int8p("int8p_example", func(val int8) *int8 { return &val }(42)),
				zap.Int16("int16_example", 42),
				zap.Int16p("int16p_example", func(val int16) *int16 { return &val }(42)),
				zap.Int32("int32_example", 42),
				zap.Int32p("int32p_example", func(val int32) *int32 { return &val }(42)),
				zap.Int64("int64_example", 42),
				zap.Int64p("int64p_example", func(val int64) *int64 { return &val }(42)),
				zap.String("string_example", "foo"),
				zap.Strings("strings_example", []string{"foo, bar, baz, qux"}),
				zap.Stringp("stringp_example", func(val string) *string { return &val }("foo")),
				zap.Uint("uint_example", 42),
				zap.Uintp("uintp_example", func(val uint) *uint { return &val }(42)),
				zap.Uint8("uint8_example", 42),
				zap.Uint8p("uint8p_example", func(val uint8) *uint8 { return &val }(42)),
				zap.Uint16("uint16_example", 42),
				zap.Uint16p("uint16p_example", func(val uint16) *uint16 { return &val }(42)),
				zap.Uint32("uint32_example", 42),
				zap.Uint32p("uint32p_example", func(val uint32) *uint32 { return &val }(42)),
				zap.Uint64("uint64_example", 42),
				zap.Uint64p("uint64p_example", func(val uint64) *uint64 { return &val }(42)),
				zap.Uintptr("uintptr_example", func(val int64) uintptr { return uintptr(unsafe.Pointer(&val)) }(42)),
				zap.Reflect("reflect_example", "foo"),
				zap.Stringer("stringer_example", &stringerMock{}),
				zap.Time("time_fullexample", time.Date(1990, time.November, 26, 17, 56, 11, 31, time.UTC).UTC()),
				zap.Timep("timep_fullexample", func(val time.Time) *time.Time { return &val }(time.Date(1990, time.November, 26, 17, 56, 11, 31, time.UTC).UTC())),
				zap.Time("time_example", time.Unix(0, math.MinInt64).Add(-1*time.Hour)),
				zap.Timep("timep_example", func(val time.Time) *time.Time { return &val }(time.Unix(0, math.MinInt64).Add(-1*time.Hour))),
				ecs.Duration("elapsed_example", 1532*time.Millisecond),
				zap.Any("any_example", map[string]interface{}{"foo": math.SqrtPhi}),
			)
			test.AssertBytesAsJSON(t, testName, SanitizeTestTimestamp(buf.Bytes()))
		})
	}
}

func Test_LoggerHeaders(t *testing.T) {
	// Set up environment
	prevGetter := applicationVersionGetter
	applicationVersionGetter = func() string {
		return "test-local-kit"
	}
	prev := os.Getenv("ENVIRONMENT")
	prev2 := os.Getenv("APPLICATION_NAME")
	prev3 := os.Getenv("LOGGING_SERVICE_NAME")
	prev4 := os.Getenv("MY_POD_NAME")
	prev5 := os.Getenv("MY_NODE_NAME")
	prev6 := os.Getenv("GO_VERSION")
	os.Setenv("ENVIRONMENT", "test-environment")
	os.Setenv("APPLICATION_NAME", "test-application")
	os.Setenv("LOGGING_SERVICE_NAME", "test-logging-service")
	os.Setenv("MY_POD_NAME", "local-pod")
	os.Setenv("MY_NODE_NAME", "local-node")
	os.Setenv("GO_VERSION", "go version go1.14.12 darwin/amd64")
	defer func() {
		applicationVersionGetter = prevGetter
		os.Setenv("ENVIRONMENT", prev)
		os.Setenv("APPLICATION_NAME", prev2)
		os.Setenv("LOGGING_SERVICE_NAME", prev3)
		os.Setenv("MY_POD_NAME", prev4)
		os.Setenv("MY_NODE_NAME", prev5)
		os.Setenv("GO_VERSION", prev6)
	}()

	// Set up log
	buf, l := NewBufferedLogger(baseTags(), baseFields())

	suts := map[string]func(msg string, fields ...zap.Field){
		"Info":  l.Info,
		"Warn":  l.Warn,
		"Error": l.Error,
		"Debug": l.Debug,
		//"Fatal": l.Fatal, fatal exits the application, so we cannot do assertions on this suite
		//"Panic": l.Panic, panic throws a panic, assert separately
		"2nd_pass_Info":  l.Info,
		"2nd_pass_Warn":  l.Warn,
		"2nd_pass_Error": l.Error,
		"2nd_pass_Debug": l.Debug,
	}

	for loggerName, sut := range suts {
		testName := fmt.Sprintf("%s_all_fields", loggerName)
		t.Run(testName, func(t *testing.T) {
			buf.Truncate(0)
			sut("this is a test message",
				ecs.HTTPRequestBodyHeaders([]http.Header{
					{"x-authorization": []string{"foo1"}},
					{"authorization": []string{"foo2"}},
					{"meh": []string{"not", "a", "secret"}},
					{"cookie": []string{"foo3"}},
					{"x-san-iatx-user-pass": []string{"foo4"}},
				}),
			)
			test.AssertBytesAsJSON(t, testName, SanitizeTestTimestamp(buf.Bytes()))
		})
	}
}
