package executor

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/VDHewei/xsh/internal/types"
	"golang.org/x/crypto/ssh"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/structpb"
)

// Executor 任务执行器
type Executor struct {
	retryCfg *RetryConfig
	sshCfg   *ssh.ClientConfig
	llmModel interface{} // LLM 模型引用, T4 阶段注入
}

// NewExecutor 创建执行器
func NewExecutor() *Executor {
	return &Executor{
		retryCfg: DefaultRetryConfig(),
	}
}

// SetLLMModel 注入 LLM 模型 (T4 阶段使用)
func (e *Executor) SetLLMModel(model interface{}) {
	e.llmModel = model
}

// ExecuteTasks 执行所有任务, 返回结果列表
func (e *Executor) ExecuteTasks(tasks []*types.Task) []types.TaskResult {
	var results []types.TaskResult
	for _, task := range tasks {
		result := e.executeTask(task)
		results = append(results, result)
	}
	return results
}

// ExecuteTasksSimple 执行任务, 返回字符串结果 (兼容旧接口)
func ExecuteTasks(tasks []*types.Task) []string {
	e := NewExecutor()
	var results []string
	for _, task := range tasks {
		r := e.executeTask(task)
		results = append(results, formatResult(r))
	}
	return results
}

func (e *Executor) executeTask(task *types.Task) types.TaskResult {
	var err error
	var output string

	switch task.Type {
	case types.TaskTypeHTTP:
		output, err = e.executeHTTP(task.HTTP)
	case types.TaskTypeSSH:
		output, err = e.executeSSH(task.SSH)
	case types.TaskTypeGRPC:
		output, err = e.executeGRPC(task.GRPC)
	case types.TaskTypeWait:
		output = e.executeWait(task.Wait)
	case types.TaskTypeAsk:
		output = fmt.Sprintf("@ask: %s - 需要LLM分析", task.Ask.Prompt)
	case types.TaskTypeCheck:
		output = fmt.Sprintf("@check: %s - 需要LLM分析", task.Check.Prompt)
	default:
		return types.TaskResult{
			Task:    task,
			Success: false,
			Output:  fmt.Sprintf("unknown task type: %s", task.Type),
			Error:   fmt.Errorf("unknown task type: %s", task.Type),
		}
	}

	if err != nil {
		return types.TaskResult{
			Task:    task,
			Success: false,
			Output:  output + " | " + err.Error(),
			Error:   err,
		}
	}

	return types.TaskResult{
		Task:    task,
		Success: true,
		Output:  output,
	}
}

// --- HTTP Execution ---

func (e *Executor) executeHTTP(httpTask *types.HTTPTask) (string, error) {
	method := strings.ToUpper(string(httpTask.Method))
	url := httpTask.URL

	fn := func() (string, error) {
		var bodyReader io.Reader
		if httpTask.Body != "" {
			bodyReader = strings.NewReader(httpTask.Body)
		}

		req, err := http.NewRequest(method, url, bodyReader)
		if err != nil {
			return "", fmt.Errorf("create request: %w", err)
		}

		// 设置 Headers
		for k, v := range httpTask.Headers {
			req.Header.Set(k, v)
		}
		if httpTask.Body != "" && req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", "application/json")
		}

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return "", classifyHTTPError(err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(io.LimitReader(resp.Body, 10240))
		bodyStr := string(body)
		if len(bodyStr) > 500 {
			bodyStr = bodyStr[:500] + "..."
		}

		if resp.StatusCode >= 400 {
			return fmt.Sprintf("[%s] %s - Status: %d, Body: %s", method, url, resp.StatusCode, bodyStr),
				fmt.Errorf("HTTP %d: %s", resp.StatusCode, bodyStr)
		}

		return fmt.Sprintf("[%s] %s - Status: %d, Body: %s", method, url, resp.StatusCode, bodyStr), nil
	}

	return e.retryCfg.Do(fn)
}

// --- SSH Execution ---

