package types

// TaskType 任务类型
type TaskType string

const (
	TaskTypeHTTP  TaskType = "http"
	TaskTypeSSH   TaskType = "ssh"
	TaskTypeGRPC  TaskType = "grpc"
	TaskTypeAsk   TaskType = "ask"
	TaskTypeWait  TaskType = "wait"
	TaskTypeCheck TaskType = "check"
)

// HTTPMethod HTTP 方法
type HTTPMethod string

const (
	GET    HTTPMethod = "GET"
	POST   HTTPMethod = "POST"
	PUT    HTTPMethod = "PUT"
	DELETE HTTPMethod = "DELETE"
	PATCH  HTTPMethod = "PATCH"
)

// Task 单个任务
type Task struct {
	Type     TaskType
	Raw      string
	HTTP     *HTTPTask
	SSH      *SSHTask
	GRPC     *GRPCTask
	Ask      *AskTask
	Wait     *WaitTask
	Check    *CheckTask
}

// HTTPTask HTTP 任务
type HTTPTask struct {
	Method  HTTPMethod
	URL     string
	Headers map[string]string
	Body    string
}

// SSHTask SSH 任务
type SSHTask struct {
	Host    string
	Port    string
	User    string
	Command string
}

// GRPCTask gRPC 任务
type GRPCTask struct {
	Host    string
	Port    string
	Method  string
	Headers map[string]string
	Body    string
}

// AskTask 询问任务
type AskTask struct {
	Prompt string
}

// WaitTask 等待任务
type WaitTask struct {
	Duration string // 如 "10min", "30s"
}

// CheckTask 检查任务
type CheckTask struct {
	Prompt string
}

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

// TaskResult 任务执行结果
type TaskResult struct {
	Task    *Task
	Success bool
	Output  string
	Error   error
}

// CustomCommand 自定义命令 (从 commands/ 目录加载)
type CustomCommand struct {
	Name    string  // 命令名 (文件名, 不含 .md)
	File    string  // 文件路径
	Content string  // 原始文件内容
	Desc    string  // 命令描述 (从 ## 描述 提取)
	Tasks   []*Task // 解析后的任务列表
}

// CommandLoader 命令加载器接口
type CommandLoader interface {
	Scan() ([]string, error)                // 扫描 commands/ 目录, 返回命令名称列表
	Load(name string) (*CustomCommand, error) // 加载命令文件并解析
}

// ModelCandidate 候选模型配置
type ModelCandidate struct {
	RepoID  string // HuggingFace 仓库 ID (如 yasserrmd/glm5.1-distill-onnx)
	Name    string // 短名称 (如 deepseek, glm5.1)
	Default bool   // 是否为默认模型
}