package llm

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	ortgenai "github.com/getcharzp/onnxruntime-genai_purego"
)

// Model LLM 模型
type Model struct {
	Name       string
	Path       string
	engine     *ortgenai.Engine
	model      *ortgenai.Model
	isLoaded   bool
	libPath    string // onnxruntime-genai shared library path
}

// Config 模型配置
type Config struct {
	ModelPath     string
	ModelType     string  // "llama", "qwen", "baichuan" 等
	MaxLength     int
	Temperature   float32
	TopP          float32
	TopK          int
	RepeatPenalty float32
	NumThreads    int32
}

// NewConfig 创建默认配置
func NewConfig() *Config {
	return &Config{
		ModelPath:     "models",
		ModelType:     "qwen",
		MaxLength:     2048,
		Temperature:   0.7,
		TopP:          0.9,
		TopK:          40,
		RepeatPenalty: 1.1,
		NumThreads:    4,
	}
}

// NewModel 创建新模型
func NewModel(name string) *Model {
	return &Model{
		Name:     name,
		Path:     "",
		engine:   nil,
		model:    nil,
		isLoaded: false,
		libPath:  "",
	}
}

// DefaultGenAILibraryPath 返回 onnxruntime-genai 动态库的默认路径
func DefaultGenAILibraryPath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join("lib", "onnxruntime-genai.dll")
	case "linux":
		if runtime.GOARCH == "arm64" {
			return filepath.Join("lib", "libonnxruntime-genai.so")
		}
		return filepath.Join("lib", "libonnxruntime-genai.so")
	case "darwin":
		return filepath.Join("lib", "libonnxruntime-genai.dylib")
	default:
		return filepath.Join("lib", "onnxruntime-genai")
	}
}

// DefaultOnnxRuntimeLibraryPath 返回 onnxruntime 动态库的默认路径
func DefaultOnnxRuntimeLibraryPath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join("lib", "onnxruntime.dll")
	case "linux":
		if runtime.GOARCH == "arm64" {
			return filepath.Join("lib", "libonnxruntime.so")
		}
		return filepath.Join("lib", "libonnxruntime.so")
	case "darwin":
		return filepath.Join("lib", "libonnxruntime.dylib")
	default:
		return filepath.Join("lib", "libonnxruntime.so")
	}
}

// ensureGenAILibrary 确保 GenAI 动态库存在
func ensureGenAILibrary() error {
	libPath := DefaultGenAILibraryPath()
	if _, err := os.Stat(libPath); err == nil {
		return nil
	}
	return DownloadOnnxRuntimeGenAILibrary()
}

// DownloadOnnxRuntimeGenAILibrary 下载 onnxruntime-genai 和 onnxruntime 动态库到 lib/ 目录
func DownloadOnnxRuntimeGenAILibrary() error {
	if err := os.MkdirAll("lib", 0755); err != nil {
		return fmt.Errorf("failed to create lib directory: %w", err)
	}

	// 下载 onnxruntime-genai
	if err := downloadGenAILib(); err != nil {
		return fmt.Errorf("failed to download onnxruntime-genai: %w", err)
	}

	// 下载 onnxruntime (GenAI 依赖)
	if err := downloadOnnxRuntimeLib(); err != nil {
		return fmt.Errorf("failed to download onnxruntime: %w", err)
	}

	return nil
}

