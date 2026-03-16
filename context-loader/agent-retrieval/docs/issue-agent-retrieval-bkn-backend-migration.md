## Issue 标题

fix: agent-retrieval 使用 bkn-backend 替代 ontology-manager 时访问配置不一致

## Issue 正文

**环境:**
- Go: 1.24+（请根据实际环境填写）
- 操作系统: macOS / Kubernetes 集群（请根据实际环境填写）
- 模块: ADP - context-loader/agent-retrieval
- 数据库: （若与本问题无关可省略）
- 相关依赖服务: bkn-backend（原 ontology-manager）、ontology-query

**问题简述:**
`agent-retrieval` 服务仍然按旧的 `ontology-manager` 服务配置（host 和 API 前缀）访问本体管理接口，而底层服务已经重命名为 `bkn-backend`。这会导致在仅部署 `bkn-backend` 的环境下，`agent-retrieval` 中所有依赖本体信息 / 行动类定义的能力在运行时调用失败（404 / service not found）。

---

**复现步骤:**
1. 在集群中部署最新的 `bkn-backend`（替代原 `ontology-manager`，服务名为 `bkn-backend-svc`，对外暴露路由前缀为 `/api/bkn-backend`）。
2. 使用当前的 Helm chart 在集群中部署 `context-loader/agent-retrieval`，保持默认配置：
   - `context-loader/agent-retrieval/helm/agent-retrieval/values.yaml` 中：
     - `depServices.ontology-manager.privateHost: ontology-manager-svc`
   - `context-loader/agent-retrieval/server/infra/config/agent-retrieval.yaml` 中：
     - `ontology_manager.private_host: "ontology-manager-svc.anyshare"`
3. 通过上游（如 Context Loader 工具链）触发一个依赖本体信息或行动类定义的请求，例如：
   - 使用 kn-logic-property-resolver / 对象类查询 / 行动类查询等能力。
4. 观察 `agent-retrieval` 日志与调用链路。

**期望行为:**
- `agent-retrieval` 通过正确的 host 和 API 前缀访问 `bkn-backend`：
  - host 指向 `bkn-backend-svc`（或对应域名，如 `bkn-backend-svc.anyshare`）。
  - API 前缀使用 `/api/bkn-backend`。
- 依赖知识网络 / 对象类 / 关系类 / 行动类等能力的接口调用成功，不出现 404 或 DNS/service not found。

**实际行为:**
- `agent-retrieval` 仍按照旧的 `ontology-manager` 服务配置访问：
  - host: `ontology-manager-svc` 或 `ontology-manager-svc.anyshare`
  - API 前缀: `/api/ontology-manager`
- 在只部署 `bkn-backend` 而不再部署 `ontology-manager` 的环境中：
  - 本体相关请求会出现 DNS 解析失败 / 连接失败（service not found），或 404（指向已废弃的 `/api/ontology-manager` 路径）。
- 导致所有依赖本体定义（知识网络详情、对象类、关系类、行动类等）的功能链路不可用。

**错误日志（示例，实际请替换为真实日志）:**
```text
[OntologyManagerAccess] GetKnowledgeNetworkDetail request failed, err: Get "http://ontology-manager-svc.anyshare/api/ontology-manager/in/v1/knowledge-networks/xxx": dial tcp: lookup ontology-manager-svc.anyshare: no such host
```

或

```text
[OntologyManagerAccess] GetObjectTypeDetail get resp failed, [http://ontology-manager-svc/api/ontology-manager/in/v1/knowledge-networks/xxx/object-types/...], 404 ...
```

**最小复现代码（MRC）:**
- 关键调用入口：
  - 文件: `context-loader/agent-retrieval/server/drivenadapters/ontology_manager.go`
  - 构造基础 URL 的位置：
    ```go
    func NewOntologyManagerAccess() interfaces.OntologyManagerAccess {
        omAccessOnce.Do(func() {
            conf := config.NewConfigLoader()
            omAccess = &ontologyManagerAccess{
                logger:     conf.GetLogger(),
                baseURL:    conf.OntologyManager.BuildURL("/api/ontology-manager"),
                httpClient: rest.NewHTTPClient(),
            }
        })
        return omAccess
    }
    ```
- 上游任意使用 `OntologyManagerAccess` 的接口（例如 `GetKnowledgeNetworkDetail`）即可触发问题。

---

**影响范围分析:**

1. **运行时代码中的 API 前缀**
   - 文件: `context-loader/agent-retrieval/server/drivenadapters/ontology_manager.go`
   - 现状:
     - `NewOntologyManagerAccess` 使用：
       ```go
       baseURL: conf.OntologyManager.BuildURL("/api/ontology-manager")
       ```
     - 后续接口均在此基础上拼接 `/in/v1/knowledge-networks/...` 等路径：
       - `GetKnowledgeNetworkDetail`
       - `SearchObjectTypes` / `GetObjectTypeDetail`
       - `SearchRelationTypes` / `GetRelationTypeDetail`
       - `SearchActionTypes` / `GetActionTypeDetail`
       - 以及本体构建任务相关接口（Create/List Jobs）
     - 当网关仅暴露 `/api/bkn-backend` 路由时，这些调用会 404。

