# Skill Runtime Profile 执行最小实现

## 文档范围

本文档只描述当前代码里已经落下来的最小执行实现。

本文不讨论：

- `skill.runtime.yaml` 的目标协议
- 注册时自动解析与自动落库
- 面向模型的能力摘要设计

这些见：

- [skill_runtime_yaml_design.md](/Users/chenshu/Code/github.com/kweaver-ai/adp/execution-factory/operator-integration/docs/design/features/skill_runtime_yaml_design.md)

## 1. 当前实现边界

- Skill 资源仍由 `t_skill_repository` 和 `t_skill_file_index` 承载
- Runtime Profile 由独立表 `t_skill_runtime_profile` 承载
- Runtime Profile 既可以通过内部接口维护，也已经支持注册 zip 时从 `skill.runtime.yaml` 自动导入
- 执行阶段复用现有 session pool
- 执行阶段不再走临时 Python runner 下载 ZIP
- AOI 直接调用 sandbox control plane 的 package materialize 与 task workspace prepare

当前内部接口：

- `GET /api/agent-operator-integration/internal-v1/skills/:skill_id/runtime-profiles/:entrypoint`
- `POST /api/agent-operator-integration/internal-v1/skills/:skill_id/runtime-profiles/:entrypoint`
- `POST /api/agent-operator-integration/internal-v1/skills/:skill_id/runtime-profiles/:entrypoint/execute`

## 2. 当前执行策略

执行链路如下：

1. 根据 `skill_id + entrypoint` 读取 runtime profile
2. 重建 Skill ZIP
3. 从 session pool 借用一个 session
4. 把 Skill ZIP 上传到该 session 的 workspace
5. 调用 sandbox `packages/materialize`
6. 调用 sandbox `tasks/prepare`
7. 解析输入并写入 task `input_dir`
8. 渲染 command template
9. 通过 `execute-sync` 以 `shell` 执行
10. 列出 `output/` 目录文件并回传

### 2.1 当前状态门禁

针对 runtime profile 状态，当前最小实现的行为是：

- `published`
  - internal `/content` 可见
  - execute 可执行
- `unpublish`
  - internal `/content` 可见
  - execute 可执行
- `offline`
  - internal `/content` 不可见
  - execute 不可执行

public `/content` 的 `runtime_capabilities` 当前只暴露 `published`。

## 3. 当前模板变量

### 3.1 保留变量

- `{{skill_id}}`
- `{{skill_version}}`
- `{{entrypoint}}`
- `{{runtime_type}}`

### 3.2 输入变量

- `{{foo}}`
- `{{inputs.foo}}`

### 3.3 输出变量

- `{{bar}}`
- `{{outputs.bar}}`

说明：

- 如果 `output_schema` 定义了 `bar`，执行服务会自动把它映射到当前 task 的 `output/` 目录
- 可用于 `{{output_path}}`、`{{outputs.images}}` 这类模板写法

### 3.4 sandbox 目录变量

- `{{package_root}}`
- `{{task_root}}`
- `{{input_dir}}`
- `{{output_dir}}`
- `{{tmp_dir}}`
- `{{logs_dir}}`

说明：

- `command[0]` 当前只允许白名单运行器：`python`、`python3`、`bash`、`sh`、`node`、`nodejs`
- `command` 中的脚本路径统一按 package root 相对路径描述，例如 `scripts/to_pdf.py`
- 执行服务固定在 `package_root` 下执行 shell
- 注册或手工写入 runtime profile 时，`command` 中如果出现 `/workspace/...`、`/runtime/...`、`.tasks/...` 这类 sandbox 固定路径会直接拒绝

## 4. 当前输入映射

### 4.1 支持的输入类型

- 普通字符串
- `text`
- `path`
- `inline_file`
- `inline_file_base64`
- `artifact_ref`
- `resource_ref`
- `oss_object`

### 4.2 输入校验

当前执行前会基于 `input_schema` 做最小校验：

- `required: true` 的输入缺失时直接返回 `400`
- `default` 会在执行前注入为有效输入，不需要调用方重复传递稳定默认值
- `enum` 当前对字符串输入生效，非法值直接返回 `400`
- `type: text|string` 仅接受普通字符串或 `{type:text, value:...}`
- `type: file` 接受字符串路径，或 `path` / `artifact_ref` / `resource_ref` / `oss_object` / `inline_file` / `inline_file_base64`
- `type: directory|dir` 接受字符串路径，或 `{type:path, value:...}`

当前仍未支持复杂 JSON Schema，只支持上述最小约束

### 4.3 当前行为

对于：

- `inline_file`
- `inline_file_base64`
- `artifact_ref`
- `resource_ref`
- `oss_object`

执行服务会：

1. 先把内容 materialize 到 task `input_dir`
2. 再把模板变量替换成 sandbox 内绝对路径

### 4.4 示例

普通字符串：

```json
{
  "inputs": {
    "input_pdf": "/workspace/data/a.pdf"
  }
}
```

内联文件：

```json
{
  "inputs": {
    "config": {
      "type": "inline_file",
      "filename": "config.json",
      "content": "{\"a\":1}"
    }
  }
}
```

资源引用：

```json
{
  "inputs": {
    "input_pdf": {
      "type": "artifact_ref",
      "storage_id": "s1",
      "storage_key": "path/to/a.pdf"
    }
  }
}
```

## 5. 当前输出

当前 `ExecuteSkillResp` 会返回：

- `skill_id`
- `skill_version`
- `entrypoint`
- `session_id`
- `runtime_type`
- `exit_code`
- `stdout`
- `stderr`
- `error_message`
- `profile`
- `return_value`

其中 `return_value` 当前包含：

- `task_id`
- `task_root`
- `package_target`
- `package_checksum`
- `directories`
- `effective_inputs`
- `input_mappings`
- `output_refs`
- `output_artifacts`
- `output_files`
- `output_file_count`
- `warnings`

说明：

- `output_refs` 是当前最小结构化输出引用，包含 `session_id`、`container_path`、`abs_path`、声明的 `type`
- 它仍然是 sandbox 级引用，还不是平台级 artifact/resource id
- `output_artifacts` 是当前最小 OSS 回收结果，当前仅对声明为 `file` 且实际生成成功的输出生效
- `output_files`、`output_refs`、`output_artifacts` 当前统一使用 workspace 相对路径形式的 `container_path`，例如 `.tasks/skill/<task_id>/output/result.pdf`
- `warnings` 用于承载非主执行链路错误，当前至少包括：
  - `output_file_list_failed`
  - `output_artifact_persist_failed`
- 这些 warning 不会改变主执行成功/失败语义，但调用方不能再把“没有 output_artifacts”直接等价成“没有输出”

## 6. 当前实现与目标设计的差距

当前已经有：

- zip 中 `skill.runtime.yaml` 的自动解析
- runtime profile 的自动写入
- `/content` 返回 `runtime_capabilities`
- `ExecuteSkill` 基于 `input_schema` 做最小必填、默认值、枚举和类型校验
- `output_schema` 自动映射到 task `output/` 目录

当前还没有：

- output 统一回收为 artifact/resource
- input/output 的平台级引用层与统一对象模型
- 复杂 schema 校验与更细粒度的类型系统

## 7. 后续直接演进点

1. output 文件上传并回写统一 artifact id
2. input/output 从路径变量升级为统一 resource/artifact ref
3. `input_schema` / `output_schema` 支持更完整的 schema 约束
