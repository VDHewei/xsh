# T4: @ask/@check LLM 集成 - 启动提示词

## 当前状态
- T1 (Model CLI) 已完成: 模型加载/下载/列表/选择 全部就绪
- T2 (Mock Servers) 已完成: HTTP/gRPC/SSH mock 服务器
- T3 (Executor) 已完成: HTTP/SSH/gRPC/Wait 执行 + 重试/超时/错误分类
- T4 前置依赖已满足 (T1 模型就绪)

## 任务目标
将 `@ask` 和 `@check` 任务类型接入 LLM 本地推理，替换现有的简单文本提示。

## 数据模型 (需新增)

在 `internal/types/types.go` 中新增:
```go
// AskResult Ask 结果
type AskResult struct {
    Prompt     string // 原始提示
    Response   string // LLM 回复
    Suggestion string // 提取的建议
}

// CheckResult Check 结果
type CheckResult struct {
    Prompt  string // 原始提示
    Passed  bool   // 是否通过
    Reason  string // 判断理由
    Context string // 上下文信息
}
```

## 需要新增的文件

### 1. `pkg/llm/ask_executor.go`
```go
type AskExecutor struct {
    model *Model
}

func NewAskExecutor(model *Model) *AskExecutor
func (e *AskExecutor) Execute(task *types.AskTask) (*types.AskResult, error)
func (e *AskExecutor) buildAskPrompt(task *types.AskTask) string
```

**关键逻辑**: 
- `buildAskPrompt` 构造提示语模板, 引导 LLM 分析问题并提供建议
- `Execute` 调用 `model.Infer(prompt)` 并解析 LLM 返回结果为结构化 `AskResult`

### 2. `pkg/llm/check_executor.go`
```go
type CheckExecutor struct {
    model *Model
}

func NewCheckExecutor(model *Model) *CheckExecutor
func (e *CheckExecutor) Execute(task *types.CheckTask, context string) (*types.CheckResult, error)
func (e *CheckExecutor) buildCheckPrompt(task *types.CheckTask, ctx string) string
```

**关键逻辑**:
- `buildCheckPrompt` 结合上下文 (前面执行结果) 构造验证提示
- `Execute` 让 LLM 判断是否通过并返回 `CheckResult{Passed, Reason}`

## 需要修改的文件

### 3. `internal/executor/executor.go`
- `executeTask()` 中 Ask/Check 分支改为调用真实 LLM (目前返回占位文本)
- `SetLLMModel(model)` 已存在 (line 34-37), 确认 `llmModel` 类型需要改为 `*llm.Model`
- 目前 `llmModel` 是 `interface{}`, 需要在 T4 中改为具体类型

### 4. `internal/tui/tui.go`
- `executeTask()` (line 188-198): Ask 任务需显示 LLM 推理结果 + 用户确认对话框
- 新增 `showAskResultDialog(result *types.AskResult)`: 展示 LLM 建议, 用户选择继续/跳过
- 新增 `showCheckResultDialog(result *types.CheckResult)`: 展示检查结果 (通过/不通过)
- 需要注入 LLM Model: `(*App).SetLLMModel(model *llm.Model)`

## 依赖关系
- `pkg/llm/ask_executor.go` 依赖 `pkg/llm/model.go` (已有)
- `pkg/llm/check_executor.go` 依赖 `pkg/llm/model.go` (已有)
- `internal/executor/executor.go` 依赖 `pkg/llm` (已有导入)
- `internal/tui/tui.go` 需要新增 `pkg/llm` 导入

## 验收标准
1. `@ask: 需要执行数据库迁移吗?` → LLM 返回分析建议
2. `@check: 验证迁移后健康状态` → 结合上下文返回 Passed/Reason
3. TUI 中 Ask 显示 LLM 推理结果 + 用户确认/跳过对话框
4. TUI 中 Check 显示检查结果 (绿色通过/红色不通过)
5. Mock 模式: 无模型时优雅降级为简单文本提示

## 参考代码位置
- 模型加载: `pkg/llm/model.go:327-339` (Infer 方法)
- 现有 Ask/Check 占位: `internal/executor/executor.go:73-76`
- TUI Ask/Check: `internal/tui/tui.go:158-165`
- Task 类型: `internal/types/types.go:63-76`
- Executor LLM 注入: `internal/executor/executor.go:34-37`
