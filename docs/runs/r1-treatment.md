# 评审结论：`job-queue`（close 排空 + drain 容错）

**结论：请求改动，不可合并。** 阻断 6 项，建议 5 项。

风险判断：中偏高。这支 diff 同时动了三处：公开 API 的失败语义（B3）、停机时的数据处置（B1/B2）、CI 的支持矩阵与覆盖率门禁（B5/B6）。前两者决定「停机时任务会不会静默消失」，后者决定「以后还能不能发现这类回归」——按 review-gate 的判级输入（数据影响 + 可逆性 + 验证覆盖 + CI 敏感面），不适合走「低风险自动集成」路径，需要作者修复后再由一位有权者合并，CI 两处改动的理由要单独给出。

行号说明：`queue.ts` 的 hunk 头与 hunk 体不一致（见「不确定」第 2 条），下文行号按「新文件从 12 起连续编号」估算，每条同时给出 hunk 与符号定位，以符号为准。

---

## 阻断项

### B1 `close()` 没有等待排空，README 的新契约当场为假

- **位置**：`src/queue.ts` close hunk（约 new-line 34–38，`async close()` 内 `this.drain(100)`）；`README.md` API 段（约 :11，本 PR 新增「`close()` 会等待所有在途任务结束再返回」）。
- **问题**：`close()` 调 `this.drain(100)`，既没 `await` 也没 `return`。`drain` 会同步执行到第一个 `await this.process(job)` 处挂起，而 `close()` 的 promise 在函数体结束时就 resolve。于是 `await queue.close()` 返回时，第一个 job 已经被 `shift()` 出队、`process()` 仍在 pending。同一 PR 里注释写「先排空再退出」、README 写「等待所有在途任务结束再返回」，两处都描述了一个没有实现的行为。
- **影响**：在 SIGTERM 停机钩子里 `await close()` 的服务，会在任务执行到一半时继续往下走并退出进程。相对改动前这**更隐蔽**：旧行为至少把任务留在 `items` 里，还有残留可观测；新行为把任务从队列摘走再丢失，队列看着是空的，没有任何地方能报告丢了什么。另外这是游离 promise，若 `drain` 因自定义 logger 抛错而 reject，无人接管。
- **依据**：failure-handling「捕获限制的是传播，不是变更」「关闭时：停止接收新工作、完成或显式放弃在途工作、报告放弃的部分」（references/state-guarantees.md 后台工作表）；README 已把它写成消费者契约，实现与契约不符即阻断（contract-prose「按位置覆盖 / README 消费者契约」）。
- **修复**：`await this.drain()`（上限问题见 B2）。注意修好之后会引入新风险：串行排空没有总时长预算，一条卡住的 job 会把「丢任务」换成「停机挂死到被 SIGKILL」。同一修复里给一个 deadline/`AbortSignal`，超时后**报告**未完成条数与 job id，而不是继续静默。

### B2 `close()` 的排空有硬上限 100，丢弃量不报告，`closed` 只写不读

- **位置**：`src/queue.ts` close hunk（约 :36–37，`this.closed = true` 与 `this.drain(100)`）。
- **问题**：
  1. 排空上限硬编码 100，超出的条目留在 `items` 里随 close 一起消失；`drain` 的返回值（处理条数）被直接丢弃，被放弃多少条、是哪些 job，没有任何出口。
  2. `closed` 从改动前就是只写字段，本次也没被读取：`enqueue()` 与 `drain()` 都不看它。close 之后仍可入队，另一个并发 `drain()` 仍会继续消费。
- **影响**：对一个以后台积压为正常工作状态的库，停机时最常见的恰恰是积压 >100。这条路径的效果是「大部分关了，但宣称全关了」。`closed` 不被读取说明「排空后退出」的状态机根本没建：close 与并发 drain 会同时 `shift()` 同一数组，各自维护独立的 `handled` 预算，互不约束，也无法回答「close 返回后队列应处于什么状态」。
- **依据**：state-guarantees.md「关闭时」行；change-evidence 把「并发、生命周期、拆卸」单列为需要专项证据的表面；归因为**本次加剧**——`closed` 的只写不读是既有问题，但本 PR 把「排空后退出」提升为公开语义，使这个死字段从无害变成契约缺口。
- **修复**：close 语义二选一并写进 README：要么排空到空；要么保留上限，但把放弃的条数与 job id 通过返回值或异常**报告**出来。`enqueue` 在 closed 之后必须有确定行为（拒绝，或入队等 reopen），不能是「还能进但永远不会被处理」。

