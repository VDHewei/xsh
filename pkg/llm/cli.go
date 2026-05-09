package llm

import (
	"fmt"
	"os"
	"strings"
)

// ModelDownload 下载模型（通过短名称或完整 RepoID）
func ModelDownload(name string, downloadDir string) error {
	repoID := resolveModelRepo(name)
	cfg := NewDownloadConfig()
	cfg.CacheDir = downloadDir

	fmt.Printf("Downloading model: %s\n", repoID)
	fmt.Printf("Target directory: %s\n", downloadDir)

	result, err := DownloadFromHuggingFace(repoID, cfg)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	fmt.Printf("\nDownload complete:\n")
	fmt.Printf("  Name:   %s\n", result.Name)
	fmt.Printf("  Path:   %s\n", result.Path)
	fmt.Printf("  RepoID: %s\n", result.RepoID)
	if result.Size > 0 {
		fmt.Printf("  Size:   %d bytes\n", result.Size)
	}

	return nil
}

func HandleModelSearch(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: xsh model search <query>")
	}

	query := strings.Join(args, " ")
	cfg := NewDownloadConfig()
	models, err := SearchModels(query, cfg)
	if err != nil {
		return fmt.Errorf("search error: %w", err)
	}

	fmt.Println("Search results:")
	for _, model := range models {
		fmt.Printf("  - %s\n", model)
	}
	return nil
}

func HandleModelList(args []string) error {
	modelDir := "models"
	for i, arg := range args {
		if arg == "-d" || arg == "--dir" {
			if i+1 < len(args) {
				modelDir = args[i+1]
			}
		}
	}
	models, err := ListModelsWithCandidates(modelDir)
	if err != nil {
		return fmt.Errorf("list error: %w", err)
	}

	if len(models) == 0 {
		fmt.Println("No models found")
		return nil
	}

	fmt.Println("Available models:")
	for _, model := range models {
		marker := ""
		if model.Candidate != nil {
			if model.Installed {
				marker = " [candidate]"
			} else {
				marker = " [not installed]"
			}
			if model.Candidate.Default {
				marker = " [candidate, default]"
			}
		}
		fmt.Printf("  - %s%s\n", model.Name, marker)
	}
	return nil
}

func HandleModelSelect(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: xsh model select <model-name> [--dir <dir>]")
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

	modelPath := GetModelPath(modelDir, modelName)

	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		repoID := ResolveCandidateName(modelName)
		if repoID != "" {
			return fmt.Errorf("model '%s' (%s) not installed. Run: xsh model download %s", modelName, repoID, modelName)
		}
		return fmt.Errorf("model not found: %s", modelPath)
	}

	model := NewModel(modelName)
	if err := model.Load(modelPath); err != nil {
		return fmt.Errorf("load error: %w", err)
	}

	fmt.Printf("Selected model: %s\n", modelName)
	fmt.Printf("Model path: %s\n", modelPath)

	cfg := NewConfig()
	cfg.ModelPath = modelPath
	_ = cfg
	return nil
}

func HandleModelDownload(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: xsh model download <repo-id|model-name> [--dir <dir>]")
	}

	modelName := args[0]
	downloadDir := "models"

	for i, arg := range args {
		if arg == "-d" || arg == "--dir" {
			if i+1 < len(args) {
				downloadDir = args[i+1]
			}
		}
	}

	return ModelDownload(modelName, downloadDir)
}

func printModelUsage() {
	fmt.Println("Usage: xsh model <command>")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  search <query>              Search models on HuggingFace")
	fmt.Println("  list [-d <dir>]            List local models")
	fmt.Println("  select <name>              Select a model to use")
	fmt.Println("  download <repo|name>       Download a model from HuggingFace")
	fmt.Println("                             Use short name (deepseek, glm5.1) or full RepoID")
}
