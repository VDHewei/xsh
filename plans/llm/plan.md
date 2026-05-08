# LLM 本地推理服务 - 任务规划 (函数粒度)

## 系统默认模型
- `deepseek-ai/DeepSeek-R1-Distill-Qwen-1.5B` (已下载, 已测试通过)
- `yasserrmd/glm5.1-distill-onnx` (候选模型, 需下载测试)

## 现有代码函数清单 (当前状态)

### 业务数据模型 (internal/types)
| 结构体 | 字段 | 状态 |
|--------|------|------|
| `Task` | Type, Raw, HTTP, SSH, GRPC, Ask, Wait, Check | 已有 |
| `HTTPTask` | Method, URL, Headers, Body | 已有 |
| `SSHTask` | Host, Port, User, Command | 已有 |
| `GRPCTask` | Host, Port, Method, Headers, Body | 已有 |
| `AskTask` | Prompt | 已有 |
| `WaitTask` | Duration | 已有 |
| `CheckTask` | Prompt | 已有 |
| `TaskResult` | Task, Success, Output, Error | 已有 |

### 功能模块/函数 (当前实现)

**pkg/llm/model.go** - 模型加载模块
| 函数 | 类型 | 状态 |
|------|------|------|
| `NewModel(name) *Model` | 创建 | 已有 |
| `NewConfig() *Config` | 创建 | 已有 |
| `(*Model).Load(dir)` | 加载 | 已有 |
| `(*Model).Unload()` | 卸载 | 已有 |
| `(*Model).Infer(prompt)` | 推理 | 已有 |
| `(*Model).InferWithConfig(prompt, cfg)` | 推理 | 已有 |
| `(*Model).InferStream(prompt, opts, cb)` | 推理 | 已有 |
| `generate(prompt, opts)` | 内部 | 已有 |
| `generateStream(prompt, opts, cb)` | 内部 | 已有 |
| `GetModelPath(dir, name)` | 查询 | 已有 |
| `ListModels(dir)` | 查询 | 已有 |
| `MockInfer(prompt)` | 测试 | 已有 |
| `Generate(dir, prompt, opts)` | 高级API | 已有 |
| `GenerateStream(dir, prompt, opts, cb)` | 高级API | 已有 |

**pkg/llm/analyzer.go** - 任务分析模块
| 函数 | 类型 | 状态 |
|------|------|------|
| `NewTaskAnalyzer()` | 创建 | 已有 |
| `(*TaskAnalyzer).SetModel(m)` | 设置 | 已有 |
| `(*TaskAnalyzer).AnalyzeContent(content)` | 分析 | 已有 |
| `(*TaskAnalyzer).AnalyzeFile(file)` | 分析 | 已有 |
| `InferWithPrompt(prompt)` | 推理 | 已有 |
| `buildAnalyzePrompt(content)` | 内部 | 已有 |
| `parseLLMResult(result)` | 内部 | 已有 |
| `parseHeadersAndBody(rest)` | 内部 | 已有 |
| `parseGRPCTarget(target)` | 内部 | 已有 |
| `extractURLs(content)` | 内部 | 已有 |

**pkg/llm/cli.go** - 模型CLI模块
| 函数 | 类型 | 状态 |
|------|------|------|
| `ModelSearch(query)` | 搜索 | 已有 |
| `ModelList(dir)` | 列表 | 已有 |
| `ModelSelect(dir, name)` | 选择 | 已有 |
| `ParseModelCommand(args)` | 路由 | 已有 |

**pkg/llm/download.go** - 模型下载模块
| 函数 | 类型 | 状态 |
|------|------|------|
| `NewDownloadConfig()` | 创建 | 已有 |
| `DownloadFromHuggingFace(repoID, cfg)` | 下载 | 已有 |
| `DownloadWithProxy(repoID, proxy)` | 下载 | 已有 |
| `DownloadWithMirror(repoID, mirror)` | 下载 | 已有 |
| `SearchModels(query, cfg)` | 搜索 | 已有 |
| `GetModelInfo(repoID, cfg)` | 查询 | 已有 |
| `SetProxy(proxy)` | 配置 | 已有 |
| `SetMirror(mirror)` | 配置 | 已有 |

**internal/parser/parser.go** - 任务解析模块
| 函数 | 类型 | 状态 |
|------|------|------|
| `ParseFile(filename)` | 解析 | 已有 |
| `parseLine(line)` | 内部 | 已有 (仅2种模式) |

