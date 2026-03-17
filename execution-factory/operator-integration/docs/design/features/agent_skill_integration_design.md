# Agent Skill 接入与管理设计方案

## 1. 文档信息

- Status: Draft
- Owner: Codex
- Last Updated: 2026-03-16
- Reference Branch: `feature/skill`

## 2. 背景与目标

执行工厂当前已经具备 `operator`、`toolbox`、`mcp` 三类核心资源管理能力，但缺少面向 Agent Skill 的统一接入与运行时读取能力。参考 `feature/skill` 分支可以确认，Skill 至少包含以下几个不可拆开的能力点：

- Skill 的注册、删除、查询
- `SKILL.md` 的解析、存储和读取
- Skill 元数据管理
- Skill 内部附属文件管理
- Agent 在执行期按需读取 Skill 文件
- Skill 文件访问边界与安全控制

本设计的目标不是直接合并 `feature/skill` 的原型实现，而是在执行工厂现有六边形架构中，为 Skill 定义一套职责清晰、可分阶段落地、可测试、可审计的正式方案。

### 2.1 Goals

- 将 Skill 作为执行工厂中的独立资源域接入
- 支持 Skill 全生命周期管理：注册、删除、详情、列表、文件读取
- 支持 `SKILL.md` 解析与标准化元数据提取
- 支持 Skill 附件文件的索引、存储与受控访问
- 为 Agent 运行时提供稳定的 Skill 绑定与读取能力
- 保证对象存储、DB 索引、访问控制之间的一致性和可回收性

### 2.2 Non-goals

- 本期不实现 Skill 编辑器 UI
- 本期不实现跨仓库 Skill 依赖解析与自动安装
- 本期不实现 Skill 执行沙箱本身，只负责 Skill 的管理和运行时读取
- 本期不把 Skill 强行并入 `toolbox` 或 `operator` 模型

## 3. 核心判断

### 3.1 Skill 在执行工厂中的定位

Skill 不是 `operator`、不是 `toolbox`、也不是 `mcp server`。它更接近“Agent 可消费的能力包”，包含两类内容：

- 指令内容：`SKILL.md`
- 附件内容：脚本、模板、配置、示例、参考文档等内部文件

因此 Skill 应作为独立资源域管理，但在运行时由 Agent 执行链按需挂载使用。

### 3.2 与 `feature/skill` 分支的关系

`feature/skill` 已验证以下抽象是成立的：

- `skill_repository` 主表保存 Skill 元数据与 `SKILL.md` 内容
- `skill_file_index` 子表保存附件文件索引
- Skill 文件内容不直接塞入 DB，而放对象存储
- `SKILL.md` 是注册入口的事实标准

但该分支仍存在以下问题，不建议直接照搬：

- `manager` 同时承担解析、注册、查询、打包、存储、文件下载、就绪检查，职责过重
- 缺少删除链路、分页查询、权限模型、补偿逻辑
- 文件查找仅依赖 `hash(rel_path)`，缺少规范化路径校验和审计字段
- DB 事务与对象存储写入存在部分成功风险，缺少状态机与补偿机制

## 4. 总体方案

本方案将 Skill 能力拆分为 4 个子域。

### 4.1 Skill Registry

负责 Skill 的资源生命周期管理：

- 注册 Skill
- 删除 Skill
- 查询 Skill 列表与详情
- 维护 Skill 状态、版本、拥有者、扩展信息

### 4.2 Skill Asset Store

负责 Skill 附件文件的落库索引与对象存储：

- 文件路径标准化
- 文件索引入库
- 文件内容写入对象存储
- 文件删除与补偿清理

### 4.3 Skill Reader

负责运行时只读访问：

- 读取 `SKILL.md`
- 读取 Skill 文件内容
- 返回文件清单
- 基于索引和策略进行文件访问校验

### 4.4 Skill Runtime Binding

负责 Skill 与 Agent 执行链的关联：

- 查询 Agent 当前绑定的 Skill 列表
- 将 Skill 摘要注入 LLM 上下文
- 在模型或执行链需要时按需读取指定文件
- 对文件读取进行最小权限控制

## 5. 架构设计

### 5.1 分层映射

- `server/interfaces`
  - 新增 `logics_skill.go`
  - 定义管理、读取、运行时绑定接口