### B3 `drain()` 吞掉失败、把失败计入 `handled`，且失败没有终态

- **位置**：`src/queue.ts` drain hunk（约 :24–31，`try/catch`、`this.logger.warn('job failed', { error })`、`handled += 1`、`return handled`）。
- **问题**：`drain()` 从「首个失败即 reject」改成「catch → 记一条日志 → 继续」，但三样东西都没有：失败的类型化/结构化、失败后的状态保证、可观察的终态。此外：
  - `handled += 1` 在 catch 之后执行，失败也计入，`drain()` 返回的数字无法区分「成功处理」与「失败丢弃」；
  - 失败的任务在进入 try 之前已被 `shift()` 出队，既没有重试，也没有死信或失败标记，失败本身只剩一行日志。
- **影响**：这是公开 API 的语义反转，而 README/变更说明里一个字都没有（README 只加了 close 那一句）。现有 `await drain()` 并靠 reject 做告警、重试、标记任务的调用方，现在全部变成「永远成功」。对队列库来说等于静默丢任务：进程不知道、监控不知道、调用方也不知道，且默认 logger 是 `console`（无 job id，见 S5）。
- **依据**：failure-handling「不要把未知失败包装成默认值返回，是比崩溃更贵的缺陷」「记一条日志然后继续——日志不是恢复」；state-guarantees.md「部分效果：已生效部分可枚举可识别、调用者知道哪些做了哪些没做」「毒消息要有终态：进死信、标记失败或降级，不能无限重投」「失败可观察：持续失败要有持续可见的信号」。
- **修复**：让失败可被调用方拿到——返回 `{ handled, failed: Job[] }`、提供 `onError` 钩子、或维持 reject 语义并在 README 声明；`limit` 计「尝试数」还是「成功数」必须选一个写清。日志只能作为辅助通道，不能是唯一通道。

### B4 两个新增测试在任何实现下都会通过，等于没有证据

- **位置**：`src/queue.test.ts` 新增两例（约 :24–28 与 :30–35）。
- **问题**：
  - **(a) `处理失败的任务不会中断队列`**：整个用例从未 `enqueue`。空队列上 `drain()` 一个 job 都遇不到，`handled` 恒为 0，`expect(handled).toBeDefined()` 对一个 number 恒真。把 try/catch 整段删掉、甚至让 `drain` 在首个失败时 reject，这个用例照样通过。
  - **(b) `close 会等待在途任务`**：断言的是 `queue.size === 0`，而 `size` 在 `await this.process(job)` **之前**（`shift()` 处）就已经是 0。也就是说，恰恰因为 close 没等，这条断言才通过；改成 `await this.drain()` 它也通过。它不区分「等了」与「没等」，也不检查那条 20ms 的 job 是否真的跑完。
- **影响**：本 PR 唯一的新证据在任一实现下都通过，而它恰好给 B1/B2/B3 盖了「已被测试覆盖」的章。另外 (b) 用 20ms `setTimeout` 当时序手段，测试结束时该 job 仍是游离 promise / 活定时器，拆卸没有到静止。
- **依据**：change-evidence「如果一条证据在任何实现下都会通过，它不是证据」「等完成信号，不等时间」「拆卸到静止」；failure-handling「验证的是结果状态，不是『异常被捕获了』」「断言『没有发生』时，同时断言一件本该发生的事确实发生了」。
- **修复**（同时作为 B1–B3 的关闭证据）：
  - (a) 入队 3 条、第 2 条抛错，断言第 1、3 条确实被执行（用完成序列/标记，不用时长），并按 B3 定下的契约断言失败可被观察；
  - (b) 让 job 在结束时置完成标记，`await close()` 之后断言标记为真；
  - 补一条 >100 条积压的用例，断言放弃量可观察；补一条 close 后 `enqueue` 行为的用例。

### B5 覆盖率门禁 90 → 85，且和功能改动挤在同一个 PR

