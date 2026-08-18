# vine-skill

[English](./README.md) | **简体中文**

[许可证](./LICENSE) · [变更日志](./CHANGELOG.md) · [贡献指南](./CONTRIBUTING.md)

[![Vine](https://img.shields.io/badge/Vine-v0.13.1-f97316)](https://github.com/yorun-ai/vine/tree/v0.13.1)
[![Go](https://img.shields.io/badge/Go-%3E%3D1.26.5-00ADD8?logo=go&logoColor=white)](./references/foundations.md)
[![skelc](https://img.shields.io/badge/skelc-v0.14.0-0097a7)](https://skel.yorun.ai/docs/)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](./LICENSE)

这是一个具备版本意识的 Codex Skill，用于开发、排障、评审、测试、升级和部署 [Yorun Vine](https://github.com/yorun-ai/vine) 服务。新项目默认先交付服务端，只有明确确认后才追加前端。

本仓库的参考基线为 Vine `v0.13.1`、skelc `v0.14.0` 和 Go `1.26.5` 或更高版本。目标项目自身固定的版本始终优先。

## 核心行为

- 选择 API 或命令前，先确定目标项目固定的 Go、Vine 与 skelc 版本。
- 契约、能力注册、生成器或版本发生变化时，必须运行固定版本的 `skelc check`、生成命令并审查生成 diff。
- 纯实现变更必须保持生成边界不变。
- 结构化同步 API 使用 Rpc/vRPC；Web/HTTP 只用于二进制流和静态资产。
- execution-scoped context、client、DAO、cache 和 locker 不得逃逸出所属 execution。
- 根据 standalone、linked 或 separated 运行拓扑判断验证证据是否充分。
- 仅要求诊断或评审时保持只读，只有请求授权实现时才修改代码。
- 新项目默认按标准骨架先只交付服务端：契约在根级 `skel/`、生成物在 `skeled/`，Hub seed 与手写代码在 `src/server/`，确认前不创建 `src/web/`。
- 服务端完成后询问是否需要前端，只有确认需要时才追加 `src/web/`。

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

默认服务端骨架以 `src/server/` 为服务端 Go module 根目录，以 `skeled/golang/` 为生成代码 module。`src/server/go.mod` 以 `v0.0.0` require 该生成 module，并用 `replace` 指向 `../../skeled/golang`。骨架不创建根级 `go.mod`、`go.sum`、`app/`、`cmd/`、`core/`、`impl/`、`repo/`、`seed/`、`web/`，也不包含 `internal/`。现有项目除非明确要求迁移，否则保持既有布局。

## 仓库结构

```text
.
├── README.md
├── README.zh-CN.md
├── LICENSE
├── CHANGELOG.md
├── CONTRIBUTING.md
├── SKILL.md
├── agents/openai.yaml
├── references/
├── scripts/
└── .github/workflows/ci.yml
```

`SKILL.md` 保存核心工作流。详细且与版本相关的内容放在 `references/`，仅在任务相关时加载。

`scripts/` 包含可选的 Linux Bash 与 Windows Command Prompt 原生启动模板，只有用户确认需要前端后才复制到项目中。

## 安装

安装前请确保本机已有 [Node.js](https://nodejs.org/) 和 OpenAI Codex。

使用 npx 安装 Skill：

```bash
npx skills add yorun-ai/vine-skill
```

新建一个 Codex 任务，让 Codex 发现刚安装的 Skill，然后通过 `$vine-skill` 显式调用，例如：

```text
使用 $vine-skill 创建一个 Vine 服务。
```

## 版本安全

本 Skill 不会自动安装或升级 Vine、skelc、Go 或运行时基础设施。它会从 `go.mod`、`go.sum`、CI 配置、生成脚本、生成文件标记和本机已有工具输出中提取目标版本证据。

当目标 revision 与本仓库参考基线不同时，以目标 revision 的公开源码、GoDoc、测试和 release notes 为准。来源优先级记录在 [official-sources.md](./references/official-sources.md)。

## 贡献

同步规则、参考资料更新要求、验证标准和 Pull Request 说明见 [CONTRIBUTING.md](./CONTRIBUTING.md)。
