# 徽章：来源目录与写法

**何时读它**：要给 README 选徽章、现有徽章裂图或数值不更新、要做自定义或私有徽章时。

选不选、放几枚、样式统一的规则在 `SKILL.md`（视觉层）；这里只回答两个问题：**徽章从哪来**、**URL 怎么写**。

URL 格式与服务能力核对于 **2026-11-02**（抽样实测，见文末）；服务方改格式时以其文档为准重核。

## 一、shields.io：最大的来源

`img.shields.io` 一个服务覆盖大多数徽章。三种用法：

| 用法 | 何时用 | 形式 |
|---|---|---|
| 服务集成 | 值由某个平台托管（npm 版本、CI 状态、覆盖率） | `img.shields.io/<服务>/<参数>`，动态更新 |
| 静态 | 值不会变（许可、标准、约定） | `img.shields.io/badge/<左>-<右>-<颜色>` |
| endpoint | 值来自你自己的接口 | 你出一个 JSON，shields 来取（字段见「私有徽章」节） |

它的官方分类目录（选徽章时照这个找，不要凭记忆造）：

> Activity · Analysis · Build · Chat · Code Coverage · Dependencies · Downloads · Funding · Issue Tracking · License · Monitoring · Platform & Version Support · Rating · Size · Social · Test Results · Version

通用查询参数（所有 shields 徽章都能加）：

