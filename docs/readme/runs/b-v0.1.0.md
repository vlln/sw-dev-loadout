# dsh-motion

DeepSeek Harness 的动画插件：agent 在工作区写 `motion/timeline.json` 和每一幕的场景代码，引擎把它逐帧渲成带音轨的 MP4。

[![npm 上的版本](https://img.shields.io/npm/v/@vlln/dsh-motion?label=npm)](https://www.npmjs.com/package/@vlln/dsh-motion)
[![MIT 许可](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

![右侧栏的动画面板正在播放一条 11 秒的片子：顶部是版本下拉，底部是场景条](docs/demo/ui.gif)

*右侧栏的「动画」面板：播放条、场景条、版本下拉都在这一屏。*

它解决的是「agent 会写代码，却交不出一条能交的片子」：时间线是数据，场景是时间的纯函数，交付有一组只断言事实的门禁。

片子住在会话的工作区里，你在右侧栏看着改，也可以直接让 agent 改。引擎不预置画风，也没有模板——画法由场景代码写，`kit/` 里的样例可以整份删掉。项目还很年轻：0.2.0 是第一个能装能用的版本。

## 安装

### 装进 dsh（在会话里做片子）

要求：

- Node.js ≥ 22（`package.json` 的 `engines`）；
- DeepSeek Harness 的 `web` profile；
- `ffmpeg` 与 `ffprobe` 在 `PATH` 上（出片、拆参考都要用）；
- 一份 Chromium。渲染由 Playwright 1.64 驱动，macOS 上会去 `~/Library/Caches/ms-playwright` 找 Chrome for Testing / Chromium；找不到或不在 macOS 上时，用 `CHROMIUM` 指向可执行文件，或先 `npx playwright install chromium`。

```bash
dsh plugin --profile web add @vlln/dsh-motion
```

装完重启 `dsh web`。pnpm 会打印一行确认：

```console
+ @vlln/dsh-motion ^0.2.0
```

重启后会话里多出三样东西：

- 一份 `dsh-motion` skill，agent 靠它学这个库；
- 右侧栏引导里的「动画」面板，显示当前会话工作区的 `motion/`；
- 命令行入口：包内 `bin/dsh-motion.mjs`；插件把它所在的 `bin/` 加进宿主进程的 `PATH`。

装好之后怎么用，见「在 dsh 里用」。想先不装插件、直接跑一条，看下一节和「五分钟做一条片子」。

本页的安装与命令在 macOS、Node 22.23、dsh 0.2.1-rc.5 上实测。仓库文档记录的验证是 dsh 0.2.1-rc.1；插件机制面（服务名、槽名）跨 rc 可能改名，升级后若启动日志出现 `entry did not activate`，先核对 `webServer` 与 `slots` 两个服务。

### 从源码跑（只想先看看）

```bash
git clone https://github.com/vlln/dsh-motion.git "$HOME/dsh-motion"
cd "$HOME/dsh-motion" && npm install
```

这一步装的是 `playwright` 与 `esbuild`，渲染与打包要用。之后所有命令都可以用 `node "$HOME/dsh-motion/bin/dsh-motion.mjs"` 调用，下面把仓库路径记作 `$MOTION`。

## 五分钟做一条片子

```bash
export MOTION="$HOME/dsh-motion"
mkdir "$HOME/my-film" && cd "$HOME/my-film"
node "$MOTION/bin/dsh-motion.mjs" init --title "First cut"
```

```console
motion/timeline.json  · 0 scene(s)
motion/镜头表.md     ← write this first: the look in three lines, then one row per shot
motion/kit/           ← samples you may copy, change or delete

next: 填镜头表 → 写 scenes/<id>.mjs（导出 mount/render）→ dsh-motion check
```

`init` 只给三样东西：空时间轴、一张镜头表、一份可删的 `kit/`。它不给场景，也不替你选画风。

先写一幕。建目录：

```bash
mkdir -p motion/scenes
```

`motion/scenes/title.mjs`：

```js
export const id = 'title';
export const slots = { title: { default: 'Hello motion', role: 'title', maxChars: 60 } };

let ctx;
export function mount(context) { ctx = context; }

export function render(t) {
  const g = ctx.canvas.getContext('2d');
  g.fillStyle = ctx.color('bg');
  g.fillRect(0, 0, ctx.width, ctx.height);

  // 一秒一个来回的呼吸：画面每帧都在变，确定性仍然只来自 t
  g.fillStyle = ctx.color('accent');
  const w = ctx.width * (0.42 + 0.12 * Math.sin(t * 1.5));
  g.fillRect(ctx.width * 0.08, ctx.height * 0.62, w, 12);

  ctx.drawText('title', {
    x: ctx.width * 0.08, y: ctx.height * 0.38, w: ctx.width * 0.6,
    size: Math.round(ctx.height * 0.09), weight: 700, lineHeight: 1.15,
  });
}
```

把它排进片子。`motion/timeline.json` 整份换成：

```json
{
  "width": 1920,
  "height": 1080,
  "fps": 60,
  "duration": 6,
  "palettes": {
    "main": {
      "bg": "#141210",
      "fg": "#f2efe8",
      "accent": "#d8a13a",
      "muted": "#8b98ad",
      "grid": "#243450",
      "panel": "#101c33"
    }
  },
  "texts": { "title": { "title": "Hello motion" } },
  "scenes": [
    { "id": "title", "name": "Title", "start": 0, "length": 6, "palette": "main", "speed": 1 }
  ]
}
```

```bash
node "$MOTION/bin/dsh-motion.mjs" check
```

```console
ok · 1 scene(s) · 6s · 1920x1080 @ 60fps
   0.23 ms  title

ok · every scene under 33 ms of draw cost (worst 0.23 ms, title)
look · 3s · mine solid 0.98 · texture 4.6 · contrast 19.6
  local detail 4.6 at contrast 19.6; the exemplars measure 14–26.
  on a dark frame a lower number is expected — this alone cannot say "painted".
```

前两段是校验与单帧耗时（数字随机器变）。末尾那三行 `look` 是画面构成的线索，不是分数，也不当门槛用。

先看静帧，别急着渲整条：

```bash
node "$MOTION/bin/dsh-motion.mjs" render --stills 0,3 --cols 2
```

```console
sheet → out/sheet.png (2 frames in out/sheet-frames)
```

接下来配音、出片、量片：

```bash
cat > motion/audio.json <<'EOF'
{
  "layers": [
    { "kind": "roomtone" },
    { "kind": "pad", "bpm": 56 },
    { "kind": "page", "at": 3 }
  ]
}
EOF
node "$MOTION/skills/dsh-motion/scripts/make-audio.mjs" --write
```

```console
assets/audio/bed.wav  6s · 48k stereo · peak -6.0 dBFS · 113 ms
wave → out/audio/wave.png  (3 layers, 1 cues)
level ▃▄▇█▆▄   1 block = 1s
timeline.json ← audio.src assets/audio/bed.wav · beats 56 BPM
```

```bash
node "$MOTION/bin/dsh-motion.mjs" render --video
```

```console
audio → assets/audio/bed.wav (trim, loudnorm -20 LUFS)
video → out/video.mp4
```

```bash
node "$MOTION/bin/dsh-motion.mjs" check --deliver
```

```console
ok · 1 scene(s) · 6s · 1920x1080 @ 60fps · audio assets/audio/bed.wav
ok   there is a picture  — 1920x1080 @ 60/1
ok   there is a sound  — aac 48000Hz 2ch
ok   it matches the timeline length  — 6.00s vs 6s
ok   nothing goes blank for long  — longest empty run 0.00s
ok   the picture moves  — 100% of frames differ
ok   not one blue-violet haze  — 0% of pixels
ok · deliverable · out/video.mp4
```

收尾存一版，用户在面板顶部的版本下拉里靠这句话找回这一版：

```bash
node "$MOTION/bin/dsh-motion.mjs" version -m "first cut"
```

```console
v1 · first cut
```

到这里片子就能交了。之后每改一处 `motion/scenes/title.mjs`，再跑一次 `check` 与 `render --stills`；整条视频是逐帧渲染，慢。

## 它是怎么工作的

一条片子是「一份数据 + 每幕一个模块」：

- `motion/timeline.json` 存尺寸、帧率、总时长、幕表（每幕的 `start` / `length`）、色板、文字槽、数值、素材和音轨。改字、改色、改时长都只动这份文件；
- `motion/scenes/<id>.mjs` 决定这一幕怎么画，导出 `id`、可选的 `slots` / `params` / `values` / `duration`，以及 `mount(ctx)` 与 `render(t)`。

硬规则：`render(t)` 必须是 `(t, 只读数据) → 像素` 的纯函数。不读 `Date.now()`、不用 `Math.random()`、不在渲染阶段发网络请求；要随机就用 `ctx.rng('key')`。同一个 `t` 渲两次必须逐像素一致，门禁 G1 守着这条——所以预览看到的那一帧就是导出的那一帧。

`ctx` 给的东西不多：画布与尺寸、`ctx.color`、`ctx.text` / `ctx.drawText`、`ctx.value` / `ctx.drawValue`、`ctx.rng`、`ctx.still`、`ctx.image` / `ctx.frame`、`ctx.beats`。接口细节见[内核契约](docs/kernel-contract.md)。

画法住在哪，是项目里的一条硬线：

| 层 | 装什么 | 规则 |
|---|---|---|
| `src/` | 引擎：时间、画布、确定性、缓存、编码、音频、字体、素材 | 没有一行画画代码；升级会覆盖 |
| `kit/` | 可抄的样例：笔触、媒介、像素、透视、骨架、后期 | 可改、可删、可以一个都不留；引擎永不 import |
| `cards/` | 一个风格一张卡：管线、色板、签名特征、坑、短板、状态 | 加一张卡，引擎不加一行 |
| `howto/` | 做法的规程：新风格、参考、角色、镜头表、审片、交付 | 给人读，程序不读 |

出片是「逐帧页面截图 → ffmpeg」：`init` 给的脚手架是 60fps，编码 H.264 + AAC；画布在创建时钉死 sRGB，否则 macOS 的 display-P3 会让颜色读回偏移。

![一条导出的成片：地铁站台、站牌，风把围巾吹起来](docs/demo/platform.gif)

*一条导出的成片，GIF 是其中 5 秒。*

引擎不认识站台、人、围巾——它们都写在场景代码里。

## 在 dsh 里用

装好插件、重启之后，工作流是「在会话里说要做一条片子」：

> 做一条 10 秒的片头：一杯茶凉掉的过程，最后一帧停在「凉了」。

skill 要求 agent 先写 `motion/镜头表.md`（画风三行 + 每镜一行），再写 `motion/scenes/*.mjs`，自己跑 `dsh-motion check`、`render`、`check --deliver`，最后落一个 `version`。它读的是插件注册的那份 skill，不需要你了解时间线格式。

- **片子住在会话的工作区里**：工作区下的 `motion/` 目录，不在插件安装目录。一个工作区一条片子。
- **没有工具面**：插件不注册任何 agent 工具。agent 的工作面是文件加命令行，面板是给人的。
- **面板**：右侧栏的「动画」标签，常驻在对话旁边，显示这个会话的片子。
  - 播放 / 暂停、拖播放头、点场景条跳转；
  - 点画面里的元素，能看它的位置与尺寸（只读）；
  - 改文字槽、数字和参数滑杆：拖动即时改画面，松手才写回 `timeline.json`，越界直接拒绝；
  - 顶部版本下拉：切换版本会把时间轴和场景源码一起换回去，换之前先把当前状态存一份隐藏副本；
  - 批注：在暂停帧上框一块、写一句，发送后它就是**你发的一条消息**，排队进这个会话；钉子（时刻、区域、元素、状态）留在片子里，拖回那一帧还能看见指的是哪儿；
  - 亮暗色跟随宿主。
- 面板只显示它所属会话的 `motion/`。独立打开预览时没有会话，批注入口不会出现。

## 命令

五个动词，一个阶段一个：

| 动词 | 做什么 |
|---|---|
| `init [--title "标题"]` | 建一条空片子：`timeline.json`、`镜头表.md`、`kit/`。没有场景，没有画风。 |
| `reference <video> [--fps 2]` | 把参考片拆成帧、接触表、切点和简报，落在 `motion/out/ref/`；加 `--shots --write` 把切点变成分镜表。 |
| `check` | 校验时间轴、量每幕的单帧绘制耗时（默认 33ms 预算）、给三行画面构成的线索。`--deliver` 量成片文件，`--perf` 只量性能。 |
| `render` | `--stills 0,3,6` 出拼图，`--still 2.5` 出单张全分辨率帧，`--video` 出 MP4，`--export [--gif]` 出交付版（可配 `--quality` / `--height` / `--fps`）。 |
| `version` | `-m "改了什么"` 存一版（时间轴 + 全部场景源码）；`version list` 列出来；`version use v3` 切回去。 |

命令都在工作区根目录跑，读写落在 `motion/` 里。`check --deliver` 的六条红灯是**事实**，不是观感：

| 红灯 | 含义 |
|---|---|
| 没有画面 | 文件里没有视频流 |
| 没有声音 | 文件里没有音轨 |
| 长度不对 | 片长与时间线差 0.5 秒以上 |
| 空画面 | 中段有连续 0.3 秒以上的近乎纯色 |
| 幻灯片 | 一半以上的相邻帧完全一样 |
| 蓝紫雾 | 蓝紫像素占全片 30% 以上（题材需要时用 `--allow-blue-purple "理由"` 放行） |

构图好不好、像不像那个风格、值不值得看，它不判。那三样走[交付流程](howto/交付.md)的第二道（并排图 + 审片四问）和第三道（独立审片）。

## 声音

默认要求有声音：`check --deliver` 把「没有音轨」判红灯。两条路。

**合成**：写一张提示单 `motion/audio.json`，交给 `make-audio` 渲染（就是上面那条片子的做法）。

- 铺底：`rain` `roomtone` `wind` `pad` `clock` `score`；事件：`page` `bell` `wheel` `click`。
- 不带 `at` 的铺底层铺满全片；事件层用 `at` 指定第几秒。铺底层带 `at` 是从那一秒铺到片尾，停不下来。
- `pad` / `score` 会把 `timeline.beats` 写上，场景里用 `ctx.beats` 踩点。
- `score` 是一小段编曲：`bpm`、`key`、`chords`、`motif`、`sections`（每段以八分音符数计，换段换音色）。

**自带一条轨**：文件放进 `motion/assets/audio/`，声明它，必要时循环：

```bash
node "$MOTION/skills/dsh-motion/scripts/make-audio.mjs" --use assets/audio/theme.wav --loop --write
```

`theme.wav` 换成你的音轨文件。两条路都落在 `timeline.audio = { src, targetLufs, fit }`。`fit: "trim"` 只放一遍、余下静音；`fit: "loop"` 循环到片尾。片的长度永远由画面决定。

## 参考与风格

引擎里没有风格可选。要做「像某个风格」，先把它变成数字，再写成代码。

```bash
node "$MOTION/bin/dsh-motion.mjs" reference ref.mp4 --fps 2
```

把 `ref.mp4` 换成你的参考片。`motion/out/ref/` 下会得到帧（480 宽）、接触表 `sheet.jpg`、`index.json`（时长、画幅、切点）和 `brief.json`。把切点变成分镜：

```bash
node "$MOTION/bin/dsh-motion.mjs" reference ref.mp4 --shots --write
```

切点之间的段落成为 Beat，过短合并、过长拆分；它写出一张 `motion/out/ref/镜头表.md`，时间段已经填好，「物」和「动词」留给你填。

色板也一样，量而不是挑：

```bash
node "$MOTION/scripts/palette-from.mjs" ref.png --colors 8
```

`ref.png` 换成一张参考帧。它按像素占比列出主色（hex + 占比）；加上 `--out motion/out/palette.png` 还会画一条色板图。

- 做一个新风格的六步：[新风格](howto/新风格.md)
- 色板、源码、人物三条「能搬就搬」的做法：[从参考里取](howto/从参考里取.md)
- 人由生图给，代码只负责合成：[角色](howto/角色.md)
- 审片问句与重复缺陷：[审片](howto/审片.md)
- 现有风格卡：[卡索引](cards/INDEX.md)——目前只有一张埃及墓室壁画，状态是「研究配方」

Agent 的手册是 [`skills/dsh-motion/SKILL.md`](skills/dsh-motion/SKILL.md)，插件把它注册成会话里的 skill。

## 已知限制

- 一个工作区一条片子；幕由 agent 写，面板只能查看和改，不提供「再加一幕」。
- 同一个场景 id 在一条时间轴里只能出现一次；要重复同一套版式，得写变体模块。
- 场景必须从 0 开始、首尾相接、正好铺满 `timeline.duration`。
- 参数只支持数值；颜色走色板，文案走文字槽。
- 一条音轨：没有旁白轨，也没有 ducking。
- 改场景代码后面板不会热应用（浏览器缓存模块），会出现「场景代码已更新」提示条，需要点刷新。
- 导出是 H.264 MP4，没有 alpha 通道。
- 参考摄取只读画面：不读音频、不做字幕 / OCR、不懂语义；`soft` 切点是启发式的，宁多勿漏。
- 长片段会先烘成一张雪碧图整段装进内存，没有分块加载。
- 它保证的是下限（有画面、有声音、长度对、能交付），不保证上限（像不像那个媒介）；后者只能靠并排图和独立审片判断。
- WebGL 的确定性只在 Apple GPU + Chrome for Testing 上验证过，换机器或驱动需要重跑门禁。
- 面板入口依赖宿主 client 的 `slots` 服务；更旧的 dsh 没有它，入口不会出现。
- 0.2.0 的 `PATH` 注入指向包内 `bin/`，而那里面只有 `dsh-motion.mjs`。如果会话里没有 `dsh-motion` 命令，用 `npx @vlln/dsh-motion`，或直接 `node` 跑插件安装目录里的 `bin/dsh-motion.mjs`。

## 开发

```bash
npm install           # playwright + esbuild
npm run check         # 串起全部门禁：时间线校验、画面、音频、面板、导出、插件安装等
npm run preview       # 本地预览 → http://127.0.0.1:4173/preview.html
npm run video         # 仓库自带的示例片 → out/demo.mp4
npm run sheet         # 静帧拼图 → out/sheet.png
npm run build:client  # 改完面板源码后重新打包 lib/
```

- `npm run check` 是唯一的回归保护：改引擎、改面板、改交付路径之前先跑它。每条门禁测什么，见[门禁清单](docs/gates.md)。
- 仓库根目录的 `timeline.json` 是引擎自检片（两幕），`npm run preview` 打开的就是它。
- 面板的源码在 `src/`，打包产物在 `lib/`；`scripts/` 是出片、抽帧、门禁与一次性工具。
- 设计取舍记在[决策记录](docs/decisions.md)，版本变化见[变更记录](CHANGELOG.md)。

## 许可

MIT（SPDX：`MIT`），见 [LICENSE](LICENSE)。
