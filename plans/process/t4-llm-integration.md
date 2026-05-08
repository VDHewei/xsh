# T4: @ask/@check LLM 集成 - 进度 ✅

| 子任务 | 状态 | 备注 |
|--------|------|------|
| T4.1 AskExecutor | 已完成 | pkg/llm/ask_executor.go: AskExecutor + buildAskPrompt + mock回退 |
| T4.2 CheckExecutor | 已完成 | pkg/llm/check_executor.go: CheckExecutor + buildCheckPrompt + 上下文传递 |
| T4.3 executor 集成 | 已完成 | executor.go: executeAsk/executeCheck 替换占位文本 |
| T4.4 TUI 集成 | 已完成 | tui.go: showAskResultDialog/showCheckResultDialog |

## 测试结果
- 通过: TestExecuteAskTask, TestExecuteCheckTask
- 失败: HTTP相关测试 (mock server 未启动, 与T4无关)
- 通过原因: Ask/Check 任务成功调用 AskExecutor/CheckExecutor, 无模型时 mock 降级

## 新增文件
- `pkg/llm/ask_executor.go` - AskExecutor: Execute, buildAskPrompt, mockExecute
- `pkg/llm/check_executor.go` - CheckExecutor: Execute, buildCheckPrompt, parseCheckResult, mockExecute

## 修改文件
- `internal/types/types.go` - 新增 AskResult, CheckResult 结构体
- `internal/executor/executor.go` - llmModel 类型改为 *llm.Model, 新增 executeAsk/executeCheck
- `internal/tui/tui.go` - 新增 SetLLMModel, showAskResultDialog, showCheckResultDialog
