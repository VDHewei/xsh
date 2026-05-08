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

// runTestMode 运行测试模式 (支持多模型对比)
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

	// 确保动态库存在
	_ = llm.DownloadOnnxRuntimeGenAILibrary()

	var results []modelTestResult

	// 测试 DeepSeek R1
	fmt.Println("--- DeepSeek R1 ---")
	dsDir := findModelDir()
	if dsDir != "" {
		results = append(results, testModelInference(dsDir, "deepseek-r1-distill", content))
	} else {
		fmt.Println("DeepSeek R1 模型未找到")
	}

	fmt.Println()

	// 测试 GLM5.1
	fmt.Println("--- GLM5.1 ---")
	glmDir := findGLM51Dir()
	if glmDir != "" {
		results = append(results, testModelInference(glmDir, "glm5.1-distill", content))
	} else {
		fmt.Println("GLM5.1 模型未找到, 尝试使用 Mock")
		analyzer := llm.NewTaskAnalyzer()
		tasks, err := analyzer.AnalyzeContent(content)
		results = append(results, modelTestResult{"glm5.1-distill (mock)", len(tasks), err})
	}

	// 打印对比结果
	fmt.Println("\n=== 对比结果 ===")
	for _, r := range results {
		status := "OK"
		if r.err != nil {
			status = fmt.Sprintf("ERROR: %v", r.err)
		}
		fmt.Printf("  %-30s %3d tasks  %s\n", r.modelName, r.taskCount, status)
	}
}

type modelTestResult struct {
	modelName string
	taskCount int
	err       error
}

// testModelInference 测试单个模型推理
func testModelInference(modelDir, modelName, content string) modelTestResult {
	fmt.Printf("模型目录: %s\n", modelDir)

	model := llm.NewModel(modelName)
	if err := model.Load(modelDir); err != nil {
		fmt.Printf("加载模型失败: %v\n", err)
		analyzer := llm.NewTaskAnalyzer()
		tasks, _ := analyzer.AnalyzeContent(content)
		return modelTestResult{modelName + " (mock)", len(tasks), nil}
	}
	defer model.Unload()

	fmt.Println("模型加载成功!")

	analyzer := llm.NewTaskAnalyzer()
	analyzer.SetModel(model)

	fmt.Println("执行任务分析...")
	tasks, err := analyzer.AnalyzeContent(content)
	if err != nil {
		fmt.Printf("分析失败: %v\n", err)
		return modelTestResult{modelName, 0, err}
	}

	fmt.Printf("提取任务: %d 个\n", len(tasks))
	return modelTestResult{modelName, len(tasks), nil}
}

// findGLM51Dir 查找 GLM5.1 模型目录
func findGLM51Dir() string {
	candidates := []string{
		"models/glm5.1-distill-onnx",
		"models/yasserrmd_glm5.1-distill-onnx",
		"models/glm5.1-distill-onnx/cpu_and_mobile/cpu-int4-rtn-block-32-acc-level-4",
	}
	for _, dir := range candidates {
		if _, err := os.Stat(dir + "/genai_config.json"); err == nil {
			return dir
		}
		if _, err := os.Stat(dir + "/model.onnx"); err == nil {
			return dir
		}
	}
	return ""
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
	prompt := llm.BuildTaskPrompt(content)

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