func downloadGenAILib() error {
	var platform, arch string

	switch runtime.GOOS {
	case "windows":
		platform = "win"
		arch = "x64"
	case "linux":
		platform = "linux"
		if runtime.GOARCH == "arm64" {
			arch = "arm64"
		} else {
			arch = "x64"
		}
	case "darwin":
		platform = "osx"
		if runtime.GOARCH == "arm64" {
			arch = "arm64"
		} else {
			arch = "x64"
		}
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	version := "0.12.0"
	var ext string
	if runtime.GOOS == "linux" {
		ext = "tar.gz"
	} else {
		ext = "zip"
	}

	filename := fmt.Sprintf("onnxruntime-genai-%s-%s-%s.%s", version, platform, arch, ext)
	downloadURL := fmt.Sprintf("https://github.com/microsoft/onnxruntime-genai/releases/download/v%s/%s", version, filename)

	fmt.Printf("Downloading onnxruntime-genai from %s\n", downloadURL)

	return downloadAndExtractArchive(downloadURL, "lib", platform, ext)
}

func downloadOnnxRuntimeLib() error {
	var platform, arch string

	switch runtime.GOOS {
	case "windows":
		platform = "win"
		arch = "x64"
	case "linux":
		platform = "linux"
		if runtime.GOARCH == "arm64" {
			arch = "aarch64"
		} else {
			arch = "x64"
		}
	case "darwin":
		platform = "osx"
		if runtime.GOARCH == "arm64" {
			arch = "aarch64"
		} else {
			arch = "x64"
		}
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	ortVersion := "1.24.1"
	var ext string
	if runtime.GOOS == "linux" {
		ext = "tgz"
	} else {
		ext = "zip"
	}

	filename := fmt.Sprintf("onnxruntime-%s-%s-%s.%s", platform, arch, ortVersion, ext)
	downloadURL := fmt.Sprintf("https://github.com/microsoft/onnxruntime/releases/download/v%s/%s", ortVersion, filename)

	fmt.Printf("Downloading onnxruntime from %s\n", downloadURL)

	return downloadAndExtractArchive(downloadURL, "lib", platform, ext)
}

func downloadAndExtractArchive(downloadURL, destDir, platform, ext string) error {
	tmpDir, err := os.MkdirTemp("", "onnx-download")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, "archive."+ext)

	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	out, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("failed to create archive file: %w", err)
	}
	_, err = io.Copy(out, resp.Body)
	out.Close()
	if err != nil {
		return fmt.Errorf("failed to write archive file: %w", err)
	}

	if ext == "zip" {
		if err := unzip(archivePath, tmpDir); err != nil {
			return fmt.Errorf("failed to unzip: %w", err)
		}
	} else {
		// tar.gz 提取
		return fmt.Errorf("tar.gz extraction not yet implemented, please manually download from: %s", downloadURL)
	}

	// 查找并复制动态库文件到 lib/
	return copyLibsFromDir(tmpDir, destDir)
}

func copyLibsFromDir(srcDir, destDir string) error {
	// 遍历解压后的目录，查找 .dll/.so/.dylib 文件并复制到 lib/
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		name := strings.ToLower(info.Name())
		// 复制动态库文件
		if strings.HasSuffix(name, ".dll") || strings.HasSuffix(name, ".so") || strings.HasSuffix(name, ".dylib") {
			dstPath := filepath.Join(destDir, info.Name())
			if err := copyFile(path, dstPath); err != nil {
				fmt.Printf("Warning: failed to copy %s: %v\n", info.Name(), err)
			} else {
				fmt.Printf("Copied: %s -> %s\n", info.Name(), dstPath)
			}
		}
		return nil
	})
}

// Load 加载模型 (modelDir 是包含 genai_config.json 等文件的模型目录)
func (m *Model) Load(modelDir string) error {
	// 确保 GenAI 动态库存在
	if err := ensureGenAILibrary(); err != nil {
		return fmt.Errorf("failed to ensure genai library: %w", err)
	}

	// 检查模型目录是否存在
	if info, err := os.Stat(modelDir); err != nil || !info.IsDir() {
		// 也可能是文件路径，兼容旧逻辑
		if _, err := os.Stat(modelDir); os.IsNotExist(err) {
			return fmt.Errorf("model path not found: %s", modelDir)
		}
	}

	libPath := DefaultGenAILibraryPath()
	m.libPath = libPath

	// 创建 GenAI 引擎
	engine, err := ortgenai.NewEngine(libPath)
	if err != nil {
		return fmt.Errorf("failed to create genai engine: %w", err)
	}
	m.engine = engine

	// 加载模型
	model, err := engine.NewModel(modelDir)
	if err != nil {
		return fmt.Errorf("failed to load model: %w", err)
	}
	m.model = model

	m.Path = modelDir
	m.isLoaded = true

	return nil
}

// Unload 卸载模型
func (m *Model) Unload() {
	if m.model != nil {
		m.model.Close()
		m.model = nil
	}
	// 注意: ortgenai.Engine 没有 Close/Destroy 方法，
	// 底层库handle会在进程结束时清理
	m.engine = nil
	m.isLoaded = false
}

// IsLoaded 检查模型是否已加载
func (m *Model) IsLoaded() bool {
	return m.isLoaded
}

// Infer 执行推理（使用默认配置）
func (m *Model) Infer(prompt string) (string, error) {
	if !m.isLoaded || m.engine == nil || m.model == nil {
		return "", fmt.Errorf("model not loaded")
	}

	return m.generate(prompt, GenerateOptions{
		MaxTokens:   2048,
		Temperature: 0.7,
		TopP:        0.9,
		DoSample:    true,
		StopOnEos:   true,
	})
}

