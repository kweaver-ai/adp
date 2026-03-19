# 🏗️ Design Doc: 执行工厂 Skill 接入与管理

> 状态: Draft
> 负责人: 待确认
> Reviewers: 待确认
> 关联 PRD: ../../product/prd/agent_skill接入与管理-prd.md

---

# 📌 1. 概述（Overview）

## 1.1 背景

- 当前现状：
  - 执行工厂已具备 `operator`、`toolbox`、`mcp` 等资源的接入与管理能力，但尚未形成 Skill 资源的统一接入、治理与运行时读取方案。
  - Skill 既包含 `SKILL.md` 指令正文，也包含模板、脚本、参考资料等附属文件，其使用方式天然不同于单一配置类资源。
  - 项目代码库内已存在 Skill 相关接口、模型、路由与测试实现，可作为本设计的实现输入，但不作为设计结构本身。

- 存在问题：
  - Skill 注册入口、元数据模型、文件索引和对象存储策略尚未在正式设计文档中统一定义。
  - Skill 的管理接口、市场接口、运行时读取接口边界不清晰，容易造成职责混用。
  - `SKILL.md` 与附件文件的读取存在渐进式加载诉求，若设计不当会导致上下文冗余、响应放大和权限暴露。
  - 删除、下载、读取等链路同时依赖 DB 与对象存储，若缺少状态机与补偿策略，容易出现部分成功和残留数据。

- 业务 / 技术背景：
  - 执行工厂需要将外部 Skill 作为独立资源域纳入六边形架构，支持标准化注册、查询、下载、删除和运行时按需读取。
  - 首版目标是建立 Skill 的整体实现基线，为后续版本治理、自动绑定、运行沙箱和市场化能力预留扩展点。
  - 本文档聚焦“整体实现及详细设计”，仅描述首版正式方案的系统边界、架构和实现细节。

---

## 1.2 目标

- 为执行工厂定义一套可落地的 Skill 资源模型、接口分层和存储方案。
- 支持 Skill 注册、列表、详情、内容读取、文件读取、zip 下载和删除的首版闭环能力。
- 将 Skill 使用链路收敛为“列表匹配 -> 内容读取 -> 文件读取”的渐进式加载模式。
- 明确 Skill 与权限体系、业务域体系、对象存储、审计日志和运行时调用链的对接方式。
- 通过 `draft/active/error/deleting` 状态机和补偿策略控制 DB 与对象存储的一致性风险。

---

## 1.3 非目标（Out of Scope）

- 首版不支持 Skill 在线编辑与内容回写。
- 首版不支持 Skill 版本对比、回滚和多分支管理。
- 首版不支持 Agent 与 Skill 的自动推荐、自动绑定和依赖求解。
- 首版不负责 Skill 沙箱执行能力，仅负责 Skill 资源管理与运行时只读访问。

---

## 1.4 术语说明（Optional）

| 术语 | 说明 |
|------|------|
| Skill | Agent 可消费的能力包，包含 `SKILL.md` 与附属文件 |
| Skill Content | Skill 的 `SKILL.md` 正文，对外以只读内容提供 |
| Skill 文件 | Skill 内部模板、脚本、参考资料、配置等附属文件 |
| 管理接口 | 面向 Skill 提供方和资源维护方的接口 |
| 市场接口 | 面向 Skill 发现与选择场景的接口 |
| 读取接口 | 面向 Guide 与文件渐进式读取的接口 |
| 业务域 | 由平台统一治理的资源可见性和归属范围 |

---

# 🏗️ 2. 整体设计（HLD）

> 本章节关注系统“怎么搭建”，不涉及具体实现细节

---

## 🌍 2.1 系统上下文（C4 - Level 1）

### 参与者
- 用户：Skill 提供方、Agent 配置方、平台运行链路
- 外部系统：统一认证/鉴权服务、业务域服务、对象存储、数据库
- 第三方服务：待确认

