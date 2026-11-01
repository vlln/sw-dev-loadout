# 实施计划：envcheck `--quiet` 与退出码 2 的收口

目标仓库：`/Users/vlln/Project/dsh-plugins/.sw-dev-experiment/round4/baseline`
基线：`c79fb5e chore: baseline 0.2.0`（tag `v0.2.0`），9 个用例全通过，Python 3.9.6，只用标准库。

需求见仓库里的 `FEATURE.md`（四条）。本计划把它拆成 4 个可独立验证的提交。

---

## 1. 现状

三个文件，总共不到 100 行：

- `envcheck/rc.py`：`parse(text) -> list[str]`。按行 `strip()`，跳过空行与 `#` 开头行，行内有空格则 `raise RcError("line %d: %r contains a space")`，其余按首见序去重后返回。
- `envcheck/cli.py`：`main(argv=None) -> int`。**只在 `argv[0] == "--rc"` 时**识别 `--rc`，取 `argv[1]` 作路径并切掉前两项；之后若还有残余参数就报 `unexpected argument` 退 2。然后读文件（`OSError` → 退 2）、`parse`（`RcError` → 退 2），再逐个打印 `missing: NAME`，最后打印汇总行，有缺失退 1，否则退 0。
- `tests/test_envcheck.py`：4 个 `parse` 单测 + 5 个 CLI 子进程端到端用例。CLI 用例对**成功路径和缺失明细用 `assertEqual` 全量比对 stdout**，对三个退 2 用例用 `assertIn` 比对 stderr 片段（`"cannot read"` / `"contains a space"` / `"unexpected argument"`）。

现状里和需求直接相关的四个事实：

1. 没有 `--quiet`。
2. `--rc` 只能出现在**第一个**位置，`envcheck --quiet --rc path` 这种顺序在现有解析器里根本走不到 `--rc` 分支。
3. `FOO=bar` 这样的行不含空格，会被 `parse` 原样当成变量名接受。
4. 四条退 2 的错误信息里，只有 `cannot read %s: %s` 带了 rc 路径；`RcError` 透传、`--rc needs a path`、`unexpected argument %r` 都不带。

## 2. 兼容性边界（先定清楚，实现时不要再放宽）

`FEATURE.md` 的约束 + 现有测试共同构成这道边界：

- **必须逐字节不变**：
  - 成功路径 stdout：`"all %d required variables present\n"`，退 0。
  - 缺失明细 stdout：`"missing: %s\n"` 每行 + `"%d of %d required variables missing\n"` 汇总，退 1。
  - `parse` 的公开签名与返回值语义（按首见序去重、空行/注释跳过）。
  - `main(argv=None)` 的签名与"返回 int 退出码"的约定。
  - 四类错误的退出码仍是 2。
- **允许改**：退 2 的 **stderr 文案**。`FEATURE.md` 只要求成功路径和缺失明细逐字节不变，而第 4 条明确要求改错误文案；两者不冲突。
- **但要保留三个子串**：`cannot read`、`contains a space`、`unexpected argument`。现有测试用 `assertIn` 依赖它们，它们也是调用方最可能 grep 的片段。改文案时让这三个子串原样留在句子里，不要改写成同义表达。

## 3. 设计决策

### 决策 A：退 2 的错误信息统一成一个后缀格式

所有退 2 的 stderr 走同一个辅助函数，格式固定为：

```
envcheck: <detail> (rc: <rc_path>)
```

`<rc_path>` 是**该时刻生效的 rc 路径**（默认 `.envcheckrc`，若已解析到 `--rc` 则为其值）。四条消息变成：

| 场景 | 新 stderr |
|---|---|
| `--rc` 缺值 | `envcheck: --rc needs a path (rc: .envcheckrc)` |
| 未知参数 | `envcheck: unexpected argument 'extra' (rc: .envcheckrc)` |
| 文件读不了 | `envcheck: cannot read: No such file or directory (rc: /nonexistent/.envcheckrc)` |
| rc 内容非法 | `envcheck: line 1: 'ALPHA BETA' contains a space (rc: /tmp/x.rc)` |

