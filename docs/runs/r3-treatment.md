# tally 0.4.0 实施报告

> 计划：`round2/with-set-plan.md`（下称"计划"）
> 仓库：`round3/repo/`（基线 `9e91cb1 chore: baseline 0.3.0`）
> 交付：`pr1`/`pr2`/`pr3`/`pr4` 四个分支，`main` 指向 `pr4`；基线锚点标签 `v0.3.0`
> 结论：计划的主干（先钉字节、再动失败路径、stdin 只在高危点小心、`--json` 共用同一份有序结果）**成立且好用**；但计划有 **4 处内部矛盾、3 处范围错位、2 处证据论证不成立**，还有 1 处把"不可验证的表面"当成了已验证（`bin` 映射）。逐条见第 3 节。

---

## 1. 四个分支各改了什么

| 分支 | 提交 | 性质 | 改动文件 | 行为变化 |
|---|---|---|---|---|
| `pr1` | `31b91a0 refactor: split cli.mjs into side-effect-free units` | 重构 | `src/{args,ext,count,render,cli}.mjs`、`test/cli.test.mjs`（新增） | **无**（12/12 成功 fixture stdout+stderr+退出码逐字节一致） |
| `pr2` | `65d304e fix(cli): readable input errors, graded exit codes, -- terminator` | 失败路径修复 | `src/args.mjs`、`src/{validate,cli}.mjs`、`test/cli.test.mjs` | 仅失败路径：可读错误、报告全部坏路径、未知 flag 1→2、`--` |
| `pr3` | `5cc768c feat(cli): read standard input when a path is "-"` | 新功能 | `src/{args,cli,count,ext,stdin}.mjs`、`test/cli.test.mjs` | 纯新增：`-` 读 stdin、聚合、重复 `-` 用法错误、空 stdin 错误 |
| `pr4` | `b0d9406 feat(cli): add --json, document the contract, bump to 0.4.0` | 新功能 + 文档 + 版本 | `README.md`、`package.json`、`src/{args,cli,render}.mjs`、`test/cli.test.mjs` | 纯新增：`--json`、README 契约、0.3.0→0.4.0 |

### 各分支的关键决定（写"为什么"）

**PR1 重构。** 拆成 `args`（纯函数：argv → 有序 inputs / 用法错误）、`ext`（扩展名派生 + 二进制后缀规则的唯一出处）、`count`（`countText`/`aggregate`/`countFiles`，签名语义不变）、`render`（有序行 → 表格字节）、`cli`（唯一碰 stdout/stderr/exitCode 的薄入口，async）。`test/count.test.mjs` **一行未改**。同时加了回归钉：T1–T6、T16，外加计划没列的 4 条（扩展名 quirk、`(none)`、无参数 usage、**二进制后缀名不碰文件系统**）。最后一条是故意的：它是 PR2 的真实陷阱（见 P4）。

**PR2 失败契约。** 两阶段：`validatePaths` 先 stat + access，收集全部坏路径，一次报完；读阶段再 catch 一次（校验后文件消失的 TOCTOU）；渲染永远最后，所以失败时 stdout 0 字节可证。退出码 `0/1/2`。`--` 终止 flag 解析。**校验跳过二进制后缀名**，因为冻结项 6 要求 `missing.png` 静默成功。

**PR3 stdin。** `-` 进 `inputs`，`readStdin` 走 `process.stdin` 流（不按 fd 0 同步读），只在 `inputs` 含 stdin 时才调用。stdin 与文件在 **argv 顺序**里交错聚合，`countFile` 是文件侧的单输入入口，`STDIN_EXT = '(none)'`。空 stdin → exit 1 可读消息；重复 `-` → exit 2。

**PR4 `--json`。** `renderJson(rows)` 与 `renderTable(rows)` 吃**同一个** `rows`；`Object.fromEntries` 保序（所有标签以 `.` 开头或是 `(none)`，不是整数样键）。README 重写成消费者契约，每条示例都真跑过（8/8 逐字节一致）。版本 0.4.0，`bin`/`engines`/包名不动。

### `git log --oneline --graph --all` 的真实输出

```
$ git log --oneline --graph --all
* b0d9406 feat(cli): add --json, document the contract, bump to 0.4.0
* 5cc768c feat(cli): read standard input when a path is "-"
* 65d304e fix(cli): readable input errors, graded exit codes, -- terminator
* 31b91a0 refactor: split cli.mjs into side-effect-free units
* 9e91cb1 chore: baseline 0.3.0
```

带装饰（含标签与分支指针）：

```
$ git log --oneline --graph --all --decorate
* b0d9406 (HEAD -> main, pr4) feat(cli): add --json, document the contract, bump to 0.4.0
* 5cc768c (pr3) feat(cli): read standard input when a path is "-"
* 65d304e (pr2) fix(cli): readable input errors, graded exit codes, -- terminator
* 31b91a0 (pr1) refactor: split cli.mjs into side-effect-free units
* 9e91cb1 (tag: v0.3.0) chore: baseline 0.3.0
```