- **位置**：`.github/workflows/ci.yml`（约 :23，`node scripts/check-coverage.js --threshold 90` → `85`）。
- **问题**：在**引入新代码的同一个 PR 里**把阈值降 5 个点，没有任何理由记录，也没有独立决定。而 B3 新增的 catch 分支正是没有测试走过的代码。
- **影响**：这道门禁本来是唯一能机械指出「新分支没人走过」的检查；降阈值把「补测试」替换成「调低标准」，同时把这个信号关掉。这类改动一旦合入，下一个 PR 不会有人再调回去，仓库能守住的下限永久下移。
- **依据**：engineering-standards/references/gates.md「为了让 CI 变绿而放宽门禁……必须是独立且有理由的决定并被记录」；change-evidence「不得用……降低阈值、或缩小被测范围来掩盖一个未被覆盖的受影响文件」。
- **修复**：恢复 90，补 B4 的用例。若确实有不可测部分需要豁免，单列 PR 说明是哪部分、由谁在评审环节核对，不要用全局阈值兜。

### B6 CI 删掉 Node 20，与 `engines.node >= 20` 冲突且与本 PR 无关

- **位置**：`.github/workflows/ci.yml`（约 :9，`node-version: [20, 22]` → `[22]`）。
- **问题**：包仍声明 `engines.node >= 20`，本 PR 却把 20 从矩阵中移除，既没有改 `engines`，也没有在 README/CHANGELOG 或 PR 描述里声明放弃该版本。这件事与「close 排空」毫无关系。
- **影响**：`engines` 继续承诺 Node 20 可用，但从此没有任何证据。而本 PR 改的正是生命周期与并发代码，Node 20 与 22 在定时器、`AbortSignal` 等能力上存在差异——下架 20 之后，「20 上还有没有这个问题」将永远无法回答。这是一次静默的兼容面收缩。
- **依据**：review-gate/references/risk-routing.md——CI 属默认敏感面，至少需要人类决定；同一文件里的功能改动不构成缩小支持矩阵的理由；contract-prose「不写死/不静默改变兼容承诺」。
- **修复**：恢复 `[20, 22]`。如确要放弃 20，另开 PR，同时改 `engines`、在 README/CHANGELOG 声明支持窗口，并给出迁移说明。

---

## 建议

### S1 默认 logger 是 `console`，库无条件向宿主 stdout 写日志

- **位置**：`src/queue.ts` 约 :14–16（`private logger: Logger` / `this.logger = logger ?? console`）。
- **问题**：字段建议 `private readonly logger`。更实质的是默认值：所有现有调用方从升级那一刻开始，会在自己的 stdout 里收到本库的 warn，没有级别、前缀、开关，只能靠注入 logger 来关掉。
- **影响**：一个小库默认污染宿主日志，通常被下游当作 bug 报回来；多实例场景下也无法归因到具体队列。
- **依据**：contract-prose「按位置覆盖 / 诊断信息」；failure-handling「诊断细节走日志或诊断通道」意味着日志通道必须可指定、可关闭。
- **建议**：默认 noop（或要求显式注入），并在 README 写明日志去向。同时核对 `Logger` 契约：`this.logger.warn('job failed', { error })` 的 meta 形状与 `console` 是否结构上满足 `Logger`（本 diff 看不到定义，见「不确定」第 1 条）。

### S2 `limit` 参数没有校验，负值/NaN 会静默退化成无限

- **位置**：`src/queue.ts` 约 :19（`limit = Number.POSITIVE_INFINITY`）与 :22（`if (handled >= limit) break`）。
- **问题**：`limit` 为负数或 `NaN` 时 `handled >= limit` 恒假，等于没传 limit；`0` 立即返回 0；非整数没有定义。这些都不报错。
- **影响**：调用方传错单位（例如把毫秒当条数）不会得到任何信号，而是拿到「全排空」的行为，与关闭路径的积压问题直接相关。
- **依据**：contract-prose「公开 API 注释必须写到返回区分、异常与拒绝」；failure-handling「预期失败要有声明的类型」。
- **建议**：`if (!Number.isInteger(limit) || limit < 0) throw new RangeError(...)`，README 写明计数口径（尝试数 or 成功数）。

### S3 返回类型从 `Promise<void>` 改成 `Promise<number>` 是公开 API 变更，未声明

- **位置**：`src/queue.ts` 约 :19。
- **问题**：原调用方拿到的是 `void`，现在拿到一个语义未定义的计数（且按 B3 还混入失败）。README 的 `drain(limit?)` 一节没提返回值。
- **影响**：读者会把它当「成功处理条数」用，而实现给的是「尝试条数」，这类误用不会报错，只会让上一层的指标悄悄偏。
- **依据**：contract-prose「公开 API 注释 / README 必须写到返回区分」；能力变更需要 semver/changelog 处置。
- **建议**：等 B3 定了口径后再写进 README 与 changelog；若改成返回结构体，趁早改，别让数字口径先被消费方依赖。

