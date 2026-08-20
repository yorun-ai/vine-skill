# Contributing to vine-skill

Thank you for improving vine-skill. Contributions are licensed under Apache-2.0 as described in `LICENSE`.

## Repository scope

The maintained deliverable is this Skill package:

- `SKILL.md`: core workflow and routing rules
- `references/`: detailed, version-sensitive guidance
- `scripts/`: native launcher templates copied into generated projects
- `agents/openai.yaml`: Codex and ChatGPT presentation adapter, optional in the Agent Skills standard and required by this package
- `.github/workflows/ci.yml`: required package validation

Generated Vine applications used for manual validation are disposable fixtures, not the primary product. Do not commit business databases, installed dependencies, build output, credentials, or other runtime artifacts as repository documentation or release content.

## Contribution principles

- Preserve version-first behavior. Never silently install or upgrade Vine, skelc, Go, or infrastructure while inspecting a target project.
- Keep Skel-generated Go and TypeScript boundaries authoritative.
- Keep new-project delivery server-first with contracts in root `skel/`, generated artifacts in `skeled/`, and Hub seed plus hand-written server code under `src/server/`.
- Add `src/web/` only after the server is complete and the user explicitly confirms a frontend.
- Keep structured browser calls on generated vRPC clients; reserve Web/HTTP for binary data and static assets.
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
6. Record notable user-visible changes under `Unreleased` in `CHANGELOG.md`.
7. Run validation that matches the affected files.

## Updating version facts

The current tested toolchain baseline is Vine `v0.13.1`, skelc `v0.14.0`, and Go `1.26.6` or later; Vine `v0.13.1` itself retains a Go `1.26.5` module floor. A version update must include:

- primary-source evidence from the matching release, source revision, GoDoc, tests, or release notes;
- synchronized version claims across `SKILL.md`, relevant references, and both repository READMEs;
- review of affected CLI flags, generated code, lifecycle behavior, and runtime topology;
- updates to `references/official-sources.md` when source priority or links change;
- an entry in `CHANGELOG.md`.

Do not use the website's `next` documentation as proof for an older pinned target revision.

## Validation

At minimum, validate the files affected by the contribution:

```bash
python3 scripts/validate_skill.py
bash -n scripts/start_vine_app.sh
```

When the paired Simplified Chinese Skill tree is present at `../../vine-skill`, the package
validator also checks matching Markdown structure, local links, launcher identity, and example
artifact behavior across both trees.

Also verify:

- `SKILL.md` has valid frontmatter and the expected Skill name;
- `SKILL.md` remains free of Codex-, Claude Code-, and OpenCode-specific paths or invocation syntax;
- every relative Markdown link resolves from the file that contains it;
- English and Simplified Chinese repository documentation stays aligned;
- Linux and Windows launcher environment variables and lifecycle behavior remain aligned;
- realistic prompts route to the correct reference material;
- no generated fixture, database, dependency directory, build output, token, or private key is included.

When agent-host compatibility changes, install the package into a disposable project for Codex,
Claude Code, and OpenCode. Confirm discovery, explicit invocation, linked-reference loading, a
read-only task, and capability-denial behavior. Record the client and installer versions; static
validation alone does not prove end-to-end host behavior.

CI regenerates `references/example-greeting` with skelc `v0.14.0`, builds and tests it against
Vine `v0.13.1`, runs `go vet`, rejects unformatted Go, and exercises the Windows launcher
`--check` path with a disposable fixture. It also installs the package into a disposable project
for Codex, Claude Code, and OpenCode and verifies both shared and Claude-specific skill copies.

When launcher behavior changes, test `--check`, first-run installation, readiness, occupied-port refusal, and cleanup in a disposable Vine browser-app fixture. Review Windows Batch changes on Windows or in Windows CI when available; otherwise report that limitation explicitly.

## Pull requests

- Keep each pull request focused on one coherent Skill change.
- Explain the behavior being changed and why it belongs in the core workflow or a specific reference.
- List the exact validation commands run and their results.
- Identify version, platform, topology, or network behavior that was not verified.
- Identify every agent host and host version that was or was not exercised.
- Avoid unrelated formatting or generated-file churn.

Unless explicitly stated otherwise, submitted contributions are licensed under Apache-2.0.
