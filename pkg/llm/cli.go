package llm

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// CLI Commands
var (
	modelSearchCmd = flag.NewFlagSet("model search", flag.ExitOnError)
	modelListCmd  = flag.NewFlagSet("model list", flag.ExitOnError)
	modelSelectCmd = flag.NewFlagSet("model select", flag.ExitOnError)
)

// ModelSearch 搜索模型
func ModelSearch(query string) {
	cfg := NewDownloadConfig()
	models, err := SearchModels(query, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Search error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Search results:")
	for _, model := range models {
		fmt.Printf("  - %s\n", model)
	}
}

// ModelList 列出本地模型
func ModelList(modelDir string) {
	models, err := ListModels(modelDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "List error: %v\n", err)
		os.Exit(1)
	}

	if len(models) == 0 {
		fmt.Println("No models found")
		return
	}

	fmt.Println("Available models:")
	for _, model := range models {
		fmt.Printf("  - %s\n", model)
	}
}

// ModelSelect 选择模型
func ModelSelect(modelDir, modelName string) {
	modelPath := GetModelPath(modelDir, modelName)

	// 检查模型是否存在
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Model not found: %s\n", modelPath)
		os.Exit(1)
	}

	// 加载模型
	model := NewModel(modelName)
	if err := model.Load(modelPath); err != nil {
		fmt.Fprintf(os.Stderr, "Load error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Selected model: %s\n", modelName)
	fmt.Printf("Model path: %s\n", modelPath)

	// 保存选中的模型配置供后续使用
	cfg := NewConfig()
	cfg.ModelPath = modelPath
	_ = cfg
}

// ParseModelCommand 解析模型命令
func ParseModelCommand(args []string) {
	if len(args) < 2 {
		printModelUsage()
		return
	}

	cmd := args[1]
	subArgs := args[2:]

	switch cmd {
	case "search":
		handleModelSearch(subArgs)
	case "list":
		handleModelList(subArgs)
	case "select":
		handleModelSelect(subArgs)
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printModelUsage()
	}
}

func handleModelSearch(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: xsh model search <query>")
		os.Exit(1)
	}

	query := strings.Join(args, " ")
	ModelSearch(query)
}

func handleModelList(args []string) {
	modelDir := "models"
	for i, arg := range args {
		if arg == "-d" || arg == "--dir" {
			if i+1 < len(args) {
				modelDir = args[i+1]
			}
		}
	}
	ModelList(modelDir)
}

func handleModelSelect(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: xsh model select <model-name>")
		os.Exit(1)
	}

	modelName := args[0]
	modelDir := "models"

	for i, arg := range args {
		if arg == "-d" || arg == "--dir" {
			if i+1 < len(args) {
				modelDir = args[i+1]
			}
		}
	}

	ModelSelect(modelDir, modelName)
}

func printModelUsage() {
	fmt.Println("Usage: xsh model <command>")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  search <query>      Search models on HuggingFace")
	fmt.Println("  list [-d <dir>]    List local models")
	fmt.Println("  select <name>      Select a model to use")
}