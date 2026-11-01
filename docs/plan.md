# 通用软件开发 loadout：地图与取舍

这套东西面向**任何软件项目**，不面向某一个仓库。它由 7 份 skill 组成，合起来覆盖做软件开发的三件事：**流程**（什么时候做什么、谁决定、怎么交接）、**质量**（凭什么算合格）、**执行**（用什么做、跑什么、产物在哪）。

定位与取舍的判断方法来自 loadout 的五份方法 skill（定位、制作、写作、改进回路、安装校验），材料来自三批只读源：deepseek-harness 的 21 份仓库级 skill、devloop（1 份 SKILL + 12 份 references）、deep-standard（1 份 SKILL + 7 份 references）。

## 地图

| skill | face | 它主要回答的问题 | 什么时候激活 | references |
|---|---|---|---|---|
| `delivery-loop` | process | 一轮交付怎么推进、改向怎么处理、谁能决定、怎么交接、每个完成声明属于哪一层 | 跨会话推进、接手在途工作、用户中途改范围、判断「这算做完了吗」 | 变更传播、记录与交接、工程准备 |
| `review-gate` | process | 谁来看、看几轮、发现怎么核验、多大风险归谁批、怎么向人汇报 | 评审 PR、判断一处发现算不算阻塞、评审多轮不收敛、决定谁批准 | 发现核验与账本、风险分级与审批 |
| `engineering-standards` | quality | 一条约定怎么才算真的存在（可执行、会拒绝、能自证、只有一个家） | 建立或审计工程规范、规范落不了地、新增门禁、文档与代码漂移 | 常驻规则分层、门禁即代码、决策记录、装进已有项目 |
| `contract-prose` | quality | 注释、文档、提示词、界面文案该写什么才算合格 | 写或审注释与文档、清理变更叙述、判断某处该不该有文档、清理思考过程泄漏 | 按位置覆盖、泄漏与保留 |
| `failure-handling` | quality | 失败路径设计得对不对：类型、恢复、状态保证、可见报告 | 加 catch、决定失败后能否继续、设计重试、审错误文案 | 状态保证与补偿 |
| `change-evidence` | quality | 这个改动凭什么算被验证：该跑哪条证据、证据怎样才可能失败 | 选测试范围、推送前取舍检查、查 CI 不稳定、性能结论 | 触达面到证据、确定性测试、性能改动的证据 |
| `change-landing` | execution | 怎么把改动安全落到主干：切片、依赖链、强推、PR 描述、多写者 | 提交、推送、改写历史、落地一组 PR、写 PR 描述、并行写同一仓库 | PR 描述 |

每份只标一个主面，写进 `metadata.face`。七份互不引用，也不提彼此的名字；共用规则各自写最小充分版，需要全文时只能从本文自己的 references 里找。

## 三面覆盖

| 面 | 由谁承担 | 覆盖到什么程度 |
|---|---|---|
| process | `delivery-loop`、`review-gate` | 从立项到交付的推进条件、输入分流、授权复用、变更传播、记录与交接、验收与交付的区别；评审的独立性、轮次收敛、风险路由、人的注意力预算 |
| quality | `engineering-standards`、`contract-prose`、`failure-handling`、`change-evidence` | 约定如何变成会拒绝违规的程序；文字如何只承载契约；失败如何被类型化、被限制、被写清保证、被做成可见报告；一次改动凭什么算被验证（最窄证据、分层归属、确定性、性能证据） |
| execution | `change-landing` | 分支与切片、官方栈、带期望值的强推、推送前后核对、PR 描述 |

**最薄的一面是 execution。** `change-evidence` 起初标为 execution，但它的判定句几乎都是「凭什么算被验证」，2026-11-01 改标 quality；它和 `engineering-standards` 原本都声称拥有「新增门禁」的触发与「门禁要证明自己会拒绝」的定义，同时改为只由后者持有。剩下的执行面停在「本地动作 + 远端动作」这一层，以下主题在材料里几乎没有通用形态，因此没有覆盖：环境与依赖搭建（容器、工具链、服务依赖）、可观测性接入（日志、指标、告警）、发布工程（制品、通道、回滚脚本、灰度）。这三类在源材料里都以「本项目已有 CI 与发布流水线」为前提，抽出来的通用部分只剩「找到入口、跑一次、报告」这一句，写不成 skill。

process 只有两份，但覆盖面最宽；quality 三份最细。这与材料分布一致：仓库级 skill 大多在写「什么叫合格」，流程材料集中在 devloop 一份，执行材料集中在两处命令清单里——后者正是最不可移植的部分。

## 为什么是这七份

拆合按「它主要回答哪个问题」判，不按活动分类。

**合并的四处**（原本是几份，问句相同）：

