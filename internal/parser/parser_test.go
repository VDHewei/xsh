package parser

import (
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
