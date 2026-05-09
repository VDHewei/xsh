package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/VDHewei/xsh/internal/executor"
	"github.com/VDHewei/xsh/internal/parser"
	"github.com/VDHewei/xsh/internal/tui"
	llm "github.com/VDHewei/xsh/pkg/llm"
	"github.com/spf13/cobra"
)

var (
	inputFile  string
	outputFile string
	llmModel   string
	llmPrompt  string
	streamMode bool
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "xsh",
	Short: "XSH - extensible shell task runner with LLM support",
	RunE: func(cmd *cobra.Command, args []string) error {
		if llmModel != "" || llmPrompt != "" {
			return runLLMMode(llmModel, llmPrompt, inputFile, streamMode)
		}

		if inputFile == "" {
			tui.RunInteractive()
			return nil
		}

		tasks, err := parser.ParseFile(inputFile)
		if err != nil {
			return fmt.Errorf("parsing file: %w", err)
		}

		results := executor.ExecuteTasks(tasks)

		if outputFile != "" {
			f, err := os.Create(outputFile)
			if err != nil {
				return fmt.Errorf("creating output file: %w", err)
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
		return nil
	},
}

// --- model 子命令 ---

var modelCmd = &cobra.Command{
	Use:   "model",
	Short: "Manage LLM models",
}

var modelSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search models on HuggingFace",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return llm.HandleModelSearch(args)
	},
}

var modelListCmd = &cobra.Command{
	Use:   "list",
	Short: "List local models",
	RunE: func(cmd *cobra.Command, args []string) error {
		modelDir, _ := cmd.Flags().GetString("dir")
		return llm.HandleModelList([]string{"-d", modelDir})
	},
}

var modelSelectCmd = &cobra.Command{
	Use:   "select <name>",
	Short: "Select a model to use",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		modelDir, _ := cmd.Flags().GetString("dir")
		return llm.HandleModelSelect(append(args, "--dir", modelDir))
	},
}

var modelDownloadCmd = &cobra.Command{
	Use:   "download <repo|name>",
	Short: "Download a model from HuggingFace",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		downloadDir, _ := cmd.Flags().GetString("dir")
		return llm.HandleModelDownload(append(args, "--dir", downloadDir))
	},
}

func init() {
	rootCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input task file (txt/md)")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output result file")
	rootCmd.Flags().StringVarP(&llmModel, "model", "m", "", "LLM model directory path or name")
	rootCmd.Flags().StringVarP(&llmPrompt, "prompt", "p", "", "LLM prompt for inference")
	rootCmd.Flags().BoolVar(&streamMode, "stream", false, "Enable streaming output for LLM inference")

	modelListCmd.Flags().StringP("dir", "d", "models", "Model directory")
	modelSelectCmd.Flags().String("dir", "models", "Model directory")
	modelDownloadCmd.Flags().StringP("dir", "d", "models", "Download target directory")

	modelCmd.AddCommand(modelSearchCmd, modelListCmd, modelSelectCmd, modelDownloadCmd)
	rootCmd.AddCommand(modelCmd)
}

// runLLMMode 运行 LLM 推理或任务分析模式
func runLLMMode(modelName string, prompt string, inputFile string, streamEnabled bool) error {
	modelDir := modelName
	if modelName != "" && !strings.Contains(modelName, "/") && !strings.Contains(modelName, "\\") {
		modelDir = llm.GetModelPath("models", modelName)
	}

	var model *llm.Model
	if modelDir != "" {
		model = llm.NewModel(modelName)
		if err := model.Load(modelDir); err != nil {
			return fmt.Errorf("load model %s: %w", modelName, err)
		}
		defer model.Unload()
	}

	// 直接推理模式
	if prompt != "" {
		if model == nil || !model.IsLoaded() {
			return fmt.Errorf("no model loaded for prompt inference")
		}
		if streamEnabled {
			if err := model.InferStream(prompt, llm.GenerateOptions{
				MaxTokens:   2048,
				Temperature: 0.7,
				TopP:        0.9,
				DoSample:    true,
				StopOnEos:   true,
			}, func(text string) error {
				fmt.Print(text)
				return nil
			}); err != nil {
				return fmt.Errorf("stream inference: %w", err)
			}
			fmt.Println()
		} else {
			result, err := model.Infer(prompt)
			if err != nil {
				return fmt.Errorf("inference: %w", err)
			}
			fmt.Println(result)
		}
		return nil
	}

	// 任务分析模式 (使用输入文件)
	if inputFile != "" {
		analyzer := llm.NewTaskAnalyzer()
		if model != nil {
			analyzer.SetModel(model)
		}
		tasks, err := analyzer.AnalyzeFile(inputFile)
		if err != nil {
			return fmt.Errorf("analyze file: %w", err)
		}
		results := executor.ExecuteTasks(tasks)
		for _, r := range results {
			fmt.Println(r)
		}
		return nil
	}

	return fmt.Errorf("-m/-p requires a model or -i input file")
}
