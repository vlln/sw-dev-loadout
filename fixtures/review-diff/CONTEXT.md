# 背景

`job-queue` 是一个小的 Node/TypeScript 任务队列库，被若干服务用作后台任务执行器。
下面是某个 PR 的完整 diff（相对主干）。请只审这份 diff 与其上下文。

包信息：`engines.node >= 20`；CI 跑 Node 20 与 22 两个版本；覆盖率阈值 90%。
