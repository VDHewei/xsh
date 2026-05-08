# ONNX 本地推理服务

## 系统默认模型
- https://huggingface.co/deepseek-ai/DeepSeek-R1-Distill-Qwen-1.5B
- https://huggingface.co/yasserrmd/glm5.1-distill-onnx

---

## 功能需求

### 1. Model CLI 接入 xsh 命令
- 支持 `model search <关键词>` 搜索 HuggingFace 模型
- 支持 `model list` 列出本地模型
- 支持 `model select <模型名>` 选择/验证本地模型
- 支持 `model download <仓库ID> --dir <目录>` 下载模型 (支持代理和镜像)

### 2. ONNX 本地 LLM 推理
- 基于 `onnxruntime_purego` 库加载 ONNX 模型并推理
- xsh cli 支持 `-m <模型目录>` 指定模型, `-p <提示语>` 指定 prompt
- 支持流式输出 `-stream`

### 3. LLM 推理 + 任务规划 + TUI 交互
- 测试 LLM 本地推理识别 `tests/data/prod-migration-form-uat.txt`
- 结合任务规划和 TUI 交互完成完整链路

---

## 测试要求

1. 测试通过记录到 `tests/data/` 目录，并记录通过原因
2. 测试分三类:
   - **通过测试**: 验证正确路径和预期行为
   - **非通过测试**: 验证错误处理和边界条件
   - **模糊测试**: 随机/异常输入验证健壮性
3. 测试形式:
   - **代码测试**: 单元测试、集成测试、场景链路测试
   - **非代码测试**: Agent 随机测试、人机交互测试
4. 测试维度:
   - **功能测试**: 所有 task 类型 [http, grpc, ssh, wait, ask, check]
   - **代码覆盖**: 达到合理覆盖率
   - **性能测试**: 推理延迟、并发、内存占用

---

## 开发要求

1. LLM 功能集中到 `pkg/llm/` 开发
2. 功能集中封装到 `pkg/<功能子目录>/`
3. 测试需有: unit 单元测试、功能步骤聚合测试、场景链路调用测试
4. `cmd/xsh/` 目录下仅有 `main.go`，其他文件归档到对应目录或测试目录
5. 任务细化拆分，子任务规划存储到 `plans/<总任务名>/**.md`
6. HTTP 方法全面覆盖: GET/POST/DELETE/PUT/PATCH/HEAD/OPTION，含 header/body 场景
7. Mock 测试服务器: `tests/servers/http`、`tests/servers/grpc`、`tests/servers/ssh`
8. 错误场景覆盖: 请求失败、失败重试、请求超时、请求成功、请求异常
9. `@ask:<提示语+意图>` 和 `@check:<意图>` 接入 LLM 推理
10. 核心任务(LLM 调用 + 任务解析)测试必须通过，失败则 loop 修复，成功才跳出
11. 业务完成进度记录到 `plans/process/<task_name>.md`
12. 优先使用成熟安全的开源仓库实现，无合适实现时再自定义 ,[开源项目一般都在github.com 中查找]
13. 功能函数必须实现最小 MVP，不能 mock 或 TODO
14. TUI 支持 xsh CLI 命令 + 自定义 `commands/` 目录 `.md` 命令文件 (LLM 解析)
15. 涉及的开源项目罗列到 README，人工审核后添加到开发规划文档
16. 注意任务规划 一定要细化粒度到 功能入口/功能控制层组装层/功能模块/功能函数/业务数据模型 [十分重要]
17. agent 人机询问 对应仓库审核结果，并记录更新文件
---

## 开源项目清单 (待审核)

| 项目 | 用途 | 仓库 | 审核状态 |
|------|------|------|----------|
| onnxruntime_purego | ONNX 模型加载推理 | https://github.com/getcharzp/onnxruntime-genai_purego | 待审核 |
| tview | 终端 UI 框架 | https://github.com/rivo/tview | 待审核 |
| tcell | 终端单元格库 | https://github.com/gdamore/tcell | 待审核 |
| purego | CGo-free FFI | https://github.com/ebitengine/purego | 待审核 |
| - | SSH 客户端 | golang.org/x/crypto (标准库扩展) | 待审核 |
| - | gRPC 客户端 | google.golang.org/grpc | 待审核 |
| openClaw/goose | 自定义命令解析参考 | 待确认 | 待调研 |