```
$ git rev-parse main pr1 pr2 pr3 pr4 v0.3.0
b0d9406d7cf6856667138e2b7477428e26c29754   main == pr4
31b91a0719438d1ea2fb688fe99c7b061dc18539   pr1
65d304eef6544c0da84bf2df2086cc791f229615   pr2
5cc768c7792f7eab65061b7a69af49ed31ee48d9   pr3
b0d9406d7cf6856667138e2b7477428e26c29754   pr4
bf3935c740c9a25751393f6309a5e40c1b237a81   v0.3.0
```

工作树干净（`git status --short` 空）；`git diff 9e91cb1 main -- test/count.test.mjs` 为空。

### 每个提交的 diff 规模

```
31b91a0  src/args.mjs +11 | src/cli.mjs ±27 | src/count.mjs ±29 | src/ext.mjs +19
         src/render.mjs +15 | test/cli.test.mjs +176        (6 files, +261/−16)
65d304e  src/args.mjs ±24 | src/cli.mjs ±35 | src/validate.mjs +51
         test/cli.test.mjs ±79                             (4 files, +183/−6)
5cc768c  src/args.mjs ±34 | src/cli.mjs ±61 | src/count.mjs ±12 | src/ext.mjs ±9
         src/stdin.mjs +16 | test/cli.test.mjs +70         (6 files, +175/−27)
b0d9406  README.md ±102 | package.json ±2 | src/args.mjs ±14 | src/cli.mjs ±6
         src/render.mjs +8 | test/cli.test.mjs +69         (6 files, +193/−8)
```

---

## 2. `npm test` 与自写用例的真实输出

### 2.1 最终 `npm test`（`main` = `pr4`）

```
$ cd round3/repo && npm test
（npm 自己的 node 进程会先打一行 UNDICI-EHPA 警告；这是本机 NODE_USE_ENV_PROXY=1 的环境噪声，
  被测子进程的环境已在测试里净化，见 2.2）
ok 1 - T1: exact table bytes for two files plus a longer extension
ok 2 - T2: ties keep first-appearance order
ok 3 - T3: an empty file counts as one line
ok 4 - T4: the same path twice is counted twice
ok 5 - T5: binary-suffix files are silently left out
ok 6 - T6: when every input is skipped the table is empty and exit is 0
ok 7 - pin: extension label comes from the whole path string, dots included
ok 8 - pin: paths without a dot use the (none) label
ok 9 - pin: no arguments is a usage error on stderr with exit 2
ok 10 - pin: a skipped name is skipped before the filesystem is consulted
ok 11 - T16: without - the CLI never reads stdin (stdin left open)
ok 12 - T14: a missing path exits 1 with something readable, stdout empty
ok 13 - T15: every bad path is reported in one run
ok 14 - T18: an unknown flag is a usage error (2), not a path
ok 15 - T18b: an unknown flag alone is still a usage error
ok 16 - T20: -- lets a flag-shaped file name through
ok 17 - T21: an unreadable file exits 1 with a readable message
ok 18 - pin: a directory argument is an input error, not a crash
ok 19 - T11: "-" reads stdin and counts by the same convention
ok 20 - T12: stdin aggregates with a file in the same (none) bucket
ok 21 - T12b: stdin keeps its (none) label next to an extension file
ok 22 - T12c: a stdin/file tie is ordered by argv position
ok 23 - T17: empty stdin is an input error with an empty stdout
ok 24 - T19: a repeated "-" is a usage error
ok 25 - pin: a bad path fails before stdin is consumed
ok 26 - T7: --json prints exactly one JSON document plus a newline
ok 27 - T8: table and JSON agree line for line on the same input
ok 28 - T9: --json is position independent
ok 29 - T10: --json without any input is a usage error with empty stdout
ok 30 - T13: JSON mode keeps stdout pure and stderr empty
ok 31 - pin: --json with every input skipped prints an empty object
ok 32 - pin: --json works together with stdin
ok 33 - pin: --json keeps input errors off stdout
ok 34 - counts lines per extension
# tests 34
# suites 0
# pass 34
# fail 0
# cancelled 0
# skipped 0
# todo 0
# duration_ms 845.790875
```

计划第 3 节的 21 条（T1–T21）全部写了，另外加了 12 条钉子：
扩展名 quirk、`(none)`、无参数 usage、二进制名不碰 fs、目录参数、T12b、T12c、坏路径先于 stdin、JSON 全跳过 `{}`、JSON+stdin、JSON 错误不污染 stdout、T18b。

### 2.2 同一套用例跑在 0.3.0 上（计划 §3.4 "先证明它真的会失败"）

把 `main` 的 `test/` 复制进只含 0.3.0 源码的临时目录，跑同一套：

