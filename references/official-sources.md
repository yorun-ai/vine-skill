# Vine Official Source Index

Priority order: source and GoDoc at the target project's pinned revision > corresponding release notes > the website's current `next` documentation. Before Vine 1.0, use tagged repository source for release-specific evidence and the website for current guidance.

## Research Baseline

Reviewed 2026-10-26. Release source is authoritative over website snippets.

- Vine: [`v0.32.5`](https://github.com/yorun-ai/vine/tree/v0.32.5); [release changes and upgrade notes](https://github.com/yorun-ai/vine/blob/v0.32.5/CHANGELOG.md).
- Skel/compiler: [`v0.34.0`](https://github.com/yorun-ai/skel/tree/v0.34.0), commit `1859e6b02251f7e32d68a8d7e52676c56897d9b2`; [release notes](https://github.com/yorun-ai/skel/blob/v0.34.0/CHANGELOG.md).
- vRPC Go: [`v0.15.1`](https://github.com/yorun-ai/vrpc-go/tree/v0.15.1); shared HTTP codecs and external Go clients.
- vRPC TS: [`v0.10.0`](https://github.com/yorun-ai/vrpc-ts/tree/v0.10.0); generated browser client runtime.
- Documentation source reviewed: [`45a9f599`](https://github.com/yorun-ai/vine-site/tree/45a9f599a6cf3cec5f365a0247536f8625cb7f35). Some older snippets still use retired names; verify against released source.
- Websites: [Vine](https://vine.yorun.ai/docs/), [Chinese Vine](https://vine.yorun.ai/zh-CN/docs/), [Skel](https://skel.yorun.ai/docs/).

## Source Checks for This Baseline

| Boundary | Pinned source |
| --- | --- |
| Standalone options and flags | [app/standalone/option.go](https://github.com/yorun-ai/vine/blob/v0.32.5/app/standalone/option.go) |
| Testkit seed and generic clients | [app/testkit](https://github.com/yorun-ai/vine/tree/v0.32.5/app/testkit) |
| Web proxy/assets facades | [core/web/api.go](https://github.com/yorun-ai/vine/blob/v0.32.5/core/web/api.go) |
| Seed entries, rules and config | [Hub seeder](https://github.com/yorun-ai/vine/tree/v0.32.5/internal/daemon/hub/src/server/mod/seeder) |
| Web timeout selection | [webtimeout](https://github.com/yorun-ai/vine/tree/v0.32.5/internal/daemon/webtimeout) |
| Runtime/compiler floor | [internal/core/skel/version.go](https://github.com/yorun-ai/vine/blob/v0.32.5/internal/core/skel/version.go) |
| Generated names and scalars | [Skel Go binding](https://github.com/yorun-ai/skel/tree/v0.34.0/internal/codegen/binding/golang), [types](https://github.com/yorun-ai/skel/tree/v0.34.0/types) |

Internal source links supply research evidence; applications depend on the public facades. For migration decisions use [baseline-upgrade.md](baseline-upgrade.md); for a new baseline, reproduce generation, selected-module compilation and actual Portal calls.

## Website Topic Routes

| Topic | English page |
| --- | --- |
| Overview | https://vine.yorun.ai/docs/ |
| Getting started | https://vine.yorun.ai/docs/getting-started |
| Compatibility | https://vine.yorun.ai/docs/compatibility |
| First application | https://vine.yorun.ai/docs/tutorial-first-app |
| First contract | https://vine.yorun.ai/docs/first-skel-contract |
| Project structure | https://vine.yorun.ai/docs/filetree |
| Application model | https://vine.yorun.ai/docs/application-model |
| Components and Modules | https://vine.yorun.ai/docs/components |
| Configuration | https://vine.yorun.ai/docs/configuration |
| Rpc guide | https://vine.yorun.ai/docs/guide/rpc |
| Web | https://vine.yorun.ai/docs/web |
| Event and Task | https://vine.yorun.ai/docs/events-and-tasks |
| Redis guide | https://vine.yorun.ai/docs/guide/redis |
| RDB guide | https://vine.yorun.ai/docs/guide/rdb |
| Logging and testing | https://vine.yorun.ai/docs/logging-and-testing |
| Architecture | https://vine.yorun.ai/docs/runtime-mechanisms |
| Lifecycle | https://vine.yorun.ai/docs/application-lifecycle |
| Execution model | https://vine.yorun.ai/docs/execution-model |
| Routing and readiness | https://vine.yorun.ai/docs/request-routing |
| Trace and timeout | https://vine.yorun.ai/docs/trace-timeout |
| Deployment modes | https://vine.yorun.ai/docs/deployment-modes |
| Production readiness | https://vine.yorun.ai/docs/production-readiness |
| Hub | https://vine.yorun.ai/docs/hub |
| Link | https://vine.yorun.ai/docs/link |
| Portal | https://vine.yorun.ai/docs/portal |
| CLI | https://vine.yorun.ai/docs/cli |
| Public packages | https://vine.yorun.ai/docs/core-packages |
| App API | https://vine.yorun.ai/docs/app |
| DI | https://vine.yorun.ai/docs/di |
| Container and Filter | https://vine.yorun.ai/docs/ctr |
| Meta, Context, and identity | https://vine.yorun.ai/docs/meta |
| Error handling | https://vine.yorun.ai/docs/ex |
| Rpc API | https://vine.yorun.ai/docs/rpc |
| vRPC over HTTP | https://vine.yorun.ai/docs/vrpc-http |
| Redis API | https://vine.yorun.ai/docs/redis |
| RDB API | https://vine.yorun.ai/docs/rdb |

Chinese pages insert `/zh-CN` before `/docs`. When reviewing documentation source, read the matching file under `docs/` in the `vine-site` repository; use `i18n/zh-CN/docusaurus-plugin-content-docs/current/` only when comparing the Chinese translation. For exact APIs, read public GoDoc and tests under `app/`, `core/`, and `infra/` at the target Vine tag. Use public package symbols in applications.
