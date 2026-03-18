.DEFAULT_GOAL := help
.PHONY: help generate-mock lint test test-cover test-race ci

help:
	@echo "vega-gateway-pro - 可用命令:"
	@echo "  make test        - 运行单元测试"
	@echo "  make test-cover  - 运行单元测试 + 覆盖率"
	@echo "  make lint        - 静态检查"
	@echo "  make ci          - CI 入口 (lint + test-cover)"

generate-mock:
	cd server && go generate ./...

lint:
	cd server && golangci-lint run ./... --exclude-dirs=interfaces/mock

UT_PACKAGES = $(shell cd server && go list ./... | grep -v /mock)

test:
	cd server && go test $(UT_PACKAGES) -gcflags=all=-l -v -count=1

test-cover:
	@mkdir -p test-result
	cd server && go test $(UT_PACKAGES) -gcflags=all=-l \
		-coverprofile=../test-result/coverage.out \
		-covermode=atomic
	@if [ -s test-result/coverage.out ]; then \
		go tool cover -func=test-result/coverage.out; \
		go tool cover -html=test-result/coverage.out -o test-result/coverage.html; \
		command -v gocover-cobertura >/dev/null 2>&1 && \
			gocover-cobertura < test-result/coverage.out > test-result/coverage.xml || \
			echo "WARN: gocover-cobertura not found, coverage.xml not generated"; \
	else \
		echo "WARN: no test coverage data (no test files?)"; \
	fi

test-race:
	cd server && go test $(UT_PACKAGES) -gcflags=all=-l -race -count=1

ci: lint test-cover