// InferWithConfig 使用配置执行推理
func (m *Model) InferWithConfig(prompt string, cfg *Config) (string, error) {
	if !m.isLoaded || m.engine == nil || m.model == nil {
		return "", fmt.Errorf("model not loaded")
	}

	return m.generate(prompt, GenerateOptions{
		MaxTokens:   cfg.MaxLength,
		Temperature: float64(cfg.Temperature),
		TopP:        float64(cfg.TopP),
		DoSample:    cfg.Temperature > 0,
		StopOnEos:   true,
	})
}

// InferStream 流式推理，通过 callback 逐 token 输出
func (m *Model) InferStream(prompt string, opts GenerateOptions, onToken func(text string) error) error {
	if !m.isLoaded || m.engine == nil || m.model == nil {
		return fmt.Errorf("model not loaded")
	}

	return m.generateStream(prompt, opts, onToken)
}

// GenerateOptions 生成选项
type GenerateOptions = ortgenai.GenerateOptions

// generate 内部推理实现（复用已加载的 model）
func (m *Model) generate(prompt string, opts GenerateOptions) (string, error) {
	if opts.MaxTokens <= 0 {
		opts.MaxTokens = 256
	}

	// 创建 tokenizer
	tokenizer, err := m.model.NewTokenizer()
	if err != nil {
		return "", fmt.Errorf("failed to create tokenizer: %w", err)
	}
	defer tokenizer.Close()

	// 创建生成参数
	params, err := m.model.NewGeneratorParams()
	if err != nil {
		return "", fmt.Errorf("failed to create generator params: %w", err)
	}
	defer params.Close()

	_ = params.SetSearchNumber("max_length", float64(opts.MaxTokens))
	if opts.Temperature > 0 {
		_ = params.SetSearchNumber("temperature", opts.Temperature)
	}
	if opts.TopP > 0 {
		_ = params.SetSearchNumber("top_p", opts.TopP)
	}
	_ = params.SetSearchBool("do_sample", opts.DoSample)

	// 编码 prompt
	sequences, err := tokenizer.NewSequences()
	if err != nil {
		return "", fmt.Errorf("failed to create sequences: %w", err)
	}
	defer sequences.Close()

	if err := tokenizer.Encode(prompt, sequences); err != nil {
		return "", fmt.Errorf("failed to encode prompt: %w", err)
	}

	// 创建生成器
	generator, err := m.model.NewGenerator(params)
	if err != nil {
		return "", fmt.Errorf("failed to create generator: %w", err)
	}
	defer generator.Close()

	if err := generator.AppendTokenSequences(sequences); err != nil {
		return "", fmt.Errorf("failed to append token sequences: %w", err)
	}

	// 创建 tokenizer stream 用于解码
	stream, err := tokenizer.NewTokenizerStream()
	if err != nil {
		return "", fmt.Errorf("failed to create tokenizer stream: %w", err)
	}
	defer stream.Close()

	// 获取 EOS token IDs
	var eosSet map[int32]struct{}
	if opts.StopOnEos {
		eosIDs, err := tokenizer.GetEosTokenIds()
		if err == nil && len(eosIDs) > 0 {
			eosSet = make(map[int32]struct{}, len(eosIDs))
			for _, id := range eosIDs {
				eosSet[id] = struct{}{}
			}
		}
	}

	// 自回归生成
	var result strings.Builder
	for i := 0; i < opts.MaxTokens && !generator.IsDone(); i++ {
		if err := generator.GenerateNextToken(); err != nil {
			break
		}
		tokens, err := generator.GetNextTokens()
		if err != nil {
			break
		}
		for _, tok := range tokens {
			if eosSet != nil {
				if _, ok := eosSet[tok]; ok {
					return result.String(), nil
				}
			}
			chunk, err := stream.DecodeToken(tok)
			if err != nil {
				continue
			}
			result.WriteString(chunk)
		}
	}

	return result.String(), nil
}

