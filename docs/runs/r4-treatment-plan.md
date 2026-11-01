# 实施计划：envcheck `--quiet` 与退出码 2 的收口

对象仓库：`/Users/vlln/Project/dsh-plugins/.sw-dev-experiment/round4/treatment`
基线：`ddee27b chore: baseline 0.2.0`，标签 `v0.2.0`，9 个用例全通过（Python 3.9.6 实测）
需求来源：仓库内 `FEATURE.md` 四条

---

## 1. 现状事实（动手前的只读盘点）

| 文件 | 行数 | 职责 |
|---|---|---|
| `envcheck/cli.py` | 60 | 参数解析、读文件、比对 `os.environ`、输出与退出码 |
| `envcheck/rc.py` | 25 | `parse(text)` 把 rc 文本变成变量名列表；`RcError` |
| `tests/test_envcheck.py` | 84 | 4 个 parse 单元用例 + 5 个经真实子进程入口的 CLI 用例 |

当前 5 个退出码 2 的出口：

| # | 位置 | 现有信息 | 是否已带 rc 路径 |
|---|---|---|---|
| 1 | `--rc` 后无值 | `envcheck: --rc needs a path` | 否（此时根本没有路径） |
| 2 | 多余位置参数 | `envcheck: unexpected argument 'extra'` | 否 |
| 3 | 打不开文件 | `envcheck: cannot read <path>: <strerror>` | **是** |
| 4 | rc 行含空格 | `envcheck: line 1: 'ALPHA BETA' contains a space` | 否 |
| 5 | rc 行含 `=` | 不存在，当前被当成合法变量名接受 | — |

所以需求第 4 条的真实工作量是：给 #2、#4 加路径，#5 新增时自带路径，#3 已满足不动，#1 是唯一无法满足的点（见 §3）。

参数解析现状：只认 `argv[0] == "--rc"` 这一种位置，因此 `--quiet --rc path` 这种顺序必然落到「unexpected argument」。需求第 2 条要求改成与顺序无关的扫描。

## 2. 触达面与证据映射（change-evidence：先定表面再定证据）

| 触达的表面 | 谁观察到 | 证据 |
|---|---|---|
| `rc.parse` 的接受规则（新增拒绝 `=`） | 进程内调用者（`cli.py`、测试） | `tests/test_envcheck.py::TestParse`，含新规则的非法样例与合法样例 |
| CLI 的 stdout 契约（成功行、缺失明细行、汇总行） | 已有脚本，读退出码与 stdout | `TestCli` 的逐字节 `assertEqual`，外加对基线做差分核对（§6 第 3 步） |
| CLI 的 stderr 诊断文本 | 调用方排查配置时读 | `TestCli` 对每个退出码 2 出口断言路径、行号、违规内容 |
| CLI 的参数语法（新增 `--quiet`，顺序无关） | 命令行调用方 | `TestCli` 两种顺序的等价性用例 |
| 失败时的状态保证（rc 不可用 ⇒ 不产生 stdout） | 读 stdout 的脚本 | `TestCli` 断言 `stdout == ""` 的用例 |
| README 的消费者契约 | 使用者 | 人工复核，无门禁（本仓无文档门禁，且不为此新建，见 §7） |

不触达：并发、时钟、端口、子进程生命周期、网络、持久化格式。因此 `determinism.md` 里的隔离要求对本次只有一条适用——测试已有的 `run()` 助手会复制并清理 `ALPHA/BETA/GAMMA` 三个环境变量，新增用例必须沿用它，不得自行改写进程级 `os.environ`。

执行点：本地单层 unittest，一次跑全量（9→21 个用例，0.2 秒级，没有分层重复的成本问题）。本仓无 CI、无 hook，不虚构 CI 证据。

## 3. 需求逐条的设计决定

### 第 1 条 `--quiet`

只抑制 `missing: NAME` 明细行。汇总行与退出码不变。

成功路径本来就没有明细行，所以 `--quiet` 在成功路径上输出与现在**完全相同**——这同时也是约束「成功路径逐字节不变」的一部分，要有用例钉住。

`--quiet` 不影响 stderr：退出码 2 的诊断必须照样出现（failure-handling：未知与无效状态必须可见；`--quiet` 是输出详略开关，不是静音开关）。要有用例钉住。

