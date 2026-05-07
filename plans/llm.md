##  onnx 本地推理服务

开发需求:
1. 使用 onnxruntime_purego 库 实现本地 llm 模型加载 与推理
2. 实现 model Hugging Face  下载（指定代理和镜像）
3. 接入 xsh cli , 支持 model search, model list model select
4. 实现后测试 llm 本地推理 识别 tests/data/prod-migration-form-uat.txt ,并调用任务规划和tui 交互
5. 测试通过记录到 tests/data/ 测试结果

## 开发目录
将 llm 功能集中到 pkg/llm 中开发