- `server/logics/skill`
  - `registry.go`
  - `asset_store.go`
  - `reader.go`
  - `parser.go`
  - `binding.go`
- `server/dbaccess`
  - `skill_repository.go`
  - `skill_file_index.go`
- `server/drivenadapters`
  - 复用对象存储适配器；若现有 S3 能力不足，则新增 skill 专用封装
- `server/driveradapters`
  - 新增 Skill 管理端和运行时读接口 handler

### 5.2 设计原则

- `SKILL.md` 是 Skill 的主事实来源，但不能作为唯一存储
- 元数据与文件索引必须可查询、可审计、可回收
- 运行时仅读，不允许通过 Reader 侧修改资源
- 文件访问必须经过“Skill 是否存在 + 文件是否登记 + 路径是否合法”三重校验
- 对象存储是内容真实源，DB 是检索与控制面

## 6. 数据模型设计

### 6.1 主表 `t_skill_repository`

建议在 `feature/skill` 原型基础上扩展，不直接沿用原表定义。

核心字段：

- `f_skill_id`
- `f_name`
- `f_description`
- `f_instructions`
- `f_version`
- `f_status`
- `f_source`
- `f_owner_type`
- `f_owner_id`
- `f_extend_info`
- `f_dependencies`
- `f_file_manifest`
- `f_create_time`
- `f_create_user`
- `f_update_time`
- `f_update_user`
- `f_delete_time`
- `f_delete_user`

新增建议：

- `f_status`
  - `draft`
  - `active`
  - `deleting`
  - `deleted`
- `f_file_manifest`
  - 保存精简后的文件清单 JSON，用于详情页和快速校验
- `f_source`
  - 区分 `upload_zip`、`raw_content`、`system_builtin`

### 6.2 子表 `t_skill_file_index`

建议保留双表结构，但增加约束与审计字段。

核心字段：

- `f_skill_id`
- `f_rel_path`
- `f_path_hash`
- `f_storage_key`
- `f_file_type`
- `f_content_sha256`
- `f_size`
- `f_mime_type`
- `f_access_level`
- `f_create_time`
- `f_update_time`

说明：

- `f_path_hash` 用于快速定位，不能替代 `f_rel_path`
- `f_content_sha256` 用于内容校验和对象存储补偿
- `f_access_level` 用于运行时控制读取范围

### 6.3 文件分类建议

- `instruction`
  - 固定保留给 `SKILL.md`
- `script`
  - 代码脚本
- `template`
  - 提示词模板、配置模板
- `reference`
  - 文档、示例、知识参考
- `asset`
  - 静态资源
- `config`
  - 配置类文件

### 6.4 运行时访问级别

- `public_manifest`
  - 可出现在文件清单中
- `runtime_read`
  - 允许运行时读取
- `restricted`
  - 注册后保留，但默认不允许 LLM 直接读取

### 6.5 `SKILL.md` 解析模型

建议约束为：

```md
---
name: xxx
description: xxx
version: 1.0.0
dependencies:
  - skill:other-skill
metadata:
  domain: customer_service
  author: alice
files:
  - path: templates/reply.md
    access: runtime_read
---

# Instructions
...
```

建议解析后形成两部分：

- 结构化 frontmatter
- markdown body 作为 `instructions`

建议校验字段：

- `name` 必填
- `description` 必填
- `version` 选填，缺省则由服务生成
- `dependencies` 选填
- `metadata` 选填
- `files` 选填，仅作为声明信息，真实文件仍以上传包内容为准

## 7. 接口设计

### 7.1 管理端接口

#### 7.1.1 注册 Skill

`POST /api/agent-operator-integration/v1/skills`

支持两种模式：

- `file_type=zip`
- `file_type=content`

请求字段建议：

- `user_id`
- `business_domain`
- `file_type`
- `file`
- `source`
- `extend_info`

响应字段建议：

- `skill_id`
- `name`
- `description`
- `version`
- `status`
- `files`

#### 7.1.2 删除 Skill

`DELETE /api/agent-operator-integration/v1/skills/:skill_id`

删除策略：

- 先标记 `deleting`
- 删除文件索引和对象存储内容
- 成功后置为 `deleted` 或执行物理删除

不建议直接无状态硬删除。

#### 7.1.3 查询 Skill 列表

