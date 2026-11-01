# 执行报告：envcheck `--quiet` 与退出码 2 的收口

计划：[treatment-plan.md](./treatment-plan.md)
仓库：`/Users/vlln/Project/dsh-plugins/.sw-dev-experiment/round4/treatment`
基线：`ddee27b`（v0.2.0，9 用例通过）→ 最终 `f8c96de`
环境：Python 3.9.6，macOS。无 CI、无 hook，下列证据全部为本地运行。

---

## 1. 提交

三个，按计划 §9 的切片执行，每个提交单独可验证：

| # | 提交 | 性质 | 内容 | 该提交上的测试结果 |
|---|---|---|---|---|
| 1 | `610b7d4` | feat | `rc.py` 拒绝含 `=` 的行；3 条 parse 用例；`parse`/`RcError` 的契约注释 | 12/12 通过 |
| 2 | `646e286` | feat | `cli.py` 的 `--quiet`、顺序无关扫描、诊断带 rc 路径；8 条 CLI 用例；模块注释重写 | 22/22 通过 |
| 3 | `f8c96de` | docs | `README.md` 消费者契约，版本号 0.2.0 → 0.3.0 | 22/22 通过 |

`git diff --stat ddee27b HEAD`：4 文件，+144 / −17。工作树干净，无未跟踪残留。

提交 1 刻意只带 parse 层的用例：CLI 用例依赖提交 2 的实现，一并放进提交 1 会让第一个提交红着落地。这是切片时实际调整的一步（见 §4.1）。

未做：**没有打 `v0.3.0` 标签，没有推送**。提交、打标签、发布是三个独立授权动作，本次授权只覆盖「提交到 git」。README 的版本号已写 0.3.0，打标签留给你。

## 2. 测试结果

```
PYTHONPATH=. python3 -m unittest discover -s tests
Ran 22 tests in 0.349s
OK
```

基线 9 条一条未改——它们是「成功路径与缺失明细输出逐字节不变」这条约束的证据，改它们等于拆掉约束。新增 13 条（计划 §5 列了 11 条，实际 13 条，见 §4.2）。

### 2.1 新证据在基线上先跑过一遍

按 change-evidence「新证据必须在改动前的版本上真的跑一遍」：先在 `ddee27b` 的 detached worktree 上跑完整的新测试文件，逐条对照计划的预期。

| 用例 | 计划预期 | 基线实测 |
|---|---|---|
| `test_rejects_equals` | 失败 | 失败 ✓ |
| `test_accepts_equals_inside_comment` | 通过 | 通过 ✓ |
| `test_space_error_takes_precedence_over_equals` | 通过 | 通过 ✓ |
| `test_quiet_prints_only_summary` | 失败 | 失败 ✓ |
| `test_quiet_keeps_success_output` | 失败 | 失败 ✓ |
| `test_quiet_and_rc_order_is_irrelevant` | 失败 | 失败 ✓ |
| `test_equals_line_exits_two` | 失败 | 失败 ✓ |
| `test_unusable_rc_prints_nothing_to_stdout` | 失败 | 失败 ✓ |
| `test_space_error_names_rc_path` | 失败 | 失败 ✓ |
| `test_unexpected_argument_names_rc_path` | 失败 | 失败 ✓ |
| `test_quiet_does_not_suppress_diagnostics` | 失败 | 失败 ✓ |
| `test_cannot_read_error_names_rc_path` | 通过 | 通过 ✓ |
| `test_rc_without_path_names_no_file` | 通过 | 通过 ✓ |

13/13 逐条吻合：9 条新行为的证据确实在旧代码上失败，4 条守卫钉确实在旧代码上就通过（所以它们不是新行为的证据，只是防过度拒绝与防回归）。

`test_quiet_keeps_success_output` 在基线上失败的原因值得点明：不是因为输出不同，而是因为基线根本不认 `--quiet`，直接退 2。所以它同时钉住两件事——`--quiet` 被接受，且它在成功路径上不改任何字节。

### 2.2 逐字节差分（约束「成功路径与缺失明细输出必须逐字节不变」的外部证据）

不信任测试自述，另写一次性脚本对基线与新代码跑同一组调用矩阵，比较 stdout/stderr/exit 的原始字节：

