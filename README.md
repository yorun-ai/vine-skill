# vine-skill

**English** | [简体中文](./README.zh-CN.md)

[License](./LICENSE) · [Changelog](./CHANGELOG.md) · [Contributing](./CONTRIBUTING.md)

[![Vine](https://img.shields.io/badge/Vine-v0.13.1-f97316)](https://github.com/yorun-ai/vine/tree/v0.13.1)
[![Go](https://img.shields.io/badge/Go-%3E%3D1.26.6-00ADD8?logo=go&logoColor=white)](./references/foundations.md)
[![skelc](https://img.shields.io/badge/skelc-v0.14.0-0097a7)](https://skel.yorun.ai/docs/)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](./LICENSE)

A version-aware Agent Skill for developing, troubleshooting, reviewing, testing, upgrading, and deploying [Yorun Vine](https://github.com/yorun-ai/vine) services with Codex, Claude Code, or OpenCode. New projects are delivered server-first; a frontend is added only after explicit confirmation.

The tested toolchain baseline is Vine `v0.13.1`, skelc `v0.14.0`, and Go `1.26.6` or later; Vine `v0.13.1` itself retains a Go `1.26.5` module floor. A target project's pinned versions always take precedence.

## Supported agent hosts

The package uses the open Agent Skills `SKILL.md` contract as its single source of workflow truth. Host-specific discovery, invocation, permissions, and optional presentation metadata stay outside the core workflow.

| Host | Project skill location | Explicit invocation |
| --- | --- | --- |
| Codex | `.agents/skills/vine-skill/` | `$vine-skill` |
| Claude Code | `.claude/skills/vine-skill/` | `/vine-skill` |
| OpenCode | `.opencode/skills/vine-skill/`, `.claude/skills/vine-skill/`, or `.agents/skills/vine-skill/` | Ask OpenCode to use `vine-skill` |

## Core behavior

- Determine the target project's pinned Go, Vine, and skelc versions before selecting APIs or commands.
- Require pinned `skelc check`, generation, and generated-diff review for contract, capability-registration, generator, or version changes.
- Preserve generated boundaries for implementation-only changes.
- Use Rpc/vRPC for structured synchronous APIs and reserve Web/HTTP for binary streams and static assets.
- Keep execution-scoped contexts, clients, DAOs, caches, and lockers inside their owning execution.
- Match validation evidence to standalone, linked, or separated runtime topology.
- Diagnose read-only unless the request authorizes implementation.
- Use only capabilities exposed by the active agent host and report missing evidence when a required tool or permission is unavailable.
- Create new projects server-first with the standard skeleton: contracts in root `skel/`, generated artifacts in `skeled/`, and Hub seed plus hand-written code under `src/server/`, with no `src/web/` until confirmed.
- After the server is complete, ask whether a frontend is needed and create `src/web/` only after confirmation.

## Generated server project layout

New projects created by the Skill initially use this server-only skeleton. Project and service names are replaced with the real names:

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
    └── web/          # created only after the user confirms the frontend
```

The default server skeleton uses `src/server/` as the server Go module root and `skeled/golang/` as the generated-code module. `src/server/go.mod` requires the generated module at `v0.0.0` and replaces it with `../../skeled/golang`. The skeleton never creates root-level `go.mod`, `go.sum`, `app/`, `cmd/`, `core/`, `impl/`, `repo/`, `seed/`, or `web/`, and contains no `internal/`. Existing projects retain their established layout unless migration is explicitly requested.

## Repository layout

```text
.
├── README.md
├── README.zh-CN.md
├── LICENSE
├── CHANGELOG.md
├── CONTRIBUTING.md
├── SKILL.md
├── agents/openai.yaml          # Codex/ChatGPT adapter metadata
├── references/
│   └── agent-hosts.md          # host compatibility details
├── scripts/
└── .github/workflows/ci.yml
```

`SKILL.md` contains the core workflow. Detailed and version-sensitive guidance lives under `references/` and is loaded only when relevant.

The files under `scripts/` are optional Linux Bash and Windows Command Prompt launcher templates copied only after the user confirms a frontend.

## Installation

Prerequisites: [Node.js](https://nodejs.org/) `>=22.20.0` for the pinned installer and at least one supported agent host.

Install `vine-skill` globally for Codex, Claude Code, and OpenCode:

```bash
npx --yes skills@1.5.23 add yorun-ai/vine-skill --global --agent codex claude-code opencode --skill vine-skill --yes
```

Remove `--global` for a project-scoped installation, or pass only one value after `--agent` to install for one host. Invoke the installed skill with the syntax supported by that host:

| Host | Example |
| --- | --- |
| Codex | `Use $vine-skill to create a Vine service.` |
| Claude Code | `/vine-skill Create a Vine service.` |
| OpenCode | `Use the vine-skill skill to create a Vine service.` |

See [agent-hosts.md](./references/agent-hosts.md) for official discovery paths, permissions, troubleshooting, and the compatibility smoke-test contract.

## Version safety

The skill does not install or upgrade Vine, skelc, Go, or runtime infrastructure automatically. It derives target versions from evidence such as `go.mod`, `go.sum`, CI configuration, generation scripts, generated-file markers, and already-installed tool output.

When the target revision differs from this repository's reference baseline, public source, GoDoc, tests, and release notes from the target revision take precedence. Source priorities are documented in [official-sources.md](./references/official-sources.md).

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for synchronization rules, reference-update requirements, validation expectations, and pull-request guidance.
