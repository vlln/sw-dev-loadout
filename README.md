<p align="center">
  <img src="docs/assets/hero.png" alt="sw-dev-loadout：一条流水线——漏斗、评审、天平、门禁、封箱五个工位" width="100%">
</p>

<h1 align="center">sw-dev-loadout</h1>

<p align="center"><b>八份 skill，覆盖软件开发里最常做的几件事。</b></p>

<p align="center">
  <a href="#八份各管什么">八份各管什么</a> ·
  <a href="#怎么配一组">怎么配一组</a> ·
  <a href="#安装">安装</a>
</p>

<p align="center">
  <a href="https://agentskills.io"><img alt="Agent Skills" src="https://img.shields.io/badge/Agent_Skills-规范-4D6BFE?style=flat-square"></a>
</p>

<br>

## 写代码时容易漏的那些事

写代码的时候，有些事每次都得做，但很容易忘：

- 需求没读完就动手，做到一半发现其中一条要求根本满足不了
- 失败路径没想清楚，出错之后不知道谁来处理、要不要重试
- 交之前不知道该跑哪些测试，凭感觉挑几个
- 提交信息和 PR 描述随手写，过一个月自己也看不懂
- 项目约定写在文档里，没人检查，慢慢就没人遵守了
- README 很久没更新，新人看不出来这东西怎么用

这八份 skill 各管一件。装好以后，遇到对应的事情，agent 会照着里面的规则做。

## 八份各管什么

| skill | face | domain | 介绍 |
|---|---|---|---|
| [delivery-loop](skills/delivery-loop/SKILL.md) | `process` | 软件开发 | 一件活从开始到交付怎么推进，什么时候算做完 |
| [review-gate](skills/review-gate/SKILL.md) | `process` | 软件开发 | 看 PR 时判断哪些要拦、谁批准、要求本身是否做得到 |
| [engineering-standards](skills/engineering-standards/SKILL.md) | `quality` | 软件开发 | 给项目定规矩，并且让规矩真的能被检查 |
| [contract-prose](skills/contract-prose/SKILL.md) | `quality` | 文档 | 写注释和文档：只写事实，不写过程 |
| [failure-handling](skills/failure-handling/SKILL.md) | `quality` | 软件开发 | 写会失败的操作：失败之后怎么办、谁来处理 |
| [change-evidence](skills/change-evidence/SKILL.md) | `quality` | 软件开发 | 这次改动要跑哪些测试，凭什么算验证过 |
| [readme-writing](skills/readme-writing/SKILL.md) | `quality` | 文档 | 写 README，或者检查一份 README 哪里不对 |
| [change-landing](skills/change-landing/SKILL.md) | `execution` | 软件开发 | 提交、推送、拆 PR、把一组改动合进去 |

八份互不引用，单独装一份也能用。

## 怎么配一组

不用全装，带上这次用得着的几份就行：

| 你在做什么 | 带哪几份 |
|---|---|
| 修个 bug、改个小东西 | `change-evidence` `change-landing` |
| 接一个新功能 | `delivery-loop` `review-gate` `change-evidence` |
| 发一版 | `delivery-loop` `change-landing` `readme-writing` |
| 给项目定规矩 | `engineering-standards` `review-gate` |
| 写模块文档或改文案 | `contract-prose` |
| 动支付、迁移、权限这类出错代价大的地方 | `failure-handling` `review-gate` `change-evidence` |

装了 [dsh-loadout](https://github.com/vlln/skill-loadout) 插件的话，上面每一行可以存成一个编组，在会话里切换。

## 安装

用 skit：

```sh
skit install vlln/sw-dev-loadout/skills/delivery-loop vlln/sw-dev-loadout/skills/review-gate      # 流程
skit install vlln/sw-dev-loadout/skills/engineering-standards vlln/sw-dev-loadout/skills/contract-prose vlln/sw-dev-loadout/skills/failure-handling vlln/sw-dev-loadout/skills/change-evidence vlln/sw-dev-loadout/skills/readme-writing   # 质量
skit install vlln/sw-dev-loadout/skills/change-landing                           # 执行
skit check
```

也可以用 npx skills（需要 Node.js 22.20 以上）：

```sh
npx skills add vlln/sw-dev-loadout --skill readme-writing    # 换成别的名字就装别的
```

什么都没装的话，把下面这段发给 agent，让它自己装：

```text
Run `npx skills use "vlln/sw-dev-loadout@readme-writing"` and follow the generated skill instructions now. Read its complete output, redirecting it to a temporary file first if necessary. Resolve relative paths from the supporting-files directory it provides.
```

本地改的时候，克隆下来用 `./skills/<名>` 装。

装完 `skit list` 里能看到这几份，`skit check` 没有指向它们的报错，就算装好了。

## 为什么是这八份

按"这份 skill 回答什么问题"分，只有三种：流程（什么时候做什么、谁决定）、质量（什么样算合格）、执行（具体怎么做）。八份里流程 2 份、质量 5 份、执行 1 份。

拆分的标准是两条规则会不会抢同一段指令，不是按开发阶段切。一个任务带 2–4 份是正常的，这也是 loadout 这个名字的意思：一次任务实际加载的那一组。

具体的分工和取舍写在 [docs/plan.md](docs/plan.md)。

## 证据

这八份不是凭空写的，是在五轮对照实验里改出来的。每轮两个 agent 做同一件事，一个能拿到这组 skill，一个拿不到。

结果大概是：需要判断力的活（比如审 diff）两边差不多；差别在流程和验证上，比如会不会先确认要求能不能满足、会不会在改动之前拿旧版本跑一遍测试。代价是有时候流程偏重。各轮的数据、被推翻的预测，还有这些结论能信多少，都在 [docs/experiment.md](docs/experiment.md)。

## 文档

| 路径 | 是什么 |
|---|---|
| [docs/plan.md](docs/plan.md) | 八份怎么分、哪些材料合并、哪些丢掉、为什么 |
| [docs/experiment.md](docs/experiment.md) | 五轮对照的度量、结论与限制 |
| [docs/readme/](docs/readme/) | readme-writing 自己的证据：同一份 README 四次产出 354→372→148→140 行 |
| [fixtures/](fixtures/) | 实验用到的项目副本和已知问题 |
| [docs/runs/](docs/runs/) | 保存下来的原始产出，含第三轮结果的 git bundle |

## 许可

MIT
