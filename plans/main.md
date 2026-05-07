## 任务大纲

概括:

1. 使用 golang TUI 框架 完成一个 读取 对应 文件(txt,md), 完成自动化 任务(http,ssh,grpc) 生成
2. 支持@{command}:{意图描述} 形式命令解析，目前执行 @ask [询问交互,可能会动态添加用户任务], @wait  [任务执行后等待时间计时器], @check: [打印输出任务响应结果，询问用户任务是否继续]
3. 任务 动态规划 和 @{command}:{意图描述} 支持接入 llm 推理服务
4. 任务验收 ： 读取 tests/data/prod-migration-form-uat.txt, 实现人机交互
5. 测试通过后, 入口 CLI 支持 指定 -i input 任务文件，-o output 记录任务执行结果
6. 测试未验收通过，继续修正和完成任务