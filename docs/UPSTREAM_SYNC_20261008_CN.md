# 2026-10-08 上游同步与 Kiro Claude 5.5 优化

本文面向本发行的维护者，说明同步范围、兼容策略、优化收益和验证边界。

## 基线与同步方式

| 来源 | 固定提交 | 用途 |
| --- | --- | --- |
| 同步前本地 main | `a49a09f5d435338ff50ef1c4fa3070064d48ac6f` | 全部本地功能与原历史保留依据 |
| Wei-Shaw main | `3f1a2ea0a760730e3bc528105c00b4ee4f23e469` | 全量上游基线，版本 `0.2.14` |
| nianzs main | `304cbe65b6268fc8d31bdb316b51272180ac7e36` | Kiro Sonnet/Opus 5.5 实现参考 |
| 实际共同祖先 | `a3eb7ef302961cba716dc78b39b93b60c467db0e` | 计算差异和三方重放基准 |

Wei-Shaw 相对共同祖先有 173 个提交、408 个变更文件；原本地分支有 40 个提交、256 个变更文件。这些是两侧差异数量，不是最终同步提交的文件数。

在隔离 worktree 中以完整 Wei-Shaw 树为基底，三方重放本地累积差异，并将冲突适配到新的上游实现。Ent 与 Wire 从合并后的源文件重新生成，不用旧生成文件覆盖上游。GPT-6.1 Sol 的公开 ID 为 `gpt-6.1-sol`，其模型支持随完整 Wei-Shaw 基线带入，不添加猜测的 `gpt-6.1` ID。

原 main 通过保留历史的同步提交采纳已验证的树，同时保留 Wei-Shaw 及隔离检查点的祖先关系。备份分支为 `backup/pre-sync-20261008`。本次不执行远程推送、release tag、应用镜像部署或共享数据库升级。

## 保留清单

- OpenAI OAuth reused refresh token 的软失败处理、继续可调度和管理员 reauth 提示。
- 既有 provider 的 stale-stream watchdog：TTFT、chunk-gap warning/timeout，以及输出提交后禁止重复 failover。
- 两条 OpenAI Responses 请求构建路径的 Harmony neutralizer，以及 `invalid_prompt` 可观测性。
- Kiro 账号管理、OAuth/直接 API Key、推理区域、模型映射、credits/cooldown、文本/工具/流式能力。
- Kiro 的 Messages、Responses 和 Chat Completions 三协议入口；三个入口继续使用平台感知的稳定 session hash。
- Kiro 分组缓存配置、逐请求 cache read/create、实际账号/平台、管理员会话指纹及 bound/kept/switched 审计；历史软删除账号 hydration。
- Image 2.5、GPT-6 prompt-cache、Responses 数组型 `function_call_output` 和本地发行/Quay/低内存多架构构建支持。
- 新上游 TypeSafe、配额枚举及其他功能与 Kiro 并存，不用二选一的平台表覆盖它们。

Kiro cache 字段仍是 **Sub2API 本地计费模拟**，不是上游返回的真实命中数据，也不意味着减少上游输入 Token 或配额消耗。不会自动开启已有分组的模拟开关。

## 本次纳入的 5.5 优化

| 优化 | 实现与收益 | 边界 |
| --- | --- | --- |
| 模型目录及显式别名 | 新增 `claude-sonnet-5-5`、`claude-opus-5-5` 及两个 `-thinking` 变体；发送 Kiro dotted ID `claude-sonnet-5.5` / `claude-opus-5.5`。客户端可发现模型，避免空映射和错误上游 ID。 | 接受对应 dotted、大小写和首尾空格别名，不使用猜测未来版本的广泛规则。 |
| 模型感知输出上限 | 两款 5.5 的 `max_tokens=-1` 使用 128,000，显式较小值保留，超过 128,000 的值限制到上限。长输出不再被既有 32,000 默认上限截断。 | 旧模型的默认值和显式大值行为不变；源码上限不等于所有账号都能实际生成该长度，长输出也可能增加成本。 |
| Adaptive thinking 字段 | 仅在两款 5.5 的 adaptive 请求中发送 `additionalModelRequestFields`，包含 adaptive/summarized thinking 和 effort；thinking 别名使用 adaptive，保留用户 effort。 | 普通请求、旧模型和不支持的模型不附加这些字段，不静默改写其他请求参数。 |
| 三协议请求校验 | 接入 Wei-Shaw 的 Claude 5.5 共享校验，非法 forced tool/thinking 或采样配置按协议返回 HTTP 400；转换器提前拒绝时也写入客户端错误。 | Messages/Chat Completions 使用既有 `error.type`，Responses 使用既有 `error.code`，不重写共享错误格式。 |
| 映射查询归一化 | 精确映射优先，其次归一化 5.5 dotted 别名，避免目录支持而账号查询拒绝的落差。 | 不扩宽账号、分组或 API Key 的显式白名单；原 restricted mapping 仍限制原模型。 |

