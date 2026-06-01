ROOT_PATH := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))
COVERAGE_PATH := $(ROOT_PATH).coverage/
POSTGRES_MODULE_PATH := $(ROOT_PATH)dberror/postgres

test:
	@rm -rf $(COVERAGE_PATH)
	@mkdir -p $(COVERAGE_PATH)
	@go test -v -coverpkg=./... ./... -coverprofile $(COVERAGE_PATH)coverage.txt
	@go tool cover -html=$(COVERAGE_PATH)coverage.txt -o $(COVERAGE_PATH)coverage.html
	@cd $(POSTGRES_MODULE_PATH) && go test -v ./...

.PHONY: test