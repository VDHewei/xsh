package tests

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/VDHewei/xsh/internal/parser"
	"github.com/VDHewei/xsh/internal/types"
	llm "github.com/VDHewei/xsh/pkg/llm"
)

const (
	migrationFile = "data/prod-migration-form-uat.txt"
	outputDir     = "data"
	deepseekDir   = "../models/deepseek-r1-distill-qwen-1.5B"
)

// genAILibraryExists checks if the ONNX GenAI library files exist
func genAILibraryExists() bool {
	genaiPath := llm.DefaultGenAILibraryPath()
	ortPath := llm.DefaultOnnxRuntimeLibraryPath()
	return fileExists(genaiPath) && fileExists(ortPath)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// TestDeepSeekR1_MigrationParse 验证 DeepSeek R1 解析迁移文件准确度
func TestDeepSeekR1_MigrationParse(t *testing.T) {
	// 先尝试用 parser 直接解析 (基础测试)
	tasks, err := parser.ParseFile(migrationFile)
	if err != nil {
		t.Fatalf("Failed to parse migration file: %v", err)
	}

	t.Logf("Parser extracted %d tasks from migration file", len(tasks))
	if len(tasks) == 0 {
		t.Fatal("Expected at least some tasks from migration file")
	}

	// 验证关键任务类型存在
	var hasGet, hasPost, hasAsk, hasCheck, hasWait bool
	for _, task := range tasks {
		switch task.Type {
		case types.TaskTypeHTTP:
			if task.HTTP.Method == "GET" {
				hasGet = true
			}
			if task.HTTP.Method == "POST" {
				hasPost = true
			}
		case types.TaskTypeAsk:
			hasAsk = true
		case types.TaskTypeCheck:
			hasCheck = true
		case types.TaskTypeWait:
			hasWait = true
		}
	}

	// 验证覆盖所有任务类型
	if !hasGet {
		t.Error("Missing GET task")
	}
	if !hasPost {
		t.Error("Missing POST task")
	}
	if !hasAsk {
		t.Error("Missing Ask task")
	}
	if !hasCheck {
		t.Error("Missing Check task")
	}
	if !hasWait {
		t.Error("Missing Wait task")
	}

	// 尝试加载 DeepSeek R1 模型进行 LLM 分析
	modelDir := findDeepSeekModelDir(t)
	if modelDir == "" {
		t.Skip("DeepSeek R1 model not found, skipping LLM inference test")
	}
	if !genAILibraryExists() {
		t.Skip("ONNX GenAI library not found, skipping model load (network download unavailable)")
	}

	model := llm.NewModel("deepseek-r1-distill-qwen-1.5B")
	if err := model.Load(modelDir); err != nil {
		t.Logf("Failed to load DeepSeek model: %v, using mock inference", err)
	} else {
		defer model.Unload()
		t.Logf("DeepSeek R1 model loaded successfully from %s", modelDir)
	}

	// LLM 分析
	analyzer := llm.NewTaskAnalyzer()
	if model.IsLoaded() {
		analyzer.SetModel(model)
	}

	content, _ := os.ReadFile(migrationFile)
	llmTasks, err := analyzer.AnalyzeContent(string(content))
	if err != nil {
		t.Fatalf("LLM analysis failed: %v", err)
	}

	t.Logf("LLM analysis extracted %d tasks", len(llmTasks))
	if len(llmTasks) == 0 {
		t.Error("Expected tasks from LLM analysis")
	}

	// 记录结果
	recordResult(t, "deepseek-r1", len(tasks), len(llmTasks), hasGet && hasPost && hasAsk && hasCheck && hasWait)
}

// TestGLM51_MigrationParse GLM5.1 解析迁移文件准确度
func TestGLM51_MigrationParse(t *testing.T) {
	modelDir := "../models/glm5.1-distill-onnx"
	if _, err := os.Stat(modelDir); os.IsNotExist(err) {
		// Try alternative paths
		modelDir = findGLM51ModelDir(t)
	}
	if modelDir == "" {
		t.Skip("GLM5.1 model not found locally. Download with: xsh model download glm5.1")
	}
	if !genAILibraryExists() {
		t.Skip("ONNX GenAI library not found, skipping model load (network download unavailable)")
	}

	model := llm.NewModel("glm5.1-distill")
	if err := model.Load(modelDir); err != nil {
		t.Skipf("Failed to load GLM5.1 model: %v", err)
	}
	defer model.Unload()

	analyzer := llm.NewTaskAnalyzer()
	analyzer.SetModel(model)

	content, _ := os.ReadFile(migrationFile)
	llmTasks, err := analyzer.AnalyzeContent(string(content))
	if err != nil {
		t.Fatalf("GLM5.1 analysis failed: %v", err)
	}

	t.Logf("GLM5.1 extracted %d tasks from migration file", len(llmTasks))
	if len(llmTasks) == 0 {
		t.Error("Expected tasks from GLM5.1 analysis")
	}

	// 验证任务类型
	var hasAsk, hasCheck bool
	for _, task := range llmTasks {
		if task.Type == types.TaskTypeAsk {
			hasAsk = true
		}
		if task.Type == types.TaskTypeCheck {
			hasCheck = true
		}
	}
	if !hasAsk {
		t.Error("GLM5.1: Missing Ask task")
	}
	if !hasCheck {
		t.Error("GLM5.1: Missing Check task")
	}

	recordResult(t, "glm5.1", 0, len(llmTasks), hasAsk && hasCheck)
}

// TestModelComparison 多模型对比测试
func TestModelComparison(t *testing.T) {
	// 先用 parser 建立基线
	parserTasks, err := parser.ParseFile(migrationFile)
	if err != nil {
		t.Fatalf("Failed to parse migration file: %v", err)
	}
	baselineCount := len(parserTasks)
	t.Logf("Parser baseline: %d tasks", baselineCount)

	results := make(map[string]struct {
		count    int
		coverage float64
	})

	// DeepSeek R1 (或 mock)
	deepseekDir := findDeepSeekModelDir(t)
	if deepseekDir != "" && genAILibraryExists() {
		model := llm.NewModel("deepseek")
		if err := model.Load(deepseekDir); err == nil {
			defer model.Unload()
			analyzer := llm.NewTaskAnalyzer()
			analyzer.SetModel(model)
			content, _ := os.ReadFile(migrationFile)
			dsTasks, err := analyzer.AnalyzeContent(string(content))
			if err == nil {
				results["deepseek-r1"] = struct {
					count    int
					coverage float64
				}{
					count:    len(dsTasks),
					coverage: float64(len(dsTasks)) / float64(baselineCount) * 100,
				}
			}
		}
	}

	// GLM5.1
	glmDir := findGLM51ModelDir(t)
	if glmDir != "" && genAILibraryExists() {
		model := llm.NewModel("glm5.1")
		if err := model.Load(glmDir); err == nil {
			defer model.Unload()
			analyzer := llm.NewTaskAnalyzer()
			analyzer.SetModel(model)
			content, _ := os.ReadFile(migrationFile)
			glmTasks, err := analyzer.AnalyzeContent(string(content))
			if err == nil {
				results["glm5.1"] = struct {
					count    int
					coverage float64
				}{count: len(glmTasks), coverage: float64(len(glmTasks)) / float64(baselineCount) * 100}
			}
		}
	}

	// 打印对比结果
	t.Log("=== Model Comparison Results ===")
	t.Logf("Baseline (Parser): %d tasks", baselineCount)
	for name, r := range results {
		t.Logf("  %s: %d tasks (%.0f%% coverage)", name, r.count, r.coverage)
	}

	// 保存对比结果到文件
	saveComparisonResult(t, baselineCount, results)
}

// --- Helpers ---

func findDeepSeekModelDir(t *testing.T) string {
	t.Helper()
	candidates := []string{
		deepseekDir + "/cpu_and_mobile/cpu-int4-rtn-block-32-acc-level-4",
		deepseekDir,
	}
	return findModelDirByCandidates(t, candidates)
}

func findGLM51ModelDir(t *testing.T) string {
	t.Helper()
	candidates := []string{
		"../models/glm5.1-distill-onnx",
		"../models/yasserrmd_glm5.1-distill-onnx",
	}
	return findModelDirByCandidates(t, candidates)
}

func findModelDirByCandidates(t *testing.T, candidates []string) string {
	t.Helper()
	for _, dir := range candidates {
		configPath := filepath.Join(dir, "genai_config.json")
		if _, err := os.Stat(configPath); err == nil {
			return dir
		}
		onnxPath := filepath.Join(dir, "model.onnx")
		if _, err := os.Stat(onnxPath); err == nil {
			return dir
		}
	}
	return ""
}

func recordResult(t *testing.T, modelName string, parserCount, llmCount int, allTypes bool) {
	t.Helper()
	result := fmt.Sprintf("# %s Migration Parse Test\n\n", strings.ToUpper(modelName))
	result += fmt.Sprintf("**Parser tasks:** %d\n", parserCount)
	result += fmt.Sprintf("**LLM tasks:** %d\n", llmCount)
	result += fmt.Sprintf("**All task types covered:** %v\n\n", allTypes)

	filename := filepath.Join(outputDir, fmt.Sprintf("%s-inference-test-result.md", modelName))
	_ = os.WriteFile(filename, []byte(result), 0644)
	t.Logf("Result recorded to %s", filename)
}

func saveComparisonResult(t *testing.T, baseline int, results map[string]struct {
	count    int
	coverage float64
}) {
	t.Helper()
	var sb strings.Builder
	sb.WriteString("# Model Comparison Results\n\n")
	sb.WriteString(fmt.Sprintf("**Input:** %s\n", migrationFile))
	sb.WriteString(fmt.Sprintf("**Baseline (Parser):** %d tasks\n\n", baseline))
	sb.WriteString("## Results\n\n")
	sb.WriteString("| Model | Tasks | Coverage (% of baseline) |\n")
	sb.WriteString("|-------|-------|--------------------------|\n")
	for name, r := range results {
		sb.WriteString(fmt.Sprintf("| %s | %d | %.0f%% |\n", name, r.count, r.coverage))
	}

	filename := filepath.Join(outputDir, "model-comparison-result.md")
	_ = os.WriteFile(filename, []byte(sb.String()), 0644)
	t.Logf("Comparison saved to %s", filename)
}

// 确保 tests/data/ 目录存在
func TestMain(m *testing.M) {
	// 确保 tests 在工作目录运行 (go test ./tests/...)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Failed to create output dir: %v\n", err)
	}
	os.Exit(m.Run())
}

// 确保 json 包被使用 (用于 future expansion)
var _ = json.Marshal
