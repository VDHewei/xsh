package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/VDHewei/xsh/pkg/llm"
)

// TaskRequest 任务请求
type TaskRequest struct {
	Task     string   `json:"task"`     // 任务描述
	Tools   []string `json:"tools"`   // 可用工具
	Context string   `json:"context"` // 上下文
}

// TaskResult 任务结果
type TaskResult struct {
	Plan    string        `json:"plan"`    // 任务规划
	Steps   []TaskStep    `json:"steps"`   // 执行步骤
	Results []StepResult `json:"results"` // 执行结果
}

// TaskStep 任务步骤
type TaskStep struct {
	ID      string `json:"id"`      // 步骤 ID
	Action  string `json:"action"`  // 执行动作
	Tool   string `json:"tool"`   // 使用的工具
	Params string `json:"params"` // 参数
	Status string `json:"status"` // 状态: pending/running/completed/failed
}

// StepResult 步骤结果
type StepResult struct {
	StepID  string `json:"step_id"`  // 步骤 ID
	Output string `json:"output"` // 输出
	Error  string `json:"error"`  // 错误
}

// ToolCall 工具调用请求
type ToolCall struct {
	Tool  string            `json:"tool"`  // 工具名
	Args  map[string]string `json:"args"`  // 参数
}

var (
	modelPath = flag.String("model", "", "Path to ONNX model")
	testFile  = flag.String("input", "tests/data/prod-migration-form-uat.txt", "Input test file")
	download  = flag.Bool("download", false, "Download model from HuggingFace")
)

func main() {
	flag.Parse()

	// 如果没有指定模型，尝试使用默认路径
	if *modelPath == "" {
		// 查找本地模型
		*modelPath = findModel()
	}

	fmt.Println("=== ONNX LLM 任务规划测试 ===")
	fmt.Printf("Model: %s\n", *modelPath)
	fmt.Printf("Input: %s\n\n", *testFile)

	// 如果指定了下载选项，下载模型
	if *download {
		downloadModel()
		return
	}

	// 加载测试数据
	task, err := os.ReadFile(*testFile)
	if err != nil {
		log.Fatalf("Failed to read test file: %v", err)
	}

	// 如果指定了模型，使用 ONNX 推理
	if *modelPath != "" && fileExists(*modelPath) {
		fmt.Println("使用 ONNX 模型推理...")
		testWithONNX(string(task), *modelPath)
	} else {
		fmt.Println("使用 Mock 推理...")
		fmt.Println("提示: 使用 -download 下载模型，或指定 -model 参数")
		testWithMock(string(task))
	}
}

// findModel 查找本地模型
func findModel() string {
	paths := []string{
		"models/deepseek-r1-distill-qwen-1.5B/cpu_and_mobile/int4-rtn/block-32_acc-level-4/decoder_model.onnx",
		"models/DeepSeek-R1-Distill-ONNX/deepseek-r1-distill-qwen-1.5B/cpu_and_mobile/int4-rtn/block-32_acc-level-4/decoder_model.onnx",
		"models/model.onnx",
	}

	for _, p := range paths {
		if fileExists(p) {
			return p
		}
	}
	return ""
}

// fileExists 检查文件是否存在
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// downloadModel 下载模型
func downloadModel() {
	fmt.Println("开始下载 DeepSeek-R1-Distill-ONNX 模型...")
	fmt.Println("模型: onnxruntime/DeepSeek-R1-Distill-ONNX")
	fmt.Println("版本: deepseek-r1-distill-qwen-1.5B CPU int4")
	fmt.Println()

	cfg := llm.NewDownloadConfig()
	cfg.CacheDir = "models"

	// 下载 1.5B CPU 版本
	repoID := "onnxruntime/DeepSeek-R1-Distill-ONNX"

	model, err := llm.DownloadFromHuggingFace(repoID, cfg)
	if err != nil {
		log.Fatalf("Failed to download model: %v", err)
	}

	fmt.Printf("模型下载完成!\n")
	fmt.Printf("模型路径: %s\n", model.Path)
	fmt.Printf("模型大小: %d bytes\n", model.Size)
	fmt.Println()
	fmt.Println("请使用以下命令运行测试:")
	fmt.Printf("  go run ./cmd/xsh onnx-test -model %s\n", filepath.Join(model.Path, "model.onnx"))
}

// executeToolCall 执行工具调用
func executeToolCall(tc ToolCall) (string, error) {
	switch tc.Tool {
	case "http-get":
		url := tc.Args["url"]
		return fmt.Sprintf("GET %s -> 200 OK (mock)", url), nil
	case "http-post":
		url := tc.Args["url"]
		return fmt.Sprintf("POST %s -> 200 OK (mock)", url), nil
	case "sleep":
		return fmt.Sprintf("Waited %s (mock)", tc.Args["duration"]), nil
	case "ask":
		return fmt.Sprintf("Asked user: %s (mock)", tc.Args["question"]), nil
	default:
		return "", fmt.Errorf("unknown tool: %s", tc.Tool)
	}
}

