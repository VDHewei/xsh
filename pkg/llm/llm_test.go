package llm

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// projectRoot returns the project root directory
func projectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	// pkg/llm/llm_test.go -> project root
	return filepath.Join(filepath.Dir(filename), "..", "..")
}

func deepseekModelDir() string {
	return filepath.Join(projectRoot(), "models", "deepseek-r1-distill-qwen-1.5B", "cpu_and_mobile", "cpu-int4-rtn-block-32-acc-level-4")
}

func TestMockInfer(t *testing.T) {
	result := MockInfer("Test input")
	if result == "" {
		t.Error("MockInfer should return non-empty string")
	}
	t.Logf("MockInfer: %s", result)
}

func TestNewModel(t *testing.T) {
	model := NewModel("test")
	if model.Name != "test" {
		t.Errorf("Expected name 'test', got '%s'", model.Name)
	}
	if model.IsLoaded() {
		t.Error("New model should not be loaded")
	}
}

func TestModelUnload(t *testing.T) {
	model := NewModel("test")
	model.Unload()
	if model.IsLoaded() {
		t.Error("After unload, model should not be loaded")
	}
}

func TestModelLoadNonexistent(t *testing.T) {
	model := NewModel("test")
	err := model.Load("/nonexistent/model_dir")
	if err == nil {
		t.Error("Loading nonexistent model should return error")
	}
	t.Logf("Expected error: %v", err)
}

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()
	if cfg.MaxLength != 2048 {
		t.Errorf("Expected MaxLength 2048, got %d", cfg.MaxLength)
	}
	if cfg.Temperature != 0.7 {
		t.Errorf("Expected Temperature 0.7, got %f", cfg.Temperature)
	}
	if cfg.TopP != 0.9 {
		t.Errorf("Expected TopP 0.9, got %f", cfg.TopP)
	}
	if cfg.TopK != 40 {
		t.Errorf("Expected TopK 40, got %d", cfg.TopK)
	}
	if cfg.NumThreads != 4 {
		t.Errorf("Expected NumThreads 4, got %d", cfg.NumThreads)
	}
}

func TestNewTaskAnalyzer(t *testing.T) {
	analyzer := NewTaskAnalyzer()
	if analyzer == nil {
		t.Fatal("NewTaskAnalyzer should return non-nil analyzer")
	}
}

