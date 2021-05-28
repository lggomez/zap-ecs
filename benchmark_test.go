// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

/* Modifications copyright (C) 2021 Luis Gabriel Gomez
-Adapted suite to author's zap core from logger implementation
-Extended test suite with more field test cases
*/

package zapecs

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"net/http"
	"runtime"
	"testing"
	"time"
	"unsafe"

	"github.com/lggomez/zap-ecs/ecs"

	pkgerrors "github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func BenchmarkCore(b *testing.B) {
	fields := []zapcore.Field{
		zap.String("string", "foo"),
		zap.Int64("int64", 1),
		zap.Int("int", 2),
		zap.Float64("float64", 1.0),
		zap.Bool("bool", true),
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
	}
	cores := map[string]func(ws zapcore.WriteSyncer) zapcore.Core{
		"zap": func(ws zapcore.WriteSyncer) zapcore.Core {
			enc := zapcore.NewJSONEncoder(zap.NewDevelopmentEncoderConfig())
			return zapcore.NewCore(enc, ws, zap.DebugLevel)
		},
		"ecs": func(ws zapcore.WriteSyncer) zapcore.Core {
			enc := zapcore.NewJSONEncoder(zap.NewDevelopmentEncoderConfig())
			core := zapcore.NewCore(enc, ws, zap.DebugLevel)
			l := &zapECSLogger{
				baseLoggerField: baseLoggerField,
				logger:          zap.New(core),
			}
			return l.AsLoggerCore()
		},
	}

	for name, coreBuilder := range cores {
		b.Run(name+"/fields", func(b *testing.B) {
			out := testWriteSyncer{}
			core := coreBuilder(&out)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				core.Write(zapcore.Entry{
					Message: "fake",
					Level:   zapcore.DebugLevel,
				}, fields)
				out.reset()
			}
		})

		b.Run(name+"/caller", func(b *testing.B) {
			caller := zapcore.NewEntryCaller(runtime.Caller(0))
			out := testWriteSyncer{}
			core := coreBuilder(&out)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				core.Write(zapcore.Entry{
					Message: "fake",
					Level:   zapcore.DebugLevel,
					Caller:  caller,
				}, fields)
				out.reset()
			}
		})

		b.Run(name+"/errors", func(b *testing.B) {
			err1 := errors.New("boom")
			err2 := pkgerrors.Wrap(err1, "crash")
			err3 := testErr{msg: "boom/crash", errors: []error{err1, err2}}
			fieldsWithErr := append(fields,
				zap.Error(err1),
				zap.Error(err2),
				zap.Error(err3),
			)
			out := testWriteSyncer{}
			core := coreBuilder(&out)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				core.Write(zapcore.Entry{
					Message: "fake",
					Level:   zapcore.DebugLevel,
				}, fieldsWithErr)
				out.reset()
			}
		})
	}
}

type testWriteSyncer struct {
	b bytes.Buffer
}

func (o *testWriteSyncer) Write(p []byte) (int, error) {
	return o.b.Write(p)
}

func (o *testWriteSyncer) Sync() error { return nil }

func (o *testWriteSyncer) reset() { o.b.Reset() }

type testErr struct {
	msg    string
	errors []error
}

func (e testErr) Error() string {
	return e.msg
}

func (e testErr) Errors() []error {
	return e.errors
}