- `style`：`flat`（默认）/ `flat-square` / `plastic` / `for-the-badge` / `social`——**一份 README 全文只用一种**（SKILL.md 自检第 22 条）。
- `color` / `labelColor`：左右两边的颜色，接受 hex、rgb()、css 颜色名。
- `label`：改写左侧文字。
- `logo`：[simple-icons](https://simpleicons.org/) 的 slug；`logoColor` 配色；`logoSize=auto` 给宽图标。
- `cacheSeconds`：缓存秒数，低于服务默认值的会被忽略。
- 高负载时 shields 会限流；在意可用性可以把静态徽章下载进仓库（见「私有徽章」），或[自托管](https://github.com/badges/shields#self-hosting)。

## 二、GitHub 原生徽章

不走第三方，GitHub 自己出图。主要用于 CI 状态：

```markdown
[![CI](https://github.com/<owner>/<repo>/actions/workflows/<workflow>.yml/badge.svg)](https://github.com/<owner>/<repo>/actions/workflows/<workflow>.yml)
```

可加 `?branch=<分支>` 与 `&event=<事件>` 参数。其余的 GitHub 数据徽章（release、tag、stars、contributors、last commit、issues、license）走 shields 的 `github/*` 集成，不属于原生。

## 三、各生态注册表徽章

包一发布就该带的"版本从哪来"徽章。URL 格式（`<...>` 换成真实值）：

| 生态 | 徽章 | URL |
|---|---|---|
| npm | 版本 / 下载量 | `img.shields.io/npm/v/<pkg>` |
| PyPI | 版本 | `img.shields.io/pypi/v/<pkg>` |
| crates.io | 版本 | `img.shields.io/crates/v/<crate>` |
| docs.rs | 文档构建 | `docs.rs/<crate>/badge.svg`（301 到 shields 的 `docsrs`，等效） |
| pkg.go.dev | 文档 | `pkg.go.dev/badge/github.com/<owner>/<repo>.svg` |
| Go Report Card | 质量报告 | `goreportcard.com/badge/github.com/<owner>/<repo>` |
| Maven Central | 版本 | `img.shields.io/maven-central/v/<group>/<artifact>`（301 到 maven-metadata，等效） |
| NuGet | 版本 | `img.shields.io/nuget/v/<pkg>` |
| RubyGems | 版本 | `img.shields.io/gem/v/<gem>` |
| Packagist | 版本 | `img.shields.io/packagist/v/<vendor>/<pkg>` |
| pub.dev | 版本 | `img.shields.io/pub/v/<pkg>` |
| Hex.pm | 版本 | `img.shields.io/hexpm/v/<pkg>` |
| JSR | 版本 | `jsr.io/badges/@<scope>/<pkg>` |
| CocoaPods | 版本 | `img.shields.io/cocoapods/v/<pod>` |
| conda-forge | 版本 | `img.shields.io/conda/vn/conda-forge/<pkg>` |
| Docker Hub | 镜像大小 / pulls | `img.shields.io/docker/image-size/<user>/<image>` |
| Homebrew | 版本 | `img.shields.io/homebrew/v/<formula>` |

## 四、质量与安全服务

值由服务方维护，需要先在服务方接入项目，再拿徽章：

| 服务 | 它回答的问题 | 徽章位置 |
|---|---|---|
| Codecov / Coveralls | 测试覆盖率多少 | 服务后台给 shields URL |
| Code Climate / Codacy / SonarCloud / CodeFactor | 代码质量评级 | 同上 |
| Snyk / Socket | 依赖有没有已知漏洞 | 同上 |
| FOSSA | 许可证合规 | 同上 |
| OpenSSF Best Practices | 工程实践达标度 | `www.bestpractices.dev/projects/<id>/badge` |
| OpenSSF Scorecard | 供应链安全分 | `api.securityscorecards.dev/projects/github.com/<owner>/<repo>/badge` |

## 五、项目状态与流程约定

这类徽章回答"这项目还活着吗、按什么规矩玩"：

- **repostatus.org**：概念 / WIP / active / inactive / suspended / unsupported / moved 七档项目状态：`www.repostatus.org/badges/latest/<状态>.svg`。比"还在维护吗"的空话强。
- **Conventional Commits** / **Commitizen friendly**：提交信息约定（shields 静态徽章）。
- **SemVer**：版本号约定（shields 静态徽章）。
- **code style**：black、prettier 等格式化约定（shields 静态徽章）。
- **PRs Welcome** / **good first issue**：社区姿态。
- **standard-readme** / **Awesome 收录**：规范符合性。

## 六、社区与赞助

- 社区入口：Discord、Slack、Matrix、Discourse、GitHub Discussions——链到真实入口，别放没人看的频道徽章。
- 赞助：GitHub Sponsors、Open Collective、Patreon、Buy Me a Coffee、LiberaPay——都是 shields 集成。

## 七、自定义与私有徽章

值只存在于你自己的系统里（内部 CI、私有指标、公司内网）时的三条路：

1. **endpoint badge**：你出一个公网可达的 JSON，shields 定时来取。JSON 字段（schemaVersion 恒为 1）：`label`、`message`（必填）；`color`、`labelColor`、`isError`、`namedLogo`（simple-icons slug）、`logoSvg`（内联自定义 SVG）、`style`（可选）。

   ```markdown
   ![内部构建](https://img.shields.io/endpoint?url=<你的JSON地址>)
   ```

2. **dynamic 系列**：shields 直接从任意 JSON / TOML / XML / YAML 文件或正则匹配里取值：`img.shields.io/dynamic/json?url=<地址>&query=<JSONPath>&label=<左>`。适合值已经在某个配置文件里的情况。
3. **完全离线**：用 [badge-maker](https://www.npmjs.com/package/badge-maker) 在本地生成 SVG 收进仓库，或自托管 shields。内网项目的徽章不要指向公网 shields——读者打不开。

## 八、skill 与 dsh 生态徽章

**Agent Skills 规范**。符合规范的 skill 仓库可以给一枚静态徽章，链到规范页：

```markdown
[![Agent Skills](https://img.shields.io/badge/Agent_Skills-规范-4D6BFE?style=flat-square)](https://agentskills.io)
```

社区里出现过第三方"verified"徽章服务，未见稳定公认的，先不放。

**DSH 官方徽章**。DeepSeek Harness 自带一个内置 skill `dsh-badge`（npm 包 `@deepseek-ai/dsh-skill-badge`），它规定的官方徽章：

```markdown
[![](https://img.shields.io/badge/powered_by-dsh-4D6BFE?style=flat-square&logo=deepseek&logoColor=white)](https://github.com/deepseek-ai/deepseek-harness)
```

使用规则（官方 skill 原文，核对 2026-11-02，本机安装的 dsh 内）：放在文档或章节的**末尾**；不改颜色、logo、措辞与链接目标；渲染尺寸 121×20；飞书这类远程图片加载不可靠的系统里改用包内 PNG 资产，不要另生成一枚。徽章 URL 实测 200。

**DSH 插件兼容徽章**（社区约定，非官方）：宿主版本与平台用静态徽章写，例如：

```markdown
![DeepSeek Harness 0.2.x](https://img.shields.io/badge/DeepSeek%20Harness-0.2.x-111?style=flat-square)
```

来源：dsh-media-desk README（核对 2026-11-01）。版本号会过期，写成 `0.2.x` 这种区间比写死小版本活得久。

`logo=deepseek` 在 simple-icons 里存在（2026-11-02 实测：渲染出的 SVG 含图标；不存在的 slug 会静默掉图标，所以用前照这个方法验一次）。

## 九、常见但过不了规则的

下列徽章都存在、都有人用，但都过不了 SKILL.md 的"每枚回答一个真问题"：

- **forthebadge**（大字娱乐徽章）与 "Made with ❤" 系列：纯装饰；
- **visitor / hits 计数器**：数字对读者没有决策价值，且依赖第三方存活（2026-11-02 实测 api.visitorbadge.io 返回 403）；
- **与许可重复的 badge**（既写 LICENSE 文件又放许可徽章可以，但同一信息放两枚徽章不行）。

## 核对记录（2026-11-02）

实测 HTTP 状态：GitHub workflow badge 200；pkg.go.dev 200；docs.rs 301→shields（等效）；crates.io / PyPI / npm / NuGet / JSR / repostatus / bestpractices.dev 各 200；maven-central 301→maven-metadata（等效）；forthebadge 301 跳转存活但属娱乐向；api.visitorbadge.io 403。shields endpoint schema 与参数表照 [shields.io 官方文档](https://shields.io/badges/endpoint-badge) 抄写。
