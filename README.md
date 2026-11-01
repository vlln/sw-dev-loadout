# readme-skill

一份 Agent Skill：让 agent 写出读者优先的项目 README，而不是一份手册。

它要挡住三类具体失败：首屏不说这是什么、工程细节混进正文、长到没人读完。

## 安装

这个仓库还没有推到远端，先克隆或下载，然后从本地目录装：

```sh
skit install ./skills/readme-writing
skit check
```

`skit check` 应当报告 `readme-writing` 的链接已就位、无报错。

## 它管什么

`readme-writing` 回答一个问题：**这份 README 凭什么算合格。**

- 读者分层：一节只服务一类人，落在"深处"的内容走链接。
- 信息层级：章节顺序照读者做决定的顺序排。
- 视觉层：徽章、logo、截图、GIF、折叠区各自什么时候用、放哪。
- 体量：≤150 行、`##` ≤8，超标时按给定顺序处理。
- 自检：交付前逐条过的可判清单，每条都能用 `grep` 或一次复制粘贴取证。

两份 reference 按需读：[self-check.md](skills/readme-writing/references/self-check.md)（逐条怎么取证）、[visual-cookbook.md](skills/readme-writing/references/visual-cookbook.md)（视觉元素的具体写法）。

## 它为什么是这个形状

同一句任务跑了四次，只改方法：354 行（不给 skill）→ 372 行（skill 初版）→ 148 行（加体量判据）→ 140 行（再模仿一个高星项目的排版）。

结论是：**能执行的判据都生效了，没写下的判据完全没生效。** 初版给出了徽章、可复制安装块与首屏纪律，但没有一条体量规则，于是写出 372 行——比不给 skill 还长。把体量写成数字之后，行数才掉下来。

夹具、四个版本的原文与度量口径在 [docs/runs/](docs/runs/)，结论与它有多硬在 [docs/experiment.md](docs/experiment.md)。

## 文档

| 路径 | 是什么 |
|---|---|
| [docs/sources.md](docs/sources.md) | 调研来源：规范、同类 skill、宿主机械事实，每条都打开过 |
| [docs/experiment.md](docs/experiment.md) | 四次对照的度量、结论与限制 |
| [docs/runs/](docs/runs/) | 夹具、任务说明与保存下来的原文 |
| [docs/trigger-tests.md](docs/trigger-tests.md) | 触发判定：该中与近邻不该中 |
| [docs/self-audit.md](docs/self-audit.md) | 正文自查记录（自查，不是独立评审） |

## 许可

MIT
