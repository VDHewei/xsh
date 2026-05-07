package llm

import (
	"fmt"
	"os"
	"regexp"
	"strings"

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
		// LLM 推理失败时降级到 mock
		fmt.Printf("LLM inference failed (%v), falling back to mock analysis\n", err)
		return a.mockAnalyze(content), nil
	}

	// 解析 LLM 输出为任务
	tasks := parseLLMResult(result)
	if len(tasks) == 0 {
		// LLM 输出无法解析时降级到 mock
		return a.mockAnalyze(content), nil
	}
	return tasks, nil
}

// AnalyzeFile 分析文件并生成任务
func (a *TaskAnalyzer) AnalyzeFile(filename string) ([]*types.Task, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}
	return a.AnalyzeContent(string(data))
}

// mockAnalyze 模拟分析（无模型时的后备方案）
func (a *TaskAnalyzer) mockAnalyze(content string) []*types.Task {
	var tasks []*types.Task

	// 检测迁移任务
	if strings.Contains(content, "migration") || strings.Contains(content, "迁移") {
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
	return fmt.Sprintf(`Analyze the following deployment/migration content and extract tasks.
Output ONLY the task list, one task per line, using this format:
[GET] <url> for GET requests
[POST] <url> for POST requests
@ask: <description> for user confirmations
@wait: <duration> for waiting periods (e.g. @wait: 10min)
@check: <description> for verification steps

Content:
%s`, content)
}

// parseLLMResult 解析 LLM 输出为任务
func parseLLMResult(result string) []*types.Task {
	var tasks []*types.Task
	lines := strings.Split(result, "\n")

	httpPattern := regexp.MustCompile(`^\[(\w+)\]\s+(.+)$`)
	cmdPattern := regexp.MustCompile(`^@(\w+):\s*(.+)$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ">") {
			continue
		}

		if match := httpPattern.FindStringSubmatch(line); match != nil {
			method := strings.ToUpper(match[1])
			url := strings.TrimSpace(match[2])
			tasks = append(tasks, &types.Task{
				Type: types.TaskTypeHTTP,
				Raw:  line,
				HTTP: &types.HTTPTask{
					Method: types.HTTPMethod(method),
					URL:    url,
				},
			})
			continue
		}

		if match := cmdPattern.FindStringSubmatch(line); match != nil {
			cmd := strings.ToLower(match[1])
			desc := strings.TrimSpace(match[2])
			switch cmd {
			case "ask":
				tasks = append(tasks, &types.Task{
					Type: types.TaskTypeAsk,
					Raw:  line,
					Ask:  &types.AskTask{Prompt: desc},
				})
			case "wait":
				tasks = append(tasks, &types.Task{
					Type: types.TaskTypeWait,
					Raw:  line,
					Wait: &types.WaitTask{Duration: desc},
				})
			case "check":
				tasks = append(tasks, &types.Task{
					Type: types.TaskTypeCheck,
					Raw:  line,
					Check: &types.CheckTask{Prompt: desc},
				})
			}
		}
	}

	return tasks
}

// extractURLs 提取 URLs
func extractURLs(content string) []string {
	var urls []string
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "http://") || strings.Contains(line, "https://") {
			// 提取 http(s):// 开头的 URL
			re := regexp.MustCompile(`https?://\S+`)
			found := re.FindAllString(line, -1)
			for _, url := range found {
				// 清除尾部标点
				url = strings.TrimRight(url, ".,;:!?)]}")
				urls = append(urls, url)
			}
		}
	}
	return urls
}

// InferWithPrompt 使用 LLM 推理（无模型时使用 mock）
func InferWithPrompt(prompt string) (string, error) {
	return MockInfer(prompt), nil
}
