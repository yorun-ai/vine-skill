# Changelog

All notable changes to the vine-skill package are documented in this file.

The project follows [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Changed

- Updated the tested baseline to Vine `v0.13.1` and skelc `v0.14.0` while retaining Vine's released Go `1.26.5` minimum
- Documented the Vine `v0.13.0` external-NATS stream-provisioning boundary, including required `VINE_EVENTS`/`VINE_TASKS` retention and deployment-owned storage
- Documented Vine `v0.13.1` and skelc `v0.12.0+` typed in-process Rpc cloning, the compatibility fallback for older generated specs, and the distinction between value-isolation and wire tests
- Refreshed the official Vine, skelc, and website source revisions and updated example/CI pins
- Replaced the manual Codex Skill clone instructions with `npx skills add yorun-ai/vine-skill`

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
