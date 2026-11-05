# 发布平台约束

**何时读**：在 GitHub 使用 HTML、提示块、自动目录或大媒体前。这里记录发布平台事实，不规定所有 README 都采用这些功能。

## GitHub 源码与渲染

以下于 2026-10-31 核对；平台规则变化后重核，不将 GitHub 行为视为所有包页面与 Markdown 渲染器的保证。

| 事实 | 来源 | 对写作的影响 |
|---|---|---|
| 从 `.github/`、仓库根、`docs/` 识别 README，按此前后取一个 | [About READMEs](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-readmes) | 多份并存时确认实际首页，不假设根目录文件一定显示 |
| 超过 500 KiB 的渲染内容截断 | 同上 | 大量内嵌内容应拆分或链接，不把全部参考塞进首页 |
| 按标题生成 Outline | 同上 | 短文不必手写完整 TOC；手写时核对实际锚点 |
| `details` / `summary` 支持折叠，`open` 默认展开 | [Collapsed sections](https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/organizing-information-with-collapsed-sections) | 次路径可折叠，核心定位与主使用路径保持可见 |
| HTML 经清洗，script、inline style、class / id 不可当作可靠布局能力 | [github/markup](https://github.com/github/markup) | 先用渲染平台支持的 Markdown / HTML，检查实际效果 |

## 提示块与媒体

- GitHub alerts 是GitHub 特有扩展，不是通用 Markdown。用前以 [Alerts 文档](https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/basic-writing-and-formatting-syntax#alerts) 核对；不支持时用普通引用块和文字标签。
- 媒体上传限制与编码支持用前查 [Attaching files](https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/attaching-files)。视觉材料应压缩，较长演示可托管，不让附件占据代码历史。
- 深浅色 `picture`、图文表和视频片段见 `$_S/references/visual-cookbook.md`。读它的时候：实际需要相应素材或布局语法，而不是为凑齐风格元素全部加入。