### 系统关系

    [Skill 提供方 / Agent 配置方] → [执行工厂 Skill 服务] → [MySQL / 对象存储 / 权限服务 / 业务域服务]

---

## 🧱 2.2 容器架构（C4 - Level 2）

| 容器 | 技术栈 | 职责 |
|------|--------|------|
| API Service | Go + Gin | 暴露 Skill 管理、市场、读取相关 HTTP 接口 |
| Core Service | Go | 实现 Skill 注册、查询、删除、下载、读取等核心业务逻辑 |
| Parser | Go | 解析 `SKILL.md`、抽取 frontmatter、校验文件清单 |
| Asset Store | Go + 对象存储 SDK | 管理 Skill 附件内容上传、下载、删除 |
| Storage | MySQL + Object Storage | 保存 Skill 主数据、文件索引与文件正文 |

---

### 容器交互

    Client → API Service → Core Service → Parser / Repository / Asset Store → MySQL / Object Storage

---

## 🧩 2.3 组件设计（C4 - Level 3）

### Skill Core 组件

| 组件 | 职责 |
|------|------|
| SkillRegistry | 处理注册、删除、列表、详情、下载等管理能力 |
| SkillMarket | 处理市场列表、市场详情与检索场景 |
| SkillReader | 提供 Guide 与 Skill 文件的只读访问 |
| SkillRuntimeBinder | 为 Agent 运行链路提供 Skill 绑定查询与只读解析能力 |
| SkillParser | 解析 `SKILL.md`、校验 frontmatter、构造标准化元数据 |
| SkillAssetStore | 管理对象存储中的附件上传、读取与删除 |
| SkillRepository | 管理 Skill 主表与文件索引表的数据访问 |
| Governance Adapter | 对接 AuthService、BusinessDomainService 和审计能力 |

---

## 🔄 2.4 数据流（Data Flow）

### 主流程

    上传 zip / content → 解析 SKILL.md → 写入主表和索引 → 上传附件到对象存储 → Skill 激活 → 列表检索 → Guide 读取 → 文件按需读取

### 子流程（可选）

    删除请求 → 状态置为 deleting → 清理对象存储 → 清理索引与主记录 → 写入审计日志

---

## ⚖️ 2.5 关键设计决策（Design Decisions）

| 决策 | 说明 |
|------|------|
| Skill 作为独立资源域 | Skill 与 `operator`、`toolbox`、`mcp` 职责不同，需要独立生命周期和接口模型 |
| `SKILL.md` 入主表，附件入对象存储 | Guide 查询频繁且体量可控，适合保留在主表；附件正文放对象存储以避免 DB 膨胀 |
| 管理/市场/读取三类接口分层 | 资源维护、资源发现、运行时读取的调用方和返回语义不同，必须拆分 |
| 采用渐进式加载 | 列表不返回正文与文件，Guide 单独读取，文件再按需读取，减少不必要装载 |
| 引入 `deleting` 和 `error` 状态 | 处理对象存储与 DB 间的部分成功问题，为补偿任务提供稳定中间态 |
| 权限与业务域通过统一服务治理 | 避免在 Skill 资源表内固化本地归属判断规则，保持与现有资源域一致 |

---

## 🚀 2.6 部署架构（Deployment）

- 部署环境：K8s
- 拓扑结构：Skill 能力作为执行工厂服务中的一组新增逻辑、路由和数据访问组件，与现有资源域共用服务进程、数据库和对象存储
- 扩展策略：API 实例水平扩展；对象存储与数据库按现有平台能力扩展

---

## 🔐 2.7 非功能设计

### 性能
- 列表查询只返回轻量摘要，避免读取 `SKILL.md` 与文件清单
- Skill 内容读取直接走主表字段，减少对象存储依赖
- 文件读取按 `skill_id + rel_path` 精确命中索引，避免全量扫描

### 可用性
- 通过 `error`、`deleting` 状态和补偿任务处理跨存储介质失败
- 对外将 `deleting` 视为不存在，避免暴露中间不一致状态