// generateStream 内部流式推理实现
func (m *Model) generateStream(prompt string, opts GenerateOptions, onToken func(text string) error) error {
	if opts.MaxTokens <= 0 {
		opts.MaxTokens = 256
	}

	tokenizer, err := m.model.NewTokenizer()
	if err != nil {
		return fmt.Errorf("failed to create tokenizer: %w", err)
	}
	defer tokenizer.Close()

	params, err := m.model.NewGeneratorParams()
	if err != nil {
		return fmt.Errorf("failed to create generator params: %w", err)
	}
	defer params.Close()

	_ = params.SetSearchNumber("max_length", float64(opts.MaxTokens))
	if opts.Temperature > 0 {
		_ = params.SetSearchNumber("temperature", opts.Temperature)
	}
	if opts.TopP > 0 {
		_ = params.SetSearchNumber("top_p", opts.TopP)
	}
	_ = params.SetSearchBool("do_sample", opts.DoSample)

	sequences, err := tokenizer.NewSequences()
	if err != nil {
		return fmt.Errorf("failed to create sequences: %w", err)
	}
	defer sequences.Close()

	if err := tokenizer.Encode(prompt, sequences); err != nil {
		return fmt.Errorf("failed to encode prompt: %w", err)
	}

	generator, err := m.model.NewGenerator(params)
	if err != nil {
		return fmt.Errorf("failed to create generator: %w", err)
	}
	defer generator.Close()

	if err := generator.AppendTokenSequences(sequences); err != nil {
		return fmt.Errorf("failed to append token sequences: %w", err)
	}

	stream, err := tokenizer.NewTokenizerStream()
	if err != nil {
		return fmt.Errorf("failed to create tokenizer stream: %w", err)
	}
	defer stream.Close()

	var eosSet map[int32]struct{}
	if opts.StopOnEos {
		eosIDs, err := tokenizer.GetEosTokenIds()
		if err == nil && len(eosIDs) > 0 {
			eosSet = make(map[int32]struct{}, len(eosIDs))
			for _, id := range eosIDs {
				eosSet[id] = struct{}{}
			}
		}
	}

	for i := 0; i < opts.MaxTokens && !generator.IsDone(); i++ {
		if err := generator.GenerateNextToken(); err != nil {
			break
		}
		tokens, err := generator.GetNextTokens()
		if err != nil {
			break
		}
		for _, tok := range tokens {
			if eosSet != nil {
				if _, ok := eosSet[tok]; ok {
					return nil
				}
			}
			chunk, err := stream.DecodeToken(tok)
			if err != nil {
				continue
			}
			if err := onToken(chunk); err != nil {
				return err
			}
		}
	}

	return nil
}

// Generate 使用高层 API 进行推理（每次调用重新加载模型，适合单次调用）
func Generate(modelDir, prompt string, opts GenerateOptions) (string, error) {
	libPath := DefaultGenAILibraryPath()
	engine, err := ortgenai.NewEngine(libPath)
	if err != nil {
		return "", fmt.Errorf("failed to create genai engine: %w", err)
	}

	return engine.Generate(modelDir, prompt, opts)
}

// GenerateStream 使用高层 API 流式推理
func GenerateStream(modelDir, prompt string, opts GenerateOptions, onToken func(text string, token int32) error) error {
	libPath := DefaultGenAILibraryPath()
	engine, err := ortgenai.NewEngine(libPath)
	if err != nil {
		return fmt.Errorf("failed to create genai engine: %w", err)
	}

	return engine.GenerateStream(modelDir, prompt, opts, onToken)
}

// GetModelPath 获取模型路径
func GetModelPath(modelDir string, modelName string) string {
	return filepath.Join(modelDir, modelName)
}

// ListModels 列出可用模型
func ListModels(modelDir string) ([]string, error) {
	var models []string

	entries, err := os.ReadDir(modelDir)
	if err != nil {
		if os.IsNotExist(err) {
			return models, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// 检查目录是否包含 genai_config.json 或 model.onnx
			configPath := filepath.Join(modelDir, entry.Name(), "genai_config.json")
			onnxPath := filepath.Join(modelDir, entry.Name(), "model.onnx")
			if _, err := os.Stat(configPath); err == nil {
				models = append(models, entry.Name())
			} else if _, err := os.Stat(onnxPath); err == nil {
				models = append(models, entry.Name())
			}
		}
	}

	return models, nil
}

// MockInfer 模拟推理（用于测试，无模型时的后备方案）
func MockInfer(prompt string) string {
	return fmt.Sprintf("[Mock LLM Response]\nProcessed: %s\n\nThis is a mock response for testing purposes.", prompt)
}

// --- 工具函数 ---

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func unzip(src, dst string) error {
	archive, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer archive.Close()

	for _, file := range archive.File {
		rc, err := file.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		path := filepath.Join(dst, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, 0755)
		} else {
			os.MkdirAll(filepath.Dir(path), 0755)
			outFile, err := os.Create(path)
			if err != nil {
				return err
			}
			_, err = io.Copy(outFile, rc)
			outFile.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
