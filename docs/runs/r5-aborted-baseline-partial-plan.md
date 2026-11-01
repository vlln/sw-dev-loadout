# 实现计划：转账接口 v0.3.0

## 基线确认

- `go version` = go1.26.4 darwin/arm64，`git describe --tags` = v0.2.0，工作树干净。
- `GOFLAGS=-mod=mod GOPROXY=off go test ./...` 全绿，`-v` 下 `=== RUN` 计数 = 17。
- 只用标准库，`go.mod` 不新增依赖。

## 核心设计决策

### 决策 1：转账逻辑放在 store，不放在 api

R1（原子性）、R5（幂等）、R7（透支检查）三条需求都要求"检查与写入之间不能被别的请求插进来"。
api 层只有 `Post` + `Balance` 两个已加锁的调用，任何组合都会在两次加锁之间留下窗口：

- 先 `Post(from)` 再 `Post(to)`，中间若第二笔失败就是 R1 禁止的"只扣款不入账"。
- 先 `Balance(from)` 查透支再 `Post`，中间别的请求可以把余额改掉，透支下限形同虚设。
- 幂等键的"查到则复用、查不到则执行"同理，并发同键会双写。

因此新增一个 store 方法，在**同一把锁**里完成：幂等命中检查 → 账户存在检查 → 透支检查 → 两条分录写入 → 余额计算。
api 层退化成纯粹的解析与错误映射，不持有任何业务规则。

签名（store 包）：

```go
type TransferRequest struct {
    From, To       string
    Amount         money.Cents
    Memo           string
    IdempotencyKey string // 空串表示不启用幂等
}

type TransferResult struct {
    SeqFrom, SeqTo           int
    FromBalance, ToBalance   money.Cents
    Replayed                 bool // true = 幂等重放，未新增分录
}

func (s *Store) Transfer(req TransferRequest) (TransferResult, error)
```

### 决策 2：校验规则（R2/R3）也放在 store

`amount <= 0` 与 `from == to` 本可以在 api 层挡掉，但放在 store 里意味着这个账本**无法**通过任何
入口被写进非法转账。api 层再把 store 的错误映射成 HTTP 状态码。代价是 api 层的 400 分支依赖 store
的错误类型，可接受——两者本来就同模块。

### 决策 3：余额唯一口径（R6）用"抽出已持锁的私有函数"实现

现在 `Balance` 自己加锁后对 `a.entries` 求和。`Transfer` 不能调用 `Balance`（会自死锁），
若复制一份求和逻辑就出现了第二个口径，正是 R6 要禁的。

做法：抽出 `balanceLocked(a *Account) money.Cents`，`Balance` 与 `Transfer` 都调它。
全服务只有这一处对分录求和的代码。

`TransferResult` 里的余额由 `Transfer` **在锁内**算出返回，api 层不再另外调一次 `Balance`——
否则写入与读取之间又有窗口，并发下 R6 会被打破。

### 决策 4：校验与副作用的先后顺序

```
HTTP 层：  方法检查 → body 解码 → amount 文本解析
store 锁内：amount>0 → from!=to → 幂等命中？→(命中则直接返回)→ 账户存在 → 透支 → 写入
```

**幂等命中检查必须早于透支检查**，这是个容易写反的坑：首次转账把余额压到 -90.00 之后，
同键重放若先过透支检查就会得到 409，而 R5 要求它返回首次的结果。

### 决策 5：透支下限

`const OverdraftLimit money.Cents = -10000`（即 -100.00）写在 store。
检查式为 `balanceLocked(from) - amount < OverdraftLimit` 则拒绝；
即**恰好等于 -100.00 放行**，-100.01 拒绝。R7 说"不得低于 -100.00"，故边界值合法。
错误里带上 `会变成多少`（`balance - amount`），供 detail 使用。

### 决策 6：幂等记录只记成功的转账

失败的请求没有"首次的结果"可返回，键不被占用，同键可以重试。
记录为 `map[string]TransferResult`，与 accounts 共用同一把锁。

### 决策 7：同键但请求体不同

FEATURE.md R5 只说"返回首次的结果"，没要求检测参数不一致。
按字面实现：**无条件返回首次结果**，不比对 from/to/amount。这是有意的取舍，会写进 REPORT.md。

### 决策 8：409 错误信封

`writeErr` 的状态码映射表新增 `"overdraft": http.StatusConflict`。
不改 `writeErr` 现有签名（它的 `code`/`errCode` 双参数设计有点绕，但改它属于未被要求的重构）。

## 工作单元与提交顺序

| # | 单元 | 内容 |
|---|---|---|
| 1 | 计划 | 本文件，实现代码之前提交 |
| 2 | store 层 | `Transfer` + 错误类型 + `balanceLocked` 重构 + store 测试 |
| 3 | api 层 | `POST /transfers` + 409 映射 + api 测试 |
| 4 | 文档 | README 补 `/transfers` 一行 |
| 5 | 报告 | REPORT.md |

先 store 后 api：store 是不变量持有者，先把它测透，api 只剩映射，缺陷定位范围小。

## 错误类型

```go
// 包装 ErrNoAccount，保留 errors.Is 兼容性，同时携带 R4 要的"是哪一个"
type MissingAccountError struct { Role, ID string } // Role = "from" | "to"

// 携带 R7 要的"会变成多少"
type OverdraftError struct { Account string; Resulting, Limit money.Cents }

var ErrSameAccount, ErrNonPositiveAmount error
```

`MissingAccountError.Unwrap()` 返回 `ErrNoAccount`，这样现有依赖 `errors.Is(err, ErrNoAccount)`
的代码路径不受影响。

## 测试计划（新增，不碰现有 17 个）

store 包：

- 成功转账：两条分录、seq 递增（from 在前）、两边余额正确
- 透支拒绝时**零副作用**：分录数与 nextSeq 都不变
- 账户不存在时零副作用，且错误能定位到 from / to
- 透支边界：恰好 -100.00 放行，-100.01 拒绝
- 金额非正：0 与负数都拒绝
- 自转拒绝
- 幂等重放：同键第二次返回相同 seq/余额、`Replayed=true`、分录不增
- 幂等不吞失败：失败请求的键可以再用
- 并发同键（`-race`）：N 个 goroutine 同键只产生一对分录

api 包：

- 201 + 四个字段取值
- **R6 回归**：响应里的 `from_balance`/`to_balance` 与随后 `GET /accounts/{id}` 的 `balance` 逐字相等
- 400：`"0"`、`"-5"`、`"abc"`、`"1.234"`、自转、坏 body
- 404：from 不存在 / to 不存在，detail 里含账户 id
- 409：detail 里含转账后的余额字符串
- 405：`GET /transfers`
- 原子性：透支被拒后两个账户的 `GET .../entries` 都是空的
- 幂等：第二次 200、body 与首次相同、`GET .../entries` 条数不变

## 验收

- `GOFLAGS=-mod=mod GOPROXY=off go test ./...` 全绿
- `go test ./... -race` 全绿
- `go vet ./...` 干净
- `git diff v0.2.0 -- '*_test.go'` 确认现有测试文件零改动
- `=== RUN` 计数从 17 增长，且 17 个原用例名全部仍在
