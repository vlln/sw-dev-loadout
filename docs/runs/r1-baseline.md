# 评审结论：job-queue PR（`drain(limit)` / `close()` 停机排空）

**结论：不能合并（request changes）。** 阻断项 6 条、建议 8 条。核心问题集中在 `src/queue.ts` 的 `close()`：它声称解决"停机丢任务"，但实现上没有等待排空、把排空上限硬编码成 100、并把任务异常静默吞掉——三处叠加后，这个 PR 既没有修掉原事故，又新增了静默丢任务与错误丢失的路径。两个新增测试都无法检测这些缺陷，CI 侧同时放宽了覆盖率阈值并删掉了 Node 20。

判定口径（只针对本 diff 与其声明）：优先级 > 正确性、生命周期、错误处理、必需行为 > 风格；"文档与代码一致"、"测试必须能在目标回归上失败"视为阻断级，不用覆盖率数字替代行为证据。

---

## 阻断项

### B1. `close()` 调用 `drain()` 却没有 `await`，关闭会在任务完成前就返回

- **位置**：`src/queue.ts` 的唯一 hunk（hunk 头见 patch.diff:L4），`close()` 体在 patch.diff:L30–L33，关键行 `this.drain(100)`（patch.diff:L32）。
- **问题**：`close()` 是 `async`，但 `this.drain(100)` 前面没有 `await`，返回的 Promise 被丢弃。`close()` 会在第一个 `await this.process(job)`（patch.diff:L20）处让出控制权后立即 resolve。
- **影响**：
  1. 调用方 `await queue.close()` 返回时，任务仍在执行；SIGTERM 处理器随后让进程退出，在途任务被截断——正是 PR 注释（patch.diff:L29）与 README 新增句（patch.diff:L63）声称修掉的"停机丢任务"。
  2. 因为这个浮动 Promise 无人 `await`，一旦 `drain` 自身抛错（例如 `logger.warn` 在 catch 里再抛，见 B3），就变成 unhandled rejection；`engines.node >= 20` 下默认 `--unhandled-rejections=throw`，会把进程直接崩掉——发生在关停路径上，代价最大。
- **依据**：本 diff 内部矛盾即可判定——README 写"`close()` 会等待所有在途任务结束再返回"，实现不等待。外部口径：`dsh-ci-test-reliability/SKILL.md:89`「Calling `abort()`, `close()`, or `kill()` without awaiting the owned completion signal is incomplete teardown.」；`dsh-code-review/SKILL.md:44` 要求对 teardown 追踪生命周期与取消。

### B2. `close()` 用硬编码 `drain(100)` 排空：超过 100 个待处理任务被静默丢弃

- **位置**：`src/queue.ts` patch.diff:L32（`this.drain(100)`）。
- **问题**：停机排空的上限被写成一个魔法数字 100，且没有任何返回值检查、日志或错误——`drain` 返回的 `handled` 被丢弃，队列里剩下的任务原地消失，无人知晓。
- **影响**：这是"丢任务"缺陷的回归，而不是修复。第 101 个及以后的任务在关停时既不执行也不报错，且调用方拿不到任何信号（`close()` 无返回值）。事故形态与 PR 想修的那两次完全同类，只是阈值从"全部"变成"100"——更难复现、更难归因。
- **依据**：PR 自述目标（patch.diff:L29）与本行直接冲突。`dsh-error-handling/SKILL.md:24`「Never claim success, rollback, or safe retry without guarantees.」；`dsh-error-handling/SKILL.md:18`「Logging alone is not recovery.」——这里连日志都没有。若确实需要上限，它必须成为显式契约（返回剩余数量、或抛错、或文档化），不能是常量。

### B3. `drain` 吞掉任务异常并计入 `handled`，调用方失去失败信号

- **位置**：`src/queue.ts` patch.diff:L19–L24（`try/catch` + `this.logger.warn('job failed', { error })` + `handled += 1`）。
- **问题**：两件事同时变了，且都没写进文档：(a) `drain()` 从"异常向上传播"变成"catch 后继续"，这是对外语义的破坏性变更；(b) `handled += 1` 在 catch 之后执行，失败任务被当成已处理计数，`return handled` 无法区分成功数。
- **影响**：
  1. 现有调用方若依赖 `drain()` 的 rejection 做失败处理（重试、告警、退出码），升级后全部变成静默成功——这类回归不会在类型检查或现有测试里暴露。
  2. 返回值语义模糊：`handled = 10` 既可能是 10 个成功，也可能是 10 个全失败。README 对 `drain(limit?)` 只写"处理队列中的任务"，未提及错误抑制。
  3. 只有 `logger.warn` 一行，没有聚合、没有 cause 保留、没有让状态所有者确认"继续是安全的"。