func TestTaskAnalyzerAnalyzeContent(t *testing.T) {
	analyzer := NewTaskAnalyzer()
	content := `[GET] http://example.com/api/health
@wait: 10min
@ask: 检查服务是否正常`

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

func TestTaskAnalyzerWithMigrationContent(t *testing.T) {
	analyzer := NewTaskAnalyzer()
	content := `# Migration Task
migration from prod to uat
http://example.com/api/health
http://example.com/api/deploy`

	tasks, err := analyzer.AnalyzeContent(content)
	if err != nil {
		t.Fatalf("AnalyzeContent failed: %v", err)
	}
	if len(tasks) == 0 {
		t.Error("Should detect at least one task in migration content")
	}
	for i, task := range tasks {
		t.Logf("Task[%d]: type=%s raw=%s", i, task.Type, task.Raw)
	}
}

func TestTaskAnalyzerAnalyzeFile(t *testing.T) {
	analyzer := NewTaskAnalyzer()

	// Test with nonexistent file
	_, err := analyzer.AnalyzeFile("/nonexistent/file.txt")
	if err == nil {
		t.Error("AnalyzeFile should return error for nonexistent file")
	}
}

func TestInferWithPrompt(t *testing.T) {
	result, err := InferWithPrompt("Test prompt")
	if err != nil {
		t.Fatalf("InferWithPrompt failed: %v", err)
	}
	if result == "" {
		t.Error("InferWithPrompt should return non-empty string")
	}
}

func TestNewDownloadConfig(t *testing.T) {
	cfg := NewDownloadConfig()
	if cfg.CacheDir != "models" {
		t.Errorf("Expected CacheDir 'models', got '%s'", cfg.CacheDir)
	}
	if cfg.Proxy != "" {
		t.Errorf("Expected empty Proxy, got '%s'", cfg.Proxy)
	}
	if cfg.Mirror != "" {
		t.Errorf("Expected empty Mirror, got '%s'", cfg.Mirror)
	}
	if cfg.Token != "" {
		t.Errorf("Expected empty Token, got '%s'", cfg.Token)
	}
}

func TestListModelsEmpty(t *testing.T) {
	models, err := ListModels(t.TempDir())
	if err != nil {
		t.Fatalf("ListModels failed: %v", err)
	}
	if len(models) != 0 {
		t.Errorf("Expected 0 models in empty dir, got %d", len(models))
	}
}

func TestListModelsNonexistentDir(t *testing.T) {
	models, err := ListModels("/nonexistent/dir")
	if err != nil {
		t.Fatalf("ListModels should not error for nonexistent dir: %v", err)
	}
	if len(models) != 0 {
		t.Errorf("Expected 0 models for nonexistent dir, got %d", len(models))
	}
}

func TestListModelsWithModelDir(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a model directory with model.onnx
	modelDir := tmpDir + "/test-model"
	os.MkdirAll(modelDir, 0755)
	os.WriteFile(modelDir+"/model.onnx", []byte("fake"), 0644)

	models, err := ListModels(tmpDir)
	if err != nil {
		t.Fatalf("ListModels failed: %v", err)
	}
	if len(models) != 1 || models[0] != "test-model" {
		t.Errorf("Expected ['test-model'], got %v", models)
	}
}

func TestGetModelPath(t *testing.T) {
	path := GetModelPath("models", "test-model")
	if path == "" {
		t.Error("GetModelPath should return non-empty string")
	}
	t.Logf("Model path: %s", path)
}

func TestSearchModels(t *testing.T) {
	cfg := NewDownloadConfig()
	models, err := SearchModels("onnx", cfg)
	if err != nil {
		t.Logf("SearchModels error (network): %v", err)
		return
	}
	if len(models) == 0 {
		t.Error("SearchModels should return at least one result")
	}
	t.Logf("Search results: %v", models)
}

func TestGetModelInfo(t *testing.T) {
	cfg := NewDownloadConfig()
	info, err := GetModelInfo("test/model", cfg)
	if err != nil {
		t.Fatalf("GetModelInfo failed: %v", err)
	}
	if info["id"] != "test/model" {
		t.Errorf("Expected id 'test/model', got %v", info["id"])
	}
}

func TestSetProxy(t *testing.T) {
	SetProxy("http://testproxy:8080")
	if os.Getenv("HTTP_PROXY") != "http://testproxy:8080" {
		t.Errorf("HTTP_PROXY not set correctly")
	}
	os.Unsetenv("HTTP_PROXY")
	os.Unsetenv("HTTPS_PROXY")
}

func TestSetMirror(t *testing.T) {
	SetMirror("https://hf-mirror.com")
}

func TestExtractURLs(t *testing.T) {
	content := `Check http://example.com/api/health
Then POST http://example.com/api/deploy
More info at https://secure.example.com/test`

	urls := extractURLs(content)
	if len(urls) == 0 {
		t.Error("Should extract at least one URL")
	}
	t.Logf("Extracted URLs: %v", urls)
}

func TestParseLLMResult(t *testing.T) {
	result := `[GET] http://example.com/api/health
[POST] http://example.com/api/deploy
@ask: 是否继续执行?
@wait: 10min
@check: 验证部署结果
# This is a comment
> This is also a comment`

	tasks := parseLLMResult(result)
	if len(tasks) != 5 {
		t.Fatalf("Expected 5 tasks, got %d", len(tasks))
	}

	// Verify task types
	expectedTypes := []string{"http", "http", "ask", "wait", "check"}
	for i, task := range tasks {
		if string(task.Type) != expectedTypes[i] {
			t.Errorf("Task[%d]: expected type '%s', got '%s'", i, expectedTypes[i], task.Type)
		}
		t.Logf("Task[%d]: type=%s raw=%s", i, task.Type, task.Raw)
	}
}

func TestParseLLMResultEmpty(t *testing.T) {
	result := `No structured output here`
	tasks := parseLLMResult(result)
	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks for non-structured output, got %d", len(tasks))
	}
}