### S4 注释是变更叙事 + 指向 PR 讨论的不可解析引用

- **位置**：`src/queue.ts` 约 :34（`// 以前 close() 只置位就返回，停机时经常丢任务；现在改成先排空再退出（见 PR 讨论里的两次事故）。`）。
- **问题**：这是「以前/现在」的变更故事，外加一个站在当前版本无法解析的引用（PR 讨论）。而且它描述的「先排空再退出」当场被实现证伪。
- **影响**：半年后读代码的人会把这句话当成契约，而它既不准确，也无法验证；真正该写的东西（close 何时返回、返回后队列处于什么状态、上限与报告方式）一个字都没有。
- **依据**：contract-prose「只写当前态」「删除变更故事与版本印花」「泄漏：已失效的设计会话引用」。
- **建议**：删掉历史，改写成契约句（时序、返回条件、失败/上限的报告方式）；事故背景放进提交信息或事故复盘，需要时用可解析的相对路径链接过去。

### S5 失败日志不含 job 身份，排障无法定位

- **位置**：`src/queue.ts` 约 :27（`this.logger.warn('job failed', { error })`）。
- **问题**：没有 job id、序号或队列标识，只有 error。
- **影响**：运维看到一行 warn 无法知道是哪条任务失败、更无法重放；调用方也无法据此去补处理。这条日志是 B3 里失败的唯一出口，因此它缺什么，就等于失败信息缺什么。
- **依据**：contract-prose「诊断信息必须写到失败的主体或路径、违反的规则、非显然的修正方式」；state-guarantees.md「调用者需要知道的部分要出现在错误类型里，而不是只出现在日志里」。
- **建议**：`{ jobId: job.id, error }`；并按 B3 让失败可从返回值/异常观察，日志降为辅助。

---

## 不确定、因此没写进结论的点

1. **看不到仓库，只有 diff。** `Logger` 的类型定义、`Job` 是否必有 `id`、`queue.size` 与 `enqueue` 的既有契约（close 后是否允许入队）、仓库是否已启用 `no-floating-promises` 一类 lint——都无法验证。因此 B1 只按「未 await」这一事实判定，没有断言「lint 本该拦住 / 类型本该报错」。
2. **`patch.diff` 的 hunk 头与 hunk 体不一致。** `src/queue.ts` 只有一条 `@@ -12,6 +12,7 @@`，但该 hunk 体实际是 16 行旧 / 29 行新，drain 与 close 两处改动都没有各自的 `@@` 头。这更像 fixture 把多个 hunk 压平了（或头部计数未更新）；真实 `git apply` 会以 corrupt patch 拒绝。我把它当格式问题，未列为代码缺陷，但所有 `queue.ts` 行号都是估算值。
3. **失败该不该继续传播，涉及产品取向。** PR 声称有两次事故背景，「一条失败不能拖垮整批」可能是真实诉求。所以我只判定「当前实现既无终态也无可观察报告，且文档未声明」，没有判定「必须改回 reject」——这需要作者给出取向，再定契约。
4. **20ms 延时在当前 CI 上是否真的抖动**，我没有运行过，无法确认它现在是否偶发失败，只按「用时间当同步手段」的规则并入 B4。
5. **`processed` 与 `handled` 的命名、`Number.POSITIVE_INFINITY` 与 `Infinity` 的写法**属风格偏好，按 review-gate/findings.md 的丢弃规则未列入。

---

## 本次读了什么

- **skill**：`review-gate`、`failure-handling`、`contract-prose`、`change-evidence`、`engineering-standards`。
- **references**：`review-gate/references/findings.md`、`review-gate/references/risk-routing.md`、`failure-handling/references/state-guarantees.md`、`engineering-standards/references/gates.md`。
- **未读**：`delivery-loop`、`change-landing`（本任务是单点评审，无跨会话交付或落地动作）；`contract-prose/references/{coverage,leakage}.md`、`change-evidence/references/{determinism,performance,surface-map}.md`、`engineering-standards` 其余 references（缺口已由各自的 SKILL 正文覆盖，且本 diff 不涉及性能证明与文档门禁建设）。
