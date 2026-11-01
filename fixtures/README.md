# 夹具

每份夹具是一次对照实验的固定输入。重跑时两组用同一份输入、同一句任务、同一条禁令（不许看另一组的产出），产出写到不同文件。

两组的差别只有一个：基线组不给本仓库的 skill，处理组完整读本仓库的 skill 并照它做。基线组要同样认真做，否则对照失效。

对照记录在 [docs/runs/](../docs/runs/)，结论在 [docs/experiment.md](../docs/experiment.md)。

## review-diff

审一份 PR diff。[CONTEXT.md](review-diff/CONTEXT.md) 给背景与包信息，[patch.diff](review-diff/patch.diff) 是被审对象。

埋了 11 个问题：6 个阻断级（README 新契约当场为假、排空有硬上限且丢弃量不报告、吞掉失败还计入成功数、两个新增测试在任何实现下都会通过、覆盖率阈值 90 → 85 且与功能改动挤在一个 PR、CI 删掉 Node 20 与 `engines` 冲突），5 个建议级（默认 logger 写宿主 stdout、`limit` 不校验、返回类型变更未声明、注释是变更叙事、失败日志不含 job 身份）。

**任务**：按 `CONTEXT.md` 审这份 diff，给出发现与结论；不要修代码。

**度量**：命中几个埋的问题、有没有误报、阻断与建议分级是否正确。

夹具自检（`patch.diff` 的 hunk 头与 hunk 体一致、能被 `git apply --check` 接受）：

```sh
cd "$(mktemp -d)" && git init -q .
# 按 hunk 头构造一个最小基线树，然后：
git apply --check -v <本仓库>/fixtures/review-diff/patch.diff
```

初版的三个 hunk 头计数写错了，`git apply` 会以 corrupt patch 拒绝；处理组把它当格式问题报了出来，基线组没报。2026-11-01 已修正。夹具自身坏掉时，所有轮次测的都是夹具。

## tally-0.3.0

一个小 CLI 的 0.3.0 源码树，加一份需求 [FEATURE.md](tally-0.3.0/FEATURE.md)：加 `--json`、支持 stdin 读、二进制跳过、错误要有非零退出码，而现有 CLI 是公开接口。

这份树被两轮用过：

- **第二轮**：只产出实施计划，不写代码。度量计划的完整性（切片、风险、回滚、用例）与体量。
- **第三轮**：把第二轮处理组的计划交给另一个 agent 执行。度量计划在执行时撞上多少矛盾与返工。

**已知缺陷（0.3.0 自带，不是实验引入）**：`src/cli.mjs` 没有 shebang，真实 `npm install` 后 `node_modules/.bin/tally` 直接失败。解包后跑 `node cli.mjs` 能通过，所以只做这一步的冒烟测不出它——第三轮正是这样暴露的。

第三轮的产出保存为 [docs/runs/r3-result.bundle](../docs/runs/r3-result.bundle)，含基线提交、`v0.3.0` 标签与四个 PR 分支：

```sh
git clone docs/runs/r3-result.bundle r3 && cd r3 && npm test
```

核对 2026-11-01：34 个用例全通过；同一套用例在 `v0.3.0` 上 12 通过 / 22 失败。

## envcheck-0.2.0

一个只用 Python 标准库的小命令行工具（两个模块共 85 行，9 个用例全绿），加一份需求 [FEATURE.md](envcheck-0.2.0/FEATURE.md)：加 `--quiet`、flag 顺序任意、rc 里带 `=` 的行按不可用处理、所有退出码 2 的信息都要带 rc 路径。

和 `tally` 那份夹具刻意不同：换语言（Python / Node）、换规模、换任务性质（四条小修改 vs 一次功能新增）。

**任务**：先写实施计划到仓库外，再按自己的计划实现、提交、验证，并报告计划里哪些地方不对。

**埋了两处**，都在 0.2.0 上实测确认过：

