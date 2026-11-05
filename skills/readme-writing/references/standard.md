# 标准开源完备：有来源的风格指南

这是从真实 README 与通用模板归纳的写作参考，不是另造一个项目。来源标记与分析不进入成品；命令、平台、功能和许可全部从目标项目取证。

## 先读哪些来源

| 来源 | 固定原文 | 借鉴重点 |
|---|---|---|
| S12 Best-README-Template | [BLANK_README.md@fc444eb](https://github.com/othneildrew/Best-README-Template/blob/fc444eb7b04b2e4863f6080b1507a219e95103fa/BLANK_README.md) | About → Getting Started（前提/安装）→ Usage → 贡献与许可等入口；原句：“Of course, no one template will serve all projects”来自其首页 |
| S11 iuricode | [repositorio/exemplo-01.md@6b590c0](https://github.com/iuricode/readme-template/blob/6b590c0aec422ef119d2448a223249d6c4461cf9/repositorio/exemplo-01.md) | 简介限长、多OS安装分块；区分个人和项目模板 |
| S04 spec-kit | [README.md@ec9a6f3](https://github.com/github/spec-kit/blob/ec9a6f3b94c0aa856f77e550785a1ddde1993de1/README.md) | 按任务给独立入口，need/process/outcome；原句：“These are independent entry points, not mandatory phases.” |
| S05 agent-desktop | [README.md@b43ac4f](https://github.com/vercel-labs/agent-desktop/blob/b43ac4f6e892ef608f8851428613c4009f0e84c7/README.md) | 安装后自检、观察/操作/再观察、预期 JSON；原句：“agent-desktop has no model inside” |
| S03 auto-ontology | [README.md@2a24f5d](https://github.com/NVIDIA/auto-ontology/blob/2a24f5d34ef6b6e168a7f22d308504d3ab36a198/README.md) | 从业务问题进入，给正例与不可回答负例、参考入口表；它是工程产品说明，不是论文复现型 |
| S02 dsh-generative-ui | [README.md@ed2db0a](https://github.com/MindLab-Research/dsh-generative-ui/blob/ed2db0a5d1f90038a5b34b1ca99ebc2b15d09c51/README.md) | 使用场景、服务端与客户端职责、故障诊断，克制解释机制 |

完整观察与许可边界见 `$_S/references/source-catalog.md`。仅分析结构与短摘，不整体复制模板、代码或素材。

## 目标结构与源文差异

建议主线：简介 → 特性与边界 → 安装 → 快速开始 → 文档导航 → 关键配置 → 贡献 / 行为准则 → License。

这不是声称每个样本都采用此顺序：Best 还包含 Roadmap、Contact、Acknowledgments；iuricode 将改进清单置于安装前；spec-kit 先给任务路由。迁移时保留真正服务读者的区块，不为了“完整”凑齐标题。

### 首屏：给定义、目标和下一步

Best 用居中 logo、标题与主 CTA 建立入口；agent-desktop 用左对齐标题、定位和责任边界直接进入安装。两者都可以用于标准型，不强制改成海报式。

首屏需要的内容槽位：

- 项目名和一句话定义：对象、动作、目标读者。
- 一段必要背景：读者为什么需要它，什么情况下不适用。
- 一个主要入口：开始使用；成熟项目可再加文档、报告问题。
- 有可核实状态才选徽章；不把贡献者数量、语言数量与下载数全部堆齐。

如果读者有不同任务，借 spec-kit 的路由表，列“我要做什么 / 从哪里开始 / 会得到什么”。只在入口真正独立时使用，不能把可选流程写成强制阶段。

### 特性：能力必须伴随边界

每条特性写“能完成的任务 + 影响使用的条件”。不要按内部模块列目录。

agent-desktop 明确模型循环在调用方；迁移时同样应说明本项目负责什么、不负责什么。auto-ontology 则通过具体问题与答案说明能力边界：负例不是缺乏自信，而是避免误用。

比较表按需求维度写，不按未经核实的“更快、更安全”排名。对竞品许可、性能、生命周期的断言必须另行取证；没有证据就只描述自己的取舍。

### 安装：前提、主路径、确认成功

Best 把 Prerequisites 与 Installation 分开；iuricode 按 OS 拆命令。迁移时：

1. 从清单与 CI 确认运行时、平台、权限及可选服务。
2. 给一个推荐安装入口；其他平台独立代码块或折叠，别在同块混三种 shell。
3. 紧接版本、状态或权限自检；只有实现存在才写该命令。
4. 指出成功信号与最常见的下一步，不把“安装包”混同于“插件已激活”。

目标项目包含应用与插件时，借 S02 的职责拆分：安装在哪一侧、哪一步需重启、什么故障表示模型不知道格式而非渲染器失效。

### 快速开始：一条完整且可检查的任务

借 agent-desktop 的 observe-act-observe，而不是继承它的 snapshot/ref 协议：先取得可操作状态 → 执行动作 → 再看结果。

目标 README 的最小闭环应包含：

- 已满足的前提或一份可公开的示例输入。
- 实际命令 / 导入 / 初始化，不能只剩省略号。
- 预期结果或产物位置，以及如何识别失败。
- 有效性边界，如权限失效、状态过期、输入无法回答；放在相关步骤旁。

库项目用“输入→API调用→返回值”替代终端循环。不要因为参考项目是 CLI，就给库虚构命令入口。

多任务项目从 spec-kit 学习区分“终端执行”和“agent 对话执行”，并指出产物位置和任务完成判据。仅展示命令名不足以教会使用。

### 文档导航：按问题分流

auto-ontology 的 Reference 和 claude-mem 的 Documentation 都将深文档按用途组织。成品可以用以下字段设计导航：

| 读者问题 | 入口需要回答的内容 |
|---|---|
| 如何跑第一个例子 | 入门指南、完整样例 |
| 参数 / API 怎么查 | 配置参考、API 契约 |
| 出错怎么定位 | 排障、权限 / 平台说明 |
| 升级有何不兼容 | 迁移指南、CHANGELOG |
| 如何参与 | CONTRIBUTING、行为准则、安全报告 |

只链接真实存在的入口。目录树不等于导航，不能把所有文件列出来就交付。

### 配置：默认值影响读者行为

仅列首次使用最关键的配置。每项给名称、默认行为、何时改、影响或限制；优先级只有在目标实现确实时写。

完整字段表链接到参考文档。账号或外部服务说明获取与配置渠道，不能复制 Best 示例里把密钥放源码的做法，也不要在文档中放真实秘密。

### 维护入口与 License

有真实计划才写 Roadmap；有实际政策才链行为准则与安全入口。贡献说明优先讲最小复现、回归测试和兼容性要求，详细开发环境移至贡献指南。

Best 原文 License 后还有 Contact/Acknowledgments；本 skill 将许可收尾作为自己的组织约定，不声称源文亦如此。目标项目许可独立核对，不继承 Best 的 Unlicense 或其他样本的 MIT。

## 不照搬清单

- Best 全套技术栈徽章、个人 LinkedIn、每节回顶、示例 URL 与许可链接分支名。
- iuricode 安装前的大改进清单、重复贡献入口、头像和低价值仓库指标。
- auto-ontology 的品牌、模型端点、数据库与接收贡献政策。
- spec-kit 的全部流程命令、生态支持矩阵及新闻 / star 图。
- 任意来源的输出、平台承诺、隐私结论或指标；它们都不自动成为目标事实。

## 验收重点

必要信息齐全而不填空；首次主路径短；文档可分流；正例和边界可核实；安装工具、模型与目标应用的责任分清。标准型“完备”不意味着强制大图，也不意味着源 README 的全部章节都要保留。

## 配套模板

读取 `$_S/assets/standard.template.md`，按该风格的实际 Markdown / HTML 结构填入目标项目事实。模板中的 `{{...}}` 为待替换槽位，HTML 注释仅供编辑；成品必须替换槽位、删除注释及不适用的可选区块。
