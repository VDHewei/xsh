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
		return nil, fmt.Errorf("no LLM model loaded for task analysis")
	}

	prompt := buildAnalyzePrompt(content)
	result, err := a.model.InferWithConfig(prompt, a.cfg)
	if err != nil {
		return nil, fmt.Errorf("LLM inference failed: %w", err)
	}

	// 解析 LLM 输出为任务
	tasks := parseLLMResult(result)
	if len(tasks) == 0 {
		return nil, fmt.Errorf("LLM output could not be parsed into tasks")
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

// buildAnalyzePrompt 构建分析提示
func buildAnalyzePrompt(content string) string {
	return fmt.Sprintf(`Analyze the following deployment/migration content and extract tasks.
Output ONLY the task list, one task per line, using this format:
[GET] <url> header:Key=Value for GET requests with optional headers
[POST] <url> header:Key=Value body:{"key":"val"} for POST requests with optional headers and body
[PUT] <url> header:Key=Value body:{"key":"val"} for PUT requests with optional headers and body
[DELETE] <url> header:Key=Value for DELETE requests with optional headers
[GRPC] <host:port/Method> header:Key=Value body:{"key":"val"} for gRPC calls with optional headers and body
@ask: <description> for user confirmations
@wait: <duration> for waiting periods (e.g. @wait: 10min)
@check: <description> for verification steps

Headers can be repeated: header:Content-Type=application/json header:Authorization=Bearer token123
Body is JSON and must be prefixed with body:

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
			rest := strings.TrimSpace(match[2])

			headers, body, target := parseHeadersAndBody(rest)

			switch method {
			case "GRPC":
				host, port, grpcMethod := parseGRPCTarget(target)
				tasks = append(tasks, &types.Task{
					Type: types.TaskTypeGRPC,
					Raw:  line,
					GRPC: &types.GRPCTask{
						Host:    host,
						Port:    port,
						Method:  grpcMethod,
						Headers: headers,
						Body:    body,
					},
				})
			default:
				tasks = append(tasks, &types.Task{
					Type: types.TaskTypeHTTP,
					Raw:  line,
					HTTP: &types.HTTPTask{
						Method:  types.HTTPMethod(method),
						URL:     target,
						Headers: headers,
						Body:    body,
					},
				})
			}
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

// parseHeadersAndBody 从行剩余部分提取 headers 和 body，返回 headers, body, target(URL或gRPC地址)
func parseHeadersAndBody(rest string) (headers map[string]string, body string, target string) {
	headers = make(map[string]string)

	// 找到第一个 header: 或 body: 的位置来切割 target
	headerIdx := strings.Index(rest, " header:")
	bodyIdx := strings.Index(rest, " body:")

	cutIdx := len(rest)
	if headerIdx >= 0 && headerIdx < cutIdx {
		cutIdx = headerIdx
	}
	if bodyIdx >= 0 && bodyIdx < cutIdx {
		cutIdx = bodyIdx
	}

	target = strings.TrimSpace(rest[:cutIdx])
	suffix := rest[cutIdx:]

	// 解析 header:Key=Value
	headerRe := regexp.MustCompile(`header:(\S+?)=(\S+)`)
	for _, m := range headerRe.FindAllStringSubmatch(suffix, -1) {
		headers[m[1]] = m[2]
	}

	// 解析 body:{...}
	bodyRe := regexp.MustCompile(`body:(\{[^}]*\})`)
	if m := bodyRe.FindStringSubmatch(suffix); m != nil {
		body = m[1]
	}

	return headers, body, target
}

// parseGRPCTarget 解析 gRPC 目标地址 host:port/Method
func parseGRPCTarget(target string) (host, port, method string) {
	// 格式: host:port/Service/Method
	slashIdx := strings.Index(target, "/")
	if slashIdx < 0 {
		return target, "", ""
	}
	addr := target[:slashIdx]
	method = target[slashIdx+1:]

	colonIdx := strings.LastIndex(addr, ":")
	if colonIdx < 0 {
		return addr, "", method
	}
	return addr[:colonIdx], addr[colonIdx+1:], method
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

