.PHONY: help
help:
	@echo "VEGA Manager - 可用命令:"
	@echo "  make test-at              - 运行所有AT测试"
	@echo "  make test-at-catalog      - 运行Catalog AT测试"
	@echo "  make test-at-setup        - 检查AT测试配置是否就绪"
	@echo ""
	@echo "运行AT测试前的准备工作:"
	@echo "  1. 启动VEGA Manager服务 (cd server && go run main.go)"
	@echo "  2. 准备测试目标MySQL"
	@echo "  3. 配置测试文件: server/tests/at/testdata/test-config.yaml"

.PHONY: test-at
test-at: test-at-setup
	@echo "运行AT测试..."
	@echo ""
	cd server && go test -v ./tests/at/... -timeout 5m

.PHONY: test-at-catalog
test-at-catalog: test-at-setup
	@echo "运行Catalog AT测试..."
	@echo ""
	cd server && go test -v ./tests/at/catalog/... -timeout 2m

.PHONY: test-at-setup
test-at-setup:
	@echo "检查AT测试配置..."
	@if [ ! -f server/tests/at/testdata/test-config.yaml ]; then \
		echo "❌ 错误: 测试配置文件不存在"; \
		echo ""; \
		echo "请执行以下步骤:"; \
		echo "  1. 复制配置模板: cp server/tests/at/testdata/test-config.yaml.example server/tests/at/testdata/test-config.yaml"; \
		echo "  2. 编辑 test-config.yaml，填入实际的连接信息"; \
		echo ""; \
		exit 1; \
	fi
	@echo "✓ 测试配置文件已存在"
	@echo ""
	@echo "请确保:"
	@echo "  1. VEGA Manager服务已启动"
	@echo "  2. 测试配置中的MySQL可访问"
	@echo ""

.PHONY: test-at-verbose
test-at-verbose: test-at-setup
	@echo "运行AT测试（详细输出）..."
	@echo ""
	cd server && go test -v -count=1 ./tests/at/... -timeout 5m

.PHONY: test-at-catalog-verbose
test-at-catalog-verbose: test-at-setup
	@echo "运行Catalog AT测试（详细输出）..."
	@echo ""
	cd server && go test -v -count=1 ./tests/at/catalog/... -timeout 2m

.PHONY: clean-test
clean-test:
	@echo "清理测试输出..."
	cd server && rm -f coverage.out

.DEFAULT_GOAL := help