理由：第 4 条说的是"**所有**退出码 2 的错误信息"。与其给 rc 类错误和用法类错误两套格式、再逐条判断哪条该带路径，不如一条规则到底——只有一个出口函数，"是否带路径"就不可能漏。代价是 `cannot read` 那条的形状变了（路径从句中挪到句尾，不再重复出现），这在允许改的范围内，且 `cannot read` 子串保留。

用法类错误（`--rc needs a path`、`unexpected argument`）带的是"**本来会用的那份配置**"。对调用方而言这不是噪音：它说明"这次运行指向的是哪份 rc"，与第 4 条的意图（让调用方知道是哪份配置出的问题）一致。

### 决策 B：参数解析改成单遍左到右扫描

```
index = 0
while index < len(argv):
    arg = argv[index]
    --rc    → 取 argv[index+1] 作 rc_path（缺值则退 2），index += 2
    --quiet → quiet = True，index += 1
    其它    → 退 2（unexpected argument）
```

由此确定的几条边界行为，实现时按这个来，并写进 README：

- `--quiet` 与 `--rc` 顺序任意，两种顺序**输出逐字节相同**（第 2 条）。
- `--rc` 重复出现时**最后一个生效**（左到右覆盖）。
- `--quiet` 重复出现无副作用。
- `--rc` 的取值**按字面消费**，不检查它是否长得像选项：`envcheck --rc --quiet` 的 rc 路径就是字符串 `--quiet`（与基线行为一致，基线 `argv[1]` 同样不做检查）。
- 遇到第一个非法 token 立即报错，**不继续往后扫**。所以 `envcheck extra --rc /tmp/x` 报的是 `(rc: .envcheckrc)` 而非 `/tmp/x`。这是标准 CLI 行为，不为了"把路径找全"而引入两遍扫描。
- 不实现 `--` 终止符、不实现 `--rc=PATH` 等价形式：需求没要求，YAGNI。

### 决策 C：`=` 的检查放在 `rc.parse` 里，且排在空格检查之后

放在 `parse` 里而不是 CLI 里：这是 rc 文件的语法规则，和空格规则同级，CLI 只负责把 `RcError` 加上路径后输出。`parse(text)` 的签名因此不变——**不要**给它加 `path` 参数，测试和任何外部调用都直接导入它。

排在空格检查**之后**：`FOO = bar` 同时含空格和 `=`，保持空格检查优先，这类行的报错文案就和基线完全一致（只有"新非法"的 `FOO=bar` 才报新错）。改动面最小。

文案与现有风格对齐，带行号与行内容（第 3 条要求"指出是哪一行"）：

```python
raise RcError("line %d: %r contains '='" % (lineno, line))
```

注释行里的 `=`（`# FOO=bar`）不受影响——注释在检查之前就 `continue` 了。

### 决策 D：`--quiet` 只关掉明细循环

`quiet` 只用来跳过 `missing: NAME` 那个 for 循环，**不影响**汇总行、不影响成功路径、不影响 stderr 上的错误信息、不影响退出码。这样成功路径的 stdout 在 `--quiet` 下也与基线逐字节相同（第 1 条"退出码不变"+ 约束"成功路径逐字节不变"自然满足）。

## 4. 落地改动

### `envcheck/rc.py`

在空格检查后插入：

```python
        if "=" in line:
            raise RcError("line %d: %r contains '='" % (lineno, line))
```

同步更新 `parse` 的 docstring，说明含空格或 `=` 的行会 `raise RcError`。

### `envcheck/cli.py`

1. 加模块级辅助函数：

```python
def _fail(detail, rc_path):
    """Write one exit-code-2 error to stderr and return 2.

    Every exit-code-2 message goes through here, so all of them name the
    rc file that was in effect.
    """
    sys.stderr.write("envcheck: %s (rc: %s)\n" % (detail, rc_path))
    return 2
```

