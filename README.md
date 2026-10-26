# vine-skill

**English** | [简体中文](./README.zh-CN.md)

[License](./LICENSE) · [Changelog](./CHANGELOG.md) · [Contributing](./CONTRIBUTING.md)

[![Vine](https://img.shields.io/badge/Vine-v0.32.5-f97316)](https://github.com/yorun-ai/vine/tree/v0.32.5)
[![Go](https://img.shields.io/badge/Go-1.27.3-00ADD8?logo=go&logoColor=white)](./references/foundations.md)
[![skelc](https://img.shields.io/badge/skelc-v0.34.0-0097a7)](https://skel.yorun.ai/docs/)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](./LICENSE)

A version-aware Agent Skill for developing, troubleshooting, reviewing, testing, upgrading, and deploying [Yorun Vine](https://github.com/yorun-ai/vine) services with Codex, Claude Code, or OpenCode. New projects are delivered server-first; a frontend is added only after explicit confirmation.

The tested toolchain baseline is Vine `v0.32.5`, skelc `v0.34.0`, and Go `1.27.3`; the Go module floor is `1.27.0`. External clients are tested with vRPC Go `v0.15.1` and vRPC TS `0.10.0`. A target project's pinned versions always take precedence.

For migration details and the dependency selection rationale, see [baseline-upgrade.md](./references/baseline-upgrade.md).

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
- Use Rpc/vRPC for structured synchronous APIs and use Web/HTTP for binary streams, static assets, and required HTTP/streaming protocols.
- Create, use and release contexts, clients, DAOs, caches and lockers within their owning execution.
- Match validation evidence to standalone, linked, or separated runtime topology.
- Use read-only diagnosis and perform implementation when requested.
- Use only capabilities exposed by the active agent host and report missing evidence when a required tool or permission is unavailable.
- Create new projects server-first with the standard skeleton: contracts in root `skel/`, generated artifacts in `skeled/`, and Hub seed plus hand-written code under `src/server/`, with a requested browser frontend under `src/web/`.
- Build a requested frontend after server validation. For a server-only request, deliver the server and ask whether a frontend is needed afterward.

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

The default server skeleton uses `src/server/` as the server Go module root and `skeled/golang/` as the generated-code module. `src/server/go.mod` requires the generated module at `v0.0.0` and replaces it with `../../skeled/golang`. Each module owns its `go.mod` and `go.sum`; application code uses the responsibility directories shown above. Existing projects retain their established layout; layout migration follows an explicit request.

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

`scripts/` contains package validation tools and optional Linux Bash and Windows Command Prompt launchers. Copy only the launcher templates into projects with a confirmed frontend; package validation fixtures remain in this repository.

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

The skill derives target versions from `go.mod`, `go.sum`, CI configuration, generation scripts, generated-file markers, and installed tool output. Installing or upgrading Vine, skelc, Go, or runtime infrastructure requires a request covering that change.

When the target revision differs from this repository's reference baseline, public source, GoDoc, tests, and release notes from the target revision take precedence. Source priorities are documented in [official-sources.md](./references/official-sources.md).

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for synchronization rules, reference-update requirements, validation expectations, and pull-request guidance.
