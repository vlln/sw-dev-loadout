# sw-dev-loadout

一组 Agent Skill（7 份），覆盖做软件开发要回答的三类问题：什么时候做什么、凭什么算合格、跑什么证据怎么落地。

面向任何软件项目，不绑定某个仓库——仓库专属的脚本名、目录约定与界面常量都留在源材料里，没有带进来。

## 安装

这个仓库还没有推到远端，先克隆或下载，然后按需装：

```sh
skit install ./skills/delivery-loop ./skills/review-gate      # 流程
skit install ./skills/engineering-standards ./skills/contract-prose ./skills/failure-handling ./skills/change-evidence   # 质量
skit install ./skills/change-landing                           # 执行
skit check
```

七份互不引用，装一份也能用。一个任务加载哪几份由任务决定，不需要装全。

## 七份各管什么

| skill | 它主要回答的问题 | 什么时候激活 |
|---|---|---|
| [delivery-loop](skills/delivery-loop/SKILL.md) | 一轮交付怎么推进、谁能决定、怎么交接、哪个完成声明算哪一层 | 跨会话推进、接手在途工作、中途改范围、判断"这算做完了吗" |
| [review-gate](skills/review-gate/SKILL.md) | 谁来看、看几轮、发现怎么核验、多大风险归谁批 | 评审 PR、判断一处发现算不算阻塞、多轮不收敛 |
| [engineering-standards](skills/engineering-standards/SKILL.md) | 一条约定怎么才算真的存在 | 建立或审计工程规范、规范落不了地、新增门禁、文档与代码漂移 |
| [contract-prose](skills/contract-prose/SKILL.md) | 注释、文档、提示词、界面文案该写什么才算合格 | 写或审文档与文案、清理变更叙述与思考过程泄漏 |
| [failure-handling](skills/failure-handling/SKILL.md) | 失败路径设计得对不对 | 加 catch、决定失败后能否继续、设计重试、审错误文案 |
| [change-evidence](skills/change-evidence/SKILL.md) | 这个改动凭什么算被验证 | 选测试范围、推送前取舍检查、查 CI 不稳定、性能结论 |
| [change-landing](skills/change-landing/SKILL.md) | 怎么把改动安全落到主干 | 提交、推送、改写历史、落地一组 PR、写 PR 描述 |

一条核心断言贯穿全组：**约定除非有程序会拒绝违规，否则等于不存在。**

## 它有多少证据

三轮同题对照，两个互不知情的 agent，同一输入，只差可用材料：

| 轮 | 任务 | 结果 |
|---|---|---|
| 1 | 审一份埋了 11 个问题的 diff | **打平**，两组都命中 6 个阻断项、零误报 |
| 2 | 为一个小 CLI 写实施计划 | 流程与落地更完整（4 个 PR 带依赖序与逐 PR 回退），代价是篇幅翻倍（336 vs 167 行） |
| 3 | 把第二轮的计划交给另一个 agent 执行 | 计划骨架成立、34 个用例全绿，但执行时撞上 16 处问题含 4 处自相矛盾 |

**这组 skill 的作用是让纪律成为默认，不是让判断变聪明。** 第一轮说明质量面的判断力在原始材料里已经足够；真正的增益在流程与落地的完整性上，风险是过度流程。

三轮各逼出一条规则：按规模裁剪流程、切片看爆炸半径、新证据必须在改动前的版本上真跑一遍。

夹具与每一轮的原始产出都在仓库里：[fixtures/](fixtures/)、[docs/runs/](docs/runs/)。结论与它有多硬见 [docs/experiment.md](docs/experiment.md)——每轮只跑过一次，补的规则都只被一个夹具逼出来过。

## 文档

| 路径 | 是什么 |
|---|---|
| [docs/plan.md](docs/plan.md) | 地图与取舍：七份怎么分、哪些材料合并、哪些丢弃、为什么 |
| [docs/experiment.md](docs/experiment.md) | 三轮对照的度量、结论与限制 |
| [fixtures/](fixtures/) | 实验输入：自检方法与已知缺陷 |
| [docs/runs/](docs/runs/) | 保存下来的原始产出，含第三轮结果的 git bundle |
| [docs/v1/](docs/v1/) | 冻结的第一版，用于对照后续改动 |

## 许可

MIT
