# 06 多租户全文检索与索引服务

纯Go REST 服务，提供租户、集合映射、文档写入/删除、内存倒排索引、匹配查询、高亮、分页、快照和本地WAL恢复。存储通过接口隔离，生产环境可替换PostgreSQL与对象存储适配器；`migrations/`包含基础表结构。

## 启动

```bash
go test ./...
go vet ./...
go run ./cmd/searchd
```

服务默认监听 `:8086`，健康检查为 `/healthz`、`/readyz`，指标为 `/metrics`。配置文件为 `configs/config.yaml`，支持`SEARCH_HTTP_ADDR`、`SEARCH_DATA_DIR`和`SEARCH_QUERY_BUDGET`环境变量覆盖。

## API流程

先 `POST /v1/tenants` 创建租户，再 `POST /v1/collections` 提交字段映射（text、keyword、integer、float、boolean、timestamp、geo_point），随后 `POST /v1/documents` 写入文档。使用 `POST /v1/search?collection_id=...` 发送 `match`、`term`、`phrase`、`prefix`、`range` 查询，`highlight:true`返回高亮字段；`POST /v1/snapshots`生成JSON快照。文档通过`DELETE /v1/documents/{id}`逻辑删除。

WAL位于`data/wal.jsonl`，服务重启会重放事件恢复租户、集合和文档；快照位于`data/snapshots`。请求包含最大4MiB体积限制、request ID、超时、恢复和统一JSON错误处理中间件。后台快照任务均接受context取消。

## 目录

`cmd`启动入口；`internal/{tenant,collection,document,analysis,index,query,snapshot,quota}`按领域分层；`api`保留协议定义；`configs`、`migrations`、`deploy`、`scripts`提供部署与运维资产。
