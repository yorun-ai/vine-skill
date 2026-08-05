# vine-skill

**English** | [简体中文](./README.zh-CN.md)

[License](./LICENSE) · [Changelog](./CHANGELOG.md) · [Contributing](./CONTRIBUTING.md)

[![Vine](https://img.shields.io/badge/Vine-v0.12.0-f97316)](https://github.com/yorun-ai/vine/tree/v0.12.0)
[![Go](https://img.shields.io/badge/Go-%3E%3D1.26.5-00ADD8?logo=go&logoColor=white)](./references/foundations.md)
[![skelc](https://img.shields.io/badge/skelc-v0.11.1-0097a7)](https://skel.yorun.ai/docs/)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](./LICENSE)

A version-aware Codex skill for developing, troubleshooting, reviewing, testing, upgrading, and deploying [Yorun Vine](https://github.com/yorun-ai/vine) services and browser frontends.

The reference baseline is Vine `v0.12.0`, skelc `v0.11.1`, and Go `1.26.5` or later. A target project's pinned versions always take precedence.

## Core behavior

- Determine the target project's pinned Go, Vine, and skelc versions before selecting APIs or commands.
- Require pinned `skelc check`, generation, and generated-diff review for contract, capability-registration, generator, or version changes.
- Preserve generated boundaries for implementation-only changes.
- Use Rpc/vRPC for structured synchronous APIs and reserve Web/HTTP for binary streams and static assets.
- Require browser frontends to use the project's vRPC client and skel-generated TypeScript services, specs, and types.
- Deliver browser applications through a generated Web capability and Portal WEBGW.
- Keep execution-scoped contexts, clients, DAOs, caches, and lockers inside their owning execution.
- Match validation evidence to standalone, linked, or separated runtime topology.
- Diagnose read-only unless the request authorizes implementation.

## Repository layout

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

`SKILL.md` contains the core workflow. Detailed and version-sensitive guidance lives under `references/` and is loaded only when relevant.

The files under `scripts/` are dependency-free Linux Bash and Windows Command Prompt launcher templates for generated browser-enabled Vine applications.

## Install

Clone the repository into the Codex skills directory:

```bash
git clone <repository-url> "${CODEX_HOME:-$HOME/.codex}/skills/vine-skill"
```

Invoke the installed skill explicitly as `$vine-skill`.

## Start a generated Vine browser app

The Skill includes dependency-free native launcher templates:

- [start_vine_app.sh](./scripts/start_vine_app.sh) for Linux Bash;
- [start_vine_app.bat](./scripts/start_vine_app.bat) for Windows Command Prompt.

When the Skill creates a browser-enabled project, it copies and adapts both launchers into that project's `scripts/` directory. On the first run use `--install`; later runs omit it, and `--check` performs preflight only.

The default local Standalone topology exposes the browser page at `http://127.0.0.1:7288/`: `/api` is handled by RPCGW and `/` by WEBGW. Vite remains a private upstream on `:5174`, while Dashboard uses the separate project-specific origin on `:7299`.

## Version safety

The skill does not install or upgrade Vine, skelc, Go, or runtime infrastructure automatically. It derives target versions from evidence such as `go.mod`, `go.sum`, CI configuration, generation scripts, generated-file markers, and already-installed tool output.

When the target revision differs from this repository's reference baseline, public source, GoDoc, tests, and release notes from the target revision take precedence. Source priorities are documented in [official-sources.md](./references/official-sources.md).

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for synchronization rules, reference-update requirements, validation expectations, and pull-request guidance.
