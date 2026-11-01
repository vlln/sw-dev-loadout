# ledger

一个内存账本服务。金额一律按整数分存储，不用浮点。

## 结构

| 包 | 职责 |
|---|---|
| `internal/money` | 金额解析与渲染（`Cents` 为 int64 分） |
| `internal/store` | 账户与分录的内存存储，全局自增 seq |
| `internal/api` | HTTP 路由与固定的 JSON 错误信封 |
| `cmd/ledgerd` | 启动服务，预置 `cash` 与 `bank` 两个账户 |

## 现有接口

| 方法与路径 | 作用 |
|---|---|
| `GET /accounts` | 列出全部账户及余额 |
| `GET /accounts/{id}` | 单个账户 |
| `GET /accounts/{id}/entries` | 账户的分录，按 seq 升序 |
| `POST /accounts/{id}/entries` | 记一笔，返回 seq 与新余额 |

错误信封固定为 `{"error": "<code>", "detail": "<human readable>"}`。

## 开发

```
go test ./...
go build ./...
go run ./cmd/ledgerd -addr 127.0.0.1:8080
```
