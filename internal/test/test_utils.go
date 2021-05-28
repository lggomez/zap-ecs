package test

import (
	"encoding/json"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/sebdah/goldie/v2"
)

var (
	nameRegistry = map[string]struct{}{}
	mux          = &sync.Mutex{}
)

func AssertBytesAsJSON(t *testing.T, testCase string, bytes []byte) {
	t.Helper()

	caseName := createCanonicalTestCaseName(t, testCase)

	g := goldie.New(t)

	object := map[string]interface{}{}
	if err := json.Unmarshal(bytes, &object); err != nil {
		panic(err.Error())
	}
	g.AssertJson(t, caseName, object)
}

// createCanonicalTestCaseName Generate a canonical name for this test case
// and validates that it is not already in use
func createCanonicalTestCaseName(t *testing.T, testCase string) string {
	mux.Lock()
	defer mux.Unlock()

	caseName := getCanonicalName(testCase)
	if _, exists := nameRegistry[caseName]; exists {
		t.Errorf("golden: test case name %s already in use", caseName)
	}

	nameRegistry[caseName] = struct{}{}

	return caseName
}

// getCanonicalName Create a canonical test case name using the caller function name
func getCanonicalName(testCase string) string {
	return getCaller() + string(os.PathSeparator) + testCase
}

func getCaller() string {
	// Skip frames up to the caller test function
	// nolint:gomnd,dogsled // runtime api & see upper comment
	pc, _, _, _ := runtime.Caller(4)
	f := runtime.FuncForPC(pc)
	fullName := f.Name()

	tokens := strings.Split(fullName, "/")
	testFuncName := tokens[len(tokens)-1]

	// Remove anonymous function names from test function name
	tokens = strings.Split(testFuncName, ".")
	if strings.HasPrefix(tokens[len(tokens)-1], "func") {
		tokens = tokens[:len(tokens)-1]
	}

	return strings.TrimSpace(strings.Join(tokens, "."))
}