- 需求 4 在两条退出码 2 的路径上**不可满足**：`--rc` 缺参数、多余参数都发生在读 rc 之前，没有真实路径可带。照字面做就会报出一个工具从未打算读的文件名。
- `cannot read` 这条信息**本来就带路径**。为需求 4 写的那条用例在旧版本上就是通过的，把它当作"新行为的证据"是错的。

**度量**：新增用例里有几条在 0.2.0 上就通过、其中几条是作者自己识别出来的守卫、几条是漏网。

自检：

```sh
cd fixtures/envcheck-0.2.0 && PYTHONPATH=. python3 -m unittest discover -s tests
```

核对 2026-11-01：9 个用例全通过。两组的产出保存为 [docs/runs/r4-baseline.bundle](../docs/runs/r4-baseline.bundle)（21 用例）与 [docs/runs/r4-treatment.bundle](../docs/runs/r4-treatment.bundle)（22 用例），`git clone` 后跑同一条命令都全绿。

## ledger-0.2.0

一个内存账本的 Go HTTP 服务（403 行代码 / 267 行测试 / 17 个用例 / 3 个包），加一份需求 [FEATURE.md](ledger-0.2.0/FEATURE.md)：加 `POST /transfers`，7 条需求（原子、金额为正、不能转给自己、账户不存在、幂等键、余额口径唯一、透支下限）。

这份夹具是为验「流程深度与改动规模成比例」而造的。前一份（`envcheck`）验不出来，因为 85 行、4 条小需求的任务不会诱发过度拆分——**夹具没给被验的规则失败的机会**。这份放大到约 4.7 倍，并让"看起来可分"与"实际耦合"冲突：7 条需求看着像 7 个单元，但 R1 / R6 / R7 共用同一段临界区，拆开交付必然返工或留下两个余额口径。

**埋了两处**，都在 0.2.0 上实测确认过：

- **锁不可重入**：`sync.Mutex` 非 reentrant，在转账的临界区里复用公开的 `Balance()` 会死锁。R1（两条分录原子）与 R6（全服务只有一个余额口径）直接冲突，必须下推到 store 或新增不加锁的内部函数。
- **错误信封没有 409**：`writeErr` 的 status map 没有 `conflict` 键，取零值后回落到 500。R7 要 409，照抄现有写法会静默变成 500；若用例只断言 `error == "overdraft"` 而不断言状态码，仍然通过。

**还有一处不是埋的，是造夹具时没发现的真缺陷**（第五轮两组都自己找到了，2026-11-01 各自实跑确认）：`money.Parse` 对 `w*100 + f` 不做范围检查，静默溢出且 `err == nil`——`"184467440737095516.20"` 解析成 0.04，`"9223372036854775807"` 解析成 -1.00。它绕过 R2 的正值检查与 R7 的透支算术。重跑这份夹具时，找到它不算任何一组的本事。

**度量**：拆成几个单元、拆分依据是需求条数还是耦合关系、计划体量、以及两处埋点各组是否踩中。

夹具自带一套[黑盒验收用例](ledger-0.2.0/acceptance/probe.py)，只打公开 HTTP 接口，对内部设计无偏好。它在实验启动前写好、两组都看不到：

```sh
cd fixtures/ledger-0.2.0 && GOFLAGS=-mod=mod GOPROXY=off go test ./...
python3 fixtures/ledger-0.2.0/acceptance/probe.py fixtures/ledger-0.2.0 8741
```

核对 2026-11-01：17 个用例全通过；验收用例在 0.2.0 上 3/41 通过。那 3 条是故意留的旧行为守卫（单账户记账仍 201、未知账户仍 404、错误方法仍 405），其余 38 条必须在 0.2.0 上失败。初版有 4 条是**空洞通过**的（R6-c、R5-c、R5-d、R7-g——两边都是 404 时"两值相等"自然成立），启动前已改成同时断言取值，否则这套验收自己就犯了它要检查的毛病。
