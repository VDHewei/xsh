# 健康检查命令

## 描述
快速检查所有服务端点是否正常响应, 包括 GET, POST, PUT, DELETE 四种方法.

## 任务
[GET] http://localhost:18080/health
[GET] http://localhost:18080/get
[POST] http://localhost:18080/post
header: Content-Type=application/json
body: {"test":"health"}
[PUT] http://localhost:18080/put
body: {"test":"health"}
[DELETE] http://localhost:18080/delete
@check: 所有端点是否正常响应?