### 安全
- 所有管理、市场、文件读取接口均接入统一权限服务
- 文件读取执行路径标准化与访问级别校验，禁止路径穿越
- 日志中禁止打印 Skill 文件正文和敏感配置

### 可观测性
- tracing：覆盖注册、删除、内容读取、文件读取、下载链路
- logging：记录关键状态变更、权限校验失败、对象存储异常
- metrics：记录注册成功率、读取成功率、删除补偿次数、接口时延

---

# 🔧 3. 详细设计（LLD）

> 本章节关注“如何实现”，开发可直接参考

---

## 🌐 3.1 API 设计

### Skill 管理接口

**Endpoint:** `POST /api/agent-operator-integration/v1/skills`

**Request:**

```json
{
  "file_type": "zip | content",
  "content": "optional when file_type=content",
  "source": "upload_zip",
  "extend_info": {
    "tag": "demo"
  }
}
```

**Response:**

```json
{
  "skill_id": "skill-xxx",
  "name": "demo-skill",
  "description": "demo",
  "version": "1.0.0",
  "status": "active",
  "files": [
    {
      "rel_path": "templates/reply.md",
      "access_level": "runtime_read"
    }
  ]
}
```

### Skill 列表接口

**Endpoint:** `GET /api/agent-operator-integration/v1/skills`

**Request:**

```json
{
  "page": 1,
  "page_size": 10,
  "name": "demo",
  "status": "active",
  "source": "upload_zip"
}
```

**Response:**

```json
{
  "total": 1,
  "data": [
    {
      "skill_id": "skill-xxx",
      "name": "demo-skill",
      "description": "demo",
      "status": "active"
    }
  ]
}
```

### Skill 详情接口

**Endpoint:** `GET /api/agent-operator-integration/v1/skills/{skill_id}`

**Request:**

```json
{
  "skill_id": "skill-xxx"
}
```

**Response:**

```json
{
  "skill_id": "skill-xxx",
  "name": "demo-skill",
  "description": "demo",
  "version": "1.0.0",
  "status": "active",
  "source": "upload_zip",
  "extend_info": {}
}
```

### Skill 删除接口

**Endpoint:** `DELETE /api/agent-operator-integration/v1/skills/{skill_id}`

**Request:**

```json
{
  "skill_id": "skill-xxx"
}
```

**Response:**

```json
{
  "success": true
}
```

### Skill 下载接口

**Endpoint:** `GET /api/agent-operator-integration/v1/skills/{skill_id}/download`

**Request:**

```json
{
  "skill_id": "skill-xxx"
}
```

**Response:**

```json
{
  "download_mode": "stream",
  "file_name": "demo-skill.zip"
}
```

### Skill 市场接口

**Endpoint:** `GET /api/agent-operator-integration/v1/skills/market`

**Request:**

```json
{
  "page": 1,
  "page_size": 10,
  "name": "demo",
  "tag": "assistant"
}
```

**Response:**

```json
{
  "total": 1,
  "data": [
    {
      "skill_id": "skill-xxx",
      "name": "demo-skill",
      "description": "demo",
      "source": "upload_zip"
    }
  ]
}
```

### Skill 内容读取接口

**Endpoint:** `GET /api/agent-operator-integration/v1/skills/{skill_id}/content`

**Request:**

```json
{
  "skill_id": "skill-xxx"
}
```

**Response:**

```json
{
  "skill_id": "skill-xxx",
  "content": "# Instructions\n...",
  "status": "active",
  "files": [
    {
      "rel_path": "templates/reply.md",
      "access_level": "runtime_read"
    }
  ]
}
```

### Skill 文件读取接口

**Endpoint:** `POST /api/agent-operator-integration/v1/skills/{skill_id}/files/read`

**Request:**

```json
{
  "rel_path": "templates/reply.md"
}
```

**Response:**

```json
{
  "skill_id": "skill-xxx",
  "rel_path": "templates/reply.md",
  "content": "file body",
  "mime_type": "text/markdown"
}
```