2. `main` 的参数解析替换成决策 B 的 while 循环，`quiet` 默认 `False`。
3. 读文件与 `parse` 的两个 `except` 改走 `_fail`：
   - `OSError` → `_fail("cannot read: %s" % error.strerror, rc_path)`
   - `RcError` → `_fail(str(error), rc_path)`
4. 明细循环加 `if not quiet:` 守卫。汇总行与返回值完全不动。
5. 更新模块 docstring 的 Usage 行为 `envcheck [--quiet] [--rc PATH]`。

目标形态（`main` 主体）：

```python
    rc_path = RC_DEFAULT
    quiet = False

    index = 0
    while index < len(argv):
        arg = argv[index]
        if arg == "--rc":
            if index + 1 >= len(argv):
                return _fail("--rc needs a path", rc_path)
            rc_path = argv[index + 1]
            index += 2
            continue
        if arg == "--quiet":
            quiet = True
            index += 1
            continue
        return _fail("unexpected argument %r" % arg, rc_path)

    try:
        with open(rc_path) as handle:
            text = handle.read()
    except OSError as error:
        return _fail("cannot read: %s" % error.strerror, rc_path)

    try:
        names = parse(text)
    except RcError as error:
        return _fail(str(error), rc_path)

    missing = [name for name in names if name not in os.environ]

    if not quiet:
        for name in missing:
            sys.stdout.write("missing: %s\n" % name)

    if missing:
        sys.stdout.write("%d of %d required variables missing\n" % (len(missing), len(names)))
        return 1

    sys.stdout.write("all %d required variables present\n" % len(names))
    return 0
```

### `tests/test_envcheck.py`

现有 9 个用例一字不改（它们就是兼容性回归网）。新增：

`TestParse`
- `test_rejects_equals`：`parse("FOO=bar\n")` 抛 `RcError`，消息里有 `line 1` 和 `'='`。
- `test_rejects_equals_reports_line_number`：`parse("ALPHA\nBETA\nFOO=bar\n")` 的消息里有 `line 3`。
- `test_equals_in_comment_is_allowed`：`parse("# FOO=bar\nALPHA\n") == ["ALPHA"]`。
- `test_space_before_equals_still_reports_space`：`parse("FOO = bar\n")` 的消息含 `contains a space`（决策 C 的顺序）。

`TestCli`
- `test_quiet_suppresses_missing_detail`：`--quiet` + 2 个缺失 → 退 1，stdout 恰为 `"2 of 2 required variables missing\n"`。
- `test_quiet_keeps_success_output_byte_identical`：`--quiet` 成功路径 stdout 与不带 `--quiet` 时**相同**（直接比对两次运行的 stdout）。
- `test_quiet_and_rc_order_independent`：`["--quiet", "--rc", p]` 与 `["--rc", p, "--quiet"]` 的 returncode / stdout / stderr 三者全等。
- `test_last_rc_wins`：`["--rc", bad, "--rc", good]` 用后一个。
- `test_equals_line_exits_two`：CLI 层 `=` 行 → 退 2，stderr 含 `line 1`、`'='` 和 rc 路径。
- `test_every_exit_two_error_names_rc_path`：把四类退 2 场景放进一个表驱动循环，逐个断言 `returncode == 2` 且 `rc 路径 in stderr`——这是第 4 条的直接守卫，以后新增错误分支漏带路径会在这里挂。
- `test_unexpected_argument_after_quiet_exits_two`：`["--quiet", "extra"]` 仍退 2（确认新解析器没把未知参数吞掉）。

注意：`run()` 辅助函数固定 `cwd=ROOT`，而仓库根目录**没有** `.envcheckrc`，所以"不带 `--rc` 时报默认路径"的用例可以依赖 `.envcheckrc` 读不到。断言只查 `(rc: .envcheckrc)` 子串，不要去断言绝对路径。

### `README.md`

- 标题版本 `0.2.0` → `0.3.0`（新增了向后兼容的 flag，语义化版本取 minor）。
- 用法行改为 `envcheck [--quiet] [--rc PATH]`，并说明：`--quiet` 只去掉缺失明细行、保留汇总行、退出码不变；两个 flag 顺序任意；`--rc` 重复时最后一个生效。
- rc 格式一节补一句：行内含空格或 `=` 时该 rc 文件按不可用处理，退 2。
- 退出码一节补一句：退 2 的错误信息一律写到 stderr 并带上 `(rc: PATH)`。
- 更新末尾的"核对"行为新的用例数。