- `dsh-error-handling` 几乎整体成为 `failure-handling`。它是三批材料里最完整、最少仓库依赖的一份，只在「报告文案」处与 `dsh-client-ui-ux` 的文案规则合并。
- `dsh-prose-standard` + `dsh-trim-cot-leakage` + `dsh-doc` 的写作部分 + devloop 的文档协作 → `contract-prose`。四者都在判「这段文字该不该存在、写成什么样」，拆成两份会让「契约 vs 转录」这条主判据出现两个权威。
- devloop 的 `work-and-delivery` + `decisions-and-change` + `records` + 六个 phase → `delivery-loop`。它们的问句是同一个：「这一步凭什么算过去了」。devloop 原文件用 13 个文件承载一份流程，对通用用途太重；本 loadout 只留判断与关口，把操作步骤压进 3 份 references。
- deep-standard 的六条法则 + `dsh-pre-push-checks` 的门禁编排部分 → `engineering-standards`。它们共同的判据是「有没有程序会拒绝」。

**拆开的一处**：deep-standard 的法则 4（最小充分证据）与法则 5（散文契约）原本与其余四法同处一份。按问句分，法则 4 是执行面（跑什么），法则 5 是质量面（写什么），而且法则 4 需要与 `dsh-pre-push-checks`、`dsh-ci-test-reliability` 合并才完整。硬留在一份里会得到一个四不像：既是规范体系又是测试选择器。

**评审为什么不拆成「判据」和「流程」两份**：评审的判据（什么算阻塞）如果脱离「谁核验、几轮收敛、谁批」就变成没有落点的清单；反过来，流程如果没有判据就只剩会议安排。合并后主面标 process，判据放正文与 references，读者能分清哪句是「该拦」哪句是「该谁决定」。

**明确没有做的**：没有做 UI/视觉 skill。`dsh-client-ui-ux`、`dsh-responsive-ui`、`dsh-official-plugin-artwork` 的规则几乎全部绑定一套具体设计系统（token 名、组件目录、平台窗口常量、13px 行高、字重上限）。抽掉这些之后只剩「复用已有组件」「暗色模式都要看」这类无法判定的空话——那正是 `skill-writing` 说的「上下文税」。

## 材料怎么处理

### 合并进某份 skill 的

| 材料 | 落点 | 合并的是什么 |
|---|---|---|
| devloop SKILL + `work-and-delivery` + `decisions-and-change` + `records` + 6 个 phase + `migration` | `delivery-loop` | 五条不变式、输入分流、三种结论与授权复用、六阶段推进条件、变更传播、记录与状态词、停止与接管、证据适用性、验收与交付、接入的完成判据 |
| devloop `review-and-risk` | `review-gate` | 五步流水线、发现三轴与账本、轮次收敛、风险分级与审批路由、三种输出、合入后回路 |
| devloop `documentation` | `contract-prose`、`change-landing` | 语句效力的五类（要求/决定/事实/假设/提议）、当前正文完整历史另有锚点、共享文档的整合责任与「Git 无冲突≠语义一致」 |
| deep-standard 六条法则 + `adoption` | `engineering-standards` | 常驻分层与预算、门禁七条契约与自证、生成物防漂移、决策记录生命周期、采纳阶梯与判档信号、审计四问 |
| deep-standard `evidence` | `change-evidence` | 最窄证据、分层唯一归属、hook/本地/CI 分工、覆盖率与豁免契约、快照与固定数据、确定性纪律 |
| `dsh-ci-test-reliability` | `change-evidence`（确定性） | 执行拓扑、原子分配、进程全局状态、平台语义、超时预算、按状态同步、拆卸到静止、拒绝掩蔽 |
| `dsh-pre-push-checks` | `change-evidence` + `change-landing` | 证据与改动面的匹配、聚焦运行、推送程序、强推保护、推送后核对 |
| `dsh-speed-up-perf` | `change-evidence`（性能） | 先证明回归、测量卡、发布形态、两侧可比、负控、放弃条件、语料隐私 |
| `dsh-code-review` | `review-gate` | 先读要求再判 diff、阻塞项与人工检查面、发现的报告格式、接收评审时的态度 |
| `dsh-pr-description` | `change-landing`（PR 描述） | 四栏目、按变更类型的测试设计、测试条目与证据、披露项 |
| `dsh-merging-stacked-prs` | `change-landing` | 官方栈机制、不可用即硬停、跨 fork 硬停、重写后重审 |
| `dsh-find-simplifications` | `review-gate` | 「足以改变决定吗」的判断、范围外与被替代的处置、可达性先于删除 |
| `dsh-prose-standard` + `dsh-trim-cot-leakage` | `contract-prose` | 保留/删除清单、完整命题、按位置覆盖、泄漏的一个测试与八类形态、保留清单、过度修剪的四种翻车 |
| `dsh-error-handling` | `failure-handling` | 类型化、捕获与恢复、状态保证、可见报告、验证方式 |
| `dsh-doc` | `contract-prose`、`engineering-standards` | 摘要说什么而非身份、教程与参考分离、字数预算机制、「跑过才算事实」 |
| `agent-experience` | `contract-prose` | 措辞即行为；提示词与可见文案的覆盖要求 |

### 丢弃的（连同理由）

