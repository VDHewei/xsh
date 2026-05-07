package parser

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/VDHewei/xsh/internal/types"
)

// ParseFile 解析任务文件
func ParseFile(filename string) ([]*types.Task, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var tasks []*types.Task
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ">") {
			continue
		}

		task, err := parseLine(line)
		if err != nil {
			return nil, fmt.Errorf("failed to parse line: %w", err)
		}
		if task != nil {
			tasks = append(tasks, task)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}

	return tasks, nil
}

// parseLine 解析单行
func parseLine(line string) (*types.Task, error) {
	// 检查是否是 HTTP 请求 [METHOD] url
	httpPattern := regexp.MustCompile(`^\[(\w+)\]\s+(.+)$`)
	if match := httpPattern.FindStringSubmatch(line); match != nil {
		method := types.HTTPMethod(strings.ToUpper(match[1]))
		url := strings.TrimSpace(match[2])
		return &types.Task{
			Type: types.TaskTypeHTTP,
			Raw:  line,
			HTTP: &types.HTTPTask{
				Method: method,
				URL:    url,
			},
		}, nil
	}

	// 检查是否是命令 @command:description
	cmdPattern := regexp.MustCompile(`^@(\w+):\s*(.+)$`)
	if match := cmdPattern.FindStringSubmatch(line); match != nil {
		cmd := strings.ToLower(match[1])
		desc := strings.TrimSpace(match[2])

		switch cmd {
		case "ask":
			return &types.Task{
				Type: types.TaskTypeAsk,
				Raw:  line,
				Ask: &types.AskTask{
					Prompt: desc,
				},
			}, nil
		case "wait":
			return &types.Task{
				Type: types.TaskTypeWait,
				Raw:  line,
				Wait: &types.WaitTask{
					Duration: desc,
				},
			}, nil
		case "check":
			return &types.Task{
				Type: types.TaskTypeCheck,
				Raw:  line,
				Check: &types.CheckTask{
					Prompt: desc,
				},
			}, nil
		}
	}

	return nil, nil
}