## 5. 提交顺序

每个提交都必须**自身测试全绿**（跑一遍第 6 节的命令再提交）。按这个顺序，后一个提交依赖前一个的辅助函数：

| # | 范围 | 内容 |
|---|---|---|
| 1 | `cli.py` + 测试 | 引入 `_fail`，四条退 2 消息全部带 `(rc: PATH)`；加表驱动守卫用例。纯错误文案收口，不碰参数解析（第 4 条） |
| 2 | `cli.py` + 测试 | 单遍扫描的参数解析 + `--quiet`（第 1、2 条） |
| 3 | `rc.py` + 测试 | 拒绝含 `=` 的行（第 3 条） |
| 4 | `README.md` | 文档与版本号 |

提交信息用 Conventional Commits（与基线 `chore: baseline 0.2.0` 一致）：

```
fix(cli): name the rc file in every exit-code-2 error
feat(cli): add --quiet and accept flags in any order
feat(rc): reject rc lines containing '='
docs: document --quiet and the 0.3.0 error format
```

全部完成后打 `v0.3.0` 标签（仓库已有 `v0.2.0` 的惯例）。

## 6. 验证

```sh
cd /Users/vlln/Project/dsh-plugins/.sw-dev-experiment/round4/baseline
PYTHONPATH=. python3 -m unittest discover -s tests
```

除了测试，手工跑一遍下面这组，确认输出与计划一致（尤其前两条的"逐字节相同"）：

```sh
printf 'ALPHA\nBETA\n' > /tmp/ec.rc
PYTHONPATH=. python3 -m envcheck.cli --rc /tmp/ec.rc            # 退 1，明细 2 行 + 汇总
PYTHONPATH=. python3 -m envcheck.cli --rc /tmp/ec.rc --quiet    # 退 1，只有汇总
PYTHONPATH=. python3 -m envcheck.cli --quiet --rc /tmp/ec.rc    # 与上一条逐字节相同
ALPHA=1 BETA=2 PYTHONPATH=. python3 -m envcheck.cli --rc /tmp/ec.rc --quiet  # 退 0，成功行不变
printf 'FOO=bar\n' > /tmp/bad.rc
PYTHONPATH=. python3 -m envcheck.cli --rc /tmp/bad.rc           # 退 2，含 line 1、'='、路径
PYTHONPATH=. python3 -m envcheck.cli --rc                       # 退 2，含 (rc: .envcheckrc)
PYTHONPATH=. python3 -m envcheck.cli extra                      # 退 2，含 unexpected argument + 路径
PYTHONPATH=. python3 -m envcheck.cli --rc /nope/.envcheckrc     # 退 2，含 cannot read + 路径
```

再用 `git stash` 对照基线跑一次成功路径与缺失明细，确认这两条 stdout 没有任何差异（或直接依赖那两个 `assertEqual` 用例——它们一字未改）。

## 7. 风险与注意

- **最容易踩的坑**：为了给 `cannot read` 加路径而重写整句时顺手把 `cannot read` 改成别的说法，直接打挂现有测试，也打挂调用方的 grep。三个子串（`cannot read` / `contains a space` / `unexpected argument`）必须原样保留。
- **第二个坑**：把 `=` 检查写在空格检查之前，`FOO = bar` 的报错就从"空格"变成"="，属于无谓的行为变更。
- **第三个坑**：给 `parse` 加 `path` 参数来满足第 4 条。`parse` 是公开接口且被测试直接导入，路径应该由 CLI 在格式化时补上。
- `--rc` 的值不做"像不像选项"的校验，是**有意**与基线保持一致，不是遗漏；写在 README 里。
- Python 3.9.6：不要用 3.10+ 的 `match`、`X | Y` 类型写法。代码里本来也没有类型标注，保持原样。