func (e *Executor) executeSSH(sshTask *types.SSHTask) (string, error) {
	if e.sshCfg == nil {
		e.sshCfg = &ssh.ClientConfig{
			User:            sshTask.User,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         10 * time.Second,
		}
	} else {
		e.sshCfg.User = sshTask.User
	}

	fn := func() (string, error) {
		addr := net.JoinHostPort(sshTask.Host, sshTask.Port)
		if sshTask.Port == "" {
			addr = net.JoinHostPort(sshTask.Host, "22")
		}

		client, err := ssh.Dial("tcp", addr, e.sshCfg)
		if err != nil {
			return "", fmt.Errorf("SSH dial: %w", err)
		}
		defer client.Close()

		session, err := client.NewSession()
		if err != nil {
			return "", fmt.Errorf("SSH session: %w", err)
		}
		defer session.Close()

		output, err := session.CombinedOutput(sshTask.Command)
		if err != nil {
			return fmt.Sprintf("[SSH] %s@%s - %s\n%s", sshTask.User, addr, sshTask.Command, string(output)),
				fmt.Errorf("SSH command: %w", err)
		}

		return fmt.Sprintf("[SSH] %s@%s - %s\n%s", sshTask.User, addr, sshTask.Command, string(output)), nil
	}

	return e.retryCfg.Do(fn)
}

// --- gRPC Execution ---

func (e *Executor) executeGRPC(grpcTask *types.GRPCTask) (string, error) {
	fn := func() (string, error) {
		addr := net.JoinHostPort(grpcTask.Host, grpcTask.Port)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		conn, err := grpc.DialContext(ctx, addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		)
		if err != nil {
			return "", fmt.Errorf("gRPC dial: %w", err)
		}
		defer conn.Close()

		// 使用动态调用方式 (反射风格)
		output, err := invokeGRPCMethod(ctx, conn, grpcTask.Method, grpcTask.Body, grpcTask.Headers)
		if err != nil {
			return fmt.Sprintf("[gRPC] %s - %s", addr, grpcTask.Method),
				fmt.Errorf("gRPC call: %w", err)
		}

		return fmt.Sprintf("[gRPC] %s - %s\n%s", addr, grpcTask.Method, output), nil
	}

	return e.retryCfg.Do(fn)
}

// invokeGRPCMethod 通过反射调用 gRPC 方法 (通用方式, 不依赖具体 proto)
func invokeGRPCMethod(ctx context.Context, conn *grpc.ClientConn, method, body string, headers map[string]string) (string, error) {
	// 使用 gRPC 的通用调用方式, 发送 JSON 并接收 JSON
	// 构造一个简单的 structpb 请求
	req := &structpb.Value{}
	if body != "" {
		req = structpb.NewStringValue(body)
	}

	resp := &structpb.Value{}
	methodPath := method
	if !strings.HasPrefix(methodPath, "/") {
		methodPath = "/" + methodPath
	}

	err := conn.Invoke(ctx, methodPath, req, resp)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%v", resp.AsInterface()), nil
}

// --- Wait Execution ---

func (e *Executor) executeWait(waitTask *types.WaitTask) string {
	normalized := normalizeDuration(waitTask.Duration)
	d, err := time.ParseDuration(normalized)
	if err != nil {
		return fmt.Sprintf("[ERROR] invalid duration: %s - %v", waitTask.Duration, err)
	}
	return fmt.Sprintf("[WAIT] %s (%v)", waitTask.Duration, d)
}

// --- Helpers ---

func normalizeDuration(d string) string {
	re := regexp.MustCompile(`(\d+)\s*min`)
	return re.ReplaceAllString(d, "${1}m")
}

func formatResult(r types.TaskResult) string {
	if r.Success {
		return r.Output
	}
	return fmt.Sprintf("[ERROR] %s: %v", r.Output, r.Error)
}

func classifyHTTPError(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()

	if strings.Contains(msg, "timeout") || strings.Contains(msg, "deadline exceeded") {
		return fmt.Errorf("HTTP_TIMEOUT: %w", err)
	}
	if strings.Contains(msg, "connection refused") || strings.Contains(msg, "actively refused") || strings.Contains(msg, "no such host") {
		return fmt.Errorf("HTTP_CONNECTION: %w", err)
	}
	if strings.Contains(msg, "DNS") || strings.Contains(msg, "name resolution") {
		return fmt.Errorf("HTTP_DNS: %w", err)
	}
	if strings.Contains(msg, "TLS") || strings.Contains(msg, "certificate") {
		return fmt.Errorf("HTTP_TLS: %w", err)
	}

	return fmt.Errorf("HTTP_ERROR: %w", err)
}
