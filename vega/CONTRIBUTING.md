# 贡献指南 (Contributing Guide)

感谢您对Vega项目的关注和支持！我们非常欢迎您为本项目做出贡献。本文档提供了参与项目开发、报告问题和提出改进建议的相关信息。

## 目录

- [行为准则](#行为准则)
- [贡献类型](#贡献类型)
- [开发环境设置](#开发环境设置)
- [提交代码流程](#提交代码流程)
- [代码规范](#代码规范)
- [测试要求](#测试要求)
- [文档规范](#文档规范)
- [分支管理策略](#分支管理策略)
- [问题报告](#问题报告)
- [联系方式](#联系方式)

## 行为准则

请遵循 [Contributor Covenant](https://www.contributor-covenant.org/version/2/0/code_of_conduct/) 行为准则，保持友好、尊重的交流环境。

## 贡献类型

您可以以以下方式为项目做贡献：

- 提交Bug报告
- 提出功能建议
- 改进文档
- 编写测试用例
- 修复Bug
- 实现新功能
- 优化性能
- 审查代码
- 推广项目

## 开发环境设置

### 前置要求

- Git
- Java 8+ (用于Java项目)
- Go 1.19+ (用于Go项目)
- Maven 3.6+
- Docker & Docker Compose
- Node.js (如需前端开发)

### 项目克隆

```bash
git clone https://github.com/eisoo/vega.git
cd vega
```

### 子项目依赖安装

根据您要开发的模块，进入相应的目录并安装依赖：

```bash
# Java项目
cd data-connection && mvn install
cd vega-gateway && mvn install
cd vega-metadata && mvn install

# Go项目
cd mdl-data-model && go mod tidy
cd mdl-uniquery && go mod tidy
```

## 提交代码流程

### 1. Fork仓库

点击GitHub页面上的"Fork"按钮，将仓库fork到您的个人账户下。

### 2. 创建分支

```bash
git checkout -b feature/your-feature-name
# 或
git checkout -b bugfix/issue-number
```

### 3. 开发与测试

编写代码并确保通过所有相关测试。

### 4. 提交更改

```bash
git add .
git commit -m "feat: 添加新功能描述"
# 或
git commit -m "fix: 修复问题描述"
```

#### 提交消息规范

提交消息应遵循 [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) 规范：

- `feat`: 新功能
- `fix`: Bug修复
- `docs`: 文档更新
- `style`: 代码格式调整
- `refactor`: 代码重构
- `test`: 测试相关
- `chore`: 构建过程或辅助工具变更

### 5. 推送到远程分支

```bash
git push origin feature/your-feature-name
```

### 6. 创建Pull Request

前往原始仓库页面，创建Pull Request。

## 代码规范

### Java代码规范

- 遵循Google Java Style Guide
- 使用UTF-8编码
- 使用4个空格缩进（不使用Tab）
- 类名使用UpperCamelCase，方法和变量使用lowerCamelCase
- 常量使用UPPER_SNAKE_CASE

### Go代码规范

- 遵循Go官方代码风格
- 使用gofmt格式化代码
- 使用golint检查代码质量
- 变量使用camelCase
- 导出的标识符需要有文档注释

### SQL代码规范

- 关键字使用大写（SELECT, FROM, WHERE等）
- 表名和字段名使用小写，单词间用下划线分隔
- 使用有意义的命名约定

### 命名约定

- 使用有意义的变量、函数和类名
- 避免缩写和简写
- 保持命名一致性

## 测试要求

### 单元测试

- 所有核心功能必须包含单元测试
- 测试覆盖率不应低于80%
- 测试用例应该覆盖正常情况和异常情况

### 集成测试

- 对于涉及多个组件的功能，需要编写集成测试
- 验证组件间的交互是否符合预期

### 性能测试

- 对于性能敏感的功能，需要提供性能测试
- 确保新功能不会显著影响系统性能

## 文档规范

### 代码注释

- 公共方法和类必须有Javadoc/GoDoc注释
- 复杂逻辑需要内联注释说明
- 注释应保持与代码同步更新

### API文档

- 更新REST API时，相应更新API文档
- 提供请求示例和响应格式

### README文件

- 每个模块应有详细的README.md文件
- 包含安装、配置和使用说明

## 分支管理策略

我们使用Git Flow工作流：

- `main`: 生产环境分支，仅接受发布标签
- `develop`: 开发主分支，合并所有功能分支
- `feature/*`: 功能开发分支
- `release/*`: 发布准备分支
- `hotfix/*`: 紧急修复分支

## 问题报告

### 报告Bug

当报告Bug时，请包含以下信息：

- 版本号
- 操作系统
- Java/Go版本
- 重现步骤
- 预期行为
- 实际行为
- 错误日志（如有）

### 提出功能建议

请提供以下信息：

- 清晰简洁的功能描述
- 使用场景
- 解决的问题
- 建议的实现方案

## 联系方式

- GitHub Issues: 用于Bug报告和功能建议
- 邮箱: [维护者邮箱地址]
- 讨论区: [讨论平台链接]

---

再次感谢您的贡献！如果您有任何疑问，请随时联系我们。