**internal/executor/executor.go** - 任务执行模块
| 函数 | 类型 | 状态 |
|------|------|------|
| `ExecuteTasks(tasks)` | 编排 | 已有 |
| `executeHTTP(task)` | 执行 | 已有 (无header/body/retry/timeout) |
| `executeWait(task)` | 执行 | 已有 |
| `executeTask(task)` | 分发 | 已有 (缺少ssh/grpc分支) |

**internal/tui/tui.go** - TUI交互模块
| 函数 | 类型 | 状态 |
|------|------|------|
| `RunInteractive()` | 入口 | 已有 |
| `NewApp()` | 创建 | 已有 |
| `(*App).setupUI()` | UI | 已有 |
| `(*App).loadTasks()` | 加载 | 已有 |
| `(*App).updateTaskList()` | UI | 已有 |
| `(*App).runNextTask()` | 执行 | 已有 |
| `(*App).executeTask()` | 分发 | 已有 (内联HTTP/Wait) |
| `(*App).showConfirmDialog()` | UI | 已有 |
| `(*App).saveResults()` | 保存 | 已有 |

---

## T1. 功能入口: Model CLI 路由接入

### T1 数据模型 (需新增)
| 结构体 | 字段 | 说明 |
|--------|------|------|
| `ModelCandidate` | RepoID string, Name string, Default bool | 默认候选模型 |

### T1 功能函数 (需新增/修改)

**层级: 功能入口 → cmd/xsh/main.go**
| 函数 | 操作 | 说明 |
|------|------|------|
| `main()` | 修改 | 增加 "model" 子命令路由分发 |
| `parseModelArgs(args []string)` | 新增 | 解析 model 子命令参数 |

**层级: 功能模块 → pkg/llm/cli.go**
| 函数 | 操作 | 说明 |
|------|------|------|
| `handleModelSearch(args)` | 已有 | 无需修改 |
| `handleModelList(args)` | 已有 | 无需修改 |
| `handleModelSelect(args)` | 已有 | 无需修改 |
| `handleModelDownload(args)` | 新增 | 解析 download 子命令参数 |
| `ParseModelCommand(args)` | 修改 | 增加 download 路由, 返回 error |
| `printModelUsage()` | 修改 | 增加 download 用法 |

**层级: 功能模块 → pkg/llm/download.go**
| 函数 | 操作 | 说明 |
|------|------|------|
| `DownloadFromHuggingFace(repoID, cfg)` | 已有 | 无需修改 |
| `resolveModelRepo(name string) string` | 新增 | 根据候选模型名解析完整 RepoID |

**层级: 功能模块 → pkg/llm/model.go**
| 函数 | 操作 | 说明 |
|------|------|------|
| `DefaultCandidateModels() []ModelCandidate` | 新增 | 返回默认候选模型列表(含 GLM5.1) |
| `GetModelPath(dir, name)` | 已有 | 需支持候选模型名匹配 |

### T1 验收标准 (入口层)
- `xsh model list` → 列出本地模型 (含候选模型标记)
- `xsh model search qwen` → 搜索 HuggingFace
- `xsh model select deepseek-r1-distill-qwen-1.5B` → 验证模型存在
- `xsh model download yasserrmd/glm5.1-distill-onnx --dir ./models` → 下载到指定目录

---

## T2. Mock 测试服务器

### T2 数据模型
复用 `internal/types` 中已有 Task/HTTPTask/GRPCTask/SSHTask 结构体

### T2 功能函数

**层级: 功能入口 → tests/servers/http/main.go**
| 函数 | 操作 | 说明 |
|------|------|------|
| `main()` | 新增 | HTTP mock server 入口 |
| `registerRoutes()` | 新增 | 注册所有路由/端点 |

