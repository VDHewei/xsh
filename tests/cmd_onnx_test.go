package tests

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/VDHewei/xsh/internal/parser"
	llm "github.com/VDHewei/xsh/pkg/llm"
)

// TaskStep 任务步骤
type TaskStep struct {
	ID     string `json:"id"`
	Action string `json:"action"`
	Tool   string `json:"tool"`
	Params string `json:"params"`
	Status string `json:"status"`
}

// TaskResult 任务结果
type TaskResult struct {
	Plan    string       `json:"plan"`
	Steps   []TaskStep   `json:"steps"`
	Results []StepResult `json:"results"`
}

// StepResult 步骤结果
type StepResult struct {
	StepID string `json:"step_id"`
	Output string `json:"output"`
	Error  string `json:"error"`
}

// TestMockAnalyze 测试 Mock 分析
func TestMockAnalyze(t *testing.T) {
	analyzer := llm.NewTaskAnalyzer()
	content := `# Migration Steps
[GET] http://example.com/api/health
@wait: 10min
@ask: 检查服务是否正常
[POST] http://example.com/api/deploy
@check: 验证部署结果`

	tasks, err := analyzer.AnalyzeContent(content)
	if err != nil {
		t.Fatalf("AnalyzeContent failed: %v", err)
	}

	if len(tasks) == 0 {
		t.Error("AnalyzeContent should return at least one task")
	}

	for i, task := range tasks {
		t.Logf("Task[%d]: type=%s raw=%s", i, task.Type, task.Raw)
	}
}

// TestModelCreation 测试模型创建
func TestModelCreation(t *testing.T) {
	model := llm.NewModel("test-model")
	if model.IsLoaded() {
		t.Error("New model should not be loaded")
	}
	if model.Name != "test-model" {
		t.Errorf("Expected name 'test-model', got '%s'", model.Name)
	}
}

// TestNewConfig 测试配置创建
func TestNewConfig(t *testing.T) {
	cfg := llm.NewConfig()
	if cfg.MaxLength != 2048 {
		t.Errorf("Expected MaxLength 2048, got %d", cfg.MaxLength)
	}
	if cfg.Temperature != 0.7 {
		t.Errorf("Expected Temperature 0.7, got %f", cfg.Temperature)
	}
}

// TestTaskAnalyzerWithFile 测试 TaskAnalyzer 文件分析
func TestTaskAnalyzerWithFile(t *testing.T) {
	analyzer := llm.NewTaskAnalyzer()
	testFile := "tests/data/prod-migration-form-uat.txt"

	tasks, err := analyzer.AnalyzeFile(testFile)
	if err != nil {
		t.Skipf("Test file not found: %s, skipping", testFile)
	}

	t.Logf("Analyzed %d tasks from file", len(tasks))
	for i, task := range tasks {
		t.Logf("  [%d] %s: %s", i, task.Type, task.Raw)
	}
}

// TestListModels 测试列出模型
func TestListModels(t *testing.T) {
	models, err := llm.ListModels("models")
	if err != nil {
		t.Logf("ListModels returned error (expected if no models dir): %v", err)
		return
	}
	t.Logf("Found %d models", len(models))
	for _, m := range models {
		t.Logf("  - %s", m)
	}
}

// TestSearchModels 测试搜索模型
func TestSearchModels(t *testing.T) {
	cfg := llm.NewDownloadConfig()
	models, err := llm.SearchModels("onnx", cfg)
	if err != nil {
		t.Logf("SearchModels returned error: %v", err)
		return
	}
	if len(models) == 0 {
		t.Error("SearchModels should return at least one result")
	}
	t.Logf("Search results: %v", models)
}

// TestGetModelInfo 测试获取模型信息
func TestGetModelInfo(t *testing.T) {
	cfg := llm.NewDownloadConfig()
	info, err := llm.GetModelInfo("onnxruntime/DeepSeek-R1-Distill-ONNX", cfg)
	if err != nil {
		t.Fatalf("GetModelInfo failed: %v", err)
	}
	t.Logf("Model info: %v", info)
}

