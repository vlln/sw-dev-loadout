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