**层级: 功能函数 → tests/servers/http/**
| 函数 | 操作 | 说明 |
|------|------|------|
| `handlerGet(w, r)` | 新增 | 成功/慢速/500 分支 |
| `handlerPost(w, r)` | 新增 | 解析body, 成功/500 分支 |
| `handlerPut(w, r)` | 新增 | 成功/慢速 分支 |
| `handlerPatch(w, r)` | 新增 | 成功/500 分支 |
| `handlerDelete(w, r)` | 新增 | 成功/500 分支 |
| `handlerHead(w, r)` | 新增 | 仅返回header |
| `handlerOptions(w, r)` | 新增 | 返回Allow header |
| `handlerTimeout(w, r)` | 新增 | sleep 后响应 |
| `handlerRetry(w, r)` | 新增 | 首次500, 后续200 |
| `handlerEcho(w, r)` | 新增 | 回显 header+body |

**层级: 功能入口 → tests/servers/grpc/main.go**
| 函数 | 操作 | 说明 |
|------|------|------|
| `main()` | 新增 | gRPC mock server 入口 |
| `newServer()` | 新增 | 创建gRPC服务实例 |

**层级: 功能函数 → tests/servers/grpc/**
| 函数 | 操作 | 说明 |
|------|------|------|
| `(*MockService).HealthCheck(ctx, req)` | 新增 | 成功/失败分支 |
| `(*MockService).Execute(ctx, req)` | 新增 | 回显参数 |
| `(*MockService).Timeout(ctx, req)` | 新增 | 延迟后返回 |

**层级: 功能入口 → tests/servers/ssh/main.go**
| 函数 | 操作 | 说明 |
|------|------|------|
| `main()` | 新增 | SSH mock server 入口 |

**层级: 功能函数 → tests/servers/ssh/**
| 函数 | 操作 | 说明 |
|------|------|------|
| `handleSession(channel, requests)` | 新增 | 处理 exec/shell 请求 |
| `handleExec(command)` | 新增 | 执行命令回显 |

### T2 验收标准
- HTTP server: `curl -X POST localhost:PORT/echo -d '{...}'` 返回 header+body
- gRPC server: `grpcurl localhost:PORT list` 列出服务
- SSH mock: connect and exec commands

---

## T3. 任务执行器完善 (SSH/gRPC/重试/超时)

### T3 数据模型 (需新增)
| 结构体 | 字段 | 说明 |
|--------|------|------|
| `RetryConfig` | MaxRetries int, RetryDelay time.Duration, BackoffFactor float64 | 重试配置 |
| `TimeoutConfig` | ConnectTimeout time.Duration, RequestTimeout time.Duration | 超时配置 |
| `ErrorCategory` | Category string, Code int, Message string | 错误分类 |

### T3 功能函数

**层级: 功能控制层 → internal/executor/executor.go**
| 函数 | 操作 | 说明 |
|------|------|------|
| `ExecuteTasks(tasks)` | 修改 | 增加 SSH/gRPC 分支 + 错误汇总 |
| `executeTask(task)` | 修改 | 增加 TaskTypeSSH/TaskTypeGRPC 分支 |
| `executeHTTP(task)` | 修改 | 增加 header/body 发送, 超时控制, 错误分类 |
| `executeSSH(task)` | 新增 | 新建SSH连接 → 执行命令 → 返回结果 |
| `executeGRPC(task)` | 新增 | 创建gRPC连接 → 调用方法 → 返回结果 |
| `executeHTTPWithRetry(task, cfg)` | 新增 | 带重试的HTTP执行 |
| `executeSSHWithRetry(task, cfg)` | 新增 | 带重试的SSH执行 |
| `executeGRPCWithRetry(task, cfg)` | 新增 | 带重试的gRPC执行 |
| `classifyError(err)` | 新增 | 将error分类为网络/超时/服务端/客户端 |
| `formatResult(success, output, err)` | 新增 | 统一结果格式化 |

**层级: 功能函数 → internal/executor/retry.go** (新文件)
| 函数 | 操作 | 说明 |
|------|------|------|
| `NewRetryConfig() *RetryConfig` | 新增 | 默认重试配置 |
| `DefaultRetryConfig() *RetryConfig` | 新增 | 3次重试, 1s间隔, 2x退避 |
| `(*RetryConfig).Do(fn)` | 新增 | 执行带重试的函数 |

**层级: 功能函数 → internal/executor/timeout.go** (新文件)
| 函数 | 操作 | 说明 |
|------|------|------|
| `NewTimeoutConfig() *TimeoutConfig` | 新增 | 默认超时配置 |
| `DefaultTimeoutConfig() *TimeoutConfig` | 新增 | 5s连接, 30s请求 |
| `withTimeout(ctx, timeout, fn)` | 新增 | 带超时的执行包装 |

### T3 依赖的外部开源库
| 包 | 用途 | 引入原因 |
|----|------|----------|
| `golang.org/x/crypto/ssh` | SSH 客户端 | Go 标准库扩展, 成熟稳定 |
| `google.golang.org/grpc` | gRPC 客户端 | 官方实现, 生态完善 |
| `google.golang.org/protobuf` | protobuf 支持 | gRPC 依赖 |

### T3 验收标准
- SSH 任务: 连接 → 执行 → 输出结果
- gRPC 任务: 连接 → 调用方法 → 输出结果
- HTTP 任务: 发送 header+body → 解析响应
- 失败场景: 自动重试3次 → 超时2次 → 失败返回
- 错误分类: network/timeout/server/client 明确

---

## T4. @ask/@check LLM 集成

### T4 数据模型 (需新增)
| 结构体 | 字段 | 说明 |
|--------|------|------|
| `AskResult` | Prompt string, Response string, Suggestion string | Ask 结果 |
| `CheckResult` | Prompt string, Passed bool, Reason string, Context string | Check 结果 |

### T4 功能函数

**层级: 功能模块 → pkg/llm/ask_executor.go** (新文件)
| 函数 | 操作 | 说明 |
|------|------|------|
| `NewAskExecutor(model *Model) *AskExecutor` | 新增 | 创建 Ask 执行器 |
| `(*AskExecutor).Execute(task *types.AskTask) (*AskResult, error)` | 新增 | 发送 prompt 给 LLM, 解析建议 |
| `(*AskExecutor).buildAskPrompt(task) string` | 新增 | 构造 ask 提示语模板 |

**层级: 功能模块 → pkg/llm/check_executor.go** (新文件)
| 函数 | 操作 | 说明 |
|------|------|------|
| `NewCheckExecutor(model *Model) *CheckExecutor` | 新增 | 创建 Check 执行器 |
| `(*CheckExecutor).Execute(task *types.CheckTask, context string) (*CheckResult, error)` | 新增 | 结合上下文让 LLM 判断 |
| `(*CheckExecutor).buildCheckPrompt(task, ctx) string` | 新增 | 构造 check 提示语模板 |

**层级: 功能组装层 → internal/executor/executor.go**
| 函数 | 操作 | 说明 |
|------|------|------|
| `executeTask(task)` | 修改 | 增加 Ask/Check 真实 LLM 调用 |
| `(*Executor).SetLLMModel(model)` | 新增 | 注入 LLM 模型到执行器 |

**层级: 功能组装层 → internal/tui/tui.go**
| 函数 | 操作 | 说明 |
|------|------|------|
| `(*App).executeTask(task)` | 修改 | Ask task: 显示 LLM 推理结果 + 用户确认 |
| `(*App).showAskResultDialog(result)` | 新增 | 展示 LLM 建议, 用户选择继续/跳过 |
| `(*App).showCheckResultDialog(result)` | 新增 | 展示检查结果 |

### T4 验收标准
- `@ask: 需要执行数据库迁移吗?` → 返回 LLM 分析建议
- `@check: 验证迁移后健康状态` → 结合上下文返回判断
- TUI 中 Ask/Check 显示 LLM 推理结果 + 用户交互确认

---

## T5. LLM 推理测试 (迁移文件识别)

### T5 功能函数

**层级: 功能入口 → cmd/xsh/main.go**
| 函数 | 操作 | 说明 |
|------|------|------|
| `runTestMode()` | 修改 | 增加 GLM5.1 模型测试对比 |

**层级: 功能测试 → tests/llm_inference_test.go** (新文件)
| 函数 | 操作 | 说明 |
|------|------|------|
| `TestDeepSeekR1_MigrationParse()` | 新增 | DeepSeek 解析迁移文件准确度 |
| `TestGLM51_MigrationParse()` | 新增 | GLM5.1 解析迁移文件准确度 |
| `TestModelComparison()` | 新增 | 多模型对比测试 |

**层级: 场景链路测试 → tests/scenario_migration_test.go** (新文件)
| 函数 | 操作 | 说明 |
|------|------|------|
| `TestMigrationFullPipeline()` | 新增 | parse → LLM分析 → execute → 结果验证 |
| `TestMigrationTUIInteraction()` | 新增 | TUI 完整交互流程 |

### T5 验收标准
- DeepSeek R1 正确识别 8 个迁移步骤
- GLM5.1 解析结果记录对比
- 完整链路 parse → analyze → execute → verify 通过
- 测试结果记录到 tests/data/

---

## T6. 全面测试覆盖

### T6 功能函数

**层级: 单元测试 → internal/parser/parser_test.go**
| 函数 | 操作 | 说明 |
|------|------|------|
| `TestParseHTTPWithHeaders()` | 新增 | 解析带 header 的 HTTP 行 |
| `TestParseHTTPWithBody()` | 新增 | 解析带 body 的 HTTP 行 |
| `TestParseGRPCTask()` | 新增 | 解析 gRPC 任务行 |
| `TestParseSSHTask()` | 新增 | 解析 SSH 任务行 |
| `TestParseMixedTasks()` | 新增 | 混合任务行解析 |

**层级: 单元测试 → internal/executor/executor_test.go**
| 函数 | 操作 | 说明 |
|------|------|------|
| `TestExecuteHTTP_GET()` | 新增 | GET 请求 |
| `TestExecuteHTTP_POST()` | 新增 | POST + body |
| `TestExecuteHTTP_PUT()` | 新增 | PUT + body |
| `TestExecuteHTTP_PATCH()` | 新增 | PATCH + body |
| `TestExecuteHTTP_DELETE()` | 新增 | DELETE |
| `TestExecuteHTTP_HEAD()` | 新增 | HEAD |
| `TestExecuteHTTP_OPTIONS()` | 新增 | OPTIONS |
| `TestExecuteHTTP_WithHeaders()` | 新增 | 自定义 header |
| `TestExecuteHTTP_Timeout()` | 新增 | 超时场景 |
| `TestExecuteHTTP_Retry()` | 新增 | 重试场景 |
| `TestExecuteHTTP_Failure()` | 新增 | 4xx/5xx |
| `TestExecuteSSH_Success()` | 新增 | SSH 成功 |
| `TestExecuteSSH_Failure()` | 新增 | SSH 失败 |
| `TestExecuteGRPC_Success()` | 新增 | gRPC 成功 |
| `TestExecuteGRPC_Failure()` | 新增 | gRPC 失败 |
| `TestRetryConfig_Do()` | 新增 | 重试逻辑 |
| `TestClassifyError()` | 新增 | 错误分类 |

**层级: 分类测试**
| 分类 | 测试函数 |
|------|----------|
| **通过测试** | TestExecuteHTTP_GET, TestExecuteHTTP_POST, TestParseFile, TestExecuteSSH_Success |
| **非通过测试** | TestExecuteHTTP_Timeout, TestExecuteHTTP_Failure, TestExecuteSSH_Failure |
| **模糊测试** | TestParseMixedTasks, TestExecuteMultipleTasks, TestRetryConfig_Do |

**层级: 非代码测试 → tests/manual_test.md** (新文件)
| 测试项 | 说明 |
|--------|------|
| Agent 随机参数 | 随机 URL/method/header 组合 |
| 人机交互 | TUI 完整流程操作记录 |

**层级: 性能测试 → tests/bench_test.go** (新文件)
| 函数 | 操作 | 说明 |
|------|------|------|
| `BenchmarkLLMInference()` | 新增 | LLM 推理延迟基准 |
| `BenchmarkStreamInference()` | 新增 | 流式推理基准 |
| `BenchmarkHTTPExecution()` | 新增 | HTTP 执行基准 |

**层级: 归档操作**
| 操作 | 说明 |
|------|------|
| 移动 `cmd/xsh/onnx_test.go` → `tests/cmd_onnx_test.go` | 归档测试文件 |
| 删除或整合 `cmd/xsh/onnx_test.go` 中 `buildTaskPrompt` 等辅助函数 | 统一到 pkg/llm |

### T6 验收标准
- `go test ./...` 全部通过
- 7 种 HTTP 方法全部有测试覆盖
- 错误场景 (失败/重试/超时/成功/异常) 全覆盖
- 测试分类: 通过/非通过/模糊 各至少一组
- cmd/xsh/ 仅有 main.go
- 性能基准记录到 tests/data/

---

## T7. TUI 自定义 Commands 支持

### T7 数据模型 (需新增)
| 结构体 | 字段 | 说明 |
|--------|------|------|
| `CustomCommand` | Name string, File string, Content string, Tasks []*types.Task | 自定义命令 |
| `CommandLoader` | dir string, commands map[string]*CustomCommand | 命令加载器 |

### T7 功能函数

**层级: 功能模块 → pkg/llm/command_loader.go** (新文件)
| 函数 | 操作 | 说明 |
|------|------|------|
| `NewCommandLoader(dir string) *CommandLoader` | 新增 | 创建加载器 |
| `(*CommandLoader).Scan() ([]string, error)` | 新增 | 扫描 commands/ 目录 |
| `(*CommandLoader).Load(name) (*CustomCommand, error)` | 新增 | 加载单个 .md 命令文件 |
| `(*CommandLoader).ParseCommand(content) (*CustomCommand, error)` | 新增 | LLM 解析 markdown 为任务列表 |

**层级: 功能组装层 → internal/tui/tui.go**
| 函数 | 操作 | 说明 |
|------|------|------|
| `(*App).setupUI()` | 修改 | 增加 commands 选择区 |
| `(*App).loadCustomCommands()` | 新增 | 加载自定义命令列表 |
| `(*App).runCustomCommand(name)` | 新增 | 执行自定义命令 |

### T7 命令文件格式规范 (commands/xxx.md)
```markdown
# 命令名
## 描述
简要描述命令用途

## 任务
[GET] http://localhost:8080/health
[POST] http://localhost:8080/backup
@ask: 是否继续?
@check: 验证结果
@wait: 5min
```

### T7 验收标准
- 放置 `commands/deploy.md` 后, TUI 中可选 "deploy" 命令
- 选中后加载并显示任务列表
- 支持执行/ask/check 交互

---

## T8. 开源项目审核与 README

### T8 数据模型
| 结构体 | 字段 | 说明 |
|--------|------|------|
| `OSSDependency` | Name, RepoURL, License, Version, Purpose, AuditStatus | 开源依赖 |

### T8 功能函数

**层级: 功能模块 → README.md**

待罗列的开源依赖:
| 项目 | 仓库 | 许可证 | 用途 |
|------|------|--------|------|
| onnxruntime-genai_purego | github.com/getcharzp/onnxruntime-genai_purego | - | ONNX 模型推理 |
| tview | github.com/rivo/tview | MIT | 终端 UI |
| tcell | github.com/gdamore/tcell | Apache-2.0 | 终端单元格 |
| purego | github.com/ebitengine/purego | Apache-2.0 | CGo-free FFI |
| golang.org/x/crypto | cs.opensource.google/go/x/crypto | BSD-3 | SSH 客户端 |
| google.golang.org/grpc | github.com/grpc/grpc-go | Apache-2.0 | gRPC 客户端 |
| google.golang.org/protobuf | github.com/protocolbuffers/protobuf-go | BSD-3 | protobuf |

### T8 验收标准
- README.md 包含完整开源依赖清单
- 仓库地址、用途、许可证明确
- plans/llm.md 审核状态已标记

---

## T9. 进度跟踪

记录到 `plans/process/t{1-9}-*.md`, 更新 `plans/process/t9-progress-track.md` 汇总表。

---

## 执行顺序

```
T1 (Model CLI 入口)
  ├── 功能入口: main.go model 路由
  ├── 功能模块: cli.go download 路由
  └── 数据模型: ModelCandidate
    
T2 (Mock Servers 入口)
  ├── 功能入口: http/grpc/ssh main.go
  ├── 功能函数: handler*/mock service
  └── 数据模型: 复用 types

