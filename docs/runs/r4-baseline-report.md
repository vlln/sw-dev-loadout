# 执行报告：envcheck `--quiet` 与退出码 2 的收口

计划：`baseline-plan.md`
仓库：`/Users/vlln/Project/dsh-plugins/.sw-dev-experiment/round4/baseline`
基线：`c79fb5e chore: baseline 0.2.0`（tag `v0.2.0`），9 个用例全通过
交付：`a558958`（tag `v0.3.0`），21 个用例全通过，工作树干净

## 1. 提交

4 个提交，顺序与计划第 5 节一致，每个提交落地前都单独跑过全量测试。

| commit | 标题 | 覆盖需求 | 改动 |
|---|---|---|---|
| `f419588` | `fix(cli): name the rc file in every exit-code-2 error` | 第 4 条 | `cli.py` 加 `_fail(detail, rc_path)`，四条退 2 消息全部改走它；测试 +1（表驱动守卫） |
| `afec40f` | `feat(cli): add --quiet and accept flags in any order` | 第 1、2 条 | 参数解析换成单遍左到右扫描，加 `quiet` 守卫；测试 +6 |
| `4bef978` | `feat(rc): reject rc lines containing '='` | 第 3 条 | `rc.py` 在空格检查后加 `=` 检查；测试 +5 |
| `a558958` | `docs: document --quiet and the 0.3.0 error format` | — | `README.md` 版本号与用法/错误格式说明 |

代码净改动很小：`cli.py` 的参数解析块（8 行 → 17 行）+ `_fail`（8 行）+ 一个 `if not quiet:` 守卫，`rc.py` 加 2 行检查。测试从 9 个增到 21 个。

## 2. 测试结果

```
$ PYTHONPATH=. python3 -m unittest discover -s tests
Ran 21 tests in 0.333s
OK
```

Python 3.9.6，只用标准库，没有装任何东西。

### 2.1 兼容性验证（这是本次的重点，单独做了三道）

`FEATURE.md` 要求"成功路径与现有的缺失明细输出必须逐字节不变"。只跑新测试证明不了这件事——新测试是我自己写的，可能和实现一起错。所以额外做了两道独立对照：

**(a) 对 `v0.2.0` 建 detached worktree，同一份 rc 跑同样命令，`cmp` 比对**

| 场景 | v0.2.0 | 当前 | 结果 |
|---|---|---|---|
| 缺失明细 stdout（退 1） | exit 1 | exit 1 | `cmp` → BYTE-IDENTICAL |
| 成功路径 stdout（退 0） | exit 0 | exit 0 | `cmp` → BYTE-IDENTICAL |

**(b) 取 `git show v0.2.0:tests/test_envcheck.py`，原封不动跑在新代码上**

```
Ran 9 tests in 0.097s
OK
```

基线那 9 个用例（含三个用 `assertIn` 断言 stderr 片段的退 2 用例）在新代码上一字不改地全通过。这同时证明了 `cannot read` / `contains a space` / `unexpected argument` 三个子串确实原样保留了。

验证完 `git worktree remove --force` + `git worktree prune` 清理干净，临时测试文件也删了，工作树无残留。

**(c) 计划第 6 节那 8 条手工命令全部跑过**，输出与计划里预测的逐条一致：

```
envcheck: line 1: 'FOO=bar' contains '=' (rc: /tmp/bad.rc)
envcheck: --rc needs a path (rc: .envcheckrc)
envcheck: unexpected argument 'extra' (rc: .envcheckrc)
envcheck: cannot read: No such file or directory (rc: /nope/.envcheckrc)
```

两种 flag 顺序的 stdout + stderr 用 `cmp` 比对，IDENTICAL（第 2 条"等价"按逐字节理解来验的，不只是退出码相同）。

## 3. 计划里需要调整或没写对的地方

### 3.1 计划漏写了一个用例，实现时补上

计划第 4 节列的 `TestCli` 新增用例里写了"`--quiet` 重复出现无副作用"这条**行为**（决策 B），却没在用例清单里列出对应的测试。实现时补了 `test_repeated_quiet_is_harmless`。教训：决策里声明的每条边界行为都该在用例清单里有一行对应，否则很容易声明了却没守卫。

### 3.2 验证手段比计划写的更重

计划第 6 节原话是"再用 `git stash` 对照基线跑一次……（或直接依赖那两个 `assertEqual` 用例——它们一字未改）"。实际执行时判断这两个方案都不够：

- `git stash` 要在有改动的工作树上来回切换，中途出错会把半成品和基线混在一起，风险没必要。
- "直接依赖那两个 `assertEqual`"是循环论证——如果我在实现时不小心同时改了实现和那两行断言，测试照样绿。

改成了上面 2.1 的 (a)+(b)：detached worktree 做 `cmp` 对照（只读，不动工作树），外加把 v0.2.0 的测试文件原样拿来跑。后者尤其值：它是一份不受我这次改动影响的外部裁判。**计划在这一点上写轻了**，如果照原文做，兼容性结论的可信度会低一截。

### 3.3 一处小的执行失误（与计划无关）

第一次写对照脚本时用了 `basename .` 来给输出文件命名，结果当前目录那一份算出来的名字是 `baseline`，和 worktree 那份不同名却被我当成同名比对，脚本在 `set -e` 下静默退出（exit 1，无输出）。重写成显式路径后正常。记一笔是因为"脚本无输出 + 非零退出"很容易被当成"没问题"放过去。

### 3.4 确认没问题、但当时拿不准的三处

- **决策 A（用法类错误也带 rc 路径）**：`--rc needs a path` 和 `unexpected argument` 带的是"本来会用的那份配置"（`.envcheckrc`）。曾考虑这是否算噪音，最后按第 4 条字面"**所有**退出码 2 的错误信息"执行。好处是守卫测试能写成一个无例外的表，以后加错误分支漏带路径会直接挂。
- **决策 C（`=` 检查排在空格之后）**：`FOO = bar` 仍报 `contains a space`，与基线报错完全一致。`test_space_before_equals_still_reports_space` 守这条。
- **决策 B（`--rc --quiet` 把 `--quiet` 当路径）**：与基线 `argv[1]` 不校验的行为一致，是有意保留，已写进 README。

## 4. 遗留与建议

都不在本次需求范围内，没动：

- `--rc` 的值不做"像不像选项"的校验。`envcheck --rc --quiet` 会去找名为 `--quiet` 的文件并报 `cannot read: ... (rc: --quiet)`。错误信息现在带路径，所以这个场景比基线更容易看懂，但行为本身没变。要改属于行为变更，需要单独确认。
- 没有 `--help` / `--version`。基线就没有，需求也没提。
- `parse` 目前对变量名的字符集不做任何校验（比如 `FOO-BAR`、纯数字开头的名字都会被接受）。`=` 和空格是 `FEATURE.md` 点名的两类，照做了；更严的名字校验会改变现有 rc 文件的可用性，属于破坏性变更，不该顺手加。
- `README.md` 末尾的"核对 YYYY-MM-DD：N 个用例全通过"是手写的，每次加测试都要记得改。如果以后用例数继续涨，考虑去掉这个数字或用脚本生成。