---

## 🗂️ 3.2 数据模型

### SkillRepository

| 字段 | 类型 | 说明 |
|------|------|------|
| skill_id | string | Skill 主键 |
| name | string | Skill 名称 |
| description | string | Skill 描述 |
| skill_content | text | `SKILL.md` 正文 |
| version | string | Skill 版本 |
| status | enum | `draft/active/error/deleting` |
| source | string | 来源，如 `upload_zip`、`raw_content` |
| extend_info | json | 扩展元数据 |
| dependencies | json | 依赖信息 |
| file_manifest | json | 精简文件清单 |
| create_user | string | 创建人 |
| update_user | string | 更新人 |
| create_time | datetime | 创建时间 |
| update_time | datetime | 更新时间 |
| delete_time | datetime | 删除时间，待确认 |

### SkillFileIndex

| 字段 | 类型 | 说明 |
|------|------|------|
| skill_id | string | Skill 主键 |
| rel_path | string | 标准化后的相对路径 |
| path_hash | string | 路径哈希，用于辅助检索 |
| storage_key | string | 对象存储键 |
| file_type | string | 文件类型，如 `instruction`、`template`、`script` |
| content_sha256 | string | 文件内容校验值 |
| size | int64 | 文件大小 |
| mime_type | string | MIME 类型 |
| access_level | enum | `public_manifest/runtime_read/restricted` |
| create_time | datetime | 创建时间 |
| update_time | datetime | 更新时间 |

### SkillSpec

| 字段 | 类型 | 说明 |
|------|------|------|
| name | string | `SKILL.md` 中的必填名称 |
| description | string | `SKILL.md` 中的必填描述 |
| version | string | 可选版本，缺省时由系统生成 |
| dependencies | array | 声明的依赖信息 |
| metadata | map | 扩展信息 |
| files | array | 文件声明，仅作为校验与展示参考 |

---

## 💾 3.3 存储设计

- 存储类型：MySQL + 对象存储
- 数据分布：
  - MySQL 保存 Skill 主数据、状态、依赖、文件索引和轻量清单
  - 对象存储保存 Skill 附件正文
  - `SKILL.md` 正文同时保存在主表中，用于 Guide 快速读取
- 索引设计：
  - 主表按 `skill_id` 主键索引
  - 列表场景按 `status`、`name`、`source`、`update_time` 建立查询索引
  - 文件索引表按 `(skill_id, rel_path)` 建唯一索引
  - `path_hash` 仅用于辅助查找，不能替代真实路径校验

---

## 🔁 3.4 核心流程（详细）

### Skill 注册流程

1. 接收 `zip` 或 `content` 注册请求并完成权限校验。
2. 对 `zip` 请求执行解包，校验包内存在 `SKILL.md`。
3. 解析 `SKILL.md` frontmatter 与正文，校验 `name`、`description` 等必填字段。
4. 标准化所有附件路径，拒绝绝对路径、空路径和路径穿越。
5. 生成 Skill 主记录、文件索引记录和文件清单，状态初始为 `draft`。
6. 在 DB 事务中写入主表和索引表。
7. 提交事务后将附件正文上传至对象存储。
8. 全部上传成功后将状态更新为 `active`；若上传失败则更新为 `error` 并记录补偿信息。

### Skill 删除流程

1. 校验 Skill 存在且调用方具有删除权限。
2. 将 Skill 状态更新为 `deleting`。
3. 查询该 Skill 对应的全部文件索引。
4. 删除对象存储中的附件内容。
5. 删除文件索引记录。
6. 删除或归档主记录，并清理业务域关联与权限策略。
7. 记录审计日志；若任一清理步骤失败，则保留 `deleting` 状态等待补偿。

### Guide 读取流程

1. 校验 Skill 存在且状态为 `active`。
2. 校验当前调用方具备读取权限。
3. 从主表读取 `skill_content` 与精简文件清单。
4. 返回 Guide 正文和可展示文件清单，不访问对象存储。