T3 (Executor 组装层)
  ├── 功能组装: ExecuteTasks → executeTask 分支
  ├── 功能模块: retry.go, timeout.go
  ├── 功能函数: executeSSH, executeGRPC, classifyError
  └── 数据模型: RetryConfig, TimeoutConfig, ErrorCategory
     ↑ 依赖 T2 (Mock Server)
     ↑ 依赖 golang.org/x/crypto, google.golang.org/grpc

T4 (LLM Integration 组装层)
  ├── 功能模块: ask_executor.go, check_executor.go
  ├── 功能组装: executor + TUI 注入 LLM
  └── 数据模型: AskResult, CheckResult
     ↑ 依赖 T1 (模型就绪)

T5 (LLM Testing 入口)
  ├── 功能入口: main.go runTestMode 扩展
  ├── 功能测试: llm_inference_test, scenario_migration_test
  └── 场景链路: parse → analyze → execute → verify
     ↑ 依赖 T1 + T4

T6 (Full Testing 模块)
  ├── 单元测试: parser_test, executor_test 补充
  ├── 分类测试: 通过/非通过/模糊
  ├── 非代码测试: manual_test.md
  ├── 性能测试: bench_test.go
  └── 归档: cmd/xsh/onnx_test.go → tests/
     ↑ 依赖 T3 + T5

T7 (TUI Commands 模块)
  ├── 功能模块: command_loader.go
  ├── 功能组装: TUI commands 选择区
  └── 数据模型: CustomCommand, CommandLoader
     ↑ 依赖 T6

T8 (OSS Review)
  └── README.md + plans/llm.md 更新

T9 (Progress Track - 贯穿全程)
  └── plans/process/ 更新
```

T1/T2 并行 → T3 依赖 T2 → T4 依赖 T1 → T5 依赖 T1+T4 → T6 依赖 T3+T5 → T7 依赖 T6
