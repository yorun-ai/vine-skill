# Changelog

All notable changes to the vine-skill package are documented in this file.

The project follows [Semantic Versioning](https://semver.org/).

## [0.4.0] - 2026-10-27

### Added

- Added an Event → Task receipt workflow to the Greeting example, including generated Emitters and Launchers, Listener and Runner registration, SQLite deduplication, result lookup, and tests for Event redelivery and Task retries.
- Added an Orders and Catalog example with public contract generation and imports, generated Rpc calls in both directions, and an Event round trip between two Apps in one Standalone process.
- Added installation and invocation guidance for Cursor, GitHub Copilot (CLI / VS Code), Gemini CLI, DSH, Cline and Qwen Code, bringing the documented agent-host matrix to nine.

### Changed

- Clarified domain and App responsibilities, public contract ownership, and the relationship between schema imports and runtime call direction.
- Extended CI to validate asynchronous receipts through generated TypeScript clients, cross-domain integration with race checks, and complete Skill installation for eight installer-supported hosts plus a native DSH directory copy.
- Documented host discovery sources, installation paths and validation coverage, including pending client-session and user-profile installation tests.

## [0.3.0] - 2026-10-26

### Added

- Added a baseline-upgrade guide covering contract changes, selected Go dependencies, and staged migration of persistent Hub databases.
- Added disposable validation for the published dependency graph, generated Go and TypeScript API clients, real Portal authentication and CRUD, direct Hub Admin calls, and executable Web proxy/assets examples.
- Added a shared baseline manifest for CI toolchain selection and repository maintenance guidance.

### Changed

- Updated the tested baseline to Vine `v0.32.5`, skelc `v0.34.0`, and Go `1.27.3`, with a Go module floor of `1.27.0`. External clients use vRPC Go `v0.15.1` and vRPC TypeScript `0.10.0`.
- Migrated the Greeting example to API service contracts, explicit authentication modes, current Actor audiences, standard-library UUIDs, and current Standalone options and Hub seed fields.
- Refreshed configuration, lifecycle, runtime cloning, Redis lock release, Hub persistence, Dashboard access, Web serving, timeout, and deployment guidance against released sources.
- Condensed the core Skill and expressed guidance as direct responsibilities, actions, and conditions across references, host metadata, launcher help, and both repository READMEs.
- Integrated the published-baseline runtime checks into CI alongside portable skill installation and Linux/Windows launcher preflight.

## [0.2.2] - 2026-09-02

### Changed

- Updated the tested baseline to Vine `v0.13.2`, skelc `v0.14.0`, and Go `1.26.6`.
- Documented Vine `v0.13.2` Rpc value-isolation semantics, Portal public-entry streaming guardrails, and refreshed the example generation pins and official source revision.

## [0.2.1] - 2026-08-20

### Added

- Added a host-compatibility reference and a pinned multi-host install path for Codex, Claude Code, and OpenCode

### Changed

- Updated the tested baseline to Vine `v0.13.1`, skelc `v0.14.0`, and Go `1.26.6` while retaining Vine's released Go `1.26.5` module floor
- Documented the Vine `v0.13.0` external-NATS stream-provisioning boundary, including required `VINE_EVENTS`/`VINE_TASKS` retention and deployment-owned storage
- Documented Vine `v0.13.1` and skelc `v0.12.0+` typed in-process Rpc cloning, the compatibility fallback for older generated specs, and the distinction between value-isolation and wire tests
- Refreshed the official Vine, skelc, and website source revisions and updated example/CI pins
- Replaced the manual Codex Skill clone instructions with `npx skills add yorun-ai/vine-skill`
- Made the core `SKILL.md` agent-host neutral while retaining `agents/openai.yaml` as the optional Codex and ChatGPT presentation adapter
- Extended package validation to enforce the portable core boundary and the three-host discovery contract

## [0.2.0] - 2026-08-07

### Added

- Added a verified `example-greeting` server template with sqlite persistence, generated
  configuration, actor authentication, session-token tests, and testkit coverage
- Added generated-example build/test/vet CI plus Linux and Windows launcher preflight jobs
- Added local English/Chinese Skill-tree structure, link, launcher, and example consistency
  validation

### Changed

- Standardized new server projects on root `skel/`, `skeled/{golang,typescript}/`,
  and Hub seed plus hand-written server code under `src/server/`
- Deferred frontend creation until after server completion and explicit user confirmation;
  confirmed frontends are added under `src/web/`
- Updated native launcher defaults and preflight checks for the server-first layout
- Replaced the unsupported JSON version probe with the target CLI's supported `vine version`
  output
- Aligned the server example with the same-project `src/web/` policy and made real
  Portal/RPCGW verification mandatory before runnable-delivery claims

### Fixed

- Removed stale root-level `web/`, `cmd/`, and `seed/` assertions from package validation
- Corrected the example's authentication evidence so unit tests cover token validation,
  testkit covers post-admission actor behavior, and Portal admission requires real E2E proof

## [0.1.0] - 2026-08-05

### Added

- Version-aware Vine development workflow with impact-scoped generation gates
- Rpc/vRPC-first capability guidance and generated TypeScript frontend boundaries
- Standalone, linked, and separated topology-aware validation rules
- References for foundations, capabilities, frontend vRPC, execution boundaries,
  data and testing, runtime operations, and official source priorities
- Native Linux Bash and Windows Command Prompt launcher templates
- Synchronized Chinese and English Skill definitions
- English and Simplified Chinese repository documentation

### Security

- Explicit read-only behavior for diagnosis and review requests
- Guidance for sensitive-field redaction, execution scope, internal endpoint exposure,
  authentication, TLS, Redis leases, and database migration boundaries

[0.4.0]: https://github.com/yorun-ai/vine-skill/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/yorun-ai/vine-skill/compare/v0.2.2...v0.3.0
[0.2.2]: https://github.com/yorun-ai/vine-skill/releases/tag/v0.2.2
