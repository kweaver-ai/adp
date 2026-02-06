---
name: go-pre-commit-check
description: 在 git commit 前对 Go 项目执行 lint 与单元测试预检查，未通过时尝试自动修复。在用户准备提交代码、要求预检查、或提及 go lint/单元测试/commit 前检查时使用。
---

# Go 项目 Commit 前预检查

在用户准备提交 Go 代码或明确要求「预检查」「commit 前检查」时，按本流程执行：先 lint，再单元测试；任一步失败则进入自动修复，直至通过或无法自动修复时给出明确说明。

## 前置条件

- 工作目录为 Go 项目根（含 `go.mod`）；若仓库多模块，在对应模块根下执行。
- 已安装 `golangci-lint`（推荐）或至少 `go vet`；单元测试使用 `go test`。

## 检查流程

按顺序执行，**全部通过**后才视为预检查通过：

1. **Lint 检查**：必须通过。
2. **单元测试**：必须全部通过。

### 1. 确定项目根并执行 Lint

若项目根有 `Makefile` 且包含 `lint` 目标，优先在项目根执行：

```bash
make lint
```

否则在项目根执行：

```bash
golangci-lint run ./...
```

若未安装 golangci-lint，可退化为：

```bash
go vet ./...
```

### 2. 执行单元测试

若有 `make test` 且与项目约定一致（如排除集成测试/ mocks），优先：

```bash
make test
```

否则：

```bash
go test ./... -gcflags=all=-l -v
```

排除目录时与项目一致（例如常见排除 `*_test` 外的集成目录、mocks），如：

```bash
go test $(go list ./... | grep -v /server/tests/ | grep -v /server/mocks) -gcflags=all=-l -v
```

---

## 自动修复策略

### Lint 失败时

1. **先尝试自动修复**（在项目根执行）：
   - 格式化：`gofmt -w .` 或 `go fmt ./...`
   - 若使用 golangci-lint：`golangci-lint run ./... --fix`
2. **再次运行 lint**：若通过，继续执行单元测试。
3. **若仍有报错**：根据 linter 报告逐条处理（未使用变量、拼写、复杂度等），修改代码后重新执行 lint，直到通过。

### 单元测试失败时

1. **根据测试输出定位**：失败用例名、文件、行号、错误信息。
2. **自动可做**：修正断言、修复因重构导致的调用错误、补全缺失的返回值或 mock、修正导入与包名等。
3. **修改后必须**：重新运行完整单元测试，确保全部通过且无新增失败。
4. **无法自动时**：明确列出失败用例与原因，并说明需要用户确认或手动修改的地方。

---

## 流程小结

```
执行 make lint 或 golangci-lint run ./...
    → 未通过：gofmt / golangci-lint --fix → 再 lint → 仍失败则按报告改代码 → 循环直到通过
    → 通过 ↓
执行 make test 或 go test ./...
    → 未通过：根据失败信息改代码/测试 → 再跑全量测试 → 循环直到通过或无法自动修复
    → 全部通过 → 预检查通过，可进行 git commit
```

---

## 更多参考

- 常用命令与常见 lint 问题处理见 [reference.md](reference.md)。

## 注意事项

- **不要跳过检查**：在用户明确要求预检查或 commit 前检查时，必须跑完 lint 与 test 并处理失败，不要只给命令让用户自己执行。
- **与项目约定一致**：若仓库有 `Makefile`、`project.sh` 或 CI 配置（如 `.golangci.yml`），排除目录、测试参数应与之一致。
- **修复后必须复测**：每次修改后都要重新执行对应的 lint 或全量 test，避免引入新问题。
