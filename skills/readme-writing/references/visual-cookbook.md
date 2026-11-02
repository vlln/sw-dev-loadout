# 装饰与排版的可复制片段

**何时读它**：准备往 README 里加徽章、深浅色 logo、演示 GIF、表格或折叠区时；或检查现有装饰的写法是否会在宿主上失效。

片段里的 `<owner>/<repo>/<pkg>` 都换成真实值。徽章的最终地址以 shields.io 页面生成为准（核对 2026-10-31）；宿主上限（图片 10 MB、渲染截断 500 KiB 等）见 `SKILL.md` 的"事实"一节。

## 首屏骨架

顺序照抄，内容换掉；不要在这一屏里塞架构说明。

```markdown
# 项目名

一句话定位：它是什么、给谁用。（≤120 字符）

[徽章 1] [徽章 2] [徽章 3]

![演示：在终端里跑 <命令>，输出 <结果>](docs/demo.gif)

## 安装
## 用法
## 文档
## 贡献
## 许可
```

章节多起来之后，定位句下可以加一行精选导航：3–6 个锚点，链到读者最常要的节，跨语言的 README 把语言切换也放这里。

```markdown
[安装](#安装) · [常见问题](#常见问题) · [English](README.en.md)
```

它是精选入口，不是目录：只放读者真的会点的几个，因此不受"手写 TOC 必须与标题逐条一致"的约束——但也意味着放进来的每一个都必须真能跳到。

## 海报式首屏（展示驱动的产品）

产品的价值要靠"看"来判断时，首屏的任务不是让人立刻复制命令，而是让人愿意往下滚。整屏居中构图：hero 图（铺满、深浅两版）、居中标题、居中粗体定位句、一行导航、一排徽章。

```html
<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="docs/assets/hero-dark.webp">
    <img src="docs/assets/hero-light.webp" alt="产品名：一句话定位" width="100%">
  </picture>
</p>

<h1 align="center">产品名</h1>

<p align="center"><b>一句话定位：它是什么、给谁用。</b></p>

<p align="center">
  <a href="releases/latest">下载安装</a> ·
  <a href="#三步上手">三步上手</a> ·
  <a href="#问答">问答</a> ·
  <a href="README.en.md">English</a>
</p>

<p align="center">
  <img alt="宿主与版本" src="https://img.shields.io/badge/<宿主>-<版本>-111?style=flat-square">
  <img alt="平台" src="https://img.shields.io/badge/<平台>-111?style=flat-square">
  <a href="LICENSE"><img alt="MIT" src="https://img.shields.io/badge/MIT-开源-111?style=flat-square"></a>
  <a href="https://github.com/<owner>/<repo>/stargazers"><img alt="Star" src="https://img.shields.io/github/stars/<owner>/<repo>?style=flat-square&color=111"></a>
</p>

<br>
```

- 徽章全部同一个 `style`、同一个色系（上例都是 `flat-square` + `111`）；排在一起要像一套。
- 导航 3–6 个，链到读者最常要的节；跨语言切换也放这里。
- 节与节之间一个 `<br>`，比空行更稳地撑开节奏。
- 命令驱动的工具不要用这套：裸排更快，海报式构图对要复制命令的人是延迟。

## 徽章

来源目录（各生态注册表、质量与安全服务、自定义与私有）见 [badges.md](badges.md)；这里只留最常用的骨架。

动态徽章（值由服务端生成，不会过期）：

```markdown
[![npm 版本](https://img.shields.io/npm/v/<pkg>.svg)](https://www.npmjs.com/package/<pkg>)
[![构建状态](https://img.shields.io/github/actions/workflow/status/<owner>/<repo>/ci.yml)](https://github.com/<owner>/<repo>/actions)
[![许可](https://img.shields.io/github/license/<owner>/<repo>.svg)](./LICENSE)
```

静态徽章（值写死，只用于不会变的事实）：

```markdown
[![标准：standard-readme](https://img.shields.io/badge/readme_style-standard-brightgreen.svg?style=flat-square)](https://github.com/RichardLitt/standard-readme)
```

- 每枚都要 `alt`（方括号里那句），并包一层链接指向能验证它的页面。
- `style=flat-square` 这类参数只用一种，混用会让一排徽章高矮不一。
- 值会变的事实一律用动态徽章；手写"v2.3.1"这种过两周就是错的。
- 想减少外部请求，就把静态徽章下载进仓库，用相对路径引用。

## 深浅色两版的图

```html
<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="docs/logo-dark.svg">
    <source media="(prefers-color-scheme: light)" srcset="docs/logo-light.svg">
    <img alt="项目名" src="docs/logo-light.svg" width="320">
  </picture>
</p>
```

只有当浅色版在深色背景上看不清时才需要两版；只做一版时，选在两种背景上都能看的那版（带自适应的透明底 PNG）。

这套写法不只属于 logo：充当功能展示主角的界面截图同样适用（dsh-media-desk 给每张展示图都配了两版，核对 2026-11-01）。代价是素材翻倍，所以判据不变——在另一种背景上看不清才配第二版，看得清就只留一版。

## 演示 GIF 与截图

```markdown
![演示：`tool init` 生成配置文件并打印下一步提示](docs/init.gif)
```

- 一张图配一句"它展示了什么"，写清操作与结果，不写"效果图"。
- 需要控制显示宽度时用 `<img>`：`<img src="docs/ui.png" alt="…" width="720">`。
- 长操作压成 GIF 的取法：只留关键 3–5 步，剪掉等待，必要时加速；超过 10 s 就剪。

