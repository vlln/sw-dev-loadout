# dsh-motion

<p align="center">dsh（DeepSeek Harness）插件：片头、讲解片、数据动画——片子写在会话工作区里，agent 用代码画，你在旁边看片、批注、改字，成片是 MP4。</p>

<p align="center">
<a href="https://www.npmjs.com/package/@vlln/dsh-motion"><img alt="npm 版本" src="https://img.shields.io/npm/v/@vlln/dsh-motion.svg"></a>
<a href="LICENSE"><img alt="许可：MIT" src="https://img.shields.io/github/license/vlln/dsh-motion.svg"></a>
</p>

<p align="center"><img src="docs/ui-panel.jpg" width="100%" alt="dsh 里的 dsh-motion：中间是正在播的片子，右边是批注、这一帧的文字与参数，底部是按幕切分的时间轴"></p>

## 装进 dsh

一条命令，装进你正在用的 profile（下面以 `web` 为例）：

```bash
dsh plugin --profile web add @vlln/dsh-motion
```

```console
$ dsh plugin --profile web add @vlln/dsh-motion
dependencies:
+ @vlln/dsh-motion ^0.2.0
Done in 314ms using pnpm v11.10.0
```

重启 dsh，右侧栏能打开一个「动画」页签，里面就是这个会话工作区的片子。它不注册工具：
agent 像写代码一样写 `motion/`，再用 `dsh-motion` 命令自己跑。

卸载用 `dsh plugin --profile web remove @vlln/dsh-motion`；只要命令行、不要面板与 skill：
`npm i -g @vlln/dsh-motion`。

## 说一句话，开始做片

在会话里说清楚要什么片子就行：

> 做一条 8 秒的片头：黑底，一枚硬币在桌上转到停下，最后停在标题上。

<p align="center"><img src="docs/demo/platform.gif" width="560" alt="一条 5 秒的示例片：地铁站台上站着一个人，随后切到隧道口"></p>

（本仓库示例片的一镜：地铁站台，不是上面那句需求的结果。）

agent 会先写 `motion/镜头表.md`（画风三行，每镜一行：一个物、一个动作），再写 `motion/scenes/*.mjs`，
自己跑 `check` 和 `render`。片子住在工作区：

```text
motion/
  timeline.json   幕、时长、调色板、文字与数值
  scenes/*.mjs    每一幕怎么画
  assets/         素材
  out/            渲染产物
```

出片用 `dsh-motion render --video`（默认 1920×1080 / 60fps，落 `out/video.mp4`）。交付前跑
`dsh-motion check --deliver`：它只断言能证明的事——有画面、有声音、长度对得上、中段不空、
不是幻灯片、不是蓝紫雾。

## 看着不对，就在面板里动手

**框一句，就是批注。** 暂停在某一帧，在画面上拖出一个框、写一句；它作为一条消息进对话，
钉子留在那一帧——agent 不用猜你说的是哪儿。

<p align="center"><img src="docs/ui-comments.jpg" width="100%" alt="暂停在 0:03.20，画面上框住一块区域，右侧批注列表里多了一条「把标题改成更短的说法」"></p>

**点画面里的字，输入框跟着跳。** 改完即时重绘，失焦落盘；每次写前留一份快照，能回退。

<p align="center"><img src="docs/ui-focus.jpg" width="100%" alt="点右侧 title 输入框时，画面上的标题文字被虚线框标出来"></p>

**数值参数，自己拖。** 拖动只改内存，松手才落盘；越界直接拒写，不静默夹到边界。

<p align="center"><img src="docs/ui-params.jpg" width="100%" alt="参数面板里六个滑杆：轴线位置、卡片离轴、双行间距、卡片宽度、卡片高度、入场时长"></p>

**每轮改完落一个版本。** agent 收尾时跑 `dsh-motion version -m "用户看得见的变化"`，面板顶部可以切回去；
切换不多出条目，只给被替换掉的状态留一份隐藏副本。

**没有模板。** 内核只有时间、画布、确定性、字体、素材、出片；风格是这次任务里现写的代码——
所以版画、信息图、三维物体能共用一个引擎。

<p align="center"><img src="docs/sheet.jpg" width="100%" alt="两条片子的抽帧：一条是信息图风格，一条是一枚金色绳结的三维片"></p>

## 命令

| 命令 | 做什么 |
|---|---|
| `dsh-motion init [--title "…"]` | 建 `motion/`：空时间轴、`镜头表.md`、可整删的 `kit/` |
| `dsh-motion reference <video> [--fps 2]` | 读参考片：抽帧、切点、配色，写一份 brief |
| `dsh-motion check [--deliver [file]]` | 查时间轴与每幕单帧耗时；`--deliver` 量成片的事实 |
| `dsh-motion render --stills 0,3 \| --still 3 \| --video \| --export [--gif]` | 拼图 / 静帧 / 成片 / 导出，默认落 `out/` |
| `dsh-motion version -m "…" \| list \| use v3` | 落版本 / 列版本 / 切回某个版本 |

## 常见问题

<details>
<summary>要装什么？</summary>

Node ≥ 22、ffmpeg；渲染用 Playwright 缓存里的 Chrome for Testing，或用 `CHROMIUM` 指一个可执行文件。
在 macOS / Apple 芯片、Node 22、ffmpeg 8.1.2 上验证过，Windows 与 Linux 未验证。

</details>

<details>
<summary>check --deliver 报「蓝紫雾」？</summary>

默认调色板的深蓝底 `#0b1426` 正好落在 30% 规则里。题材确实需要时用
`dsh-motion check --deliver --allow-blue-purple "理由"`。

</details>

<details>
<summary>改了场景代码，面板怎么没变？</summary>

浏览器缓存了场景模块，点一次刷新（会先弹一条「场景代码已更新」）。如果顶部标着「制作中 · 显示 vN」，
说明工作区领先于面板载入的那一版——让 agent 落一个新版本就追上了。

</details>

<details>
<summary>一个工作区能放几条片子？</summary>

一条。幕由 agent 写，面板能看和改，不能加幕；同一场景模块在一条时间轴里只能出现一次，复用版式要写变体。
参数只有数值——颜色走调色板、文案走文本槽。

</details>

## 文档

| 去哪 | 读什么 |
|---|---|
| [`skills/dsh-motion/SKILL.md`](skills/dsh-motion/SKILL.md) | agent 的完整手册：目录、时间轴字段、场景契约、`ctx` 原语 |
| [`docs/kernel-contract.md`](docs/kernel-contract.md) | 内核契约：场景模块、文本槽、参数、调色板、错误与降级 |
| [`docs/limits.md`](docs/limits.md) | 已知限制全集 |
| [`docs/gates.md`](docs/gates.md) | 仓库自己的回归门禁清单，`npm run check` 一次跑完 |
| [`howto/交付.md`](howto/交付.md) | 交付三步：事实门禁 → 并排对照 → 独立审片 |

画法样例在 [`kit/`](kit/)，风格卡在 [`cards/INDEX.md`](cards/INDEX.md)，做法规程在 [`howto/`](howto/)。
改引擎或面板：`npm install`，`npm run preview` 起本地预览，`npm run check` 跑全部门禁。

## 许可

MIT。见 [LICENSE](LICENSE)。
