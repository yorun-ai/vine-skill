# vine-skill

[English](./README.md) | **简体中文**

[许可证](./LICENSE) · [变更日志](./CHANGELOG.md) · [贡献指南](./CONTRIBUTING.md)

[![Vine](https://img.shields.io/badge/Vine-v0.32.5-f97316)](https://github.com/yorun-ai/vine/tree/v0.32.5)
[![Go](https://img.shields.io/badge/Go-1.27.3-00ADD8?logo=go&logoColor=white)](./references/foundations.md)
[![skelc](https://img.shields.io/badge/skelc-v0.34.0-0097a7)](https://skel.yorun.ai/docs/)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](./LICENSE)

这是一个具备版本意识的 Agent Skill，可通过 Codex、Claude Code 或 OpenCode 开发、排障、评审、测试、升级和部署 [Yorun Vine](https://github.com/yorun-ai/vine) 服务。新项目默认先交付服务端，只有明确确认后才追加前端。

本仓库的实测工具链基线为 Vine `v0.32.5`、skelc `v0.34.0` 和 Go `1.27.3`；Go module 下限为 `1.27.0`。外部客户端使用 vRPC Go `v0.15.1` 和 vRPC TS `0.10.0` 验证。目标项目自身固定的版本始终优先。

迁移要点与实际依赖选择依据见 [baseline-upgrade.md](./references/baseline-upgrade.md)。

## 支持的 Agent Host

本包以开放 Agent Skills 的 `SKILL.md` 契约作为工作流唯一事实来源。Host 专属的发现方式、调用语法、权限和可选展示元数据不进入核心工作流。

| Host | 项目级 Skill 目录 | 显式调用 |
| --- | --- | --- |
| Codex | `.agents/skills/vine-skill/` | `$vine-skill` |
| Claude Code | `.claude/skills/vine-skill/` | `/vine-skill` |
| OpenCode | `.opencode/skills/vine-skill/`、`.claude/skills/vine-skill/` 或 `.agents/skills/vine-skill/` | 要求 OpenCode 使用 `vine-skill` |

## 核心行为

- 选择 API 或命令前，先确定目标项目固定的 Go、Vine 与 skelc 版本。
- 契约、能力注册、生成器或版本发生变化时，必须运行固定版本的 `skelc check`、生成命令并审查生成 diff。
- 纯实现变更必须保持生成边界不变。
- 结构化同步 API 使用 Rpc/vRPC；Web/HTTP 用于二进制流、静态资产以及确需 HTTP 或流式协议的场景。
- 在所属 execution 内创建、使用并释放 context、client、DAO、cache 和 locker。
- 根据 standalone、linked 或 separated 运行拓扑判断验证证据是否充分。
- 诊断和评审采用只读检查；按用户请求实施修改。
- 只使用当前 agent host 提供的能力；必需工具或权限不可用时，明确报告缺失证据。
- 新项目默认按标准骨架先交付服务端：契约在根级 `skel/`、生成物在 `skeled/`，Hub seed 与手写代码在 `src/server/`，用户要求的前端放在 `src/web/`。
- 服务端验证完成后，实现用户要求的前端。对于纯服务端请求，先交付服务端，再询问是否需要前端。

## 生成服务端项目结构

Skill 新建项目时先统一使用以下纯服务端骨架，项目名和服务名替换为真实名称：

```text
demo/
├── skel/
│   ├── domain.skel
│   └── greeting_service.skel
├── skeled/
│   ├── golang/
│   │   ├── go.mod
│   │   └── go.sum
│   └── typescript/
└── src/
    ├── server/
    │   ├── app/
    │   │   └── app.go
    │   ├── cmd/
    │   │   └── demo/
    │   │       └── main.go
    │   ├── core/
    │   ├── impl/
    │   ├── repo/
    │   ├── seed/
    │   │   └── hub.yaml
    │   ├── go.mod
    │   └── go.sum
    └── web/          # 用户确认前端后才创建
```

默认服务端骨架以 `src/server/` 为服务端 Go module 根目录，以 `skeled/golang/` 为生成代码 module。`src/server/go.mod` 以 `v0.0.0` require 该生成 module，并用 `replace` 指向 `../../skeled/golang`。各 module 分别维护自己的 `go.mod` 和 `go.sum`，应用代码按上图的职责目录组织。现有项目保持既有布局；布局迁移按用户的明确请求执行。

## 仓库结构

```text
.
├── README.md
├── README.zh-CN.md
├── LICENSE
├── CHANGELOG.md
├── CONTRIBUTING.md
├── SKILL.md
├── agents/openai.yaml          # Codex/ChatGPT 适配元数据
├── references/
│   └── agent-hosts.md          # host 兼容详情
├── scripts/
└── .github/workflows/ci.yml
```

`SKILL.md` 保存核心工作流。详细且与版本相关的内容放在 `references/`，仅在任务相关时加载。

`scripts/` 包含 Skill 包验证工具，以及可选的 Linux Bash 与 Windows Command Prompt 原生启动器。仅将启动模板复制到已确认需要前端的项目；Skill 包验证 fixture 保留在本仓库。

## 安装方式

安装前请确保固定版本安装器所需的 [Node.js](https://nodejs.org/) `>=22.20.0`，并至少安装了一个受支持的 agent host。

为 Codex、Claude Code 和 OpenCode 全局安装 `vine-skill`：

```bash
npx --yes skills@1.5.23 add yorun-ai/vine-skill --global --agent codex claude-code opencode --skill vine-skill --yes
```

项目级安装时移除 `--global`；只安装一个 host 时，仅在 `--agent` 后传一个值。按 host 支持的语法调用已安装 Skill：

| Host | 示例 |
| --- | --- |
| Codex | `使用 $vine-skill 创建一个 Vine 服务。` |
| Claude Code | `/vine-skill 创建一个 Vine 服务。` |
| OpenCode | `使用 vine-skill Skill 创建一个 Vine 服务。` |

官方发现目录、权限、故障排查和兼容性冒烟测试契约见 [agent-hosts.md](./references/agent-hosts.md)。

## 版本安全

本 Skill 从 `go.mod`、`go.sum`、CI 配置、生成脚本、生成文件标记和本机已有工具输出中提取目标版本证据。安装或升级 Vine、skelc、Go 或运行时基础设施，需要用户的请求涵盖相应变更。

当目标 revision 与本仓库参考基线不同时，以目标 revision 的公开源码、GoDoc、测试和 release notes 为准。来源优先级记录在 [official-sources.md](./references/official-sources.md)。

## 贡献

同步规则、参考资料更新要求、验证标准和 Pull Request 说明见 [CONTRIBUTING.md](./CONTRIBUTING.md)。
