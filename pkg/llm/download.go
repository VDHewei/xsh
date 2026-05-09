package llm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// DownloadConfig 下载配置
type DownloadConfig struct {
	Proxy    string // HTTP 代理地址
	Mirror   string // HuggingFace 镜像站
	Token    string // HuggingFace API Token
	CacheDir string // 缓存目录
}

// defaultMirror 全局默认镜像站
var defaultMirror string

// NewDownloadConfig 创建默认下载配置
func NewDownloadConfig() *DownloadConfig {
	return &DownloadConfig{
		Proxy:    "",
		Mirror:   defaultMirror,
		Token:    os.Getenv("HF_TOKEN"),
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

// lfsPointer LFS 指针文件内容
type lfsPointer struct {
	OID  string
	Size int64
}

// DownloadFromHuggingFace 从 HuggingFace 下载模型
func DownloadFromHuggingFace(repoID string, cfg *DownloadConfig) (*DownloadedModel, error) {
	baseURL := "https://huggingface.co"
	if cfg.Mirror != "" {
		baseURL = cfg.Mirror
	}

	modelName := strings.ReplaceAll(repoID, "/", "_")
	downloadPath := filepath.Join(cfg.CacheDir, modelName)

	if err := os.MkdirAll(downloadPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// 获取模型文件列表
	treeURL := fmt.Sprintf("%s/api/%s/revision/main", baseURL, repoID)
	if cfg.Mirror != "" {
		treeURL = cfg.Mirror + "/api/" + repoID + "/revision/main"
	}

	body, err := doHTTPGet(treeURL, cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("fetch file list for %s: %w", repoID, err)
	}

	var treeResponse struct {
		Tree []struct {
			Path string `json:"path"`
			Type string `json:"type"`
		} `json:"tree"`
	}

	if err := json.Unmarshal(body, &treeResponse); err != nil {
		return nil, fmt.Errorf("parse API response for %s: %w", repoID, err)
	}

	httpClient := &http.Client{}
	totalSize := int64(0)

	for _, item := range treeResponse.Tree {
		if item.Type == "blob" {
			filePath := filepath.Join(downloadPath, item.Path)
			size, err := downloadFileReal(httpClient, repoID, item.Path, filePath, cfg)
			if err != nil {
				fmt.Printf("Warning: failed to download %s: %v\n", item.Path, err)
				continue
			}
			totalSize += size
		}
	}

	if totalSize == 0 {
		return nil, fmt.Errorf("no files downloaded for %s", repoID)
	}

	return &DownloadedModel{
		Name:   modelName,
		Path:   downloadPath,
		Size:   totalSize,
		RepoID: repoID,
	}, nil
}

// downloadFileReal 下载文件，自动处理 LFS 指针文件
func downloadFileReal(client *http.Client, repoID, filename, destPath string, cfg *DownloadConfig) (int64, error) {
	baseURL := "https://huggingface.co"
	if cfg.Mirror != "" {
		baseURL = cfg.Mirror
	}
	baseURL = strings.TrimSuffix(baseURL, "/raw")
	baseURL = strings.TrimSuffix(baseURL, "/tree")

	url := fmt.Sprintf("%s/%s/resolve/main/%s", baseURL, repoID, filename)

	data, err := doHTTPGet(url, cfg.Token)
	if err != nil {
		return 0, fmt.Errorf("fetch: %w", err)
	}

	// 如果是 LFS 指针文件，通过 LFS API 获取真实内容
	if ptr := parseLFSPointer(data); ptr != nil {
		data, err = downloadLFSObject(client, repoID, ptr, cfg)
		if err != nil {
			return 0, fmt.Errorf("lfs: %w", err)
		}
	}

	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return 0, err
	}

	if err := os.WriteFile(destPath, data, 0644); err != nil {
		return 0, err
	}

	return int64(len(data)), nil
}

// parseLFSPointer 检测 LFS 指针文件并解析
func parseLFSPointer(data []byte) *lfsPointer {
	if !bytes.HasPrefix(data, []byte("version https://git-lfs.github.com")) {
		return nil
	}

	var ptr lfsPointer
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "oid sha256:") {
			ptr.OID = strings.TrimPrefix(line, "oid sha256:")
		}
		if strings.HasPrefix(line, "size ") {
			ptr.Size, _ = strconv.ParseInt(strings.TrimPrefix(line, "size "), 10, 64)
		}
	}

	if ptr.OID == "" {
		return nil
	}
	return &ptr
}