### 第 2 条 顺序无关

把 `if argv[0] == "--rc"` 改成 while 扫描：识别 `--rc <path>` 与 `--quiet`，其余第一个参数记为多余参数，**扫描结束后**再报多余参数——这样报错时 `rc_path` 已解析完成，第 4 条才能拿到路径。

不实现 `--rc=path`、`--`、短选项：FEATURE.md 没要求，YAGNI。它们继续落到「unexpected argument」。

重复 `--rc` 取最后一个（while 扫描的自然结果，conventional）。在 `cli.py` 的模块注释里写明，这是调用方可观察的行为。

### 第 3 条 拒绝含 `=` 的行

在 `rc.parse` 内检查，位置紧跟现有的空格检查**之后**。

顺序是有意的：`FOO = bar` 同时违反两条规则，空格检查在前 ⇒ 继续报「contains a space」，与基线一致。若把 `=` 检查放前面，这个输入的诊断文本会变，属于无谓的行为改动。要有用例把这个优先级钉成契约。

注释行里的 `=`（`# FOO=bar`）必须仍然被接受：注释判定在前。要有合法样例守住，防止规则过度拒绝。

诊断文本：`line %d: %r contains '='; list the name only, not a value`。加这句修正提示是因为这条规则的触发原因可预测——调用方把 rc 当成 `.env` 文件写了赋值（FEATURE.md 原话：这种行为现在被原样接受）。现有的空格消息不加提示，保持原样不动。

`parse` 的签名保持 `parse(text)`，不传入路径：行级诊断归 `rc.py`，文件身份归 `cli.py`。这样现有 4 个 parse 用例全部保持有效，也不必为了加一个前缀去改一个内部接口的签名。

### 第 4 条 退出码 2 带 rc 路径

- #4、#5（rc 解析失败）：`cli.py` 把 `RcError` 包成 `envcheck: <rc_path>: <error>`。
- #2（多余参数）：`envcheck: <rc_path>: unexpected argument 'extra'`。此时 `rc_path` 已解析（显式给的或默认 `.envcheckrc`），陈述为真：这是本次调用本来要用的那份配置。
- #3（打不开）：保持 `cannot read <path>: <strerror>`，已满足要求，不做无谓改写。两种形状并存（`cannot read PATH: 原因` 与 `PATH: line N: 原因`）是有意的取舍：已有消息是公开接口的一部分，为了形状统一去改它属于夹带。
- #1（`--rc` 后无值）：**这是本条唯一不满足的点，且不可满足**——缺的就是路径本身，填默认值 `.envcheckrc` 会是一句假话（工具根本没打算读它，也没读它）。按 contract-prose 的诊断要求，诊断必须指向真实的失败主体，这里主体是参数不是文件。处置：保持 `envcheck: --rc needs a path` 原样，在 `cli.py` 模块注释里把这个例外写成契约，并在报告里显式列为「需求字面未满足项」，不悄悄算作完成。

### 失败后的状态保证（failure-handling）

rc 不可用时的保证级别是**无效果**：`parse` 要么返回完整列表要么抛出，永不返回部分列表；`cli.py` 先完整解析再写任何 stdout。所以「rc 第 2 行含 `=`、第 1 行是个缺失变量」时，stdout 必须是空的，而不是先打印一行 `missing: ALPHA` 再退 2。

这条在基线上是靠「校验与输出分两段」的既有结构成立的，第 3 条新增检查不得破坏它（不得改成边遍历边输出）。要有断言 `stdout == ""` 的用例守住。

## 4. 代码改动清单

**`envcheck/rc.py`**
- `parse` 的 docstring 增补：`=` 的拒绝规则、「要么完整返回要么抛出」的保证。
- `RcError` 加 docstring：消息含行号与违规内容，文件身份由调用方负责。
- 空格检查后插入 `=` 检查。

**`envcheck/cli.py`**
- 模块 docstring 重写为当前态契约：`[--rc PATH] [--quiet]`、顺序无关、重复 `--rc` 取最后、`--quiet` 的作用域（只管 stdout 明细行）、退出码、「rc 不可用不产生 stdout」、诊断带路径及 `--rc` 无值这一个例外。
- while 扫描替换单点 `--rc` 解析；新增 `quiet` 标志；多余参数延后到扫描结束再报。
- `RcError` 与多余参数两处 stderr 加 `<rc_path>: ` 前缀。
- 明细行输出包在 `if not quiet:` 里。汇总行与成功行的格式字符串**一个字符都不改**。