- `dsh-session-projection`、`dsh-official-plugin-artwork`、`dsh-responsive-ui`、`cordis-styling-nitpicking`：内容全部绑定一个产品、一套图形规范或一个框架的生命周期语义。`cordis-styling-nitpicking` 的三条 P0（单例状态、不可逆副作用、拆卸顺序）是通用的，但它已经以「拆卸到静止」「副作用的所有权」的形式存在于 `change-evidence` 与 `failure-handling` 里，单独成篇只会重复。
- `dsh-translate-docs`、`dsh-archive-agent-notes`、`dsh-create-upgrade-guide`、`record-browser-gif`：它们的价值在**机制**（配对记录、取代检查、迁移指南、可播放证据），不在**体裁与工具链**。机制已经并入，工具流程没有。`record-browser-gif` 只留下一条判据（界面可见变更要附可播放证据），录制与上传的步骤没带。
- `dsh-pr-description` 的 AI Usage 分级表与 AI-DECLARATION 徽章：这是一个组织的声明制度，不是通用软件工程判断。留在通用版里会让所有使用者被迫接受一套外部治理格式。
- `dsh-error-handling` 末尾的三本参考书：它们是好的来源，但不是可执行判据。
- devloop 的编号规则（0001 递增、BL-0001）、frontmatter 字段表、目录树（`docs/vision.md`、`docs/deliveries/000x`、`docs/plans/000x`）、`$_S` 变量：它们是**一个具体流程包的安装形态**。通用版只说「项目入口、交付记录、工作记录」，让使用者按自己的载体映射。

### 刻意没带过来的仓库事实

下列内容只对某个仓库成立，全部留在源材料里，没有进入这七份：

| 类别 | 具体留下的内容 |
|---|---|
| 变更范围命令 | `change-scope --base/--head`、它对 staged/unstaged/untracked 的语义、版本化 JSON 报告 |
| 仓库脚本 | `doc-sync`、`test:docs`、`verify-doc-budgets`、`verify-translation-pairing`、`verify-type-equiv`、`build:artifacts`、`build:check`、`test:snapshot`、`test:web` |
| 覆盖率参数 | `--coverage.include='packages/<group>/<package>/src/**/*.ts'`、按文件 100% 阈值、`vitest related` 的用法 |
| 运行器语义 | `--testTimeout`、describe 级超时对默认值的覆盖关系、`--passWithNoTests` |
| 栈工具 | `gh stack link/merge/sync/checkout/rebase/push` 的逐个用法、`PullRequest.stack` 与 `stackEntry.position` 的 GraphQL 查询、`entries` 分页 |
| 快照布局 | `snapshots/` 与 `tests/expected/` 的归属规则、`session[.vN].jsonl` 与 `session.<ordinal>[.vN].jsonl` 命名、生成代次选择 |
| 决策记录树 | `.agents/notes/{proposed,implemented,rejected,archived}/` 的路径、triplet（正文/译文/一致性记录）、只追加的内容哈希清单文件 |
| 目录约定 | `packages/*/*/src`、`apps/cli/...`、`docs/`、`website/docs.ts` 的投影、`scripts/doc-budgets.manifest.json` |
| 宿主机制 | `cordis.yml` 行、`dsh plugin --profile <name> add`、bundle 的 `dsh.bundle.patch`、Loader 的具名导出要求 |
| 界面常量 | `--dsh-frame-top-clearance`、`[data-platform='darwin']`、视觉 token 名、13px 行高、字重上限 500、`shell.overlay` |
| 双语机制 | `foo.i18n.yaml` sidecar、逐行对齐、`verify-translation-pairing --write` |
| 媒体与发布 | `gh --attach`、5 MiB 分片、浏览器控制工作流、录制帧率 |
| 流程包安装形态 | devloop 的六阶段英文名与编号规则、deep-standard 的 `assets/templates/` 与 `doc-budgets.manifest.json` 文件名、`REVIEW.md` 作为默认载体 |

保留下来的是**机制与判据**：路径即状态、门禁必须自证、证据要能被改动面覆盖、重写要带期望值、发布要看消费者实际拿到的东西。

## 事实与来源

这些 skill 里唯一带版本与日期的断言，都标了来源与核对日期（2026-10-31）：

| 事实 | 来源 |
|---|---|
| GitHub stacked pull request 于 2026-10-06 正式可用 | github.blog 的 changelog（Stacked pull requests generally available），核对 2026-10-31 |
| 不带期望值的 `--force-with-lease` 与后台隐式 `git fetch` 交互很差，保护会失效 | git-push 手册（git-scm.com/docs/git-push），核对 2026-10-31 |

其余规则是判断，不是事实，因此不标日期：它们是「应当怎样」，不是「某版本是这样」。宿主行为与工具行为的断言在换工具链后要重核。

## 维护

复查触发：换工具链或平台（git、CI、PR 平台）、出现漏触发或误触发、满 60 天。

复查时只做三件事：重核带日期的事实、看描述是否仍命中真实请求、删掉不再成立的判断。不为显得在维护而加内容。
