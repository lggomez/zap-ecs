echo "** running formatters"
goimports -e -w $(find . -type f -name '*.go' -not -path "./vendor/*")

# use gofumpt if available, see https://github.com/mvdan/gofumpt
if ! command -v gofumpt -version &> /dev/null
then
    echo "gofumpt could not be found, defaulting to gofmt"
    gofmt -s -w $(find . -type f -name '*.go' -not -path "./vendor/*")
else
    echo "gofumpt found, executing as formatter"
    gofumpt -l -w .
fi

echo "** running golangci-lint"
golangci-lint run ./...

echo "** running tests"
go test $(go list ./...) -cover