- **依据**：`dsh-code-review/SKILL.md:43`「Reject catch-and-continue unless the state owner establishes that subsequent work's preconditions hold, including after recovery.」；`dsh-error-handling/SKILL.md:12,18`（日志不是恢复；每个失败要有处理与上报的 owner）。**如果目标就是"单任务失败不中断队列"，那么它是合理需求，但必须同时提供失败计数/聚合结果，并在 README 与 PR 说明里明确这是行为变更**——现状是三者都缺。这是本 PR 中"可能是有意设计"的一项，我按阻断处理，理由是它对既有消费者的影响面完全未披露。

### B4. 两个新增测试都无法在它们声称的场景上失败

- **位置**：`src/queue.test.ts` 新 hunk（patch.diff:L38–L55）。
- **问题**：
  1. `处理失败的任务不会中断队列`（patch.diff:L42–46）：从未 `enqueue`，队列为空 → `while` 循环体一次都不执行 → `catch` 分支根本没被走到；断言是 `expect(handled).toBeDefined()`，对 `0` 也成立。这条测试只证明"空队列 `drain` 返回了某个数"。
  2. `close 会等待在途任务`（patch.diff:L48–53）：因为 async 函数体同步执行到第一个 `await`，`this.items.shift()`（L18）在 `close()` 返回前就已经把元素取走，`size` 立刻为 0——所以即使 B1 的缺失 `await` 存在，`expect(queue.size).toBe(0)` 依然通过。它测到的是"shift 发生了"，不是"等待结束了"。
- **影响**：两条测试给出的绿灯与声称的行为无关，恰好把 B1/B3 掩护过去；同时新增的 `catch` 分支很可能仍然是 0 覆盖（这也让 B5 的阈值下调显得可疑，见下）。测试还留下一个 20ms 后完成的浮动 Promise 与真实定时器，属于"用 sleep 当同步"。
- **依据**：`dsh-code-review/SKILL.md:53`「assertions fail on the intended regression and verify external state … rather than restating the implementation」；`deep-standard/evidence.md:90`（断言要盯权威来源，而非"方法存在/返回值有定义"这类几乎不会失败的静态存在性）；`deep-standard/evidence.md:82`「等完成信号，不等时间。禁止用 sleep/延时作为同步手段」。改法：手工排入会抛错的任务并断言 `handled`/失败计数与后续任务确实被执行；用 deferred promise（暴露 resolve 句柄）证明 `close()` 的 Promise 只有在任务完成后才 settle。

### B5. 覆盖率阈值 90 → 85，与本 PR 的性质无关且没有理由

- **位置**：`.github/workflows/ci.yml` patch.diff:L84（`--threshold 85`）。
- **问题**：门槛被下调 5 个百分点，diff 里没有任何说明；同一 PR 又恰好新增了未被测试覆盖的 `catch` 分支（B4）。CONTEXT 明确记载"覆盖率阈值 90%"。
- **影响**：这是把项目的质量门禁改弱以换取绿灯，而不是让代码满足门禁。一旦合并，后续所有 PR 都在 85 的基线上运行，且没有任何记录说明这是有意的策略变更；如果真实动因是新增代码测不到，那等于用降低全局门槛掩盖一个可修的测试缺陷。
- **依据**：`deep-standard/gates-as-code.md:84`「**禁止为了让 CI 变绿而放宽门禁。** 抬阈值、加豁免、跳过检查都必须是独立且有理由的决定并被记录。」；`deep-standard/evidence.md:57-63`（覆盖率是独立门禁，未覆盖的行通常说明测试缺陷或死代码，应先问"这行还需要存在吗"）。要求：从本 PR 移除，或拆成独立 PR 并用决策记录写明理由。

### B6. CI 矩阵删掉 Node 20，但 `engines.node` 仍声明 `>= 20`