func TestParseHTTPRequestWithHeaders(t *testing.T) {
	result := `[GET] http://example.com/api/health header:Authorization=Bearer_token123 header:X-Request-Id=abc`
	tasks := parseLLMResult(result)
	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks))
	}

	task := tasks[0]
	if task.Type != "http" {
		t.Errorf("Expected http type, got %s", task.Type)
	}
	if task.HTTP == nil {
		t.Fatal("HTTP task should not be nil")
	}
	if task.HTTP.URL != "http://example.com/api/health" {
		t.Errorf("Expected URL 'http://example.com/api/health', got '%s'", task.HTTP.URL)
	}
	if task.HTTP.Method != "GET" {
		t.Errorf("Expected GET, got %s", task.HTTP.Method)
	}
	if len(task.HTTP.Headers) != 2 {
		t.Fatalf("Expected 2 headers, got %d", len(task.HTTP.Headers))
	}
	if task.HTTP.Headers["Authorization"] != "Bearer_token123" {
		t.Errorf("Expected Authorization=Bearer_token123, got %s", task.HTTP.Headers["Authorization"])
	}
	if task.HTTP.Headers["X-Request-Id"] != "abc" {
		t.Errorf("Expected X-Request-Id=abc, got %s", task.HTTP.Headers["X-Request-Id"])
	}
}

func TestParseHTTPPostWithHeadersAndBody(t *testing.T) {
	result := `[POST] http://example.com/api/deploy header:Content-Type=application/json header:Authorization=Bearer_token body:{"name":"test","version":"1.0"}`
	tasks := parseLLMResult(result)
	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks))
	}

	task := tasks[0]
	if task.HTTP.Method != "POST" {
		t.Errorf("Expected POST, got %s", task.HTTP.Method)
	}
	if task.HTTP.URL != "http://example.com/api/deploy" {
		t.Errorf("Expected URL 'http://example.com/api/deploy', got '%s'", task.HTTP.URL)
	}
	if len(task.HTTP.Headers) != 2 {
		t.Fatalf("Expected 2 headers, got %d: %v", len(task.HTTP.Headers), task.HTTP.Headers)
	}
	if task.HTTP.Headers["Content-Type"] != "application/json" {
		t.Errorf("Expected Content-Type=application/json, got %s", task.HTTP.Headers["Content-Type"])
	}
	if task.HTTP.Body != `{"name":"test","version":"1.0"}` {
		t.Errorf(`Expected body '{"name":"test","version":"1.0"}', got '%s'`, task.HTTP.Body)
	}
}

func TestParseHTTPPutWithBody(t *testing.T) {
	result := `[PUT] http://example.com/api/config body:{"key":"value"}`
	tasks := parseLLMResult(result)
	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks))
	}

	task := tasks[0]
	if task.HTTP.Method != "PUT" {
		t.Errorf("Expected PUT, got %s", task.HTTP.Method)
	}
	if task.HTTP.Body != `{"key":"value"}` {
		t.Errorf(`Expected body '{"key":"value"}', got '%s'`, task.HTTP.Body)
	}
	if len(task.HTTP.Headers) != 0 {
		t.Errorf("Expected 0 headers, got %d", len(task.HTTP.Headers))
	}
}

func TestParseHTTPDeleteWithHeaders(t *testing.T) {
	result := `[DELETE] http://example.com/api/resource/123 header:Authorization=Bearer_token`
	tasks := parseLLMResult(result)
	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks))
	}

	task := tasks[0]
	if task.HTTP.Method != "DELETE" {
		t.Errorf("Expected DELETE, got %s", task.HTTP.Method)
	}
	if task.HTTP.Headers["Authorization"] != "Bearer_token" {
		t.Errorf("Expected Authorization=Bearer_token, got %s", task.HTTP.Headers["Authorization"])
	}
}

func TestParseGRPCTaskWithHeadersAndBody(t *testing.T) {
	result := `[GRPC] localhost:50051/api.Service/Deploy header:Authorization=Bearer_token body:{"name":"test"}`
	tasks := parseLLMResult(result)
	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks))
	}

	task := tasks[0]
	if task.Type != "grpc" {
		t.Errorf("Expected grpc type, got %s", task.Type)
	}
	if task.GRPC == nil {
		t.Fatal("GRPC task should not be nil")
	}
	if task.GRPC.Host != "localhost" {
		t.Errorf("Expected host 'localhost', got '%s'", task.GRPC.Host)
	}
	if task.GRPC.Port != "50051" {
		t.Errorf("Expected port '50051', got '%s'", task.GRPC.Port)
	}
	if task.GRPC.Method != "api.Service/Deploy" {
		t.Errorf("Expected method 'api.Service/Deploy', got '%s'", task.GRPC.Method)
	}
	if task.GRPC.Headers["Authorization"] != "Bearer_token" {
		t.Errorf("Expected Authorization=Bearer_token, got %s", task.GRPC.Headers["Authorization"])
	}
	if task.GRPC.Body != `{"name":"test"}` {
		t.Errorf(`Expected body '{"name":"test"}', got '%s'`, task.GRPC.Body)
	}
}

