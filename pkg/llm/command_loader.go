package llm

import (
	"fmt"
	"os"
	"path/filepath"
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

	// 尝试用 LLM analyzer 解析, 无模型时直接用 regex
	var tasks []*types.Task
	var err error
	if l.analyzer.model != nil && l.analyzer.model.IsLoaded() {
		tasks, err = l.analyzer.AnalyzeContent(taskContent)
		// LLM 失败时降级
		if err != nil || len(tasks) == 0 {
			tasks, err = parseTaskLines(taskContent)
		}
	} else {
		// 无 LLM 模型时直接用 regex 解析（命令文件中的任务已有明确结构）
		tasks, err = parseTaskLines(taskContent)
	}
	if err != nil {
		return cmd, fmt.Errorf("parse command tasks: %w", err)
	}

	cmd.Tasks = tasks
	return cmd, nil
}

// extractDescription 从 markdown 中提取 ## 描述 区块
func extractDescription(content string) string {
	return extractSection(content, "描述")
}

// extractTaskSection 从 markdown 中提取 ## 任务 区块
func extractTaskSection(content string) string {
	return extractSection(content, "任务")
}

// extractSection 从 markdown 中提取指定的 ## 区块 (通用实现, 避免 Go regex 不支持 lookahead)
func extractSection(content, keyword string) string {
	lines := strings.Split(content, "\n")
	var result []string
	inSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") && strings.Contains(trimmed, keyword) {
			inSection = true
			continue
		}
		if inSection {
			// 遇到下一个标题行则结束
			if strings.HasPrefix(trimmed, "#") {
				break
			}
			result = append(result, line)
		}
	}

	return strings.TrimSpace(strings.Join(result, "\n"))
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
