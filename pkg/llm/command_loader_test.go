package llm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewCommandLoader(t *testing.T) {
	loader := NewCommandLoader("testdir")
	if loader == nil {
		t.Fatal("NewCommandLoader should return non-nil loader")
	}
	if loader.dir != "testdir" {
		t.Errorf("Expected dir 'testdir', got '%s'", loader.dir)
	}
	if loader.analyzer == nil {
		t.Error("Expected non-nil analyzer")
	}
}

func TestCommandLoaderScan(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .md files
	os.WriteFile(filepath.Join(tmpDir, "deploy.md"), []byte("# deploy"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "health.md"), []byte("# health"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "migration.md"), []byte("# migration"), 0644)
	// Non-.md files should be ignored
	os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("text"), 0644)
	// Subdirectory should be ignored
	os.MkdirAll(filepath.Join(tmpDir, "sub"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "sub", "nested.md"), []byte("# nested"), 0644)

	loader := NewCommandLoader(tmpDir)
	names, err := loader.Scan()
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(names) != 3 {
		t.Errorf("Expected 3 commands, got %d: %v", len(names), names)
	}

	// Names should be sorted
	if !strings.EqualFold(names[0], "deploy") {
		t.Errorf("Expected 'deploy' first, got '%s'", names[0])
	}
	if !strings.EqualFold(names[1], "health") {
		t.Errorf("Expected 'health' second, got '%s'", names[1])
	}
	if !strings.EqualFold(names[2], "migration") {
		t.Errorf("Expected 'migration' third, got '%s'", names[2])
	}
}

func TestCommandLoaderScan_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewCommandLoader(tmpDir)
	names, err := loader.Scan()
	if err != nil {
		t.Fatalf("Scan should not fail for empty dir: %v", err)
	}
	if len(names) != 0 {
		t.Errorf("Expected 0 commands in empty dir, got %d", len(names))
	}
}

func TestCommandLoaderScan_NonexistentDir(t *testing.T) {
	loader := NewCommandLoader("/nonexistent/commands")
	_, err := loader.Scan()
	if err == nil {
		t.Error("Scan should return error for nonexistent directory")
	}
}

func TestCommandLoaderLoad(t *testing.T) {
	tmpDir := t.TempDir()

	content := `# 部署命令

## 描述
自动部署到生产环境

## 任务
[GET] http://example.com/health
[POST] http://example.com/deploy
@ask: 确认部署?
@wait: 5min
@check: 验证部署结果`

	os.WriteFile(filepath.Join(tmpDir, "deploy.md"), []byte(content), 0644)

	loader := NewCommandLoader(tmpDir)
	cmd, err := loader.Load("deploy")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cmd == nil {
		t.Fatal("Expected non-nil command")
	}
	if cmd.Name != "deploy" {
		t.Errorf("Expected name 'deploy', got '%s'", cmd.Name)
	}
	if cmd.Desc != "自动部署到生产环境" {
		t.Errorf("Expected desc '自动部署到生产环境', got '%s'", cmd.Desc)
	}
	if len(cmd.Tasks) == 0 {
		t.Error("Expected at least one task")
	}
	if len(cmd.Tasks) != 5 {
		t.Errorf("Expected 5 tasks, got %d", len(cmd.Tasks))
	}

	// Verify task types
	expectedTypes := []string{"http", "http", "ask", "wait", "check"}
	for i, task := range cmd.Tasks {
		if string(task.Type) != expectedTypes[i] {
			t.Errorf("Task[%d]: expected type '%s', got '%s'", i, expectedTypes[i], task.Type)
		}
	}
	t.Logf("Loaded command: %s with %d tasks", cmd.Name, len(cmd.Tasks))
}

func TestCommandLoaderLoad_Nonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewCommandLoader(tmpDir)
	_, err := loader.Load("nonexistent")
	if err == nil {
		t.Error("Load should return error for nonexistent file")
	}
}

func TestParseCommand(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewCommandLoader(tmpDir)

	content := `# 健康检查

## 描述
检查服务健康状态

## 任务
[GET] http://example.com/api/health
@check: 验证返回状态码为200`

	cmd, err := loader.ParseCommand(content, "health", "health.md")
	if err != nil {
		t.Fatalf("ParseCommand failed: %v", err)
	}
	if cmd.Name != "health" {
		t.Errorf("Expected name 'health', got '%s'", cmd.Name)
	}
	if cmd.Desc != "检查服务健康状态" {
		t.Errorf("Expected desc '检查服务健康状态', got '%s'", cmd.Desc)
	}
	if len(cmd.Tasks) != 2 {
		t.Fatalf("Expected 2 tasks, got %d", len(cmd.Tasks))
	}
	if cmd.Tasks[0].Type != "http" {
		t.Errorf("Expected http task, got %s", cmd.Tasks[0].Type)
	}
	if cmd.Tasks[1].Type != "check" {
		t.Errorf("Expected check task, got %s", cmd.Tasks[1].Type)
	}
}

func TestParseCommand_MissingTaskSection(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewCommandLoader(tmpDir)

	content := `# 空命令

## 描述
这个命令没有任务部分`

	_, err := loader.ParseCommand(content, "empty", "empty.md")
	if err == nil {
		t.Error("ParseCommand should return error when no ## 任务 section")
	}
	if !strings.Contains(err.Error(), "no ## 任务 section") {
		t.Errorf("Expected 'no ## 任务 section' error, got: %v", err)
	}
}

func TestParseCommand_NoDescription(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewCommandLoader(tmpDir)

	content := `# 无描述命令

## 任务
[GET] http://example.com/api`

	cmd, err := loader.ParseCommand(content, "nodesc", "nodesc.md")
	if err != nil {
		t.Fatalf("ParseCommand failed: %v", err)
	}
	if cmd.Desc != "" {
		t.Errorf("Expected empty desc, got '%s'", cmd.Desc)
	}
}

func TestExtractDescription(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "standard description",
			content:  "## 描述\n自动部署到生产环境\n\n## 任务",
			expected: "自动部署到生产环境",
		},
		{
			name:     "multiline description",
			content:  "## 描述\n第一行描述\n第二行描述\n\n## 任务",
			expected: "第一行描述\n第二行描述",
		},
		{
			name:     "no description section",
			content:  "# 标题\n## 任务\n[GET] http://example.com",
			expected: "",
		},
		{
			name:     "empty description",
			content:  "## 描述\n\n## 任务",
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := extractDescription(tc.content)
			if got != tc.expected {
				t.Errorf("extractDescription: expected '%s', got '%s'", tc.expected, got)
			}
		})
	}
}

func TestExtractTaskSection(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "standard task section",
			content:  "# 标题\n\n## 任务\n[GET] http://example.com/api\n@ask: 确认\n\n## 备注",
			expected: "[GET] http://example.com/api\n@ask: 确认",
		},
		{
			name:     "task section at end",
			content:  "# 标题\n\n## 描述\n一些描述\n\n## 任务\n[GET] http://example.com/api",
			expected: "[GET] http://example.com/api",
		},
		{
			name:     "no task section",
			content:  "# 标题\n\n## 描述\n一些描述",
			expected: "",
		},
		{
			name:     "tasks with headers and body",
			content:  "## 任务\n[POST] http://example.com/deploy header:Content-Type=application/json body:{\"key\":\"val\"}\n@check: 验证",
			expected: "[POST] http://example.com/deploy header:Content-Type=application/json body:{\"key\":\"val\"}\n@check: 验证",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := extractTaskSection(tc.content)
			if got != tc.expected {
				t.Errorf("extractTaskSection: expected '%s', got '%s'", tc.expected, got)
			}
		})
	}
}