2. **Helm values 中的下游服务 host**
   - 文件: `context-loader/agent-retrieval/helm/agent-retrieval/values.yaml`
   - 现状:
     ```yaml
     depServices:
       ontology-manager:
         privateHost: ontology-manager-svc
         privatePort: 13014
         privateProtocol: http
     ```
   - `context-loader/agent-retrieval/helm/agent-retrieval/templates/configmap.yaml` 将其渲染为应用配置：
     ```yaml
     ontology_manager:
       private_protocol: {{ index .Values "depServices" "ontology-manager" "privateProtocol" | quote }}
       private_host: {{ index .Values "depServices" "ontology-manager" "privateHost" | quote }}
       private_port: {{ index .Values "depServices" "ontology-manager" "privatePort" }}
     ```
   - 最终生成的 `agent-retrieval.yaml` 中，`ontology_manager.private_host` 仍为 `ontology-manager-svc`。

3. **本地/非 Helm 部署配置**
   - 文件: `context-loader/agent-retrieval/server/infra/config/agent-retrieval.yaml`
   - 现状:
     ```yaml
     ontology_manager:
       private_protocol: "http"
       private_host: "ontology-manager-svc.anyshare"
       private_port: 13014
     ```
   - 若使用该文件直接启动本地服务，同样会连接到已废弃的 `ontology-manager-svc.anyshare`。

4. **文档与设计说明（次要影响）**
   - 多处 PRD / 设计文档以及 `context-loader/agent-retrieval/README*.md` 中仍以 `ontology-manager` 命名描述依赖。
   - 这是命名一致性问题，对运行时行为影响较小，可在后续文档重构中统一为 `bkn-backend`。

---

**建议的修复方案:**

1. **更新运行时代码的 API 前缀**
   - 文件: `context-loader/agent-retrieval/server/drivenadapters/ontology_manager.go`
   - 将 `NewOntologyManagerAccess` 中的基础 URL 前缀从：
     ```go
     baseURL: conf.OntologyManager.BuildURL("/api/ontology-manager")
     ```
     修改为：
     ```go
     baseURL: conf.OntologyManager.BuildURL("/api/bkn-backend")
     ```
   - 前提假设：`bkn-backend` 在 `/api/bkn-backend` 前缀下，对外暴露与 `ontology-manager` 等价的 `/in/v1/knowledge-networks/...` 路由（知识网络 / 对象类 / 关系类 / 行动类 / 任务管理等）。

2. **更新 Helm 部署配置中的服务 host**
   - 文件: `context-loader/agent-retrieval/helm/agent-retrieval/values.yaml`
   - 将：
     ```yaml
     depServices:
       ontology-manager:
         privateHost: ontology-manager-svc
     ```
     更新为：
     ```yaml
     depServices:
       ontology-manager:
         privateHost: bkn-backend-svc
     ```
   - 保持 key 名为 `ontology-manager` 以兼容现有配置加载逻辑，仅变更实际 host。
   - 通过 `helm template` 或实际部署验证生成的配置文件中：
     - `ontology_manager.private_host` 应为 `bkn-backend-svc`。

3. **更新本地配置文件**
   - 文件: `context-loader/agent-retrieval/server/infra/config/agent-retrieval.yaml`
   - 将：
     ```yaml
     ontology_manager:
       private_host: "ontology-manager-svc.anyshare"
     ```
     更新为：
     ```yaml
     ontology_manager:
       private_host: "bkn-backend-svc.anyshare"
     ```
     或者根据实际 DNS 命名规范调整。

4. **回归验证建议**
   - 在应用上述修改后：
     1. 部署/启动更新后的 `agent-retrieval` 与 `bkn-backend`。
     2. 通过上游调用一个依赖本体定义的典型链路（如逻辑属性解析 / 对象类查询 / 行动类调度等）。
     3. 预期：
        - 日志中不再出现访问 `ontology-manager-svc` 或 `/api/ontology-manager` 的请求；
        - 所有本体相关请求指向 `bkn-backend-svc` 且使用 `/api/bkn-backend` 前缀；
        - 链路整体返回成功结果。

---

**补充说明（可选）：**
- 目前 Issue 主要聚焦于运行时访问失败问题。后续可以单独起一个文档与命名重构的 Issue，将文档中的 `ontology-manager` 逐步迁移为 `bkn-backend`，并在 BKN 体系内统一命名。

