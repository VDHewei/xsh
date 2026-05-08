package parser

import (
	"fmt"
	"os"
	"testing"
)

func TestParseFile(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := tmpDir + "/test_tasks.txt"

	content := `# Comment line
> Another comment
[GET] http://example.com/api/health
@wait: 10min
@ask: Is the service healthy?
[POST] http://example.com/api/deploy
@check: Verify deployment result
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tasks, err := ParseFile(testFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(tasks) != 5 {
		t.Fatalf("Expected 5 tasks, got %d", len(tasks))
	}

	// Verify task types
	expectedTypes := []string{"http", "wait", "ask", "http", "check"}
	for i, task := range tasks {
		if string(task.Type) != expectedTypes[i] {
			t.Errorf("Task[%d]: expected type '%s', got '%s'", i, expectedTypes[i], task.Type)
		}
		t.Logf("Task[%d]: type=%s raw=%s", i, task.Type, task.Raw)
	}
}

func TestParseHTTPGet(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := tmpDir + "/http_test.txt"

	content := `[GET] http://example.com/api/test`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tasks, err := ParseFile(testFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks))
	}

	if tasks[0].Type != "http" {
		t.Errorf("Expected http type, got %s", tasks[0].Type)
	}
	if tasks[0].HTTP == nil {
		t.Fatal("HTTP task should not be nil")
	}
	if tasks[0].HTTP.Method != "GET" {
		t.Errorf("Expected GET method, got %s", tasks[0].HTTP.Method)
	}
	if tasks[0].HTTP.URL != "http://example.com/api/test" {
		t.Errorf("Expected URL 'http://example.com/api/test', got '%s'", tasks[0].HTTP.URL)
	}
}

func TestParseHTTPPost(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := tmpDir + "/http_post_test.txt"

	content := `[POST] http://example.com/api/deploy`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tasks, err := ParseFile(testFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks))
	}
	if tasks[0].HTTP.Method != "POST" {
		t.Errorf("Expected POST method, got %s", tasks[0].HTTP.Method)
	}
}

