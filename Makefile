TEST_PACKAGE := ./...

filter :=

ifneq "$(TEST_FOCUS)" ""
	filter := $(filter) -goblin.run='$(TEST_FOCUS)'
endif

build:
	go build -o chipi-gen ./chipi-gen/gen.go

# test-tools are binaries required to execute the tests
# ex:
#   go install github.com/gogo/protobuf/protoc-min-version
test-tools:

test: test-tools
	go test --tags=test $(TEST_PACKAGE) $(filter)


COVERAGE_OUT:=/tmp/cover
COVERAGE_RESULT:=/tmp/cover.html
coverage:
	go test -coverprofile $(COVERAGE_OUT) ./...
	go tool cover -html=$(COVERAGE_OUT) -o $(COVERAGE_RESULT)


# run the tests and run them again when a source file is changed
watch:
	find . -name "*.go" | entr -c make test

.PHONY: example
example: build
	cd example && make run

validate-example:
	npx @redocly/openapi-cli lint http://127.0.0.1:2121/doc.json

lint:
	golangci-lint run