// downloadLFSObject 通过 HuggingFace LFS 批量 API 下载真实文件
func downloadLFSObject(client *http.Client, repoID string, ptr *lfsPointer, cfg *DownloadConfig) ([]byte, error) {
	baseURL := "https://huggingface.co"
	if cfg.Mirror != "" {
		baseURL = cfg.Mirror
	}

	batchURL := fmt.Sprintf("%s/api/%s.git/info/lfs/objects/batch", baseURL, repoID)

	reqBody := map[string]interface{}{
		"operation": "download",
		"objects": []map[string]interface{}{
			{
				"oid":  ptr.OID,
				"size": ptr.Size,
			},
		},
	}

	jsonBody, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", batchURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.git-lfs+json")
	if cfg.Token != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.Token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("lfs batch request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lfs batch: %s - %s", resp.Status, string(body))
	}

	var batchResp struct {
		Objects []struct {
			OID    string `json:"oid"`
			Size   int64  `json:"size"`
			Actions *struct {
				Download *struct {
					Href string `json:"href"`
				} `json:"download"`
			} `json:"actions"`
			Error *struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		} `json:"objects"`
	}

	if err := json.Unmarshal(body, &batchResp); err != nil {
		return nil, fmt.Errorf("parse lfs batch response: %w", err)
	}

	if len(batchResp.Objects) == 0 {
		return nil, fmt.Errorf("no objects in lfs batch response")
	}

	obj := batchResp.Objects[0]
	if obj.Error != nil {
		return nil, fmt.Errorf("lfs object error: %s (code %d)", obj.Error.Message, obj.Error.Code)
	}

	if obj.Actions == nil || obj.Actions.Download == nil {
		return nil, fmt.Errorf("no download action for lfs object")
	}

	downloadURL := obj.Actions.Download.Href
	data, err := doHTTPGet(downloadURL, cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("lfs download: %w", err)
	}

	return data, nil
}

// doHTTPGet 发起 GET 请求并返回 body
func doHTTPGet(url, token string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		if len(body) > 200 {
			body = body[:200]
		}
		return nil, fmt.Errorf("%s: %s", resp.Status, string(body))
	}

	return body, nil
}

// downloadFileWithURL 使用指定 URL 下载文件
func downloadFileWithURL(client *http.Client, url, destPath string, cfg *DownloadConfig) error {
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

	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
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

// resolveModelRepo 根据候选模型名解析完整 HuggingFace RepoID
func resolveModelRepo(candidate string) string {
	candidates := DefaultCandidateModels()
	for _, c := range candidates {
		if c.Name == candidate {
			return c.RepoID
		}
	}
	return candidate
}

// SearchModels 搜索 HuggingFace 模型
func SearchModels(query string, cfg *DownloadConfig) ([]string, error) {
	baseURL := "https://huggingface.co"
	if cfg.Mirror != "" {
		baseURL = cfg.Mirror
	}

	apiURL := fmt.Sprintf("%s/api/models?search=%s&sort=downloads&direction=-1&limit=10", baseURL, query)

	body, err := doHTTPGet(apiURL, cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("search API: %w", err)
	}

	var models []struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &models); err != nil {
		return nil, fmt.Errorf("parse search results: %w", err)
	}

	var names []string
	for _, m := range models {
		names = append(names, m.ID)
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("no models found for query: %s", query)
	}
	return names, nil
}

// GetModelInfo 获取模型信息
func GetModelInfo(repoID string, cfg *DownloadConfig) (map[string]interface{}, error) {
	baseURL := "https://huggingface.co"
	if cfg.Mirror != "" {
		baseURL = cfg.Mirror
	}

	apiURL := fmt.Sprintf("%s/api/models/%s", baseURL, repoID)

	body, err := doHTTPGet(apiURL, cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("model info API: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse model info: %w", err)
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
	defaultMirror = mirror
}