- **位置**：`.github/workflows/ci.yml` patch.diff:L76（`node-version: [22]`）。
- **问题**：支持面收窄了，但包元数据（`engines.node >= 20`）没有同步，diff 里也没有理由。CONTEXT 记载 CI 原本跑 20 与 22 两个版本。
- **影响**：包对用户仍然宣称支持 Node 20，却不再有任何 20 上的证据；此外本 PR 改的是 async/关闭语义与 `unhandled-rejection` 相关路径，恰好是最吃运行时版本差异的部分。若确实要放弃 20，那是 breaking change：需要改 `engines`、进 CHANGELOG、按 major 发布，而不是顺手删一行矩阵。
- **依据**：`deep-standard/evidence.md:18,22`（平台矩阵层回答"在支持的每个运行时版本上都成立吗"，是 CI 独占职责）；`dsh-code-review/SKILL.md:37`（要求 CI 覆盖穷尽的矩阵，证据与声明必须一致）。与 B5 一样，这属于"门禁/矩阵被削弱的决定必须显式且被记录"。

---

## 建议

### S1. `handled` 的语义应在签名上固定下来
- **位置**：patch.diff:L13（`limit` 默认值）、L24（`handled += 1`）、L26（`return handled`）。
- **问题与影响**：成功与失败合并计数（见 B3）。建议返回 `{ processed, failed }`，或让 `handled` 只计成功、失败另行计数；否则调用方无论怎么用这个数字都会误解。
- **依据**：`dsh-error-handling/SKILL.md:10`（预期失败要有声明的类型与处理相关信息）。

### S2. `drain(limit)` 的边界没有测试，README 也没有真的描述它
- **位置**：patch.diff:L13、L17；README hunk L65–L67。
- **问题与影响**：README 原有 `drain(limit?)` 段落未更新以说明"失败不中断"和"返回值是计数"；`limit = 0`、负数、`Infinity`、恰好等于队列长度这四种边界一个新测试都没有。此外 `handled >= limit` 的检查放在 `while` 条件之后（L15–L17），语义正确但读起来像漏了边界，可移到循环首行。
- **依据**：`dsh-code-review/SKILL.md:34`（公开行为变更在同 diff 更新 README/JSDoc）。

### S3. `closed` 变成了只写不读的字段
- **位置**：patch.diff:L31（`this.closed = true`）。
- **问题与影响**：改造后 `closed` 被置位，但本 diff 内没有任何读取点；`drain` 也不检查它，于是"关闭后继续入队"的任务既不会被这次 close 处理，也不会被拒绝，静默悬空。要么让 `enqueue` 在 `closed` 时拒绝/抛错，要么删掉这个字段。
- **依据**：`dsh-find-simplifications` 式的必要性判断（`dsh-code-review/SKILL.md:45`：先找出每个状态字段背后的必需行为，再决定保留）。

### S4. `close()` 的重入与幂等未定义
- **位置**：patch.diff:L30–33。
- **问题与影响**：重复调用 `close()` 会再排空一次；并发 `close()` 会得到两个并行 `drain`，两个循环同时 `shift` 同一队列——顺序与计数都不确定。建议加一个 `closing?: Promise<void>` 做单飞（single-flight），或在入口短路。
- **依据**：`dsh-code-review/SKILL.md:44`（teardown 要检查重入前的所有权与竞态）。

### S5. `catch` 里调用 logger 可能二次抛错，且默认 logger 的适配未经确认
- **位置**：patch.diff:L22（`this.logger.warn('job failed', { error })`）、L7–L10（`logger ?? console`）。
- **问题与影响**：在 catch 块内部再次调用可能抛错的外部对象，会把"任务失败"升级为"`drain` 失败"（并在 B1 的情况下变成 unhandled rejection）。建议把日志调用本身包起来，或至少把它当作不可信的副作用边界。默认值 `console` 与 `Logger` 接口的形状匹配只能靠赋值兼容性侥幸成立（见"未确认"）。
- **依据**：`dsh-error-handling/SKILL.md:24`「Reporting failures need an independent fallback.」

### S6. 测试里的 20ms `setTimeout` 是延时同步
- **位置**：patch.diff:L49、L52。
- **问题与影响**：断言依赖"20ms 足够"，在 CI 并发下不稳定，且它并不能证明 overlap。建议改成显式的 deferred：测试持有 `resolve`，先断言 `close()` 尚未 settle，再 `resolve()`，await 完成后断言终态。
- **依据**：`deep-standard/evidence.md:82`；`dsh-ci-test-reliability/SKILL.md:75-80`（用显式就绪信号/屏障，禁止用固定 sleep 当同步）。