func TestParseAskCommand(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := tmpDir + "/ask_test.txt"

	content := `@ask: 检查服务是否正常?`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tasks, err := ParseFile(testFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Type != "ask" {
		t.Errorf("Expected ask type, got %s", tasks[0].Type)
	}
	if tasks[0].Ask == nil {
		t.Fatal("Ask task should not be nil")
	}
	if tasks[0].Ask.Prompt != "检查服务是否正常?" {
		t.Errorf("Expected prompt '检查服务是否正常?', got '%s'", tasks[0].Ask.Prompt)
	}
}

func TestParseWaitCommand(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := tmpDir + "/wait_test.txt"

	content := `@wait: 10min`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tasks, err := ParseFile(testFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Type != "wait" {
		t.Errorf("Expected wait type, got %s", tasks[0].Type)
	}
	if tasks[0].Wait.Duration != "10min" {
		t.Errorf("Expected duration '10min', got '%s'", tasks[0].Wait.Duration)
	}
}

func TestParseCheckCommand(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := tmpDir + "/check_test.txt"

	content := `@check: Verify deployment is complete`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tasks, err := ParseFile(testFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Type != "check" {
		t.Errorf("Expected check type, got %s", tasks[0].Type)
	}
}

func TestParseEmptyLines(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := tmpDir + "/empty_test.txt"

	content := `

[GET] http://example.com/api/health

@ask: Continue?

`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tasks, err := ParseFile(testFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("Expected 2 tasks, got %d", len(tasks))
	}
}

func TestParseNonexistentFile(t *testing.T) {
	_, err := ParseFile("/nonexistent/file.txt")
	if err == nil {
		t.Error("Should return error for nonexistent file")
	}
}

func TestParseFileWithOnlyComments(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := tmpDir + "/comments_only.txt"

	content := `# Comment 1
> Comment 2
# Comment 3`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tasks, err := ParseFile(testFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks for comments-only file, got %d", len(tasks))
	}
}

func TestParseHTTPPut(t *testing.T) {
	testParseHTTPMethod(t, "PUT", "http://example.com/api/update")
}

func TestParseHTTPPatch(t *testing.T) {
	testParseHTTPMethod(t, "PATCH", "http://example.com/api/partial-update")
}

func TestParseHTTPDelete(t *testing.T) {
	testParseHTTPMethod(t, "DELETE", "http://example.com/api/resource/123")
}

func TestParseHTTPHead(t *testing.T) {
	testParseHTTPMethod(t, "HEAD", "http://example.com/api/health")
}

func TestParseHTTPOptions(t *testing.T) {
	testParseHTTPMethod(t, "OPTIONS", "http://example.com/api/resource")
}

func testParseHTTPMethod(t *testing.T, method, url string) {
	t.Helper()
	tmpDir := t.TempDir()
	testFile := tmpDir + "/http_method_test.txt"

	content := fmt.Sprintf("[%s] %s", method, url)
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tasks, err := ParseFile(testFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Type != "http" {
		t.Errorf("Expected http type, got %s", tasks[0].Type)
	}
	if tasks[0].HTTP == nil {
		t.Fatal("HTTP task should not be nil")
	}
	if string(tasks[0].HTTP.Method) != method {
		t.Errorf("Expected %s method, got %s", method, tasks[0].HTTP.Method)
	}
	if tasks[0].HTTP.URL != url {
		t.Errorf("Expected URL '%s', got '%s'", url, tasks[0].HTTP.URL)
	}
}

func TestParseMixedTasks(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := tmpDir + "/mixed_tasks.txt"

	// 混合多种任务类型，包含不同的 HTTP 方法
	content := `# Migration workflow
[GET] http://example.com/api/health
@wait: 1min
[POST] http://example.com/api/create
@ask: Is the deployment complete?
[PUT] http://example.com/api/config
@check: Verify configuration update
[DELETE] http://example.com/api/temp
@wait: 30s
[HEAD] http://example.com/api/status
[POST] http://example.com/api/notify
@ask: Final confirmation?`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tasks, err := ParseFile(testFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	// 1 comment + 11 tasks = 11
	expectedCount := 11
	if len(tasks) != expectedCount {
		t.Fatalf("Expected %d tasks, got %d", expectedCount, len(tasks))
	}

	var methodCounts = map[string]int{}
	var taskTypeCounts = map[string]int{}
	for _, task := range tasks {
		taskTypeCounts[string(task.Type)]++
		if task.HTTP != nil {
			methodCounts[string(task.HTTP.Method)]++
		}
		t.Logf("Task: type=%s raw=%s", task.Type, task.Raw)
	}

	// 验证 HTTP 方法覆盖
	expectedMethods := []string{"GET", "POST", "PUT", "DELETE", "HEAD"}
	for _, m := range expectedMethods {
		if methodCounts[m] == 0 {
			t.Errorf("Missing HTTP method: %s", m)
		}
	}
	if methodCounts["POST"] < 2 {
		t.Errorf("Expected at least 2 POST tasks, got %d", methodCounts["POST"])
	}

	// 验证任务类型覆盖
	if taskTypeCounts["http"] < 6 {
		t.Errorf("Expected at least 6 HTTP tasks, got %d", taskTypeCounts["http"])
	}
	if taskTypeCounts["wait"] < 2 {
		t.Errorf("Expected at least 2 wait tasks, got %d", taskTypeCounts["wait"])
	}
	if taskTypeCounts["ask"] < 2 {
		t.Errorf("Expected at least 2 ask tasks, got %d", taskTypeCounts["ask"])
	}
	if taskTypeCounts["check"] < 1 {
		t.Errorf("Expected at least 1 check task, got %d", taskTypeCounts["check"])
	}
}

func TestParseUnknownCommand(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := tmpDir + "/unknown_cmd.txt"

	content := `@unknown: this should be skipped
[GET] http://example.com/api/health
@notask: also skipped`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tasks, err := ParseFile(testFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	// Unknown commands should be silently skipped, only GET remains
	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task (unknown commands skipped), got %d", len(tasks))
	}
	if tasks[0].Type != "http" {
		t.Errorf("Expected http type from GET line, got %s", tasks[0].Type)
	}
}

func TestParseMalformedLines(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := tmpDir + "/malformed.txt"

	content := `# Valid lines mixed with malformed
[GET] http://example.com/api/good
[MISSING_BRACKET http://example.com/bad
[POST] http://example.com/api/good2
@@ double at sign
[GET] http://example.com/api/good3
@ invalid space after at
just plain text line
@ask: A valid ask after malformed`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tasks, err := ParseFile(testFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	// Only well-formed lines should parse: 2 GET + 1 POST + 1 ask = 4
	if len(tasks) != 4 {
		t.Fatalf("Expected 4 tasks (malformed lines skipped), got %d", len(tasks))
	}

	// Verify all tasks are valid types
	for i, task := range tasks {
		if task.Type != "http" && task.Type != "ask" && task.Type != "wait" && task.Type != "check" {
			t.Errorf("Task[%d]: unexpected type %s", i, task.Type)
		}
	}
}