func TestParseHTTPWithoutHeadersOrBody(t *testing.T) {
	result := `[GET] http://example.com/api/health`
	tasks := parseLLMResult(result)
	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks))
	}

	task := tasks[0]
	if task.HTTP.URL != "http://example.com/api/health" {
		t.Errorf("Expected URL 'http://example.com/api/health', got '%s'", task.HTTP.URL)
	}
	if len(task.HTTP.Headers) != 0 {
		t.Errorf("Expected 0 headers, got %d", len(task.HTTP.Headers))
	}
	if task.HTTP.Body != "" {
		t.Errorf("Expected empty body, got '%s'", task.HTTP.Body)
	}
}

func TestParseMixedTasksWithHeadersAndBody(t *testing.T) {
	result := `[GET] http://example.com/api/health header:Accept=application/json
[POST] http://example.com/api/deploy header:Content-Type=application/json header:Authorization=Bearer_tok body:{"app":"web","version":"2.0"}
[PUT] http://example.com/api/config body:{"debug":true}
[DELETE] http://example.com/api/old header:Authorization=Bearer_tok
[GRPC] grpc-server:9000/tx.Transaction/Commit header:token=abc123 body:{"txid":"001"}
@ask: 是否继续执行?
@wait: 10min
@check: 验证部署结果`

	tasks := parseLLMResult(result)
	if len(tasks) != 8 {
		t.Fatalf("Expected 8 tasks, got %d", len(tasks))
	}

	// GET with Accept header
	if tasks[0].HTTP.Headers["Accept"] != "application/json" {
		t.Errorf("GET: Expected Accept=application/json, got %s", tasks[0].HTTP.Headers["Accept"])
	}

	// POST with 2 headers + body
	if len(tasks[1].HTTP.Headers) != 2 {
		t.Errorf("POST: Expected 2 headers, got %d", len(tasks[1].HTTP.Headers))
	}
	if tasks[1].HTTP.Body != `{"app":"web","version":"2.0"}` {
		t.Errorf("POST: Expected body, got '%s'", tasks[1].HTTP.Body)
	}

	// PUT with body only
	if tasks[2].HTTP.Body != `{"debug":true}` {
		t.Errorf(`PUT: Expected body '{"debug":true}', got '%s'`, tasks[2].HTTP.Body)
	}

	// DELETE with Auth header
	if tasks[3].HTTP.Headers["Authorization"] != "Bearer_tok" {
		t.Errorf("DELETE: Expected Authorization=Bearer_tok, got %s", tasks[3].HTTP.Headers["Authorization"])
	}

	// GRPC with headers + body
	if tasks[4].GRPC.Host != "grpc-server" {
		t.Errorf("GRPC: Expected host 'grpc-server', got '%s'", tasks[4].GRPC.Host)
	}
	if tasks[4].GRPC.Port != "9000" {
		t.Errorf("GRPC: Expected port '9000', got '%s'", tasks[4].GRPC.Port)
	}
	if tasks[4].GRPC.Headers["token"] != "abc123" {
		t.Errorf("GRPC: Expected token=abc123, got %s", tasks[4].GRPC.Headers["token"])
	}
	if tasks[4].GRPC.Body != `{"txid":"001"}` {
		t.Errorf("GRPC: Expected body, got '%s'", tasks[4].GRPC.Body)
	}

	// @ask
	if tasks[5].Type != "ask" {
		t.Errorf("Expected ask type, got %s", tasks[5].Type)
	}

	// @wait
	if tasks[6].Type != "wait" {
		t.Errorf("Expected wait type, got %s", tasks[6].Type)
	}

	// @check
	if tasks[7].Type != "check" {
		t.Errorf("Expected check type, got %s", tasks[7].Type)
	}
}

