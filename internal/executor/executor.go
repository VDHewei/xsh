package executor

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/VDHewei/xsh/internal/types"
)

// normalizeDuration 转换 duration 格式，如 "10min" -> "10m"
func normalizeDuration(d string) string {
	re := regexp.MustCompile(`(\d+)\s*min`)
	return re.ReplaceAllString(d, "${1}m")
}

// ExecuteTasks 执行所有任务
func ExecuteTasks(tasks []*types.Task) []string {
	var results []string
	for _, task := range tasks {
		result := executeTask(task)
		results = append(results, result)
	}
	return results
}

func executeTask(task *types.Task) string {
	switch task.Type {
	case types.TaskTypeHTTP:
		return executeHTTP(task.HTTP)
	case types.TaskTypeWait:
		return executeWait(task.Wait)
	case types.TaskTypeAsk:
		return fmt.Sprintf("%s - 需要用户确认", task.Raw)
	case types.TaskTypeCheck:
		return fmt.Sprintf("%s - 需要用户确认", task.Raw)
	default:
		return fmt.Sprintf("[SKIP] unknown task type: %s", task.Type)
	}
}

func executeHTTP(httpTask *types.HTTPTask) string {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest(string(httpTask.Method), httpTask.URL, nil)
	if err != nil {
		return fmt.Sprintf("[ERROR] %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf("[ERROR] %v", err)
	}
	defer resp.Body.Close()

	return fmt.Sprintf("[%s] %s - Status: %d", httpTask.Method, httpTask.URL, resp.StatusCode)
}

func executeWait(waitTask *types.WaitTask) string {
	normalized := normalizeDuration(waitTask.Duration)
	_, err := time.ParseDuration(normalized)
	if err != nil {
		return fmt.Sprintf("[ERROR] invalid duration: %v", err)
	}
	// CLI 模式下跳过实际等待，只输出信息
	return fmt.Sprintf("[WAIT] %s (skipped in CLI mode)", waitTask.Duration)
}