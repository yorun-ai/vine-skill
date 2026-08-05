# vine-skill

[English](./README.md) | **简体中文**

[许可证](./LICENSE) · [变更日志](./CHANGELOG.md) · [贡献指南](./CONTRIBUTING.md)

[![Vine](https://img.shields.io/badge/Vine-v0.12.0-f97316)](https://github.com/yorun-ai/vine/tree/v0.12.0)
[![Go](https://img.shields.io/badge/Go-%3E%3D1.26.5-00ADD8?logo=go&logoColor=white)](./references/foundations.md)
[![skelc](https://img.shields.io/badge/skelc-v0.11.1-0097a7)](https://skel.yorun.ai/docs/)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](./LICENSE)

这是一个具备版本意识的 Codex Skill，用于开发、排障、评审、测试、升级和部署 [Yorun Vine](https://github.com/yorun-ai/vine) 服务及其浏览器前端。

本仓库的参考基线为 Vine `v0.12.0`、skelc `v0.11.1` 和 Go `1.26.5` 或更高版本。目标项目自身固定的版本始终优先。

## 核心行为

- 选择 API 或命令前，先确定目标项目固定的 Go、Vine 与 skelc 版本。
- 契约、能力注册、生成器或版本发生变化时，必须运行固定版本的 `skelc check`、生成命令并审查生成 diff。
- 纯实现变更必须保持生成边界不变。
- 结构化同步 API 使用 Rpc/vRPC；Web/HTTP 只用于二进制流和静态资产。
- 浏览器前端必须使用项目 vRPC client 与 skel 生成的 TypeScript service、spec 和类型。
- 浏览器应用通过生成的 Web 能力与 Portal WEBGW 交付。
- execution-scoped context、client、DAO、cache 和 locker 不得逃逸出所属 execution。
- 根据 standalone、linked 或 separated 运行拓扑判断验证证据是否充分。
- 仅要求诊断或评审时保持只读，只有请求授权实现时才修改代码。

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

`scripts/` 包含面向生成项目的 Linux Bash 与 Windows Command Prompt 原生启动模板，不要求 Python。

## 安装

把仓库克隆到 Codex skills 目录：

```bash
git clone <repository-url> "${CODEX_HOME:-$HOME/.codex}/skills/vine-skill"
```

安装后通过 `$vine-skill` 显式调用。

## 启动生成的 Vine 浏览器应用

本 Skill 包含不依赖额外运行时的原生启动模板：

- Linux Bash 使用 [start_vine_app.sh](./scripts/start_vine_app.sh)；
- Windows Command Prompt 使用 [start_vine_app.bat](./scripts/start_vine_app.bat)。

Skill 创建带浏览器前端的项目时，会把两个启动器复制并适配到目标项目的 `scripts/` 目录。首次运行使用 `--install`，后续启动省略该参数，`--check` 只执行预检。

默认本地 Standalone 拓扑通过 `http://127.0.0.1:7288/` 提供浏览器页面：`/api` 由 RPCGW 处理，`/` 由 WEBGW 处理。Vite 作为私有上游保留在 `:5174`，Dashboard 使用独立的项目专属 `:7299` origin。

## 版本安全

本 Skill 不会自动安装或升级 Vine、skelc、Go 或运行时基础设施。它会从 `go.mod`、`go.sum`、CI 配置、生成脚本、生成文件标记和本机已有工具输出中提取目标版本证据。

当目标 revision 与本仓库参考基线不同时，以目标 revision 的公开源码、GoDoc、测试和 release notes 为准。来源优先级记录在 [official-sources.md](./references/official-sources.md)。

## 贡献

同步规则、参考资料更新要求、验证标准和 Pull Request 说明见 [CONTRIBUTING.md](./CONTRIBUTING.md)。
