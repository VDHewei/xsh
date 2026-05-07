package main

import (
	"flag"
	"fmt"
	"os"

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
)

func main() {
	flag.Parse()

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
