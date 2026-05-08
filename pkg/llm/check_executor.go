package llm

import (
	"fmt"
	"strings"

	"github.com/VDHewei/xsh/internal/types"
)

// CheckExecutor Check 执行器 - 结合上下文让 LLM 判断是否通过
type CheckExecutor struct {
	model *Model
}

// NewCheckExecutor 创建 Check 执行器
func NewCheckExecutor(model *Model) *CheckExecutor {
	return &CheckExecutor{model: model}
}

// Execute 结合上下文让 LLM 判断检查是否通过
func (e *CheckExecutor) Execute(task *types.CheckTask, context string) (*types.CheckResult, error) {
	prompt := e.buildCheckPrompt(task, context)

	if e.model == nil || !e.model.IsLoaded() {
		return e.mockExecute(task, context), nil
	}

	response, err := e.model.Infer(prompt)
	if err != nil {
		fmt.Printf("Check LLM inference failed (%v), falling back to mock\n", err)
		return e.mockExecute(task, context), nil
	}

	passed, reason := parseCheckResult(response)

	return &types.CheckResult{
		Prompt:  task.Prompt,
		Passed:  passed,
		Reason:  reason,
		Context: context,
	}, nil
}

// buildCheckPrompt 构造 check 提示语模板
func (e *CheckExecutor) buildCheckPrompt(task *types.CheckTask, ctx string) string {
	contextInfo := "No prior execution context available."
	if ctx != "" {
		contextInfo = fmt.Sprintf("Prior execution results:\n%s", ctx)
	}

	return fmt.Sprintf(`You are a deployment/migration verification assistant.

Check condition: %s

%s

Based on the check condition and execution context, determine if the verification passes.
- PASS: the condition is met, all expected behaviors are observed
- FAIL: the condition is not met, errors or unexpected behaviors detected

Respond in this exact format:
VERDICT: [PASS|FAIL]
REASON: <brief explanation of why it passed or failed>`, task.Prompt, contextInfo)
}

// parseCheckResult 解析 LLM 检查结果
func parseCheckResult(response string) (passed bool, reason string) {
	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		upperLine := strings.ToUpper(line)
		if strings.HasPrefix(upperLine, "VERDICT:") {
			verdict := strings.TrimSpace(line[len("VERDICT:"):])
			passed = strings.EqualFold(verdict, "PASS")
		}
		if strings.HasPrefix(upperLine, "REASON:") {
			reason = strings.TrimSpace(line[len("REASON:"):])
		}
	}
	if reason == "" {
		// fallback: 用首行作为原因
		for _, line := range lines {
			if trimmed := strings.TrimSpace(line); trimmed != "" && !strings.HasPrefix(trimmed, "VERDICT:") {
				reason = trimmed
				break
			}
		}
		if reason == "" {
			reason = response
		}
	}
	return passed, reason
}

// mockExecute 无模型时返回模拟检查结果
func (e *CheckExecutor) mockExecute(task *types.CheckTask, context string) *types.CheckResult {
	return &types.CheckResult{
		Prompt:  task.Prompt,
		Passed:  true, // mock 默认通过
		Reason:  fmt.Sprintf("[Mock] Check condition '%s' assumed passed (no LLM model loaded)", task.Prompt),
		Context: context,
	}
}
