package llm

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	ort "github.com/getcharzp/onnxruntime_purego"
)

// Model LLM 模型
type Model struct {
	Name     string
	Path     string
	Engine   *ort.Engine
	Session  *ort.Session
	isLoaded bool
}

// Config 模型配置
type Config struct {
	ModelPath     string
	ModelType    string // "llama", "qwen", "baichuan" 等
	MaxLength    int
	Temperature float32
	TopP         float32
	TopK         int
	RepeatPenalty float32
	NumThreads   int32
}

// NewConfig 创建默认配置
func NewConfig() *Config {
	return &Config{
		ModelPath:     "models",
		ModelType:     "llama",
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
		Path:    "",
		Engine:  nil,
		Session: nil,
		isLoaded: false,
	}
}

// ensureLibrary 确保 onnxruntime 动态库存在
func ensureLibrary() error {
	// 检查动态库是否已存在
	libPath := ort.DefaultLibraryPath()
	if _, err := os.Stat(libPath); err == nil {
		return nil // 库已存在
	}

	// 尝试下载动态库
	return DownloadOnnxRuntimeLibrary()
}

// DownloadOnnxRuntimeLibrary 下载 onnxruntime 动态库到系统目录
func DownloadOnnxRuntimeLibrary() error {
	// 确定平台和架构
	var platform, arch string
	var libName string

	switch runtime.GOOS {
	case "windows":
		platform = "windows"
		arch = "x64"
		libName = "onnxruntime.dll"
	case "linux":
		platform = "linux"
		if runtime.GOARCH == "arm64" {
			arch = "aarch64"
			libName = "onnxruntime_arm64.so"
		} else {
			arch = "x64"
			libName = "onnxruntime_amd64.so"
		}
	case "darwin":
		platform = "darwin"
		if runtime.GOARCH == "arm64" {
			arch = "aarch64"
			libName = "onnxruntime_arm64.dylib"
		} else {
			arch = "x64"
			libName = "onnxruntime_amd64.dylib"
		}
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	// 下载链接 - 下载 zip 文件
	version := "v1.24.1"
	zipURL := fmt.Sprintf("https://github.com/microsoft/onnxruntime/releases/download/%s/onnxruntime-%s-%s-%s.zip",
		version, platform, arch, version)

	// 创建临时目录下载
	tmpDir, err := os.MkdirTemp("", "onnxruntime")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	zipPath := filepath.Join(tmpDir, "onnxruntime.zip")

	// 下载 zip 文件
	resp, err := http.Get(zipURL)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	// 保存 zip 文件
	out, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	_, err = io.Copy(out, resp.Body)
	out.Close()
	if err != nil {
		return fmt.Errorf("failed to write zip file: %w", err)
	}

	// 解压 zip 文件
	if err := unzip(zipPath, tmpDir); err != nil {
		return fmt.Errorf("failed to unzip: %w", err)
	}

	// 找到解压后的库文件
	srcLibPath := filepath.Join(tmpDir, fmt.Sprintf("onnxruntime-%s-%s-%s", platform, arch, version), libName)
	if _, err := os.Stat(srcLibPath); os.IsNotExist(err) {
		return fmt.Errorf("library not found in zip: %s", srcLibPath)
	}

	// 确定系统目录并创建
	sysLibDir, err := getSystemLibDir()
	if err != nil {
		return fmt.Errorf("failed to get system lib directory: %w", err)
	}
	if err := os.MkdirAll(sysLibDir, 0755); err != nil {
		return fmt.Errorf("failed to create system lib directory: %w", err)
	}

	// 复制到系统目录
	dstLibPath := filepath.Join(sysLibDir, libName)
	if err := copyFile(srcLibPath, dstLibPath); err != nil {
		return fmt.Errorf("failed to copy library to system directory: %w", err)
	}

	fmt.Printf("Downloaded onnxruntime library to: %s\n", dstLibPath)
	return nil
}

func getSystemLibDir() (string, error) {
	switch runtime.GOOS {
	case "windows":
		// Windows 系统目录
		windir := os.Getenv("SystemRoot")
		if windir == "" {
			windir = "C:\\Windows"
		}
		return filepath.Join(windir, "System32"), nil
	case "linux":
		// Linux 系统目录
		return "/usr/local/lib", nil
	case "darwin":
		// macOS 系统目录
		return "/usr/local/lib", nil
	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

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
	// 简单的 unzip 实现
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

// Load 加载模型
func (m *Model) Load(modelPath string) error {
	// 确保动态库存在
	if err := ensureLibrary(); err != nil {
		return fmt.Errorf("failed to ensure library: %w", err)
	}

	// 检查模型文件是否存在
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return fmt.Errorf("model file not found: %s", modelPath)
	}

	// 创建推理引擎
	engine, err := ort.NewEngine(ort.DefaultLibraryPath())
	if err != nil {
		return fmt.Errorf("failed to create engine: %w", err)
	}
	m.Engine = engine

	// 创建会话配置
	opts, err := engine.NewSessionOptions()
	if err != nil {
		return fmt.Errorf("failed to create session options: %w", err)
	}
	opts.SetIntraOpNumThreads(4)

	// 创建推理会话 (加载模型)
	session, err := engine.NewSession(modelPath, opts)
	if err != nil {
		engine.Destroy()
		return fmt.Errorf("failed to create session: %w", err)
	}
	m.Session = session

	m.Path = modelPath
	m.isLoaded = true

	return nil
}

// LoadWithConfig 使用配置加载模型
func (m *Model) LoadWithConfig(modelPath string, cfg *Config) error {
	// 确保动态库存在
	if err := ensureLibrary(); err != nil {
		return fmt.Errorf("failed to ensure library: %w", err)
	}

	// 检查模型文件是否存在
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return fmt.Errorf("model file not found: %s", modelPath)
	}

	// 创建推理引擎
	engine, err := ort.NewEngine(ort.DefaultLibraryPath())
	if err != nil {
		return fmt.Errorf("failed to create engine: %w", err)
	}
	m.Engine = engine

	// 创建会话配置
	opts, err := engine.NewSessionOptions()
	if err != nil {
		return fmt.Errorf("failed to create session options: %w", err)
	}
	if cfg.NumThreads > 0 {
		opts.SetIntraOpNumThreads(cfg.NumThreads)
	}

	// 创建推理会话 (加载模型)
	session, err := engine.NewSession(modelPath, opts)
	if err != nil {
		engine.Destroy()
		return fmt.Errorf("failed to create session: %w", err)
	}
	m.Session = session

	m.Path = modelPath
	m.isLoaded = true

	return nil
}

// Unload 卸载模型
func (m *Model) Unload() {
	if m.Session != nil {
		m.Session.Destroy()
		m.Session = nil
	}
	if m.Engine != nil {
		m.Engine.Destroy()
		m.Engine = nil
	}
	m.isLoaded = false
}

// IsLoaded 检查模型是否已加载
func (m *Model) IsLoaded() bool {
	return m.isLoaded
}

// Infer 执行推理
func (m *Model) Infer(prompt string) (string, error) {
	if !m.isLoaded {
		return "", fmt.Errorf("model not loaded")
	}

	// TODO: 实现实际的 tokenize 和 detokenize
	// 目前是占位实现，返回格式化的提示

	return fmt.Sprintf("[LLM Inference]\nPrompt: %s\nModel: %s\nStatus: Not fully implemented - requires onnxruntime_purego integration", prompt, m.Name), nil
}

// InferWithConfig 使用配置执行推理
func (m *Model) InferWithConfig(prompt string, cfg *Config) (string, error) {
	if !m.isLoaded {
		return "", fmt.Errorf("model not loaded")
	}

	// 使用配置进行推理
	result, err := m.Infer(prompt)
	if err != nil {
		return "", err
	}

	return result, nil
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
			models = append(models, entry.Name())
		}
	}

	return models, nil
}

// MockInfer 模拟推理（用于测试）
func MockInfer(prompt string) string {
	// 简单的 mock 实现
	return fmt.Sprintf("[Mock LLM Response]\nProcessed: %s\n\nThis is a mock response for testing purposes.", prompt)
}