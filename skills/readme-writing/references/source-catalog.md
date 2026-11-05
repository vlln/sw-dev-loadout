# README 来源目录与分类证据

取证：2026-11-05 UTC。实际读取固定提交的 README 源码与所列关键模板；未渲染浏览器视口、未检查全部媒体或执行安装。布局指 HTML / Markdown 源序；产品能力、指标和支持范围仅是原文自述。

本目录分析结构并保留一行短摘，不分发原文、图片或代码。仓库许可证不自动覆盖品牌和第三方素材；复用实现 / 素材另行核权。无许可的来源仅分析，不整体复制。

## 项目 README 样本

| ID / 固定原文 | 分类与实际结构 | 一行短摘 | 可迁移 / 不照搬 | 公开许可 |
|---|---|---|---|---|
| S01 [omdsh-dev/dsh-genui@c2624ce:README.md](https://github.com/omdsh-dev/dsh-genui/blob/c2624ce0e08fd18510bef640168fcb81e817ed58/README.md) | 强视觉展示 + 完备技术说明；居中 logo/标题文字/徽章/导航 → Showcase → Before vs. After → Quick Start → 能力矩阵 → 例子 → 架构/开发 → FAQ/路线图/社区/许可，许可后另有 maps 小节 | “Buttons without an action render disabled to eliminate ‘dead clicks’” | 学真实产出、前后对比、限时验证和视频退路；不继承性能承诺、包名迁移史、长开发细节与许可后追加功能 | [MIT](https://github.com/omdsh-dev/dsh-genui/blob/c2624ce0e08fd18510bef640168fcb81e817ed58/LICENSE) |
| S02 [MindLab-Research/dsh-generative-ui@ed2db0a:README.md](https://github.com/MindLab-Research/dsh-generative-ui/blob/ed2db0a5d1f90038a5b34b1ca99ebc2b15d09c51/README.md) | 克制技术型；普通标题/两徽章 → Inline/Canvas 场景 → Install → How it works → Development → License；无 hero 或演示图 | “It streams — the component renders while the model is still typing it.” | 学按使用位置解释、host/client 职责、可区分的故障诊断；不是纯极简，也不能因机构名归论文型 | [Apache-2.0](https://github.com/MindLab-Research/dsh-generative-ui/blob/ed2db0a5d1f90038a5b34b1ca99ebc2b15d09c51/LICENSE) |
| S03 [NVIDIA/auto-ontology@2a24f5d:README.md](https://github.com/NVIDIA/auto-ontology/blob/2a24f5d34ef6b6e168a7f22d308504d3ab36a198/README.md) | 业务问题驱动的工程产品型；普通标题/企业问题 → Benchmarks 外链 → 特性/组件 → Quickstart → 原理 → agent 接入 → 连接器 → 集成/参考/许可/安全 | “It complements the semantic models customers already own.” | 学 Quickstart 的问题、答案、SQL 和不可回答负例，参考文档路由；未见论文标题/作者/BibTeX/实验配方，不归论文复现型 | [Apache-2.0](https://github.com/NVIDIA/auto-ontology/blob/2a24f5d34ef6b6e168a7f22d308504d3ab36a198/LICENSE) |
| S04 [github/spec-kit@ec9a6f3:README.md](https://github.com/github/spec-kit/blob/ec9a6f3b94c0aa856f77e550785a1ddde1993de1/README.md) | 品牌首屏 + 任务路由型；居中 logo/标题/定位/徽章 → Choose your process → 能力 → Get started → 各流程 → 自定义 → 文档/新闻/社区 | “These are independent entry points, not mandatory phases.” | 学 need/process/outcome 表、终端与 agent 对话命令分离、产物和成功判据；不将独立入口写成强制流水线 | [MIT](https://github.com/github/spec-kit/blob/ec9a6f3b94c0aa856f77e550785a1ddde1993de1/LICENSE) |
| S05 [vercel-labs/agent-desktop@b43ac4f:README.md](https://github.com/vercel-labs/agent-desktop/blob/b43ac4f6e892ef608f8851428613c4009f0e84c7/README.md) | 安装优先的技术实用型；普通标题/定位/徽章/机制与隐私 → Install → Quick start → 原理 → AI agents → 命令/平台/C库/开发/文档/许可 | “The loop lives in your agent; agent-desktop has no model inside.” | 学权限自检、observe-act-observe 命令配 JSON 输出、限制贴近步骤；有完整命令与平台表，不是纯极简；不继承协议或支持承诺 | [Apache-2.0](https://github.com/vercel-labs/agent-desktop/blob/b43ac4f6e892ef608f8851428613c4009f0e84c7/LICENSE) |
| S06 [thedotmack/claude-mem@d8bf121:README.md](https://github.com/thedotmack/claude-mem/blob/d8bf121ae98063e5e3158befff33a9d635c0246b/README.md) | 品牌/社会证明 + 文档门户型；居中深浅 logo、多语言/徽章、GIF与star图、导航 → Quick Start → 文档 → 原理/搜索/要求/配置/开发/故障/贡献/许可，后有 Support 等 | “Context from previous sessions will automatically appear in new sessions.” | 学深浅品牌图、按读者任务分文档、隐私选择说明；不照搬密集社会证明、主路径漂移、工具数前后不一致及无关推广 | [Apache-2.0](https://github.com/thedotmack/claude-mem/blob/d8bf121ae98063e5e3158befff33a9d635c0246b/LICENSE) |
| S07 [floci-io/floci@143e80d:README.md](https://github.com/floci-io/floci/blob/143e80d52b580ce885a6ed876463b1d6bdc2b2a9/README.md) | 品牌首屏 + 操作手册型；居中双 logo/口号/六徽章/导航 → What is → Quick Start/视频 → Console → 特性/比较/架构/服务 → Docker/存储/隔离/SDK/测试/迁移/配置 → 社区/许可 | “Floci is a free, open-source local AWS emulator for development, testing, and CI.” | 学主安装路径与折叠次路径、服务实现边界、最小迁移差异；它先安装后视频，不是展示优先规则的证据；竞品/性能/永久承诺不能继承 | [MIT](https://github.com/floci-io/floci/blob/143e80d52b580ce885a6ed876463b1d6bdc2b2a9/LICENSE) |
| S08 [max-sixty/worktrunk@78e1357:README.md](https://github.com/max-sixty/worktrunk/blob/78e1357583f1cc6bdfd29ae34f177430244b5e9c/README.md) | CLI 展示 + 完备工作流混合；行内 logo/徽章/定位/GIF → worktree 问题 → 与 git 对比 → 高级工作流 → Install → Quick start → Next steps → Contributing | “Create a worktree for a new feature” | 学任务/新旧命令表、运行反馈与状态解释、主平台外的折叠安装、下一步分流；不继承完整长文、时效宣传或星图参数 | README 标示 MIT OR Apache-2.0；许可全文未在此记录中核验，不作复用授权结论 |

## 通用模板与个人展示集合

| ID / 固定原文 | 真实对象 / 结构 | 一行短摘 | 可迁移 / 不照搬 | 许可边界 |
|---|---|---|---|---|
| S09 [dbader/readme-template@f89960a:README.md](https://github.com/dbader/readme-template/blob/f89960a33fc952bcfe811675540347dfb361a6d5/README.md) | 轻量项目模板；标题/短句/徽章/介绍/图 → Installation → Usage example → Development setup → Release History → Meta → Contributing | “A few motivating and useful examples” | 学安装后给有动机的用例、多平台分块、深例子分流；旧 Windows 命令、Travis、手写历史和空 alt 不照搬 | 根目录未见许可证；不整体复制 |
| S10 [kautukkundan/Awesome-Profile-README-templates@72a8f8e:README.md](https://github.com/kautukkundan/Awesome-Profile-README-templates/blob/72a8f8ed3dca4829ec77c0eb0afa04d70b5e441f/README.md) | 个人 Profile 灵感集合；banner/收集说明/视频/邀请提交。目录按 code-styled、short-and-sweet、multimedia 等视觉手法分类 | “developers around the world are using on their own profiles” | 只借身份简介、可信联系入口、视觉分类；不把个人履历、统计卡、社交信息替代项目安装使用 | API license=null，根目录未见许可证；个别素材另核 |
| S11 [iuricode/readme-template@6b590c0:README.md](https://github.com/iuricode/readme-template/blob/6b590c0aec422ef119d2448a223249d6c4461cf9/README.md) | Profile + 项目模板混合集合；居中标题/定位/徽章/导航 → 资源用途表 → 使用 → Profile/项目入口 → 结构/贡献/许可/致谢 | “Modelos de README prontos para copiar, colar e personalizar.” | 学用途分流、项目与Profile物理分离、多OS安装；不继承低价值指标徽章、重复贡献入口或安装前的大路线图 | [MIT](https://github.com/iuricode/readme-template/blob/6b590c0aec422ef119d2448a223249d6c4461cf9/LICENSE.md)；avancado 改编内容另有 CC BY 4.0 来源说明 |
| S12 [othneildrew/Best-README-Template@fc444eb:README.md](https://github.com/othneildrew/Best-README-Template/blob/fc444eb7b04b2e4863f6080b1507a219e95103fa/README.md) | 通用项目模板；徽章 → 居中logo/标题/CTA → 折叠目录 → About/Built With → Getting Started → Usage → Roadmap → Contributing → License → Contact/Acknowledgments | “Of course, no one template will serve all projects” | 学主/次CTA、前提与安装分开、引用式链接、统一替换清单；不照搬个人品牌、框架徽章、每段回顶、把密钥写源码的示例 | [Unlicense](https://github.com/othneildrew/Best-README-Template/blob/fc444eb7b04b2e4863f6080b1507a219e95103fa/LICENSE.txt)，不是 MIT |

### 集合中实际读取的子模板

- S10：[code-styled/Thaiane.md](https://github.com/kautukkundan/Awesome-Profile-README-templates/blob/72a8f8ed3dca4829ec77c0eb0afa04d70b5e441f/code-styled/Thaiane.md)：问候/GIF → 职位/社交徽章 → JS对象个人简介 → 联系邀请。仅分析表达形式，不复制身份资料。
- S10：[short-and-sweet/M0nica.md](https://github.com/kautukkundan/Awesome-Profile-README-templates/blob/72a8f8ed3dca4829ec77c0eb0afa04d70b5e441f/short-and-sweet/M0nica.md)：问候 → 身份使命 → 联系入口；是“短个人简介”，不是小工具模板。
- S11：[repositorio/exemplo-01.md](https://github.com/iuricode/readme-template/blob/6b590c0aec422ef119d2448a223249d6c4461cf9/repositorio/exemplo-01.md)：标题/徽章/图/简介 → 改进清单 → 前提 → 分OS安装 → 用法 → 贡献/作者/许可。
- S11：[perfil/exemplo-02.md](https://github.com/iuricode/readme-template/blob/6b590c0aec422ef119d2448a223249d6c4461cf9/perfil/exemplo-02.md)：JS类个人介绍 + 技能；没有项目安装章节。
- S12：[BLANK_README.md](https://github.com/othneildrew/Best-README-Template/blob/fc444eb7b04b2e4863f6080b1507a219e95103fa/BLANK_README.md)：应参考的空白项目模板，不只看介绍模板本身的首页；许可链接中的分支名也需迁移核对。

## 极简与科研样本

S13 展示小型库的极简用法组织；S14 展示论文代码的演示与复现路径。科研分类依据论文身份与实验配方，不依据机构名。

| ID / 固定原文 | 观察与短摘 | 用途 / 边界 | 许可 |
|---|---|---|---|
| S13 [sindresorhus/strip-bom@c691ec5:readme.md](https://github.com/sindresorhus/strip-bom/blob/c691ec5724be8bb61e9df269a2577fcba41353f7/readme.md) | 一句任务定位 → 必要背景 → Install → Usage（输入与返回值）→ Related → 支持尾注；“Strip UTF-8 byte order mark (BOM) from a string” | 极简主样本；源文有支持HTML、安装提示符、未标语言围栏、无独立许可节，不把它说成完全无HTML或满足所有skill规则 | [MIT](https://github.com/sindresorhus/strip-bom/blob/c691ec5724be8bb61e9df269a2577fcba41353f7/license) |
| S14 [bmild/nerf@14c5556:README.md](https://github.com/bmild/nerf/blob/14c55567a6d0fbd75d3fd12b0411f98160ba3237/README.md) | 论文/作者/数据/方法图 → TL;DR → Setup → 方法解释 → 训练/渲染/论文复现 → 自有数据 → Citation；“The example config files run at lower resolutions” | 科研主样本；学快速演示与论文配置分开、完成信号与自有数据前提；不继承旧依赖、耗时或数据授权，不冒称原文已有种子统计/局限表 | [MIT](https://github.com/bmild/nerf/blob/14c55567a6d0fbd75d3fd12b0411f98160ba3237/LICENSE) |

## 从观察到规则：不能混淆的层次

- **来源观察**：按上述固定原文陈述，允许不符合本 skill 的规则。品牌首屏可与技术型、完备型、安装优先同时存在，分类不是排他标签。
- **归纳建议**：任务路由（S04）、输入/操作/结果（S05/S08/S13）、主次安装分层（S07/S08）、文档分流（S03/S06/S12）、演示与复现分开（S14）。
- **本 skill 风格约定**：海报式要求居中定位、展示优先与内容驱动的视觉层级；交替图文是可选手法；首图可选，采用时居顶；样本提供各部件的依据，不表示每仓都采用完整链条。S07 先安装、S12 徽章先于logo、S01无标准H1都是实际反例，不能改写观察以迎合规则。
- **本 skill 质量门槛**：License 收尾、徽章上限、可复制无提示符命令、代码语言标签、真实验证与证据边界；不是这批样本的一致事实。

每种风格由 `references/<style>.md` 说明适用场景、来源归纳与验收要求，由 `assets/<style>.template.md` 提供可填充的 README 成品结构。来源分析不写进模板正文；占位符只填目标项目的可核实事实。新增风格应先研究真实来源，再建立对应 reference 与 template。
