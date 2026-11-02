<p align="center">
  <img src="docs/assets/hero.png" alt="sw-dev-loadout：一条流水线——漏斗、评审、天平、门禁、封箱五个工位" width="100%">
</p>

<h1 align="center">sw-dev-loadout</h1>

<p align="center"><b>软件开发里那些躲不开的活，一样一份 skill。</b></p>

<p align="center">
  <a href="#八份各管什么">八份各管什么</a> ·
  <a href="#怎么配一组">怎么配一组</a> ·
  <a href="#安装">安装</a>
</p>

<p align="center">
  <a href="https://agentskills.io"><img alt="Agent Skills" src="https://img.shields.io/badge/Agent_Skills-规范-4D6BFE?style=flat-square"></a>
</p>

<br>

做一个功能，来回就那些事：先把需求吃透、别做完才发现不可行；写的时候想清楚失败路径；交之前知道该跑哪些测试；提交、推送、开 PR；别人看的时候能一眼判出哪里该拦；项目里的约定不能只写在文档里；README 要能让陌生人看懂。

每一件都有一套该守的纪律，但每次都靠记性。这八份 skill 就是把它们一样一份固定下来。

## 八份各管什么

| skill | face | domain | 介绍 |
|---|---|---|---|
| [delivery-loop](skills/delivery-loop/SKILL.md) | `process` | 软件开发 | 一轮活怎么推进，做到哪儿算完成 |
| [review-gate](skills/review-gate/SKILL.md) | `process` | 软件开发 | 看别人的 PR：哪些该拦、谁批准、要求本身是否可满足 |
| [engineering-standards](skills/engineering-standards/SKILL.md) | `quality` | 软件开发 | 给项目立规矩：怎么让约定真的会拒绝违规 |
| [contract-prose](skills/contract-prose/SKILL.md) | `quality` | 文档 | 写注释与文案：只写契约，不写流水账 |
| [failure-handling](skills/failure-handling/SKILL.md) | `quality` | 软件开发 | 写会失败的操作：失败类型、谁处理、能不能继续 |
| [change-evidence](skills/change-evidence/SKILL.md) | `quality` | 软件开发 | 这次改动该跑哪些测试、凭什么算验证过 |
| [readme-writing](skills/readme-writing/SKILL.md) | `quality` | 文档 | 写或重写 README：读者优先、可交付自查 |
| [change-landing](skills/change-landing/SKILL.md) | `execution` | 软件开发 | 提交、推送、切片、落地一组 PR |

八份互不引用，装一份也能用。一条断言贯穿全组：**约定除非有程序会拒绝违规，否则等于不存在。**

## 怎么配一组

不用全装。按你在做什么带几份——这组就是一次任务的 loadout：

| 你在做什么 | 带哪几份 |
|---|---|
| 修个 bug、改个小东西 | `change-evidence` `change-landing` |
| 接一个新功能 | `delivery-loop` `review-gate` `change-evidence` |
| 发一版 | `delivery-loop` `change-landing` `readme-writing` |
| 给项目立规矩 | `engineering-standards` `review-gate` |
| 写模块文档或改文案 | `contract-prose` |
| 动支付、迁移、权限这些出错代价大的地方 | `failure-handling` `review-gate` `change-evidence` |

装了 [dsh-loadout](https://github.com/vlln/skill-loadout) 插件的话，上面每一行可以直接存成一个编组，会话里一键切换。

## 安装

推荐 skit：

```sh
skit install vlln/sw-dev-loadout/skills/delivery-loop vlln/sw-dev-loadout/skills/review-gate      # 流程
skit install vlln/sw-dev-loadout/skills/engineering-standards vlln/sw-dev-loadout/skills/contract-prose vlln/sw-dev-loadout/skills/failure-handling vlln/sw-dev-loadout/skills/change-evidence vlln/sw-dev-loadout/skills/readme-writing   # 质量
skit install vlln/sw-dev-loadout/skills/change-landing                           # 执行
skit check
```

生态通用 CLI（需 Node.js ≥ 22.20）：

```sh
npx skills add vlln/sw-dev-loadout --skill readme-writing    # 例；换 --skill 装别的
```

什么都没装时，把这段直接发给 agent，让它自己装自己用：

```text
Run `npx skills use "vlln/sw-dev-loadout@readme-writing"` and follow the generated skill instructions now. Read its complete output, redirecting it to a temporary file first if necessary. Resolve relative paths from the supporting-files directory it provides.
```

本地开发时克隆后用 `./skills/<名>` 直装。

跑通的样子：`skit list` 里出现刚装的几份，且 `skit check` 没有指向它们的报错。

## 为什么是这八份

按"这份 skill 主要回答哪个问题"分，只有三种：**流程**（什么时候做什么、谁决定）、**质量**（凭什么算合格）、**执行**（具体怎么动手）。八份里流程 2、质量 5、执行 1，拆分依据是"两条规则会不会抢同一段指令"，不是按开发阶段切。

一个任务带 2–4 份是常态——这也是 loadout 这个名字的来源：一次任务实际加载的那一组。完整分工与取舍见 [docs/plan.md](docs/plan.md)。

## 它有多少证据

这八份不是推演出来的，是五轮同题对照实验里逼出来的：每个实验两组 agent 拿同样的输入，只差给不给这组 skill。

一句话结论：**它让纪律成为默认，不是让判断变聪明**——质量面的判断力本来就不差（第一轮审 diff 打平），真正的增益在流程完整性与验证纪律上，风险是过度流程。各轮的数据、被推翻的预测与"这些结论有多硬"都在 [docs/experiment.md](docs/experiment.md)。

## 文档

| 路径 | 是什么 |
|---|---|
| [docs/plan.md](docs/plan.md) | 八份怎么分、哪些材料合并、哪些丢弃、为什么 |
| [docs/experiment.md](docs/experiment.md) | 五轮对照的度量、结论与限制 |
| [docs/readme/](docs/readme/) | readme-writing 自己的证据线：同一 README 四次产出 354→372→148→140 行 |
| [fixtures/](fixtures/) | 实验输入与已知缺陷 |
| [docs/runs/](docs/runs/) | 保存下来的原始产出，含第三轮结果的 git bundle |

## 许可

MIT
