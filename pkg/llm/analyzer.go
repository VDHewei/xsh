package llm

import (
	"fmt"

	"github.com/VDHewei/xsh/internal/types"
)

// TaskAnalyzer 任务分析器
type TaskAnalyzer struct {
	model *Model
	cfg   *Config
}

// NewTaskAnalyzer 创建新任务分析器
func NewTaskAnalyzer() *TaskAnalyzer {
	return &TaskAnalyzer{
		model: nil,
		cfg:   NewConfig(),
	}
}

// SetModel 设置模型
func (a *TaskAnalyzer) SetModel(model *Model) {
	a.model = model
}

// AnalyzeContent 分析内容并生成任务
func (a *TaskAnalyzer) AnalyzeContent(content string) ([]*types.Task, error) {
	if a.model == nil || !a.model.IsLoaded() {
		// 使用 Mock 推理
		return a.mockAnalyze(content), nil
	}

	prompt := buildAnalyzePrompt(content)
	result, err := a.model.InferWithConfig(prompt, a.cfg)
	if err != nil {
		return nil, err
	}

	// 解析 LLM 输出为任务
	return a.parseTasks(result), nil
}

// AnalyzeFile 分析文件并生成任务
func (a *TaskAnalyzer) AnalyzeFile(filename string) ([]*types.Task, error) {
	content, err := readFileContent(filename)
	if err != nil {
		return nil, err
	}

	return a.AnalyzeContent(content)
}

// mockAnalyze 模拟分析
func (a *TaskAnalyzer) mockAnalyze(content string) []*types.Task {
	// 简单的 mock 实现：从内容中提取 URL
	var tasks []*types.Task

	// 检测迁移任务
	if contains(content, "migration") {
		task := &types.Task{
			Type: types.TaskTypeAsk,
			Raw:  "@ask: 检测到迁移任务，是否继续执行?",
			Ask: &types.AskTask{
				Prompt: "检测到迁移流程，是否继续执行?",
			},
		}
		tasks = append(tasks, task)
	}

	// 检测 HTTP 请求
	urls := extractURLs(content)
	for _, url := range urls {
		task := &types.Task{
			Type: types.TaskTypeHTTP,
			Raw:  fmt.Sprintf("[GET] %s", url),
			HTTP: &types.HTTPTask{
				Method: "GET",
				URL:   url,
			},
		}
		tasks = append(tasks, task)
	}

	return tasks
}

// buildAnalyzePrompt 构建分析提示
func buildAnalyzePrompt(content string) string {
	return fmt.Sprintf(`Please analyze the following deployment migration content and extract the tasks.

Content:
%s

Extract the tasks as a structured list in this format:
- [HTTP] <URL> for HTTP requests
- @ask: <description> for user interactions  
- @wait:<duration> for waiting periods
- @check: <description> for verification steps`, content)
}

// parseTasks 解析任务
func (a *TaskAnalyzer) parseTasks(result string) []*types.Task {
	// TODO: 使用更智能的解析
	// 这里使用简单的 mock 实现
	return a.mockAnalyze(result)
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(len(s) >= len(substr)) &&
		(s[:len(substr)] == substr ||
			(len(s) > len(substr) && findSubstring(s, substr)))
}

// findSubstring 查找子串
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// extractURLs 提取 URLs
func extractURLs(content string) []string {
	var urls []string
	var current []rune
	var inURL bool

	for _, ch := range content {
		if ch == 'h' && len(current) == 0 {
			current = append(current, ch)
			inURL = true
		} else if inURL {
			current = append(current, ch)
			if ch == ' ' || ch == '\n' || ch == '\t' {
				if len(current) > 4 && string(current[:4]) == "http" {
					urls = append(urls, string(current[:len(current)-1]))
				}
				current = nil
				inURL = false
			}
		}
	}

	if len(current) > 4 && string(current[:4]) == "http" {
		urls = append(urls, string(current))
	}

	return urls
}

// readFileContent 读取文件内容
func readFileContent(filename string) (string, error) {
	// 这个实现需要在外部调用 parser.ParseFile
	// 这里只做占位
	return "", fmt.Errorf("use parser.ParseFile instead")
}

// InferWithPrompt 使用 LLM 推理
func InferWithPrompt(prompt string) (string, error) {
	// 全局推理函数
	return MockInfer(prompt), nil
}