| 调用 | exit | stdout | stderr |
|---|---|---|---|
| 全部已设、1 变量 | 0 | 相同 | 相同 |
| 2 个缺失、无 `--quiet` | 1 | 相同 | 相同 |
| 空 rc | 0 | 相同 | 相同 |
| 注释与空行混排、1 个已设 | 1 | 相同 | 相同 |
| 注释与空行混排、全缺失 | 1 | 相同 | 相同 |
| rc 行含空格 | 2 | 相同 | 仅多出路径（第 4 条要求） |
| rc 路径不存在 | 2 | 相同 | 相同 |
| 多余参数 | 2 | 相同 | 仅多出路径（第 4 条要求） |
| `--rc` 后无值 | 2 | 相同 | 相同 |

**预期外差异 0 项。** 受保护的 stdout 与退出码全部逐字节一致；stderr 的两处差异正是 FEATURE.md 第 4 条主动要求的，且只是多出路径前缀，原有文本原样保留。脚本与 worktree 用完即删，未入库。

### 2.3 真实入口核验

四条需求各自在真实 CLI 入口（`python3 -m envcheck.cli`，独立临时目录，子进程）跑过一遍，而不只是靠单元测试：

- 第 1 条：`--rc good.rc --quiet` → `2 of 2 required variables missing`，exit 1；同一 rc 不带 `--quiet` 仍输出两行明细；成功路径 `--quiet` → `all 2 required variables present`，exit 0。
- 第 2 条：`--quiet --rc good.rc` 与 `--rc good.rc --quiet` 的 stdout/stderr/exit 三者全等。
- 第 3 条：含 `FOO=bar` 的 rc → exit 2，`envcheck: eq.rc: line 2: 'FOO=bar' contains '='; list the name only, not a value`。
- 第 4 条：上述消息含路径；`extra` → `envcheck: .envcheckrc: unexpected argument 'extra'`；`--rc` 无值 → `envcheck: --rc needs a path`（声明的例外）。

## 3. 需求逐条对照

| 需求 | 状态 | 证据 |
|---|---|---|
| 1. `--quiet` 只打印汇总行，退出码不变 | 完成 | `test_quiet_prints_only_summary`、`test_quiet_keeps_success_output` |
| 2. `--quiet` 与 `--rc` 顺序任意 | 完成 | `test_quiet_and_rc_order_is_irrelevant`（断言两种顺序三项输出全等） |
| 3. rc 行含 `=` ⇒ 退出码 2，指出哪一行 | 完成 | `test_rejects_equals`、`test_equals_line_exits_two` |
| 4. 所有退出码 2 的错误信息带 rc 路径 | **4/5 个出口完成，1 个显式不满足** | 见下 |

### 第 4 条的未满足项

`--rc` 后面漏了路径时，错误信息**不带** rc 路径，保持 `envcheck: --rc needs a path`。

这不是遗漏，是计划 §3 里就判定为不可满足并写进代码注释的例外：此时缺失的值就是路径本身，填默认 `.envcheckrc` 会是一句假话——工具既没打算读它也没读它，而诊断信息必须指向真实的失败主体（这里是参数，不是文件）。`test_rc_without_path_names_no_file` 把这个例外钉成契约，`cli.py` 的模块注释写明理由。

需要的话有两种处置：接受这个例外（当前状态），或把消息改成同时点明「默认会读 `.envcheckrc`」这类不做虚假断言的措辞。后者是一次新的措辞决定，不在本次授权内。

## 4. 计划里不对或需要调整的地方

### 4.1 提交 1 会红着落地（计划缺一步）

计划 §9 把「`rc.py` + 3 条 parse 用例」切成提交 1，但 §5 把全部 13 条新用例写成一个测试文件，没说这个文件要分两次入库。照字面做，提交 1 会带上依赖 `--quiet` 的 8 条 CLI 用例，`git checkout 610b7d4` 跑测试会红 7 条——一个红着的提交不满足「每个切片可独立验证」。

实际做法：提交 1 只入 parse 层的 3 条用例（12/12 绿），CLI 用例随提交 2 入库（22/22 绿）。计划 §9 应当写明「测试用例随它验证的实现层一起入库」。

### 4.2 用例数与「基线失败数」两个汇总数字写错

计划 §5 说新增 11 条合计 20 条、§10 说 8 条在基线失败。实际新增 13 条合计 22 条，9 条在基线失败。

