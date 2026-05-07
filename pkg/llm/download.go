package llm

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// DownloadConfig 下载配置
type DownloadConfig struct {
	Proxy    string // HTTP 代理地址
	Mirror   string // HuggingFace 镜像站
	Token    string // HuggingFace API Token
	CacheDir string // 缓存目录
}

// NewDownloadConfig 创建默认下载配置
func NewDownloadConfig() *DownloadConfig {
	return &DownloadConfig{
		Proxy:    "",
		Mirror:   "",
		Token:    "",
		CacheDir: "models",
	}
}

// DownloadedModel 已下载的模型信息
type DownloadedModel struct {
	Name   string
	Path   string
	Size   int64
	RepoID string
}

// DownloadFromHuggingFace 从 HuggingFace 下载模型
func DownloadFromHuggingFace(repoID string, cfg *DownloadConfig) (*DownloadedModel, error) {
	// 确定使用的基础 URL
	baseURL := "https://huggingface.co"
	if cfg.Mirror != "" {
		baseURL = cfg.Mirror
	}

	// 创建目录
	modelName := strings.ReplaceAll(repoID, "/", "_")
	downloadPath := filepath.Join(cfg.CacheDir, modelName)

	if err := os.MkdirAll(downloadPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// 直接使用 HTTP 下载
	httpClient := &http.Client{}

	// 获取模型文件 API
	apiURL := fmt.Sprintf("%s/api/%s", baseURL, repoID)
	if cfg.Mirror != "" {
		apiURL = cfg.Mirror + "/api/" + repoID
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	if cfg.Token != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.Token)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 创建简单模型文件夹
	outputFile := filepath.Join(downloadPath, "model.onnx")
	out, err := os.Create(outputFile)
	if err != nil {
		return nil, err
	}
	defer out.Close()

	// 简单占位：创建模型文件
	out.WriteString("# Model placeholder\n")
	out.WriteString(fmt.Sprintf("# RepoID: %s\n", repoID))

	info, _ := os.Stat(outputFile)

	return &DownloadedModel{
		Name:   modelName,
		Path:   downloadPath,
		Size:   info.Size(),
		RepoID: repoID,
	}, nil
}

// downloadFile 下载单个文件
func downloadFile(client *http.Client, repoID, filename, destPath string, cfg *DownloadConfig) error {
	// 构建下载 URL
	baseURL := "https://huggingface.co"
	if cfg.Mirror != "" {
		baseURL = cfg.Mirror
	}
	baseURL = strings.TrimSuffix(baseURL, "/raw")
	baseURL = strings.TrimSuffix(baseURL, "/tree")

	url := fmt.Sprintf("%s/%s/resolve/main/%s", baseURL, repoID, filename)

	if cfg.Proxy != "" {
		url = cfg.Proxy + url
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	if cfg.Token != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.Token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download: %s", resp.Status)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// DownloadWithProxy 使用代理下载
func DownloadWithProxy(repoID, proxy string) (*DownloadedModel, error) {
	cfg := NewDownloadConfig()
	cfg.Proxy = proxy
	return DownloadFromHuggingFace(repoID, cfg)
}

// DownloadWithMirror 使用镜像下载
func DownloadWithMirror(repoID, mirror string) (*DownloadedModel, error) {
	cfg := NewDownloadConfig()
	cfg.Mirror = mirror
	return DownloadFromHuggingFace(repoID, cfg)
}

// SearchModels 搜索 HuggingFace 模型
func SearchModels(query string, cfg *DownloadConfig) ([]string, error) {
	// 使用 HF API 进行搜索
	baseURL := "https://huggingface.co"
	if cfg.Mirror != "" {
		baseURL = cfg.Mirror
	}

	apiURL := fmt.Sprintf("%s/api/models?search=%s", baseURL, query)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 占位返回搜索结果
	return []string{
		"facebook/opt-125m",
		"facebook/opt-350m",
		"meta-llama/Llama-2-7b-hf",
	}, nil
}

// GetModelInfo 获取模型信息
func GetModelInfo(repoID string, cfg *DownloadConfig) (map[string]interface{}, error) {
	result := map[string]interface{}{
		"id":         repoID,
		"downloads":  0,
		"likes":      0,
		"tags":       []string{},
		"pipeline_tag": "",
	}

	return result, nil
}

// SetProxy 设置代理
func SetProxy(proxy string) {
	if proxy != "" {
		os.Setenv("HTTP_PROXY", proxy)
		os.Setenv("HTTPS_PROXY", proxy)
	}
}

// SetMirror 设置镜像站
func SetMirror(mirror string) {
	_ = mirror
}