`GET /api/agent-operator-integration/v1/skills`

支持筛选：

- `page`
- `page_size`
- `sort_by`
- `sort_order`
- `name`
- `status`
- `create_user`
- `source`

#### 7.1.4 查询 Skill 详情

`GET /api/agent-operator-integration/v1/skills/:skill_id`

返回：

- 基本元数据
- `SKILL.md` 摘要
- 文件清单
- 扩展元数据

### 7.2 读取接口

#### 7.2.1 读取 Skill Guide

`GET /api/agent-operator-integration/v1/skills/:skill_id/guide`

返回：

- `content`
- `files`

#### 7.2.2 读取 Skill 文件

`POST /api/agent-operator-integration/v1/skills/:skill_id/files/read`

请求：

- `rel_path`

返回：

- `skill_id`
- `rel_path`
- `content`
- `mime_type`

### 7.3 运行时绑定接口

如果执行工厂需要显式提供运行时绑定能力，建议新增内部接口：

- `GET /internal-v1/agents/:agent_id/skills`
- `POST /internal-v1/agents/:agent_id/skills/resolve`

其中 `resolve` 返回：

- Skill 摘要
- 可读文件清单
- Guide 内容

这个接口的意义是把“Skill 管理”与“Agent 如何绑定 Skill”解耦。

## 8. 核心时序

### 8.1 注册时序

1. 接收 zip 或原始 `SKILL.md`
2. 解包并查找 `SKILL.md`
3. 标准化所有相对路径
4. 解析 frontmatter 与正文
5. 构造 Skill 主记录
6. 构造文件索引记录
7. 开启 DB 事务，写入主表和索引表
8. 提交事务后上传对象存储
9. 若对象存储上传失败，触发补偿任务并标记状态异常
10. 返回注册结果

### 8.2 删除时序

1. 校验 Skill 是否存在
2. 状态更新为 `deleting`
3. 查询全部文件索引
4. 删除对象存储文件
5. 删除索引记录
6. 删除或归档主记录
7. 写入审计日志

### 8.3 运行时读取时序

1. 校验 Skill 是否存在且状态为 `active`
2. 标准化请求路径
3. 校验路径不包含越权片段
4. 查询文件索引
5. 校验 `access_level`
6. 从对象存储取内容
7. 返回内容

## 9. 文件访问控制设计

这是本方案的高风险部分，必须单独定义。

### 9.1 基础规则

- 禁止绝对路径
- 禁止 `..` 路径穿越
- 禁止空路径
- 禁止目录读取，只允许文件读取
- 请求路径必须命中索引表中的标准化 `rel_path`

### 9.2 标准化规则

对 `rel_path` 执行：

