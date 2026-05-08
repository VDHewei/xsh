package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/VDHewei/xsh/internal/executor"
	"github.com/VDHewei/xsh/internal/parser"
	"github.com/VDHewei/xsh/internal/tui"
	llm "github.com/VDHewei/xsh/pkg/llm"
)

var (
	inputFile  = flag.String("i", "", "Input task file (txt/md)")
	outputFile = flag.String("o", "", "Output result file")
	llmModel   = flag.String("m", "", "LLM model directory path or name")
	llmPrompt  = flag.String("p", "", "LLM prompt for inference")
	testMode   = flag.Bool("test", false, "Run ONNX LLM test mode")
	streamMode = flag.Bool("stream", false, "Enable streaming output for LLM inference")
)

func main() {
	// model 子命令路由
	if len(os.Args) > 1 && os.Args[1] == "model" {
		if err := llm.ParseModelCommand(os.Args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

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
	fmt.Println("=== ONNX GenAI LLM 任务规划测试 ===")

	testFile := "tests/data/prod-migration-form-uat.txt"
	data, err := os.ReadFile(testFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read test file: %v\n", err)
		os.Exit(1)
	}
	content := string(data)

	fmt.Printf("输入: %s\n\n", testFile)

	// 尝试查找模型目录
	modelDir := findModelDir()
	if modelDir == "" {
		fmt.Println("未找到模型目录，使用 Mock 推理...")
		result := llm.MockInfer(content)
		fmt.Printf("推理结果:\n%s\n", result)
		return
	}

	fmt.Printf("模型目录: %s\n", modelDir)

	// 确保动态库存在
	if err := llm.DownloadOnnxRuntimeGenAILibrary(); err != nil {
		fmt.Printf("Warning: failed to ensure genai library: %v\n", err)
		fmt.Println("使用 Mock 推理...")
		result := llm.MockInfer(content)
		fmt.Printf("推理结果:\n%s\n", result)
		return
	}

	// 使用 GenAI 进行推理
	testWithGenAI(modelDir, content)
}

// findModelDir 查找模型目录
func findModelDir() string {
	// 候选模型目录路径
	candidates := []string{
		"models/deepseek-r1-distill-qwen-1.5B/cpu_and_mobile/cpu-int4-rtn-block-32-acc-level-4",
		"models/DeepSeek-R1-Distill-ONNX/deepseek-r1-distill-qwen-1.5B/cpu_and_mobile/cpu-int4-rtn-block-32-acc-level-4",
		"models/deepseek-r1-distill-qwen-1.5B",
	}

	for _, dir := range candidates {
		// 检查目录是否包含 genai_config.json 或 model.onnx
		configPath := dir + "/genai_config.json"
		onnxPath := dir + "/model.onnx"
		if _, err := os.Stat(configPath); err == nil {
			return dir
		}
		if _, err := os.Stat(onnxPath); err == nil {
			return dir
		}
	}

	// 搜索 models/ 下的子目录
	entries, err := os.ReadDir("models")
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				subEntries, err := os.ReadDir("models/" + entry.Name())
				if err == nil {
					for _, sub := range subEntries {
						if sub.IsDir() {
							dir := "models/" + entry.Name() + "/" + sub.Name()
							if _, err := os.Stat(dir + "/model.onnx"); err == nil {
								return dir
							}
						}
					}
				}
				// 顶层目录也检查
				dir := "models/" + entry.Name()
				if _, err := os.Stat(dir + "/model.onnx"); err == nil {
					return dir
				}
			}
		}
	}

	return ""
}

// testWithGenAI 使用 GenAI 测试
func testWithGenAI(modelDir, content string) {
	model := llm.NewModel("deepseek-r1-distill")
	defer model.Unload()

	fmt.Println("加载模型...")
	if err := model.Load(modelDir); err != nil {
		fmt.Printf("加载模型失败: %v\n使用 Mock 推理\n", err)
		result := llm.MockInfer(content)
		fmt.Printf("推理结果:\n%s\n", result)
		return
	}

	fmt.Println("模型加载成功!")

	// 构建任务分析提示
	prompt := buildTaskPrompt(content)

	fmt.Println("\n执行推理...")
	if *streamMode {
		fmt.Println("--- 流式输出 ---")
		err := model.InferStream(prompt, llm.GenerateOptions{
			MaxTokens:   2048,
			Temperature: 0.7,
			TopP:        0.9,
			DoSample:    true,
			StopOnEos:   true,
		}, func(text string) error {
			fmt.Print(text)
			return nil
		})
		if err != nil {
			fmt.Printf("\n推理错误: %v\n", err)
		}
		fmt.Println()
	} else {
		result, err := model.Infer(prompt)
		if err != nil {
			fmt.Printf("推理失败: %v\n", err)
			return
		}
		fmt.Printf("\n推理结果:\n%s\n", result)
	}
}

// buildTaskPrompt 构建任务分析提示
func buildTaskPrompt(taskContent string) string {
	return fmt.Sprintf(`你是一个任务规划和执行助手。请分析以下迁移步骤，提取出结构化任务列表。

迁移步骤:
%s

请按以下格式输出，每行一个任务：
[GET] <url> 用于 GET 请求
header: Xxxx=xxx
[POST] <url> 用于 POST 请求
header: XXX=xxx
body: {}
@ask: <描述> 用于需要用户确认的步骤
@wait: <时长> 用于等待步骤 (如 @wait: 10min)
@check: <描述> 用于验证步骤

只输出任务列表，不要其他内容。`, taskContent)
}

// runLLMMode 运行 LLM 模式
func runLLMMode(modelName string, prompt string, inputFile string) {
	// 解析模型路径：可能是名称或路径
	modelDir := modelName
	if modelName != "" && !strings.Contains(modelName, "/") && !strings.Contains(modelName, "\\") {
		// 看起来是模型名称，在 models/ 下查找
		modelDir = llm.GetModelPath("models", modelName)
	}

	// 如果指定了模型路径，加载模型
	var model *llm.Model
	if modelDir != "" {
		model = llm.NewModel(modelName)
		if err := model.Load(modelDir); err != nil {
			fmt.Fprintf(os.Stderr, "Load model warning: %v, using mock inference\n", err)
			model = nil
		} else {
			defer model.Unload()
		}
	}

	// 使用模型推理
	if prompt != "" {
		if model != nil && model.IsLoaded() {
			if *streamMode {
				err := model.InferStream(prompt, llm.GenerateOptions{
					MaxTokens:   2048,
					Temperature: 0.7,
					TopP:        0.9,
					DoSample:    true,
					StopOnEos:   true,
				}, func(text string) error {
					fmt.Print(text)
					return nil
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "\nInference error: %v\n", err)
					os.Exit(1)
				}
				fmt.Println()
			} else {
				result, err := model.Infer(prompt)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Inference error: %v\n", err)
					os.Exit(1)
				}
				fmt.Println(result)
			}
		} else {
			fmt.Println(llm.MockInfer(prompt))
		}
		return
	}

	// 如果有输入文件，分析文件内容
	if inputFile != "" {
		analyzer := llm.NewTaskAnalyzer()
		if model != nil {
			analyzer.SetModel(model)
		}

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