```
$ node --test        # 在只含 9e91cb1 源码 + 现测试的目录
ok 1 - T1 ...                       ← 回归钉
ok 2 - T2 ...
ok 3 - T3 ...
ok 4 - T4 ...
ok 5 - T5 ...
ok 6 - T6 ...
ok 7 - pin: extension label ...
ok 8 - pin: paths without a dot ...
ok 9 - pin: no arguments ...
ok 10 - pin: a skipped name is skipped before the filesystem is consulted
ok 11 - T16 ...
not ok 12 - T14: a missing path exits 1 with something readable, stdout empty
not ok 13 - T15: every bad path is reported in one run
not ok 14 - T18: an unknown flag is a usage error (2), not a path
not ok 15 - T18b: an unknown flag alone is still a usage error
not ok 16 - T20: -- lets a flag-shaped file name through
not ok 17 - T21: an unreadable file exits 1 with a readable message
not ok 18 - pin: a directory argument is an input error, not a crash
not ok 19 - T11: "-" reads stdin and counts by the same convention
not ok 20 - T12: stdin aggregates with a file in the same (none) bucket
not ok 21 - T12b: stdin keeps its (none) label next to an extension file
not ok 22 - T12c: a stdin/file tie is ordered by argv position
not ok 23 - T17: empty stdin is an input error with an empty stdout
not ok 24 - T19: a repeated "-" is a usage error
not ok 25 - pin: a bad path fails before stdin is consumed
not ok 26 - T7: --json prints exactly one JSON document plus a newline
not ok 27 - T8: table and JSON agree line for line on the same input
not ok 28 - T9: --json is position independent
not ok 29 - T10: --json without any input is a usage error with empty stdout
not ok 30 - T13: JSON mode keeps stdout pure and stderr empty
not ok 31 - pin: --json with every input skipped prints an empty object
not ok 32 - pin: --json works together with stdin
not ok 33 - pin: --json keeps input errors off stdout
ok 34 - counts lines per extension
# tests 34
# pass 12
# fail 22
```

即：11 条回归钉 + 原有 count 单测在 0.3.0 上通过；22 条新证据在 0.3.0 上失败，失败原因都是"功能缺失"，不是测试写错。这与计划 §3.4 的最后一列一致，**除了 T14（见 P9）和 T21（见 P8）**。

### 2.3 成功路径字节对照（对 0.3.0）

用 `9e91cb1` 的 `src/cli.mjs` 与 `main` 的 `src/cli.mjs` 跑同一组 fixture，比较 **stdout 字节 + stderr 字节 + 退出码**：

```
SAME  exit=0 stdout=14B  /tmp/tally-probe/a.ts /tmp/tally-probe/b.md
SAME  exit=0 stdout=24B  /tmp/tally-probe/a.ts /tmp/tally-probe/b.md /tmp/tally-probe/long.jsx
SAME  exit=0 stdout=7B   /tmp/tally-probe/b.md
SAME  exit=0 stdout=8B   /tmp/tally-probe/empty.txt
SAME  exit=0 stdout=0B   /tmp/tally-probe/pic.png
SAME  exit=0 stdout=7B   /tmp/tally-probe/a.ts /tmp/tally-probe/pic.png
SAME  exit=0 stdout=7B   /tmp/tally-probe/a.ts /tmp/tally-probe/a.ts
SAME  exit=0 stdout=9B   /tmp/tally-probe/dir.v1/f
SAME  exit=0 stdout=10B  /tmp/tally-probe/noext
SAME  exit=0 stdout=0B   /tmp/tally-probe/ghost.png
SAME  exit=0 stdout=30B  /tmp/tally-probe/empty.txt /tmp/tally-probe/noext /tmp/tally-probe/a.ts
SAME  exit=0 stdout=16B  /tmp/tally-probe/empty.txt /tmp/tally-probe/pic.png /tmp/tally-probe/b.md

success-path fixtures (table mode): 12 byte-identical, 0 different
```

覆盖了冻结清单 9 条：行格式/两个空格/补齐、降序+并列首次出现、扩展名 quirk、`(none)`、尾换行多算 1、空文件 1 行、二进制静默跳过、全部跳过 0 字节 exit 0、重复路径重复计数。

**注意**：失败路径的 stderr **不可能**逐字节一致（P3），所以这张表对失败 fixture 只比较"退出码 + stdout 字节 + 错误类别"。

### 2.4 README 示例逐条执行（A11 的机械部分）

```
OK    table example
OK    json example
OK    stdin example
OK    empty stdin example (stderr)
OK    skip example
OK    usage example (stderr)
OK    usage example (empty stdout)
OK    input error example (stderr)

README examples: 8 ok, 0 failed
```

### 2.5 `npm pack` 冒烟 + 真实消费者安装（A12）

