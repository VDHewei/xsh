package llm

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/VDHewei/xsh/internal/parser"
	"github.com/VDHewei/xsh/internal/types"
)

// CommandLoaderImpl 命令加载器实现
type CommandLoaderImpl struct {
	dir      string
	analyzer *TaskAnalyzer
}

// NewCommandLoader 创建命令加载器
func NewCommandLoader(dir string) *CommandLoaderImpl {
	return &CommandLoaderImpl{
		dir:      dir,
		analyzer: NewTaskAnalyzer(),
	}
}

// Scan 扫描 commands/ 目录, 返回命令名称列表
func (l *CommandLoaderImpl) Scan() ([]string, error) {
	entries, err := os.ReadDir(l.dir)
	if err != nil {
		return nil, fmt.Errorf("scan directory %s: %w", l.dir, err)
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".md") {
			name := strings.TrimSuffix(entry.Name(), ".md")
			names = append(names, name)
		}
	}

	sort.Strings(names)
	return names, nil
}

// Load 加载命令文件并解析
func (l *CommandLoaderImpl) Load(name string) (*types.CustomCommand, error) {
	filePath := filepath.Join(l.dir, name+".md")
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read command file %s: %w", filePath, err)
	}

	return l.ParseCommand(string(content), name, filePath)
}

// ParseCommand 解析 .md 内容为自定义命令
func (l *CommandLoaderImpl) ParseCommand(content, name, filePath string) (*types.CustomCommand, error) {
	cmd := &types.CustomCommand{
		Name:    name,
		File:    filePath,
		Content: content,
		Desc:    extractDescription(content),
	}

	// 提取 ## 任务 区块
	taskContent := extractTaskSection(content)
	if taskContent == "" {
		return cmd, fmt.Errorf("no ## 任务 section found in %s", filePath)
	}

	// 尝试用 LLM analyzer 解析
	tasks, err := l.analyzer.AnalyzeContent(taskContent)
	if err != nil || len(tasks) == 0 {
		// LLM 失败或无模型, 降级为 regex 解析
		tasks, err = parseTaskLines(taskContent)
		if err != nil {
			return cmd, fmt.Errorf("parse command tasks: %w", err)
		}
	}

	cmd.Tasks = tasks
	return cmd, nil
}

// extractDescription 从 markdown 中提取 ## 描述 区块
func extractDescription(content string) string {
	re := regexp.MustCompile(`(?m)^##\s*描述\s*\n([\s\S]*?)(?=\n##|$)`)
	matches := re.FindStringSubmatch(content)
	if len(matches) < 2 {
		return ""
	}
	return strings.TrimSpace(matches[1])
}

// extractTaskSection 从 markdown 中提取 ## 任务 区块
func extractTaskSection(content string) string {
	re := regexp.MustCompile(`(?m)^##\s*任务\s*\n([\s\S]*?)(?=\n##|\n#|$)`)
	matches := re.FindStringSubmatch(content)
	if len(matches) < 2 {
		return ""
	}
	return strings.TrimSpace(matches[1])
}

// parseTaskLines 使用 regex 解析任务行 (LLM 降级方案)
func parseTaskLines(content string) ([]*types.Task, error) {
	// 写入临时文件, 复用 parser.ParseFile
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("xsh_cmd_%d.txt", os.Getpid()))
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("write temp file: %w", err)
	}
	defer os.Remove(tmpFile)

	tasks, err := parser.ParseFile(tmpFile)
	if err != nil {
		return nil, fmt.Errorf("parse task lines: %w", err)
	}

	return tasks, nil
}