func TestParseGRPCTarget(t *testing.T) {
	tests := []struct {
		input       string
		expectHost  string
		expectPort  string
		expectMethod string
	}{
		{"localhost:50051/api.Service/Deploy", "localhost", "50051", "api.Service/Deploy"},
		{"10.0.0.1:9000/tx.Commit", "10.0.0.1", "9000", "tx.Commit"},
		{"my-server:8080/v1/Create", "my-server", "8080", "v1/Create"},
		{"noservice", "noservice", "", ""},
		{"host:port/method/path", "host", "port", "method/path"},
	}

	for _, tc := range tests {
		host, port, method := parseGRPCTarget(tc.input)
		if host != tc.expectHost {
			t.Errorf("parseGRPCTarget(%q): expected host=%q, got %q", tc.input, tc.expectHost, host)
		}
		if port != tc.expectPort {
			t.Errorf("parseGRPCTarget(%q): expected port=%q, got %q", tc.input, tc.expectPort, port)
		}
		if method != tc.expectMethod {
			t.Errorf("parseGRPCTarget(%q): expected method=%q, got %q", tc.input, tc.expectMethod, method)
		}
	}
}

func TestParseHeadersAndBody(t *testing.T) {
	// URL only
	headers, body, target := parseHeadersAndBody("http://example.com/api")
	if target != "http://example.com/api" {
		t.Errorf("Expected target 'http://example.com/api', got '%s'", target)
	}
	if len(headers) != 0 {
		t.Errorf("Expected 0 headers, got %d", len(headers))
	}
	if body != "" {
		t.Errorf("Expected empty body, got '%s'", body)
	}

	// URL + header
	headers, body, target = parseHeadersAndBody("http://example.com/api header:Auth=token123")
	if target != "http://example.com/api" {
		t.Errorf("Expected target 'http://example.com/api', got '%s'", target)
	}
	if headers["Auth"] != "token123" {
		t.Errorf("Expected Auth=token123, got %s", headers["Auth"])
	}

	// URL + body
	headers, body, target = parseHeadersAndBody(`http://example.com/api body:{"key":"val"}`)
	if target != "http://example.com/api" {
		t.Errorf("Expected target 'http://example.com/api', got '%s'", target)
	}
	if body != `{"key":"val"}` {
		t.Errorf("Expected body, got '%s'", body)
	}

	// URL + multiple headers + body
	headers, body, target = parseHeadersAndBody(`http://example.com/api header:Content-Type=application/json header:Auth=Bearer_tok body:{"name":"test"}`)
	if target != "http://example.com/api" {
		t.Errorf("Expected target 'http://example.com/api', got '%s'", target)
	}
	if len(headers) != 2 {
		t.Errorf("Expected 2 headers, got %d", len(headers))
	}
	if headers["Content-Type"] != "application/json" {
		t.Errorf("Expected Content-Type=application/json, got %s", headers["Content-Type"])
	}
	if headers["Auth"] != "Bearer_tok" {
		t.Errorf("Expected Auth=Bearer_tok, got %s", headers["Auth"])
	}
	if body != `{"name":"test"}` {
		t.Errorf("Expected body, got '%s'", body)
	}
}

func TestDefaultGenAILibraryPath(t *testing.T) {
	path := DefaultGenAILibraryPath()
	if path == "" {
		t.Error("DefaultGenAILibraryPath should return non-empty string")
	}
	t.Logf("GenAI library path: %s", path)
}

func TestDefaultOnnxRuntimeLibraryPath(t *testing.T) {
	path := DefaultOnnxRuntimeLibraryPath()
	if path == "" {
		t.Error("DefaultOnnxRuntimeLibraryPath should return non-empty string")
	}
	t.Logf("ONNX Runtime library path: %s", path)
}

func TestBuildAnalyzePrompt(t *testing.T) {
	content := "Sample task content"
	prompt := buildAnalyzePrompt(content)
	if prompt == "" {
		t.Error("buildAnalyzePrompt should return non-empty string")
	}
	t.Logf("Prompt length: %d", len(prompt))
}

