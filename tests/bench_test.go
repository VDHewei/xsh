package tests

import (
	"fmt"
	"testing"

	"github.com/VDHewei/xsh/internal/executor"
	"github.com/VDHewei/xsh/internal/types"
	llm "github.com/VDHewei/xsh/pkg/llm"
)

// ===== LLM Benchmarks =====

// BenchmarkMockInfer 测试 Mock 推理性能
func BenchmarkMockInfer(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		llm.MockInfer("Benchmark test prompt for performance measurement")
	}
}

// BenchmarkTaskAnalyzer 测试任务分析器性能
func BenchmarkTaskAnalyzer(b *testing.B) {
	analyzer := llm.NewTaskAnalyzer()
	content := `# Migration Steps
[GET] http://example.com/api/health
[POST] http://example.com/api/items
@wait: 5min
@ask: 检查数据库连接
[PUT] http://example.com/api/items/1
@check: 验证 API 响应
[DELETE] http://example.com/api/items/1
@wait: 10min`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tasks, err := analyzer.AnalyzeContent(content)
		if err != nil {
			b.Fatalf("AnalyzeContent failed: %v", err)
		}
		if len(tasks) == 0 {
			b.Error("no tasks returned")
		}
	}
}

// BenchmarkBuildTaskPrompt 测试构建提示词性能
func BenchmarkBuildTaskPrompt(b *testing.B) {
	content := "Sample migration content for benchmarking the prompt building function"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := llm.BuildTaskPrompt(content)
		if result == "" {
			b.Error("BuildTaskPrompt returned empty string")
		}
	}
}

// ===== HTTP Benchmarks =====

// BenchmarkExecutorHTTP 测试 HTTP 任务执行性能
// 需要 mock HTTP server 在 localhost:18080 运行
func BenchmarkExecutorHTTP(b *testing.B) {
	e := executor.NewExecutor()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		task := httpTask(types.GET, "http://localhost:18080/health", nil, "")
		result := e.ExecuteTasks([]*types.Task{task})[0]
		if !result.Success {
			b.Fatalf("HTTP health check failed: %v", result.Error)
		}
	}
}

// BenchmarkExecutorHTTPParallel 测试 HTTP 并发执行性能
func BenchmarkExecutorHTTPParallel(b *testing.B) {
	e := executor.NewExecutor()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			task := httpTask(types.GET, "http://localhost:18080/health", nil, "")
			result := e.ExecuteTasks([]*types.Task{task})[0]
			if !result.Success {
				b.Errorf("HTTP health check failed: %v", result.Error)
			}
		}
	})
}

// BenchmarkExecutorMultipleHTTP 测试批量 HTTP 任务执行性能
func BenchmarkExecutorMultipleHTTP(b *testing.B) {
	e := executor.NewExecutor()
	tasks := make([]*types.Task, 5)
	endpoints := []string{"/health", "/get", "/echo?q=1", "/echo?q=2", "/echo?q=3"}
	for i, ep := range endpoints {
		method := types.GET
		if i == 3 {
			method = types.POST
		}
		tasks[i] = httpTask(method, "http://localhost:18080"+ep, nil, `{"bench":"true"}`)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results := e.ExecuteTasks(tasks)
		for j, r := range results {
			if !r.Success {
				b.Fatalf("task[%d] failed: %v", j, r.Error)
			}
		}
	}
}

// ===== Wait/Ask/Check Benchmarks =====

// BenchmarkExecuteWait 测试 Wait 任务性能
func BenchmarkExecuteWait(b *testing.B) {
	e := executor.NewExecutor()
	task := &types.Task{
		Type: types.TaskTypeWait,
		Raw:  "@wait: 10min",
		Wait: &types.WaitTask{Duration: "10min"},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := e.ExecuteTasks([]*types.Task{task})[0]
		if !result.Success {
			b.Fatalf("wait task failed: %v", result.Error)
		}
	}
}

// BenchmarkExecuteAsk 测试 Ask 任务性能
func BenchmarkExecuteAsk(b *testing.B) {
	e := executor.NewExecutor()
	task := &types.Task{
		Type: types.TaskTypeAsk,
		Raw:  "@ask: Check service status?",
		Ask:  &types.AskTask{Prompt: "Check service status?"},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := e.ExecuteTasks([]*types.Task{task})[0]
		if !result.Success {
			b.Fatalf("ask task failed: %v", result.Error)
		}
	}
}

// BenchmarkExecuteCheck 测试 Check 任务性能
func BenchmarkExecuteCheck(b *testing.B) {
	e := executor.NewExecutor()
	task := &types.Task{
		Type:  types.TaskTypeCheck,
		Raw:   "@check: Verify deployment",
		Check: &types.CheckTask{Prompt: "Verify deployment"},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := e.ExecuteTasks([]*types.Task{task})[0]
		if !result.Success {
			b.Fatalf("check task failed: %v", result.Error)
		}
	}
}

// ===== RetryConfig Benchmarks =====

// BenchmarkRetry_Success 测试重试成功性能
func BenchmarkRetry_Success(b *testing.B) {
	cfg := executor.NewRetryConfig(3, 0, 1.0)
	callCount := 0
	fn := func() (string, error) {
		callCount++
		if callCount < 3 {
			return "", &benchError{msg: "temp error"}
		}
		return "success", nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		callCount = 0
		result, err := cfg.Do(fn)
		if err != nil {
			b.Fatalf("retry should succeed: %v", err)
		}
		_ = result
	}
}

// BenchmarkRetry_AllFail 测试重试全部失败性能
func BenchmarkRetry_AllFail(b *testing.B) {
	cfg := executor.NewRetryConfig(3, 0, 1.0)
	fn := func() (string, error) {
		return "", &benchError{msg: "always fail"}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cfg.Do(fn)
		if err == nil {
			b.Fatal("retry should fail")
		}
	}
}

// ===== Helpers =====

func httpTask(method types.HTTPMethod, url string, headers map[string]string, body string) *types.Task {
	taskType := types.TaskTypeHTTP
	return &types.Task{
		Type: taskType,
		Raw:  fmt.Sprintf("[%s] %s", method, url),
		HTTP: &types.HTTPTask{
			Method:  method,
			URL:     url,
			Headers: headers,
			Body:    body,
		},
	}
}

type benchError struct{ msg string }

func (e *benchError) Error() string { return e.msg }