## 演示视频：GitHub 附件嵌入

视频文件不进代码分支（进了历史就永远在克隆里）。走 GitHub 的附件托管：上传后得到一个永不过期的 `user-attachments` URL，README 里裸 URL 独占一行即渲染为播放器。

用 `gh` 上传并拿回 URL（核对 gh 2.102.0，2026-11-01）：

```bash
# 建一个专门的托管 issue，只做一次；置顶，永不关闭
gh issue create --title "README assets" --body "README 媒体托管，勿删。"

# 每个视频发一条评论；正文里写了相对路径引用的话，会被改写成上传后的 URL
gh issue comment <issue号> --attach './demo.mp4' --body 'demo.mp4'

# 读回评论正文，取出 https://github.com/user-attachments/assets/<uuid>
gh issue view <issue号> --json comments --jq '.comments[-1].body'
```

README 里这样嵌（裸 URL，独占一行）：

```markdown
https://github.com/user-attachments/assets/<uuid>
```

- 视频渲染为播放器，**没有 alt 文本**（`gh` 帮助原文），所以视频前后各写一句它展示了什么。
- 上传上限与编码建议见 `SKILL.md` 的"事实"一节；超限就剪短、降码率，或压回 GIF。
- 附件的生命周期不在仓库控制内：上传后源文件自己留档（素材库或孤儿 assets 分支），否则原视频丢失时 URL 里的那份无法从仓库重建。要求媒体可审计、可重建的项目，改用孤儿 assets 分支发布，README 里用带 `?raw=true` 的 blob URL。

## 功能展示单元：截图 + 说明文字

有多个功能要逐个展示时，每个功能一节，节内一个两格 `<table>`：一格截图，一格文字；相邻节左右互换，形成之字形。宿主清洗 class 与 style，这是图文并列唯一剩下的手段——它是这条禁令的唯一例外：表格不得用于纯文字排版。

```html
## 功能名

<table>
<tr>
<td width="62%">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="docs/assets/f-dark.webp">
    <img src="docs/assets/f-light.webp" width="100%" alt="界面哪一部分、正在发生什么">
  </picture>
</td>
<td valign="top">
<p><b>粗体领句：这个功能的一句话价值。</b></p>
<p>一两句展开。引号里给真实用户话术，如「把第 3 镜换成俯拍」。</p>
</td>
</tr>
</table>
```

- 图格约 62%、文格约 38%；`valign="top"` 让文字顶对齐。
- 文字侧一段一个粗体领句，整节读完领句就能读懂功能；引号里的用户话术要真实，不要编。
- 节与节之间一个 `<br>`，比空行更稳地撑开间距。
- 文字格不写实现细节；它回答"这功能替我做什么"，不回答"它怎么实现"。

## 折叠区

```html
<details>
<summary>在 Windows / WSL 上安装</summary>

（这里可以照常写 Markdown：标题、列表、代码块、图片）

</details>
```

- `<summary>` 写清"展开后能得到什么"，不写"点击展开"。
- FAQ 的 `<summary>` 用读者会问的原话（"能直接发布吗？""我的东西存在哪？"），不用主题名（"发布能力说明"）——读者带着自己的问题来，不用你的术语来。
- 折叠区内前后各留一个空行，否则部分渲染器会把 Markdown 当纯 HTML 文本处理。
- 需要默认展开时用 `<details open>`；但默认展开就等于没折叠，只在确实想让所有人先看到时才用。

## 表格

```markdown
| 平台 | 安装方式 | 备注 |
|---|---|---|
| macOS | `brew install <pkg>` | 需要 Xcode CLI 工具 |
| Linux | `sh <(curl -LsSf <url>)` | 不要用 `sudo` |
| Windows | `winget install <pkg>` | PowerShell 5.1 及以上 |
```

- 一格一句话；格子里放代码用反引号，别放整段代码。
- 单元格里的 `|` 要转义成 `\|`，否则整行错位。
- 列宽靠内容自然撑开，不要为了对齐塞空格——渲染后宽度由宿主决定。

## 外部依赖与卸载

读者决定"敢不敢装"之前，想知道两样东西：要花多少钱、接什么账号，以及能不能干净地退掉。

有可选的外部依赖（付费 API、第三方服务、账号）时，给一张依赖表，每个依赖一行，关键列是**不接会怎样**——写降级后的行为，不写"不可用"：

```markdown
## 要花钱吗

| | 用什么 | 不接会怎样 |
| --- | --- | --- |
| **写脚本** | 你已配好的模型 | 只能出结构草稿 |
| **画封面** | 你自己的生图 API Key，在设置里填 | 封面变成纯文字版 |
```

安装节后补一句怎么卸载、数据留在哪。知道能干净地退，是读者敢装的前提之一：

```markdown
卸载时把 `add` 换成 `remove <名>`。你的内容一直在 `~/.<tool>/`，不会跟着删。
```

## 代码块：命令与输出分开

````markdown
```bash
<tool> init --name demo
```

```console
$ <tool> init --name demo
Initialized demo at ./demo
```
````

- 要读者复制的那份用 `bash` / `sh`，不带 `$` 提示符。
- 展示运行结果的用 `console`，可以带 `$`。
- 一份代码块只放一个平台或一个场景，不要把三种安装方式塞进同一块。