### Skill 文件读取流程

1. 校验 Skill 存在且状态为 `active`。
2. 对 `rel_path` 执行路径标准化。
3. 校验调用方具有文件读取权限。
4. 按 `(skill_id, rel_path)` 查询索引记录。
5. 校验 `access_level` 允许当前读取场景。
6. 从对象存储读取文件正文并校验 `content_sha256`。
7. 返回文件内容与 MIME 信息。

---

## 🧠 3.5 关键逻辑设计

### 路径标准化模块
- 使用 `filepath.Clean` 处理原始路径。
- 将 `\` 统一替换为 `/`。
- 去除前导 `/`。
- 拒绝 `.`、`..`、空字符串和目录路径。

### `SKILL.md` 解析模块
- 将文档拆分为 frontmatter 与 markdown body。
- frontmatter 抽取 `name`、`description`、`version`、`dependencies`、`metadata`、`files`。
- markdown body 直接保存为 `skill_content`。
- 若关键字段缺失或 YAML 非法，返回明确参数错误。

### 状态机模块
- `draft`：主记录和索引已创建，附件尚未全部就绪。
- `active`：主记录、索引和对象存储均已就绪，可对外提供能力。
- `error`：注册或补偿失败，需要人工或异步任务修复。
- `deleting`：删除流程中间态，对外不可见。

### 权限与业务域治理模块
- 注册前校验 `create` 权限。
- 查询、读取、下载、删除分别校验对应资源操作权限。
- 业务域关联通过统一服务建立和解除，不在 Skill 表内固化归属字段。
- 市场接口在权限过滤后叠加业务域可见性过滤。

### 下载打包模块
- 以主表中的 `skill_content` 和文件索引为数据源重建 zip 包。
- 打包时必须写入 `SKILL.md`。
- 仅打包已登记且允许导出的文件。
- 任一对象缺失时整体失败，不返回不完整压缩包。

---

## ❗ 3.6 错误处理

- `SKILL.md` 缺失或解析失败：返回参数错误，不创建 Skill。
- 文件路径非法：返回参数错误，拒绝注册。
- 权限校验失败：返回无权限错误并记录审计日志。
- Skill 不存在或处于 `deleting`：对外返回未找到。
- 对象存储上传失败：状态置为 `error`，触发补偿。
- 对象存储删除失败：保留 `deleting` 状态，等待补偿任务处理。
- 下载阶段存在索引缺失或对象缺失：返回下载失败，不返回部分成功结果。

---

## ⚙️ 3.7 配置设计

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| skill.storage.bucket | 待确认 | Skill 附件对象存储桶 |
| skill.storage.prefix | `skills/` | Skill 对象存储路径前缀 |
| skill.register.max_zip_size | 待确认 | 注册允许的最大 zip 大小 |
| skill.register.max_file_count | 待确认 | 注册允许的最大文件数量 |
| skill.download.max_package_size | 待确认 | 下载打包允许的最大包体积 |
| skill.compensation.enabled | `true` | 是否启用补偿任务 |
| skill.compensation.scan_interval | 待确认 | 补偿任务扫描周期 |
| skill.file.read.allowed_levels | `runtime_read,public_manifest` | 默认允许运行时读取的文件级别 |

---

## 📊 3.8 可观测性实现

- tracing：
  - 在注册、删除、Guide 读取、文件读取、下载链路创建 span
  - 记录 `skill_id`、操作类型、状态变更和对象存储调用结果

- metrics：
  - `skill_register_total{result}`
  - `skill_read_total{type,result}`
  - `skill_delete_total{result}`
  - `skill_compensation_total{status,result}`
  - `skill_request_latency_ms{api}`

- logging：
  - 记录注册解析失败、路径非法、权限拒绝、对象存储异常、补偿执行结果
  - 日志仅记录 `skill_id`、`rel_path`、状态和错误摘要，不记录文件正文
