package llm

import (
	"fmt"
	"strings"

	"github.com/VDHewei/xsh/internal/types"
)

// AskExecutor Ask 执行器 - 将 @ask 任务发送给 LLM 获取建议
type AskExecutor struct {
	model *Model
}

// NewAskExecutor 创建 Ask 执行器
func NewAskExecutor(model *Model) *AskExecutor {
	return &AskExecutor{model: model}
}

// Execute 执行 Ask 任务: 发送 prompt 给 LLM, 解析建议
func (e *AskExecutor) Execute(task *types.AskTask) (*types.AskResult, error) {
	if e.model == nil || !e.model.IsLoaded() {
		return nil, fmt.Errorf("no LLM model loaded for ask execution")
	}

	prompt := e.buildAskPrompt(task)
	response, err := e.model.Infer(prompt)
	if err != nil {
		return nil, fmt.Errorf("ask LLM inference failed: %w", err)
	}

	suggestion := extractSuggestion(response)

	return &types.AskResult{
		Prompt:     task.Prompt,
		Response:   response,
		Suggestion: suggestion,
	}, nil
}

// buildAskPrompt 构造 ask 提示语模板
func (e *AskExecutor) buildAskPrompt(task *types.AskTask) string {
	return fmt.Sprintf(`You are a task execution assistant analyzing the following question.

Question: %s

Based on the question, determine what action should be taken. 
Consider the context: this is a deployment/migration/production environment.
Provide your analysis and a clear recommendation:
- If the action seems safe and routine, recommend "CONTINUE"
- If the action requires caution, recommend "REVIEW"
- If the action is dangerous, recommend "SKIP"

Respond in this format:
RECOMMENDATION: [CONTINUE|REVIEW|SKIP]
REASON: <brief explanation>
SUGGESTION: <detailed suggestion>`, task.Prompt)
}

// extractSuggestion 从 LLM 回复中提取建议
func extractSuggestion(response string) string {
	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		upperLine := strings.ToUpper(line)
		if strings.HasPrefix(upperLine, "SUGGESTION:") {
			return strings.TrimSpace(line[len("SUGGESTION:"):])
		}
	}
	// fallback: 返回最后一行非空内容
	for i := len(lines) - 1; i >= 0; i-- {
		if trimmed := strings.TrimSpace(lines[i]); trimmed != "" {
			return trimmed
		}
	}
	return response
}