原因是 §5 的表格里本来就列了 13 行（3 条 parse + 10 条 CLI），11 和 20 是手算汇总时漏掉 parse 表的结果；§10 的「8 条」同样漏掉了 `test_rejects_equals`。**逐条预期全部正确**，错的只是汇总。影响面：如果只按 §10 的「8 条基线失败」验收，会以为多了一条预期外的失败。计划文件**保持原样未回改**，以便这处偏差可被核对；正确的数字是本报告 §2.1 的实测表。

教训是计划里的汇总数字应当由逐条清单推出，不另手算一遍——两处数字不一致时，先修计划再继续，正是 change-evidence 要求的「标注与实际不符说明计划内部有矛盾」。

### 4.3 差分矩阵需要补一组用例

计划 §6 第 3 步列了 8 组调用，但「注释与空行混排」只列了一组。实际跑时发现它只覆盖「有变量已设」的分支，没覆盖同一个 rc 全部缺失时的明细行顺序（去重后 `ALPHA` 保持首见位置）。补成两组，共 9 组。这是对「按首见序、去重语义不变」这条约束更直接的外部证据。

### 4.4 `--quiet` 与 stderr 的关系在计划里是对的，但实现时值得记一笔

计划 §3 判定 `--quiet` 不影响 stderr。实现时确认这是零成本的：`quiet` 只包住 stdout 的明细行循环，所有 `return 2` 的分支都在 `quiet` 被使用之前就返回了。`test_quiet_does_not_suppress_diagnostics` 守住它，防止以后有人把 `--quiet` 误改成全局静音。

## 5. 披露

- **stderr 文本变了。** 第 4 条主动要求的。调用方若在匹配 stderr **全文**（而不是子串），rc 解析错误与多余参数错误这两处会被打破。无法与第 4 条同时满足。原有文本本身原样保留，只是前面多了 `<rc_path>: `。
- **rc 的接受规则收紧了。** 以前 `FOO=bar` 会被当成变量名、走 exit 0/1；现在走 exit 2。这正是第 3 条要求的，但对「rc 里确实写了赋值且依赖旧行为」的调用方是破坏性变更。
- **没有抬任何阈值、没有加豁免、没有跳过任何检查、没有改动既有的 9 条用例。**

## 6. 既有债（本次未修，非本次引入，记录不夹带）

| 问题 | 现状 | 为什么不在本次修 |
|---|---|---|
| rc 行里的制表符 | `ALPHA\tBETA` 的 `strip()` 不去中间的制表符，`" " in line` 也查不到它，于是被当成一个变量名接受 | FEATURE.md 只点了空格与 `=`。修它是收紧公开接口的接受规则，属于独立的行为变更 |
| rc 是二进制文件 | `handle.read()` 抛 `UnicodeDecodeError`，traceback 冒到用户面前，**退出码是 1 而不是 2** | 修它会改公开接口的退出码语义（1 → 2），需要独立授权。这是四项里最值得接着做的一项：它违反了 README 写的「2 = rc 文件不可用」 |
| `cannot read` 与 `line N` 两种消息形状 | 并存 | 前者已满足第 4 条。为形状统一去改一条公开接口消息是夹带 |

## 7. 评审能力缺口（声明降级，不虚报）

本次由单个 agent 执行，委派深度已用尽，**无法起独立上下文做 review-gate 的并行发现（B）与交叉核验（C）**。这两步降级为「作者自审完整 diff + 确定性工具」：全量 unittest、对基线的逐字节差分、真实入口核验。

不声称「多视角已独立评审」。缺口的后果：语义判断类的问题（例如「`--quiet` 该不该也抑制汇总行」这种意图问题）只有我单方面的判断，没有第二个独立上下文检查。

风险判级**中**（计划 §8 的依据不变：范围局限、完全可逆、无敏感面，但改动了公开 CLI 的诊断文本与参数语法）。按中风险的默认路径，合入需要有权的人决定——「实现它，提交到 git」覆盖了「在本仓提交」这个动作与这个候选，未覆盖打标签与发布，故两者都没做。

## 8. 完成声明的分层

- **实施完成**：是。四条需求中三条完整完成，第 4 条有一个声明在案的例外。
- **验证通过**：是，限本地范围——22/22 用例、9 组逐字节差分零预期外差异、四条需求在真实 CLI 入口各自核验过。
- **验收接受**：待你判断，特别是第 4 条的例外与 §5 的两项披露。
- **实际交付**：未发生。未打标签、未推送、未发布。