前端默认目录、账号映射预设和白名单来源继续集中在 `frontend/src/kiro/models.ts`。自定义账号映射不会自动添加新模型，需要管理员显式调整。

Sonnet 5.5 的共享 effort、协议兼容、请求校验和定价支持属于 Wei-Shaw 基线收益，不应记为 nianzs 独有优化。Kiro dotted/thinking ID 继续复用共享模型识别和定价逻辑。

## 暂未纳入的 nianzs 扩展

- 动态模型目录、分页和区域回退：可更准确反映账号报告的能力，但涉及账号 metadata、管理 API 和白名单语义；目录存在也不能证明调用可用。
- 900 KiB 历史裁剪：可降低超大请求失败概率，但可能丢失上下文、改变工具配对和 cache/session 行为，本次不覆盖现有稳定实现。
- 普通请求中的隐式 thinking 标签解析：可能减少标签混入正文，但也可能改变字面文本或暴露未请求 reasoning，暂保留既有契约。
- 广泛输出上限修改、forced tool 自动改写及结构化输出合成工具：影响其他模型或用户请求语义，不属于两款 5.5 的必要增量。

因此本次是完整同步 Wei-Shaw，并选择性纳入 nianzs 的直接相关 5.5 优化，不是合并 nianzs 的整个分支。

## 迁移兼容

原已发布的 Kiro `201`、`235`、`236`、`241` SQL 文件与同步前内容完全一致。不能改写这些文件，否则已执行实例会发生 checksum 不匹配。

Wei-Shaw 的 `241_add_typesafe_platform.sql` 和本地 `241_user_platform_quotas_add_kiro.sql` 都重建配额 CHECK。直接顺序执行会相互排除另一个平台，且仅新增后续迁移不能避免已有 Kiro 行在前一条迁移中失败。

采用仅针对这两个文件的执行层适配：先按原 SQL 计算和核对 checksum，再将配额 CHECK 的执行内容扩为 TypeSafe/Kiro 并集；其他约束、文件及已执行记录不变。新增 `242_user_platform_quotas_kiro_typesafe.sql` 作前向收敛。真实 PostgreSQL 测试覆盖新建、已有 Kiro、已有 TypeSafe 三条路径、幂等性及原 checksum 记录。

## 发行与验证结果

CI 与依赖跟随新上游；`.github` 只保留 CLA 仓库引用及 DockerHub 描述步骤 `continue-on-error` 两项既有差异。前端 package/lockfile 与 Wei-Shaw 相同，使用 pnpm 9 冻结安装。

保留完整上游安装器、systemd 单元和安装器测试；本发行默认下载地址改为 `tamakiramimy/sub2api-kiro`，可用 `SUB2API_RELEASE_REPO` / `SUB2API_DEPLOY_BASE_URL` 显式覆盖。只执行测试，不执行安装或发布。

最终已通过：

- Go 1.27.0：`go build ./...`、`go vet ./...`、全量 `go test -tags=unit ./...`。
- 全量 `go test -tags=integration ./...`；`CI=true`，使用隔离 PostgreSQL 18.1 / Redis 8.4 容器，不允许 Docker 不可用时静默跳过。
- golangci-lint 2.13.2：0 issues，与 CI v2.13 系列一致。
- govulncheck：0 可达漏洞、0 导入包漏洞；依赖模块层仍报告 11 项本代码未调用的漏洞，不将其描述为依赖完全无漏洞。
- Node 20.20.2：typecheck、全量 ESLint、339 个测试文件 / 2622 项 Vitest、生产构建。
- Python 3.12 / Linux Bash：10 项 release-matrix 测试；安装器 GitHub-token 测试、脚本语法检查通过，未运行真实发布命令。
- 四个已发布 Kiro 迁移内容不变；没有删除上游文件；补丁格式检查通过。

现有 TypeScript-eslint 版本支持范围提示和构建大 chunk warning 仍存在，不是本次新增失败。原共享应用、PostgreSQL 和 Redis 没有替换；只清理本次测试创建并核对 ID 的容器。

未使用真实获授权账号调用 GPT-6.1 Sol / Kiro 两款 5.5，故账号可用性、上游对 adaptive 字段及 128,000 输出的实际接受能力仍需实际调用验收。单元/协议/集成测试通过不能代替这些真实 provider 能力验证。

## 本地检查点

- `46d1268e5`：后端上游适配、本地 gateway 保留和 Kiro 5.5 核心支持。
- `b156a675c`：前端 Kiro 管理能力保留、模型目录及映射预设。
- `6b0bdc153`：迁移双平台兼容、别名 lookup 和三协议 5.5 请求错误处理。

这些检查点均为本地提交；最终同步提交的树由隔离 worktree 的已验证成果组成，并保留同步前 main 的祖先关系。