### S7. 注释在叙述评审历史，而不是契约
- **位置**：patch.diff:L29「以前 close() 只置位就返回，停机时经常丢任务；现在改成先排空再退出（见 PR 讨论里的两次事故）」。
- **问题与影响**：注释讲"改之前怎样、PR 讨论里有什么"，属于评审历史；读者从代码里拿不到契约（排空上限多少、失败怎么办、关闭后能否入队）。建议改为陈述契约，历史放 PR 描述或决策记录。
- **依据**：`dsh-code-review/SKILL.md:34`（注释陈述非显然的契约；实现叙述/评审历史应删除或链接到唯一归属）。

### S8. PR 混装了四种性质，CI 变更应拆出去
- **位置**：整个 diff：功能修复（`queue.ts`）+ 测试 + 文档（`README.md`）+ CI 配置（`ci.yml`）。
- **问题与影响**：`ci.yml` 的两处改动（B5、B6）既不属于这个 bug-fix，也无法被单独回滚或单独评审；混在一起时，评审者很容易把"门禁变弱"当成功能改动的附带噪声。即使最终接受，也应在 PR 描述里为每一项单独给理由。
- **依据**：`deep-standard/landing.md:5-9`（一个 PR 一种性质；独立改动拆开，否则无法单独回滚/归因）；`deep-standard/landing.md:56`（加了门禁豁免、跳过检查——在 PR 里为每一项给理由，不要混在功能改动里悄悄过去）。

---

## 未确认、因此没有写进结论的点

以下都需要 `src/queue.ts` 的其余部分或仓库配置才能判定，本 diff 不含这些信息，所以只作提示、不计入上文的阻断/建议：

1. **`enqueue()` 与 `size` 是否存在、语义是否如测试所用**：diff 只暴露 `items`/`closed`/`process`/`logger`。若 `size` 不是 getter 或不反映在途任务，B4 第 2 条的"误过"论证需要改口径（结论方向不变：`shift` 先于 `await`）。
2. **`Logger` 接口的确切形状**：无法确认 `console` 是否可赋值（`console.warn(...data: any[])` 通常可赋给 `(msg: string, meta?: X) => void`，但若 `Logger` 带 brand、必需字段或更窄的参数类型，这里就是编译错误——那会升级为阻断项）。
3. **是否有其它 `close()` 调用方依赖其立即返回**（如 SIGTERM 处理器、测试收尾钩子）：看不到调用点，无法评估 B1 的爆炸半径；只看本 diff 已足够定为阻断。
4. **阈值 85 是否就是为了掩盖新增的未覆盖 `catch`**：这是因果推测，我没有把它写成事实，只写了"可疑"与"必须给理由"。
5. **`engines.node` 与 `package.json` 是否已被其它 PR 改过**：若已放弃 Node 20，B6 只剩"本 diff 未同步说明"的问题；若没有，B6 成立。CONTEXT 的 `engines.node >= 20` 是我唯一的依据。
6. **`scripts/check-coverage.js` 的阈值是否另有默认值/配置文件**：若 90 还硬编码在别处，会出现两处阈值漂移；未读该脚本，故未单列。
7. **是否存在其它 workflow 仍跑 Node 20**：只看到 `ci.yml` 这一处矩阵改动，无法排除别的 job 覆盖了 20。

---

### 附：本次实际读取的材料

| 文件 | 行数 |
|---|---|
| `fixture/CONTEXT.md` | 6 |
| `fixture/patch.diff` | 85 |
| `deepseek-harness/.agents/skills/dsh-code-review/SKILL.md` | 62 |
| `deepseek-harness/.agents/skills/dsh-error-handling/SKILL.md` | 30 |
| `deepseek-harness/.agents/skills/dsh-ci-test-reliability/SKILL.md` | 131 |
| `devloop/skills/devloop/references/review-and-risk.md` | 115 |
| `deep-standard/references/landing.md` | 63 |
| `deep-standard/references/gates-as-code.md` | 94 |
| `deep-standard/references/evidence.md` | 91 |

合计 9 份、677 行（另浏览了三个参考目录的清单，未计入行数）。未读其余 20 份 dsh skill、其余 11 份 devloop references、其余 4 份 deep-standard references，也未读 `.sw-dev-experiment/with-set-review.md`。
