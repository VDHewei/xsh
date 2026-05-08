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

// TaskResult 任务执行结果
type TaskResult struct {
	Task    *Task
	Success bool
	Output  string
	Error   error
}