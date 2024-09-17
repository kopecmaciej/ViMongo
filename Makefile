BUILD_DIR := .build
SVC_NAME := vi-mongo
INC_VERSION := $(shell git describe --tags --abbrev=0 | awk -F. '{OFS="."; $$NF+=1; print}')

.PHONY: build run

all: build run

build:
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(SVC_NAME) .

run:
	env $$(cat .env) $(BUILD_DIR)/$(SVC_NAME)

test:
	go test -v ./...

debug:
	if [ -f /proc/sys/kernel/yama/ptrace_scope ]; then \
		sudo sysctl kernel.yama.ptrace_scope=0; \
	fi
	go build -gcflags="all=-N -l" -o $(BUILD_DIR)/$(SVC_NAME) .
	$(BUILD_DIR)/$(SVC_NAME)

lint:
	golangci-lint run

release: check-version
	git tag -a $(INC_VERSION) -m "Release $(INC_VERSION)"
	git push origin $(INC_VERSION)

check-version:
	@if [ -z "$(INC_VERSION)" ]; then \
		echo "Error: INC_VERSION is not set"; \
		exit 1; \
	fi

bump-version:
	@git describe --tags --abbrev=0 | awk -F. '{OFS="."; $NF+=1; print $0}'