func TestDeepSeekR1RealInference(t *testing.T) {
	modelDir := deepseekModelDir()

	// Verify model files exist
	configPath := filepath.Join(modelDir, "genai_config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Skipf("Model not found at %s, skipping real inference test", modelDir)
	}

	model := NewModel("deepseek-r1-distill-qwen-1.5B")
	if err := model.Load(modelDir); err != nil {
		t.Fatalf("Failed to load model: %v", err)
	}
	defer model.Unload()

	if !model.IsLoaded() {
		t.Fatal("Model should be loaded after Load()")
	}

	// Test real inference with a simple prompt
	result, err := model.Infer("What is 2+3?")
	if err != nil {
		t.Fatalf("Inference failed: %v", err)
	}

	if result == "" {
		t.Error("Inference should return non-empty string")
	}

	t.Logf("Inference result: %s", result)
}

func TestDeepSeekR1StreamInference(t *testing.T) {
	modelDir := deepseekModelDir()

	configPath := filepath.Join(modelDir, "genai_config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Skipf("Model not found at %s, skipping stream inference test", modelDir)
	}

	model := NewModel("deepseek-r1-distill-qwen-1.5B")
	if err := model.Load(modelDir); err != nil {
		t.Fatalf("Failed to load model: %v", err)
	}
	defer model.Unload()

	var tokens []string
	opts := GenerateOptions{
		MaxTokens:   256,
		Temperature: 0.6,
		TopP:        0.95,
		DoSample:    true,
		StopOnEos:   true,
	}

	err := model.InferStream("Hello", opts, func(text string) error {
		t.Logf("Token: %s", text)
		tokens = append(tokens, text)
		return nil
	})

	if err != nil {
		t.Fatalf("Stream inference failed: %v", err)
	}

	if len(tokens) == 0 {
		t.Error("Stream inference should produce at least one token")
	}

	t.Logf("Stream result: %s", strings.Join(tokens, ""))
}

func TestDeepSeekR1TaskAnalysis(t *testing.T) {
	modelDir := deepseekModelDir()

	configPath := filepath.Join(modelDir, "genai_config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Skipf("Model not found at %s, skipping task analysis test", modelDir)
	}

	model := NewModel("deepseek-r1-distill-qwen-1.5B")
	if err := model.Load(modelDir); err != nil {
		t.Fatalf("Failed to load model: %v", err)
	}
	defer model.Unload()

	analyzer := NewTaskAnalyzer()
	analyzer.SetModel(model)

	content := `Migration from prod to UAT:
1. Check service health at http://example.com/api/health
2. Deploy new version at http://example.com/api/deploy
3. Wait 10 minutes for deployment
4. Verify deployment result`

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

func TestDeepSeekR1ProdMigrationAnalysis(t *testing.T) {
	modelDir := deepseekModelDir()

	configPath := filepath.Join(modelDir, "genai_config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Skipf("Model not found at %s, skipping prod migration test", modelDir)
	}

	model := NewModel("deepseek-r1-distill-qwen-1.5B")
	if err := model.Load(modelDir); err != nil {
		t.Fatalf("Failed to load model: %v", err)
	}
	defer model.Unload()

	// Read the real prod-migration-form-uat.txt
	migrationFile := filepath.Join(projectRoot(), "tests", "data", "prod-migration-form-uat.txt")
	content, err := os.ReadFile(migrationFile)
	if err != nil {
		t.Fatalf("Failed to read migration file: %v", err)
	}

	analyzer := NewTaskAnalyzer()
	analyzer.SetModel(model)

	tasks, err := analyzer.AnalyzeContent(string(content))
	if err != nil {
		t.Fatalf("AnalyzeContent failed: %v", err)
	}

	if len(tasks) == 0 {
		t.Error("AnalyzeContent should return at least one task from prod migration file")
	}

	t.Logf("Total tasks extracted: %d", len(tasks))
	for i, task := range tasks {
		t.Logf("Task[%d]: type=%s raw=%s", i, task.Type, task.Raw)
	}

	// Also test direct inference with migration content as prompt
	cfg := NewConfig()
	cfg.MaxLength = 512
	cfg.Temperature = 0.6

	prompt := buildAnalyzePrompt(string(content))
	result, err := model.InferWithConfig(prompt, cfg)
	if err != nil {
		t.Fatalf("Direct inference failed: %v", err)
	}

	if result == "" {
		t.Error("Direct inference should return non-empty string")
	}

	t.Logf("Direct inference result:\n%s", result)
}