- `filepath.Clean`
- 替换 `\` 为 `/`
- 去除前导 `/`
- 拒绝 `.`、`..`、空字符串

### 9.3 访问判定

运行时读取文件必须同时满足：

- Skill 存在
- Skill 状态可用
- 文件已登记
- 请求路径与登记路径一致
- `access_level` 允许读取

### 9.4 LLM 最小权限原则

不建议将所有 Skill 文件暴露给 Agent。默认策略应为：

- 仅暴露 `SKILL.md`
- 仅暴露 `runtime_read` 文件
- `restricted` 文件只能被后端内部逻辑读取，不能直接拼入模型上下文

## 10. 存储与一致性

### 10.1 存储策略

- DB 存控制面数据
- 对象存储存正文内容
- `SKILL.md` 正文可同时保存在主表，便于快速查询
- 附件文件正文只放对象存储

### 10.2 一致性问题

`feature/skill` 的主要隐患是：

- DB 成功
- 对象存储部分失败

因此正式方案建议：

- 主记录状态增加 `draft/active/error/deleting`
- 注册成功条件改为“DB 成功且对象存储全部成功”
- 对象存储失败时将主记录标记 `error`
- 由异步补偿任务清理残留索引和对象

### 10.3 补偿任务

建议新增定时任务扫描：

- `status=error`
- `status=deleting`

处理内容：

- 重试上传
- 清理残留对象
- 修复状态

## 11. 与现有资源域的关系

### 11.1 与 Operator

- Operator 是可执行单元
- Skill 是 Agent 的能力包与上下文包
- Skill 可以引用 Operator，但不应变成 Operator 的一个变种

### 11.2 与 Toolbox

- Toolbox 是工具集合
- Skill 可声明推荐工具或依赖工具箱
- 但 Skill 不负责工具托管

### 11.3 与 MCP

- MCP 是工具服务的运行协议与托管方式
- Skill 可以依赖 MCP 工具
- Skill 不应管理 MCP 生命周期

### 11.4 推荐关系建模

本期建议先将外部依赖放入 `f_dependencies` JSON：

- `operator_ids`
- `toolbox_ids`
- `mcp_ids`
- `skill_ids`

等 Skill 规模变大后，再独立关系表。

## 12. 模块拆分建议

为避免一次改动超过 3 个文件后失控，后续实现建议拆成 4 个子任务。

### 12.1 子任务一：Skill 领域骨架

- `interfaces/logics_skill.go`
- `interfaces/model/skill_repository.go`
- `interfaces/model/skill_file_index.go`

### 12.2 子任务二：注册与查询

- `logics/skill/parser.go`
- `logics/skill/registry.go`
- `dbaccess/skill_repository.go`

### 12.3 子任务三：文件索引与读取

- `logics/skill/asset_store.go`
- `logics/skill/reader.go`
- `dbaccess/skill_file_index.go`

### 12.4 子任务四：接口接入与运行时绑定

- `driveradapters/skill/*.go`
- `driveradapters/rest_public_handler.go` 或新增专属 handler
- `logics/skill/binding.go`

## 13. 测试与验证

### 13.1 单元测试

- `SKILL.md` 解析成功
- 缺少 frontmatter 解析失败
- 缺少 `name/description` 解析失败
- zip 中缺失 `SKILL.md` 注册失败
- 路径标准化正确处理 `a/../b.txt`
- 拒绝绝对路径与穿越路径
- `runtime_read` 文件可读，`restricted` 文件拒绝
- 删除后再次读取返回 404

### 13.2 集成测试

- 注册 zip Skill 后可查询详情
- 注册后可读取 `guide`
- 注册后可读取指定文件
- 删除 Skill 后对象存储文件被清理
- 对象存储上传失败时 Skill 状态为 `error`

### 13.3 回归点

- 不影响现有 `operator`、`toolbox`、`mcp` 路由
- 不引入新的公共执行路径权限漏洞
- 不把大文件正文写入 DB 导致查询退化

## 14. 风险与权衡

### 14.1 将 `SKILL.md` 全文存在 DB

优点：

- 查询快
- 不依赖对象存储即可返回 guide

缺点：

- 主表膨胀

结论：

- 保留 `SKILL.md` 正文入主表
- 其它附件禁止入主表

### 14.2 是否直接沿用 `feature/skill` 的 S3 Engine

优点：

- 可快速复用

缺点：

- 当前接口过于薄，未体现状态、校验、补偿和命名策略

结论：

- 可复用底层对象存储 client
- 不建议直接复用整个 manager

### 14.3 是否使用物理删除

优点：

- 简单

缺点：

- 审计差
- 补偿难

结论：

- 先软删除或状态删除
- 异步清理后再物理回收

## 15. 验收清单

- Skill 具备注册、删除、列表、详情能力
- `SKILL.md` 可被解析并返回 guide
- Skill 附件文件可索引、可读取、可删除
- 文件访问具备路径标准化和访问级别控制
- DB 与对象存储具备失败补偿策略
- Skill 与 `operator/toolbox/mcp` 的职责边界清晰
- 运行时可以最小权限绑定并读取 Skill 内容

## 16. 失败条件

- Skill 被实现成 `toolbox` 的别名资源
- `SKILL.md` 只存对象存储，导致 guide 查询依赖对象存储可用性
- 文件读取不做路径规范化，允许穿越或越权
- 对象存储写失败后仍返回成功且无补偿
- 把 Skill 管理、运行时绑定、对象存储访问全部堆进一个 manager

## 17. 建议的首期实现范围

建议首期只落地最小闭环：

- 注册 Skill
- 查询列表/详情
- 读取 `guide`
- 读取单文件
- 删除 Skill

暂缓：

- Skill 打包导出
- 跨 Skill 依赖解析
- Agent 侧自动发现推荐 Skill
- Skill 版本多分支管理

这样可以先把资源模型和安全边界做对，再逐步扩展运行时能力。