// testWithONNX 使用 ONNX 模型测试
func testWithONNX(task, modelPath string) {
	model := llm.NewModel("deepseek-r1-distill")
	defer model.Unload()

	fmt.Printf("加载模型: %s\n", modelPath)
	if err := model.Load(modelPath); err != nil {
		log.Fatalf("Failed to load model: %v", err)
	}

	prompt := buildPrompt(task)

	fmt.Println("\n执行推理...")
	result, err := model.Infer(prompt)
	if err != nil {
		log.Fatalf("Failed to infer: %v", err)
	}

	fmt.Printf("\n推理结果:\n%s\n", result)

	executeTaskResult(result)
}

// testWithMock 使用 Mock 测试
func testWithMock(task string) {
	result := llm.MockInfer(task)
	fmt.Printf("Mock 推理结果:\n%s\n", result)

	executeTaskResult(result)
}

// buildPrompt 构建推理提示
func buildPrompt(taskContent string) string {
	return fmt.Sprintf(`你是一个任务规划和执行助手。请分析以下迁移步骤，生成任务规划并执行。

迁移步骤:
%s

请生成 JSON 格式的任务规划，包含:
1. plan: 任务总体描述
2. steps: 执行步骤数组，每个步骤包含:
   - id: 步骤ID
   - action: 执行动作
   - tool: 使用的工具 (http-get, http-post, sleep, ask)
   - params: 参数
   - status: 状态 (pending)

请只输出 JSON，不要其他内容。`, taskContent)
}

// executeTaskResult 执行任务结果
func executeTaskResult(result string) {
	fmt.Println("\n=== 执行任务 ===")

	var taskResult TaskResult
	if err := json.Unmarshal([]byte(result), &taskResult); err != nil {
		fmt.Println("使用模拟执行...")
		mockExecute(result)
		return
	}

	fmt.Printf("计划: %s\n", taskResult.Plan)
	fmt.Printf("步骤数: %d\n\n", len(taskResult.Steps))

	for i, step := range taskResult.Steps {
		fmt.Printf("[%d/%d] Step: %s\n", i+1, len(taskResult.Steps), step.ID)
		fmt.Printf("    Action: %s\n", step.Action)
		fmt.Printf("    Tool: %s\n", step.Tool)

		tc := ToolCall{
			Tool: step.Tool,
			Args: map[string]string{
				"url":      step.Params,
				"question": step.Action,
			},
		}
		output, err := executeToolCall(tc)
		if err != nil {
			fmt.Printf("    Error: %v\n", err)
		} else {
			fmt.Printf("    Result: %s\n", output)
		}
		fmt.Println()
	}
}

// mockExecute 模拟执行
func mockExecute(prompt string) {
	file, err := os.Open(*testFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var steps []TaskStep
	scanner := bufio.NewScanner(file)
	stepID := 1

	for scanner.Scan() {
		line := scanner.Text()

		var tool, params string

		switch {
		case contains(line, "[GET]"):
			tool = "http-get"
			params = extractURL(line)
		case contains(line, "[POST]"):
			tool = "http-post"
			params = extractURL(line)
		case contains(line, "@wait:"):
			tool = "sleep"
			params = extractWait(line)
		case contains(line, "@ask:"):
			tool = "ask"
			params = extractQuestion(line)
		default:
			continue
		}

		if tool != "" {
			steps = append(steps, TaskStep{
				ID:      fmt.Sprintf("step-%d", stepID),
				Action:  line,
				Tool:   tool,
				Params: params,
				Status: "pending",
			})
			stepID++
		}
	}

	fmt.Printf("解析到 %d 个步骤:\n\n", len(steps))

	for i, step := range steps {
		fmt.Printf("[%d] %s\n", i+1, step.Tool)
		fmt.Printf("    %s\n\n", step.Params)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s[:len(substr)] == substr || bytes.Contains([]byte(s), []byte(substr)))
}

func extractURL(line string) string {
	start := 0
	for i := 0; i < len(line)-4; i++ {
		if line[i:i+5] == "[GET]" || line[i:i+5] == "[POST]" {
			start = i + 5
			break
		}
	}
	for i := start; i < len(line); i++ {
		if i+4 <= len(line) && line[i:i+4] == "http" {
			start = i
			break
		}
	}
	for i := start; i < len(line); i++ {
		if line[i] == ' ' || line[i] == '\n' {
			return line[start:i]
		}
	}
	return line[start:]
}

func extractWait(line string) string {
	return "10min"
}

func extractQuestion(line string) string {
	return "询问用户是否执行"
}