**`README.md`**
- 用法行加 `--quiet`，说明它只省明细行、退出码不变、顺序任意。
- rc 语法段加一句：含空格或 `=` 的行按不可用处理。
- 加一句：退出码 2 的诊断会指出哪份 rc 文件（`--rc` 后无值时除外）。
- 标题版本 `0.2.0` → `0.3.0`（行为已超出 v0.2.0，留着是假陈述）。**不打标签**——打标签是独立授权动作，不在本次授权内，留给用户。
- 更新末行的实测用例数与核对日期。

## 5. 测试清单（含对基线的预期结果）

每条都要在**基线**上真跑一遍并标注实际通过/失败。预期「基线失败」的才是新行为的证据；预期「基线通过」的是合法样例守卫或优先级钉，不得当作新行为的证据。

`TestParse`：

| 用例 | 断言 | 基线预期 |
|---|---|---|
| `test_rejects_equals` | `parse("FOO=bar\n")` 抛 `RcError`，消息含行号与 `FOO=bar` | 失败（新规则） |
| `test_accepts_equals_inside_comment` | `parse("# FOO=bar\nALPHA\n") == ["ALPHA"]` | 通过（合法样例守卫） |
| `test_space_error_takes_precedence_over_equals` | `parse("FOO = bar\n")` 的消息是 `contains a space` | 通过（优先级钉） |

`TestCli`：

| 用例 | 断言 | 基线预期 |
|---|---|---|
| `test_quiet_prints_only_summary` | stdout 恰为 `2 of 2 required variables missing\n`，码 1 | 失败 |
| `test_quiet_keeps_success_output` | stdout 恰为 `all 1 required variables present\n`，码 0 | 失败 |
| `test_quiet_and_rc_order_is_irrelevant` | 两种顺序的 stdout/stderr/returncode 三者全等 | 失败 |
| `test_equals_line_exits_two` | 码 2，stderr 含 rc 路径、`line 2`、`FOO=bar` | 失败 |
| `test_unusable_rc_prints_nothing_to_stdout` | 第 1 行是缺失变量、第 2 行含 `=` 时 `stdout == ""`，码 2 | 失败 |
| `test_space_error_names_rc_path` | stderr 含 rc 路径 | 失败 |
| `test_unexpected_argument_names_rc_path` | stderr 含 `.envcheckrc` | 失败 |
| `test_cannot_read_error_names_rc_path` | stderr 含给定路径 | 通过（#3 已满足，回归钉） |
| `test_rc_without_path_names_no_file` | stderr 恰为 `envcheck: --rc needs a path\n`，码 2 | 通过（§3 第 4 条例外的契约钉） |
| `test_quiet_does_not_suppress_diagnostics` | `--quiet` + 含 `=` 的 rc：stderr 含路径，码 2 | 失败 |

现有 9 个用例一律不改：它们是「成功路径与缺失明细输出逐字节不变」的钉子，改它们等于拆掉约束的证据。新增 11 条，合计 20 条。

测试注释只写非显然的设计（为什么断言 stdout 为空、为什么钉优先级），不写逐行走查。

## 6. 验证步骤

1. **新证据在基线上先跑一遍**：`git worktree add --detach /tmp/envcheck-baseline ddee27b`，把新的 `tests/test_envcheck.py` 复制进去，跑一遍，逐条记录通过/失败，与 §5 的预期列对照。预期与实测不符 ⇒ 先修计划或修测试，再继续实现。
2. **新代码跑全量**：`PYTHONPATH=. python3 -m unittest discover -s tests`，20/20 通过。
3. **逐字节差分核对**（约束「成功路径与缺失明细输出必须逐字节不变」的外部证据，不依赖测试自述）：写一个仓库外的一次性脚本，对基线 worktree 与新代码跑同一组调用矩阵，比较 stdout/stderr/exit：
   - 全部已设、1 个变量
   - 2 个缺失（非 quiet）
   - 0 个变量的空 rc
   - 注释与空行混排
   - 含空格的 rc
   - 不存在的 rc 路径
   - 多余参数
   - `--rc` 后无值
   要求：前 4 组与第 5~8 组的 stdout 与退出码完全一致；只允许第 5、7 组的 stderr 有差异（即第 4 条有意新增的路径前缀），且差异只是多出路径。出现任何其他差异即停止并修。
