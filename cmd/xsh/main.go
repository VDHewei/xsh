package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/VDHewei/xsh/internal/executor"
	"github.com/VDHewei/xsh/internal/parser"
	"github.com/VDHewei/xsh/internal/tui"
	llm "github.com/VDHewei/xsh/pkg/llm"
)

var (
	inputFile  = flag.String("i", "", "Input task file (txt/md)")
	outputFile = flag.String("o", "", "Output result file")
	llmModel  = flag.String("m", "", "LLM model to use")
	llmPrompt = flag.String("p", "", "LLM prompt for inference")
	testMode  = flag.Bool("test", false, "Run ONNX test mode")
)

func main() {
	flag.Parse()

	// 测试模式
	if *testMode {
		runTestMode()
		return
	}

	// LLM 模式
	if *llmModel != "" || *llmPrompt != "" {
		runLLMMode(*llmModel, *llmPrompt, *inputFile)
		return
	}

	// 无参数时启动 TUI 交互模式
	if *inputFile == "" {
		tui.RunInteractive()
		return
	}

	// CLI 模式：读取并执行任务
	tasks, err := parser.ParseFile(*inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing file: %v\n", err)
		os.Exit(1)
	}

	results := executor.ExecuteTasks(tasks)

	// 输出结果
	if *outputFile != "" {
		f, err := os.Create(*outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		for _, r := range results {
			fmt.Fprintln(f, r)
		}
	} else {
		for _, r := range results {
			fmt.Println(r)
		}
	}
}

// runTestMode 运行测试模式
func runTestMode() {
	fmt.Println("=== ONNX LLM 任务规划测试 ===")
	fmt.Println("输入: tests/data/prod-migration-form-uat.txt")
	fmt.Println()

	// 尝试下载模型
	modelPath := downloadModelWithPython()
	if modelPath != "" {
		fmt.Println("\n使用 ONNX 模型推理...")
		testWithONNXModel(modelPath)
	} else {
		fmt.Println("\n使用 Mock 推理...")
		task, err := os.ReadFile("tests/data/prod-migration-form-uat.txt")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read test file: %v\n", err)
			os.Exit(1)
		}
		result := llm.MockInfer(string(task))
		fmt.Printf("推理结果:\n%s\n", result)
	}
}

// downloadModelWithPython 使用 Python 下载模型
func downloadModelWithPython() string {
	fmt.Println("开始下载 DeepSeek-R1-Distill-ONNX 模型...")
	fmt.Println("这可能需要几分钟时间...")

	// 创建模型目录
	os.MkdirAll("models", 0755)

	// 使用 Python 脚本下载模型
	script := `
import os
from huggingface_hub import snapshot_download

model_id = "onnxruntime/DeepSeek-R1-Distill-ONNX"
local_dir = "models"

try:
    # 下载 1.5B CPU 版本
    path = snapshot_download(
        repo_id=model_id,
        allow_patterns=["deepseek-r1-distill-qwen-1.5B/cpu_and_mobile/*"],
        local_dir=local_dir,
        local_dir_use_symlinks=False
    )
    print(f"SUCCESS:{path}")
except Exception as e:
    print(f"ERROR:{e}")
`
	cmd := exec.Command("python", "-c", script)
	output, err := cmd.CombinedOutput()
	result := string(output)
	fmt.Println(result)

	if err != nil {
		fmt.Printf("下载失败: %v\n", err)
		return ""
	}

	// 解析输出
	if strings.Contains(result, "SUCCESS:") {
		path := strings.TrimPrefix(result, "SUCCESS:")
		path = strings.TrimSpace(path)
		return path
	}

	return ""
}

// testWithONNXModel 使用 ONNX 模型测试
func testWithONNXModel(modelPath string) {
	task, err := os.ReadFile("tests/data/prod-migration-form-uat.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read test file: %v\n", err)
		return
	}

	// 查找 ONNX 模型文件
	onnxPath := findONNXFile(modelPath)
	if onnxPath == "" {
		fmt.Println("未找到 ONNX 模型文件，使用 Mock 推理")
		result := llm.MockInfer(string(task))
		fmt.Printf("推理结果:\n%s\n", result)
		return
	}

	fmt.Printf("模型路径: %s\n", onnxPath)

	// 创建模型
	model := llm.NewModel("deepseek-r1-distill")
	defer model.Unload()

	// 加载模型
	fmt.Println("加载模型...")
	if err := model.Load(onnxPath); err != nil {
		fmt.Printf("加载模型失败: %v\n", err)
		fmt.Println("使用 Mock 推理")
		result := llm.MockInfer(string(task))
		fmt.Printf("推理结果:\n%s\n", result)
		return
	}

	fmt.Println("模型加载成功!")

	// 构建提示
	prompt := string(task)

	// 执行推理
	fmt.Println("执行推理...")
	result, err := model.Infer(prompt)
	if err != nil {
		fmt.Printf("推理失败: %v\n", err)
		return
	}

	fmt.Printf("\n推理结果:\n%s\n", result)
}

// findONNXFile 查找 ONNX 模型文件
func findONNXFile(dir string) string {
	var onnxFiles []string

	// 递归查找 .onnx 文件
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(path, ".onnx") {
			onnxFiles = append(onnxFiles, path)
		}
		return nil
	})

	if len(onnxFiles) > 0 {
		// 返回第一个找到的 ONNX 文件
		return onnxFiles[0]
	}
	return ""
}

// runLLMMode 运行 LLM 模式
func runLLMMode(modelName string, prompt string, inputFile string) {
	// 创建模型
	model := llm.NewModel(modelName)

	// 如果指定了模型路径，加载模型
	if modelName != "" {
		modelPath := llm.GetModelPath("models", modelName)
		if err := model.Load(modelPath); err != nil {
			fmt.Fprintf(os.Stderr, "Load model warning: %v, using mock inference\n", err)
			// 使用 mock 推理
		}
	}

	// 使用模型推理
	if prompt != "" {
		if model.IsLoaded() {
			result, err := model.Infer(prompt)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Inference error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(result)
		} else {
			// 使用 mock 推理
			fmt.Println(llm.MockInfer(prompt))
		}
		return
	}

	// 如果有输入文件，分析文件内容
	if inputFile != "" {
		analyzer := llm.NewTaskAnalyzer()
		analyzer.SetModel(model)

		tasks, err := analyzer.AnalyzeFile(inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Analyze error: %v\n", err)
			os.Exit(1)
		}

		// 执行并输出结果
		results := executor.ExecuteTasks(tasks)
		for _, r := range results {
			fmt.Println(r)
		}
		return
	}

	// 无参数时使用 mock 推理
	fmt.Println(llm.MockInfer("Hello from LLM mode"))
}