```
$ npm pack --pack-destination /tmp/tally-pack
package/package.json     package/FEATURE.md   package/README.md
package/src/args.mjs     package/src/cli.mjs  package/src/count.mjs
package/src/ext.mjs      package/src/render.mjs package/src/stdin.mjs
package/src/validate.mjs package/test/cli.test.mjs package/test/count.test.mjs
total files: 12
```

解包后在仓库外运行（**通过**）：

```
$ node package/src/cli.mjs --json a.ts b.md
{".ts":3,".md":2}
$ printf 'a\nb\n' | node package/src/cli.mjs -
(none)  3
$ node package/src/cli.mjs a.ts b.md logo.png
.ts  3
.md  2
```

但真实消费者安装后**以命令方式运行失败**（0.3.0 就有的缺陷，见 P10）：

```
$ npm install --offline /tmp/tally-pack/tally-0.4.0.tgz
$ ./node_modules/.bin/tally --json a.ts b.md
./node_modules/.bin/tally: line 1: //: is a directory
./node_modules/.bin/tally: line 81: syntax error near unexpected token `('
（shell 把 JS 当脚本解释；`src/cli.mjs` 没有 `#!/usr/bin/env node`）
```

### 2.6 环境与确定性

- 本机 `NODE_USE_ENV_PROXY=1` 确实让每个 node 进程往 stderr 打 `UNDICI-EHPA` 警告（计划 §0.2 的实测成立）。测试对所有子进程删除 `NODE_USE_ENV_PROXY` / `NODE_OPTIONS`，没有用 `--no-warnings`。
- T16 用 `spawn` 且**不关闭 stdin**，只等 `close` 事件；5s 定时器只做上界，断言的是退出码。手工复核：子进程 27ms 退出、exit 0、`.ts  3\n`。
- 测试 fixture 全部走 `mkdtempSync`；`T12c` 一开始复用了 `tie2.md` 并在测试里重写它，会造成顺序依赖，已改成独立文件名。

---

## 3. 计划与现实的差异（最重要）

### 3.1 计划成立、且真的好用的部分

| # | 计划原句 | 实测结论 |
|---|---|---|
| 1 | §0.2 那张"现状事实"表 | **逐条实测成立**：行格式、降序+稳定排序、`split('\n').length` 的 off-by-one、`dir.v1/f → .v1/f`、后缀跳过、全跳过 exit 0、无参数 usage+2、坏路径栈 exit 1 + stdout 0、重复路径、`--json`/`-` 被当路径、`NODE_USE_ENV_PROXY` 噪声 |
| 2 | 步骤 0 "用一组固定 fixture 记录 stdout 字节…它的长期形态是 PR1 里的断言字面量，不引入生成物" | 成立且够用。字面量直接成长在 `test/cli.test.mjs` 里，没有快照机制；12 条 fixture 精确挡住了重构期间的每一处字节变化 |
| 3 | 步骤 2 "先解析并校验…再读取计数，最后才渲染。渲染前不往 stdout 写一个字节" | 成立。两阶段让"失败时 stdout 0 字节"成为结构保证，而不是每处 `return` 前的自觉 |
| 4 | 步骤 3 "不给 `-` 时绝不读 stdin…这是本步唯一的高危点" | 判断准确，T16 真的能拦住（我在 PR1 就把它作为回归钉放进去了，因为 0.3.0 本来就满足它） |
| 5 | 步骤 1 "`test/count.test.mjs` **一行不改**且仍绿…的未被修改本身就是证据" | 成立（`git diff 9e91cb1 main -- test/count.test.mjs` 为空），但证据强度被 P16 削弱 |
| 6 | 步骤 4 "与表格共用同一份有序聚合结果，只是第二个渲染器" | 成立。T8 直接 `parseTable(table) === Object.entries(JSON.parse(json))`，两种输出不可能只在一侧改 |
| 7 | §3.3 "净化子进程环境…为什么不用 `--no-warnings`" | 成立。净化后 `stderr === ''` 稳定；`--no-warnings` 的理由（别把噪声变成不可见）也成立 |
| 8 | §3.2 "T1–T6 与 T16 是回归钉…T7–T15、T17–T21 是新证据" | 判断准确：11 条钉在 0.3.0 全绿，其余 22 条全红，且红的原因都是功能缺失 |
| 9 | §4.2 四个 PR 是依赖顺序 | 成立。没有发生"要回头改前面 PR"的情况；每层只碰自己范围内的文件 |
| 10 | §3.3 "临时目录走内建 `mkdtempSync`；不依赖 TTY；等完成信号不等时间" | 全部照做 |
| 11 | 步骤 5 "README 只写当前态" | 照做；我多加了一条扩展名 quirk 的说明（可观察输出，不写反而害脚本作者） |
| 12 | §3.6 第 5 条 "stdin 读取中途失败这类分支没有覆盖" | 诚实且成立：`readStdin` 的流错误确实没有 catch，会变成未捕获 rejection；本环境无法在进程级黑盒里制造 EIO（见第 5 节） |

### 3.2 计划错了 / 内部矛盾 / 漏了 / 多余的地方

> 每条都指到计划里的具体句子，并给出我的处理。

**P1〔过时〕§0.2 "仓库状态：**不是 git 仓库**（无 `.git`、无远端、无 CI…）"，以及 §4.1 整节"硬前置：这个目录不是 git 仓库"、D9。**
round3 的 repo **是** git 仓库：`git log` 有 `9e91cb1 chore: baseline 0.3.0`，`main` 存在，工作树干净。这条事实（以及建立在它上面的 D9 二选一）在 round3 不成立；如果再"做一次性初始化"会破坏基线。
处理：按"已初始化"执行，不重新 init；只补了计划 D9 推荐值里真正缺的那一半——在基线提交上打 `v0.3.0` 锚点标签（回退安排 §4.4 需要一个可命名的锚点）。没有远端，所以"核对主干/已发布版本"没有做。

**P2〔内部矛盾〕步骤 1 "参数解析（纯函数：argv → `{ json, paths, stdinRequested }` 或一个用法类错误）" 与冻结项 2、D3 冲突。**
`{ paths, stdinRequested }` 丢掉了 argv 顺序。没有 stdin 时，并列顺序靠 `Map` 插入序天然正确；但 D3 让 stdin 和无扩展名文件同属 `(none)` 桶，一旦并列（`tally - x.md` vs `tally x.md -`），就必须知道 `-` 在命令行里的位置，否则"并列按首次出现序"要么退化成"stdin 总是最先"，要么"总是最后"。
处理：parser 返回**有序** `inputs: [{type:'stdin'} | {type:'path', path}]`（外加 `json`）。加了 T12c 钉住两个方向的相反顺序。

**P3〔内部矛盾 + 不可满足〕步骤 1 "它需要变成 async" 与同一步的完成信号 "基线表逐字节重放一致（含 stderr 与退出码）"。**
0.3.0 的坏路径 stderr 是未捕获栈，栈里嵌**源码文件路径和行号**。任何重构（哪怕只改 `cli.mjs` 一行）都会改这些字节。实测 sync 版与 async 版的 ENOENT 栈只在栈帧行不同（`at main (…)` / `at file:///…`），错误类型、消息、退出码相同。计划自己在 §0.2 把 stderr 记成"**摘要**"，与步骤 1 的"含 stderr 逐字节"自相矛盾。
处理：把"逐字节"限定到 stdout（12/12 一致）、成功路径的 stderr（都为空）、退出码；失败路径的 stderr 只按"错误类别 + 退出码 + stdout 为空"对照，并在提交信息里写明。

**P4〔与冻结项冲突〕步骤 2 "先解析并校验（flag 合法性、路径存在性/可读性）" 与冻结项 6 / §4.3 "PR2 爆炸半径限于失败路径"。**
0.3.0 对 `missing.png` **从不 stat/read**（跳过规则先命中），所以是 `exit 0`。无差别预校验会把它变成 `exit 1`——这不是"失败路径变化"，而是**成功路径变失败**，直接违反 §4.3 对 PR2 的定性。
处理：`validatePaths` 对二进制后缀名 `continue`，一行也不碰 fs。加了回归钉 "a skipped name is skipped before the filesystem is consulted"（`tally ghost.png` → exit 0、stdout 空、stderr 空）。**这条钉子计划里没有**，而它正是这个计划最容易被实现者踩坏的地方。

**P5〔顺序错〕§0.4 表格 "usage 文本：追加 `[--json]`、`-`、`--`、退出码提示" 被写在步骤 2（PR2）。**
PR2 时 `--json` 还不存在（会被当未知 flag → exit 2），`-` 也还是普通路径。按计划在 PR2 写这段 usage，等于宣传两个非法的用法。
处理：usage 文本按 PR 递增——PR2 加 `--` + 退出码提示；PR3 加 `-`；PR4 加 `[--json]`。每个 PR 的 usage 都只描述当时真实存在的东西。

**P6〔范围错位〕步骤 1 "拆成四块无副作用单元……渲染（表格渲染器 / JSON 渲染器）"，但 JSON 是 PR4 的功能。**
在 PR1 写 `renderJson` 就是死代码，而且违反 §4.2 "每个 PR 恰好一种性质（PR1 = 重构）"和"只碰自己范围内的文件"。
处理：PR1 只落 args/ext/count/render(表格)/薄入口；`renderJson` 在 PR4 落，`countFile`/`STDIN_EXT` 在 PR3 落。"四个无副作用单元"到 PR4 才真正齐。

**P7〔内部矛盾〕§3.2 T12 "`tally - <file.ts>` | 同扩展名相加成一行" 与 D3 "stdin 标签复用 `(none)`"。**
两者不能同时成立：`-` 归 `(none)`、`.ts` 归 `.ts`，必然两行。A5 的"同一个扩展名桶相加"只在**无扩展名文件**时成立。
处理：按 D3 实现；T12 改成 `tally - noext`（同 `(none)` 相加成一行），另加 T12b 钉住 `-` + `.ts` 的**两行**事实。计划需要在 T12 措辞和 D3 之间改一个。

**P8〔证据不足〕§3.2 T21 "不可读文件（chmod 000）| exit 1 + 可读消息 | 在 v0.3.0 上：**失败**（栈）"。**
第一版 T21 只断言 `exit 1`、stdout 空、stderr 含路径、含 `permission denied`——它在 0.3.0 上**通过**，因为未捕获栈里本来就有 `EACCES: permission denied, open '<path>'`。实测基线输出：`ok 17 - T21`。
处理：加强为同时断言"无 `at …` 栈帧"和"含下一步提示 `check the file mode`"，它才在 0.3.0 上红。计划说的"失败"依赖一条没写出来的断言。

**P9〔表述歧义〕T14 在 §3.2 被列进"新证据"，§3.4 又说 "T14 … 这条不算失败，只验证'已成立的保证别退化'"，而 §3.2 给 T14 的三条断言（exit 1、stdout 0 字节、stderr 含路径）在 0.3.0 上全部通过（栈里当然有路径）。**
处理：把 T14 写成"可读消息"形式（无栈 + `check the path`），于是它在 0.3.0 上红。这与 §3.4 的"不算失败"表述不一致——但这个不一致是计划对 T14 的期望值本身没定死，不是实现问题。

**P10〔过度声称，最严重的一条〕§3.1 "发布形态（`bin` 映射、产物完整性）| `npm i` 的消费者 | `npm pack` → 解包 → 在临时目录跑一次 | 唯一能覆盖'新模块是否被发布出去'的证据"。**
这条证据**覆盖不到 `bin` 映射**：`npm pack` 只证明文件入包，解包后按 `node <path>` 跑也绕开了 `bin` 与 shebang。我按计划做完全通过，但多走了一步真实消费者安装，结果 `node_modules/.bin/tally` **跑不起来**：`src/cli.mjs` 没有 `#!/usr/bin/env node`，shell 把 JS 当脚本解释（`//: is a directory`、`syntax error`）。这是 0.3.0 就存在的缺陷，不是本次引入，但 A12 因此是"假验证"：A12 说"解包后直接运行 `--json` 与 `-` 两种用法都正常"——用 `node` 跑正常，**用 `tally` 命令跑不正常**。
处理：**未修**。修它是发布入口文件的变化（而且需要决定是否补 shebang、是否顺带 `files` 瘦身），计划 §0.5 明确把这些排除在外，D 列表也没有这一项。记为缺口（第 5 节），建议单开一行 PR：给 `src/cli.mjs` 加 shebang + 一条"安装 tarball 后用 `node_modules/.bin/tally` 跑一次"的门禁。

**P11〔计划不精确〕§0.4 "未知 flag | 当路径 → exit 1 | exit 2" 对二进制后缀的未知 flag 不成立。**
0.3.0 里 `tally --json.png` 会被后缀跳过规则先命中：`exit 0`、stdout 0 字节、stderr 空。新实现先解析 flag → `exit 2`。也就是说这条是"**成功路径变失败**"，不是计划表里说的"exit 1→2"。这是 D6 的必然结果（flag 解析必须在校验/跳过之前），但计划没披露这一格。
实测：
```
tally --json.png   0.3.0: exit=0 stdout=0B   0.4.0: exit=2 stdout=0B
tally --foo.png    0.3.0: exit=0 stdout=0B   0.4.0: exit=2 stdout=0B
```
处理：按 D6 执行（未知 flag 就该是 2），在报告里披露；README 不需要为这个边角加内容。

**P12〔计划没决定〕步骤 3 没有说 `--` 与 `-` 的组合。**
`tally -- -` 是读 stdin 还是打开名为 `-` 的文件？计划只说了"`--` 终止符 + 形似 flag 的文件名"。
处理：按 `cat` 语义——`-` 是操作数不是 flag，`--` 之后仍然是 stdin，因此无法指定一个真名叫 `-` 的文件。README 明写这条。

**P13〔计划没写〕stdin 与"先校验"的先后。**
步骤 2 说"先解析并校验…再读取计数"，但没说 stdin 属于"校验"还是"读取"。
处理：先 validate 路径，再读 stdin。好处是 `tally missing.ts -` 在坏路径时**不消费 stdin**（加了 pin "a bad path fails before stdin is consumed"）。

**P14〔计划漏了〕读阶段的 TOCTOU 错误。**
校验通过后文件仍可能在读取时消失/变成目录。计划 §3.6 只承认"stdin 读取中途失败"没覆盖，没提文件读错误。
处理：读阶段包 try/catch，用同一套 `describeFsError` 输出可读消息 + exit 1。加了一条 `pin: a directory argument is an input error, not a crash`（0.3.0 是 `EISDIR` 栈）。

**P15〔计划漏了 / "多余"的入口）步骤 1 "计数组合（`countFiles` 保持现有签名与语义，另外补一条只吃文本的入口供 stdin 用）"。**
"只吃文本的入口"（`countText`）确实够了行数，但**不够让 stdin 与文件共用扩展名/跳过规则**——stdin 也需要"一个输入 → 一个 `[标签, 行数]` entry"，而且必须能和文件在 argv 顺序里交错。计划没预见到需要 `countFile`（单输入入口）。
处理：PR3 补 `countFile`，`countFiles` 改为它的薄包装（签名/语义不变）。

**P16〔证据强度被削弱〕步骤 1/§3.1 "`test/count.test.mjs` 不改它本身就是'语义没动'的证据"。**
CLI 现在走的是 `countFile` + `aggregate`（为了 argv 顺序），**不再调用 `countFiles`**。所以"未修改的 count 单测"钉住的是模块级 API，而不是 CLI 的实际路径；`countFiles` 目前只有测试在调用。黑盒用例钉住 CLI 的实际路径，但从"唯一契约测试"这个论证看，计划把证据绑定在了一个 CLI 已不再经过的函数上。
处理：保留 `countFiles`（计划的签名承诺）+ 不改单测；在报告里说明这层缝。若要更强，应该给 `countFile` 也加一条模块级单测，或让 CLI 在无 stdin 时走 `countFiles`——后者会引入两条计数路径，我选择不引入。

**P17〔计划成立，补充一条环境事实〕§0.2 "环境噪声…任何 `stderr === ''` 断言必须先净化环境" 成立；补充：`npm test` 自身也会打这条警告（npm 自己的 node 进程），所以原始 `npm test` 输出里会有，不影响被测量的子进程。**

### 3.3 计划没写但我必须做的决定（汇总）

| 决定 | 为什么 | 证据 |
|---|---|---|
| parser 返回有序 `inputs` 而非 `{paths, stdinRequested}` | 并列顺序（P2） | T12c |
| usage 文本按 PR 递增 | 不能宣传不存在的 flag（P5） | 各 PR 的 usage |
| 校验跳过二进制后缀名 | 冻结项 6（P4） | pin "skipped name…" |
| `countFile` 单输入入口 | stdin 要按 argv 顺序交错（P15） | T12c、T12 |
| `--` 后 `-` 仍是 stdin | 与 `cat` 一致（P12） | README |
| 先 validate 再读 stdin | 坏路径不消费 stdin（P13） | pin "bad path fails before stdin" |
| 读阶段 catch TOCTOU | 不让边缘 IO 错误变栈（P14） | pin "directory argument" |
| T21 加强、T12 改写 | 让新证据真的在 0.3.0 上红（P8/P7） | 2.2 的基线输出 |
| 失败用 `process.exitCode` 而非 `process.exit()` | 管道下不截断 stdout/stderr；可观察退出码不变 | 全部用例 |
| `renderTable` 先拼字符串再单次 `write` | 空行集合时 stdout 恰好 0 字节；字节与逐行 `console.log` 相同 | 2.3 |

---

## 4. D1–D10 执行情况

| # | 推荐值 | 我的执行 | 证据/备注 |
|---|---|---|---|
| D1 | 三种"空"分开：无参数→2；`-` 且 0 字节→1+可读消息；零字节文件仍 1 行 | **按推荐值执行** | T17（空 stdin exit 1）、T3（空文件 1 行）、pin 无参数 exit 2；README 明写这条不一致是有意的 |
| D2 | 单对象 `{".ts":3,".md":2}`，键序与表格一致**并作为契约** | **按推荐值执行** | `renderJson` 吃同一 `rows`；T7 断言精确字节 `{".ts":3,".md":2}\n`，T8 逐行比对，README 写明"键序是契约" |
| D3 | stdin 标签复用 `(none)` | **按推荐值执行** | `STDIN_EXT = '(none)'`；T11；同时发现与 T12 措辞冲突（P7），T12 按 D3 改写，T12b 钉住两桶事实 |
| D4 | 不修扩展名派生缺陷，单开 PR | **按推荐值执行（未修）** | 回归钉 pin `.v1/f`；README 把这条规则写成现状（含缺陷），没有偷偷修 |
| D5 | 采用 0/1/2 分级；一次报告全部坏路径 | **按推荐值执行** | T14/T15/T18/T19/T21；T15 断言两条坏路径都在 stderr |
| D6 | 支持 `--`；未知 flag → 2 | **按推荐值执行** | T20（`--` 放行 `--weird.ts`）、T18/T18b；补充：`--` 后 `-` 仍读 stdin（计划未定，按 `cat` 语义） |
| D7 | TTY 下 `-` 阻塞等输入（与 `wc -` 一致） | **按推荐值执行（实现天然阻塞）** | 用 `script` 伪终端观察：不给输入时进程一直等到 pty 关闭（实测等待 1.5s ≈ 管道保持时长），随后按"空 stdin"报 exit 1、无栈；真正的交互 Ctrl-D 未验证（`script` 不是可信 pty，见第 5 节） |
| D8 | 0.4.0；发布与打标签单独授权 | **按推荐值执行** | `package.json` 0.4.0；**没有** `npm publish`、**没有** `v0.4.0` 标签（授权不在本任务内）；只给基线打了 `v0.3.0` 锚点 |
| D9 | 先确认真实仓库/主干；若这是源仓库，先做一次性初始化并落 0.3.0 基线锚点 | **改为"已满足 + 补锚点"**：round3 已初始化且基线提交存在，不重新 init；补打 annotated `v0.3.0`；无远端，无法核对"已发布版本" | 见 P1；`git log --decorate` 显示 `tag: v0.3.0` |
| D10 | stdin 按文本，不做 NUL 嗅探 | **按推荐值执行** | `readStdin` 一律 utf8 拼接；README 写明理由（跳过规则基于文件名，stdin 无文件名） |

---

## 5. 无法验证的验收条件（以及为什么）

| 条件 | 状态 | 为什么无法验证 |
|---|---|---|
| A11 的后半句（"一个没读过代码的人能凭 README 写出 `tally --json … \| jq` 并正确解释 1/2 退出码"） | **部分**：8/8 示例逐字节核过 | 我能证明"示例与实现一致"，不能证明"读者能凭它写出脚本"——那是人工可读性判断，计划 §3.5 也把它交给人 |
| A12 的意图（"`npm i` 的消费者能用"） | **未达成**：pack → 解包 → `node <path>` 全过；真实安装后 `node_modules/.bin/tally` 失败 | `src/cli.mjs` 缺 shebang，0.3.0 已有（P10）。不在计划范围内，未修；只做到本地 tarball 安装 |
| Node 20.x（`engines` 下界） | **未验证** | 本机只有 v22.23.1，无 nvm/其他运行时；计划 §3.5 已列为缺口 |
| Windows / 非 POSIX | **未验证** | 无环境；代码走 `process.stdin` 流、测试用 `node:path`，只是降低风险，不是证据 |
| 交互式 TTY 下 `-` 的完整人工手测（计划 §3.5 / §3.6 第 3 条） | **只观察到"阻塞"** | `script` 的 pty 不忠实：第二次实验里管道输入被回显但没送到子进程，子进程看到的是 EOF → "空 stdin"错误。不能据此断言交互行为 |
| stdin 读取中途失败（EIO）分支（计划 §3.6 第 5 条） | **未覆盖，且未 catch** | 进程级黑盒里造不出 EIO（管道关闭是 EOF 不是错误，pty 不可信）；`readStdin` 抛错会变成未捕获 rejection。计划自己已承认这个缺口 |
| 覆盖率 / CI / 发布后真实安装 | **没有** | 仓库无 CI、无覆盖率门禁；未 `npm publish`。所有证据标注"本地 v22.23.1 / macOS" |
| "当前已发布版本是否就是 0.3.0"（D9 的核对） | **无法** | 无远端、无 `npm view` 授权，不对公网包名做假设 |
| 性能 | 不需要 | 本改动没有性能声明（计划 §3.5 的判断成立） |

---

## 6. 残余风险与建议

1. **发布入口是坏的（最高优先）**：`bin` 指向的文件没有 shebang。加一行 `#!/usr/bin/env node` 即修，但需要单独 PR + 一条"安装 tarball 后用 `.bin/tally` 跑一次"的证据（`npm pack` 冒烟查不出来）。
2. **`countFiles` 与 CLI 实际路径已分离**（P16）：要么给 `countFile` 加模块级单测，要么在无 stdin 时走 `countFiles`。当前黑盒测试覆盖了真实路径，但 `countFiles` 只剩测试在调。
3. **JSON 键序依赖实现细节**：所有标签以 `.` 开头或是 `(none)`，所以 `JSON.stringify` 不会把键当整数重排。如果将来允许整数样标签，键序会静默变；建议在 `renderJson` 旁边留一条断言（现在只有注释）。
4. **`--json` 与未知 flag 的边角**：`--json.png` 这类"二进制后缀的 flag 形状名"从 exit 0 变 exit 2（P11）。如果真有用户这么用，需要一次显式兼容决定。
5. **`-` 在 `--` 之后仍是 stdin**：这是与 `cat` 一致的约定，但计划没决定过；如果评审者想要"`--` 之后一切都是文件名"，改这一处即可（T20 不受影响）。
6. **stderr 的字节契约**：失败文案已定型（`tally: <path>: <reason> (<hint>)`），一旦发布就是兼容面；PR2 的 sign-off 应该按 §4.3 的要求真的给。
