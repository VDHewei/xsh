# T1 Entry Layer & GLM5.1 Support - 任务执行提示词

## Step 1: 定义 ModelCandidate 结构体 ✅
```
在 internal/types/types.go 末尾新增 ModelCandidate 结构体:
- RepoID string  (HuggingFace 仓库 ID)
- Name string    (短名称)
- Default bool   (是否为默认模型)
```

## Step 2: 修复 DefaultCandidateModels() 返回类型 ✅
```
修改 pkg/llm/model.go:
1. 导入 internal/types 包
2. DefaultCandidateModels() 返回类型从 map[string]string 改为 []types.ModelCandidate
3. 新增 ResolveCandidateName(name) 辅助函数

修改 pkg/llm/download.go:
1. resolveModelRepo 改用新的 DefaultCandidateModels() 返回类型
```

## Step 3: 增强 GetModelPath 候选名匹配 ✅
```
修改 pkg/llm/model.go GetModelPath:
1. 先尝试直接路径 modelDir/name
2. 直接路径不存在时，通过 ResolveCandidateName 查找
3. 将 RepoID (yasserrmd/xxx) 转换为目录名 (yasserrmd_xxx)
4. 检查该目录是否存在，存在则返回
5. 都找不到则返回直接路径(保持向后兼容)
```

## Step 4: 增强 xsh model list 候选标记 🔄
```
修改 pkg/llm/model.go:
1. 新增 ModelInfo 结构体: Name, Installed, Candidate
2. 新增 ListModelsWithCandidates(dir) 函数
   - 扫描本地已安装模型
   - 合并候选模型列表
   - 标记已安装/未安装状态

修改 pkg/llm/cli.go:
1. handleModelList() 调用 ListModelsWithCandidates
2. 输出格式: 已安装候选显示 [candidate], 未安装候选显示 [not installed]
3. 非候选模型不显示额外标记

修改 pkg/llm/cli.go:
1. ModelList() 同步更新
```

## Step 5: 修复动态库重复下载
```
修改 pkg/llm/model.go:
1. DownloadOnnxRuntimeGenAILibrary() 开头增加:
   - 检查 DefaultGenAILibraryPath() 文件是否存在
   - 检查 DefaultOnnxRuntimeLibraryPath() 文件是否存在
   - 两者都存在则跳过下载
2. downloadGenAILib() 开头增加 DefaultGenAILibraryPath() 检查
3. downloadOnnxRuntimeLib() 开头增加 DefaultOnnxRuntimeLibraryPath() 检查
```

## Step 6: 修复 runTestMode 动态库调用
```
修改 cmd/xsh/main.go runTestMode():
1. 将 _ = llm.DownloadOnnxRuntimeGenAILibrary() 改为:
   if err := llm.DownloadOnnxRuntimeGenAILibrary(); err != nil {
       fmt.Fprintf(os.Stderr, "Dynamic library error: %v\n", err)
       // 继续执行，模型加载时还会再次检查
   }
(Step 5 已增加缓存检查，所以直接调用也是安全的)
```

## Step 7: GetModelPath 候选名测试
```
在 pkg/llm/llm_test.go 新增:
1. TestGetModelPath_CandidateShortName
   - 创建测试目录 models/yasserrmd_deepseek-r1-distill-qwen-onnx/
   - 放入 genai_config.json
   - 验证 GetModelPath("models", "deepseek") 返回正确路径
2. TestGetModelPath_DirectPath
   - 创建普通目录
   - 验证直接路径不变
```

## Step 8: model list 候选标记测试
```
在 pkg/llm/llm_test.go 新增:
1. TestListModelsWithCandidates_Installed
   - 创建候选模型目录
   - 验证 ListModelsWithCandidates 返回 Installed=true
2. TestListModelsWithCandidates_NotInstalled
   - 只有候选列表、无本地目录
   - 验证返回 Installed=false, Candidate!=nil
```

## Step 9: 动态库缓存测试
```
在 pkg/llm/llm_test.go 新增:
1. TestDownloadLibraryCache
   - 首次调用下载(模拟)
   - 二次调用应跳过下载
   - 验证 DefaultGenAILibraryPath 文件存在后 DownloadOnnxRuntimeGenAILibrary 跳过
```

## Step 10: 运行全部测试
```
go test ./... -count=1
检查所有测试通过，包括新增测试
```

## Step 11: 更新 README (中英双版)
```
修改 README.md 和 README_CN.md:
1. 补充 GLM5.1 模型支持说明
2. 补充 xsh model list 候选模型标记示例
3. 补充动态库缓存说明(首次下载，后续自动复用)
```

---

生成时间: 2026-05-09
