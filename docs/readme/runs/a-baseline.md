<img src="icon.svg" width="44" align="right" alt="dsh-motion">

# dsh-motion

给 [dsh](https://github.com/deepseek-ai/deepseek-harness) 用的动画插件：**agent 在工作区里写一条时间轴和几段场景代码，你拿到一条逐帧可重放的 MP4。**

`@vlln/dsh-motion` 0.2.0 · MIT · Node ≥ 22 · [更新记录](CHANGELOG.md)

画面不来自模板，也不来自一组「加动画」的工具调用，而来自 agent 写的 `render(t)`——一个把秒数映射成像素的纯函数。插件负责把这个函数跑起来、把结果编码成片、把它交给人看和改。

| 部分 | 给谁 | 是什么 |
|---|---|---|
| 内核 | 场景代码 | 时间轴、画布、确定性、缓存、字体、素材、音频、出片（`src/`） |
| 手册 | agent | [`skills/dsh-motion/SKILL.md`](skills/dsh-motion/SKILL.md)，装成插件后注册为 skill |
| 命令 | agent | `dsh-motion`：`init` / `reference` / `check` / `render` / `version` |
| 面板 | 人 | 对话右侧的「动画」页：看片、改这一帧的文字与数值、拖参数、批注、切版本、导出 |

一条片子住在**会话工作区**的 `motion/` 里，不在插件的安装目录里——所以所有会话不共享一部片，改画面也不用改插件源码。

## 适合谁

| 适合，如果 | 不适合，如果 |
|---|---|
| 你要的是**能编程控制的画面**：版画、图解、数据动画、片头，而不是拖时间轴剪素材 | 你要的是一个 GUI 剪辑器（这里没有时间轴编辑器，也无法在面板里「再加一幕」） |
| 你接受「风格要自己写、或者从参考里量出来」 | 你想要一个内置模板库，选一个就出片 |
| agent 能执行命令（有 shell） | agent 不能跑命令：这套形态靠「写文件 + 跑命令」闭环 |
| 机器上能装 `ffmpeg`，并有可用的 Chromium | 纯浏览器沙箱、装不了外部二进制 |

`docs/demo/` 里的《站台》来自一次冷启动实验——一个没参与制作这个仓库的 agent，用这套东西做出的一条 11 秒片子（下面是无声 gif）：

![示例片《站台》的一个片段](docs/demo/platform.gif)

## 安装

### 前置

| 依赖 | 用途 | 说明 |
|---|---|---|
| Node ≥ 22 | 全部 | `engines` 里写死 |
| `ffmpeg` + `ffprobe` 在 PATH | 编码、混音、拆参考片、量成片 | 本仓库实测 8.1.2；`brew install ffmpeg` |
| 一个 Chromium | 逐帧截屏 | 默认在 `~/Library/Caches/ms-playwright` 找 Chrome for Testing（**只认 macOS 布局**）；其它情况用 `CHROMIUM=/path/to/chrome` 指定 |
| dsh 0.2.x | 只有走插件路线时需要 | 命令、skill、面板由插件注册 |

### 装成 dsh 插件（推荐）

```bash
dsh plugin --profile web add @vlln/dsh-motion
```

装完之后：

- profile 的组件树里多一行 `@vlln/dsh-motion`（`dsh --profile web --dump-config | grep dsh-motion` 能看到）；
- 宿主进程的 `PATH` 里多出包内的 `bin/`——agent 直接能跑 `dsh-motion`，不需要知道包在哪；
- skill `dsh-motion` 注册进 agent 的 skill 列表（正文就是 `skills/dsh-motion/SKILL.md`）；
- 对话右侧栏出现「动画」页，也可以从右侧栏的引导行打开。

卸载用 `dsh plugin --profile web remove @vlln/dsh-motion`。片子住在工作区里，卸载不会删 `motion/`。

### 只当命令行用

```bash
npm install -g @vlln/dsh-motion     # 或者：npx --yes @vlln/dsh-motion <命令>
dsh-motion init --title "Hello"
```

这条路不需要 dsh，`init` / `check` / `render` 都能用；没有 skill、没有面板、没有批注。

## 五分钟：从空目录到一张静帧

```bash
mkdir hello && cd hello
dsh-motion init --title "Hello"
```

`init` 打印三行，落下 `motion/`：`timeline.json`（空时间轴）、`镜头表.md`（镜头表模板）、`kit/`（可抄可删的画法样例），外加一个空的 `scenes/` 和一个只忽略 `out/`、`.snapshots/` 的 `.gitignore`。**它不生成场景，也不给画风**——预置的画风会变成片子的画风，这是有意的。

然后写两个文件。`motion/scenes/hello.mjs`：

```js
export const id = 'hello';
export const slots = { title: { default: 'Hello world', role: 'title' } };
export const textUsage = { title: 1 };

let ctx;
export function mount(context) { ctx = context; }

export function render(t) {
  const g = ctx.canvas.getContext('2d');
  g.fillStyle = ctx.color('bg');
  g.fillRect(0, 0, ctx.width, ctx.height);
  g.fillStyle = ctx.color('accent');
  g.fillRect(ctx.width * (0.1 + 0.2 * t), ctx.height * 0.5, ctx.width * 0.1, ctx.height * 0.05);
  ctx.drawText('title', {
    x: ctx.width * 0.1, y: ctx.height * 0.3, w: ctx.width * 0.8, size: 96, weight: 700,
  });
}
```

`motion/timeline.json`（`init` 已经写好了尺寸、帧率和色板，这里补上幕与总时长）：

```json
{
  "width": 1920, "height": 1080, "fps": 60,
  "duration": 3,
  "palettes": {
    "main": { "bg": "#0b1426", "fg": "#f2f2f0", "accent": "#e0b040",
              "muted": "#8b98ad", "grid": "#243450", "panel": "#101c33" }
  },
  "scenes": [
    { "id": "hello", "name": "Hello", "start": 0, "length": 3, "palette": "main", "speed": 1 }
  ]
}
```

跑：

```bash
dsh-motion check                                   # ok · 1 scene(s) · 3s · 1920x1080 @ 60fps
dsh-motion render --stills 0,1.5 --out out/try.png # 两张静帧拼成一张对照表
dsh-motion render --video --out out/film.mp4       # 180 帧逐帧截屏，再编码
dsh-motion check --deliver                         # 六条事实：画面、声音、长度、空画面、幻灯片、蓝紫
```

这套 `init` → 写场景 → `check` → 静帧对照表的流程，在 0.2.0 的 npm 包上跑通过（macOS / node 22 / ffmpeg 8.1.2）。

一个前置提醒：`check` 默认会在最后多跑一次**性能基准**，而 npm 包里没有它要加载的 `preview.html`，所以从 npm 安装的 CLI 跑 `check` 时这一项会以 90 秒超时结束（不计入失败，见「[已知限制](#已知限制)」）。只想校验时间轴就用 `check --fast`。

## 命令

五个动词，一个阶段一个；没有第六个。

| 命令 | 做什么 |
|---|---|
| `dsh-motion init [--title "…"]` | 建 `motion/`：空时间轴 + `镜头表.md` + 一份 `kit/`。没有场景 |
| `dsh-motion reference <video> [--fps 2]` | 把参考片拆进 `motion/out/ref/`：帧、接触表、切点、brief |
| `dsh-motion reference --shots [--write]` | 把切点变成分镜（`--min 1.2` / `--max 8` 合并或拆分节拍）；`--write` 会按节拍重排 `timeline.scenes`，并为每一拍复制一份场景模块、渲一张静帧 |
| `dsh-motion check [--fast] [--strict] [--budget 33]` | 校验时间轴（含音频谱、字体、`audio.src` 是否存在），再量每幕单帧耗时与画面构成。`--fast` 只校验 |
| `dsh-motion check --perf [--budget 33] [--strict]` | 只量单帧耗时：默认预算 33ms（60fps 的一帧），超了退出码 1 |
| `dsh-motion check --deliver [--file x.mp4] [--allow-blue-purple "理由"]` | 先校验时间轴，再只对成片断言能证明的事实，不判好坏 |
| `dsh-motion render --stills 0,3,6 [--cols 3] [--out out/sheet.png]` | 静帧对照表（审片用，快） |
| `dsh-motion render --still 7.5 [--out out/frame.png]` | 一张全分辨率静帧 |
| `dsh-motion render --video [--out out/film.mp4] [--fps N]` | 整条片子，逐帧渲染，慢 |
| `dsh-motion render --export [--gif] [--out path] [--quality high\|medium\|low] [--height N]` | 交付用导出。默认 CRF 18；`--gif` 走 15fps、960 宽、调色板量化 |
| `dsh-motion version -m "用户看得见的变化"` | 把**当前 timeline + 全部场景源码**存成一个版本（`motion/.versions/`） |
| `dsh-motion version list` / `version use v3` | 列版本 / 切版本。切换只是导航：不新增版本；被替换的状态若不同，会先存进 `.versions/safety/` 作为撤销用 |

所有命令都读写**当前目录下的 `motion/`**。想指到别处就设 `DSH_MOTION_PROJECT`。

## 一部片子长什么样

```
motion/
  timeline.json      片子：尺寸、帧率、幕与时长、色板、文字、数值、素材、音频
  scenes/<id>.mjs    每一幕怎么画（导出 mount(ctx) / render(t)）；`_` 开头是共享模块，不算一幕
  assets/            图片与片段（在 timeline.assets 里声明）
  audio.json         音频提示单（可选）；make-audio 渲染成 assets/audio/bed.wav
  fonts/             自带的字体文件（可选，src 必须以 fonts/ 开头）
  out/               渲染产物、参考拆解（建片时已 gitignore）
  .snapshots/        每次写 timeline 前的副本，最多 20 份（已 gitignore）
  .versions/         版本：v1.json、v2.json…（建议进 git——这是片子的历史）
  镜头表.md          给人看的镜头表，程序不读它
```

## timeline.json

| 字段 | 类型 | 说明 |
|---|---|---|
| `width` / `height` | number | 画面尺寸，默认 1920×1080 |
| `fps` | number | 1–120；捕获帧率默认取它 |
| `duration` | number | 片长。**必须等于所有幕长之和**（首幕从 0 开始、首尾相接），校验会拒绝 |
| `scenes[]` | array | `{ id, name, start, length, palette, speed }`；`speed` 只允许 `0.5 / 1 / 2` |
| `palettes` | object | `main` 必含 `bg / fg / accent / muted / grid / panel` 六个 token，值必须是 `#rrggbb` |
| `overrides` | object | 颜色覆盖：`"<sceneId>"` 或 `"<sceneId>/<elementId>"` |
| `texts` | object | `{ "<sceneId>": { "<slot>": "文本" } }`；槽必须由场景声明，超 `maxChars` 拒绝 |
| `values` | object | 数值槽：`{ "<sceneId>": { "<key>": number } }`，面板可以改 |
| `params` | object | 参数（滑杆）的当前值：`{ "<sceneId>": { "<key>": number } }`，由面板写入 |
| `assets` | object | `{ "<name>": { "kind": "image" \| "clip", "src": "assets/…", "fps": 24 } }` |
| `audio` | object | `{ src, start?, targetLufs?, fit }`；`fit` 为 `trim`（放一遍，之后静音）或 `loop` |
| `beats` | object | `{ bpm, beats: [...] }`；由 `audio.json` 里的 `pad` / `score` 生成，不要手写 |
| `poster` | number | 面板打开时停在哪一帧（0..duration） |
| `seed` | number | 随机种子，缺省 `746337560` |

**铁律**：场景代码里不出现用户文案，也不出现 `#rrggbb`。两者只能来自数据——这样改字、改色不碰代码，改构图不碰数据。

### 声音：一张提示单

默认要有声音。`motion/audio.json` 写一张提示单，用 `node skills/dsh-motion/scripts/make-audio.mjs --write` 合成 `assets/audio/bed.wav` 并写回 `timeline.audio`：

```json
{
  "layers": [
    { "kind": "rain", "gain": 0.13 },
    { "kind": "pad", "bpm": 56 },
    { "kind": "page", "at": 34 }
  ]
}
```

- 不带 `at` 的层铺满全片（环境）；带 `at` 的层从那一秒开始——事件类只响一次，铺底类会一直铺到片尾；
- 可用的层：`rain` `roomtone` `wind` `pad` `score` `clock`（铺底）与 `page` `bell` `wheel` `click`（事件）。全部是合成的，没有采样与授权问题；
- 也支持自带一条轨：`node skills/dsh-motion/scripts/make-audio.mjs --use motion/assets/audio/track.wav`——文件必须在片子目录内，否则拒绝；
- **片长永远由画面决定**：混音时视频流直接 copy，音频按 `targetLufs`（默认 −14）做 loudnorm，AAC 192k / 48kHz，并补成整片长度。

## 场景模块

场景只导出 `mount` / `render` 和声明，没有工具面可以学，也没有 DOM 可依赖：

```js
export const id = 'hello';           // 必须与 timeline.scenes[].id、文件名一致
export const kind = 'canvas2d';      // 或 'webgl'，缺省 canvas2d
export const slots = { title: { default: '…', role: 'title', maxChars: 60 } };
export const textUsage = { title: 1 };          // 这个词每帧读几次，用于审计
export const params = { axisY: { default: 0, min: -1, max: 1, step: 0.01, label: '轴' } };
export const values = { speed: { default: 12, min: 0, max: 99, format: 'int' } };
export const assets = ['plate'];

export function mount(ctx) { /* 只调一次：预加载、预绘图、预计算都放这里 */ }
export function render(t) { /* 逐帧；t 是这一幕自己的秒表 */ }
export function overlay(g, ctx, t) { /* 可选：合成之后的 2D 层，给 WebGL 场景配排版 */ }
export function hitTest(x, y, t) { return null; }  // 可选：只读命中
```

`ctx` 能给什么：

| | |
|---|---|
| `ctx.canvas` / `ctx.width` / `ctx.height` / `ctx.dpr` | 本幕独占的画布与尺寸 |
| `ctx.color(token, elementId?)` | 调色板颜色；解析链为 元素覆盖 → 场景覆盖 → 场景色板 → `main` → **抛错**（不静默回退黑） |
| `ctx.text(slot)` / `ctx.drawText(slot, opts)` | 取文字槽 / 排版并绘制（断行、禁则、省略号由内核负责），同时登记为可编辑、可命中 |
| `ctx.param(key)` / `ctx.value(key)` / `ctx.drawValue(key, opts)` | 数值旋钮 / 数值槽，越界是错误不是钳制 |
| `ctx.rng(key)` | 可复现的随机（按 `seed` + 场景 id + key 派生） |
| `ctx.sceneStart` / `ctx.beats` / `ctx.bpm` | 本幕在整片中的起点、节拍网格 |
| `ctx.font(role)` / `ctx.still(draw, { scale })` | 字体栈 / 画一次存起来，之后每帧贴回 |
| `ctx.asset(name)` / `ctx.image(name, opts)` / `ctx.frame(name, at, opts)` | 素材 URL、图片、片段某时刻的帧 |

**确定性是硬规则**，因为整条片子是「逐帧重放 + 截屏」得到的：

| 禁止 | 替代 |
|---|---|
| `Math.random()` | `ctx.rng('名字')` |
| `Date.now()` / `performance.now()` | 只用 `t` |
| `requestAnimationFrame` | 宿主每帧调 `render(t)` |
| 渲染阶段发网络请求 | 放进 `mount` |
| 跨帧可变状态（闭包计数器） | 预计算到 `mount` |
| 不钉 sRGB 的 context | `getContext('2d', { colorSpace: 'srgb' })`；macOS 的 display-P3 会让读回值偏移 |

同一 `t` 渲两次必须逐像素一致，这条由门禁守着。段长变化时：没声明 `export const duration` 的场景**被裁剪**，声明了的**被变速**走完自己的自然时长；`speed` 再额外乘一次。

## 面板

装成插件后，面板是对话右侧的一栏（iframe，地址前缀是 `/plugins/dsh-motion/p/<会话>/`，所以时间轴、场景、素材、接口都落在同一个工作区目录里）。

![面板：版本、播放头、这一幕的分段](docs/demo/ui.gif)

- 播放 / 暂停、拖播放头、按场景分段跳转；点画面里的元素能看到它这一帧的几何（只读检视器）；
- 改**这一帧实际画出来的**文字与数值：聚焦某个字段，画面上会框出它对应的区域；拖动即时重绘，**失焦才落盘**；
- 参数滑杆：拖动改内存、松手才写文件；越界直接 400 拒绝并说明范围，文件与快照都不动；
- 回退一步：读 `.snapshots/` 里的上一版（每次写入前留一份，最多 20 份）；
- 版本下拉：面板跟的是**已发布的版本**，不是工作区。有未落版的修改时顶部显示「制作中」角标并只读——所以 agent 改完要跑一次 `dsh-motion version -m "…"`；
- 导出：从面板直接出片；
- 批注：在暂停帧上框一块、写一句、发送。话由插件外壳用这个会话的身份发进对话（作为人说的话），片子里只留那颗钉子（时刻、区域、元素、处理状态）。**没有会话就没有批注入口**：独立打开预览时批注只会显示「未发送」。

在命令行里看批注：`node scripts/comments.mjs list --json --crops` 会按锚点把每个线程所在的时刻、`rect` 外扩 20% 截一张图（`out/comments/<id>.png`）——图比坐标好读；改完用 `resolve --id <id>` 把钉子标成已处理。

## 它保证什么，不保证什么

`check --deliver` 只断言能证明的事实，六条红灯：

| 红灯 | 含义 |
|---|---|
| 没有画面 / 没有声音 | 文件里缺流 |
| 长度不对 | 片长与时间线差 0.5s 以上 |
| 空画面 | 中段有连续 0.3s 以上的近乎纯色（降到 96×54，亮度标准差 < 4） |
| 幻灯片 | 一半以上的相邻帧完全一样 |
| 蓝紫雾 | 蓝紫像素占全片 30% 以上（题材需要时用 `--allow-blue-purple "理由"`） |

构图好不好、像不像那个风格、值不值得看——**这三样不在这里判**。它们走并排图、审片四问和独立盲评（[`howto/交付.md`](howto/交付.md)、[`howto/审片.md`](howto/审片.md)）。

这套系统的实测边界写在 [`docs/cold-start.md`](docs/cold-start.md)：三个没参与制作的 agent 各自从 `init` 出发（上面那条《站台》就是其中一条），都交出了带音轨、`check --deliver` 全绿的 10.8–11.4s 片子，并且各自现写了一种印刷媒介；但盲评只有 37–55 分，全部被「签名特征」一票否决，而自评比盲评高 20–35 分。结论是：**它保证下限（不空、不静、长度对、能交付），不保证上限（像不像那个媒介）**；上限靠参考、审片和不信任自评。

性能上，仓库记录过一次实测（2026-10-19，M 系列 Mac / node 22 / ffmpeg 8.1.2）：1920×1080、60fps 的成片捕获约 **20ms/帧**（含 WebGL 与封装），单帧预算默认 33ms——超过就是播放会卡的信号。

## 已知限制

写在这里的是**确实做不到**的，不是待办：

- **只有一条音轨**，没有旁白轨与 ducking；`audio.json` 的层做不到「中途出现又消失」的铺底；
- **一个工作区一部片子**。幕由 agent 写，面板只读和改，不提供「再加一幕」；
- 同一个场景模块在一条时间轴里只能出现一次（`id` 唯一）；要重复同一套版式，得写变体模块；
- **改场景代码后面板不能热应用**（浏览器已缓存模块），只出现「场景代码已更新」提示条，需要刷新；
- 参数只支持数值：颜色走 `palettes`，文案走 `texts`，没有枚举 / 布尔 / 颜色参数；
- 数值槽只有一个数：没有数组 / 时间序列，复杂图表要自己在场景里组织数据；
- 片段素材先烘成一张雪碧图整段装进内存（`npm run assets`），长片段会吃内存，没有分块加载；片段自带音轨不入片；
- 参考摄取只读画面：不读音频（拍点靠 `audio.json` 生成的 `beats`）、不做字幕 / OCR；
- 回退只有整份快照，没有逐字段 diff；
- 字体默认只用系统族（`sans` / `display` / `serif` / `mono` 四个角色），自带字体要自己放进 `motion/fonts/`；
- WebGL 的确定性只在一台机器上验证过（Apple GPU + Chrome for Testing），换机器或驱动需要重跑；
- 面板入口依赖宿主 client 的 `slots` / `sidebarRight` / `sessions` / `conversation` 服务；更旧的 dsh 没有这些服务时面板不会出现，但内核、命令、skill 仍然可用；
- **npm 包里没有 `preview.html`**（它只存在于仓库里，用于 `npm run preview` 和性能基准）。因此从 npm 安装的 CLI 跑 `check` 时，性能那一项会以 90 秒超时结束——不计入失败，想跳过就用 `check --fast`。

## 平台与版本

- 开发与实测环境：macOS（Apple Silicon）、node 22、ffmpeg 8.1.2、Playwright 缓存的 Chrome for Testing。Chromium 的自动查找只认 macOS 的 Playwright 缓存布局，其它平台请给 `CHROMIUM`。
- 插件形式在 dsh `0.2.1-rc.1` 上验证过安装、启动与面板（见 [`docs/features.md`](docs/features.md) 的兼容性表）；本仓库的 `0.2.0` 包在 dsh `0.2.1-rc.5` 上验证过 `dsh plugin --profile web add`、组件树合成与 `dsh-motion` 可执行文件落盘。
- 插件机制面（服务名、槽名、manifest 字段）跨 rc 会改名。升级 dsh 后若启动日志出现 `entry did not activate`，先核对 `webServer`、`sidebarRight` 这些服务名。

## 仓库地图

| 目录 | 是什么 |
|---|---|
| `src/` | 引擎：时间轴、画布、确定性、缓存、出片、音频、字体、素材。**没有一行画画的代码** |
| `lib/` | 插件的 Node half 与面板的打包产物（`lib/app.js` 由 `src/preview.mjs` 打包而来） |
| `bin/` | `dsh-motion` 命令行 |
| `skills/dsh-motion/` | 给 agent 的说明书（唯一手册）与 `make-audio` / `find-asset` 脚本 |
| `kit/` | 可整删的代码样例：笔触、媒介、像素、透视、骨架、后期。**引擎从不 import 它** |
| `cards/` | 一个风格一张卡：管线、色板、签名特征、坑、短板、状态 |
| `howto/` | 做法规程：新风格六步、从参考里取、角色、镜头表、六种做坏、审片、交付 |
| `showcase/` | 面板空状态用的样板库（引擎不引用） |
| `scenes/` + `timeline.json` | 仓库自己那条示例片（不属于引擎） |
| `examples/` | 四个案例的证据：对照帧与并排图（片子本身已移出仓库） |
| `fixtures/` | 门禁用时间轴夹具 |
| `docs/` | 决策记录、内核契约、冷启动记录、限制、基准、门禁清单 |
| `scripts/` | 出片、抽帧、音频、参考拆解、门禁 |

## 在仓库里开发

```bash
npm install              # 装 playwright 与 esbuild
npm run build:client     # 重建面板（面板跑的是 lib/ 里的打包产物，改 src/preview.mjs 后必须重建）
npm run check            # 时间轴测试校验 + 28 组门禁
npm run preview          # 本地预览 → http://127.0.0.1:4173/preview.html
npm run video            # 渲仓库的示例片 → out/demo.mp4（含音频）
npm run audio            # 合成音床 + 节拍分析 + 波形图
npm run dev:host         # 起一个私有、持久的 dsh 宿主（home 默认 ~/.dsh-motion，DSH_MOTION_HOME 可改）
```

门禁是这份仓库唯一的回归保护：改引擎、改面板、改交付路径时都要跑。装成插件使用则用不到它们。清单见 [`docs/gates.md`](docs/gates.md)。

## 文档

| | |
|---|---|
| [`docs/kernel-contract.md`](docs/kernel-contract.md) | 引擎契约：场景能拿到什么、必须导出什么 |
| [`docs/decisions.md`](docs/decisions.md) | 决策记录：备选、代价、**什么会让我们改主意** |
| [`docs/limits.md`](docs/limits.md) | 已知限制 |
| [`docs/gates.md`](docs/gates.md) | 门禁清单与实测数字 |
| [`docs/features.md`](docs/features.md) | 每项功能的细节与兼容性表 |
| [`docs/cold-start.md`](docs/cold-start.md) · [`docs/benchmark.md`](docs/benchmark.md) | 冷启动记录与基准：系统自己交了什么、别人怎么评 |
| [`howto/`](howto/README.md) · [`cards/`](cards/INDEX.md) | 做法规程与风格卡片 |
| [`skills/dsh-motion/SKILL.md`](skills/dsh-motion/SKILL.md) | agent 的完整手册（也是人的参考手册） |

## 许可

MIT © vlln