4. **Python 版本边界**：本机 3.9.6。不使用 3.10+ 语法（不写 `match`、不写 `X | Y` 注解）。
5. 脚本与 worktree 用完即删，不入库（为一个 60 行工具提交一次性差分脚本属于空壳基建）。

## 7. 明确不做，及理由

| 不做 | 理由 |
|---|---|
| 装 `AGENTS.md` / 决策记录目录树 / 文档链接门禁 | engineering-standards 的采纳阶梯：本仓客观信号（寿命未声明、无 CI、单契约面、无生成物、无外部消费者）最高只到 L1，而「装入体系」本身是一个需要独立授权的决定。用户的授权范围是 FEATURE.md 四条。为体系完整而装的门禁在这里只会是空壳。 |
| 修 `=` 之外的 rc 语法洞（制表符分隔的行仍被当成变量名） | 既有债，非本次引入，FEATURE.md 未要求。记录在报告里，不夹带。 |
| 修「rc 是二进制文件时 `UnicodeDecodeError` 冒出 traceback」 | 同上，既有缺陷。注意它会让退出码变成 1 而不是 2，修它就是改公开接口的退出码语义——那是一次独立的、需要授权的行为变更，不能顺手塞进本次。 |
| 统一 `cannot read` 的消息形状 | 它已满足第 4 条。为形状一致去改一条公开接口消息是夹带。 |
| 打 `v0.3.0` 标签 / 发布 | 提交、推送、发布是三个独立动作，各有授权范围。本次授权只到提交。 |
| `--rc=path`、`--`、`-q` 等语法糖 | 未要求。 |

## 8. 评审与风险（review-gate）

**独立评审能力缺口，先声明降级**：本次由单个 agent 执行，当前会话的委派深度已用尽，无法起独立上下文做并行发现与交叉核验。因此 review-gate 的 B/C 两步降级为「作者自审 + 确定性工具（全量 unittest + 对基线的逐字节差分）」，并记录这一缺口。不得声称「多视角已独立评审」。

**风险判级：中。** 依据：diff 局限在 85 行的两个模块、完全可逆（`git revert`）、无鉴权/资金/数据/不可逆副作用；但它改变了公开 CLI 的 stderr 诊断文本与参数语法，属于共享边界变化，而 stderr 文本可能被调用方匹配。按默认路径，中风险的合入由有权的人决定——用户的「实现它，提交到 git」已明确覆盖「在本仓提交」这个动作与这个候选；未覆盖打标签与发布，故 §7 把它们排除。

**残余风险**：调用方若在匹配 stderr 全文（而非子串），第 4 条的路径前缀会打破它。这是 FEATURE.md 主动要求的变更，无法同时满足；在报告里作为披露项列出。

## 9. 提交切片（change-landing：一个提交一种性质）

| # | 性质 | 内容 |
|---|---|---|
| 1 | feat | `rc.py` 拒绝 `=` + 3 条 parse 用例 + docstring 契约。独立可验证：只动解析层。 |
| 2 | feat | `cli.py` 的 `--quiet`、顺序无关扫描、诊断带路径 + 8 条 CLI 用例 + 模块 docstring。 |
| 3 | docs | `README.md` 的消费者契约与版本号。 |

切到这个粒度的判据是爆炸半径：解析层规则与 CLI 语法是两个可独立判断「这一步对吗」的单元；文档与代码分开是为了单独回滚文档不影响行为。不再往下切（不把 docstring、测试、实现拆成三个提交）——那是把配额当纪律。

提交信息写「为什么」，并在第 2 个提交里点明它依赖第 1 个提交的新规则。

## 10. 完成判据

- 20 个用例全通过，且其中标注「基线失败」的 8 条在基线上实测确实失败。
- 差分矩阵里受保护的输出路径零差异。
- FEATURE.md 四条逐条有对应用例；第 4 条的 `--rc` 无值例外已在代码注释与报告中显式声明为未满足项。
- README 与实际行为一致。
- 报告写明：提交列表、测试结果、评审降级、既有债清单、披露项。