// TestDownloadConfig 测试下载配置
func TestDownloadConfig(t *testing.T) {
	cfg := llm.NewDownloadConfig()
	if cfg.CacheDir != "models" {
		t.Errorf("Expected default CacheDir 'models', got '%s'", cfg.CacheDir)
	}

	llm.SetProxy("http://proxy:8080")
	proxy := os.Getenv("HTTP_PROXY")
	if proxy != "http://proxy:8080" {
		t.Errorf("Expected HTTP_PROXY 'http://proxy:8080', got '%s'", proxy)
	}
	os.Unsetenv("HTTP_PROXY")
	os.Unsetenv("HTTPS_PROXY")
}

// TestGetModelPath 测试获取模型路径
func TestGetModelPath(t *testing.T) {
	path := llm.GetModelPath("models", "test-model")
	if path != "models/test-model" && path != "models\\test-model" {
		t.Errorf("Expected path 'models/test-model' or 'models\\test-model', got '%s'", path)
	}
}

// TestExecuteTaskResult 测试执行任务结果解析
func TestExecuteTaskResult(t *testing.T) {
	jsonResult := `{
		"plan": "Execute migration steps",
		"steps": [
			{"id": "step-1", "action": "健康检查", "tool": "http-get", "params": "http://example.com/health", "status": "pending"},
			{"id": "step-2", "action": "等待", "tool": "sleep", "params": "10min", "status": "pending"}
		]
	}`

	var taskResult TaskResult
	if err := json.Unmarshal([]byte(jsonResult), &taskResult); err != nil {
		t.Fatalf("Failed to parse task result JSON: %v", err)
	}

	if taskResult.Plan != "Execute migration steps" {
		t.Errorf("Expected plan 'Execute migration steps', got '%s'", taskResult.Plan)
	}
	if len(taskResult.Steps) != 2 {
		t.Errorf("Expected 2 steps, got %d", len(taskResult.Steps))
	}
}

// TestModelLoadNonexistent 测试加载不存在的模型
func TestModelLoadNonexistent(t *testing.T) {
	model := llm.NewModel("nonexistent")
	err := model.Load("/nonexistent/path/model_dir")
	if err == nil {
		t.Error("Loading nonexistent model should return error")
	}
	t.Logf("Expected error: %v", err)
}

// TestModelUnload 测试模型卸载
func TestModelUnload(t *testing.T) {
	model := llm.NewModel("test-unload")
	model.Unload()
	if model.IsLoaded() {
		t.Error("After unload, model should not be loaded")
	}
}

// TestParserWithSampleData 测试解析器配合示例数据
func TestParserWithSampleData(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := tmpDir + "/test_tasks.txt"

	content := `# Test migration file
> This is a comment
[GET] http://example.com/api/health
@wait: 10min
@ask: 检查服务是否正常
[POST] http://example.com/api/deploy
@check: 验证部署结果
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tasks, err := parser.ParseFile(testFile)
	if err != nil {
		t.Fatalf("parseFile failed: %v", err)
	}

	if len(tasks) == 0 {
		t.Error("Should parse at least one task")
	}

	t.Logf("Parsed %d tasks", len(tasks))
	for i, task := range tasks {
		t.Logf("  [%d] Type=%s Raw=%s", i, task.Type, task.Raw)
	}
}

// TestDefaultGenAILibraryPath 测试默认 GenAI 动态库路径
func TestDefaultGenAILibraryPath(t *testing.T) {
	path := llm.DefaultGenAILibraryPath()
	if path == "" {
		t.Error("DefaultGenAILibraryPath should return non-empty string")
	}
	t.Logf("GenAI lib path: %s", path)
}

// TestDefaultOnnxRuntimeLibraryPath 测试默认 onnxruntime 动态库路径
func TestDefaultOnnxRuntimeLibraryPath(t *testing.T) {
	path := llm.DefaultOnnxRuntimeLibraryPath()
	if path == "" {
		t.Error("DefaultOnnxRuntimeLibraryPath should return non-empty string")
	}
	t.Logf("OnnxRuntime lib path: %s", path)
}

// TestBuildTaskPrompt 测试构建任务提示
func TestBuildTaskPrompt(t *testing.T) {
	prompt := llm.BuildTaskPrompt("Sample migration content")
	if prompt == "" {
		t.Error("BuildTaskPrompt should return non-empty string")
	}
	t.Logf("Prompt length: %d", len(prompt))
}
