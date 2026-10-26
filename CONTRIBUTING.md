# Contributing to vine-skill

Thank you for improving vine-skill. Contributions are licensed under Apache-2.0 as described in `LICENSE`.

## Repository scope

The maintained deliverable is this Skill package:

- `SKILL.md`: core workflow and routing rules
- `references/`: detailed, version-sensitive guidance
- `scripts/`: package validation tools, fixtures, and native launcher templates
- `agents/openai.yaml`: Codex and ChatGPT presentation adapter, optional in the Agent Skills standard and required by this package
- `.github/workflows/ci.yml`: required package validation

Generate validation applications in disposable directories. Keep business databases, installed dependencies, build output, credentials and runtime artifacts within those directories. Repository and release content consist of the maintained Skill sources and fixtures.

## Contribution principles

- Preserve version-first behavior. Inspect the target project's current toolchain; install or upgrade tools and infrastructure when the user's request covers those changes.
- Keep Skel-generated Go and TypeScript boundaries authoritative.
- Keep new-project delivery server-first with contracts in root `skel/`, generated artifacts in `skeled/`, and Hub seed plus hand-written server code under `src/server/`.
- Add `src/web/` after server validation when the user has requested a frontend; an explicit request already supplies confirmation.
- Keep structured browser calls on generated vRPC clients; use Web/HTTP for binary data, static assets, and required HTTP/streaming protocols.
- Match validation claims to the actual runtime topology and commands executed.
- Keep the core workflow concise and route detailed material into the appropriate reference file.
- Keep `SKILL.md` agent-host neutral. Put discovery paths, invocation syntax, permissions, and host-specific metadata in `references/agent-hosts.md` or the host adapter.
- Keep `README.md` and `README.zh-CN.md` aligned.
- Keep launcher templates dependency-free and behaviorally aligned across Linux and Windows.

## Making a change

1. Identify whether the change affects core routing, a version-specific reference, agent-host compatibility, launcher behavior, metadata, or translation only.
2. Update `SKILL.md` when the core workflow or routing table changes.
3. Update only the relevant file under `references/` for detailed guidance.
4. Update `references/agent-hosts.md` when a supported host changes discovery, invocation, permissions, or its tested compatibility baseline.
5. Update both `README.md` and `README.zh-CN.md` when repository-level documentation changes.
6. Describe user-visible changes in the review/PR. Update `CHANGELOG.md` only during release preparation, summarizing the final changes since the previous tag.
7. Run validation that matches the affected files.

## Updating version facts

The current tested toolchain baseline is Vine `v0.32.5`, skelc `v0.34.0`, and Go `1.27.3`; the module floor is `1.27.0`. External clients are tested with vRPC Go `v0.15.1` and vRPC TS `0.10.0`. A version update must include:

- primary-source evidence from the matching release, source revision, GoDoc, tests, or release notes;
- synchronized version claims across `SKILL.md`, relevant references, and both repository READMEs;
- review of affected CLI flags, generated code, lifecycle behavior, and runtime topology;
- updates to `references/official-sources.md` when source priority or links change;
- reproducible generation/build/admission evidence; a changelog entry only when preparing a release.

Use the pinned target revision's source, tests and release notes as compatibility evidence.

## Validation

At minimum, validate the files affected by the contribution:

```bash
python3 scripts/validate_skill.py
python3 -m unittest discover -s scripts -p 'test_*.py'
bash -n scripts/start_vine_app.sh
# For baseline, contract, example, or public API changes:
bash scripts/validate_example.sh
```

When the paired Simplified Chinese Skill tree is present at `../../vine-skill`, the package
validator also checks matching Markdown structure, local links, launcher identity, and example
artifact behavior across both trees.

Also verify:

- `SKILL.md` has valid frontmatter and the expected Skill name;
- `SKILL.md` describes the shared workflow, with host-specific paths and invocation syntax in the host reference and adapter;
- every relative Markdown link resolves from the file that contains it;
- English and Simplified Chinese repository documentation stays aligned;
- Linux and Windows launcher environment variables and lifecycle behavior remain aligned;
- realistic prompts route to the correct reference material;
- generated fixtures, databases, dependencies, build output and credentials stay in disposable validation directories.

When agent-host compatibility changes, install the package into a disposable project for Codex,
Claude Code, and OpenCode. Confirm discovery, explicit invocation, linked-reference loading, a
read-only task, and capability-denial behavior. Record the client and installer versions, static
format checks and observed host behavior as separate validation results.

The example validation script works in a temporary directory, checks the actual module graph, compiles the frontend reference snippets, and exercises generated TypeScript clients and the real Hub Admin listener. It also generates `references/example-domains` with separate public/regular Go modules and tests bidirectional Rpc and Event delivery between two bundled Apps, including race checks. CI regenerates `references/example-greeting` with skelc `v0.34.0`, builds and tests it against
Vine `v0.32.5`, runs `go vet`, rejects unformatted Go, and exercises the Windows launcher
`--check` path with a disposable fixture. It also installs the package into a disposable project
for Codex, Claude Code, and OpenCode and verifies both shared and Claude-specific skill copies.

When launcher behavior changes, test `--check`, first-run installation, readiness, occupied-port refusal, and cleanup in a disposable Vine browser-app fixture. Review Windows Batch changes on Windows or in Windows CI when available; otherwise report that limitation explicitly.

## Pull requests

- Keep each pull request focused on one coherent Skill change.
- Explain the behavior being changed and why it belongs in the core workflow or a specific reference.
- List the exact validation commands run and their results.
- Record validation coverage and gaps for versions, platforms, topologies and network behavior.
- Record each agent host and version exercised, along with remaining host coverage.
- Keep formatting and generated-file changes traceable to the requested change.

Unless explicitly stated otherwise, submitted contributions are licensed under Apache-2.0.
