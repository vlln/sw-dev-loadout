# 调研来源与结论

调研日期 **2026-10-31**。方法：先 `web_search` 找"README 怎么写"这一类里最常被引用的规范、模板、agent skill 与高 star 实例，再 `web_fetch` 读原文。下表每条都实际打开过；读不到的记在最后一节。

## 规范与模板

| # | 来源 | URL | 一句话结论 |
|---|---|---|---|
| 1 | standard-readme（README） | https://github.com/RichardLitt/standard-readme | 它把"README 是别人接触代码的第一入口"讲成一条硬理由：章节标准化只为让读者少找，徽章建议靠顶部且明确提醒"太多徽章会显得杂乱"。 |
| 2 | standard-readme（spec.md） | https://github.com/RichardLitt/standard-readme/blob/main/spec.md | 章节的必需 / 可选与**先后顺序**在这里是硬约束：Short Description 必须 <120 字符，Table of Contents 只对超过 100 行的 README 必需，License 必须最后，且"README 里不该包含断链"。 |
| 3 | Make a README | https://www.makeareadme.com/ | 写法按读者的问题组织（是什么 / 为什么 / 给谁 / 何时 / 放哪 / 怎么装 / 怎么用），并补了 Requirements、Support、Roadmap、Project status 这几节；明确说太长时该转 wiki 或文档站，而不是硬塞。 |
| 4 | awesome-readme | https://github.com/matiassingers/awesome-readme | 一百多份高 star README 的**元素清单**：logo、banner、GIF、TOC、徽章、折叠区、贡献者、架构图各出现在哪些项目里。它是"装饰库存"，也顺手构成反例库——同一条目里既有"精准 3 枚徽章"，也有"徽章墙 + 三屏 banner"。 |

## 同类 agent skill

| # | 来源 | URL | 一句话结论 |
|---|---|---|---|
| 5 | better-readme（FFloriani/skills） | https://github.com/FFloriani/skills/tree/main/skills/better-readme | 一个成熟的 README skill：把"GitHub 机械规则"与"最佳实践"分开陈述；徽章 3–5 枚且每枚回答一个真问题；首屏先答"是什么、解决什么痛、怎么获取"，装饰与支持信息一律后置；并记下了 GitHub 会移除 `style`、`class`、内联 `<svg>`、`<center>`、`<iframe>`、`<script>`。 |
| 6 | crafting-effective-readmes（aiskillstore marketplace） | https://github.com/aiskillstore/marketplace/tree/main/skills/softaworks/crafting-effective-readmes | 按项目类型（OSS / 个人 / 内部 / 配置）分模板，但所有类型都收敛到三节最小集：Name、Description、Usage；先把"谁会读"问清楚再动笔。 |
| 7 | anthropics/skills（仓库 README） | https://github.com/anthropics/skills | 既是 Agent Skills 规范的官方实现仓库，也是一份高 star README 的形态样本：1 枚徽章、无 logo、短段落 + 按使用场景分节（Claude Code / Claude.ai / API），末尾是许可与免责声明，全文没有实现细节。 |

## 宿主与服务的机械事实

| # | 来源 | URL | 一句话结论 |
|---|---|---|---|
| 8 | GitHub Docs：About READMEs | https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-readmes | README 从 `.github/`、根、`docs/` 三处识别且按此优先级取一个；**渲染超过 500 KiB 内容被截断**；渲染页按标题自动生成 Outline 目录；并明确"README 只放让开发者上手所需的信息，更长的文档归 wiki"。 |
| 9 | GitHub Docs：Collapsed sections | https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/organizing-information-with-collapsed-sections | `<details>` + `<summary>` 是官方支持的折叠写法，`open` 可默认展开，折叠区内的 Markdown、代码块、图片照常渲染。 |
| 10 | github/markup | https://github.com/github/markup | 渲染管线会**主动清洗 HTML**：`script`、inline style、`class` / `id` 属性都被移除；代码块的高亮由 `github/linguist` 提供，所以不标语言的围栏没有高亮。 |
| 11 | Shields.io | https://shields.io/ | 动态徽章与静态徽章同一入口，提供 badge-maker 库与自托管镜像——这解释了为什么"值会变的事实必须用动态徽章、静态徽章可以本地托管"。 |
| 12 | GitHub Docs：Attaching files | https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/attaching-files | 图片与 GIF 上传上限 10 MB；视频按计划 10 MB / 100 MB，编码依浏览器，官方推荐 H.264——这是"GIF 要小、视频给链接"的量化依据。 |

## 高 star 实例（观察形态，不是规范）

| # | 来源 | URL | 一句话结论 |
|---|---|---|---|
| 13 | uv（astral-sh） | https://github.com/astral-sh/uv | 首屏 = 标题 + 3 枚徽章 + 一句定义 + 居中深浅色 `<picture>` 演示图 + 斜体说明；安装按平台各给一个独立代码块；功能节用 `console` 块给"命令 + 真实输出"；FAQ 用 `####` 当问句；致谢与许可收尾。全篇没有一行实现细节。 |
| 14 | sindresorhus/awesome | https://github.com/sindresorhus/awesome | 反面参照：正文极少，靠 HTML 居中的 logo、一行导航链接、覆盖全部章节的 TOC 撑起层级；说明"装饰可以很多，但每一件都在替读者做导航"，而不是替作者做展示。 |

## 未能读取的来源

| 来源 | URL | 情况 |
|---|---|---|
| Art of Readme（noffle） | https://github.com/noffle/art-of-readme | 两次尝试（`master/README.md`、`master/readme.md`）均返回 HTTP 404，仓库路径已变；未使用，改以 #1 #2 #3 的规范文本替代它的位置。 |

## 这些结论落到了哪里

- "首屏只回答是什么" → `SKILL.md` 信息层级 + 自检第 1、2 条（依据 #1 #2 #3 #5 #13）。
- "徽章 1–3 枚、上限 5 枚、每枚回答一个真问题" → 视觉层 + 自检第 6 条（依据 #1 #5 #11）。
- "实现细节 / 逐文件说明 / 开发流水 / 手写数字不该进 README" → 替换表与反模式表（依据 #3 #8 #5）。
- "`<details>`、`<picture>`、HTML 清洗、500 KiB 截断、10 MB 图片上限" → 事实一节，全部带核对日期（依据 #8 #9 #10 #12）。
- "装饰要小、要少、要贴住它说明的那句话" → 视觉层与 `references/visual-cookbook.md`（依据 #4 #13 #14）。
