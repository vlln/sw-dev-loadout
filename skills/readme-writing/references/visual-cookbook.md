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

## 徽章

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

## 深浅色两版 logo

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

## 演示 GIF 与截图

```markdown
![演示：`tool init` 生成配置文件并打印下一步提示](docs/init.gif)
```

- 一张图配一句"它展示了什么"，写清操作与结果，不写"效果图"。
- 需要控制显示宽度时用 `<img>`：`<img src="docs/ui.png" alt="…" width="720">`。
- 长操作压成 GIF 的取法：只留关键 3–5 步，剪掉等待，必要时加速；超过 10 s 就剪。

## 折叠区

```html
<details>
<summary>在 Windows / WSL 上安装</summary>

（这里可以照常写 Markdown：标题、列表、代码块、图片）

</details>
```

- `<summary>` 写清"展开后能得到什么"，不写"点击展开"。
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
